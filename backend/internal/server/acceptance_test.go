package server_test

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/scenarios"
	"geolineai/internal/testutil"
)

// repoPath is a path from the repository root, as seen from this package.
func repoPath(parts ...string) string {
	return filepath.Join(append([]string{"..", "..", ".."}, parts...)...)
}

// issueLabelOf is the label an import issue result's explanation ends with, as the
// page shows it: the wire code is the test's input, the label the reviewer's text.
func issueLabelOf(code string) string {
	switch code {
	case "missing":
		return "Missing value."
	case "id_collision":
		return "Repeated acquisition ID."
	}
	panic(code)
}

// shippedLine imports the shipped line and the shipped scenario manifest, and reads
// every reading's detail in file order.
func shippedLine(t *testing.T) (*testutil.App, []map[string]any) {
	t.Helper()
	a := testutil.NewApp(t)
	csv, err := os.ReadFile(repoPath("data", "line-1.csv"))
	require.NoError(t, err)
	a.Import("/data/line-1.csv", string(csv))
	a.Manifest, err = scenarios.Load(repoPath("data", "scenarios.json"))
	require.NoError(t, err)
	a.Reopen()
	return a, allDetails(t, a)
}

// results returns a detail's results whose state is not Met, as "criterion:state".
func unresolved(d map[string]any) []string {
	var out []string
	for _, r := range d["advisory"].(map[string]any)["results"].([]any) {
		m := r.(map[string]any)
		if m["state"] != "met" {
			out = append(out, m["criterion"].(string)+":"+m["state"].(string))
		}
	}
	return out
}

// The shipped line, imported once for both checks: reading every detail is the slow
// part, and neither check writes.
func TestShippedLine(t *testing.T) {
	a, details := shippedLine(t)
	t.Run("counts", func(t *testing.T) { shippedLineCounts(t, details) })
	t.Run("scenarios", func(t *testing.T) { shippedScenarios(t, a, details) })
}

// The crafted fixtures pin each rule on its own; only the shipped line pins them
// together. These counts were computed independently of this code, from the CSV, by
// the rules the criteria document; a change to a rule, a limit, the grouping key or
// the median that keeps every fixture green still moves them.
func shippedLineCounts(t *testing.T, details []map[string]any) {
	require.Len(t, details, 1603, "every row retained")

	advisories := map[string]int{}
	readingsWith := map[string]int{}
	groups := map[any]bool{}
	disagreeingGroups := map[any]bool{}
	disagreeing := 0
	issueReadings, issueResults := map[string]int{}, map[string]int{}
	named := regexp.MustCompile(`^Electrode (\S+) \(`)
	suspect := map[string]bool{}
	for _, d := range details {
		advisories[d["advisory"].(map[string]any)["value"].(string)]++
		seen := map[string]bool{}
		for _, r := range d["advisory"].(map[string]any)["results"].([]any) {
			m := r.(map[string]any)
			c := m["criterion"].(string)
			if m["state"] == "met" {
				continue
			}
			if !seen[c] {
				readingsWith[c]++
				seen[c] = true
			}
			if c == "suspect_electrode" {
				suspect[named.FindStringSubmatch(m["explanation"].(string))[1]] = true
			}
			if c == "import_issue" {
				for _, code := range []string{"missing", "id_collision"} {
					if strings.Contains(m["explanation"].(string), issueLabelOf(code)) {
						issueResults[code]++
					}
				}
			}
		}
		for _, code := range []string{"missing", "id_collision"} {
			for _, i := range d["import_issues"].([]any) {
				if i.(map[string]any)["code"] == code {
					issueReadings[code]++
					break
				}
			}
		}
		if g := d["repeat_group_id"]; g != nil {
			groups[g] = true
			if d["disagrees"] == true {
				disagreeing++
				disagreeingGroups[g] = true
			}
		}
	}
	assert.Equal(t, map[string]int{"pass": 1326, "needs_review": 163, "fail": 114}, advisories, "advisories")
	assert.Equal(t, map[string]int{
		"import_issue": 14, "resistivity_sign": 23, "resistivity_relative_error": 34, "chargeability_relative_error": 52,
		"transmitter_contact": 19, "decay_trend": 46, "frozen_repeat": 5, "repeat_agreement": 24, "suspect_electrode": 194,
	}, readingsWith, "readings with an unresolved result of each criterion")
	assert.Equal(t, map[string]bool{"line-1-e31": true}, suspect, "the one suspect electrode")
	assert.Equal(t, 268, len(groups), "repeat groups among readings with all four slots")
	assert.Equal(t, 24, disagreeing, "disagreeing readings")
	assert.Equal(t, 7, len(disagreeingGroups), "groups with a disagreeing reading")
	assert.Equal(t, map[string]int{"missing": 6, "id_collision": 8}, issueReadings, "readings with each import issue")
	assert.Equal(t, map[string]int{"missing": 18, "id_collision": 8}, issueResults, "import-issue results")
}

// The three scenarios satisfy their selection rules on the shipped line's computed
// results, each being the first row that does, and resolve to three readings.
func shippedScenarios(t *testing.T, a *testutil.App, details []map[string]any) {
	status, out := a.Do(http.MethodGet, "/api/scenarios", "")
	require.Equal(t, http.StatusOK, status, out)
	list := out["scenarios"].([]any)
	require.Len(t, list, 3)

	advisory := func(d map[string]any) string { return d["advisory"].(map[string]any)["value"].(string) }
	violatedMeasurement := func(d map[string]any) bool {
		for _, u := range unresolved(d) {
			if strings.HasSuffix(u, ":violated") && !strings.HasPrefix(u, "import_issue") {
				return true
			}
		}
		return false
	}
	// The repeat comparison is the reason the reading needs review: its only
	// unresolved result.
	disagreesAlone := func(d map[string]any) bool {
		u := unresolved(d)
		return len(u) == 1 && u[0] == "repeat_agreement:review"
	}
	first := func(rule func(map[string]any) bool, skip ...any) any {
		for _, d := range details {
			if rule(d) && !slices.Contains(skip, d["reading_id"]) {
				return d["reading_id"]
			}
		}
		return nil
	}
	pass := first(func(d map[string]any) bool { return advisory(d) == "pass" })
	violated := first(violatedMeasurement, pass)
	repeats := first(disagreesAlone, pass, violated)
	assert.Equal(t, []any{
		map[string]any{"label": "Passing reading", "reading_id": pass},
		map[string]any{"label": "Violated criterion", "reading_id": violated},
		map[string]any{"label": "Disagreeing repeats", "reading_id": repeats},
	}, list)
	for _, d := range details {
		if d["reading_id"] == repeats {
			assert.Equal(t, "needs_review", advisory(d), "the third scenario's advisory")
		}
		if d["reading_id"] == violated {
			assert.Equal(t, "fail", advisory(d), "the second scenario's advisory")
		}
	}
}
