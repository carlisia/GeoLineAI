package criterion

// The limits the criteria compare against; each rests on the domain information or
// on a gap in the shipped data. Each is both recorded on the result and compared
// against, so the two can never differ.
const (
	// ResistivityMin is the apparent resistivity, in ohm-m, a reading must exceed:
	// at or below it, the value is physically implausible.
	ResistivityMin = 0.0
	// ResistivityRelativeErrorLimit is the largest reported resistivity error, as a
	// fraction of the absolute resistivity, that is still Met.
	ResistivityRelativeErrorLimit = 0.50
	// ChargeabilityRelativeErrorLimit is the same for chargeability: an error equal to
	// the value itself carries no information beyond its own uncertainty.
	ChargeabilityRelativeErrorLimit = 1.00
	// ContactResistanceMax is the largest transmitter contact resistance, in ohm, that
	// is still Met; below ContactResistanceMin a contact resistance is impossible.
	ContactResistanceMax = 10000.0
	ContactResistanceMin = 0.0
	// DecaySlopeMax is the slope, in V/V per ms, a decay must stay below: at or above
	// it the curve does not fall.
	DecaySlopeMax = 0.0
	// SuspectPartners is how many different partner electrodes an electrode must have
	// transmitted with, each with failing contact, to be suspect: one failing partner
	// marks the partner's own fault, two mark a fault shared through this electrode.
	SuspectPartners = 2
)
