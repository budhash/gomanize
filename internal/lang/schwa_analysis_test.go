package lang

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	newHindi "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestCompareEngines(t *testing.T) {
	newEng := core.NewEngine(newHindi.Hindi{}, colloquial.Colloquial{})
	oldEng := Hindi{}

	file, err := os.Open("../../testbed/dakshina/native_hindi.tsv")
	if err != nil {
		t.Skip("Dakshina test data not available")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	oldMatch, newMatch, total := 0, 0, 0
	oldOnly, newOnly := 0, 0

	type regression struct {
		word, oldGot, newGot, expected string
	}
	var regressions []regression

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) < 2 {
			continue
		}
		devanagari := parts[0]
		expected := parts[1]
		total++

		oldGot := oldEng.Transliterate(devanagari)
		newGot := newEng.Transliterate(devanagari)

		oldOk := oldGot == expected
		newOk := newGot == expected

		if oldOk {
			oldMatch++
		}
		if newOk {
			newMatch++
		}

		if oldOk && !newOk {
			oldOnly++
			if len(regressions) < 30 {
				regressions = append(regressions, regression{devanagari, oldGot, newGot, expected})
			}
		}
		if !oldOk && newOk {
			newOnly++
		}
	}

	t.Logf("Total: %d", total)
	t.Logf("Old engine: %d (%.1f%%)", oldMatch, float64(oldMatch)/float64(total)*100)
	t.Logf("New engine: %d (%.1f%%)", newMatch, float64(newMatch)/float64(total)*100)
	t.Logf("Old only (regressions): %d", oldOnly)
	t.Logf("New only (improvements): %d", newOnly)

	if len(regressions) > 0 {
		t.Log("\nRegressions (old correct, new wrong):")
		for _, r := range regressions {
			t.Logf("  %s: old=%q new=%q expected=%q", r.word, r.oldGot, r.newGot, r.expected)
		}
	}
}
