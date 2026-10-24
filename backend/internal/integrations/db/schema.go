package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"strings"
)

//go:embed schema.sql
var schema string

// Open connects to the file at path and makes sure the database holds exactly the
// schema: an empty database gets it, and any other must already match it.
func Open(ctx context.Context, path string) (DB, error) {
	d, err := connect(ctx, path)
	if err != nil {
		return DB{}, err
	}
	if err := ensureSchema(ctx, d.write); err != nil {
		_ = d.Close()
		return DB{}, err
	}
	return d, nil
}

// errSchemaDiffers says what to do about a database from another schema.
var errSchemaDiffers = fmt.Errorf("the database schema differs from schema.sql; move it aside with scripts/recreate-db.sh, or set GEOLINEAI_DB_PATH to a new file, then run scripts/setup.sh")

func ensureSchema(ctx context.Context, conn *sql.DB) error {
	have, err := objects(ctx, conn)
	if err != nil {
		return err
	}
	if len(have) == 0 {
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
		return tx.Commit()
	}

	ref, err := sql.Open("sqlite", "file::memory:")
	if err != nil {
		return err
	}
	defer func() { _ = ref.Close() }()
	ref.SetMaxOpenConns(1)
	if _, err := ref.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("create reference schema: %w", err)
	}
	want, err := objects(ctx, ref)
	if err != nil {
		return err
	}
	if !maps.Equal(have, want) {
		return errSchemaDiffers
	}
	return nil
}

// objects maps each table, index, view and trigger to its definition with
// comments and whitespace normalized away.
func objects(ctx context.Context, conn *sql.DB) (map[string]string, error) {
	rows, err := conn.QueryContext(ctx,
		`SELECT type, name, COALESCE(sql, '') FROM sqlite_master WHERE name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var typ, name, def string
		if err := rows.Scan(&typ, &name, &def); err != nil {
			return nil, err
		}
		out[typ+" "+name] = normalize(def)
	}
	return out, rows.Err()
}

var (
	lineComment  = regexp.MustCompile(`--[^\n]*`)
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	space        = regexp.MustCompile(`\s+`)
)

func normalize(def string) string {
	def = blockComment.ReplaceAllString(lineComment.ReplaceAllString(def, ""), "")
	return strings.TrimSpace(space.ReplaceAllString(def, " "))
}
