package server

import "geolineai/internal/core/closed"

// ErrorCode is the kind of an error response, which the frontend reads to tell a
// refusal it handles from one it only shows; the frontend's copy is generated.
type ErrorCode string

const (
	CodeValidation      ErrorCode = "validation_error"
	CodeReadingNotFound ErrorCode = "reading_not_found"
	CodeNotFound        ErrorCode = "not_found"
	CodeBodyTooLarge    ErrorCode = "body_too_large"
	CodeInternal        ErrorCode = "internal_error"
)

// ErrorCodes is every kind an error response can carry.
var ErrorCodes = closed.Set[ErrorCode]{CodeValidation, CodeReadingNotFound, CodeNotFound, CodeBodyTooLarge, CodeInternal}
