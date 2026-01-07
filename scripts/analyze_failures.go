//go:build ignore
// +build ignore

// This script analyzes all failures in detail for manual review.
// Run with: go run scripts/analyze_failures.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/budhash/gomanize/core"
	hindi "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

type FailureInfo struct {
	Hindi    string
	Got      string
	Expected string
	Category string
	Notes    string
}

func categorizeFailure(hindi, got, expected string) (string, string) {
	// v vs w issues
	gotNoW := strings.ReplaceAll(got, "w", "v")
	expectedNoW := strings.ReplaceAll(expected, "w", "v")
	if gotNoW == expectedNoW {
		return "V_VS_W", "व mapped to v, Dakshina expects w"
	}

	// Double vowel issues (aa vs a, ee vs i, etc)
	gotNorm := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(got, "aa", "A"), "ee", "E"), "oo", "O")
	expNorm := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(expected, "aa", "A"), "ee", "E"), "oo", "O")
	if gotNorm == expNorm {
		return "VOWEL_LENGTH", "Long vowel difference (aa/a, ee/i, oo/u)"
	}

	// Word-final 'a' missing
	if got+"a" == expected {
		return "MISSING_FINAL_A", "Missing final schwa"
	}

	// ein vs en pattern (में → mein vs men)
	if strings.Contains(got, "ein") && strings.Contains(expected, "en") {
		gotFixed := strings.ReplaceAll(got, "ein", "en")
		if gotFixed == expected {
			return "EIN_VS_EN", "Gomanize uses 'ein', Dakshina uses 'en'"
		}
	}

	// Extra schwa
	if len(got) > len(expected) {
		gotNoA := strings.ReplaceAll(got, "a", "")
		expNoA := strings.ReplaceAll(expected, "a", "")
		if gotNoA == expNoA {
			return "EXTRA_SCHWA", "Gomanize retains schwa that Dakshina deletes"
		}
	}

	// Missing schwa
	if len(got) < len(expected) {
		gotNoA := strings.ReplaceAll(got, "a", "")
		expNoA := strings.ReplaceAll(expected, "a", "")
		if gotNoA == expNoA {
			return "MISSING_SCHWA", "Gomanize deletes schwa that Dakshina retains"
		}
	}

	return "OTHER", "Complex difference requiring manual review"
}

func main() {
	file, err := os.Open("testbed/dakshina/native_hindi.tsv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	engine := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
	var failures []FailureInfo
	passCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		hindiWord := parts[0]
		expected := parts[1]
		// Use override if present
		if len(parts) >= 4 && parts[3] != "" {
			expected = parts[3]
		}

		result := engine.Transliterate(hindiWord)
		if result == expected {
			passCount++
		} else {
			cat, notes := categorizeFailure(hindiWord, result, expected)
			failures = append(failures, FailureInfo{
				Hindi:    hindiWord,
				Got:      result,
				Expected: expected,
				Category: cat,
				Notes:    notes,
			})
		}
	}

	// Sort by category
	sort.Slice(failures, func(i, j int) bool {
		if failures[i].Category != failures[j].Category {
			return failures[i].Category < failures[j].Category
		}
		return failures[i].Hindi < failures[j].Hindi
	})

	// Count by category
	catCounts := make(map[string]int)
	for _, f := range failures {
		catCounts[f.Category]++
	}

	total := passCount + len(failures)
	fmt.Println("# Failure Analysis Report")
	fmt.Println()
	fmt.Printf("- **Total tested**: %d\n", total)
	fmt.Printf("- **Passed**: %d (%.1f%%)\n", passCount, float64(passCount)*100/float64(total))
	fmt.Printf("- **Failed**: %d (%.1f%%)\n", len(failures), float64(len(failures))*100/float64(total))
	fmt.Println()

	// Summary by category
	fmt.Println("## Summary by Category")
	fmt.Println()
	fmt.Println("| Category | Count | % of Failures | Notes |")
	fmt.Println("|----------|-------|---------------|-------|")

	var cats []string
	for c := range catCounts {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		return catCounts[cats[i]] > catCounts[cats[j]]
	})

	for _, cat := range cats {
		count := catCounts[cat]
		pct := float64(count) * 100 / float64(len(failures))
		// Find first failure in this category for notes
		var notes string
		for _, f := range failures {
			if f.Category == cat {
				notes = f.Notes
				break
			}
		}
		fmt.Printf("| %s | %d | %.1f%% | %s |\n", cat, count, pct, notes)
	}
	fmt.Println()

	// Detailed failures by category
	fmt.Println("## Detailed Failures")
	fmt.Println()

	currentCat := ""
	for _, f := range failures {
		if f.Category != currentCat {
			currentCat = f.Category
			fmt.Printf("### %s (%d cases)\n\n", currentCat, catCounts[currentCat])
			fmt.Println("| Hindi | Gomanize | Dakshina |")
			fmt.Println("|-------|----------|----------|")
		}
		fmt.Printf("| %s | %s | %s |\n", f.Hindi, f.Got, f.Expected)
	}
}
