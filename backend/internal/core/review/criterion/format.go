package criterion

import "strconv"

// The explanations' number formatting: exact where the source is exact, and apart
// from a limit wherever a value is printed against one.

// num writes a number for an explanation: exact for the short decimals the source
// holds, and to six significant digits for a computed ratio or slope.
func num(x float64) string {
	return sig(x, 6)
}

func sig(x float64, digits int) string {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	if len(s) > 12 {
		s = strconv.FormatFloat(x, 'g', digits, 64)
	}
	return s
}

// numAgainst writes a value compared with a limit so that the two read differently
// whenever they differ: digits are added until they do, so an explanation never says
// a value is above a limit it prints as equal to.
func numAgainst(value, limit float64) string {
	if s, ok := apart(value, limit, 6, 17, sig); ok {
		return s
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// apart writes value with format at the fewest digits, from one count to another,
// that read differently from limit when the two differ; ok is false when no count
// does, and the caller falls back to an exact form.
func apart(value, limit float64, from, to int, format func(x float64, digits int) string) (s string, ok bool) {
	for digits := from; digits <= to; digits++ {
		if s := format(value, digits); value == limit || s != format(limit, digits) {
			return s, true
		}
	}
	return "", false
}

// percent writes a fraction as a percentage with one decimal, as the page shows one.
func percent(fraction float64) string {
	return percentTo(fraction, 1)
}

func percentTo(fraction float64, decimals int) string {
	return strconv.FormatFloat(fraction*100, 'f', decimals, 64) + "%"
}

// percentAgainst is percent for a value compared with a limit, with the digits it
// takes to read differently from the limit whenever they differ.
func percentAgainst(fraction, limit float64) string {
	if s, ok := apart(fraction, limit, 1, 12, percentTo); ok {
		return s
	}
	return strconv.FormatFloat(fraction*100, 'g', -1, 64) + "%"
}
