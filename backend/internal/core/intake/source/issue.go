package source

// The frontend's copy of the types in this file is generated from it.

// Issue is a problem recorded on a retained reading.
type Issue struct {
	Field      Column    `json:"field"`
	Code       IssueCode `json:"code"`
	SourceText string    `json:"source_text"`
}
