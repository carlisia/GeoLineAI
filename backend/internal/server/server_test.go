package server_test

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/intake/sourcetest"
	"geolineai/internal/core/scenarios"
	"geolineai/internal/server"
	"geolineai/internal/testutil"
)

// fixture imports line-1 rows out of time order, one without a time, one with a
// whitespace-only value and two sharing an acquisition ID.
func fixture(t *testing.T) *testutil.App {
	a := testutil.NewApp(t)
	a.Import("/data/line-1.csv", sourcetest.CSV(
		sourcetest.Row(map[string]string{"id": "late", "timestamp": "2026-08-17T15:00:00Z"}),
		sourcetest.Row(map[string]string{"id": "untimed", "timestamp": ""}),
		sourcetest.Row(map[string]string{"id": "twin", "timestamp": "2026-08-17T14:00:00Z", "apparent_resistivity": "  "}),
		sourcetest.Row(map[string]string{"id": "twin", "timestamp": "2026-08-17T14:00:00Z"}),
	))
	return a
}

// ids lists the fixture's reading IDs in file order.
func ids(a *testutil.App) []string {
	var out []string
	for _, r := range a.ReadingIDs() {
		out = append(out, r.ID)
	}
	return out
}

// allDetails reads every reading's detail, in file order.
func allDetails(t *testing.T, a *testutil.App) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, id := range ids(a) {
		out = append(out, detailOf(t, a, id))
	}
	return out
}

// detailOf reads one reading's detail, which must exist.
func detailOf(t *testing.T, a *testutil.App, id string) map[string]any {
	t.Helper()
	status, d := a.Do(http.MethodGet, "/api/readings/"+id, "")
	require.Equal(t, http.StatusOK, status, d)
	return d
}

// A row with any one usable context is a reading the reviewer must see, and a row with
// none is rejected; dropping a context would silently lose every reading that has only
// that one.
func TestEachUsableContextAloneKeepsItsRow(t *testing.T) {
	blank := map[string]string{
		"tx1_id": "", "tx2_id": "", "rx1_id": "", "rx2_id": "", "timestamp": "", "apparent_resistivity": "",
		"chargeability": "", "input_current": "", "contact_resistance": "", "decay_curve": "",
	}
	only := func(id, column string) string {
		set := maps.Clone(blank)
		set["id"] = id
		delete(set, column)
		return sourcetest.Row(set)
	}
	a := testutil.NewApp(t)
	a.Import("/data/line-1.csv", sourcetest.CSV(
		only("electrode", "rx2_id"),
		only("timestamp", "timestamp"),
		only("resistivity", "apparent_resistivity"),
		only("chargeability", "chargeability"),
		only("current", "input_current"),
		only("contact", "contact_resistance"),
		only("decay", "decay_curve"),
		only("none", ""),
	))

	var kept []string
	for _, r := range a.ReadingIDs() {
		kept = append(kept, r.AcquisitionID)
	}
	assert.Equal(t, "electrode,timestamp,resistivity,chargeability,current,contact,decay", strings.Join(kept, ","), "each context alone kept, the row with none rejected")
}

// A blank acquisition ID is missing, not repeated, and IDs collide only when exactly
// equal; a false collision would show the reviewer an issue the file does not have.
func TestOnlyEqualNonblankAcquisitionIDsCollide(t *testing.T) {
	a := testutil.NewApp(t)
	a.Import("/data/line-1.csv", sourcetest.CSV(
		sourcetest.Row(map[string]string{"id": ""}),
		sourcetest.Row(map[string]string{"id": ""}),
		sourcetest.Row(map[string]string{"id": "a"}),
		sourcetest.Row(map[string]string{"id": "a "}),
	))
	var counts []string
	for _, id := range ids(a) {
		counts = append(counts, fmt.Sprint(len(detailOf(t, a, id)["import_issues"].([]any))))
	}
	assert.Equal(t, "1,1,0,0", strings.Join(counts, ","), `each blank ID missing only, a and "a " distinct`)
}

// values is a detail without its provenance, issues and verdict: the fields a
// measurement or an electrode lands in.
func values(detail map[string]any) map[string]any {
	out := maps.Clone(detail)
	for _, k := range []string{"reading_id", "acquisition_id", "line_id", "source_filename", "row_number", "timestamp", "import_issues", "verdict",
		"repeat_group_id", "repeat_difference", "disagrees", "repeats", "history", "advisory", "reason", "would_save_reasons"} {
		delete(out, k)
	}
	return out
}

// The page shows provenance and why a value is unavailable; the detail must carry
// both. And every value must come back from its own column: the fixture gives each
// column a distinct value, so two columns swapped between the insert and the read
// show here, where two plausible numbers would not.
func TestDetailCarriesEveryFieldFromItsOwnColumn(t *testing.T) {
	a := fixture(t)
	all := ids(a)
	d := detailOf(t, a, all[0]) // late
	assert.Equal(t, map[string]any{
		"apparent_resistivity": 120.0, "apparent_resistivity_err": 1.2, "chargeability": 0.01, "chargeability_err": 0.001,
		"input_current": 1.5, "contact_resistance": 800.0, "stacks": 4.0, "symmetry": 1.0, "array_type": "dipole-dipole",
		"electrodes": []any{
			map[string]any{"slot": "tx1", "electrode_id": "line-1-e01", "lat": 33.201, "lon": -117.101, "alt": 101.0},
			map[string]any{"slot": "tx2", "electrode_id": "line-1-e02", "lat": 33.202, "lon": -117.102, "alt": 102.0},
			map[string]any{"slot": "rx1", "electrode_id": "line-1-e03", "lat": 33.203, "lon": -117.103, "alt": 103.0},
			map[string]any{"slot": "rx2", "electrode_id": "line-1-e04", "lat": 33.204, "lon": -117.104, "alt": 104.0},
		},
		"decay": map[string]any{"samples": []any{10.0, 9.0, 8.0, 7.0, 6.0, 5.0, 4.0, 3.0, 2.0, 1.0}},
	}, values(d), "every value of the complete reading, from its own column")

	d = detailOf(t, a, all[2]) // the first twin
	assert.Equal(t, "line-1.csv", d["source_filename"])
	assert.Equal(t, 4.0, d["row_number"])
	assert.Equal(t, "unreviewed", d["verdict"])
	assert.Equal(t, []any{
		map[string]any{"code": "id_collision", "field": "id", "source_text": "twin"},
		map[string]any{"code": "missing", "field": "apparent_resistivity", "source_text": "  "},
	}, d["import_issues"], "in source-column order, the blank cell's text kept")
}

// A rejected request must say which field broke which rule, because the message is
// all a caller gets, and must never answer an API path in HTML.
func TestErrorsNameTheirFieldInOneEnvelope(t *testing.T) {
	a := fixture(t)
	cases := []struct {
		method, path, body string
		status             int
		kind, message      string
	}{
		{"GET", "/api/readings/007", "", 400, "validation_error", `id must be a reading ID, a decimal string such as "1".`},
		{"GET", "/api/readings/99999", "", 404, "reading_not_found", "Reading 99999 does not exist."},
		{"GET", "/api/nope", "", 404, "not_found", "No API route at /api/nope."},
		{"GET", "/api/scenarios/", "", 404, "not_found", "No API route at /api/scenarios/."},
		{"GET", "/api", "", 404, "not_found", "No API route at /api."},
		{"POST", "/api/reviews", `[]`, 400, "validation_error", "The body must be a JSON object."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"1"}],"verdict":"pass"} x`, 400, "validation_error", "The body must be a JSON object."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":1,"reason":"x"}]}`, 400, "validation_error",
			`targets[0].reason is not a known field. targets[0].reading_id must be a reading ID, a decimal string such as "1". verdict is required.`},
		{"POST", "/api/reviews", `{"verdict":"pass"}`, 400, "validation_error", "targets is required."},
		{"POST", "/api/reviews", `{"targets":[],"verdict":"maybe"}`, 400, "validation_error",
			"targets must hold exactly one reading; it holds 0. verdict must be pass, fail or unreviewed."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"1"},{"reading_id":"2"}],"verdict":"pass"}`, 400, "validation_error",
			"targets must hold exactly one reading; it holds 2."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"1"},{"reading_id":"1"}],"verdict":"pass"}`, 400, "validation_error",
			"targets must hold exactly one reading; it holds 2."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"007"}],"verdict":"maybe"}`, 400, "validation_error",
			`targets[0].reading_id must be a reading ID, a decimal string such as "1". verdict must be pass, fail or unreviewed.`},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"99999"}],"verdict":"pass"}`, 404, "reading_not_found", "Reading 99999 does not exist."},
		{"POST", "/api/reviews", `{"verdict":"` + strings.Repeat("x", 4<<20) + `"}`, 413, "body_too_large", "The request body exceeds 4 MiB."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":"1","reading_id":"2"}],"verdict":"pass","verdict":"fail"}`, 400, "validation_error",
			"verdict must be given once. targets[0].reading_id must be given once."},
		{"POST", "/api/reviews", `{"targets":[{"reading_id":null}],"verdict":null}`, 400, "validation_error",
			`targets[0].reading_id must be a reading ID, a decimal string such as "1". verdict must be a string.`},
		{"POST", "/api/readings/1/explanations", `{"question":"Why?","extra":1}`, 400, "validation_error", "extra is not a known field."},
		{"POST", "/api/readings/1/explanations", `{"question":"   "}`, 400, "validation_error", "question must be 1 to 300 characters."},
		{"POST", "/api/readings/1/explanations", `{"question":"` + strings.Repeat("x", 301) + `"}`, 400, "validation_error", "question must be 1 to 300 characters."},
		{"POST", "/api/readings/99999/explanations", `{"question":"Why?"}`, 404, "reading_not_found", "Reading 99999 does not exist."},
	}
	for _, c := range cases {
		t.Run(c.method+" "+c.path+" "+c.body[:min(len(c.body), 40)], func(t *testing.T) {
			status, body := a.Do(c.method, c.path, c.body)
			assert.Equal(t, c.status, status)
			assert.Equal(t, c.kind, body["error"])
			assert.Equal(t, c.message, body["message"])
		})
	}
	assert.Zero(t, a.Count("review_actions"), "no refused request wrote an action")
	assert.Zero(t, a.Count("review_history"), "no refused request wrote history")
}

// Every route of the out-of-scope features is gone, answered with the JSON not-found
// envelope, never the page: a hidden route would be a capability nobody reviews.
func TestOutOfScopeRoutesAreNotFound(t *testing.T) {
	a := fixture(t)
	for _, r := range []struct{ method, path, body string }{
		{"GET", "/api/lines", ""},
		{"GET", "/api/readings?line=line-1", ""},
		{"GET", "/api/repeat-groups?line=line-1", ""},
		{"GET", "/api/review-actions/latest", ""},
		{"POST", "/api/review-actions/1/undo", "{}"},
		{"GET", "/api/investigations?line=line-1&by=n_factor&issue=advisory_needs_review", ""},
		{"POST", "/api/investigation-requests", `{"request":"x","current_issue":""}`},
		{"POST", "/api/handover-drafts", `{"line":"line-1"}`},
		{"GET", "/api/exports", ""},
		{"POST", "/api/imports", "{}"},
	} {
		status, body := a.Do(r.method, r.path, r.body)
		assert.Equal(t, http.StatusNotFound, status, r.path)
		assert.Equal(t, "not_found", body["error"], r.path)
	}
}

// A save must land on exactly its reading, report an unchanged save as such, and
// survive the database being reopened.
func TestSaveRoundTripPersists(t *testing.T) {
	a := fixture(t)
	all := ids(a)
	target := all[2]
	status, out := save(a, "fail", target)
	require.Equal(t, http.StatusOK, status, out)
	assert.Equal(t, []any{target}, out["affected_reading_ids"], "first save")
	assert.Equal(t, []any{}, out["unchanged_reading_ids"], "first save")
	status, out = save(a, "fail", target)
	require.Equal(t, http.StatusOK, status, out)
	assert.Equal(t, []any{}, out["affected_reading_ids"], "repeated save, want it unchanged")
	assert.Equal(t, []any{target}, out["unchanged_reading_ids"], "repeated save, want it unchanged")
	assert.Nil(t, out["action_id"], "an unchanged save is no action")
	assert.Equal(t, 1, a.Count("review_actions"))

	a.Reopen()
	var got []string
	for _, d := range allDetails(t, a) {
		got = append(got, d["verdict"].(string))
	}
	assert.Equal(t, "unreviewed,unreviewed,fail,unreviewed", strings.Join(got, ","), "after reopening, only the saved reading failed")
}

// The scenario picker offers the manifest's readings, each resolved from its file and
// row to whatever internal ID the import gave it; a row that matches nothing, or
// matches twice because the file was imported twice, fails rather than offering the
// wrong reading.
func TestScenariosResolveFromFileAndRow(t *testing.T) {
	a := fixture(t)
	a.Manifest = []scenarios.Entry{{Label: "Late", File: "line-1.csv", Row: 2}, {Label: "Twin", File: "line-1.csv", Row: 5}}
	a.Reopen()
	all := ids(a)
	status, out := a.Do(http.MethodGet, "/api/scenarios", "")
	require.Equal(t, http.StatusOK, status, out)
	assert.Equal(t, []any{
		map[string]any{"label": "Late", "reading_id": all[0]},
		map[string]any{"label": "Twin", "reading_id": all[3]},
	}, out["scenarios"])

	a.Manifest = []scenarios.Entry{{Label: "Gone", File: "line-1.csv", Row: 99}}
	a.Reopen()
	status, out = a.Do(http.MethodGet, "/api/scenarios", "")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "The server could not complete the request.", out["message"], "the cause is logged, not sent")

	a.Manifest = []scenarios.Entry{{Label: "Late", File: "line-1.csv", Row: 2}}
	a.Import("/elsewhere/line-1.csv", sourcetest.CSV(sourcetest.Row(nil)))
	a.Reopen()
	status, _ = a.Do(http.MethodGet, "/api/scenarios", "")
	assert.Equal(t, http.StatusInternalServerError, status, "the file imported twice: row 2 is ambiguous")
}

// The built UI is served from the web root: the page at / and at /index.html, its files
// by path, and nothing for a directory or a path with no file. A browser holding an
// old page must get the new one whole, since the old one names scripts a new build no
// longer has, and would show a blank page; a script, named by its hash, may be kept.
func TestStaticFilesServeTheBuiltUI(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("<p>page</p>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("js"), 0o644))
	routes := server.Routes(nil, server.Models{}, server.Config{WebRoot: root})
	for _, c := range []struct {
		path  string
		stale bool // the request claims a copy newer than the file
		code  int
		body  string // checked when set
	}{
		{"/", false, 200, "<p>page</p>"},
		{"/index.html", false, 200, "<p>page</p>"},
		{"/", true, 200, "<p>page</p>"},
		{"/index.html", true, 200, "<p>page</p>"},
		{"/assets/app.js", false, 200, "js"},
		{"/assets/app.js", true, 304, ""},
		{"/assets", false, 404, ""},
		{"/assets/", false, 404, ""},
		{"/nope", false, 404, ""},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		if c.stale {
			req.Header.Set("If-Modified-Since", "Sat, 01 Jan 2050 00:00:00 GMT")
		}
		routes.ServeHTTP(rec, req)
		assert.Equal(t, c.code, rec.Code, "%s stale=%v", c.path, c.stale)
		if c.body != "" {
			assert.Equal(t, c.body, rec.Body.String(), c.path)
		}
		if c.code == 200 && c.body == "<p>page</p>" {
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), "%s is never kept", c.path)
		}
	}
}

// A check reads the status code alone, so a database the server cannot read must not
// answer 200; the fingerprint identifies the model configuration the build runs with.
func TestHealthReportsTheDatabaseAndTheConfiguration(t *testing.T) {
	a := fixture(t)
	health := func() (int, string) {
		rec := httptest.NewRecorder()
		a.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_/health", nil))
		return rec.Code, rec.Body.String()
	}
	status, body := health()
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"error":null,"fingerprint":"sha256:test","imports":1,"status":"ok"}`, body)

	a.CloseDatabase()
	status, body = health()
	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Equal(t, `{"error":"The database could not be read.","fingerprint":"sha256:test","imports":null,"status":"unavailable"}`, body)
}
