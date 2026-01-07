//go:build ignore

// Evaluate gomanize against Aksharantar Hindi test set
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/budhash/gomanize/core"
	hindi "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

type Entry struct {
	ID      string `json:"unique_identifier"`
	Native  string `json:"native word"`
	English string `json:"english word"`
	Source  string `json:"source"`
}

func categorize(hindi, got, expected string) string {
	// Normalize for comparison
	gotLower := strings.ToLower(got)
	expLower := strings.ToLower(expected)

	if gotLower == expLower {
		return "MATCH"
	}

	// v vs w
	if strings.ReplaceAll(gotLower, "w", "v") == strings.ReplaceAll(expLower, "w", "v") {
		return "V_VS_W"
	}

	// aa vs a
	if strings.ReplaceAll(gotLower, "aa", "a") == strings.ReplaceAll(expLower, "aa", "a") {
		return "AA_VS_A"
	}

	// ee vs i
	if strings.ReplaceAll(gotLower, "ee", "i") == strings.ReplaceAll(expLower, "ee", "i") ||
		strings.ReplaceAll(gotLower, "i", "ee") == expLower {
		return "EE_VS_I"
	}

	// oo vs u
	if strings.ReplaceAll(gotLower, "oo", "u") == strings.ReplaceAll(expLower, "oo", "u") ||
		strings.ReplaceAll(gotLower, "u", "oo") == expLower {
		return "OO_VS_U"
	}

	// ph vs f
	if strings.ReplaceAll(gotLower, "ph", "f") == expLower ||
		strings.ReplaceAll(gotLower, "f", "ph") == expLower {
		return "PH_VS_F"
	}

	// Schwa differences (length mismatch often indicates schwa)
	gotNoA := strings.ReplaceAll(gotLower, "a", "")
	expNoA := strings.ReplaceAll(expLower, "a", "")
	if gotNoA == expNoA {
		if len(gotLower) > len(expLower) {
			return "EXTRA_SCHWA"
		}
		return "MISSING_SCHWA"
	}

	return "OTHER"
}

func main() {
	file, err := os.Open("datasets/aksharantar/hin_test.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	engine := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})

	passCount := 0
	total := 0
	categories := make(map[string]int)
	examples := make(map[string][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		result := engine.Transliterate(entry.Native)
		total++

		// Case-insensitive comparison
		if strings.ToLower(result) == strings.ToLower(entry.English) {
			passCount++
			categories["MATCH"]++
		} else {
			cat := categorize(entry.Native, result, entry.English)
			categories[cat]++
			if len(examples[cat]) < 10 {
				examples[cat] = append(examples[cat],
					fmt.Sprintf("%s → %s (expected: %s)", entry.Native, result, entry.English))
			}
		}
	}

	fmt.Println("# Aksharantar Hindi Test Set Evaluation")
	fmt.Println()
	fmt.Printf("**Total**: %d entries\n", total)
	fmt.Printf("**Exact Match**: %d (%.1f%%)\n", passCount, float64(passCount)*100/float64(total))
	fmt.Println()

	// Sort categories by count
	var cats []string
	for c := range categories {
		if c != "MATCH" {
			cats = append(cats, c)
		}
	}
	sort.Slice(cats, func(i, j int) bool {
		return categories[cats[i]] > categories[cats[j]]
	})

	fmt.Println("## Failure Categories")
	fmt.Println()
	fmt.Println("| Category | Count | % of Failures |")
	fmt.Println("|----------|-------|---------------|")
	failures := total - passCount
	for _, cat := range cats {
		count := categories[cat]
		pct := float64(count) * 100 / float64(failures)
		fmt.Printf("| %s | %d | %.1f%% |\n", cat, count, pct)
	}

	fmt.Println()
	fmt.Println("## Examples by Category")
	for _, cat := range cats {
		if exs, ok := examples[cat]; ok && len(exs) > 0 {
			fmt.Printf("\n### %s\n", cat)
			for _, ex := range exs {
				fmt.Printf("- %s\n", ex)
			}
		}
	}
}
