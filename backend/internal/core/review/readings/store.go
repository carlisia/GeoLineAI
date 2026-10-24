package readings

import (
	"context"
	"errors"

	"geolineai/internal/core/ids"
	"geolineai/internal/core/review"
)

// ErrNotFound is returned for a reading ID no reading has.
var ErrNotFound = errors.New("reading not found")

// Store reads readings as storage holds them; the derivations the page shows are
// made here, from what it returns.
type Store interface {
	Detail(ctx context.Context, id ids.ID) (Stored, error)
}

// Stored is one reading's detail as storage holds it: the import's source path in
// place of SourceFilename, the import issues in no particular order, its whole repeat
// group in reading order, the reading itself included, and its history entries
// oldest first. New verdicts and the would-save reasons are derived on read, so
// storage has no field for them.
type Stored struct {
	Detail
	SourcePath string
	Group      []Repeat
	History    []StoredEntry
}

// StoredEntry is one history entry as storage holds it: the verdict the action
// replaced, the reason saved with the verdict it set, and the action's reviewer,
// time and size.
type StoredEntry struct {
	ActionID        ids.ID
	Reviewer        string
	RecordedAt      string
	PreviousVerdict review.Verdict
	Reason          *string
	EntryCount      int
}
