package verdict

import (
	"context"
	"fmt"
	"strings"
	"time"

	"geolineai/internal/core/ids"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
)

// Store sets one reading's verdict and records the change, all of it or none.
type Store interface {
	// ApplyVerdict reads the reading's current verdict, and when v differs from it,
	// appends a review action stamped by with the reading's history
	// entry, holding the replaced verdict and reason(the stored evaluation), then sets
	// the verdict, all in one write transaction. It reports whether the reading exists
	// and, when it changed, the action's ID; nothing is written for a reading that does
	// not exist or already holds v.
	ApplyVerdict(ctx context.Context, id ids.ID, v review.Verdict, by Stamp, reason func(criterion.Evaluation) *string) (Applied, error)
}

// Stamp is who recorded a review action and when: the fixed reviewer label, and the
// time in the layout the server writes every time it generates.
type Stamp struct {
	Reviewer   string
	RecordedAt string
}

// Applied is what one ApplyVerdict did.
type Applied struct {
	Found bool
	// The action the change recorded, nil when nothing changed.
	ActionID *ids.ID
}

// Clock tells the time a write is recorded at; a variable so tests can fix it.
var Clock = time.Now

// Target is one reading a submission names. Index is its position among the request's
// targets, which a message about it names.
type Target struct {
	Index     int
	ReadingID string
}

// Submission is one save request, as decoded from the wire.
type Submission struct {
	Targets []Target
	Verdict string
}

// InvalidError lists every broken rule, one sentence each.
type InvalidError struct{ Messages []string }

func (e *InvalidError) Error() string { return strings.Join(e.Messages, " ") }

// MissingError names the target that does not exist; nothing is written.
type MissingError struct{ ID ids.ID }

func (e *MissingError) Error() string {
	return fmt.Sprintf("reading %s does not exist", e.ID)
}

// Save applies a submission: validation, then one write that records the change only
// when the verdict differs. A Pass or Fail saves a reason built from the reading's
// stored advisory; Unreviewed saves none.
func Save(ctx context.Context, s Store, sub Submission) (Outcome, error) {
	id, v, err := validate(sub)
	if err != nil {
		return Outcome{}, err
	}
	reason := func(e criterion.Evaluation) *string {
		if !review.RecordsReason(v) {
			return nil
		}
		r := review.Reason(e, v)
		return &r
	}
	applied, err := s.ApplyVerdict(ctx, id, v, Stamp{Reviewer: review.Reviewer, RecordedAt: Clock().UTC().Format(review.TimeLayout)}, reason)
	switch {
	case err != nil:
		return Outcome{}, err
	case !applied.Found:
		return Outcome{}, &MissingError{ID: id}
	case applied.ActionID == nil:
		return Outcome{Affected: []ids.ID{}, Unchanged: []ids.ID{id}}, nil
	}
	return Outcome{Affected: []ids.ID{id}, Unchanged: []ids.ID{}, ActionID: applied.ActionID}, nil
}

// validate checks the targets' count and IDs, and the verdict, reporting every broken
// rule at once, and returns the one target. A save names exactly one reading: there is
// no bulk write.
func validate(sub Submission) (ids.ID, review.Verdict, error) {
	var msgs []string
	if len(sub.Targets) != 1 {
		msgs = append(msgs, fmt.Sprintf("targets must hold exactly one reading; it holds %d.", len(sub.Targets)))
	}
	var target ids.ID
	for _, t := range sub.Targets {
		id, err := ids.Parse(t.ReadingID)
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("targets[%d].reading_id %s", t.Index, ids.ReadingIDRule))
			continue
		}
		target = id
	}
	v, ok := review.Verdicts.Parse(sub.Verdict)
	if !ok {
		msgs = append(msgs, "verdict must be "+review.Verdicts.Choices()+".")
	}
	if len(msgs) > 0 {
		return 0, "", &InvalidError{Messages: msgs}
	}
	return target, v, nil
}
