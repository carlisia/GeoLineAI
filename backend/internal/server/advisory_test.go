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

// advisoryCSV holds line-1 rows that land on every advisory for a different reason.
// In the default geometry: neg (negative resistivity, so it fails and is left out of
// its repeats' reference), ok1 and ok2 (each exactly a fifth off its reference, at the
// limit and so agreeing), far (half off the reference, so it needs review) and missing (no resistivity, an import issue). In a second geometry,
// frozen1 and frozen2 copy each other to every digit. contact1 and contact2 transmit
// from e10 with failing contact to two different partners, making e10 suspect, and
// uses10 receives on e10 with sound values of its own.
var advisoryCSV = sourcetest.CSV(
	sourcetest.Row(map[string]string{"id": "neg", "apparent_resistivity": "-50", "chargeability": "0.011"}),
	sourcetest.Row(map[string]string{"id": "ok1", "apparent_resistivity": "100", "chargeability": "0.012"}),
	sourcetest.Row(map[string]string{"id": "ok2", "apparent_resistivity": "100", "chargeability": "0.013"}),
	sourcetest.Row(map[string]string{"id": "far", "apparent_resistivity": "150", "chargeability": "0.014"}),
	sourcetest.Row(map[string]string{"id": "missing", "apparent_resistivity": "", "chargeability": "0.015"}),
	sourcetest.Row(map[string]string{"id": "frozen1", "tx1_id": "line-1-e05", "apparent_resistivity": "90", "chargeability": "0.02"}),
	sourcetest.Row(map[string]string{"id": "frozen2", "tx1_id": "line-1-e05", "apparent_resistivity": "90", "chargeability": "0.02"}),
	sourcetest.Row(map[string]string{"id": "contact1", "tx1_id": "line-1-e10", "tx2_id": "line-1-e11", "contact_resistance": "20000"}),
	sourcetest.Row(map[string]string{"id": "contact2", "tx1_id": "line-1-e10", "tx2_id": "line-1-e12", "contact_resistance": "20000"}),
	sourcetest.Row(map[string]string{"id": "uses10", "tx1_id": "line-1-e13", "tx2_id": "line-1-e14", "rx1_id": "line-1-e10", "rx2_id": "line-1-e15"}),
)

func advisoryFixture(t *testing.T) (*testutil.App, map[string]string) {
	a := testutil.NewApp(t)
	a.Import("/data/line-1.csv", advisoryCSV)
	return a, a.IDsByAcquisition()
}

// unresolvedExplanations lists a detail's unresolved results' explanations: the
// evidence a saved reason must quote for that reading, whatever words surround it.
func unresolvedExplanations(d map[string]any) []string {
	var out []string
	for _, r := range d["advisory"].(map[string]any)["results"].([]any) {
		m := r.(map[string]any)
		if m["state"] != "met" {
			out = append(out, m["explanation"].(string))
		}
	}
	return out
}

// quotesOwnEvidence asserts a reason quotes every unresolved explanation of its reading.
func quotesOwnEvidence(t *testing.T, reason any, d map[string]any, msg string) {
	t.Helper()
	for _, e := range unresolvedExplanations(d) {
		assert.Contains(t, reason, e, msg)
	}
}

// resultStates writes a detail's results as "criterion:state" in order.
func resultStates(d map[string]any) string {
	var out []string
	for _, r := range d["advisory"].(map[string]any)["results"].([]any) {
		m := r.(map[string]any)
		out = append(out, fmt.Sprintf("%v:%v", m["criterion"], m["state"]))
	}
	return strings.Join(out, " ")
}

// The page shows the advisory with its results and the reason a save would
// record; a preview that differs from what the save then writes, or a history entry
// whose reason is not the one saved, misreports the decision.
func TestDetailCarriesTheAdvisoryAndTheReasonsASaveRecords(t *testing.T) {
	a, ids := advisoryFixture(t)
	d := detailOf(t, a, ids["neg"])
	assert.Equal(t, "fail", d["advisory"].(map[string]any)["value"])
	assert.Equal(t, "resistivity_sign:violated resistivity_relative_error:met chargeability_relative_error:met transmitter_contact:met decay_trend:met frozen_repeat:met repeat_agreement:review suspect_electrode:met", resultStates(d))
	assert.Nil(t, d["reason"], "no save yet")
	// The template's wording is pinned where it is built; here, that the preview of each
	// verdict quotes this reading's own evidence.
	preview := d["would_save_reasons"].(map[string]any)
	quotesOwnEvidence(t, preview["pass"], d, "the Pass preview")
	quotesOwnEvidence(t, preview["fail"], d, "the Fail preview")
	negFailReason := preview["fail"]
	var advisories []string
	for _, d := range allDetails(t, a) {
		advisories = append(advisories, d["advisory"].(map[string]any)["value"].(string))
	}
	assert.Equal(t, "fail,pass,pass,needs_review,needs_review,fail,fail,fail,fail,needs_review", strings.Join(advisories, ","))
	// far's repeats: neg is left out for its violation, missing for having no
	// resistivity, so the reference is the median of ok1 and ok2 alone.
	var leftOut []string
	for _, r := range detailOf(t, a, ids["far"])["repeats"].([]any) {
		m := r.(map[string]any)
		leftOut = append(leftOut, fmt.Sprintf("%v:%v", m["acquisition_id"], m["left_out"]))
	}
	assert.Equal(t, []string{"neg:true", "ok1:false", "ok2:false", "missing:true"}, leftOut)

	status, out := save(a, "fail", ids["neg"])
	require.Equal(t, http.StatusOK, status, out)
	d = detailOf(t, a, ids["neg"])
	assert.Equal(t, negFailReason, d["reason"], "the saved reason is the previewed one")
	_, entries := history(t, a, ids["neg"])
	assert.Equal(t, negFailReason, entries[0]["reason"])

	status, out = save(a, "unreviewed", ids["neg"])
	require.Equal(t, http.StatusOK, status, out)
	d = detailOf(t, a, ids["neg"])
	assert.Nil(t, d["reason"], "Unreviewed clears the reason")
	_, entries = history(t, a, ids["neg"])
	assert.Nil(t, entries[0]["reason"])
	assert.Equal(t, negFailReason, entries[1]["reason"], "the earlier entry keeps its reason")

	a.Reopen()
	_, entries = history(t, a, ids["neg"])
	assert.Equal(t, negFailReason, entries[1]["reason"], "after reopening the database")
}

// Each reading's saved reason is built from its own evidence by the server, a repeated
// verdict records nothing, and a caller can write the reason neither at the top level
// nor on a target.
func TestSaveBuildsEachReadingsOwnReason(t *testing.T) {
	a, ids := advisoryFixture(t)
	for _, name := range []string{"ok1", "far"} {
		before := detailOf(t, a, ids[name])
		status, out := save(a, "pass", ids[name])
		require.Equal(t, http.StatusOK, status, out)
		saved := detailOf(t, a, ids[name])["reason"]
		assert.Equal(t, before["would_save_reasons"].(map[string]any)["pass"], saved, "%s: the saved reason is its preview", name)
		quotesOwnEvidence(t, saved, before, name)
	}
	assert.NotEqual(t, detailOf(t, a, ids["ok1"])["reason"], detailOf(t, a, ids["far"])["reason"], "two readings, two reasons")

	status, out := save(a, "pass", ids["far"])
	require.Equal(t, http.StatusOK, status, out)
	assert.Nil(t, out["action_id"], "repeating the verdict is unchanged")
	got, _ := history(t, a, ids["far"])
	assert.Len(t, got, 1)

	for _, body := range []string{
		fmt.Sprintf(`{"targets":[{"reading_id":%q}],"verdict":"fail","reason":"mine"}`, ids["ok2"]),
		fmt.Sprintf(`{"targets":[{"reading_id":%q,"reason":"mine"}],"verdict":"fail"}`, ids["ok2"]),
	} {
		status, out = a.Do(http.MethodPost, "/api/reviews", body)
		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, out["message"], "reason is not a known field.", body)
	}
	assert.Nil(t, detailOf(t, a, ids["ok2"])["reason"])
	assert.Equal(t, "unreviewed", detailOf(t, a, ids["ok2"])["verdict"])
}
