package source

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"geolineai/internal/core/closed"
	"geolineai/internal/core/survey"
)

// Column is one of the 30 source columns, spelled as in the header and on the wire.
type Column string

// Columns are the 30 source columns, in source-column order.
var Columns = closed.Set[Column]{
	"id", "line_id",
	"tx1_id", "tx1_lat", "tx1_lon", "tx1_alt",
	"tx2_id", "tx2_lat", "tx2_lon", "tx2_alt",
	"rx1_id", "rx1_lat", "rx1_lon", "rx1_alt",
	"rx2_id", "rx2_lat", "rx2_lon", "rx2_alt",
	"timestamp", "array_type", "symmetry", "stacks",
	"input_current", "contact_resistance",
	"apparent_resistivity", "apparent_resistivity_err",
	"chargeability", "chargeability_err",
	"decay_curve", "qc_review",
}

// ColumnIndex is a column's position in source-column order; an unknown column
// is a programming error.
func ColumnIndex(column Column) int {
	i := slices.Index(Columns, column)
	if i < 0 {
		panic(fmt.Sprintf("source: unknown column %q", column))
	}
	return i
}

// Slot is one of the four electrode slots, spelled as in its column names and on
// the wire.
type Slot string

// The four electrode slots: the transmitting pair and the receiving pair.
const (
	SlotTx1 Slot = "tx1"
	SlotTx2 Slot = "tx2"
	SlotRx1 Slot = "rx1"
	SlotRx2 Slot = "rx2"
)

// Slots are the four electrode slots, in source-column order.
var Slots = [...]Slot{SlotTx1, SlotTx2, SlotRx1, SlotRx2}

// SlotIndex is a slot's position in source-column order, where a reading and a
// geometry hold their electrodes; an unknown slot is a programming error.
func SlotIndex(slot Slot) int {
	i := slices.Index(Slots[:], slot)
	if i < 0 {
		panic(fmt.Sprintf("source: unknown slot %q", slot))
	}
	return i
}

// SlotNames lists every slot's spelling, in source-column order.
func SlotNames() []string { return closed.Set[Slot](Slots[:]).Names() }

// SlotColumns are a slot's four columns: its electrode ID, latitude, longitude
// and altitude.
type SlotColumns struct{ ID, Lat, Lon, Alt Column }

// Columns names a slot's four columns, in source-column order.
func (s Slot) Columns() SlotColumns {
	return SlotColumns{Column(s + "_id"), Column(s + "_lat"), Column(s + "_lon"), Column(s + "_alt")}
}

// IssueCode is an import issue code, spelled as persisted and sent.
type IssueCode string

// Import issue codes.
const (
	IssueMissing     IssueCode = "missing"
	IssueIDCollision IssueCode = "id_collision"
)

// IssueCodes lists every import issue code.
var IssueCodes = closed.Set[IssueCode]{IssueMissing, IssueIDCollision}

// Electrode is one slot of a reading's geometry; nil means unavailable.
type Electrode struct {
	ID            *string
	Lat, Lon, Alt *float64
}

// Reading is one retained row. A nil value is unavailable, and its field
// carries a missing issue.
type Reading struct {
	RowNumber     int
	LineID        survey.Line
	AcquisitionID string                // verbatim, including blank text
	Electrodes    [len(Slots)]Electrode // in Slots order
	Timestamp     *string               // UTC, nine fractional digits
	ArrayType     *string
	Symmetry      *int64
	Stacks        *int64

	InputCurrent           *float64
	ContactResistance      *float64
	ApparentResistivity    *float64
	ApparentResistivityErr *float64
	Chargeability          *float64
	ChargeabilityErr       *float64
	DecayCurve             []float64 // survey.DecaySamples values, or nil

	Issues []Issue // in no particular order; a reader that needs one sorts them

	// RepeatGroup indexes the file's Groups; nil when a slot is blank, so the reading
	// belongs to no group.
	RepeatGroup *int
}

// Rejection is a row that failed retention.
type Rejection struct {
	Row    int
	Reason string
}

// File is what a CSV parses into.
type File struct {
	Readings   []Reading
	Rejections []Rejection
	// Groups are the file's repeat groups, in order of first appearance; a reading's
	// RepeatGroup indexes them.
	Groups []Group
}

// Geometry is the four electrode IDs of a reading, in slot order.
type Geometry [len(Slots)]string

// Group is one repeat group of a file: the line and geometry its readings share.
type Group struct {
	Line survey.Line
	Geometry
}

// fileError fails the whole file: an unreadable header or framing. line is the
// physical line on which the failing record starts.
type fileError struct {
	line    int
	message string
	header  bool // the failing record is the header
}

func (e *fileError) Error() string {
	if e.header {
		return fmt.Sprintf("header (line %d): %s", e.line, e.message)
	}
	return fmt.Sprintf("record starting on line %d: %s", e.line, e.message)
}

// Parse reads one CSV. On a fileError it also returns the rejections found
// before the failure; they are diagnostics only, since nothing is retained.
func Parse(r io.Reader) (File, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return File{}, fmt.Errorf("read: %w", err)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1 // parseRow counts the fields, to name the count it found

	var out File
	var header map[string]int
	for {
		rec, cells, reason, err := readRecord(reader, data)
		if err == io.EOF {
			break
		}
		switch {
		case err != nil:
			return File{Rejections: out.Rejections}, err
		case header == nil && reason != "":
			return File{}, &fileError{line: rec.line, message: reason, header: true}
		case header == nil:
			if header, err = parseHeader(cells, rec.line); err != nil {
				return File{}, err
			}
			continue
		}
		var reading Reading
		if reason == "" {
			reading, reason = parseRow(cells, header, rec.line)
		}
		if reason != "" {
			out.Rejections = append(out.Rejections, Rejection{Row: rec.line, Reason: reason})
			continue
		}
		out.Readings = append(out.Readings, reading)
	}
	if header == nil {
		return File{}, &fileError{line: 1, message: "the file is empty", header: true}
	}
	markCollisions(out.Readings)
	out.Groups = numberGroups(out.Readings)
	return out, nil
}

// record is one CSV record: its text in the file, line endings included, and the
// physical line it starts on.
type record struct {
	text string
	line int
}

// strayCR is the rejection reason for a record with a carriage return still at the end
// of a physical line once its line ending has been removed.
const strayCR = "stray carriage return at the end of a line"

// A CR is stray before another line ending, or at the end of the file with no LF after it.
var strayCRForm = regexp.MustCompile(`\r(\r\n|$)`)

// neverClosed is the file failure for a quoted field still open at the end of its record.
const neverClosed = "a quoted field is never closed"

// readRecord reads the next record: its text and line, and then its cells, the reason
// it is rejected, or the error that fails the file. At the end of the file the error
// is io.EOF.
func readRecord(reader *csv.Reader, data []byte) (record, []string, string, error) {
	start := reader.InputOffset()
	cells, err := reader.Read()
	if err == io.EOF {
		return record{}, nil, "", err
	}
	rec := record{text: string(data[start:reader.InputOffset()])}
	var pe *csv.ParseError
	switch {
	case err == nil:
		rec.line, _ = reader.FieldPos(0)
	case errors.As(err, &pe):
		rec.line = pe.StartLine
	default:
		return rec, nil, "", fmt.Errorf("read: %w", err)
	}
	// encoding/csv reads a line ending in CR CR LF as a cell ending in a carriage return,
	// and one holding nothing else as a one-cell record, so the file's own text is checked.
	if strayCRForm.MatchString(rec.text) {
		return rec, nil, strayCR, nil
	}
	if err == nil {
		return rec, cells, "", nil
	}

	var why string
	switch pe.Err {
	case csv.ErrBareQuote:
		why = "bare quote in an unquoted field"
	case csv.ErrQuote:
		why = "text after a closing quote"
		// encoding/csv reports a quoted field left open at the end of the file with the
		// same error. Only then does closing the quote make the record readable.
		if _, err := csv.NewReader(strings.NewReader(rec.text + `"`)).Read(); err == nil {
			return rec, nil, "", &fileError{line: rec.line, message: neverClosed}
		}
	default:
		// Field counts are not checked by the reader, so it has no other error to give;
		// one that appears anyway fails the file.
		return rec, nil, "", &fileError{line: rec.line, message: "unexpected CSV error: " + err.Error()}
	}
	// A quote error confined to one physical line rejects its record; in a record
	// spanning lines it fails the file.
	if pe.Line != pe.StartLine {
		return rec, nil, "", &fileError{line: rec.line, message: why}
	}
	// encoding/csv stops at the first quote error, so the line itself says whether a
	// later field on it opens a quote. If the rest of the file never closes it, the
	// quoted field is never closed; if it does, the record spans lines with its error.
	// The text starts where the read began, so blank lines the reader skipped open it;
	// their line endings would put the record's first quote after a field's start.
	text := strings.TrimLeft(rec.text, "\r\n")
	if stillQuoted(text) {
		if stillQuoted(strings.TrimLeft(string(data[start:]), "\r\n")) {
			return rec, nil, "", &fileError{line: rec.line, message: neverClosed}
		}
		return rec, nil, "", &fileError{line: rec.line, message: why}
	}
	return rec, nil, why, nil
}

// stillQuoted reports whether a quoted field is open at the end of one record's text.
func stillQuoted(text string) bool {
	open, fieldStart := false, true
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case open && c == '"' && i+1 < len(text) && text[i+1] == '"':
			i++
		case open && c == '"':
			open = false
		case open:
		case c == ',':
			fieldStart = true
			continue
		case c == '"' && fieldStart:
			open = true
		}
		fieldStart = false
	}
	return open
}

// parseHeader maps each column to its position in the file.
func parseHeader(names []string, line int) (map[string]int, error) {
	pos := map[string]int{}
	var problems []string
	for i, name := range names {
		switch {
		case !slices.Contains(Columns, Column(name)):
			problems = append(problems, fmt.Sprintf("unexpected column %q", name))
		case hasKey(pos, name):
			problems = append(problems, fmt.Sprintf("duplicate column %q", name))
		default:
			pos[name] = i
		}
	}
	for _, c := range Columns {
		if !hasKey(pos, string(c)) {
			problems = append(problems, fmt.Sprintf("missing column %q", c))
		}
	}
	if len(problems) > 0 {
		return nil, &fileError{line: line, message: strings.Join(problems, "; "), header: true}
	}
	return pos, nil
}

func hasKey(m map[string]int, k string) bool { _, ok := m[k]; return ok }

// markCollisions flags every reading whose nonblank acquisition ID equals
// another's in the file.
func markCollisions(rs []Reading) {
	count := map[string]int{}
	for _, r := range rs {
		if !isBlank(r.AcquisitionID) {
			count[r.AcquisitionID]++
		}
	}
	for i := range rs {
		if count[rs[i].AcquisitionID] > 1 {
			rs[i].Issues = append(rs[i].Issues, Issue{Field: "id", Code: IssueIDCollision, SourceText: rs[i].AcquisitionID})
		}
	}
}

// numberGroups gives each reading with four electrode IDs its repeat group, one per
// distinct geometry of the file, numbered in order of first appearance; a reading
// with a blank slot gets none. The IDs are compared as given.
func numberGroups(rs []Reading) []Group {
	var groups []Group
	index := map[Group]int{}
	for i := range rs {
		var g Geometry
		complete := true
		for slot, e := range rs[i].Electrodes {
			if e.ID == nil {
				complete = false
				break
			}
			g[slot] = *e.ID
		}
		if !complete {
			continue
		}
		key := Group{Line: rs[i].LineID, Geometry: g}
		n, seen := index[key]
		if !seen {
			n = len(groups)
			groups = append(groups, key)
			index[key] = n
		}
		rs[i].RepeatGroup = &n
	}
	return groups
}
