package server_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"geolineai/internal/core/intake/sourcetest"
	"geolineai/internal/testutil"
)

// groupCSV holds line-1 rows in six geometries: three that agree (d, e, f);
// four of which one differs by a tenth (k, l, m, q); a pair, too few for a difference
// (g, h); three of which one disagrees (a, b, c, c at 0.5 and the others at the 0.20
// limit); four of which one disagrees just over the limit (n, o, p, r at 0.21); a
// five of which two disagree, the others at the limit (s, t, u, v, w); a single
// reading (i); and one with a blank slot (j). Every row shares a time, so reading
// order is import order.
var groupCSV = sourcetest.CSV(
	sourcetest.Row(map[string]string{"id": "d", "tx1_id": "line-1-e05", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "e", "tx1_id": "line-1-e05", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "f", "tx1_id": "line-1-e05", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "k", "tx1_id": "line-1-e08", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "l", "tx1_id": "line-1-e08", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "m", "tx1_id": "line-1-e08", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "q", "tx1_id": "line-1-e08", "apparent_resistivity": "110"}),
	sourcetest.Row(map[string]string{"id": "g", "tx1_id": "line-1-e06", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "h", "tx1_id": "line-1-e06", "apparent_resistivity": "300"}),
	sourcetest.Row(map[string]string{"id": "a", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "b", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "c", "apparent_resistivity": "150"}),
	sourcetest.Row(map[string]string{"id": "n", "tx1_id": "line-1-e09", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "o", "tx1_id": "line-1-e09", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "p", "tx1_id": "line-1-e09", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "r", "tx1_id": "line-1-e09", "apparent_resistivity": "121"}),
	sourcetest.Row(map[string]string{"id": "s", "tx1_id": "line-1-e10", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "t", "tx1_id": "line-1-e10", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "u", "tx1_id": "line-1-e10", "apparent_resistivity": "100"}),
	sourcetest.Row(map[string]string{"id": "v", "tx1_id": "line-1-e10", "apparent_resistivity": "150"}),
	sourcetest.Row(map[string]string{"id": "w", "tx1_id": "line-1-e10", "apparent_resistivity": "150"}),
	sourcetest.Row(map[string]string{"id": "i", "tx1_id": "line-1-e07"}),
	sourcetest.Row(map[string]string{"id": "j", "tx1_id": ""}),
)

func groupFixture(t *testing.T) *testutil.App {
	a := testutil.NewApp(t)
	a.Import("/data/line-1.csv", groupCSV)
	return a
}

// A repeat listed against the wrong group, or a difference computed against the
// wrong set, would have the reviewer judge a reading against readings that are not
// its repeats.
func TestDetailListsRepeatsWithTheirDifferences(t *testing.T) {
	a := groupFixture(t)
	ids := a.IDsByAcquisition()
	detail := func(name string) map[string]any { return detailOf(t, a, ids[name]) }
	repeats := func(d map[string]any) []string {
		var out []string
		for _, r := range d["repeats"].([]any) {
			m := r.(map[string]any)
			out = append(out, fmt.Sprintf("%v %v %v %v", m["acquisition_id"], m["apparent_resistivity"], m["repeat_difference"], m["disagrees"]))
		}
		return out
	}

	c := detail("c")
	assert.Equal(t, detail("a")["repeat_group_id"], c["repeat_group_id"], "a, b and c share a group")
	assert.NotEqual(t, detail("d")["repeat_group_id"], c["repeat_group_id"], "d is in another")
	assert.Equal(t, 0.5, c["repeat_difference"])
	assert.Equal(t, true, c["disagrees"])
	assert.Equal(t, []string{"a 100 0.2 false", "b 100 0.2 false"}, repeats(c), "the repeats in reading order, each against its own other repeats")

	g := detail("g")
	assert.Nil(t, g["repeat_difference"], "one repeat is too few for a difference")
	assert.Equal(t, false, g["disagrees"])
	assert.Equal(t, []string{"h 300 <nil> false"}, repeats(g))

	i := detail("i")
	assert.NotNil(t, i["repeat_group_id"], "a single reading still has its group")
	assert.Equal(t, []any{}, i["repeats"])

	j := detail("j")
	assert.Nil(t, j["repeat_group_id"], "a blank slot leaves the reading without a group")
	assert.Equal(t, []any{}, j["repeats"])
}
