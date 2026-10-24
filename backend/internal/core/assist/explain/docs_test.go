package explain

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"geolineai/internal/core/assist/prompts"
)

// A passage is shown as a quotation with its document and heading; once the document
// is edited and the passage no longer stands in it, or the heading is renamed, an
// answer would cite words the documentation does not hold, or point to nowhere.
func TestPassagesQuoteTheirDocumentsVerbatim(t *testing.T) {
	heading := regexp.MustCompile(`(?m)^#{2,3} (.+)$`)
	drop := regexp.MustCompile(`[^a-z0-9 -]`)
	for _, p := range prompts.Passages {
		raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", p.Path))
		if err != nil {
			t.Fatalf("%s: %v", p.Path, err)
		}
		doc := strings.NewReplacer("**", "", "`", "").Replace(string(raw))
		if !strings.Contains(doc, p.Text) {
			t.Errorf("%s no longer holds the passage %q", p.Path, p.Title)
		}
		anchors := map[string]bool{}
		for _, h := range heading.FindAllStringSubmatch(doc, -1) {
			anchors[strings.ReplaceAll(drop.ReplaceAllString(strings.ToLower(h[1]), ""), " ", "-")] = true
		}
		if !anchors[p.Anchor] {
			t.Errorf("%s has no heading with the anchor %s", p.Path, p.Anchor)
		}
	}
}
