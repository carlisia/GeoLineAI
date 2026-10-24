package grounded

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"golang.org/x/sync/errgroup"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/assist/prompts"
)

// The judgment thresholds: a sentence passes when its supported probability is at
// least supportedMin and its tells-a-verdict probability is below verdictMax.
// Changing either changes the configuration fingerprint.
var (
	supportedMin = prompts.Settings.Jev.SupportedMin
	verdictMax   = prompts.Settings.Jev.VerdictMax
)

// blankAnswerRetries is how many more times the text model is asked when its answer
// claims to answer but holds no sentence with text and a citation.
var blankAnswerRetries = prompts.Settings.LLM.BlankAnswerRetries

// Request is one question for the text model: the system message saying what to
// write, the task fields that go beside the sources in the user message, the sources
// it may cite, and the most sentences it may write.
type Request struct {
	System       string
	Task         map[string]any
	Sources      []Source
	MaxSentences int
}

// schema holds the text model to an answer the checks can read: whether the sources
// answer the question, and each sentence with the IDs of the sources it relies on.
var schema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"answerable", "sentences"},
	"properties": map[string]any{
		"answerable": map[string]any{"type": "boolean"},
		"sentences": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"text", "sources"},
				"properties": map[string]any{
					"text":    map[string]any{"type": "string"},
					"sources": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
		},
	},
}

// The judgment questions, asked of each sentence over only the sources it cites.
var (
	supportedQuestion = prompts.JudgeSupported
	verdictQuestion   = prompts.JudgeVerdict
)

// writtenAnswer is the object the text model writes, as decoded: every field is a
// pointer, so an omitted one is told from false or empty, and parse refuses it. A
// reply without "answerable" is not a refusal, and one without "sentences" is not an
// empty answer.
type writtenAnswer struct {
	Answerable *bool `json:"answerable"`
	Sentences  *[]struct {
		Text    *string   `json:"text"`
		Sources *[]string `json:"sources"`
	} `json:"sentences"`
}

// parsed is a written answer once parse has found every required field.
type parsed struct {
	answerable bool
	sentences  []writtenSentence
}

type writtenSentence struct {
	text    string
	sources []string
}

// parse decodes the text model's object and refuses one that does not decode or omits
// a required field, as an invalid response.
func parse(raw json.RawMessage) (parsed, error) {
	var a writtenAnswer
	if err := json.Unmarshal(raw, &a); err != nil {
		return parsed{}, fmt.Errorf("%w: the answer does not decode: %v", assist.ErrInvalidResponse, err)
	}
	missing := func(field string) (parsed, error) {
		return parsed{}, fmt.Errorf("%w: the answer has no %s field", assist.ErrInvalidResponse, field)
	}
	switch {
	case a.Answerable == nil:
		return missing("answerable")
	case a.Sentences == nil:
		return missing("sentences")
	}
	out := parsed{answerable: *a.Answerable}
	for i, s := range *a.Sentences {
		switch {
		case s.Text == nil:
			return missing(fmt.Sprintf("sentences[%d].text", i))
		case s.Sources == nil:
			return missing(fmt.Sprintf("sentences[%d].sources", i))
		}
		out.sentences = append(out.sentences, writtenSentence{text: *s.Text, sources: *s.Sources})
	}
	return out, nil
}

// Ask has the text model answer from the sources and checks every sentence it
// writes. With no models (no provider key), a model that cannot be reached, or an
// answer that cannot be read, the outcome is Unavailable with its reason; the error is
// for a failure of the server's own, such as a cancelled request.
func Ask(ctx context.Context, w assist.Writer, j assist.Judge, req Request) (Answer, error) {
	if w == nil || j == nil {
		return unavailable(req, NoKey, nil), nil
	}
	out, err := ask(ctx, w, j, req)
	switch {
	case errors.Is(err, assist.ErrInvalidResponse):
		return unavailable(req, InvalidResponse, err), nil
	case errors.Is(err, assist.ErrUnavailable):
		return unavailable(req, ProviderError, err), nil
	}
	return out, err
}

// unavailable is the answer when no answer could be obtained: the sources, so the
// evidence can still be read, and the reason. The cause, which the answer never
// carries, is logged here, the one place every model failure passes through.
func unavailable(req Request, reason UnavailableReason, cause error) Answer {
	if cause != nil {
		slog.Error("model answer unavailable", slog.String("reason", string(reason)), slog.String("error", cause.Error()))
	}
	return Answer{Outcome: Unavailable, Reason: &reason, Sentences: []Sentence{}, Sources: req.Sources}
}

// ask is Ask once the models are known to be configured.
func ask(ctx context.Context, w assist.Writer, j assist.Judge, req Request) (Answer, error) {
	type offered struct {
		ID    string     `json:"id"`
		Kind  SourceKind `json:"kind"`
		Title string     `json:"title"`
		Text  string     `json:"text"`
	}
	input := maps.Clone(req.Task)
	list := make([]offered, len(req.Sources))
	byID := make(map[string]Source, len(req.Sources))
	for i, s := range req.Sources {
		list[i] = offered{s.ID, s.Kind, s.Title, s.Text}
		byID[s.ID] = s
	}
	input["sources"] = list
	user, err := json.Marshal(input)
	if err != nil {
		return Answer{}, err
	}
	var got parsed
	var writer string
	// The text model now and then claims an answer but writes only blank or uncited
	// sentences; asked again, it answers. Such an answer could only be withheld, so it is
	// asked for again, up to the configured retries, and the last answer meets the
	// same checks.
	for attempt := 0; attempt <= blankAnswerRetries && (attempt == 0 || degenerate(got)); attempt++ {
		var raw json.RawMessage
		if raw, writer, err = w.Write(ctx, req.System, string(user), schema); err != nil {
			return Answer{}, err
		}
		if got, err = parse(raw); err != nil {
			return Answer{}, err
		}
	}
	out := Answer{Outcome: Shown, Sentences: []Sentence{}, Sources: req.Sources, Models: Models{Writer: writer}}
	if !got.answerable {
		out.Outcome = Unanswerable
		return out, nil
	}
	// An answer that says it can answer and writes nothing is not a refusal and has no
	// sentence to withhold: the reply cannot be read as either.
	if len(got.sentences) == 0 {
		return Answer{}, fmt.Errorf("%w: the answer claims to answer but has no sentences", assist.ErrInvalidResponse)
	}
	// Each sentence's cited sources, resolved once for the code checks and the judge.
	cited := make([][]Source, len(got.sentences))
	for i, s := range got.sentences {
		for _, id := range s.sources {
			if src, ok := byID[id]; ok {
				cited[i] = append(cited[i], src)
			}
		}
		out.Sentences = append(out.Sentences, Sentence{Text: s.text, Sources: s.sources, Problems: codeProblems(s.text, cited[i], byID, s.sources)})
	}
	if err := judge(ctx, j, cited, &out); err != nil {
		return Answer{}, err
	}
	if len(out.Sentences) > req.MaxSentences {
		out.Outcome = Withheld
		out.Sentences[0].Problems = append(out.Sentences[0].Problems, fmt.Sprintf("The answer has %d sentences; at most %d are allowed.", len(out.Sentences), req.MaxSentences))
	}
	for _, s := range out.Sentences {
		if len(s.Problems) > 0 {
			out.Outcome = Withheld
		}
	}
	return out, nil
}

// degenerate reports an answer that says it can answer but whose every sentence is
// blank or cites no source: nothing a check could pass.
func degenerate(a parsed) bool {
	if !a.answerable {
		return false
	}
	for _, s := range a.sentences {
		if strings.TrimSpace(s.text) != "" && len(s.sources) > 0 {
			return false
		}
	}
	return true
}

// judge asks the judgment model about every sentence at once, one request per
// sentence over only the sources it cites, and adds a problem to each sentence that
// fails a threshold. The first failure cancels the requests still running, since one
// failure leaves the whole answer unavailable.
func judge(ctx context.Context, j assist.Judge, cited [][]Source, out *Answer) error {
	g, ctx := errgroup.WithContext(ctx)
	models := make([]string, len(out.Sentences))
	for i := range out.Sentences {
		g.Go(func() error {
			s := &out.Sentences[i]
			sources := make([]map[string]string, len(cited[i]))
			for k, src := range cited[i] {
				sources[k] = map[string]string{"title": src.Title, "text": src.Text}
			}
			answers, model, err := j.Judge(ctx, map[string]any{"sources": sources, "sentence": s.Text},
				map[string]assist.Question{"supported": supportedQuestion, "verdict": verdictQuestion})
			if err != nil {
				return err
			}
			models[i] = model
			s.Supported, s.TellsVerdict = answers["supported"].Noul, answers["verdict"].Noul
			if s.Supported < supportedMin {
				s.Problems = append(s.Problems, "The judgment model did not find it supported by its sources.")
			}
			if s.TellsVerdict >= verdictMax {
				s.Problems = append(s.Problems, "The judgment model found it may tell the reviewer what verdict to record.")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	if len(models) > 0 {
		out.Models.Judge = models[0]
	}
	return nil
}
