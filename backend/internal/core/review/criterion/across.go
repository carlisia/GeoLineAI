package criterion

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/repeats"
	"geolineai/internal/core/survey"
)

// Evaluate evaluates every criterion for every reading of a parsed file, the
// single-reading criteria first and then those across readings, which look at the
// file's repeat groups and electrodes; a line's repeats and electrodes all arrive in
// its one file. The evaluations are in the file's reading order.
func Evaluate(f *source.File) []Evaluation {
	singles := make([][]Result, len(f.Readings))
	leftOut := make([]bool, len(f.Readings))
	for i := range f.Readings {
		singles[i] = single(&f.Readings[i])
		// A repeat already known to be bad must not drag the reference toward itself.
		leftOut[i] = f.Readings[i].ApparentResistivity == nil || violated(singles[i])
	}
	groups := groupMembers(f)
	suspect := suspectElectrodes(f, singles)
	out := make([]Evaluation, len(f.Readings))
	for i := range f.Readings {
		r := &f.Readings[i]
		results := singles[i]
		var members []int
		if r.RepeatGroup != nil {
			members = groups[*r.RepeatGroup]
		}
		if res, ok := frozenRepeat(f, i, members); ok {
			results = append(results, res)
		}
		if res, ok := repeatAgreement(f, leftOut, i, members); ok {
			results = append(results, res)
		}
		if res, ok := suspectElectrode(r, suspect); ok {
			results = append(results, res)
		}
		out[i] = Evaluation{Value: decide(results), Results: results, LeftOut: leftOut[i]}
	}
	return out
}

// violated reports whether a result of the given criteria, or of any criterion when
// none is given, is violated.
func violated(results []Result, criteria ...Criterion) bool {
	for _, res := range results {
		if res.State == Violated && (len(criteria) == 0 || slices.Contains(criteria, res.Criterion)) {
			return true
		}
	}
	return false
}

// groupMembers lists each repeat group's readings, as indexes into the file, in
// reading order: acquisition time with none last, then row. Timestamps compare as
// text, which is time order only because the import stores them in UTC with a fixed
// width; that must hold for this order to be reading order. Storage orders a group
// the same way with each reading's ID in place of its row, which agrees only because
// an import stores a file's readings in row order; that must hold too.
func groupMembers(f *source.File) map[int][]int {
	out := map[int][]int{}
	for i, r := range f.Readings {
		if r.RepeatGroup != nil {
			out[*r.RepeatGroup] = append(out[*r.RepeatGroup], i)
		}
	}
	for _, members := range out {
		slices.SortStableFunc(members, func(a, b int) int {
			ra, rb := f.Readings[a], f.Readings[b]
			switch {
			case ra.Timestamp == nil && rb.Timestamp == nil:
			case ra.Timestamp == nil:
				return 1
			case rb.Timestamp == nil:
				return -1
			case *ra.Timestamp != *rb.Timestamp:
				return cmp.Compare(*ra.Timestamp, *rb.Timestamp)
			}
			return cmp.Compare(ra.RowNumber, rb.RowNumber)
		})
	}
	return out
}

// complete reports whether a reading has every value a frozen comparison needs.
func complete(r *source.Reading) bool {
	return r.ApparentResistivity != nil && r.Chargeability != nil && len(r.DecayCurve) == survey.DecaySamples
}

// frozenRepeat compares a reading with each repeat that has every value, and is
// recorded when the reading has them all and at least one such repeat exists: a
// reading that exactly equals a repeat in resistivity, chargeability and all ten
// decay samples is frozen, since independent measurements never agree to every digit.
func frozenRepeat(f *source.File, i int, members []int) (Result, bool) {
	r := &f.Readings[i]
	if !complete(r) {
		return Result{}, false
	}
	compared := 0
	for _, j := range members {
		if j == i || !complete(&f.Readings[j]) {
			continue
		}
		compared++
		o := &f.Readings[j]
		if *o.ApparentResistivity == *r.ApparentResistivity && *o.Chargeability == *r.Chargeability && slices.Equal(o.DecayCurve, r.DecayCurve) {
			return Result{Criterion: FrozenRepeat, State: Violated, Explanation: fmt.Sprintf("Resistivity, chargeability and all ten decay samples equal those of repeat %s.", o.AcquisitionID), Unit: UnitNone}, true
		}
	}
	if compared == 0 {
		return Result{}, false
	}
	return Result{Criterion: FrozenRepeat, State: Met, Explanation: fmt.Sprintf("No compared repeat has the same resistivity, chargeability and decay samples (%d compared).", compared), Unit: UnitNone}, true
}

// repeatAgreement compares a reading's resistivity with its repeats' reference, the
// median of the repeats not left out; it has no result when the reference does not
// exist.
func repeatAgreement(f *source.File, leftOut []bool, i int, members []int) (Result, bool) {
	r := &f.Readings[i]
	if r.ApparentResistivity == nil {
		return Result{}, false
	}
	var rs []repeats.Repeat
	for _, j := range members {
		if j != i {
			rs = append(rs, repeats.Repeat{Resistivity: f.Readings[j].ApparentResistivity, LeaveOut: leftOut[j]})
		}
	}
	ref, ok := repeats.ReferenceOf(rs)
	if !ok {
		return Result{}, false
	}
	d := ref.Difference(*r.ApparentResistivity)
	limit := repeats.Limit
	res := Result{Criterion: RepeatAgreement, Limit: &limit, Unit: UnitFraction}
	// A difference too large to represent disagrees and stores no value, as a relative
	// error does; it cannot be NaN, since the value is finite and the median finite and
	// above zero.
	by, within := "more than can be represented", "above"
	res.State = RepeatDisagreementState
	if !math.IsInf(d, 0) {
		res.Value = &d
		by = percentAgainst(d, limit)
		if !repeats.Disagrees(d) {
			within, res.State = "within", RepeatAgreementState
		}
	}
	res.Explanation = fmt.Sprintf("Resistivity %s ohm-m differs from the reference %s ohm-m, the median of %d repeats, by %s, %s the %s limit.",
		num(*r.ApparentResistivity), num(ref.Median), ref.Count, by, within, percent(limit))
	return res, true
}

// suspectElectrodes finds the electrodes that transmitted with failing contact to
// at least SuspectPartners different partner electrodes, with each one's partner
// count. Contact is measured on the transmitting pair, so only the tx slots count.
func suspectElectrodes(f *source.File, singles [][]Result) map[string]int {
	partners := map[string]map[string]bool{}
	for i, r := range f.Readings {
		if !violated(singles[i], TransmitterContact) {
			continue
		}
		tx1, tx2 := r.Electrodes[source.SlotIndex(source.SlotTx1)].ID, r.Electrodes[source.SlotIndex(source.SlotTx2)].ID
		if tx1 == nil || tx2 == nil {
			continue
		}
		for _, pair := range [][2]string{{*tx1, *tx2}, {*tx2, *tx1}} {
			if partners[pair[0]] == nil {
				partners[pair[0]] = map[string]bool{}
			}
			partners[pair[0]][pair[1]] = true
		}
	}
	out := map[string]int{}
	for id, p := range partners {
		if len(p) >= SuspectPartners {
			out[id] = len(p)
		}
	}
	return out
}

// suspectElectrode is recorded for every reading with at least one electrode ID: it
// needs review when any slot uses a suspect electrode, named with its partner count.
func suspectElectrode(r *source.Reading, suspect map[string]int) (Result, bool) {
	hasID := false
	for s, e := range r.Electrodes {
		if e.ID == nil {
			continue
		}
		hasID = true
		if n, ok := suspect[*e.ID]; ok {
			return Result{Criterion: SuspectElectrode, State: Review, Explanation: fmt.Sprintf("Electrode %s (%s) transmitted with failing contact to %d different partners.", *e.ID, source.Slots[s], n), Unit: UnitNone}, true
		}
	}
	if !hasID {
		return Result{}, false
	}
	return Result{Criterion: SuspectElectrode, State: Met, Explanation: "No electrode of this reading is suspect.", Unit: UnitNone}, true
}
