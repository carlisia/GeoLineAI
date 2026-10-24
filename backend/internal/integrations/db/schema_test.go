package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/survey"
)

// A database from another schema must be refused before anything reads or
// writes it, so its contents survive for the operator to recreate.
func TestDifferentSchemaIsRefusedAndLeftUnchanged(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "geolineai.db")
	d, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("empty database: %v", err)
	}
	_ = d.Close()
	if d, err = Open(ctx, path); err != nil {
		t.Fatalf("matching database refused: %v", err)
	}
	_ = d.Close()

	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	if _, err := raw.Exec(`ALTER TABLE imports ADD COLUMN note TEXT; INSERT INTO imports (source_path, imported_at, retained_count, rejected_count, note) VALUES ('x', 'y', 0, 0, 'kept')`); err != nil {
		t.Fatal(err)
	}

	// The operator reads this text; it must say to recreate and reimport.
	_, err = Open(ctx, path)
	require.EqualError(t, err, "the database schema differs from schema.sql; move it aside with scripts/recreate-db.sh, or set GEOLINEAI_DB_PATH to a new file, then run scripts/setup.sh")
	var note string
	require.NoError(t, raw.QueryRow(`SELECT note FROM imports`).Scan(&note), "contents after refusal")
	assert.Equal(t, "kept", note, "contents after refusal")
}

// Each closed value set lives in Go; the schema's CHECK list must not drift
// from it.
func TestCheckListsMatchGo(t *testing.T) {
	for column, c := range map[string]struct {
		want  []string
		times int // how many tables list the column
	}{
		"qc_review":        {review.Verdicts.Names(), 1},
		"advisory":         {criterion.Advisories.Names(), 1},
		"criterion":        {criterion.Criteria.Names(), 1},
		"state":            {criterion.ResultStates.Names(), 1},
		"unit":             {criterion.Units.Names(), 1},
		"previous_verdict": {review.Verdicts.Names(), 1},
		"code":             {source.IssueCodes.Names(), 1},
		"line_id":          {survey.Lines.Names(), 2},
		"array_type":       {[]string{survey.ArrayDipoleDipole}, 1},
		"field":            {slices.DeleteFunc(source.Columns.Names(), func(c string) bool { return c == "line_id" || c == "qc_review" }), 1},
	} {
		// Every table that lists the column, or one of them could drift unseen.
		lists := regexp.MustCompile(column+` IN \(([^)]*)\)`).FindAllStringSubmatch(schema, -1)
		if !assert.Len(t, lists, c.times, "CHECK lists for %s", column) {
			continue
		}
		for _, m := range lists {
			var got []string
			for _, v := range strings.Split(m[1], ",") {
				got = append(got, strings.Trim(strings.TrimSpace(v), "'"))
			}
			assert.Equal(t, c.want, got, "%s CHECK list against Go", column)
		}
	}
}

// The review record is append-only and the database enforces it; a trigger dropped
// or mistyped would let an action or an entry be rewritten with nothing else
// noticing.
func TestReviewRecordRefusesUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "geolineai.db")
	d, err := Open(ctx, path)
	require.NoError(t, err)
	_ = d.Close()
	raw, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	defer func() { _ = raw.Close() }()
	_, err = raw.Exec(`INSERT INTO review_actions (id, reviewer, recorded_at) VALUES (1, 'r', 't');
		INSERT INTO review_history (action_id, reading_id, previous_verdict) VALUES (1, 1, 'pass')`)
	require.NoError(t, err)

	for _, c := range []struct{ statement, refusal string }{
		{`UPDATE review_actions SET reviewer = 'x' WHERE id = 1`, "review actions are never rewritten"},
		{`DELETE FROM review_actions WHERE id = 1`, "review actions are never rewritten"},
		{`UPDATE review_history SET previous_verdict = 'fail' WHERE action_id = 1`, "review history is never rewritten"},
		{`DELETE FROM review_history WHERE action_id = 1`, "review history is never rewritten"},
	} {
		_, err := raw.Exec(c.statement)
		assert.ErrorContains(t, err, c.refusal, c.statement)
	}
	var actions, entries int
	require.NoError(t, raw.QueryRow(`SELECT (SELECT COUNT(*) FROM review_actions), (SELECT COUNT(*) FROM review_history)`).Scan(&actions, &entries))
	assert.Equal(t, [2]int{1, 1}, [2]int{actions, entries}, "the record after every refusal")
}

// Each schema limit listed here is one the import enforces in Go; the two must
// not drift.
func TestCheckLimitsMatchGo(t *testing.T) {
	positiveInt := strconv.Itoa(source.MinPositiveInt)
	reportedError := strconv.FormatFloat(source.MinReportedError, 'f', -1, 64)
	for _, c := range []struct{ column, want string }{
		{"symmetry", positiveInt},
		{"stacks", positiveInt},
		{"apparent_resistivity_err", reportedError},
		{"chargeability_err", reportedError},
	} {
		m := regexp.MustCompile(`CHECK \(` + c.column + ` >= ([0-9.]+)\)`).FindStringSubmatch(schema)
		if assert.NotNil(t, m, "no CHECK limit for %s", c.column) {
			assert.Equal(t, c.want, m[1], "%s: schema limit against Go", c.column)
		}
	}
}
