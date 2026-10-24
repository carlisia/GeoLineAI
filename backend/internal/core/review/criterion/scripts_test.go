package criterion

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geolineai/internal/core/repeats"
	"geolineai/internal/core/survey"
)

// scriptsDir is the repository's scripts directory, as seen from this package.
var scriptsDir = filepath.Join("..", "..", "..", "..", "..", "scripts")

// The dataset generator and the dataset-facts script keep their own copies of the
// limits and the gate centres, so the facts they print stay an independent reading of
// the line; a copy that drifts from the values the criteria use would make those facts,
// and the documentation built on them, describe rules the application does not apply.
func TestScriptConstantsMatchTheCriteria(t *testing.T) {
	facts, err := os.ReadFile(filepath.Join(scriptsDir, "dataset-facts.py"))
	require.NoError(t, err)
	generator, err := os.ReadFile(filepath.Join(scriptsDir, "generate-line", "generate.py"))
	require.NoError(t, err)

	for name, want := range map[string]float64{
		"RESISTIVITY_ERROR_LIMIT":   ResistivityRelativeErrorLimit,
		"CHARGEABILITY_ERROR_LIMIT": ChargeabilityRelativeErrorLimit,
		"CONTACT_MAX":               ContactResistanceMax,
		"REPEAT_LIMIT":              repeats.Limit,
	} {
		m := regexp.MustCompile(`(?m)^` + name + ` = ([0-9.]+)$`).FindSubmatch(facts)
		require.NotNil(t, m, "dataset-facts.py has no %s", name)
		got, err := strconv.ParseFloat(string(m[1]), 64)
		require.NoError(t, err)
		assert.Equal(t, want, got, name)
	}

	gates := func(text []byte, pattern string) []float64 {
		m := regexp.MustCompile(pattern).FindSubmatch(text)
		require.NotNil(t, m, "no gate centres matching %s", pattern)
		var out []float64
		for _, f := range strings.Split(string(m[1]), ",") {
			v, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
			require.NoError(t, err)
			out = append(out, v)
		}
		return out
	}
	assert.Equal(t, survey.DecayGateCentresMs[:], gates(facts, `(?m)^GATES = \[([^\]]+)\]`), "dataset-facts.py gate centres")
	assert.Equal(t, survey.DecayGateCentresMs[:], gates(generator, `(?s)DECAY_GATES_MS = np\.array\(\s*\[([^\]]+)\]`), "generate.py gate centres")
}
