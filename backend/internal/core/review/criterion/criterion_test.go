package criterion_test

import (
	"fmt"
	"strings"
	"testing"

	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/intake/sourcetest"
	"geolineai/internal/core/review/criterion"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(t *testing.T, rows ...string) *source.File {
	t.Helper()
	f, err := source.Parse(strings.NewReader(sourcetest.CSV(rows...)))
	require.NoError(t, err)
	require.Empty(t, f.Rejections, "the fixture rows must all be retained")
	return &f
}

// states writes results as "criterion:state" in order.
func states(results []criterion.Result) string {
	var out []string
	for _, r := range results {
		out = append(out, fmt.Sprintf("%s:%s", r.Criterion, r.State))
	}
	return strings.Join(out, " ")
}

func find(results []criterion.Result, name criterion.Criterion) (criterion.Result, bool) {
	for _, r := range results {
		if r.Criterion == name {
			return r, true
		}
	}
	return criterion.Result{}, false
}

// Every limit is a boundary the reviewer cannot see: a value exactly at it, just over
// it, at zero, or absent must land on the right side, or readings are flagged or
// cleared for the wrong reason.
func TestSingleCriteriaAtTheirBoundaries(t *testing.T) {
	for _, c := range []struct {
		name      string
		set       map[string]string
		criterion criterion.Criterion
		want      string // "" when no result is recorded
		value     *float64
		anyValue  bool // the value is a computed slope, not checked here
	}{
		{"resistivity at 0 is violated", map[string]string{"apparent_resistivity": "0", "apparent_resistivity_err": "0"}, criterion.ResistivitySign, "violated", f(0), false},
		{"resistivity just above 0 is met", map[string]string{"apparent_resistivity": "0.000001"}, criterion.ResistivitySign, "met", f(0.000001), false},
		{"negative resistivity is violated", map[string]string{"apparent_resistivity": "-5"}, criterion.ResistivitySign, "violated", f(-5), false},
		{"resistivity error at the limit is met", map[string]string{"apparent_resistivity": "100", "apparent_resistivity_err": "50"}, criterion.ResistivityRelativeError, "met", f(0.5), false},
		{"resistivity error just over the limit is violated", map[string]string{"apparent_resistivity": "100", "apparent_resistivity_err": "50.0001"}, criterion.ResistivityRelativeError, "violated", f(0.500001), false},
		{"resistivity error divides by the magnitude of a negative value", map[string]string{"apparent_resistivity": "-100", "apparent_resistivity_err": "50"}, criterion.ResistivityRelativeError, "met", f(0.5), false},
		{"a zero resistivity has no relative error", map[string]string{"apparent_resistivity": "0", "apparent_resistivity_err": "1"}, criterion.ResistivityRelativeError, "", nil, false},
		{"a zero error is met at 0", map[string]string{"apparent_resistivity": "100", "apparent_resistivity_err": "0"}, criterion.ResistivityRelativeError, "met", f(0), false},
		{"chargeability error at the limit is met", map[string]string{"chargeability": "0.01", "chargeability_err": "0.01"}, criterion.ChargeabilityRelativeError, "met", f(1), false},
		{"chargeability error just over the limit is violated", map[string]string{"chargeability": "0.01", "chargeability_err": "0.0100001"}, criterion.ChargeabilityRelativeError, "violated", f(1.00001), false},
		{"a zero chargeability has no relative error", map[string]string{"chargeability": "0", "chargeability_err": "0.001"}, criterion.ChargeabilityRelativeError, "", nil, false},
		{"a ratio too large to represent is violated with no value", map[string]string{"chargeability": "1e-300", "chargeability_err": "1e300"}, criterion.ChargeabilityRelativeError, "violated", nil, false},
		{"negative contact resistance is violated", map[string]string{"contact_resistance": "-0.0001"}, criterion.TransmitterContact, "violated", f(-0.0001), false},
		{"zero contact resistance is met", map[string]string{"contact_resistance": "0"}, criterion.TransmitterContact, "met", f(0), false},
		{"contact resistance at the limit is met", map[string]string{"contact_resistance": "10000"}, criterion.TransmitterContact, "met", f(10000), false},
		{"contact resistance just over the limit is violated", map[string]string{"contact_resistance": "10000.0001"}, criterion.TransmitterContact, "violated", f(10000.0001), false},
		{"a falling decay is met", map[string]string{"decay_curve": "10;9;8;7;6;5;4;3;2;1"}, criterion.DecayTrend, "met", nil, true},
		{"a rising decay is violated", map[string]string{"decay_curve": "1;2;3;4;5;6;7;8;9;10"}, criterion.DecayTrend, "violated", nil, true},
		{"a flat decay has slope 0 and is violated", map[string]string{"decay_curve": "5;5;5;5;5;5;5;5;5;5"}, criterion.DecayTrend, "violated", f(0), false},
		{"a decay that dips but falls overall is met", map[string]string{"decay_curve": "10;11;8;9;6;7;4;5;2;1"}, criterion.DecayTrend, "met", nil, true},
		{"an absent resistivity gives no sign result", map[string]string{"apparent_resistivity": ""}, criterion.ResistivitySign, "", nil, false},
		{"an absent resistivity gives no relative error", map[string]string{"apparent_resistivity": ""}, criterion.ResistivityRelativeError, "", nil, false},
		{"an absent contact resistance gives no result", map[string]string{"contact_resistance": ""}, criterion.TransmitterContact, "", nil, false},
		{"an absent decay gives no result", map[string]string{"decay_curve": ""}, criterion.DecayTrend, "", nil, false},
		{"an absent reported error gives no relative error, and the sign still judges", map[string]string{"apparent_resistivity_err": ""}, criterion.ResistivityRelativeError, "", nil, false},
		{"an absent reported error leaves the sign result", map[string]string{"apparent_resistivity_err": ""}, criterion.ResistivitySign, "met", f(120), false},
	} {
		results := criterion.Evaluate(parse(t, sourcetest.Row(c.set)))[0].Results
		got, ok := find(results, c.criterion)
		if c.want == "" {
			assert.False(t, ok, "%s: no result wanted, got %+v", c.name, got)
			continue
		}
		if !assert.True(t, ok, "%s: no %s result", c.name, c.criterion) {
			continue
		}
		assert.Equal(t, c.want, string(got.State), c.name)
		if c.anyValue {
			assert.NotNil(t, got.Value, c.name)
		} else if c.value == nil {
			assert.Nil(t, got.Value, c.name)
		} else if assert.NotNil(t, got.Value, c.name) {
			assert.InDelta(t, *c.value, *got.Value, 1e-9, c.name)
		}
	}
}

// A missing value is judged by the reviewer, not failed: its import issue is the
// only result it leaves, and the results keep table order.
func TestImportIssuesNeedReviewAndResultsKeepTableOrder(t *testing.T) {
	f := parse(t, sourcetest.Row(map[string]string{"apparent_resistivity": " ", "contact_resistance": "20000"}))
	results := criterion.Evaluate(f)[0].Results
	assert.Equal(t, "import_issue:review chargeability_relative_error:met transmitter_contact:violated decay_trend:met suspect_electrode:met", states(results))
	assert.Equal(t, "Import issue on apparent_resistivity: Missing value.", results[0].Explanation)
	assert.Equal(t, "Transmitter contact resistance 20000 ohm is above 10000 ohm.", results[2].Explanation)
}

// An explanation that prints a value as equal to the limit it says the value is
// above contradicts itself, and the text is saved into every reason; a negative
// contact resistance was judged against the floor, so that is the limit it records.
func TestExplanationsReadApartFromTheirLimits(t *testing.T) {
	f := parse(t,
		sourcetest.Row(map[string]string{"id": "a", "apparent_resistivity": "256.035"}),
		sourcetest.Row(map[string]string{"id": "b", "apparent_resistivity": "256.035"}),
		sourcetest.Row(map[string]string{"id": "c", "apparent_resistivity": "204.773"}),
		sourcetest.Row(map[string]string{"id": "d", "tx1_id": "line-1-e05", "apparent_resistivity": "100", "apparent_resistivity_err": "50.00001"}),
		sourcetest.Row(map[string]string{"id": "e", "tx1_id": "line-1-e06", "contact_resistance": "-1"}),
		sourcetest.Row(map[string]string{"id": "f", "tx1_id": "line-1-e07", "contact_resistance": "10000.00000001"}),
	)
	evaluations := criterion.Evaluate(f)
	agreement, _ := find(evaluations[2].Results, criterion.RepeatAgreement)
	assert.Equal(t, criterion.Review, agreement.State)
	assert.Equal(t, "Resistivity 204.773 ohm-m differs from the reference 256.035 ohm-m, the median of 2 repeats, by 20.02%, above the 20.0% limit.", agreement.Explanation)
	err, _ := find(evaluations[3].Results, criterion.ResistivityRelativeError)
	assert.Equal(t, criterion.Violated, err.State)
	assert.Equal(t, "Resistivity error 50.00001 ohm-m is 0.5000001 of the value 100 ohm-m, above the 0.5 allowed.", err.Explanation)
	contact, _ := find(evaluations[4].Results, criterion.TransmitterContact)
	assert.Equal(t, "Transmitter contact resistance -1 ohm is below 0 ohm.", contact.Explanation)
	assert.Equal(t, 0.0, *contact.Limit, "the floor it was judged against")
	over, _ := find(evaluations[5].Results, criterion.TransmitterContact)
	assert.Equal(t, "Transmitter contact resistance 10000.00000001 ohm is above 10000 ohm.", over.Explanation)
}

// Frozen means every value equal: a reading that matches a repeat to every digit is
// copied, one that differs in a single sample is not, and a reading missing a value
// is neither compared nor comparable.
func TestFrozenRepeatIsAnExactMatchOfEveryValue(t *testing.T) {
	f := parse(t,
		sourcetest.Row(map[string]string{"id": "a"}),
		sourcetest.Row(map[string]string{"id": "b"}),
		sourcetest.Row(map[string]string{"id": "c", "decay_curve": "10;9;8;7;6;5;4;3;2;1.0001"}),
		sourcetest.Row(map[string]string{"id": "d", "chargeability": ""}),
	)
	evaluations := criterion.Evaluate(f)
	byID := map[string]criterion.Evaluation{}
	for i, r := range f.Readings {
		byID[r.AcquisitionID] = evaluations[i]
	}
	a, _ := find(byID["a"].Results, criterion.FrozenRepeat)
	assert.Equal(t, criterion.Violated, a.State)
	assert.Equal(t, "Resistivity, chargeability and all ten decay samples equal those of repeat b.", a.Explanation)
	assert.Equal(t, criterion.Fail, byID["a"].Value, "a frozen repeat fails the advisory")
	c, _ := find(byID["c"].Results, criterion.FrozenRepeat)
	assert.Equal(t, criterion.Met, c.State)
	assert.Equal(t, "No compared repeat has the same resistivity, chargeability and decay samples (2 compared).", c.Explanation, "d, missing a value, is not compared")
	_, ok := find(byID["d"].Results, criterion.FrozenRepeat)
	assert.False(t, ok, "a reading missing a value has no frozen result")
}

// A repeat already known to be bad must not drag the reference toward itself: with
// the negative repeat counted the median would be 100 and d would disagree; left
// out, the reference is 150 and d agrees.
func TestRepeatAgreementLeavesOutViolatedRepeats(t *testing.T) {
	f := parse(t,
		sourcetest.Row(map[string]string{"id": "a", "apparent_resistivity": "100"}),
		sourcetest.Row(map[string]string{"id": "b", "apparent_resistivity": "200"}),
		sourcetest.Row(map[string]string{"id": "c", "apparent_resistivity": "-300"}),
		sourcetest.Row(map[string]string{"id": "d", "apparent_resistivity": "150"}),
	)
	evaluations := criterion.Evaluate(f)
	d, ok := find(evaluations[3].Results, criterion.RepeatAgreement)
	require.True(t, ok)
	assert.Equal(t, criterion.Met, d.State)
	assert.Equal(t, "Resistivity 150 ohm-m differs from the reference 150 ohm-m, the median of 2 repeats, by 0.0%, within the 20.0% limit.", d.Explanation)
	c, ok := find(evaluations[2].Results, criterion.RepeatAgreement)
	require.True(t, ok, "the bad reading is still compared with its good repeats")
	assert.Equal(t, criterion.Review, c.State)
	assert.InDelta(t, 3, *c.Value, 1e-12, "-300 against the median 150 of 100, 150 and 200")
	assert.Equal(t, criterion.Fail, evaluations[2].Value, "the sign violation decides c's advisory")

	// Two agreeing readings and one bad one: with the bad one left out, one repeat is
	// too few for a reference.
	g := parse(t,
		sourcetest.Row(map[string]string{"id": "a", "apparent_resistivity": "100"}),
		sourcetest.Row(map[string]string{"id": "b", "apparent_resistivity": "100"}),
		sourcetest.Row(map[string]string{"id": "c", "apparent_resistivity": "-300"}),
	)
	_, ok = find(criterion.Evaluate(g)[0].Results, criterion.RepeatAgreement)
	assert.False(t, ok, "one counting repeat gives no reference")
}

// Repeats near the float64 maximum are valid, finite input, but arithmetic on them
// can overflow: averaging the middle two once became +Inf, the difference Inf/Inf was
// can overflow: averaging the middle two once became +Inf, the difference Inf/Inf was
// NaN, and NaN is above no limit, so the reading was recorded as agreeing. A
// difference too large to represent must still disagree, and store no value.
func TestRepeatAgreementDisagreesWhenItsArithmeticWouldOverflow(t *testing.T) {
	cases := []struct {
		name, own   string
		value       *float64
		explanation string
	}{
		{"the median of the repeats", "200", new(1.0),
			"Resistivity 200 ohm-m differs from the reference 1e+308 ohm-m, the median of 2 repeats, by 100.0%, above the 20.0% limit."},
		{"the difference from the median", "-1e308", nil,
			"Resistivity -1e+308 ohm-m differs from the reference 1e+308 ohm-m, the median of 2 repeats, by more than can be represented, above the 20.0% limit."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := parse(t,
				sourcetest.Row(map[string]string{"id": "a", "apparent_resistivity": c.own}),
				sourcetest.Row(map[string]string{"id": "b", "apparent_resistivity": "1e308"}),
				sourcetest.Row(map[string]string{"id": "c", "apparent_resistivity": "1e308"}),
			)
			a, ok := find(criterion.Evaluate(f)[0].Results, criterion.RepeatAgreement)
			require.True(t, ok)
			assert.Equal(t, criterion.Review, a.State)
			if c.value == nil {
				assert.Nil(t, a.Value, "no value is stored for a difference that cannot be represented")
			} else {
				require.NotNil(t, a.Value)
				assert.InDelta(t, *c.value, *a.Value, 1e-12)
			}
			assert.Equal(t, c.explanation, a.Explanation)
		})
	}
}

// A faulty electrode fails contact with every partner it transmits with; its
// neighbours fail only with it. Counting distinct failing partners tells them apart,
// and every reading using the suspect electrode, in any slot, needs review.
func TestSuspectElectrodeNeedsTwoFailingPartners(t *testing.T) {
	f := parse(t,
		sourcestestRow("r1", "line-1-e01", "line-1-e02", "line-1-e03", "line-1-e04", "20000"),
		sourcestestRow("r2", "line-1-e01", "line-1-e05", "line-1-e06", "line-1-e07", "20000"),
		sourcestestRow("r3", "line-1-e08", "line-1-e09", "line-1-e01", "line-1-e10", "800"),
		sourcestestRow("r4", "line-1-e02", "line-1-e05", "line-1-e06", "line-1-e07", "800"),
		sourcestestRow("r5", "", "", "", "", "800"),
	)
	evaluations := criterion.Evaluate(f)
	get := func(i int) (criterion.Result, bool) { return find(evaluations[i].Results, criterion.SuspectElectrode) }
	r1, _ := get(0)
	assert.Equal(t, "Electrode line-1-e01 (tx1) transmitted with failing contact to 2 different partners.", r1.Explanation)
	assert.Equal(t, criterion.Review, r1.State)
	assert.Equal(t, criterion.Fail, evaluations[0].Value, "its own failing contact decides r1")
	r3, _ := get(2)
	assert.Equal(t, criterion.Review, r3.State, "the suspect electrode in a receiver slot")
	assert.Equal(t, "Electrode line-1-e01 (rx1) transmitted with failing contact to 2 different partners.", r3.Explanation)
	assert.Equal(t, criterion.NeedsReview, evaluations[2].Value)
	r4, _ := get(3)
	assert.Equal(t, criterion.Met, r4.State, "e02 and e05 each failed with one partner only")
	assert.Equal(t, "No electrode of this reading is suspect.", r4.Explanation)
	assert.Equal(t, criterion.Pass, evaluations[3].Value, "r4's import issues aside, nothing is unresolved")
	_, ok := get(4)
	assert.False(t, ok, "a reading with no electrode ID has no suspect result")
}

func sourcestestRow(id, tx1, tx2, rx1, rx2, contact string) string {
	return sourcetest.Row(map[string]string{"id": id, "tx1_id": tx1, "tx2_id": tx2, "rx1_id": rx1, "rx2_id": rx2, "contact_resistance": contact})
}

func f(v float64) *float64 { return &v }
