package repeats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"geolineai/internal/core/repeats"
)

func f(v float64) *float64 { return &v }

func rs(values ...*float64) []repeats.Repeat {
	out := make([]repeats.Repeat, len(values))
	for i, v := range values {
		out[i] = repeats.Repeat{Resistivity: v}
	}
	return out
}

// A difference computed against the wrong set, or given where the reference is too
// thin, would mark agreement as disagreement or hide a real one; the limit is a
// boundary, so a difference exactly at it agrees and one just over disagrees.
func TestDifferenceNeedsTwoCountingRepeatsAndDisagreesOnlyAboveTheLimit(t *testing.T) {
	for _, c := range []struct {
		name      string
		own       float64
		rs        []repeats.Repeat
		want      *float64
		disagrees bool
	}{
		{"half above the median of two equal repeats", 150, rs(f(100), f(100)), f(0.5), true},
		{"a quarter off a median between two repeats", 150, rs(f(100), f(300)), f(0.25), true},
		{"exactly a fifth below three equal repeats", 80, rs(f(100), f(100), f(100)), f(0.2), false},
		{"just over a fifth above two equal repeats", 120.00000001, rs(f(100), f(100)), f(0.2000000001), true},
		{"the median is the middle of three unequal repeats", 200, rs(f(300), f(100), f(200)), f(0), false},
		{"a negative resistivity is a whole span from a positive median", -100, rs(f(100), f(100)), f(2), true},
		{"one repeat is too few", 150, rs(f(100)), nil, false},
		{"a repeat without a resistivity does not count", 150, rs(f(100), nil), nil, false},
		{"no repeats", 150, nil, nil, false},
	} {
		ref, ok := repeats.ReferenceOf(c.rs)
		if c.want == nil {
			assert.False(t, ok, c.name)
			continue
		}
		if assert.True(t, ok, c.name) {
			got := ref.Difference(c.own)
			assert.InDelta(t, *c.want, got, 1e-12, c.name)
			assert.Equal(t, c.disagrees, repeats.Disagrees(got), c.name)
		}
	}
}
