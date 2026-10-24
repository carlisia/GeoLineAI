package server

import "geolineai/internal/core/ids"

// The frontend's copy of the types in this file is generated from it.

// ErrorBody is every error response: the error's kind, the message the reviewer
// is shown word for word, and the details of that kind.
type ErrorBody struct {
	Error   ErrorCode `json:"error"`
	Message string    `json:"message"`
	// The reading a reading_not_found error names.
	MissingReadingIDs []ids.ID `json:"missing_reading_ids,omitempty"`
}
