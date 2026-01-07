//go:build ignore
// +build ignore

// Deep analysis of failure patterns
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

type Failure struct {
	Hindi    string
	Got      string
	Expected string
	Category string
	SubCat   string
}

func categorize(hindi, got, expected string) (string, string) {
	// ी→i vs ee (long vowel)
	if strings.ReplaceAll(got, "i", "ee") == expected ||
		strings.ReplaceAll(got, "ee", "i") == expected {
		if strings.Contains(expected, "ee") && strings.Contains(got, "i") {
			return "VOWEL_II", "ी→i, Dakshina uses ee"
		}
		return "VOWEL_II", "ी→ee, Dakshina uses i"
	}

	// ू→u vs oo
	if strings.ReplaceAll(got, "u", "oo") == expected ||
		strings.ReplaceAll(got, "oo", "u") == expected {
		if strings.Contains(expected, "oo") && strings.Contains(got, "u") {
			return "VOWEL_UU", "ू→u, Dakshina uses oo"
		}
		return "VOWEL_UU", "ू→oo, Dakshina uses u"
	}

	// ड़/ढ़ → r vs d
	if strings.Contains(hindi, "ड़") || strings.Contains(hindi, "ढ़") {
		gotD := strings.ReplaceAll(strings.ReplaceAll(got, "rh", "dh"), "r", "d")
		if gotD == expected {
			return "NUKTA_DR", "ड़/ढ़→r, Dakshina uses d"
		}
	}

	// ं before प/ब/म → n vs m
	if strings.Contains(got, "np") || strings.Contains(got, "nb") || strings.Contains(got, "nm") {
		gotM := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(got, "np", "mp"), "nb", "mb"), "nm", "mm")
		if gotM == expected {
			return "ANUSVARA_M", "ं→n before labial, should be m"
		}
	}
	if strings.Contains(got, "nk") || strings.Contains(got, "ng") {
		// Check if swapping helps - sometimes ं before velar should be ṅ
	}

	// फ → ph vs f
	if strings.Contains(hindi, "फ") {
		gotF := strings.ReplaceAll(got, "ph", "f")
		if gotF == expected {
			return "PH_VS_F", "फ→ph, Dakshina uses f"
		}
	}

	// ए → e vs ai/ay at word end
	if strings.Contains(hindi, "ए") || strings.Contains(hindi, "िए") || strings.Contains(hindi, "ये") {
		if strings.HasSuffix(got, "e") && (strings.HasSuffix(expected, "ay") || strings.HasSuffix(expected, "ey")) {
			return "E_VS_AY", "ए→e, Dakshina uses ay/ey"
		}
		if strings.HasSuffix(got, "ie") && strings.HasSuffix(expected, "iye") {
			return "IE_VS_IYE", "िए→ie, Dakshina uses iye"
		}
	}

	// ां + व pattern (gaanv vs gaon)
	if strings.Contains(hindi, "ांव") || strings.Contains(hindi, "ाँव") {
		return "AANV_VS_AON", "ांव→aanv, Dakshina uses aon"
	}

	// v vs w
	gotNoW := strings.ReplaceAll(got, "w", "v")
	expectedNoW := strings.ReplaceAll(expected, "w", "v")
	if gotNoW == expectedNoW {
		return "V_VS_W", "व→v, Dakshina uses w"
	}

	// Double consonant differences
	if strings.ReplaceAll(got, "chh", "ch") == expected {
		return "CHH_VS_CH", "छ→chh, Dakshina uses ch"
	}

	// Word-final 'a' missing
	if got+"a" == expected {
		return "MISSING_FINAL_A", "Missing final schwa (Sanskrit)"
	}

	// aa vs a differences
	gotAA := strings.ReplaceAll(got, "aa", "ā")
	expAA := strings.ReplaceAll(expected, "aa", "ā")
	if gotAA == expAA {
		return "AA_MISMATCH", "aa placement differs"
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

	return "OTHER", "Complex - manual review needed"
}

func main() {
	file, err := os.Open("testbed/dakshina/native_hindi.tsv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	engine := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
	var failures []Failure
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
		if len(parts) >= 4 && parts[3] != "" {
			expected = parts[3]
		}

		result := engine.Transliterate(hindiWord)
		if result == expected {
			passCount++
		} else {
			cat, subcat := categorize(hindiWord, result, expected)
			failures = append(failures, Failure{
				Hindi:    hindiWord,
				Got:      result,
				Expected: expected,
				Category: cat,
				SubCat:   subcat,
			})
		}
	}

	// Count by category
	catCounts := make(map[string]int)
	catExamples := make(map[string][]Failure)
	for _, f := range failures {
		catCounts[f.Category]++
		if len(catExamples[f.Category]) < 10 {
			catExamples[f.Category] = append(catExamples[f.Category], f)
		}
	}

	// Sort by count
	var cats []string
	for c := range catCounts {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		return catCounts[cats[i]] > catCounts[cats[j]]
	})

	total := passCount + len(failures)
	fmt.Println("# Deep Failure Analysis")
	fmt.Println()
	fmt.Printf("**Passed**: %d / %d (%.1f%%)\n", passCount, total, float64(passCount)*100/float64(total))
	fmt.Printf("**Failed**: %d (%.1f%%)\n", len(failures), float64(len(failures))*100/float64(total))
	fmt.Println()
	fmt.Println("## Categories")
	fmt.Println()
	fmt.Println("| Category | Count | % | Description | Actionable? |")
	fmt.Println("|----------|-------|---|-------------|-------------|")

	for _, cat := range cats {
		count := catCounts[cat]
		pct := float64(count) * 100 / float64(len(failures))
		desc := ""
		action := ""
		if len(catExamples[cat]) > 0 {
			desc = catExamples[cat][0].SubCat
		}

		// Determine if actionable
		switch cat {
		case "VOWEL_II", "VOWEL_UU":
			action = "Preference - could add option"
		case "NUKTA_DR":
			action = "**FIX**: ड़→d not r"
		case "ANUSVARA_M":
			action = "**FIX**: ं→m before labials"
		case "PH_VS_F":
			action = "Preference - फ़ vs फ"
		case "V_VS_W":
			action = "Preference"
		case "AANV_VS_AON":
			action = "**FIX**: ांव→aon pattern"
		case "EXTRA_SCHWA":
			action = "**FIX**: Schwa deletion"
		case "MISSING_SCHWA":
			action = "Mixed - some correct"
		case "MISSING_FINAL_A":
			action = "**FIX**: Sanskrit endings"
		case "IE_VS_IYE", "E_VS_AY":
			action = "**FIX**: ए ending"
		case "CHH_VS_CH":
			action = "Preference"
		default:
			action = "Manual review"
		}

		fmt.Printf("| %s | %d | %.1f%% | %s | %s |\n", cat, count, pct, desc, action)
	}

	fmt.Println()
	fmt.Println("## Examples by Category")
	fmt.Println()

	for _, cat := range cats {
		examples := catExamples[cat]
		fmt.Printf("### %s (%d cases)\n\n", cat, catCounts[cat])
		fmt.Println("| Hindi | Gomanize | Dakshina |")
		fmt.Println("|-------|----------|----------|")
		for _, f := range examples {
			fmt.Printf("| %s | %s | %s |\n", f.Hindi, f.Got, f.Expected)
		}
		fmt.Println()
	}
}
