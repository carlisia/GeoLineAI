// Package closed is the closed value set: a fixed list of wire spellings, with the
// one parser and the one list of names every such set shares.
package closed

import (
	"fmt"
	"slices"

	"github.com/dustin/go-humanize/english"
)

// Set is every value of a closed set, in the order a message or a menu lists them.
type Set[T ~string] []T

// Parse returns s as a value of the set when it is exactly one of them. Any other
// text gives the zero value, so a rejected spelling never travels on as a value.
func (set Set[T]) Parse(s string) (T, bool) {
	if v := T(s); slices.Contains(set, v) {
		return v, true
	}
	var none T
	return none, false
}

// Choices lists the set's spellings as a sentence offers them: "a, b or c".
func (set Set[T]) Choices() string { return english.WordSeries(set.Names(), "or") }

// Names lists the set's wire and storage spellings, in the set's order.
func (set Set[T]) Names() []string {
	names := make([]string, len(set))
	for i, v := range set {
		names[i] = string(v)
	}
	return names
}

// ParseStored parses a value storage holds, which the schema confines to the set, so
// one outside it is an error naming what it should have been.
func (set Set[T]) ParseStored(s, what string) (T, error) {
	v, ok := set.Parse(s)
	if !ok {
		return v, fmt.Errorf("stored %s %q is not one", what, s)
	}
	return v, nil
}

// Entry is m's entry for k. Each map it serves holds an entry for every value of its
// set, so a missing one is a programming error and panics, naming the value.
func Entry[K ~string, V any](m map[K]V, k K, what string) V {
	v, ok := m[k]
	if !ok {
		panic(fmt.Sprintf("%s %q has no entry", what, k))
	}
	return v
}
