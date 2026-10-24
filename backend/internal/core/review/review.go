package review

import "geolineai/internal/core/closed"

// Verdict is the reviewer's judgment on one reading. Its values are exactly
// the wire and storage spellings.
type Verdict string

const (
	Pass       Verdict = "pass"
	Fail       Verdict = "fail"
	Unreviewed Verdict = "unreviewed"
)

// Verdicts lists every verdict, in display order.
var Verdicts = closed.Set[Verdict]{Pass, Fail, Unreviewed}

// VerdictLabels are the verdicts' display labels, as a saved reason and the page both
// write them.
var VerdictLabels = map[Verdict]string{Pass: "Pass", Fail: "Fail", Unreviewed: "Unreviewed"}

// TimeLayout is how the server writes the times it generates: UTC, RFC 3339
// with exactly three fractional digits.
const TimeLayout = "2006-01-02T15:04:05.000Z"

// Reviewer is the fixed label every review action is attributed to: the one human
// recording verdicts, not an authenticated identity.
const Reviewer = "reviewer"
