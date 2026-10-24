package readings

import (
	"context"
	"path/filepath"
	"slices"

	"geolineai/internal/core/ids"
	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review"
	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/survey"
)

// The frontend's copy of the types in this file is generated from it.

type Electrode struct {
	Slot        source.Slot `json:"slot"`
	ElectrodeID *string     `json:"electrode_id"`
	Lat         *float64    `json:"lat"`
	Lon         *float64    `json:"lon"`
	Alt         *float64    `json:"alt"`
}

type Decay struct {
	Samples *[]float64 `json:"samples"` // nil, sent as null, when unavailable
}

type Detail struct {
	ReadingID              ids.ID         `json:"reading_id"`
	AcquisitionID          string         `json:"acquisition_id"`
	LineID                 survey.Line    `json:"line_id"`
	SourceFilename         string         `json:"source_filename"`
	RowNumber              int            `json:"row_number"`
	Timestamp              *string        `json:"timestamp"`
	ArrayType              *string        `json:"array_type"`
	Symmetry               *int64         `json:"symmetry"`
	Stacks                 *int64         `json:"stacks"`
	InputCurrent           *float64       `json:"input_current"`
	ContactResistance      *float64       `json:"contact_resistance"`
	ApparentResistivity    *float64       `json:"apparent_resistivity"`
	ApparentResistivityErr *float64       `json:"apparent_resistivity_err"`
	Chargeability          *float64       `json:"chargeability"`
	ChargeabilityErr       *float64       `json:"chargeability_err"`
	Electrodes             []Electrode    `json:"electrodes"`
	Decay                  Decay          `json:"decay"`
	ImportIssues           []source.Issue `json:"import_issues"`
	Verdict                review.Verdict `json:"verdict"`
	// The reading's repeat group, nil when a slot is blank; its repeat difference, nil
	// when its repeats give nothing sound to compare against; and its repeats, in
	// reading order.
	RepeatGroupID    *ids.ID  `json:"repeat_group_id"`
	RepeatDifference *float64 `json:"repeat_difference"`
	Disagrees        bool     `json:"disagrees"`
	Repeats          []Repeat `json:"repeats"`
	// The reading's verdict changes, newest first.
	History []HistoryEntry `json:"history"`
	// The advisory with its results; the reason the latest save recorded, nil when
	// the verdict is Unreviewed; and the reason each of Pass and Fail would save now,
	// built the way the save builds it.
	Advisory         criterion.Evaluation `json:"advisory"`
	Reason           *string              `json:"reason"`
	WouldSaveReasons WouldSaveReasons     `json:"would_save_reasons"`
}

// WouldSaveReasons are the reasons a save of each decided verdict would record.
type WouldSaveReasons struct {
	Pass string `json:"pass"`
	Fail string `json:"fail"`
}

// HistoryEntry is one review action's change to the reading: the verdict it replaced
// and the one it set, the action's reviewer and time, and how many readings the action
// changed in all, which is one for every save.
type HistoryEntry struct {
	ActionID        ids.ID         `json:"action_id"`
	Reviewer        string         `json:"reviewer"`
	RecordedAt      string         `json:"recorded_at"`
	PreviousVerdict review.Verdict `json:"previous_verdict"`
	NewVerdict      review.Verdict `json:"new_verdict"`
	// The reason saved with the new verdict, nil for Unreviewed.
	Reason     *string `json:"reason"`
	EntryCount int     `json:"entry_count"`
}

// Repeat is one of a reading's repeats, with its own repeat difference against the
// group's other readings.
type Repeat struct {
	ReadingID           ids.ID         `json:"reading_id"`
	AcquisitionID       string         `json:"acquisition_id"`
	Timestamp           *string        `json:"timestamp"`
	ApparentResistivity *float64       `json:"apparent_resistivity"`
	Chargeability       *float64       `json:"chargeability"`
	Verdict             review.Verdict `json:"verdict"`
	RepeatDifference    *float64       `json:"repeat_difference"`
	Disagrees           bool           `json:"disagrees"`
	// LeftOut marks a repeat that does not count toward its group's reference: it has
	// no resistivity, or a single-reading criterion found it violated.
	LeftOut bool `json:"left_out"`
}

// Get reads one reading's detail: the stored values with the file's name in place of
// its path, its import issues in source-column order, its repeats with the differences
// the import recorded, its history with each change's new verdict derived, and the
// reasons a save of Pass or Fail would record now.
func Get(ctx context.Context, s Store, id ids.ID) (Detail, error) {
	stored, err := s.Detail(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	d := stored.Detail
	d.SourceFilename = filepath.Base(stored.SourcePath)
	if d.ImportIssues == nil {
		d.ImportIssues = []source.Issue{}
	}
	// Storage holds only the verdict each change replaced; the verdict it set is the
	// next change's replaced one, or the current verdict for the latest change.
	// The current reason is the one saved with the latest change, nil for none.
	d.Reason = nil
	if n := len(stored.History); n > 0 {
		d.Reason = stored.History[n-1].Reason
	}
	d.History = make([]HistoryEntry, 0, len(stored.History))
	for i := len(stored.History) - 1; i >= 0; i-- {
		s := stored.History[i]
		e := HistoryEntry{ActionID: s.ActionID, Reviewer: s.Reviewer, RecordedAt: s.RecordedAt, PreviousVerdict: s.PreviousVerdict, Reason: s.Reason, EntryCount: s.EntryCount}
		if i+1 < len(stored.History) {
			e.NewVerdict = stored.History[i+1].PreviousVerdict
		} else {
			e.NewVerdict = d.Verdict
		}
		d.History = append(d.History, e)
	}
	// The group's other readings, each with the difference and agreement its import
	// recorded.
	d.Repeats = []Repeat{}
	for _, m := range stored.Group {
		if m.ReadingID == d.ReadingID {
			continue
		}
		d.Repeats = append(d.Repeats, m)
	}
	if d.Advisory.Results == nil {
		d.Advisory.Results = []criterion.Result{}
	}
	d.WouldSaveReasons = WouldSaveReasons{Pass: review.Reason(d.Advisory, review.Pass), Fail: review.Reason(d.Advisory, review.Fail)}
	// ColumnIndex panics on an unknown field. It relies on the schema holding the
	// issue's field to the source columns, so every stored field is one of them.
	slices.SortStableFunc(d.ImportIssues, func(a, b source.Issue) int {
		return source.ColumnIndex(a.Field) - source.ColumnIndex(b.Field)
	})
	return d, nil
}
