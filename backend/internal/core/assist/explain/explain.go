package explain

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"geolineai/internal/core/assist"
	"geolineai/internal/core/assist/grounded"
	"geolineai/internal/core/assist/prompts"
	"geolineai/internal/core/closed"
	"geolineai/internal/core/ids"
	"geolineai/internal/core/intake/source"
	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/review/readings"
	"geolineai/internal/core/survey"
)

// MaxQuestion is the longest question, in characters, a reviewer may ask.
var MaxQuestion = prompts.Settings.LLM.QuestionCharacters

// maxSentences is the most sentences an answer may have: an answer is a short
// explanation beside the evidence, not a report.
var maxSentences = prompts.Settings.LLM.ExplainSentences

// system tells the text model what to write and what it may not.
var system = prompts.ExplainSystem

// InvalidError is a question the rules refuse; its message names the rule.
type InvalidError struct{ Message string }

func (e *InvalidError) Error() string { return e.Message }

// Store is what an explanation reads: the reading's detail, and the transmitting pairs
// of its line's readings with unresolved Transmitter contact, which say how its
// electrodes fared.
type Store interface {
	readings.Store
	FailingContacts(ctx context.Context, line survey.Line) ([]assist.FailingContact, error)
}

// Ask answers a question about one reading's criteria. A question that is blank
// after trimming or longer than MaxQuestion is an InvalidError and an unknown reading
// is the readings package's not-found error; with no models, or a model that cannot
// be reached or read, the answer is Unavailable with the reading's sources.
func Ask(ctx context.Context, s Store, w assist.Writer, j assist.Judge, id ids.ID, question string) (grounded.Answer, error) {
	q := strings.TrimSpace(question)
	if q == "" || utf8.RuneCountInString(q) > MaxQuestion {
		return grounded.Answer{}, &InvalidError{fmt.Sprintf("question must be 1 to %d characters.", MaxQuestion)}
	}
	d, err := readings.Get(ctx, s, id)
	if err != nil {
		return grounded.Answer{}, err
	}
	line, err := s.FailingContacts(ctx, d.LineID)
	if err != nil {
		return grounded.Answer{}, err
	}
	return grounded.Ask(ctx, w, j, grounded.Request{
		System:       system,
		Task:         map[string]any{"question": q, "reading": shown(d.AcquisitionID)},
		Sources:      sources(d, line),
		MaxSentences: maxSentences,
	})
}

// sources lists what an answer about one reading may cite, numbered in the order
// offered: the reading's results, facts read from its stored values, every
// criterion's rule with the advisory's, and the documentation's passages.
func sources(d readings.Detail, line []assist.FailingContact) []grounded.Source {
	acq := shown(d.AcquisitionID)
	var out []grounded.Source
	for _, r := range d.Advisory.Results {
		out = append(out, grounded.Source{Kind: grounded.KindResult, Title: "Result: " + label(r.Criterion),
			Text: fmt.Sprintf("%s is %s for %s. %s", label(r.Criterion), stateLabel(r.State), acq, r.Explanation)})
	}
	out = append(out, facts(d, line)...)
	out = append(out, grounded.Source{Kind: grounded.KindRule, Title: "Rule: the advisory", Text: criterion.AdvisoryRule()})
	for _, c := range criterion.Criteria {
		out = append(out, grounded.Source{Kind: grounded.KindRule, Title: "Rule: " + label(c), Text: criterion.Rule(c)})
	}
	for _, p := range prompts.Passages {
		doc := p.Path + "#" + p.Anchor
		out = append(out, grounded.Source{Kind: grounded.KindDoc, Title: p.Title, Text: p.Text, Doc: &doc})
	}
	for i := range out {
		out[i].ID = "S" + strconv.Itoa(i+1)
	}
	return out
}

// facts are what the reading's stored values say, each in one source: its advisory
// and how many results of each state decided it, which criteria recorded a result,
// its values, electrodes, decay samples, import issues and repeats, and how each of
// its electrodes fared when it transmitted with failing contact on its line.
func facts(d readings.Detail, line []assist.FailingContact) []grounded.Source {
	acq := shown(d.AcquisitionID)
	fact := func(title, text string) grounded.Source {
		return grounded.Source{Kind: grounded.KindFact, Title: title, Text: text}
	}
	count := map[criterion.ResultState]int{}
	recorded := map[criterion.Criterion]bool{}
	for _, r := range d.Advisory.Results {
		count[r.State]++
		recorded[r.Criterion] = true
	}
	var with, without []string
	for _, c := range criterion.Criteria {
		if c == criterion.ImportIssue {
			continue
		}
		if recorded[c] {
			with = append(with, label(c))
		} else {
			without = append(without, label(c))
		}
	}
	coverage := fmt.Sprintf("Criteria with a recorded result for %s: %s.", acq, listOr(with, "none"))
	if len(without) > 0 {
		coverage += fmt.Sprintf(" Criteria with no recorded result for %s: %s.", acq, strings.Join(without, ", "))
	} else {
		coverage += fmt.Sprintf(" Every criterion other than Import issue recorded a result for %s.", acq)
	}
	out := []grounded.Source{
		fact("Fact: the advisory of this reading", fmt.Sprintf("The advisory of %s is %s: of its %d recorded results, %d are Violated, %d are Review and %d are Met.",
			acq, advisoryLabel(d.Advisory.Value), len(d.Advisory.Results), count[criterion.Violated], count[criterion.Review], count[criterion.Met])),
		fact("Fact: which criteria recorded a result", coverage),
		fact("Fact: values of this reading", values(d)),
		fact("Fact: electrodes of this reading", electrodes(d)),
		fact("Fact: decay samples of this reading", decay(d)),
		fact("Fact: import issues of this reading", issues(d)),
		fact("Fact: repeats of this reading", repeatsOf(d)),
	}
	return append(out, contacts(d, line)...)
}

func values(d readings.Detail) string {
	v := func(x *float64, unit string) string {
		if x == nil {
			return "unavailable"
		}
		return number(*x) + " " + unit
	}
	n := func(x *int64) string {
		if x == nil {
			return "unavailable"
		}
		return strconv.FormatInt(*x, 10)
	}
	return fmt.Sprintf("Values of %s: apparent resistivity %s, reported resistivity error %s; chargeability %s, reported chargeability error %s; transmitter-pair contact resistance %s; input current %s; stack count %s; n factor %s.",
		shown(d.AcquisitionID), v(d.ApparentResistivity, "ohm-m"), v(d.ApparentResistivityErr, "ohm-m"), v(d.Chargeability, "V/V"), v(d.ChargeabilityErr, "V/V"),
		v(d.ContactResistance, "ohm"), v(d.InputCurrent, "A"), n(d.Stacks), n(d.Symmetry))
}

func electrodes(d readings.Detail) string {
	var parts []string
	for _, e := range d.Electrodes {
		id := "blank"
		if e.ElectrodeID != nil {
			id = *e.ElectrodeID
		}
		parts = append(parts, string(e.Slot)+" "+id)
	}
	return fmt.Sprintf("Electrodes of %s: %s. tx1 and tx2 are the transmitting pair; rx1 and rx2 the receiving pair.", shown(d.AcquisitionID), strings.Join(parts, ", "))
}

// decay lists the samples and every gate where a sample rises above the one before:
// the comparisons a question about the decay asks for, made here rather than left to
// a model to make.
func decay(d readings.Detail) string {
	acq := shown(d.AcquisitionID)
	if d.Decay.Samples == nil {
		return fmt.Sprintf("The decay curve of %s is unavailable.", acq)
	}
	samples := *d.Decay.Samples
	parts := make([]string, len(samples))
	var rises []string
	for i, s := range samples {
		parts[i] = number(s)
		if i > 0 && s > samples[i-1] {
			rises = append(rises, fmt.Sprintf("from gate %d to gate %d (%s to %s V/V)", i, i+1, number(samples[i-1]), number(s)))
		}
	}
	text := fmt.Sprintf("Decay samples of %s, gate 1 to gate %d at %s to %s ms: %s V/V.", acq, survey.DecaySamples,
		number(survey.DecayGateCentresMs[0]), number(survey.DecayGateCentresMs[survey.DecaySamples-1]), strings.Join(parts, ", "))
	if len(rises) == 0 {
		return text + " No sample is above the one before it."
	}
	return text + " A sample rises above the one before it " + strings.Join(rises, "; ") + "."
}

func issues(d readings.Detail) string {
	acq := shown(d.AcquisitionID)
	if len(d.ImportIssues) == 0 {
		return fmt.Sprintf("No import issues are recorded for %s.", acq)
	}
	var parts []string
	for _, i := range d.ImportIssues {
		parts = append(parts, fmt.Sprintf("%s %s", i.Field, issueLabel(i.Code)))
	}
	return fmt.Sprintf("Import issues of %s: %s.", acq, strings.Join(parts, "; "))
}

// repeatsOf names each repeat with its resistivity and whether it counts toward the
// reference, as the import recorded it.
func repeatsOf(d readings.Detail) string {
	acq := shown(d.AcquisitionID)
	switch {
	case d.RepeatGroupID == nil:
		return fmt.Sprintf("%s has no repeat group: an electrode slot is blank.", acq)
	case len(d.Repeats) == 0:
		return fmt.Sprintf("%s has no repeats: no other reading of its import has the same four electrode IDs.", acq)
	}
	var parts []string
	counting := 0
	for _, r := range d.Repeats {
		rho := "with no resistivity"
		if r.ApparentResistivity != nil {
			rho = "at " + number(*r.ApparentResistivity) + " ohm-m"
		}
		role := "counts toward the reference"
		if r.LeftOut {
			role = "left out of the reference"
		} else {
			counting++
		}
		parts = append(parts, fmt.Sprintf("%s %s, %s", shown(r.AcquisitionID), rho, role))
	}
	return fmt.Sprintf("Repeats of %s, the other readings of its import with the same four electrode IDs: %s. %d of its %d repeats count toward the reference.",
		acq, strings.Join(parts, "; "), counting, len(d.Repeats))
}

// contacts says, for each electrode of the reading that transmitted with violated
// Transmitter contact anywhere on its line, in how many such readings and with which
// partners: the evidence behind a suspect electrode, read from the stored results.
// Contact belongs to the transmitting pair, so only the tx slots count.
func contacts(d readings.Detail, line []assist.FailingContact) []grounded.Source {
	var out []grounded.Source
	seen := map[string]bool{}
	for _, e := range d.Electrodes {
		if e.ElectrodeID == nil || seen[*e.ElectrodeID] {
			continue
		}
		id := *e.ElectrodeID
		seen[id] = true
		partners := map[string]int{}
		total := 0
		for _, p := range line {
			if p.Tx1 == nil || p.Tx2 == nil {
				continue
			}
			a, b := *p.Tx1, *p.Tx2
			switch id {
			case a:
				partners[b]++
			case b:
				partners[a]++
			default:
				continue
			}
			total++
		}
		if total == 0 {
			continue
		}
		names := slices.Sorted(maps.Keys(partners))
		var parts []string
		for _, n := range names {
			parts = append(parts, fmt.Sprintf("%d with partner %s", partners[n], n))
		}
		// Transmitter contact is only ever Met or Violated, never Review, so every
		// unresolved result the line reports here is a violation and the text says so.
		out = append(out, grounded.Source{
			Kind:  grounded.KindFact,
			Title: "Fact: failing transmitter contact of " + id,
			Text: fmt.Sprintf("On %s, %s transmits in %d readings with Violated Transmitter contact: %s. It has %d different partners in those readings.",
				d.LineID, id, total, strings.Join(parts, ", "), len(names)),
		})
	}
	return out
}

// shown is an acquisition ID as the page shows it, (blank) for blank text.
func shown(acq string) string {
	if strings.TrimSpace(acq) == "" {
		return "(blank)"
	}
	return acq
}

// number writes a stored value in its shortest exact decimal form.
func number(x float64) string {
	return strconv.FormatFloat(x, 'f', -1, 64)
}

func listOr(items []string, none string) string {
	if len(items) == 0 {
		return none
	}
	return strings.Join(items, ", ")
}

func label(c criterion.Criterion) string { return closed.Entry(criterion.Labels, c, "criterion") }

func stateLabel(s criterion.ResultState) string {
	return closed.Entry(criterion.ResultStateLabels, s, "result state")
}

func advisoryLabel(a criterion.Advisory) string {
	return closed.Entry(criterion.AdvisoryLabels, a, "advisory")
}

func issueLabel(c source.IssueCode) string {
	return closed.Entry(criterion.IssueLabels, c, "import issue code")
}
