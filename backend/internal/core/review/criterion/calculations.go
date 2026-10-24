package criterion

import (
	"fmt"
	"math"

	"geolineai/internal/core/survey"
)

// decaySlope is the slope of the unweighted least-squares line, with a free
// intercept, through the samples against the gate centres, both on linear scales,
// in V/V per ms. The samples are one per gate, which the import guarantees.
func decaySlope(samples []float64) float64 {
	if len(samples) != survey.DecaySamples {
		panic(fmt.Sprintf("criterion: a decay curve has %d samples, not %d", len(samples), survey.DecaySamples))
	}
	var meanX, meanY float64
	for i, y := range samples {
		meanX += survey.DecayGateCentresMs[i]
		meanY += y
	}
	meanX /= survey.DecaySamples
	meanY /= survey.DecaySamples
	var cov, varX float64
	for i, y := range samples {
		dx := survey.DecayGateCentresMs[i] - meanX
		cov += dx * (y - meanY)
		varX += dx * dx
	}
	return cov / varX
}

// relativeRatio is a reported error as a fraction of its measured value's magnitude.
// It has no answer for a value of zero, so ok is false then; a ratio too large to
// represent is returned as +Inf for the caller to name.
func relativeRatio(value, reported float64) (ratio float64, ok bool) {
	if value == 0 {
		return 0, false
	}
	return reported / math.Abs(value), true
}
