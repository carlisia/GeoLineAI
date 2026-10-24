package criterion

import (
	"fmt"

	"geolineai/internal/core/closed"

	"geolineai/internal/core/repeats"
	"geolineai/internal/core/survey"
)

// Rule states, for a reader, the rule a criterion applies: when it is met, when it is
// not, and when it records no result. The limits are the ones the criterion compares
// against, so the text and the comparison cannot differ.
func Rule(c Criterion) string {
	switch c {
	case ImportIssue:
		return "Import issue is recorded once for each import issue on the reading, a missing value or a repeated acquisition ID, and is always Review: a cell the import cannot store rejects its whole row instead, so no import issue is Violated."
	case ResistivitySign:
		return fmt.Sprintf("Resistivity sign is Met when the apparent resistivity is above %s ohm-m and Violated when it is at or below %s ohm-m. It records no result when the apparent resistivity is unavailable.", num(ResistivityMin), num(ResistivityMin))
	case ResistivityRelativeError:
		return relativeErrorRule(resistivityError, "apparent resistivity")
	case ChargeabilityRelativeError:
		return relativeErrorRule(chargeabilityError, "chargeability")
	case TransmitterContact:
		return fmt.Sprintf("Transmitter contact is Met when the contact resistance is from %s to %s ohm and Violated when it is below %s ohm or above %s ohm. The contact resistance belongs to the transmitting pair of the reading, its tx1 and tx2 electrodes, and says nothing about its receiving pair. It records no result when the contact resistance is unavailable.",
			num(ContactResistanceMin), num(ContactResistanceMax), num(ContactResistanceMin), num(ContactResistanceMax))
	case DecayTrend:
		return fmt.Sprintf("Decay trend fits an unweighted least-squares line, with a free intercept, through the ten decay samples against the ten gate centres, from %s to %s ms, both on linear scales. It is Met when the slope is below %s V/V per ms, so the curve falls, and Violated when the slope is at or above %s V/V per ms or is not finite. It records no result when the decay curve is unavailable.",
			num(survey.DecayGateCentresMs[0]), num(survey.DecayGateCentresMs[survey.DecaySamples-1]), num(DecaySlopeMax), num(DecaySlopeMax))
	case FrozenRepeat:
		return "Frozen repeat compares the reading with each of its repeats that has an apparent resistivity, a chargeability and all ten decay samples. It is Violated when all of those values exactly equal a repeat's, since independent measurements never agree to every digit, and Met when they differ from every compared repeat. It records no result when the reading lacks one of those values or no repeat has them all."
	case RepeatAgreement:
		return fmt.Sprintf("Repeat agreement compares the apparent resistivity with the reference, the median resistivity of the reading's repeats, leaving out repeats without a resistivity and repeats with a Violated single-reading result. The repeat difference is the absolute difference between the resistivity and the reference, divided by the reference. It is Met when the repeat difference is at most %s and Review when it is above %s. It records no result when the reading has no resistivity or fewer than two repeats count toward the reference.",
			percent(repeats.Limit), percent(repeats.Limit))
	case SuspectElectrode:
		return fmt.Sprintf("An electrode is suspect when the readings in which it transmits, in the tx1 or tx2 slot, violate Transmitter contact with at least %d different partner electrodes. Suspect electrode is Review when the reading uses a suspect electrode in any slot and Met when it uses none; it marks an association to review, not a proven fault. It records no result when the reading has no electrode ID.", SuspectPartners)
	}
	panic("criterion: " + string(c) + " is not a criterion")
}

// relativeErrorRule is the rule of a relative-error criterion.
func relativeErrorRule(c relativeErrorCriterion, value string) string {
	return fmt.Sprintf("%s error divides the reported %s error by the absolute %s. It is Met when that fraction is at most %s and Violated when it is above %s or too large to represent. It records no result when either value is unavailable or the %s is 0.",
		c.label, value, value, num(c.limit), num(c.limit), value)
}

// AdvisoryRule states how the results decide the advisory, and what the advisory
// never does.
func AdvisoryRule() string {
	return fmt.Sprintf("The advisory is %s when any result is Violated, %s when no result is Violated but some are Review, and %s otherwise, so %s means every recorded result is Met. It is decided from the results alone at import and never changes afterwards, and it never sets the verdict, which only the reviewer's save changes.",
		advisoryLabel(Fail), advisoryLabel(NeedsReview), advisoryLabel(Pass), advisoryLabel(Pass))
}

// advisoryLabel is an advisory value's display label; every value has one.
func advisoryLabel(a Advisory) string {
	return closed.Entry(AdvisoryLabels, a, "advisory")
}
