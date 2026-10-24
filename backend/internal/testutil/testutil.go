// Package testutil builds what the seam tests drive: the router the server
// runs, over a temporary on-disk database.
package testutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/intake/importer"
	"geolineai/internal/core/scenarios"
	"geolineai/internal/integrations/db"
	"geolineai/internal/server"
)

// App is the router over one database file, which Reopen closes and opens
// again to show what persisted, with fake models a test sets before asking, or none
// when Models is nil at the last open, and the scenario manifest in Manifest.
type App struct {
	t        *testing.T
	path     string
	store    db.DB
	router   *gin.Engine
	Models   *Models
	Manifest []scenarios.Entry
}

// Models stands in for both models: scripted answers, for tests only. The writer
// returns Written, raw, so a test can script malformed JSON; the judge answers each
// noul with Nouls[name]. Unavailable makes every call fail as a provider that does not
// answer, and Invalid as one whose answer cannot be read.
type Models struct {
	Written     string
	Nouls       map[string]float64
	Unavailable bool
	Invalid     bool
}

func (m *Models) Write(_ context.Context, _, _ string, _ map[string]any) (json.RawMessage, string, error) {
	if err := m.failure(); err != nil {
		return nil, "", err
	}
	return json.RawMessage(m.Written), "fake-writer", nil
}

func (m *Models) Judge(_ context.Context, _ any, questions map[string]assist.Question) (map[string]assist.Answer, string, error) {
	if err := m.failure(); err != nil {
		return nil, "", err
	}
	out := map[string]assist.Answer{}
	for name, q := range questions {
		out[name] = assist.Answer{Kind: q.Kind, Noul: m.Nouls[name]}
	}
	return out, "fake-judge", nil
}

func (m *Models) failure() error {
	switch {
	case m.Unavailable:
		return assist.ErrUnavailable
	case m.Invalid:
		return assist.ErrInvalidResponse
	}
	return nil
}

// NewApp opens a fresh database in a temporary directory.
func NewApp(t *testing.T) *App {
	t.Helper()
	gin.SetMode(gin.TestMode)
	a := &App{t: t, path: filepath.Join(t.TempDir(), "geolineai.db"), Models: &Models{}}
	a.open()
	t.Cleanup(func() { _ = a.store.Close() })
	return a
}

func (a *App) open() {
	d, err := db.Open(context.Background(), a.path)
	if err != nil {
		a.t.Fatalf("open database: %v", err)
	}
	a.store = d
	var models server.Models
	if a.Models != nil {
		models = server.Models{Writer: a.Models, Judge: a.Models}
	}
	a.router = server.Routes(d, models, server.Config{Manifest: a.Manifest, WebRoot: a.t.TempDir(), Fingerprint: "sha256:test"})
}

// CloseDatabase closes the database under the running router, so a test can see what
// the server answers when it cannot read it.
func (a *App) CloseDatabase() {
	a.t.Helper()
	if err := a.store.Close(); err != nil {
		a.t.Fatalf("close database: %v", err)
	}
}

// Reopen closes the database and opens the same file again.
func (a *App) Reopen() {
	a.t.Helper()
	if err := a.store.Close(); err != nil {
		a.t.Fatalf("close database: %v", err)
	}
	a.open()
}

// Import imports one CSV text as the file sourcePath.
func (a *App) Import(sourcePath string, csv string) {
	a.t.Helper()
	if _, err := importer.Import(context.Background(), a.store, sourcePath, strings.NewReader(csv)); err != nil {
		a.t.Fatalf("import %s: %v", sourcePath, err)
	}
}

// ReadingIDs lists every reading's internal ID with its acquisition ID, in the order
// the file held them, read straight from the database file: the API no longer lists
// readings, and a test must not assume the IDs the database hands out.
func (a *App) ReadingIDs() []Reading {
	a.t.Helper()
	raw := a.readOnly()
	defer func() { _ = raw.Close() }()
	rows, err := raw.Query(`SELECT id, acquisition_id FROM readings ORDER BY import_id, row_number`)
	if err != nil {
		a.t.Fatalf("list readings: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Reading
	for rows.Next() {
		var r Reading
		if err := rows.Scan(&r.ID, &r.AcquisitionID); err != nil {
			a.t.Fatalf("list readings: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		a.t.Fatalf("list readings: %v", err)
	}
	return out
}

// Reading is one reading's internal ID, as the API writes it, and acquisition ID.
type Reading struct {
	ID            string
	AcquisitionID string
}

// IDsByAcquisition maps each acquisition ID to its reading's internal ID; a test whose
// rows share an acquisition ID uses ReadingIDs instead.
func (a *App) IDsByAcquisition() map[string]string {
	a.t.Helper()
	out := map[string]string{}
	for _, r := range a.ReadingIDs() {
		if _, dup := out[r.AcquisitionID]; dup {
			a.t.Fatalf("acquisition ID %q is not unique; use ReadingIDs", r.AcquisitionID)
		}
		out[r.AcquisitionID] = r.ID
	}
	return out
}

// Count counts the rows of a table, for a test that checks nothing was written.
func (a *App) Count(table string) int {
	a.t.Helper()
	raw := a.readOnly()
	defer func() { _ = raw.Close() }()
	var n int
	if err := raw.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		a.t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// Exec runs one statement on the database file through a connection of its own, for
// a test that changes how the database behaves under the running app.
func (a *App) Exec(stmt string) {
	a.t.Helper()
	raw, err := sql.Open("sqlite", "file:"+a.path)
	if err != nil {
		a.t.Fatalf("open database: %v", err)
	}
	defer func() { _ = raw.Close() }()
	if _, err := raw.Exec(stmt); err != nil {
		a.t.Fatalf("exec %q: %v", stmt, err)
	}
}

// readOnly opens the database file read-only, beside the app's own connections.
func (a *App) readOnly() *sql.DB {
	a.t.Helper()
	raw, err := sql.Open("sqlite", "file:"+a.path+"?mode=ro")
	if err != nil {
		a.t.Fatalf("open database: %v", err)
	}
	return raw
}

// Handler is the router, for a test that reads a response's raw body.
func (a *App) Handler() http.Handler { return a.router }

// Do sends one request and returns the status and the decoded JSON body.
func (a *App) Do(method, path, body string) (int, map[string]any) {
	a.t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, httptest.NewRequest(method, path, r))
	var out map[string]any
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			a.t.Fatalf("%s %s answered non-JSON %q", method, path, rec.Body.String())
		}
	}
	return rec.Code, out
}
