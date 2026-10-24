package scenarios

import "geolineai/internal/core/ids"

// The frontend's copy of the types in this file is generated from it.

// Scenario is one scenario resolved to the reading it names.
type Scenario struct {
	Label     string `json:"label"`
	ReadingID ids.ID `json:"reading_id"`
}

// List is the answer to the scenarios request.
type List struct {
	Scenarios []Scenario `json:"scenarios"`
}
