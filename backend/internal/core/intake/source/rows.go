package source

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"geolineai/internal/core/survey"
)

// row validates one record's cells under the cell rules, collecting every
// invalid cell and every blank one.
type row struct {
	cells   map[string]string
	invalid []string
	issues  []Issue
}

// parseRow validates one record. A non-empty reason rejects it.
func parseRow(cells []string, header map[string]int, line int) (Reading, string) {
	if len(cells) != len(Columns) {
		return Reading{}, fmt.Sprintf("expected %d fields, found %d", len(Columns), len(cells))
	}
	v := row{cells: map[string]string{}}
	for name, i := range header {
		v.cells[name] = cells[i]
	}

	lineID, lineOK := survey.Lines.Parse(v.cell("line_id"))
	if !lineOK {
		v.reject("line_id", "is not one of "+survey.Lines.Choices())
	}

	r := Reading{RowNumber: line, LineID: lineID, AcquisitionID: v.cell("id")}
	v.blank("id")
	for i, slot := range Slots {
		var e Electrode
		c := slot.Columns()
		if lineOK {
			e.ID = v.electrode(string(c.ID), lineID)
		}
		e.Lat = v.number(string(c.Lat))
		e.Lon = v.number(string(c.Lon))
		e.Alt = v.number(string(c.Alt))
		r.Electrodes[i] = e
	}
	r.Timestamp = v.timestamp("timestamp")
	r.ArrayType = v.arrayType("array_type")
	r.Symmetry = v.positiveInt("symmetry")
	r.Stacks = v.positiveInt("stacks")
	r.InputCurrent = v.number("input_current")
	r.ContactResistance = v.number("contact_resistance")
	r.ApparentResistivity = v.number("apparent_resistivity")
	r.ApparentResistivityErr = v.reportedError("apparent_resistivity_err")
	r.Chargeability = v.number("chargeability")
	r.ChargeabilityErr = v.reportedError("chargeability_err")
	r.DecayCurve = v.decay("decay_curve")

	if len(v.invalid) > 0 {
		return Reading{}, strings.Join(v.invalid, "; ")
	}
	if !usable(r) {
		return Reading{}, unusable
	}
	r.Issues = v.issues
	return r, ""
}

// usableContext is one thing that makes a reading worth reviewing.
type usableContext struct {
	label   string
	present func(Reading) bool
}

// contexts are the usable contexts; a reading needs any one of them.
var contexts = []usableContext{
	{"electrode ID", func(r Reading) bool {
		return slices.ContainsFunc(r.Electrodes[:], func(e Electrode) bool { return e.ID != nil })
	}},
	{"timestamp", func(r Reading) bool { return r.Timestamp != nil }},
	{"apparent resistivity", func(r Reading) bool { return r.ApparentResistivity != nil }},
	{"chargeability", func(r Reading) bool { return r.Chargeability != nil }},
	{"input current", func(r Reading) bool { return r.InputCurrent != nil }},
	{"contact resistance", func(r Reading) bool { return r.ContactResistance != nil }},
	{"decay curve", func(r Reading) bool { return r.DecayCurve != nil }},
}

// usable reports whether a reading has any context worth reviewing.
func usable(r Reading) bool {
	return slices.ContainsFunc(contexts, func(c usableContext) bool { return c.present(r) })
}

// unusable is the rejection reason for a reading with none of the contexts.
var unusable = func() string {
	labels := make([]string, len(contexts))
	for i, c := range contexts {
		labels[i] = c.label
	}
	return "no usable context: no " + english.WordSeries(labels, "or")
}()

// cell is a column's text. The header check puts every column in cells, so an
// unknown name is a programming error, never a blank cell.
func (v *row) cell(field string) string {
	s, ok := v.cells[field]
	if !ok {
		panic(fmt.Sprintf("source: unknown column %q", field))
	}
	return s
}

func (v *row) reject(field, why string) {
	v.invalid = append(v.invalid, fmt.Sprintf("%s \"%s\" %s", field, v.cell(field), why))
}

// isBlank reports whether a cell's text is blank: empty or whitespace only.
func isBlank(s string) bool { return strings.TrimSpace(s) == "" }

// blank records a missing issue for a blank or whitespace-only cell.
func (v *row) blank(field string) bool {
	if !isBlank(v.cell(field)) {
		return false
	}
	v.issues = append(v.issues, Issue{Field: Column(field), Code: IssueMissing, SourceText: v.cell(field)})
	return true
}

func (v *row) electrode(field string, line survey.Line) *string {
	if v.blank(field) {
		return nil
	}
	s := v.cell(field)
	if !survey.IsElectrodeOf(s, line) {
		v.reject(field, "is not an electrode of "+string(line))
		return nil
	}
	return &s
}

var timestampForm = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,9})?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$`)

// timestampLayout stores UTC with nine fractional digits, so text order is
// time order.
const timestampLayout = "2006-01-02T15:04:05.000000000Z"

func (v *row) timestamp(field string) *string {
	if v.blank(field) {
		return nil
	}
	s := strings.TrimSpace(v.cell(field))
	t, err := time.Parse(time.RFC3339Nano, s)
	if !timestampForm.MatchString(s) || err != nil || t.UTC().Year() < 0 || t.UTC().Year() > 9999 {
		v.reject(field, "is not an RFC 3339 date-time")
		return nil
	}
	out := t.UTC().Format(timestampLayout)
	return &out
}

func (v *row) arrayType(field string) *string {
	if v.blank(field) {
		return nil
	}
	s := v.cell(field)
	if s != survey.ArrayDipoleDipole {
		v.reject(field, "is not "+survey.ArrayDipoleDipole)
		return nil
	}
	return &s
}

// MinPositiveInt is the least symmetry (n factor) or stack count a reading has.
const MinPositiveInt = 1

// MinReportedError is the least reported error; a value below it is rejected as
// negative.
const MinReportedError = 0.0

var integerForm = regexp.MustCompile(`^[+-]?[0-9]+$`)

func (v *row) positiveInt(field string) *int64 {
	if v.blank(field) {
		return nil
	}
	s := strings.TrimSpace(v.cell(field))
	n, err := strconv.ParseInt(s, 10, 64)
	if !integerForm.MatchString(s) || err != nil || n < MinPositiveInt {
		v.reject(field, fmt.Sprintf("is not an integer of at least %d", MinPositiveInt))
		return nil
	}
	return &n
}

func (v *row) number(field string) *float64 {
	if v.blank(field) {
		return nil
	}
	f, ok := finite(v.cell(field))
	if !ok {
		v.reject(field, "is not a finite number")
		return nil
	}
	return &f
}

func (v *row) reportedError(field string) *float64 {
	f := v.number(field)
	if f != nil && *f < MinReportedError {
		v.reject(field, "is negative")
		return nil
	}
	return f
}

func (v *row) decay(field string) []float64 {
	if v.blank(field) {
		return nil
	}
	tokens := strings.Split(v.cell(field), ";")
	out := make([]float64, 0, survey.DecaySamples)
	for _, t := range tokens {
		f, ok := finite(t)
		if !ok {
			break
		}
		out = append(out, f)
	}
	if len(tokens) != survey.DecaySamples || len(out) != survey.DecaySamples {
		v.reject(field, fmt.Sprintf("is not %d finite numbers", survey.DecaySamples))
		return nil
	}
	return out
}

// decimalForm is decimal notation only: no hexadecimal, digit separators,
// NaN or infinity.
var decimalForm = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)

var nonZeroDigit = regexp.MustCompile(`[1-9]`)

// finite parses a trimmed decimal number, refusing overflow and underflow (a
// non-zero written value that parses to 0).
func finite(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if !decimalForm.MatchString(s) {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, false
	}
	mantissa, _, _ := strings.Cut(strings.ToLower(s), "e")
	if f == 0 && nonZeroDigit.MatchString(mantissa) {
		return 0, false
	}
	return f, true
}
