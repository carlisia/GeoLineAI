package assist

import (
	"context"
	"encoding/json"
	"errors"
)

// FailingContact is the transmitting pair of one reading whose Transmitter contact
// result is unresolved; a blank slot is nil.
type FailingContact struct {
	Tx1, Tx2 *string
}

// Kind is the type of a typed question, as the judgment model reads it.
type Kind string

// KindNoul is the one kind asked: how likely a question's instructions hold.
const KindNoul Kind = "noul"

// Question is one typed question: a Noul answers how likely its Instructions hold.
// Instructions may be any JSON value, so a question can carry the data it refers to
// beside its wording.
type Question struct {
	Kind         Kind
	Instructions any
	Criteria     map[string]string
}

// Answer is one typed answer: a Noul's probability of yes, from 0 to 1.
type Answer struct {
	Kind Kind
	Noul float64
}

// Judge answers typed questions over one state in one request, and names the model
// version that answered.
type Judge interface {
	Judge(ctx context.Context, state any, questions map[string]Question) (answers map[string]Answer, model string, err error)
}

// Writer asks a text model for one JSON object held to schema, the system message
// saying what to write and the user message holding the input, and names the model
// version that wrote it.
type Writer interface {
	Write(ctx context.Context, system, user string, schema map[string]any) (object json.RawMessage, model string, err error)
}

// ErrUnavailable wraps every failure to reach a model: a refused or failed request,
// including one still refused after the configured retries.
var ErrUnavailable = errors.New("the model could not be reached")

// ErrInvalidResponse wraps every answer a model gave that cannot be read: JSON that
// does not parse, a required field missing, a value of the wrong type, or a
// probability missing or outside 0 to 1. It is never read as a valid answer.
var ErrInvalidResponse = errors.New("the model's answer could not be read")
