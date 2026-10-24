package server_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/testutil"
)

// recordedAt is a server-generated time: UTC, RFC 3339 with three fractional digits.
var recordedAt = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

// save posts one verdict for one reading and returns the status and body.
func save(a *testutil.App, verdict, id string) (int, map[string]any) {
	return a.Do(http.MethodPost, "/api/reviews", fmt.Sprintf(`{"targets":[{"reading_id":%q}],"verdict":%q}`, id, verdict))
}

// history reads a reading's history entries, newest first, as
// "previous->new reviewer entries".
func history(t *testing.T, a *testutil.App, id string) ([]string, []map[string]any) {
	t.Helper()
	d := detailOf(t, a, id)
	var out []string
	var entries []map[string]any
	for _, e := range d["history"].([]any) {
		m := e.(map[string]any)
		out = append(out, fmt.Sprintf("%v->%v %v %v", m["previous_verdict"], m["new_verdict"], m["reviewer"], m["entry_count"]))
		entries = append(entries, m)
	}
	return out, entries
}

// A history entry that records a verdict the reading never had, or derives the wrong
// new verdict, tells the reviewer a decision evolved in a way it did not; a save that
// changed nothing must leave no trace, or history would record a change that did not
// happen; and a save touches only its own reading.
func TestSaveRecordsOneActionWithOneEntry(t *testing.T) {
	a := groupFixture(t)
	ids := a.IDsByAcquisition()

	status, out := save(a, "fail", ids["a"])
	require.Equal(t, http.StatusOK, status, out)
	first := out["action_id"]
	require.NotNil(t, first, "a save that changed a reading is an action")

	status, out = save(a, "fail", ids["c"])
	require.Equal(t, http.StatusOK, status, out)
	assert.NotEqual(t, first, out["action_id"])

	status, out = save(a, "fail", ids["a"])
	require.Equal(t, http.StatusOK, status, out)
	assert.Nil(t, out["action_id"], "a save that changed nothing is no action")
	assert.Equal(t, 2, a.Count("review_actions"))
	assert.Equal(t, 2, a.Count("review_history"))

	got, entries := history(t, a, ids["a"])
	assert.Equal(t, []string{"unreviewed->fail reviewer 1"}, got, "reading a: changed once")
	assert.Equal(t, first, entries[0]["action_id"])
	assert.Regexp(t, recordedAt, entries[0]["recorded_at"])
	got, _ = history(t, a, ids["c"])
	assert.Equal(t, []string{"unreviewed->fail reviewer 1"}, got, "reading c: its own save")
	got, _ = history(t, a, ids["b"])
	assert.Empty(t, got, "a reading in the same group, untouched, has no history")
	assert.Equal(t, "unreviewed", detailOf(t, a, ids["b"])["verdict"])

	a.Reopen()
	got, _ = history(t, a, ids["a"])
	assert.Equal(t, []string{"unreviewed->fail reviewer 1"}, got, "after reopening the database")
}

// Each save on one reading must record the verdict it replaced, appended in order, or
// the derived new verdicts would skip a step the reviewer took; no later save
// rewrites an earlier entry.
func TestEachSaveAppendsTheVerdictItReplaced(t *testing.T) {
	a := groupFixture(t)
	ids := a.IDsByAcquisition()
	var actions []any
	for _, v := range []string{"fail", "pass", "unreviewed", "fail"} {
		status, out := save(a, v, ids["a"])
		require.Equal(t, http.StatusOK, status, out)
		actions = append(actions, out["action_id"])
	}
	got, entries := history(t, a, ids["a"])
	assert.Equal(t, []string{
		"unreviewed->fail reviewer 1",
		"pass->unreviewed reviewer 1",
		"fail->pass reviewer 1",
		"unreviewed->fail reviewer 1",
	}, got, "newest first, each new verdict the next entry's previous, the latest the current verdict")
	for i, e := range entries {
		assert.Equal(t, actions[len(actions)-1-i], e["action_id"], "entry %d belongs to its own save", i)
	}
}

// A save commits whole or not at all: a failure at its last write, after the action and
// its history entry are written, must leave no action, no entry and the old verdict,
// or the record would show a decision the reading does not hold. The database itself
// refuses the verdict's update, so the save runs as in production up to that point.
func TestSaveThatFailsPartwayWritesNothing(t *testing.T) {
	a := groupFixture(t)
	id := a.IDsByAcquisition()["a"]
	a.Exec(`CREATE TRIGGER disk_full BEFORE UPDATE OF qc_review ON readings BEGIN SELECT RAISE(ABORT, 'disk full'); END`)
	status, out := save(a, "pass", id)
	require.Equal(t, http.StatusInternalServerError, status, out)
	assert.Equal(t, "The server could not complete the request.", out["message"], "the cause is logged, not sent")
	assert.Zero(t, a.Count("review_actions"), "no action")
	assert.Zero(t, a.Count("review_history"), "no history entry")
	assert.Equal(t, "unreviewed", detailOf(t, a, id)["verdict"], "the verdict unchanged")

	a.Exec(`DROP TRIGGER disk_full`)
	status, out = save(a, "pass", id)
	require.Equal(t, http.StatusOK, status, out, "the same save succeeds once the database does")
	assert.Equal(t, 1, a.Count("review_actions"))
	assert.Equal(t, "pass", detailOf(t, a, id)["verdict"])
}
