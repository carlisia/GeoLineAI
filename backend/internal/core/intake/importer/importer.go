// Package importer is the import use case: it parses one CSV and hands the
// retained readings to storage, which keeps the whole file or none of it.
package importer

import (
	"context"
	"io"

	"geolineai/internal/core/ids"
	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
)

// Counts are how many of one file's rows were retained and how many rejected. Import
// counts them once, so the stored import row and the printed summary cannot disagree.
type Counts struct {
	Retained int
	Rejected int
}

// File is everything one import commits.
type File struct {
	SourcePath string // absolute and resolved, as the importing environment saw it
	Readings   []source.Reading
	Groups     []source.Group // the readings' repeat groups, which they index
	// Evaluations are the readings' advisories with their results, one per reading in
	// the same order, decided at import and never changed afterwards.
	Evaluations []criterion.Evaluation
	Verdict     review.Verdict // every reading's starting verdict
	Counts
}

// Committer stores one file's import row, readings, issues, groups and results, all
// of them or none, and returns the new import's ID.
type Committer interface {
	Commit(ctx context.Context, f File) (ids.ID, error)
}

// Result is what one file's import did. Rejections are diagnostics even when
// the file fails.
type Result struct {
	ImportID   ids.ID
	Counts     // zero when the file failed
	Rejections []source.Rejection
}

// Import parses r and commits it. A file that fails commits nothing.
func Import(ctx context.Context, c Committer, sourcePath string, r io.Reader) (Result, error) {
	parsed, err := source.Parse(r)
	if err != nil {
		return Result{Rejections: parsed.Rejections}, err
	}
	counts := Counts{Retained: len(parsed.Readings), Rejected: len(parsed.Rejections)}
	id, err := c.Commit(ctx, File{SourcePath: sourcePath, Readings: parsed.Readings, Groups: parsed.Groups, Evaluations: criterion.Evaluate(&parsed), Counts: counts, Verdict: review.Unreviewed})
	if err != nil {
		return Result{Rejections: parsed.Rejections}, err
	}
	return Result{ImportID: id, Counts: counts, Rejections: parsed.Rejections}, nil
}
