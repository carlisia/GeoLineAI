// Package scenarios is the fixed set of sample readings the page offers. A manifest
// names each by a label and a stable source identity, the source file's name and the
// row number the import recorded, and storage resolves that identity to the reading's
// internal ID whenever it is asked, so no internal ID is assumed to survive a reseed.
package scenarios
