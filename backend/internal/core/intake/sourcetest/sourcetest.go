// Package sourcetest builds survey CSV text for tests: the header and valid line-1
// rows, with any cell replaced.
package sourcetest

import (
	"fmt"
	"slices"
	"strings"

	"geolineai/internal/core/intake/source"
)

// Header is the source columns, in source-column order.
var Header = strings.Join(source.Columns.Names(), ",")

// CSV joins the header and rows into a file's text.
func CSV(rows ...string) string {
	return strings.Join(append([]string{Header}, rows...), "\n") + "\n"
}

// Row is one valid line-1 row with cells replaced by column name. A key that is
// not a column is a mistake in the test, so it panics.
func Row(set map[string]string) string {
	for c := range set {
		if !slices.Contains(source.Columns, source.Column(c)) {
			panic(fmt.Sprintf("sourcetest: %q is not a column", c))
		}
	}
	cells := make([]string, len(source.Columns))
	for i, c := range source.Columns.Names() {
		cells[i] = valid(c)
		if v, ok := set[c]; ok {
			cells[i] = v
		}
	}
	return strings.Join(cells, ",")
}

// valid is each column's cell in a valid row. Values differ between columns, so a
// test catches one column's value landing in another.
func valid(column string) string {
	switch column {
	case "id":
		return "line-1-r00001"
	case "line_id":
		return "line-1"
	case "tx1_id":
		return "line-1-e01"
	case "tx2_id":
		return "line-1-e02"
	case "rx1_id":
		return "line-1-e03"
	case "rx2_id":
		return "line-1-e04"
	case "tx1_lat":
		return "33.201"
	case "tx1_lon":
		return "-117.101"
	case "tx1_alt":
		return "101"
	case "tx2_lat":
		return "33.202"
	case "tx2_lon":
		return "-117.102"
	case "tx2_alt":
		return "102"
	case "rx1_lat":
		return "33.203"
	case "rx1_lon":
		return "-117.103"
	case "rx1_alt":
		return "103"
	case "rx2_lat":
		return "33.204"
	case "rx2_lon":
		return "-117.104"
	case "rx2_alt":
		return "104"
	case "timestamp":
		return "2026-08-17T14:00:12Z"
	case "array_type":
		return "dipole-dipole"
	case "symmetry":
		return "1"
	case "stacks":
		return "4"
	case "input_current":
		return "1.5"
	case "contact_resistance":
		return "800"
	case "apparent_resistivity":
		return "120"
	case "apparent_resistivity_err":
		return "1.2"
	case "chargeability":
		return "0.01"
	case "chargeability_err":
		return "0.001"
	case "decay_curve":
		return "10;9;8;7;6;5;4;3;2;1"
	case "qc_review":
		return ""
	}
	panic(fmt.Sprintf("sourcetest: no valid value for column %q", column))
}
