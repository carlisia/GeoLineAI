package criterion

import (
	"geolineai/internal/core/closed"
	"geolineai/internal/core/intake/source"
)

// The value sets in this file are generated into the frontend's vocabulary.

// Criterion names a standard every reading is evaluated against. Its values
// are exactly the wire and storage spellings.
type Criterion string

const (
	ImportIssue                Criterion = "import_issue"
	ResistivitySign            Criterion = "resistivity_sign"
	ResistivityRelativeError   Criterion = "resistivity_relative_error"
	ChargeabilityRelativeError Criterion = "chargeability_relative_error"
	TransmitterContact         Criterion = "transmitter_contact"
	DecayTrend                 Criterion = "decay_trend"
	FrozenRepeat               Criterion = "frozen_repeat"
	RepeatAgreement            Criterion = "repeat_agreement"
	SuspectElectrode           Criterion = "suspect_electrode"
)

// Criteria lists every criterion in table order, the order a reading's results are
// recorded and shown in: the single-reading criteria, then those across readings.
var Criteria = closed.Set[Criterion]{
	ImportIssue, ResistivitySign, ResistivityRelativeError, ChargeabilityRelativeError, TransmitterContact, DecayTrend,
	FrozenRepeat, RepeatAgreement, SuspectElectrode,
}

// Labels are the criteria's display labels.
var Labels = map[Criterion]string{
	ImportIssue:                "Import issue",
	ResistivitySign:            "Resistivity sign",
	ResistivityRelativeError:   "Resistivity error",
	ChargeabilityRelativeError: "Chargeability error",
	TransmitterContact:         "Transmitter contact",
	DecayTrend:                 "Decay trend",
	FrozenRepeat:               "Frozen repeat",
	RepeatAgreement:            "Repeat agreement",
	SuspectElectrode:           "Suspect electrode",
}

// ResultState is one criterion's evaluation of one reading. Met and Violated are decided;
// Review needs a human. A Violated or Review result is unresolved.
type ResultState string

const (
	Met      ResultState = "met"
	Violated ResultState = "violated"
	Review   ResultState = "review"
)

// ResultStates lists every result state.
var ResultStates = closed.Set[ResultState]{Met, Violated, Review}

// The states a repeat agreement result takes: a reading that disagrees with its
// repeats needs review, one that agrees has met the criterion. The import records
// them once; the reading detail and every Disagrees flag read them from the record.
const (
	RepeatDisagreementState = Review
	RepeatAgreementState    = Met
)

// UnresolvedStates are the states that need the reviewer's attention, violations
// first: the order a reason spells them out in, and the set a contact read uses.
var UnresolvedStates = []ResultState{Violated, Review}

// Unit is what a result's value and limit are measured in, "" when it compares none.
// Its values are exactly the wire and storage spellings.
type Unit string

const (
	UnitOhmM     Unit = "ohm-m"
	UnitOhm      Unit = "ohm"
	UnitVPerV    Unit = "V/V"
	UnitFraction Unit = "fraction"
	UnitSlope    Unit = "V/V per ms"
	UnitNone     Unit = ""
)

// Units lists every unit a result can carry.
var Units = closed.Set[Unit]{UnitOhmM, UnitOhm, UnitVPerV, UnitFraction, UnitSlope, UnitNone}

// ResultStateLabels are the states' display labels.
var ResultStateLabels = map[ResultState]string{Met: "Met", Violated: "Violated", Review: "Review"}

// Advisory is what the results say about one reading, never what the reviewer
// decided. Its values are exactly the wire and storage spellings.
type Advisory string

const (
	Pass        Advisory = "pass"
	Fail        Advisory = "fail"
	NeedsReview Advisory = "needs_review"
)

// Advisories lists every advisory value, in display order.
var Advisories = closed.Set[Advisory]{Pass, Fail, NeedsReview}

// AdvisoryLabels are the advisory values' display labels.
var AdvisoryLabels = map[Advisory]string{Pass: "Pass", Fail: "Fail", NeedsReview: "Needs review"}

// AdvisoryMeanings say what each advisory value means, as a saved reason and the page
// both phrase it.
var AdvisoryMeanings = map[Advisory]string{
	Pass:        "every criterion met",
	Fail:        "a criterion was violated",
	NeedsReview: "no criterion was violated but some need review",
}

// IssueLabels are the import issue codes' display labels, which an import issue
// result's explanation names.
var IssueLabels = map[source.IssueCode]string{source.IssueMissing: "Missing value", source.IssueIDCollision: "Repeated acquisition ID"}
