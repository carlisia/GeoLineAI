// Package grounded asks a text model to answer from a closed list of sources and
// shows its answer only when every sentence survives two checks: code finds every
// number, identifier and sensitive word of the sentence in the sources it cites, and
// the judgment model finds the sentence supported by those sources and telling no
// verdict. One sentence that fails withholds the whole answer, so a reader never
// meets a paragraph with a claim cut out of it.
package grounded
