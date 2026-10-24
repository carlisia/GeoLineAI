// Package criterion evaluates every retained reading against fixed criteria at
// import and decides its advisory from the results alone: Fail when a criterion is
// violated, Needs review when some result needs a human, Pass otherwise. The
// advisory never sets a verdict; the reviewer does, and may disagree with it.
package criterion
