package prompts

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"geolineai/internal/core/assist"
)

//go:embed jev/*.json llm/*.txt sources/*.json settings.toml
var files embed.FS

// Values are the tunable values, by the model they tune; every one is required, and
// an unknown one is an error.
type Values struct {
	Jev struct {
		Model        string  `toml:"model"`
		SupportedMin float64 `toml:"supported_min"`
		VerdictMax   float64 `toml:"verdict_max"`
	} `toml:"jev"`
	LLM struct {
		Model              string  `toml:"model"`
		Temperature        float64 `toml:"temperature"`
		QuestionCharacters int     `toml:"question_characters"`
		ExplainSentences   int     `toml:"explain_sentences"`
		BlankAnswerRetries int     `toml:"blank_answer_retries"`
		TimeoutSeconds     int     `toml:"timeout_seconds"`
		RetryStatuses      []int   `toml:"retry_statuses"`
		Retries            int     `toml:"retries"`
		RetryWaitSeconds   int     `toml:"retry_wait_seconds"`
	} `toml:"llm"`
}

// Passage is a piece of the repository's documentation an explanation may cite, quoted
// verbatim apart from its emphasis and code marks, with the document and the heading
// anchor it sits under and a title saying what it is.
type Passage struct {
	Path   string `json:"path"`
	Anchor string `json:"anchor"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

var (
	// Settings are the tunable values.
	Settings Values
	// ExplainSystem tells the text model what to write and what it may not.
	ExplainSystem string
	// JudgeSupported and JudgeVerdict are asked of each written sentence over only the
	// sources it cites: whether they support it, and whether it tells a verdict.
	JudgeSupported, JudgeVerdict assist.Question
	// Passages are the documentation an explanation may cite, in the order offered.
	Passages []Passage
	// Fingerprint identifies the configuration built into the binary: the models, the
	// writer's temperature, the Jev thresholds, the sentence, question and retry limits,
	// the system prompt, the judgment questions and the passages. It changes exactly
	// when one of them does. Call retries and the timeout are operational, left out.
	Fingerprint string
)

// The files are built into the binary, so one that does not read is a mistake in the
// build, and the process stops naming it.
func init() {
	raw, err := files.ReadFile("settings.toml")
	if err == nil {
		dec := toml.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		err = dec.Decode(&Settings)
	}
	if err == nil {
		err = Settings.check()
	}
	if err != nil {
		panic("prompts: the settings: " + err.Error())
	}
	ExplainSystem = text("llm/explain-system.txt")
	JudgeSupported, JudgeVerdict = question("jev/judge-supported.json"), question("jev/judge-verdict.json")
	decode("sources/explain-passages.json", &Passages)
	Fingerprint = fingerprint()
}

// check names every value that is missing or out of its range.
func (v Values) check() error {
	var problems []string
	need := func(ok bool, rule string) {
		if !ok {
			problems = append(problems, rule)
		}
	}
	need(v.Jev.Model != "" && v.LLM.Model != "", "jev.model and llm.model are required")
	for _, t := range []float64{v.Jev.SupportedMin, v.Jev.VerdictMax} {
		need(t > 0 && t <= 1, fmt.Sprintf("threshold %v must be above 0 and at most 1", t))
	}
	need(v.LLM.QuestionCharacters > 0, "llm.question_characters must be above 0")
	need(v.LLM.Temperature >= 0, "llm.temperature must be 0 or more")
	need(v.LLM.ExplainSentences > 0, "llm.explain_sentences must be above 0")
	need(v.LLM.BlankAnswerRetries >= 0, "llm.blank_answer_retries must be 0 or more")
	need(v.LLM.TimeoutSeconds > 0, "llm.timeout_seconds must be above 0")
	need(v.LLM.Retries >= 0 && v.LLM.RetryWaitSeconds >= 0, "llm.retries and llm.retry_wait_seconds must be 0 or more")
	if len(problems) > 0 {
		return fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return nil
}

// text is a prompt file's text, without the newline the file ends with.
func text(name string) string {
	raw, err := files.ReadFile(name)
	if err != nil {
		panic("prompts: " + err.Error())
	}
	return strings.TrimSuffix(string(raw), "\n")
}

// question is a judgment question file: its kind, which must be noul, its
// instructions, which may be any JSON value, and the criteria naming each answer.
func question(name string) assist.Question {
	var q struct {
		Type         assist.Kind       `json:"type"`
		Instructions any               `json:"instructions"`
		Criteria     map[string]string `json:"criteria"`
	}
	decode(name, &q)
	if q.Type != assist.KindNoul || q.Instructions == nil || len(q.Criteria) == 0 {
		panic("prompts: " + name + " needs a type of noul, instructions and criteria")
	}
	return assist.Question{Kind: q.Type, Instructions: q.Instructions, Criteria: q.Criteria}
}

// decode reads a JSON file into v, refusing a field v does not have.
func decode(name string, v any) {
	raw, err := files.ReadFile(name)
	if err == nil {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		err = dec.Decode(v)
	}
	if err != nil {
		panic("prompts: " + name + ": " + err.Error())
	}
}

// fingerprint hashes the configuration, keyed by logical names rather than by file,
// folder or settings section, in an encoding that orders every map by key: moving a
// file, renaming a section, reordering a file's keys or reindenting it leaves it
// unchanged, while any change in what a model receives does not. The names below are
// part of the fingerprint, so renaming one is a new fingerprint.
func fingerprint() string {
	type models struct {
		Judge, Writer     string
		WriterTemperature float64
	}
	type thresholds struct{ SupportedMin, VerdictMax float64 }
	type limits struct{ ExplainSentences, QuestionCharacters, BlankAnswerRetries int }
	configuration, err := json.Marshal(struct {
		Models                       models
		Thresholds                   thresholds
		Limits                       limits
		ExplainSystem                string
		JudgeSupported, JudgeVerdict assist.Question
		Passages                     []Passage
	}{models{Settings.Jev.Model, Settings.LLM.Model, Settings.LLM.Temperature},
		thresholds{Settings.Jev.SupportedMin, Settings.Jev.VerdictMax},
		limits{Settings.LLM.ExplainSentences, Settings.LLM.QuestionCharacters, Settings.LLM.BlankAnswerRetries},
		ExplainSystem, JudgeSupported, JudgeVerdict, Passages})
	if err != nil {
		panic("prompts: " + err.Error())
	}
	sum := sha256.Sum256(configuration)
	return "sha256:" + hex.EncodeToString(sum[:])
}
