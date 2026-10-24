package verdict

import "geolineai/internal/core/ids"

// The frontend's copy of the types in this file is generated from it.

// Outcome lists which targets changed and which already had the verdict, and names
// the review action the save recorded, nil when every target was unchanged.
type Outcome struct {
	Affected  []ids.ID `json:"affected_reading_ids"`
	Unchanged []ids.ID `json:"unchanged_reading_ids"`
	ActionID  *ids.ID  `json:"action_id"`
}
