package server_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gzuidhof/tygo/tygo"

	"geolineai/internal/core/assist/explain"
	"geolineai/internal/core/assist/grounded"
	"geolineai/internal/core/assist/prompts"
	"geolineai/internal/core/closed"
	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/review/verdict"
	"geolineai/internal/core/scenarios"
	"geolineai/internal/core/survey"
	"geolineai/internal/server"
)

var update = flag.Bool("update", false, "rewrite the generated frontend files")

const rerunCommand = "cd backend && go test ./internal/server -run TestFrontend -update"

const rerun = "// Rerun: " + rerunCommand + "\n"

// generatedHeader is the first lines of every generated file: the test that
// writes it and the command that reruns it.
func generatedHeader(test string) string {
	return "// Generated from Go by " + test + " (backend/internal/server/frontend_test.go); do not edit.\n" + rerun
}

// check compares a generated frontend file with what Go produces now, or
// rewrites it under -update.
func check(t *testing.T, path, want string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Errorf("%s is stale; rerun: %s", path, rerunCommand)
	}
}

func tsList(name, typ string, values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("'%s'", v)
	}
	return fmt.Sprintf("export const %s = [%s] as const\nexport type %s = (typeof %s)[number]\n", name, strings.Join(quoted, ", "), typ, name)
}

// The frontend holds no copy of a value set Go owns; it imports this file.
func TestFrontendVocabulary(t *testing.T) {
	check(t, repoPath("frontend", "src", "lib", "vocabulary.gen.ts"),
		generatedHeader("TestFrontendVocabulary")+"\n"+
			tsList("LINES", "LineId", survey.Lines.Names())+"\n"+
			tsList("VERDICTS", "Verdict", review.Verdicts.Names())+"\n"+
			tsList("ISSUE_CODES", "IssueCode", source.IssueCodes.Names())+"\n"+
			tsList("COLUMNS", "Column", source.Columns.Names())+"\n"+
			tsList("SLOTS", "Slot", source.SlotNames())+"\n"+
			tsList("ERROR_CODES", "ErrorCode", server.ErrorCodes.Names())+"\n"+
			tsList("ADVISORIES", "Advisory", criterion.Advisories.Names())+"\n"+
			tsList("CRITERIA", "Criterion", criterion.Criteria.Names())+"\n"+
			tsList("RESULT_STATES", "ResultState", criterion.ResultStates.Names())+"\n"+
			tsList("UNITS", "Unit", criterion.Units.Names())+"\n"+
			tsRecord("ISSUE_LABELS", "IssueCode", source.IssueCodes, criterion.IssueLabels)+"\n"+
			tsRecord("CRITERION_LABELS", "Criterion", criterion.Criteria, criterion.Labels)+"\n"+
			tsRecord("RESULT_STATE_LABELS", "ResultState", criterion.ResultStates, criterion.ResultStateLabels)+"\n"+
			tsRecord("ADVISORY_LABELS", "Advisory", criterion.Advisories, criterion.AdvisoryLabels)+"\n"+
			tsRecord("ADVISORY_MEANINGS", "Advisory", criterion.Advisories, criterion.AdvisoryMeanings)+"\n"+
			tsRecord("VERDICT_LABELS", "Verdict", review.Verdicts, review.VerdictLabels)+"\n"+
			tsList("RESULT_STATE_ORDER", "ResultStateInOrder", resultStateOrder())+"\n"+
			tsList("SOURCE_KINDS", "SourceKind", grounded.SourceKinds.Names())+"\n"+
			tsList("OUTCOMES", "Outcome", grounded.Outcomes.Names())+"\n"+
			tsList("UNAVAILABLE_REASONS", "UnavailableReason", grounded.UnavailableReasons.Names())+"\n"+
			fmt.Sprintf("export const MAX_QUESTION = %d\n", explain.MaxQuestion)+"\n"+
			recordsReason())
}

// resultStateOrder is every result state in the order the page lists results: the
// unresolved states as the criteria rank them, then the rest.
func resultStateOrder() []string {
	var out []string
	for _, s := range criterion.UnresolvedStates {
		out = append(out, string(s))
	}
	for _, s := range criterion.ResultStates {
		if !slices.Contains(criterion.UnresolvedStates, s) {
			out = append(out, string(s))
		}
	}
	return out
}

// The documentation the explanations may cite is shown on the page before any
// question, from the same passages the models receive, so the page never shows a
// passage they do not have.
func TestFrontendDocumentation(t *testing.T) {
	type passage struct {
		Title  string `json:"title"`
		Path   string `json:"path"`
		Anchor string `json:"anchor"`
		Text   string `json:"text"`
	}
	var list []passage
	for _, p := range prompts.Passages {
		list = append(list, passage{p.Title, p.Path, p.Anchor, p.Text})
	}
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	check(t, repoPath("frontend", "src", "lib", "documentation.gen.ts"),
		generatedHeader("TestFrontendDocumentation")+"\nexport interface Passage {\n  title: string\n  path: string\n  anchor: string\n  text: string\n}\n\nexport const PASSAGES: readonly Passage[] = "+string(b)+"\n")
}

// recordsReason writes, per verdict, whether a save of it records a reason, from the
// rule the save follows.
func recordsReason() string {
	var entries []string
	for _, v := range review.Verdicts {
		entries = append(entries, fmt.Sprintf("'%s': %t", v, review.RecordsReason(v)))
	}
	return "export const RECORDS_REASON: Record<Verdict, boolean> = { " + strings.Join(entries, ", ") + " }\n"
}

// tsRecord writes a value set's display labels as a TypeScript record keyed by the
// set's type, in the set's order; every value must have a label.
func tsRecord[K ~string](name, typ string, keys closed.Set[K], labels map[K]string) string {
	var entries []string
	for _, k := range keys {
		label, ok := labels[k]
		if !ok {
			panic(fmt.Sprintf("%s: no label for %q", name, k))
		}
		entries = append(entries, fmt.Sprintf("'%s': '%s'", k, label))
	}
	return fmt.Sprintf("export const %s: Record<%s, string> = { %s }\n", name, typ, strings.Join(entries, ", "))
}

// vocabularyImport is the import line a generated file needs for the value sets it names.
func vocabularyImport(names ...string) string {
	return "import type { " + strings.Join(names, ", ") + " } from './vocabulary.gen'\n"
}

// The frontend reads every response through TypeScript generated from the Go
// types the server answers with, so a response's shape is written once, in Go.
// Each package's wire types sit in one file, the only file generated from.
func TestFrontendTypes(t *testing.T) {
	out := t.TempDir()
	generated := []struct {
		name, path, file, frontmatter string
		mappings                      map[string]string
		// A nil pointer is sent as JSON null, so a package whose types hold pointers
		// reads them as "null" rather than absent.
		absent string
	}{
		{"source", "geolineai/internal/core/intake/source", "issue.go", vocabularyImport("Column", "IssueCode"), map[string]string{}, "null"},
		{"criterion", "geolineai/internal/core/review/criterion", "criterion.go", vocabularyImport("Advisory", "Criterion", "ResultState", "Unit"),
			// Its value types are named as the vocabulary names them, so references resolve to the import.
			map[string]string{}, "null"},
		{"readings", "geolineai/internal/core/review/readings", "readings.go",
			vocabularyImport("LineId", "Slot", "Verdict") + "import type { Issue } from './source.gen'\nimport type { Evaluation } from './criterion.gen'\n",
			map[string]string{"ids.ID": "string", "survey.Line": "LineId", "review.Verdict": "Verdict", "source.Slot": "Slot", "source.Issue": "Issue", "criterion.Advisory": "Advisory", "criterion.Evaluation": "Evaluation"}, "null"},
		{"verdict", "geolineai/internal/core/review/verdict", "outcome.go", "", map[string]string{"ids.ID": "string"}, "null"},
		{"grounded", "geolineai/internal/core/assist/grounded", "source.go",
			vocabularyImport("Outcome", "SourceKind", "UnavailableReason"),
			map[string]string{"SourceKind": "SourceKind", "Outcome": "Outcome", "UnavailableReason": "UnavailableReason"}, "null"},
		{"scenarios", "geolineai/internal/core/scenarios", "list.go", "", map[string]string{"ids.ID": "string"}, "null"},
		// The envelope's detail field is omitted, not null, when the error has none.
		// ErrorCode lives in its own file, so it is a reference to the vocabulary and not a second declaration.
		{"server", "geolineai/internal/server", "envelope.go", vocabularyImport("ErrorCode"), map[string]string{"ids.ID": "string", "ErrorCode": "ErrorCode"}, "undefined"},
	}
	conf := &tygo.Config{}
	for _, g := range generated {
		conf.Packages = append(conf.Packages, &tygo.PackageConfig{
			Path:         g.path,
			OutputPath:   filepath.Join(out, g.name+".gen.ts"),
			IncludeFiles: []string{g.file},
			TypeMappings: g.mappings,
			OptionalType: g.absent,
			Frontmatter:  generatedHeader("TestFrontendTypes") + g.frontmatter,
		})
	}
	if err := tygo.New(conf).Generate(); err != nil {
		t.Fatal(err)
	}
	for _, g := range generated {
		got, err := os.ReadFile(filepath.Join(out, g.name+".gen.ts"))
		if err != nil {
			t.Fatal(err)
		}
		check(t, repoPath("frontend", "src", "lib", g.name+".gen.ts"), string(got))
	}
}

// Every response the frontend reads is pinned as a literal the frontend's types
// must accept, so a Go change to a response fails the frontend build.
func TestFrontendResponses(t *testing.T) {
	// A recorded time in the golden file must not change from run to run.
	verdict.Clock = func() time.Time { return time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { verdict.Clock = time.Now })
	a, ids := advisoryFixture(t)
	a.Manifest = []scenarios.Entry{{Label: "Negative resistivity", File: "line-1.csv", Row: 2}}
	a.Reopen()
	reading := ids["neg"]
	get := func(method, path, body string) string {
		_, out := a.Do(method, path, body)
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	// The fake models answer with one sentence that passes both checks, then with a
	// reply that cannot be read.
	a.Models.Nouls = map[string]float64{"supported": 0.9, "verdict": 0.1}
	a.Models.Written = `{"answerable":true,"sentences":[{"text":"See the cited source.","sources":["S1"]}]}`
	shown := get(http.MethodPost, "/api/readings/"+reading+"/explanations", `{"question":"Why?"}`)
	a.Models.Written = `{"sentences":[]}`
	unavailable := get(http.MethodPost, "/api/readings/"+reading+"/explanations", `{"question":"Why?"}`)

	var b strings.Builder
	b.WriteString("// Generated from real responses by TestFrontendResponses (backend/internal/server/frontend_test.go); do not edit.\n" + rerun)
	b.WriteString("import type { ErrorBody, ExplanationResponse, ReadingDetail, SaveResponse, ScenariosResponse } from '../lib/api'\n")
	for _, r := range []struct{ name, typ, body string }{
		{"scenariosResponse", "ScenariosResponse", get(http.MethodGet, "/api/scenarios", "")},
		{"saveResponse", "SaveResponse", get(http.MethodPost, "/api/reviews", `{"targets":[{"reading_id":"`+reading+`"}],"verdict":"fail"}`)},
		// Read after the save above, so the detail holds a history entry and a reason.
		{"detailResponse", "ReadingDetail", get(http.MethodGet, "/api/readings/"+reading, "")},
		{"errorResponse", "ErrorBody", get(http.MethodPost, "/api/reviews", `{"targets":[{"reading_id":"99999"}],"verdict":"fail"}`)},
		{"shownExplanationResponse", "ExplanationResponse", shown},
		{"unavailableExplanationResponse", "ExplanationResponse", unavailable},
	} {
		fmt.Fprintf(&b, "\nexport const %s = %s satisfies %s\n", r.name, r.body, r.typ)
	}
	check(t, repoPath("frontend", "src", "test", "responses.golden.ts"), b.String())
}
