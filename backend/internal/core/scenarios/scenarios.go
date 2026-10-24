package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"geolineai/internal/core/ids"
)

// Count is how many scenarios the manifest holds.
const Count = 3

// Entry is one scenario as the manifest names it.
type Entry struct {
	Label string `json:"label"`
	File  string `json:"file"`
	// Row is the row number the import records: the physical line the record starts
	// on, counting the header as line 1.
	Row int `json:"row"`
}

// Store finds the readings imported from a file, named by its base name, at a row.
type Store interface {
	ReadingsAt(ctx context.Context, file string, row int) ([]ids.ID, error)
}

// UnresolvedError names a scenario whose identity matches no reading, or more than one.
type UnresolvedError struct {
	Entry   Entry
	Matches int
}

func (e *UnresolvedError) Error() string {
	if e.Matches == 0 {
		return fmt.Sprintf("scenario %q names %s row %d, which no imported reading has", e.Entry.Label, e.Entry.File, e.Entry.Row)
	}
	return fmt.Sprintf("scenario %q names %s row %d, which %d imported readings have; import the file once", e.Entry.Label, e.Entry.File, e.Entry.Row, e.Matches)
}

// Load reads and checks the manifest at path: exactly Count entries, each with a
// label, a file name without a directory and a row past the header, no two alike.
func Load(path string) ([]Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&entries); err != nil {
		return nil, fmt.Errorf("read scenario manifest %s: %w", path, err)
	}
	var problems []string
	if len(entries) != Count {
		problems = append(problems, fmt.Sprintf("it must list exactly %d scenarios, not %d", Count, len(entries)))
	}
	seen := map[Entry]bool{}
	for i, e := range entries {
		key := Entry{File: e.File, Row: e.Row}
		switch {
		case strings.TrimSpace(e.Label) == "":
			problems = append(problems, fmt.Sprintf("entry %d has no label", i))
		case e.File == "" || strings.ContainsAny(e.File, `/\`):
			problems = append(problems, fmt.Sprintf("entry %d must name a file by its base name", i))
		case e.Row < 2:
			problems = append(problems, fmt.Sprintf("entry %d must name a row of 2 or more", i))
		case seen[key]:
			problems = append(problems, fmt.Sprintf("entry %d repeats %s row %d", i, e.File, e.Row))
		}
		seen[key] = true
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("scenario manifest %s: %s", path, strings.Join(problems, "; "))
	}
	return entries, nil
}

// Resolve finds the one reading each entry names, in manifest order. An entry that
// matches no reading, or several, is an UnresolvedError.
func Resolve(ctx context.Context, s Store, entries []Entry) (List, error) {
	out := List{Scenarios: make([]Scenario, 0, len(entries))}
	for _, e := range entries {
		found, err := s.ReadingsAt(ctx, e.File, e.Row)
		if err != nil {
			return List{}, err
		}
		if len(found) != 1 {
			return List{}, &UnresolvedError{Entry: e, Matches: len(found)}
		}
		out.Scenarios = append(out.Scenarios, Scenario{Label: e.Label, ReadingID: found[0]})
	}
	return out, nil
}
