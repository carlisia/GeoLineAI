package grounded

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"geolineai/internal/core/review/criterion"
	"geolineai/internal/core/survey"
)

// The code checks: what a sentence may say only when a source it cites says it too.
// They are literal, so a true sentence that rewords a number or an identifier fails;
// failing withholds the answer, which the reviewer sees, while passing a false one
// would not be seen.

var (
	// identifier matches the IDs a sentence may name, in any case: a line, an
	// electrode or an acquisition ID, an electrode by its short form, a slot, and a
	// domain answer.
	identifier = regexp.MustCompile(`(?i)\b(?:` + lineNames() + `)(?:-[er]\d+)?\b|\be\d{2}\b|\b[tr]x[12]\b|\bC\d{1,2}\b`)
	// number matches a written number: a sign, digits with optional thousands commas,
	// a fraction, an exponent and a percent sign.
	number = regexp.MustCompile(`[-+]?(?:(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?|\.\d+)(?:[eE][-+]?\d+)?%?`)
	// wordNumbers are the counts a sentence may spell out. "one" is left out: as a
	// pronoun it would fail sentences that count nothing.
	wordNumbers = map[string]float64{"zero": 0, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
		"eleven": 11, "twelve": 12, "thirteen": 13, "fourteen": 14, "fifteen": 15, "sixteen": 16, "seventeen": 17, "eighteen": 18, "nineteen": 19, "twenty": 20}
	word = regexp.MustCompile(`[A-Za-z]+(?:'[a-z]+)?`)
)

// sensitive are word stems that carry a recommendation, a hedge, a diagnosis, a cause,
// a judgment of quality or a rank: the claims a model adds when it explains more than
// its sources say. A sentence may use one only when a source it cites uses the same
// stem.
var sensitive = []string{
	"should", "must", "ought", "need", "recommend", "advis", "suggest",
	"likely", "probabl", "possibl", "perhaps", "may", "might", "could", "appear", "seem", "indicat", "impli", "imply",
	"fault", "malfunction", "broken", "glitch", "nois", "poor", "bad", "junk", "problem", "issue", "defect",
	"impossible", "implausible", "physical", "cause", "due",
	"reliab", "unreliab", "trust", "valid", "invalid", "wrong", "correct", "accura", "inaccura",
	"suspicious", "anomal", "outlier", "concern", "worr", "serious", "severe", "significan", "substantial", "dramatic", "extreme", "notabl",
	"accept", "reject", "discard", "exclude", "keep", "remov", "drop", "approv", "safe", "only",
	"largest", "biggest", "smallest", "highest", "lowest", "worst", "best", "greatest", "fewest", "least", "most", "top",
	"majority", "minority", "heaviest", "strongest", "weakest",
}

// lineNames is the survey's line names as one alternation, so the identifiers a
// sentence may name follow the lines the survey has.
func lineNames() string {
	names := make([]string, len(survey.Lines))
	for i, l := range survey.Lines {
		names[i] = regexp.QuoteMeta(string(l))
	}
	return strings.Join(names, "|")
}

// stemOf is the sensitive stem a word starts with, or "" when none does. "need" in
// "needs review" is the advisory's own label, so a source naming the advisory allows it.
func stemOf(w string) string {
	w = strings.ToLower(w)
	for _, s := range sensitive {
		if strings.HasPrefix(w, s) {
			return s
		}
	}
	return ""
}

// stems lists the sensitive stems a text uses.
func stems(text string) map[string]bool {
	out := map[string]bool{}
	for _, w := range word.FindAllString(text, -1) {
		if s := stemOf(w); s != "" {
			out[s] = true
		}
	}
	return out
}

// written is one number as written: its value, with the percent sign kept apart, so
// "20%" equals "20.0%" but not "0.20".
type written struct {
	value   float64
	percent bool
}

// numbers lists the numbers a text writes outside its identifiers, and the spelled-out
// counts as their values.
func numbers(text string) []written {
	masked := identifier.ReplaceAllStringFunc(text, func(m string) string { return strings.Repeat(" ", len(m)) })
	var out []written
	for _, m := range number.FindAllString(masked, -1) {
		percent := strings.HasSuffix(m, "%")
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSuffix(m, "%"), ",", ""), 64)
		if err != nil {
			continue
		}
		out = append(out, written{v, percent})
	}
	for _, w := range word.FindAllString(masked, -1) {
		if v, ok := wordNumbers[strings.ToLower(w)]; ok {
			out = append(out, written{v, false})
		}
	}
	return out
}

// sameNumber reports whether two written numbers are equal to the precision the
// shorter one is written in, so a source's 205.1 and a sentence's 205.10 agree.
func sameNumber(a, b written) bool {
	if a.percent != b.percent {
		return false
	}
	return a.value == b.value || math.Abs(a.value-b.value) <= 1e-9*math.Max(math.Abs(a.value), math.Abs(b.value))
}

// codeProblems checks one sentence against the sources it cites, which must exist:
// each problem is a sentence the page can show.
func codeProblems(text string, cited []Source, byID map[string]Source, ids []string) []string {
	problems := []string{}
	if len(ids) == 0 {
		problems = append(problems, "It cites no source.")
	}
	for _, id := range ids {
		if _, ok := byID[id]; !ok {
			problems = append(problems, fmt.Sprintf("It cites %s, which is not one of its sources.", id))
		}
	}
	var corpus strings.Builder
	for _, s := range cited {
		fmt.Fprintf(&corpus, "%s\n%s\n", s.Title, s.Text)
	}
	all := corpus.String()
	have := numbers(all)
	for _, n := range numbers(text) {
		if !slices.ContainsFunc(have, func(h written) bool { return sameNumber(h, n) }) {
			problems = append(problems, fmt.Sprintf("The number %s is not in its sources.", format(n)))
		}
	}
	lower := strings.ToLower(all)
	for _, id := range identifier.FindAllString(text, -1) {
		if !strings.Contains(lower, strings.ToLower(id)) {
			problems = append(problems, fmt.Sprintf("%q is not in its sources.", id))
		}
	}
	for _, label := range criterion.Labels {
		if strings.Contains(strings.ToLower(text), strings.ToLower(label)) && !strings.Contains(lower, strings.ToLower(label)) {
			problems = append(problems, fmt.Sprintf("The criterion %q is not in its sources.", label))
		}
	}
	allowed := stems(all)
	var unsupported []string
	for s := range stems(text) {
		if !allowed[s] {
			unsupported = append(unsupported, s)
		}
	}
	slices.Sort(unsupported)
	for _, s := range unsupported {
		problems = append(problems, fmt.Sprintf("It uses a word starting %q, which its sources do not use.", s))
	}
	return problems
}

func format(n written) string {
	s := strconv.FormatFloat(n.value, 'g', -1, 64)
	if n.percent {
		s += "%"
	}
	return s
}
