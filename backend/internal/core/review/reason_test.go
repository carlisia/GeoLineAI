package review_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
)

func result(state criterion.ResultState, explanation string) criterion.Result {
	return criterion.Result{State: state, Explanation: explanation}
}

// A reason that drops a violation behind a review, miscounts what it left out, or
// stays silent when the reviewer overrides the advisory records a decision wrongly,
// and the record is all a later reader has.
func TestReasonOrdersCapsAndNamesTheOverride(t *testing.T) {
	for _, c := range []struct {
		name   string
		e      criterion.Evaluation
		chosen review.Verdict
		want   string
	}{
		{
			"violations before reviews, three shown, the rest counted, and the override named",
			criterion.Evaluation{Value: criterion.Fail, Results: []criterion.Result{
				result(criterion.Review, "R1."), result(criterion.Violated, "V1."), result(criterion.Met, "M."), result(criterion.Review, "R2."), result(criterion.Violated, "V2."),
			}},
			review.Pass,
			"Advisory: Fail, a criterion was violated. V1. V2. R1. 1 more unresolved result is recorded with the advisory. Reviewer selected Pass. This differs from the advisory.",
		},
		{
			"more than one left out is plural, and agreeing with the advisory says so by silence",
			criterion.Evaluation{Value: criterion.Fail, Results: []criterion.Result{
				result(criterion.Violated, "V1."), result(criterion.Violated, "V2."), result(criterion.Violated, "V3."), result(criterion.Review, "R1."), result(criterion.Review, "R2."),
			}},
			review.Fail,
			"Advisory: Fail, a criterion was violated. V1. V2. V3. 2 more unresolved results are recorded with the advisory. Reviewer selected Fail.",
		},
		{
			"no verdict differs from needs review",
			criterion.Evaluation{Value: criterion.NeedsReview, Results: []criterion.Result{result(criterion.Review, "R1.")}},
			review.Fail,
			"Advisory: Needs review, no criterion was violated but some need review. R1. Reviewer selected Fail.",
		},
		{
			"a pass with nothing unresolved",
			criterion.Evaluation{Value: criterion.Pass, Results: []criterion.Result{result(criterion.Met, "M.")}},
			review.Pass,
			"Advisory: Pass, every criterion met. Reviewer selected Pass.",
		},
		{
			"failing a passing advisory differs",
			criterion.Evaluation{Value: criterion.Pass},
			review.Fail,
			"Advisory: Pass, every criterion met. Reviewer selected Fail. This differs from the advisory.",
		},
	} {
		assert.Equal(t, c.want, review.Reason(c.e, c.chosen), c.name)
	}
	assert.Panics(t, func() { review.Reason(criterion.Evaluation{Value: criterion.Pass}, review.Unreviewed) }, "Unreviewed saves no reason")
}
