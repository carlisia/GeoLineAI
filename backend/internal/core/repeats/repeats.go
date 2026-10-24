// Package repeats compares a reading with its repeats, the other readings of its
// repeat group: how far its resistivity lies from their reference, and whether
// that is a disagreement. The caller says which repeats to leave out of the
// reference, so one definition serves every caller.
package repeats

import (
	"math"
	"slices"
)

// Limit is the repeat difference above which a reading disagrees with its repeats:
// a fifth of their reference resistivity.
const Limit = 0.20

// Repeat is one of a reading's repeats as the comparison sees it: its resistivity,
// nil when unavailable, and whether the caller leaves it out of the reference, as
// it does for a repeat already known to be bad.
type Repeat struct {
	Resistivity *float64
	LeaveOut    bool
}

// Reference is what a reading is compared against: the median resistivity of the
// repeats that count and how many they were.
type Reference struct {
	Median float64
	Count  int
}

// ReferenceOf is the reference of the repeats that count, those with a resistivity
// and not left out. It exists only when at least two count; otherwise there is
// nothing sound to compare against and ok is false. The median is above zero because
// the caller leaves out every repeat whose resistivity is not; this does not check it.
func ReferenceOf(rs []Repeat) (ref Reference, ok bool) {
	var values []float64
	for _, r := range rs {
		if r.Resistivity != nil && !r.LeaveOut {
			values = append(values, *r.Resistivity)
		}
	}
	if len(values) < 2 {
		return Reference{}, false
	}
	return Reference{Median: medianOf(values), Count: len(values)}, true
}

// Difference is how far a reading's resistivity lies from the reference, as a
// fraction of it: |own - median| / median.
func (ref Reference) Difference(own float64) float64 {
	return math.Abs(own-ref.Median) / ref.Median
}

// Disagrees reports whether a repeat difference marks its reading as disagreeing.
func Disagrees(difference float64) bool {
	return difference > Limit
}

// medianOf at least one value: the middle one, or the mean of the middle two.
func medianOf(values []float64) float64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	// Halving each before adding keeps two values near the float64 maximum from
	// overflowing to +Inf, and gives the same result as halving their sum otherwise.
	return sorted[n/2-1]/2 + sorted[n/2]/2
}
