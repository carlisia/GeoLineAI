package review

import (
	"fmt"
	"strings"

	"geolineai/internal/core/closed"
	"geolineai/internal/core/review/criterion"
)

// reasonResults is how many unresolved results a reason spells out; the rest are
// counted, since every result stays recorded with the advisory.
const reasonResults = 3

// RecordsReason reports whether a save of the verdict records a reason on each reading:
// Pass and Fail do, Unreviewed clears it. The frontend's copy is generated from it.
func RecordsReason(v Verdict) bool {
	return v != Unreviewed
}

// Reason is the text a save records with a Pass or Fail on one reading, built from
// its evaluation: the advisory, up to reasonResults unresolved results (Violated
// first, then Review, each in criterion table order) with a count of the rest, and
// the verdict chosen, saying when it differs from the advisory. The same text is
// previewed before a save and recorded by it, so the two cannot differ.
func Reason(e criterion.Evaluation, chosen Verdict) string {
	parts := []string{advisoryClause(e.Value)}
	var unresolved []criterion.Result
	for _, state := range criterion.UnresolvedStates {
		for _, r := range e.Results {
			if r.State == state {
				unresolved = append(unresolved, r)
			}
		}
	}
	for i, r := range unresolved {
		if i == reasonResults {
			break
		}
		parts = append(parts, r.Explanation)
	}
	if more := len(unresolved) - reasonResults; more == 1 {
		parts = append(parts, "1 more unresolved result is recorded with the advisory.")
	} else if more > 1 {
		parts = append(parts, fmt.Sprintf("%d more unresolved results are recorded with the advisory.", more))
	}
	if !RecordsReason(chosen) {
		panic("review: a reason is built only for a verdict that records one, not " + string(chosen))
	}
	parts = append(parts, fmt.Sprintf("Reviewer selected %s.", closed.Entry(VerdictLabels, chosen, "verdict")))
	if contradicts[e.Value] == chosen {
		parts = append(parts, "This differs from the advisory.")
	}
	return strings.Join(parts, " ")
}

// contradicts is the verdict that contradicts each advisory: Fail against Pass and
// Pass against Fail. No verdict contradicts Needs review, so it has no entry.
var contradicts = map[criterion.Advisory]Verdict{criterion.Pass: Fail, criterion.Fail: Pass}

// advisoryClause names the advisory by its display label and says what it means.
func advisoryClause(a criterion.Advisory) string {
	return fmt.Sprintf("Advisory: %s, %s.", closed.Entry(criterion.AdvisoryLabels, a, "advisory"), closed.Entry(criterion.AdvisoryMeanings, a, "advisory"))
}
