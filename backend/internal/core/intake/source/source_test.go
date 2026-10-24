package source_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/intake/sourcetest"
)

func parse(t *testing.T, lines ...string) (source.File, error) {
	t.Helper()
	return source.Parse(strings.NewReader(strings.Join(lines, "\n") + "\n"))
}

// A cell stored dishonestly would put an invented value in front of the reviewer,
// and a valid one rejected would silently lose a reading.
func TestCellRulesDecideWhatIsStored(t *testing.T) {
	rejected := []struct {
		name, field, cell string
		reject            string // the reason the row is rejected, in full
	}{
		{"digit separator", "apparent_resistivity", "1_0", `apparent_resistivity "1_0" is not a finite number`},
		{"hexadecimal", "input_current", "0x1p-2", `input_current "0x1p-2" is not a finite number`},
		{"overflow", "chargeability", "1e400", `chargeability "1e400" is not a finite number`},
		{"underflow", "chargeability", "1e-400", `chargeability "1e-400" is not a finite number`},
		{"not a number", "tx1_lat", "NaN", `tx1_lat "NaN" is not a finite number`},
		{"negative error", "chargeability_err", "-0.1", `chargeability_err "-0.1" is negative`},
		{"n factor below one", "symmetry", "0", `symmetry "0" is not an integer of at least 1`},
		{"stack count not an integer", "stacks", "1.0", `stacks "1.0" is not an integer of at least 1`},
		{"timestamp with a space", "timestamp", "2026-08-17 14:00:12Z", `timestamp "2026-08-17 14:00:12Z" is not an RFC 3339 date-time`},
		{"ten fractional digits", "timestamp", "2026-08-17T14:00:12.1234567890Z", `timestamp "2026-08-17T14:00:12.1234567890Z" is not an RFC 3339 date-time`},
		{"offset hour past 23", "timestamp", "2026-08-17T14:00:12+24:00", `timestamp "2026-08-17T14:00:12+24:00" is not an RFC 3339 date-time`},
		{"offset minute past 59", "timestamp", "2026-08-17T14:00:12+23:60", `timestamp "2026-08-17T14:00:12+23:60" is not an RFC 3339 date-time`},
		{"UTC year before 0000", "timestamp", "0000-01-01T00:00:00+01:00", `timestamp "0000-01-01T00:00:00+01:00" is not an RFC 3339 date-time`},
		{"UTC year after 9999", "timestamp", "9999-12-31T23:59:59-01:00", `timestamp "9999-12-31T23:59:59-01:00" is not an RFC 3339 date-time`},
		{"electrode of another line", "tx1_id", "line-2-e01", `tx1_id "line-2-e01" is not an electrode of line-1`},
		{"electrode index zero", "tx1_id", "line-1-e00", `tx1_id "line-1-e00" is not an electrode of line-1`},
		{"electrode index past 40", "rx2_id", "line-1-e41", `rx2_id "line-1-e41" is not an electrode of line-1`},
		{"electrode untrimmed", "tx2_id", " line-1-e02", `tx2_id " line-1-e02" is not an electrode of line-1`},
		{"array type untrimmed", "array_type", " dipole-dipole", `array_type " dipole-dipole" is not dipole-dipole`},
		{"nine decay samples", "decay_curve", "9;8;7;6;5;4;3;2;1", `decay_curve "9;8;7;6;5;4;3;2;1" is not 10 finite numbers`},
		{"eleven decay samples", "decay_curve", "11;10;9;8;7;6;5;4;3;2;1", `decay_curve "11;10;9;8;7;6;5;4;3;2;1" is not 10 finite numbers`},
		{"tenth decay token not a number", "decay_curve", "10;9;8;7;6;5;4;3;2;x", `decay_curve "10;9;8;7;6;5;4;3;2;x" is not 10 finite numbers`},
		{"eleventh decay token not a number", "decay_curve", "10;9;8;7;6;5;4;3;2;1;x", `decay_curve "10;9;8;7;6;5;4;3;2;1;x" is not 10 finite numbers`},
		{"cell text named as written", "apparent_resistivity", `"1""2"`, `apparent_resistivity "1"2" is not a finite number`},
	}
	for _, c := range rejected {
		t.Run(c.name, func(t *testing.T) {
			f, err := parse(t, sourcetest.Header, sourcetest.Row(map[string]string{c.field: c.cell}))
			require.NoError(t, err)
			assert.Equal(t, []source.Rejection{{Row: 2, Reason: c.reject}}, f.Rejections)
		})
	}

	retained := []struct {
		name, field, cell string
		stored            string // what the reading holds for the field, as text
	}{
		{"implausible kept", "apparent_resistivity", "-5.2", "-5.2"},
		{"trimmed number", "apparent_resistivity", " 3.5 ", "3.5"},
		{"exponent", "apparent_resistivity", "1e-3", "0.001"},
		{"zero error kept", "chargeability_err", "0", "0"},
		{"offset timestamp", "timestamp", "2026-08-17T16:00:12+02:00", "2026-08-17T14:00:12.000000000Z"},
		{"first UTC year", "timestamp", "0000-01-01T00:00:00Z", "0000-01-01T00:00:00.000000000Z"},
		{"last UTC year", "timestamp", "9999-12-31T23:59:59Z", "9999-12-31T23:59:59.000000000Z"},
		{"source verdict ignored", "qc_review", "pass", ""},
	}
	for _, c := range retained {
		t.Run(c.name, func(t *testing.T) {
			f, err := parse(t, sourcetest.Header, sourcetest.Row(map[string]string{c.field: c.cell}))
			require.NoError(t, err)
			require.Len(t, f.Readings, 1, "rejected %v, want retained", f.Rejections)
			assert.Equal(t, c.stored, stored(f.Readings[0], c.field))
		})
	}
}

// rowNumbers lists the source row of each retained reading, in order.
func rowNumbers(readings []source.Reading) []int {
	var rows []int
	for _, r := range readings {
		rows = append(rows, r.RowNumber)
	}
	return rows
}

// stored is what a retained reading holds for field, as text. A field the retained
// cases do not use is a mistake in the test, so it panics.
func stored(r source.Reading, field string) string {
	switch field {
	case "apparent_resistivity":
		return fmt.Sprint(*r.ApparentResistivity)
	case "chargeability_err":
		return fmt.Sprint(*r.ChargeabilityErr)
	case "timestamp":
		return *r.Timestamp
	case "qc_review":
		return "" // ignored: a reading holds nothing from this column
	}
	panic(fmt.Sprintf("stored: no case for field %q", field))
}

// The reviewer fixes a file from its reasons, so one read must name every bad cell.
func TestRejectionNamesEveryInvalidCell(t *testing.T) {
	f, _ := parse(t, sourcetest.Header, sourcetest.Row(map[string]string{"stacks": "x", "decay_curve": "1;2"}))
	want := `stacks "x" is not an integer of at least 1; decay_curve "1;2" is not 10 finite numbers`
	assert.Equal(t, []source.Rejection{{Row: 2, Reason: want}}, f.Rejections)
}

// Row numbers send the reviewer to a line of the file, so they must be the
// physical line a record starts on.
func TestRowNumbersAreThePhysicalStartLine(t *testing.T) {
	quoted := sourcetest.Row(map[string]string{"id": "\"line-1\nr2\""})
	f, err := parse(t, "\xef\xbb\xbf"+sourcetest.Header, "", sourcetest.Row(nil), `bad"quote,`+sourcetest.Row(nil)[1:], quoted, sourcetest.Row(nil), sourcetest.Row(map[string]string{"id": `"x"y`}))
	require.NoError(t, err)
	assert.Equal(t, []int{3, 5, 7}, rowNumbers(f.Readings), "retained rows")
	assert.Equal(t, []source.Rejection{{Row: 4, Reason: "bare quote in an unquoted field"}, {Row: 8, Reason: "text after a closing quote"}}, f.Rejections, "row 4 bare quote and row 8 (last line) text after a closing quote")
}

// encoding/csv drops a carriage return left at the end of a line, which failed the whole
// file on a line holding nothing else and stored a cell without it on any other; the
// record must be rejected alone, and a CRLF line beside it kept.
func TestStrayCarriageReturnRejectsOnlyItsRecord(t *testing.T) {
	row := sourcetest.Row(nil)
	quoted := sourcetest.Row(map[string]string{"id": "\"a\r\r\nb\""})
	f, err := parse(t, sourcetest.Header, row, "\r\r", row+"\r\r", quoted, row+"\r")
	require.NoError(t, err)
	const stray = "stray carriage return at the end of a line"
	assert.Equal(t, []source.Rejection{{Row: 3, Reason: stray}, {Row: 4, Reason: stray}, {Row: 5, Reason: stray}}, f.Rejections)
	assert.Equal(t, []int{2, 7}, rowNumbers(f.Readings), "retained rows, the CRLF line kept")

	// A CR at the very end of the file, with no LF after it, is a stray one too.
	f, err = source.Parse(strings.NewReader(sourcetest.Header + "\n" + row + "\r"))
	require.NoError(t, err)
	assert.Equal(t, []source.Rejection{{Row: 2, Reason: stray}}, f.Rejections)
	assert.Empty(t, f.Readings)
}

// A broken header or quote fails the whole file, naming the line it starts on.
func TestFileFailuresNameTheirLine(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  string // what the import command prints
	}{
		{"quote error across lines", []string{sourcetest.Header, sourcetest.Row(nil), `"a` + "\n" + `b"x,` + sourcetest.Row(nil)}, "record starting on line 3: text after a closing quote"},
		{"header after a blank line", []string{"", sourcetest.Header + ",extra", sourcetest.Row(nil)}, `header (line 2): unexpected column "extra"`},
		{"unterminated quote", []string{sourcetest.Header, sourcetest.Row(nil), `"never closed,` + sourcetest.Row(nil)}, "record starting on line 3: a quoted field is never closed"},
		{"unterminated quote after a quote error", []string{sourcetest.Header, sourcetest.Row(nil), `a"b,"c`, sourcetest.Row(nil)}, "record starting on line 3: a quoted field is never closed"},
		{"quote error, the quote closed on a later line", []string{sourcetest.Header, sourcetest.Row(nil), `a"b,"c`, `d"` + sourcetest.Row(nil)}, "record starting on line 3: bare quote in an unquoted field"},
		// A blank line before the record is skipped by the reader, so the record's text
		// opens with a line ending; it must not read as text before the record's first quote.
		{"unterminated quote after a quote error, blank line before", []string{sourcetest.Header, sourcetest.Row(nil), "", `"a,"b,"c`, sourcetest.Row(nil)}, "record starting on line 4: a quoted field is never closed"},
		{"quote error closed on a later line, blank line before", []string{sourcetest.Header, sourcetest.Row(nil), "", `"a,"b,"c`, `d"` + sourcetest.Row(nil)}, "record starting on line 4: text after a closing quote"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parse(t, c.lines...)
			assert.EqualError(t, err, c.want)
		})
	}
}
