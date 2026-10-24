package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/ids"
	"geolineai/internal/core/intake/importer"
	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/review/readings"
	"geolineai/internal/core/review/verdict"
	"geolineai/internal/core/survey"
)

// Commit stores one file, all of it or none: its import row, its readings,
// each with the file's verdict, and their issues.
func (d DB) Commit(ctx context.Context, f importer.File) (ids.ID, error) {
	tx, err := d.write.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO imports (source_path, imported_at, retained_count, rejected_count) VALUES (?, ?, ?, ?)`,
		f.SourcePath, time.Now().UTC().Format(review.TimeLayout), f.Retained, f.Rejected)
	if err != nil {
		return 0, fmt.Errorf("insert import: %w", err)
	}
	importID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The file's groups first, so each reading can name its group's row.
	groupIDs := make([]int64, len(f.Groups))
	for i, g := range f.Groups {
		args := []any{importID, string(g.Line)}
		for _, id := range g.Geometry {
			args = append(args, id)
		}
		res, err := tx.ExecContext(ctx,
			`INSERT INTO repeat_groups (import_id, line_id, `+strings.Join(slotIDColumns, ", ")+`) VALUES (`+placeholders(len(args))+`)`, args...)
		if err != nil {
			return 0, fmt.Errorf("insert repeat group: %w", err)
		}
		if groupIDs[i], err = res.LastInsertId(); err != nil {
			return 0, err
		}
	}

	insReading, err := tx.PrepareContext(ctx, `INSERT INTO readings (`+strings.Join(readingColumns, ", ")+`)
		VALUES (`+placeholders(len(readingColumns))+`)`)
	if err != nil {
		return 0, err
	}
	insIssue, err := tx.PrepareContext(ctx,
		`INSERT INTO import_issues (reading_id, field, code, source_text) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}

	if len(f.Evaluations) != len(f.Readings) {
		return 0, fmt.Errorf("%d evaluations for %d readings", len(f.Evaluations), len(f.Readings))
	}
	insResult, err := tx.PrepareContext(ctx,
		`INSERT INTO results (reading_id, position, criterion, state, explanation, value, limit_value, unit) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}

	for i, r := range f.Readings {
		var groupID any // NULL for a reading with no group
		if r.RepeatGroup != nil {
			groupID = groupIDs[*r.RepeatGroup]
		}
		args := []any{importID, groupID, r.RowNumber, string(r.LineID), r.AcquisitionID}
		for _, e := range r.Electrodes {
			args = append(args, e.ID, e.Lat, e.Lon, e.Alt)
		}
		decay, err := decayJSON(r.DecayCurve)
		if err != nil {
			return 0, err
		}
		args = append(args, r.Timestamp, r.ArrayType, r.Symmetry, r.Stacks, r.InputCurrent, r.ContactResistance,
			r.ApparentResistivity, r.ApparentResistivityErr, r.Chargeability, r.ChargeabilityErr,
			decay, string(f.Verdict), string(f.Evaluations[i].Value), f.Evaluations[i].LeftOut)
		res, err := insReading.ExecContext(ctx, args...)
		if err != nil {
			return 0, fmt.Errorf("insert reading from row %d: %w", r.RowNumber, err)
		}
		readingID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		for _, is := range r.Issues {
			if _, err := insIssue.ExecContext(ctx, readingID, is.Field, is.Code, is.SourceText); err != nil {
				return 0, fmt.Errorf("insert issue from row %d: %w", r.RowNumber, err)
			}
		}
		for position, res := range f.Evaluations[i].Results {
			if _, err := insResult.ExecContext(ctx, readingID, position, string(res.Criterion), string(res.State), res.Explanation, res.Value, res.Limit, string(res.Unit)); err != nil {
				return 0, fmt.Errorf("insert result from row %d: %w", r.RowNumber, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return ids.ID(importID), nil
}

// slotIDColumns are the repeat_groups table's electrode ID columns, in slot order,
// the order the import's geometry array holds the IDs in.
var slotIDColumns = func() []string {
	var out []string
	for _, s := range source.Slots {
		out = append(out, string(s.Columns().ID))
	}
	return out
}()

// electrodeColumns are the readings table's electrode columns: ID, latitude,
// longitude and altitude for each slot, in slot order, which is the same order a
// reading holds its electrodes in.
var electrodeColumns = func() []string {
	var out []string
	for _, s := range source.Slots {
		c := s.Columns()
		out = append(out, string(c.ID), string(c.Lat), string(c.Lon), string(c.Alt))
	}
	return out
}()

// readingColumns are the columns Commit writes, in the order it binds them.
var readingColumns = slices.Concat(
	[]string{"import_id", "repeat_group_id", "row_number", "line_id", "acquisition_id"},
	electrodeColumns,
	[]string{"timestamp", "array_type", "symmetry", "stacks", "input_current", "contact_resistance",
		"apparent_resistivity", "apparent_resistivity_err", "chargeability", "chargeability_err",
		"decay_curve", "qc_review", "advisory", "left_out"},
)

// resultOfReading is the one result of reading r for the criterion bound to its
// placeholder, which every subselect over a reading's results is built on.
const resultOfReading = "FROM results x WHERE x.reading_id = r.id AND x.criterion = ?"

// repeatColumns select a reading's recorded repeat difference, the value of its
// repeat agreement result, NULL when it has none, and whether that result's state is
// the disagreeing one, 0 when it has none. The columns bind repeatArgs, in order, and
// the statement's own arguments follow them.
const repeatColumns = "(SELECT x.value " + resultOfReading + "), " +
	"COALESCE((SELECT x.state = ? " + resultOfReading + "), 0)"

func repeatArgs(args ...any) []any {
	return append([]any{string(criterion.RepeatAgreement), string(criterion.RepeatDisagreementState), string(criterion.RepeatAgreement)}, args...)
}

func decayJSON(samples []float64) (any, error) {
	if samples == nil {
		return nil, nil
	}
	b, err := json.Marshal(samples)
	return string(b), err
}

// ImportCount counts the imports the database holds, through the read pool, so a
// health check proves the database is open and says whether anything is in it.
func (d DB) ImportCount(ctx context.Context) (int, error) {
	var n int
	err := d.read.QueryRowContext(ctx, `SELECT COUNT(*) FROM imports`).Scan(&n)
	return n, err
}

// ReadingsAt finds the readings imported from a file at a row: the file is named by
// its base name, and an import records its absolute path, so the base name is compared
// here rather than in SQL, where LIKE would read an underscore as a wildcard.
func (d DB) ReadingsAt(ctx context.Context, file string, row int) ([]ids.ID, error) {
	rows, err := d.read.QueryContext(ctx, `
		SELECT r.id, i.source_path FROM readings r JOIN imports i ON i.id = r.import_id
		WHERE r.row_number = ? ORDER BY r.id`, row)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []ids.ID
	for rows.Next() {
		var id ids.ID
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		if filepath.Base(path) == file {
			out = append(out, id)
		}
	}
	return out, rows.Err()
}

// Detail reads one reading and its issues as they stood at one moment.
func (d DB) Detail(ctx context.Context, id ids.ID) (readings.Stored, error) {
	tx, err := d.read.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return readings.Stored{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var r readings.Stored
	var line, v, a string
	var decay sql.NullString
	e := make([]readings.Electrode, len(source.Slots))
	dest := []any{&r.ReadingID, &r.AcquisitionID, &line, &r.SourcePath, &r.RowNumber, &r.RepeatGroupID}
	for i, slot := range source.Slots {
		e[i].Slot = slot
		dest = append(dest, &e[i].ElectrodeID, &e[i].Lat, &e[i].Lon, &e[i].Alt)
	}
	dest = append(dest, &r.Timestamp, &r.ArrayType, &r.Symmetry, &r.Stacks, &r.InputCurrent, &r.ContactResistance,
		&r.ApparentResistivity, &r.ApparentResistivityErr, &r.Chargeability, &r.ChargeabilityErr, &decay, &v, &a,
		&r.RepeatDifference, &r.Disagrees)
	err = tx.QueryRowContext(ctx, `
		SELECT r.id, r.acquisition_id, r.line_id, i.source_path, r.row_number, r.repeat_group_id,
		       r.`+strings.Join(electrodeColumns, ", r.")+`,
		       r.timestamp, r.array_type, r.symmetry, r.stacks, r.input_current, r.contact_resistance,
		       r.apparent_resistivity, r.apparent_resistivity_err, r.chargeability, r.chargeability_err,
		       r.decay_curve, r.qc_review, r.advisory, `+repeatColumns+`
		FROM readings r JOIN imports i ON i.id = r.import_id
		WHERE r.id = ?`, repeatArgs(id)...).Scan(dest...)
	if errors.Is(err, sql.ErrNoRows) {
		return readings.Stored{}, readings.ErrNotFound
	}
	if err != nil {
		return readings.Stored{}, err
	}
	r.Electrodes = e
	if r.LineID, err = lineOf(line); err != nil {
		return readings.Stored{}, err
	}
	if r.Verdict, err = verdictOf(v); err != nil {
		return readings.Stored{}, err
	}
	if r.Advisory.Value, err = advisoryOf(a); err != nil {
		return readings.Stored{}, err
	}
	if r.Advisory.Results, err = resultsOf(ctx, tx, id); err != nil {
		return readings.Stored{}, err
	}
	if decay.Valid {
		if err := json.Unmarshal([]byte(decay.String), &r.Decay.Samples); err != nil {
			return readings.Stored{}, fmt.Errorf("decode decay of reading %s: %w", id, err)
		}
	}

	issues, err := tx.QueryContext(ctx, `SELECT field, code, source_text FROM import_issues WHERE reading_id = ?`, id)
	if err != nil {
		return readings.Stored{}, err
	}
	defer func() { _ = issues.Close() }()
	for issues.Next() {
		var is source.Issue
		if err := issues.Scan(&is.Field, &is.Code, &is.SourceText); err != nil {
			return readings.Stored{}, err
		}
		r.ImportIssues = append(r.ImportIssues, is)
	}
	if err := issues.Err(); err != nil {
		return readings.Stored{}, err
	}
	if r.RepeatGroupID != nil {
		if r.Group, err = groupMembers(ctx, tx, *r.RepeatGroupID); err != nil {
			return readings.Stored{}, err
		}
	}
	if r.History, err = historyOf(ctx, tx, id); err != nil {
		return readings.Stored{}, err
	}
	return r, tx.Commit()
}

// historyOf reads a reading's history entries oldest first, each with its action's
// reviewer, time and entry count.
func historyOf(ctx context.Context, tx *sql.Tx, reading ids.ID) ([]readings.StoredEntry, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT h.action_id, a.reviewer, a.recorded_at, h.previous_verdict, h.reason,
		       (SELECT COUNT(*) FROM review_history n WHERE n.action_id = h.action_id)
		FROM review_history h JOIN review_actions a ON a.id = h.action_id
		WHERE h.reading_id = ?
		ORDER BY h.action_id`, int64(reading))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []readings.StoredEntry
	for rows.Next() {
		var e readings.StoredEntry
		var v string
		if err := rows.Scan(&e.ActionID, &e.Reviewer, &e.RecordedAt, &v, &e.Reason, &e.EntryCount); err != nil {
			return nil, err
		}
		if e.PreviousVerdict, err = verdictOf(v); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// groupMembers reads a repeat group's readings in reading order, each with the repeat
// difference, the agreement and the left-out flag its import recorded.
func groupMembers(ctx context.Context, tx *sql.Tx, group ids.ID) ([]readings.Repeat, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.id, r.acquisition_id, r.timestamp, r.apparent_resistivity, r.chargeability, r.qc_review, `+repeatColumns+`, r.left_out
		FROM readings r WHERE r.repeat_group_id = ?
		ORDER BY r.timestamp IS NULL, r.timestamp, r.id`, repeatArgs(int64(group))...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []readings.Repeat
	for rows.Next() {
		var m readings.Repeat
		var v string
		if err := rows.Scan(&m.ReadingID, &m.AcquisitionID, &m.Timestamp, &m.ApparentResistivity, &m.Chargeability, &v, &m.RepeatDifference, &m.Disagrees, &m.LeftOut); err != nil {
			return nil, err
		}
		if m.Verdict, err = verdictOf(v); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ApplyVerdict sets one reading's verdict and records the change in one write
// transaction, which takes the write lock when it begins, so the read of the current
// verdict and the writes that follow it see the same state.
func (d DB) ApplyVerdict(ctx context.Context, id ids.ID, v review.Verdict, by verdict.Stamp, reason func(criterion.Evaluation) *string) (verdict.Applied, error) {
	tx, err := d.write.BeginTx(ctx, nil)
	if err != nil {
		return verdict.Applied{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var current, advisory string
	err = tx.QueryRowContext(ctx, `SELECT qc_review, advisory FROM readings WHERE id = ?`, int64(id)).Scan(&current, &advisory)
	if errors.Is(err, sql.ErrNoRows) {
		return verdict.Applied{}, nil
	}
	if err != nil {
		return verdict.Applied{}, err
	}
	was, err := verdictOf(current)
	if err != nil {
		return verdict.Applied{}, err
	}
	if was == v {
		return verdict.Applied{Found: true}, nil
	}
	e := criterion.Evaluation{}
	if e.Value, err = advisoryOf(advisory); err != nil {
		return verdict.Applied{}, err
	}
	if e.Results, err = resultsOf(ctx, tx, id); err != nil {
		return verdict.Applied{}, err
	}

	// The action and its history entry are written before the verdict they explain.
	res, err := tx.ExecContext(ctx, `INSERT INTO review_actions (reviewer, recorded_at) VALUES (?, ?)`, by.Reviewer, by.RecordedAt)
	if err != nil {
		return verdict.Applied{}, err
	}
	action, err := res.LastInsertId()
	if err != nil {
		return verdict.Applied{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO review_history (action_id, reading_id, previous_verdict, reason) VALUES (?, ?, ?, ?)`,
		action, int64(id), string(was), reason(e)); err != nil {
		return verdict.Applied{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE readings SET qc_review = ? WHERE id = ?`, string(v), int64(id)); err != nil {
		return verdict.Applied{}, err
	}
	if err := tx.Commit(); err != nil {
		return verdict.Applied{}, err
	}
	actionID := ids.ID(action)
	return verdict.Applied{Found: true, ActionID: &actionID}, nil
}

// resultsOf reads a reading's criterion results in the order they were recorded.
func resultsOf(ctx context.Context, tx *sql.Tx, reading ids.ID) ([]criterion.Result, error) {
	rows, err := tx.QueryContext(ctx, `SELECT criterion, state, explanation, value, limit_value, unit FROM results WHERE reading_id = ? ORDER BY position`, int64(reading))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []criterion.Result
	for rows.Next() {
		var r criterion.Result
		var name, state, unit string
		if err := rows.Scan(&name, &state, &r.Explanation, &r.Value, &r.Limit, &unit); err != nil {
			return nil, err
		}
		// All three rely on the schema holding the columns to the criteria, the states
		// and the units.
		if r.Criterion, err = criterion.Criteria.ParseStored(name, "criterion"); err != nil {
			return nil, err
		}
		if r.State, err = criterion.ResultStates.ParseStored(state, "result state"); err != nil {
			return nil, err
		}
		if r.Unit, err = criterion.Units.ParseStored(unit, "unit"); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// advisoryOf reads a stored advisory. It relies on the schema holding the column to
// the advisory values, so any other value is a fault.
func advisoryOf(s string) (criterion.Advisory, error) {
	return criterion.Advisories.ParseStored(s, "advisory value")
}

// verdictOf reads a stored verdict. It relies on the schema holding the column to
// the verdicts, so any other value is a fault, never a verdict.
func verdictOf(s string) (review.Verdict, error) { return review.Verdicts.ParseStored(s, "verdict") }

// lineOf reads a stored line. It relies on the schema holding the column to the
// lines, so any other value is a fault, never a line.
func lineOf(s string) (survey.Line, error) { return survey.Lines.ParseStored(s, "survey line") }

func placeholders(n int) string { return strings.TrimSuffix(strings.Repeat("?, ", n), ", ") }

// FailingContacts reads the transmitting pair of every reading of a line whose
// Transmitter contact result is unresolved.
func (d DB) FailingContacts(ctx context.Context, line survey.Line) ([]assist.FailingContact, error) {
	args := []any{string(line), string(criterion.TransmitterContact)}
	for _, s := range criterion.UnresolvedStates {
		args = append(args, string(s))
	}
	rows, err := d.read.QueryContext(ctx, `
		SELECT r.`+string(source.SlotTx1.Columns().ID)+`, r.`+string(source.SlotTx2.Columns().ID)+` FROM readings r
		WHERE r.line_id = ? AND EXISTS (
			SELECT 1 FROM results x WHERE x.reading_id = r.id AND x.criterion = ? AND x.state IN (`+placeholders(len(criterion.UnresolvedStates))+`))
		ORDER BY r.id`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []assist.FailingContact
	for rows.Next() {
		var c assist.FailingContact
		if err := rows.Scan(&c.Tx1, &c.Tx2); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
