package criterion

import (
	"fmt"
	"math"

	"geolineai/internal/core/closed"
	"geolineai/internal/core/intake/source"
)

// relativeErrorCriterion is what a relative-error criterion compares: which criterion
// it is, how it names its measurement, the measurement's unit and the limit.
type relativeErrorCriterion struct {
	name  Criterion
	label string
	unit  Unit
	limit float64
}

var (
	resistivityError   = relativeErrorCriterion{ResistivityRelativeError, "Resistivity", UnitOhmM, ResistivityRelativeErrorLimit}
	chargeabilityError = relativeErrorCriterion{ChargeabilityRelativeError, "Chargeability", UnitVPerV, ChargeabilityRelativeErrorLimit}
)

// single evaluates the criteria that look at one reading alone, in table order. A
// criterion whose input is unavailable produces no result: the input's import issue
// already records that fact. Values are compared at full precision.
func single(r *source.Reading) []Result {
	var out []Result
	for _, i := range r.Issues {
		out = append(out, importIssue(i))
	}
	if r.ApparentResistivity != nil {
		out = append(out, resistivitySign(*r.ApparentResistivity))
		if r.ApparentResistivityErr != nil {
			if res, ok := relativeError(resistivityError, *r.ApparentResistivity, *r.ApparentResistivityErr); ok {
				out = append(out, res)
			}
		}
	}
	if r.Chargeability != nil && r.ChargeabilityErr != nil {
		if res, ok := relativeError(chargeabilityError, *r.Chargeability, *r.ChargeabilityErr); ok {
			out = append(out, res)
		}
	}
	if r.ContactResistance != nil {
		out = append(out, transmitterContact(*r.ContactResistance))
	}
	if r.DecayCurve != nil {
		out = append(out, decayTrend(r.DecayCurve))
	}
	return out
}

func importIssue(i source.Issue) Result {
	label := closed.Entry(IssueLabels, i.Code, "import issue code")
	return Result{Criterion: ImportIssue, State: Review, Explanation: fmt.Sprintf("Import issue on %s: %s.", i.Field, label), Unit: UnitNone}
}

func resistivitySign(rho float64) Result {
	limit := ResistivityMin
	res := Result{Criterion: ResistivitySign, Value: &rho, Limit: &limit, Unit: UnitOhmM}
	res.State, res.Explanation = Met, "above"
	if rho <= limit {
		res.State, res.Explanation = Violated, "at or below"
	}
	res.Explanation = fmt.Sprintf("Apparent resistivity %s ohm-m is %s %s.", num(rho), res.Explanation, num(limit))
	return res
}

// relativeError records a reported error against its value only when the value is
// not zero, since dividing by zero has no answer.
func relativeError(c relativeErrorCriterion, value, reported float64) (Result, bool) {
	ratio, ok := relativeRatio(value, reported)
	if !ok {
		return Result{}, false
	}
	limit := c.limit
	res := Result{Criterion: c.name, Limit: &limit, Unit: UnitFraction}
	if math.IsInf(ratio, 0) {
		res.State = Violated
		res.Explanation = fmt.Sprintf("%s error %s %s against the value %s %s is too large to represent; the limit is %s.", c.label, num(reported), c.unit, num(value), c.unit, num(limit))
		return res, true
	}
	res.Value = &ratio
	res.State, res.Explanation = Met, "at most"
	if ratio > limit {
		res.State, res.Explanation = Violated, "above"
	}
	res.Explanation = fmt.Sprintf("%s error %s %s is %s of the value %s %s, %s the %s allowed.", c.label, num(reported), c.unit, numAgainst(ratio, limit), num(value), c.unit, res.Explanation, num(limit))
	return res, true
}

// transmitterContact records the limit the value was judged against: the floor for a
// negative value, the ceiling otherwise.
func transmitterContact(ohm float64) Result {
	res := Result{Criterion: TransmitterContact, Value: &ohm, Unit: UnitOhm}
	floor, ceiling := ContactResistanceMin, ContactResistanceMax
	switch {
	case ohm < floor:
		res.Limit = &floor
		res.State, res.Explanation = Violated, fmt.Sprintf("Transmitter contact resistance %s ohm is below %s ohm.", numAgainst(ohm, floor), num(floor))
	case ohm > ceiling:
		res.Limit = &ceiling
		res.State, res.Explanation = Violated, fmt.Sprintf("Transmitter contact resistance %s ohm is above %s ohm.", numAgainst(ohm, ceiling), num(ceiling))
	default:
		res.Limit = &ceiling
		res.State, res.Explanation = Met, fmt.Sprintf("Transmitter contact resistance %s ohm is within %s to %s ohm.", num(ohm), num(floor), num(ceiling))
	}
	return res
}

func decayTrend(samples []float64) Result {
	limit := DecaySlopeMax
	res := Result{Criterion: DecayTrend, Limit: &limit, Unit: UnitSlope}
	slope := decaySlope(samples)
	if math.IsNaN(slope) || math.IsInf(slope, 0) {
		res.State, res.Explanation = Violated, "The decay has no finite slope."
		return res
	}
	res.Value = &slope
	res.State, res.Explanation = Met, "below %s: the curve falls"
	if slope >= limit {
		res.State, res.Explanation = Violated, "at or above %s: the curve does not fall"
	}
	res.Explanation = fmt.Sprintf("Decay slope %s V/V per ms is "+res.Explanation+".", num(slope), num(limit))
	return res
}
