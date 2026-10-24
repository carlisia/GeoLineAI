package grounded

// The frontend's copy of the types in this file is generated from it.

// Source is one piece of evidence an answer may cite. Its text is built by code or
// quoted verbatim, so it is true by construction, and the page shows it whole in the
// source's card.
type Source struct {
	ID    string     `json:"id"`
	Kind  SourceKind `json:"kind"`
	Title string     `json:"title"`
	Text  string     `json:"text"`
	// The repository path of the quoted document, with the heading anchor it sits
	// under; nil for every kind but a doc.
	Doc *string `json:"doc"`
}

// Sentence is one written sentence with what the checks found. The page shows its
// text only when the whole answer is shown, and shows Problems when it is withheld.
type Sentence struct {
	Text    string   `json:"text"`
	Sources []string `json:"sources"`
	// One sentence per failed code check, and one per failed judgment.
	Problems []string `json:"problems"`
	// The judgment model's probability that the cited sources support every claim of
	// the sentence, and that the sentence tells the reviewer what verdict to record.
	Supported    float64 `json:"supported"`
	TellsVerdict float64 `json:"tells_verdict"`
}

// Answer is a written answer after the checks, with every source it was offered, so
// each citation resolves to a source the page can show. Reason is set exactly when
// the outcome is Unavailable; Sentences is then empty.
type Answer struct {
	Outcome   Outcome            `json:"outcome"`
	Reason    *UnavailableReason `json:"reason"`
	Sentences []Sentence         `json:"sentences"`
	Sources   []Source           `json:"sources"`
	Models    Models             `json:"models"`
}

// Models names the model versions that produced one answer.
type Models struct {
	Writer string `json:"writer"`
	Judge  string `json:"judge"`
}
