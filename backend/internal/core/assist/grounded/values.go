package grounded

import "geolineai/internal/core/closed"

// The value sets in this file are generated into the frontend's vocabulary.

// SourceKind says where a source's text comes from.
type SourceKind string

const (
	// KindRule is a criterion's rule as the code implements it, built from its limits.
	KindRule SourceKind = "rule"
	// KindDoc is a passage quoted verbatim from the repository's documentation.
	KindDoc SourceKind = "doc"
	// KindResult is a criterion result recorded for a reading at import.
	KindResult SourceKind = "result"
	// KindFact is a value or count read or computed from what storage holds.
	KindFact SourceKind = "fact"
)

// SourceKinds lists every kind of source.
var SourceKinds = closed.Set[SourceKind]{KindRule, KindDoc, KindResult, KindFact}

// Outcome is what became of one written answer.
type Outcome string

const (
	// Shown: every sentence passed every check.
	Shown Outcome = "shown"
	// Withheld: a sentence failed a check, so no sentence is shown.
	Withheld Outcome = "withheld"
	// Unanswerable: the text model found no answer in the sources.
	Unanswerable Outcome = "unanswerable"
	// Unavailable: no answer could be obtained; the answer's reason says why.
	Unavailable Outcome = "unavailable"
)

// Outcomes lists every outcome.
var Outcomes = closed.Set[Outcome]{Shown, Withheld, Unanswerable, Unavailable}

// UnavailableReason is why an answer is unavailable.
type UnavailableReason string

const (
	// NoKey: the server has no provider key, so no model is asked.
	NoKey UnavailableReason = "no_key"
	// ProviderError: a model could not be reached, after the configured retries.
	ProviderError UnavailableReason = "provider_error"
	// InvalidResponse: a model answered in a shape that cannot be read.
	InvalidResponse UnavailableReason = "invalid_response"
)

// UnavailableReasons lists every reason an answer can be unavailable.
var UnavailableReasons = closed.Set[UnavailableReason]{NoKey, ProviderError, InvalidResponse}
