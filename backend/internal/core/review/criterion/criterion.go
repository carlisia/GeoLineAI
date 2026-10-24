package criterion

// The frontend's copy of the types in this file is generated from it; the value
// sets they name are generated where they are declared, into the shared vocabulary.

// Result is one criterion's evaluation of one reading: its state, an explanation
// built from its evidence, and the value and limit compared when it has them, in Unit.
type Result struct {
	Criterion   Criterion   `json:"criterion"`
	State       ResultState `json:"state"`
	Explanation string      `json:"explanation"`
	Value       *float64    `json:"value"`
	Limit       *float64    `json:"limit"`
	Unit        Unit        `json:"unit"`
}

// Evaluation is one reading's advisory with the results it was decided from, in
// criterion table order, and whether the reading is left out of its repeats'
// reference: it has no resistivity, or a violated single-reading result. LeftOut is
// decided here once, stored with the reading and sent as each repeat's own flag, so
// it is not part of the advisory on the wire.
type Evaluation struct {
	Value   Advisory `json:"value"`
	Results []Result `json:"results"`
	LeftOut bool     `json:"-"`
}

// decide is the advisory the results alone give: Fail when any is Violated, Needs
// review when none is Violated but some need review, Pass otherwise.
func decide(results []Result) Advisory {
	value := Pass
	for _, r := range results {
		switch r.State {
		case Violated:
			return Fail
		case Review:
			value = NeedsReview
		case Met:
		default:
			panic("criterion: " + string(r.State) + " is not a result state")
		}
	}
	return value
}
