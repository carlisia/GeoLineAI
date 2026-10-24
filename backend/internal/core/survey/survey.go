// Package survey holds the facts of the survey itself: its lines, how an
// electrode is named on a line, the decay sample count and the array type.
package survey

import (
	"fmt"
	"slices"

	"geolineai/internal/core/closed"
)

// Line is one survey line. Its values are exactly the wire and storage spellings,
// and one that came from Lines.Parse or Lines needs no second check.
type Line string

// Lines are the survey lines: the one line the application ships.
var Lines = closed.Set[Line]{"line-1"}

// electrodesPerLine is how many electrodes a line has: electrode IDs run from e01 to
// e40.
const electrodesPerLine = 40

// ElectrodeIDs lists a line's electrode IDs in index order, e01 to the last: the one
// spelling of an electrode ID, which the page offers and IsElectrodeOf accepts.
func ElectrodeIDs(line Line) []string {
	out := make([]string, electrodesPerLine)
	for i := range out {
		out[i] = fmt.Sprintf("%s-e%02d", line, i+1)
	}
	return out
}

// IsElectrodeOf reports whether id, taken untrimmed, is one of line's electrode IDs.
func IsElectrodeOf(id string, line Line) bool {
	return slices.Contains(ElectrodeIDs(line), id)
}

// DecaySamples is the number of decay samples a reading carries, one per gate.
const DecaySamples = 10

// DecayGateCentresMs are the centres of the ten time gates after current shutoff, in
// milliseconds: a fact of the acquisition, the same for every reading, so no reading
// carries them.
var DecayGateCentresMs = [DecaySamples]float64{24.32, 35.96, 53.18, 78.64, 116.30, 171.97, 254.31, 376.06, 556.10, 822.34}

// ArrayDipoleDipole is the only electrode arrangement the data uses.
const ArrayDipoleDipole = "dipole-dipole"
