// Package cli is the import command: it opens the files named on the command
// line, imports each through the core, and reports rejections, a summary and
// the exit status.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"geolineai/internal/core/intake/importer"
)

const usage = `usage: geolineai import <csv-path>...

Loads survey readings from CSV into the database at GEOLINEAI_DB_PATH, one
transaction per file; a failed file saves nothing and later files still import.
Importing a file twice stores its readings twice.

	geolineai import data/line-1.csv
`

// Opener opens storage for the import and returns how to close it.
type Opener func(ctx context.Context) (importer.Committer, func() error, error)

// Import runs the command and returns its exit status: 2 for usage, 1 when any
// file failed, 0 when every file imported.
func Import(ctx context.Context, paths []string, open Opener, stdout, stderr io.Writer) int {
	if len(paths) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	c, closeStore, err := open(ctx)
	if err != nil {
		// Nothing imports, so every file is reported as failed with the usual summary.
		for _, path := range paths {
			reportFailed(stdout, stderr, filepath.Base(path), fmt.Errorf("open database: %w", err))
		}
		summarizeFailures(stdout, paths, len(paths))
		return 1
	}
	defer func() {
		if err := closeStore(); err != nil {
			fmt.Fprintf(stderr, "failed to close database: %v\n", err)
		}
	}()

	var failed []string
	for _, path := range paths {
		name := filepath.Base(path)
		res, err := importFile(ctx, c, path)
		for _, r := range res.Rejections {
			fmt.Fprintf(stderr, "%s:%d: rejected: %s\n", name, r.Row, r.Reason)
		}
		if err != nil {
			reportFailed(stdout, stderr, name, err)
			failed = append(failed, path)
			continue
		}
		fmt.Fprintf(stdout, "%s: imported, %d retained, %d rejected (import %s)\n", name, res.Retained, res.Rejected, res.ImportID)
	}

	if len(failed) > 0 {
		summarizeFailures(stdout, failed, len(paths))
		return 1
	}
	fmt.Fprintf(stdout, "%d of %d files imported\n", len(paths), len(paths))
	return 0
}

// reportFailed says why one file failed and that none of it was saved.
func reportFailed(stdout, stderr io.Writer, name string, err error) {
	fmt.Fprintf(stderr, "%s: failed: %v\n", name, err)
	fmt.Fprintf(stdout, "%s: FAILED, nothing saved from this file\n", name)
}

// summarizeFailures names the files that did not import, so only those are retried.
func summarizeFailures(stdout io.Writer, failed []string, total int) {
	fmt.Fprintf(stdout, "%d of %d files not imported; fix and retry only these:\n", len(failed), total)
	for _, p := range failed {
		fmt.Fprintf(stdout, "  %s\n", p)
	}
}

// importFile opens one CSV and records its absolute, resolved path.
func importFile(ctx context.Context, c importer.Committer, path string) (importer.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return importer.Result{}, err // os.Open's error names the path
	}
	defer func() { _ = f.Close() }()
	abs, err := filepath.Abs(path)
	if err == nil {
		abs, err = filepath.EvalSymlinks(abs)
	}
	if err != nil {
		return importer.Result{}, fmt.Errorf("resolve %s: %w", path, err)
	}
	return importer.Import(ctx, c, abs, f)
}
