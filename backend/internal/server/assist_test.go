package server_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/intake/sourcetest"
	"geolineai/internal/testutil"
)

// sourceTexts maps each source's title to its text, and names the ID of the source
// with the given title.
func sourceTexts(t *testing.T, answer map[string]any, title string) (map[string]string, string) {
	t.Helper()
	out, id := map[string]string{}, ""
	for _, s := range answer["sources"].([]any) {
		src := s.(map[string]any)
		out[src["title"].(string)] = src["text"].(string)
		if src["title"] == title {
			id = src["id"].(string)
		}
	}
	return out, id
}

const contactTitle = "Fact: failing transmitter contact of line-1-e10"

// ask posts a question about one reading and returns the answer, which must be a 200.
func ask(t *testing.T, a *testutil.App, id string) map[string]any {
	t.Helper()
	status, out := a.Do(http.MethodPost, "/api/readings/"+id+"/explanations", `{"question":"Why?"}`)
	require.Equal(t, http.StatusOK, status, out)
	return out
}

// A fact an answer cites is shown to the reviewer as true; a wrong count of failing
// partners, a repeat marked as counting that is left out, or a missed decay rise
// would be believed.
func TestExplanationFactsStateTheStoredEvidence(t *testing.T) {
	a, ids := advisoryFixture(t)
	a.Models.Written = `{"answerable":false,"sentences":[]}`
	for _, c := range []struct{ reading, title, want string }{
		{"uses10", contactTitle, "On line-1, line-1-e10 transmits in 2 readings with Violated Transmitter contact: 1 with partner line-1-e11, 1 with partner line-1-e12. It has 2 different partners in those readings."},
		{"neg", "Fact: repeats of this reading", "Repeats of neg, the other readings of its import with the same four electrode IDs: ok1 at 100 ohm-m, counts toward the reference; ok2 at 100 ohm-m, counts toward the reference; far at 150 ohm-m, counts toward the reference; missing with no resistivity, left out of the reference. 3 of its 4 repeats count toward the reference."},
		{"missing", "Fact: which criteria recorded a result", "Criteria with a recorded result for missing: Chargeability error, Transmitter contact, Decay trend, Suspect electrode. Criteria with no recorded result for missing: Resistivity sign, Resistivity error, Frozen repeat, Repeat agreement."},
	} {
		out := ask(t, a, ids[c.reading])
		texts, _ := sourceTexts(t, out, c.title)
		assert.Equal(t, c.want, texts[c.title], c.reading)
		assert.Equal(t, "unanswerable", out["outcome"])
		assert.Nil(t, out["reason"], "only an unavailable answer has a reason")
	}
	rise := testutil.NewApp(t)
	rise.Import("/data/line-1.csv", sourcetest.CSV(sourcetest.Row(map[string]string{"id": "rise", "decay_curve": "5;6;4;3;2;1;1;1;1;0"})))
	rise.Models.Written = `{"answerable":false,"sentences":[]}`
	out := ask(t, rise, rise.IDsByAcquisition()["rise"])
	texts, _ := sourceTexts(t, out, "")
	assert.Equal(t, "Decay samples of rise, gate 1 to gate 10 at 24.32 to 822.34 ms: 5, 6, 4, 3, 2, 1, 1, 1, 1, 0 V/V. A sample rises above the one before it from gate 1 to gate 2 (5 to 6 V/V).", texts["Fact: decay samples of this reading"])
}

// Every source an answer is offered can be inspected on the page: a doc source names
// the repository document it quotes, and no source points anywhere else.
func TestExplanationSourcesResolveOnThePage(t *testing.T) {
	a, ids := advisoryFixture(t)
	a.Models.Written = `{"answerable":false,"sentences":[]}`
	out := ask(t, a, ids["neg"])
	docs := 0
	for i, s := range out["sources"].([]any) {
		src := s.(map[string]any)
		assert.Equal(t, fmt.Sprintf("S%d", i+1), src["id"], "numbered in order")
		assert.NotContains(t, src, "listing", "no link into an out-of-scope listing")
		if src["kind"] == "doc" {
			docs++
			doc := src["doc"].(string)
			assert.True(t, strings.HasPrefix(doc, "docs/") && strings.Contains(doc, ".md#"), "a repository-relative doc path with its section: %s", doc)
		} else {
			assert.Nil(t, src["doc"], src["title"])
		}
	}
	assert.Positive(t, docs, "the documentation passages are offered")
}

// An answer is shown only when every sentence passes every check; one sentence that
// states what its sources do not, cites what is not a source, recommends a verdict or
// that the judgment model doubts, must withhold the whole answer, or the reviewer
// reads an unsupported claim as checked.
func TestExplanationIsShownOnlyWhenEverySentencePasses(t *testing.T) {
	a, ids := advisoryFixture(t)
	a.Models.Written = `{"answerable":false,"sentences":[]}`
	_, fact := sourceTexts(t, ask(t, a, ids["uses10"]), contactTitle)
	require.NotEmpty(t, fact)
	sound := "line-1-e10 transmits in 2 readings with Violated Transmitter contact."
	for _, c := range []struct {
		name               string
		sentences          []string
		cite               string
		supported, verdict float64
		outcome, problem   string
	}{
		{"sound, at both thresholds", []string{sound}, fact, 0.5, 0.49, "shown", ""},
		{"judged unsupported", []string{sound}, fact, 0.49, 0.1, "withheld", "The judgment model did not find it supported by its sources."},
		{"a verdict recommendation", []string{"Mark this reading Fail."}, fact, 0.9, 0.5, "withheld", "The judgment model found it may tell the reviewer what verdict to record."},
		{"a number its source does not state", []string{"line-1-e10 transmits in 3 readings with Violated Transmitter contact."}, fact, 0.9, 0.1, "withheld", "The number 3 is not in its sources."},
		{"an identifier its source does not name", []string{"line-1-e13 transmits with Violated Transmitter contact."}, fact, 0.9, 0.1, "withheld", `"line-1-e13" is not in its sources.`},
		{"a citation that is not a source", []string{sound}, "S9999", 0.9, 0.1, "withheld", "It cites S9999, which is not one of its sources."},
		{"a hedged diagnosis", []string{"line-1-e10 is likely faulty."}, fact, 0.9, 0.1, "withheld", `It uses a word starting "likely", which its sources do not use.`},
		{"a rank its source does not state", []string{"line-1-e10 is the worst electrode."}, fact, 0.9, 0.1, "withheld", `It uses a word starting "worst", which its sources do not use.`},
		{"one bad sentence among good ones", []string{sound, "line-1-e10 transmits in 4 readings."}, fact, 0.9, 0.1, "withheld", "The number 4 is not in its sources."},
	} {
		var sentences []string
		for _, s := range c.sentences {
			sentences = append(sentences, fmt.Sprintf(`{"text":%q,"sources":[%q]}`, s, c.cite))
		}
		a.Models.Written = `{"answerable":true,"sentences":[` + strings.Join(sentences, ",") + `]}`
		a.Models.Nouls = map[string]float64{"supported": c.supported, "verdict": c.verdict}
		out := ask(t, a, ids["uses10"])
		assert.Equal(t, c.outcome, out["outcome"], c.name)
		if c.problem != "" {
			assert.Contains(t, fmt.Sprint(out["sentences"]), c.problem, c.name)
		}
		if c.name == "one bad sentence among good ones" {
			first := out["sentences"].([]any)[0].(map[string]any)
			assert.Empty(t, first["problems"], "the good sentence passed on its own")
		}
	}
}

// Unanswerable is a valid refusal; unavailable is no answer at all, and says why. A
// malformed reply or one missing a required field must never read as a refusal, and a
// missing key or a provider failure must not hide the evidence.
func TestExplanationOutcomesTellRefusalFromFailure(t *testing.T) {
	a, ids := advisoryFixture(t)
	for _, c := range []struct {
		name            string
		written         string
		unavailable     bool
		invalid         bool
		outcome, reason string
	}{
		{"a valid refusal", `{"answerable":false,"sentences":[]}`, false, false, "unanswerable", ""},
		{"malformed JSON", `{"answerable":tru`, false, false, "unavailable", "invalid_response"},
		{"answerable missing", `{"sentences":[]}`, false, false, "unavailable", "invalid_response"},
		{"an answer with no sentence", `{"answerable":true,"sentences":[]}`, false, false, "unavailable", "invalid_response"},
		{"sentences missing", `{"answerable":true}`, false, false, "unavailable", "invalid_response"},
		{"a sentence without sources", `{"answerable":true,"sentences":[{"text":"x"}]}`, false, false, "unavailable", "invalid_response"},
		{"answerable of the wrong type", `{"answerable":"no","sentences":[]}`, false, false, "unavailable", "invalid_response"},
		{"an unreadable judgment", `{"answerable":true,"sentences":[{"text":"x","sources":["S1"]}]}`, false, true, "unavailable", "invalid_response"},
		{"the provider failing", "", true, false, "unavailable", "provider_error"},
	} {
		a.Models.Written, a.Models.Unavailable, a.Models.Invalid = c.written, c.unavailable, c.invalid
		out := ask(t, a, ids["neg"])
		assert.Equal(t, c.outcome, out["outcome"], c.name)
		if c.reason == "" {
			assert.Nil(t, out["reason"], c.name)
		} else {
			assert.Equal(t, c.reason, out["reason"], c.name)
			assert.Equal(t, []any{}, out["sentences"], "%s: no sentence of a failed answer", c.name)
		}
		assert.NotEmpty(t, out["sources"], "%s: the evidence is still offered", c.name)
	}

	a.Models = nil
	a.Reopen()
	out := ask(t, a, ids["neg"])
	assert.Equal(t, "unavailable", out["outcome"])
	assert.Equal(t, "no_key", out["reason"])
	assert.NotEmpty(t, out["sources"])
	assert.Equal(t, "fail", detailOf(t, a, ids["neg"])["advisory"].(map[string]any)["value"], "inspection works without a key")
	status, saved := save(a, "pass", ids["neg"])
	require.Equal(t, http.StatusOK, status, saved)
	assert.NotNil(t, saved["action_id"], "saving works without a key")
}

// Asking never writes: whatever the models answer, every verdict and the review
// record stay exactly as they were.
func TestAskingChangesNoReviewState(t *testing.T) {
	a, ids := advisoryFixture(t)
	save(a, "fail", ids["far"])
	verdicts := func() string {
		var out []string
		for _, d := range allDetails(t, a) {
			out = append(out, d["verdict"].(string))
		}
		return strings.Join(out, ",")
	}
	before := verdicts()
	_, fact := sourceTexts(t, ask(t, a, ids["uses10"]), contactTitle)
	a.Models.Nouls = map[string]float64{"supported": 0.9, "verdict": 0.1}
	for _, written := range []string{
		fmt.Sprintf(`{"answerable":true,"sentences":[{"text":"line-1-e10 transmits in 2 readings with Violated Transmitter contact.","sources":[%q]}]}`, fact),
		fmt.Sprintf(`{"answerable":true,"sentences":[{"text":"Save Pass for 9 readings.","sources":[%q]}]}`, fact),
		`{"answerable":false,"sentences":[]}`,
		`not json`,
	} {
		a.Models.Written = written
		ask(t, a, ids["uses10"])
	}
	a.Models.Unavailable = true
	ask(t, a, ids["uses10"])
	assert.Equal(t, before, verdicts())
	assert.Equal(t, 1, a.Count("review_actions"))
	assert.Equal(t, 1, a.Count("review_history"))
}
