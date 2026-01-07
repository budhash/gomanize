//go:build ignore

// Deep analysis of Aksharantar failures
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

type Failure struct {
	Native   string
	Got      string
	Expected string
	Category string
	SubCat   string
}

func hasEnglishChars(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func isAcronymLike(s string) bool {
	// Check for patterns like पीएच, डब्ल्यू, etc.
	acronymParts := []string{"एच", "डब्ल्यू", "सी", "एस", "आर", "एन", "पी", "बी", "टी", "डी", "एम", "एल", "के", "जी", "एफ"}
	count := 0
	for _, part := range acronymParts {
		if strings.Contains(s, part) {
			count++
		}
	}
	return count >= 2
}

func hasNukta(s string) bool {
	return strings.ContainsRune(s, '़') ||
		strings.ContainsAny(s, "क़ख़ग़ज़ड़ढ़फ़य़")
}

func categorize(native, got, expected string) (string, string) {
	gotLower := strings.ToLower(got)
	expLower := strings.ToLower(expected)

	if gotLower == expLower {
		return "MATCH", ""
	}

	// Check for acronyms first
	if isAcronymLike(native) {
		return "ACRONYM", "Hindi acronym spelling"
	}

	// Check for English loanwords (expected has English-style spelling)
	englishPatterns := []string{"tion", "sion", "ght", "ous", "ious", "ness", "ment", "able", "ible", "ist", "ism"}
	for _, pat := range englishPatterns {
		if strings.Contains(expLower, pat) && !strings.Contains(gotLower, pat) {
			return "ENGLISH_LOANWORD", "English spelling expected"
		}
	}

	// Nukta handling (फ़ vs फ, ज़ vs ज)
	if hasNukta(native) {
		gotNoNukta := strings.ReplaceAll(gotLower, "z", "j")
		gotNoNukta = strings.ReplaceAll(gotNoNukta, "f", "ph")
		if gotNoNukta == expLower || strings.ReplaceAll(expLower, "z", "j") == gotLower {
			return "NUKTA_HANDLING", "ज़/फ़ variation"
		}
	}

	// v vs w - check if only difference
	gotV := strings.ReplaceAll(gotLower, "w", "v")
	expV := strings.ReplaceAll(expLower, "w", "v")
	if gotV == expV {
		if strings.Count(gotLower, "w") > strings.Count(expLower, "w") {
			return "V_VS_W", "We use 'w', they use 'v'"
		}
		return "V_VS_W", "We use 'v', they use 'w'"
	}

	// aa vs a variations
	// Normalize both to single 'a' and compare
	gotA := strings.ReplaceAll(gotLower, "aa", "A")
	expA := strings.ReplaceAll(expLower, "aa", "A")
	if gotA == expA {
		if strings.Count(gotLower, "aa") > strings.Count(expLower, "aa") {
			return "LONG_VOWEL_AA", "We use 'aa', they use 'a'"
		}
		return "LONG_VOWEL_AA", "We use 'a', they use 'aa'"
	}

	// ee vs i variations
	gotI := strings.ReplaceAll(strings.ReplaceAll(gotLower, "ee", "I"), "i", "I")
	expI := strings.ReplaceAll(strings.ReplaceAll(expLower, "ee", "I"), "i", "I")
	if gotI == expI {
		if strings.Contains(expLower, "ee") && !strings.Contains(gotLower, "ee") {
			return "LONG_VOWEL_II", "We use 'i', they use 'ee'"
		}
		return "LONG_VOWEL_II", "We use 'ee', they use 'i'"
	}

	// oo vs u variations
	gotU := strings.ReplaceAll(strings.ReplaceAll(gotLower, "oo", "U"), "u", "U")
	expU := strings.ReplaceAll(strings.ReplaceAll(expLower, "oo", "U"), "u", "U")
	if gotU == expU {
		if strings.Contains(expLower, "oo") && !strings.Contains(gotLower, "oo") {
			return "LONG_VOWEL_UU", "We use 'u', they use 'oo'"
		}
		return "LONG_VOWEL_UU", "We use 'oo', they use 'u'"
	}

	// ph vs f
	gotF := strings.ReplaceAll(gotLower, "ph", "f")
	if gotF == expLower {
		return "PH_VS_F", "We use 'ph', they use 'f'"
	}
	gotPh := strings.ReplaceAll(gotLower, "f", "ph")
	if gotPh == expLower {
		return "PH_VS_F", "We use 'f', they use 'ph'"
	}

	// chh vs ch
	if strings.ReplaceAll(gotLower, "chh", "ch") == expLower {
		return "CHH_VS_CH", "We use 'chh', they use 'ch'"
	}

	// Schwa analysis - remove all 'a' and compare consonant skeleton
	gotNoA := strings.ReplaceAll(gotLower, "a", "")
	expNoA := strings.ReplaceAll(expLower, "a", "")
	if gotNoA == expNoA {
		if len(gotLower) > len(expLower) {
			return "EXTRA_SCHWA", "We retain more schwas"
		}
		return "MISSING_SCHWA", "We delete more schwas"
	}

	// ड़/ढ़ → d vs r
	if strings.Contains(native, "ड़") || strings.Contains(native, "ढ़") {
		gotD := strings.ReplaceAll(gotLower, "r", "d")
		gotD = strings.ReplaceAll(gotD, "rh", "dh")
		if gotD == expLower {
			return "NUKTA_DR", "ड़/ढ़: we use 'd', they use 'r'"
		}
		expD := strings.ReplaceAll(expLower, "r", "d")
		expD = strings.ReplaceAll(expD, "rh", "dh")
		if gotLower == expD {
			return "NUKTA_DR", "ड़/ढ़: we use 'r', they use 'd'"
		}
	}

	// Check for doubled consonants
	if strings.Contains(expLower, "dd") || strings.Contains(expLower, "tt") ||
		strings.Contains(expLower, "ll") || strings.Contains(expLower, "mm") ||
		strings.Contains(expLower, "nn") || strings.Contains(expLower, "pp") {
		return "GEMINATION", "Consonant doubling difference"
	}

	// Specific ending patterns
	if strings.HasSuffix(expLower, "ey") && strings.HasSuffix(gotLower, "e") {
		return "ENDING_PATTERN", "Word-final 'e' vs 'ey'"
	}
	if strings.HasSuffix(expLower, "ay") && strings.HasSuffix(gotLower, "e") {
		return "ENDING_PATTERN", "Word-final 'e' vs 'ay'"
	}

	// Multiple differences - complex
	diffCount := 0
	if gotV != expV {
		diffCount++
	}
	if strings.Count(gotLower, "aa") != strings.Count(expLower, "aa") {
		diffCount++
	}
	if strings.Contains(gotLower, "ee") != strings.Contains(expLower, "ee") {
		diffCount++
	}
	if strings.Contains(gotLower, "oo") != strings.Contains(expLower, "oo") {
		diffCount++
	}
	if len(gotLower) != len(expLower) {
		diffCount++
	}

	if diffCount >= 3 {
		return "COMPLEX_MULTI", "Multiple differences combined"
	}

	return "OTHER", "Uncategorized"
}

func main() {
	file, err := os.Open("datasets/aksharantar/hin_test.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	engine := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})

	var failures []Failure
	passCount := 0
	total := 0

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		result := engine.Transliterate(entry.Native)
		total++

		if strings.ToLower(result) == strings.ToLower(entry.English) {
			passCount++
		} else {
			cat, subcat := categorize(entry.Native, result, entry.English)
			failures = append(failures, Failure{
				Native:   entry.Native,
				Got:      result,
				Expected: entry.English,
				Category: cat,
				SubCat:   subcat,
			})
		}
	}

	// Count by category
	catCounts := make(map[string]int)
	catExamples := make(map[string][]Failure)
	subcatCounts := make(map[string]map[string]int)

	for _, f := range failures {
		catCounts[f.Category]++
		if len(catExamples[f.Category]) < 15 {
			catExamples[f.Category] = append(catExamples[f.Category], f)
		}
		if subcatCounts[f.Category] == nil {
			subcatCounts[f.Category] = make(map[string]int)
		}
		subcatCounts[f.Category][f.SubCat]++
	}

	// Sort by count
	var cats []string
	for c := range catCounts {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		return catCounts[cats[i]] > catCounts[cats[j]]
	})

	fmt.Println("# Aksharantar Deep Failure Analysis")
	fmt.Println()
	fmt.Printf("**Total**: %d entries\n", total)
	fmt.Printf("**Passed**: %d (%.1f%%)\n", passCount, float64(passCount)*100/float64(total))
	fmt.Printf("**Failed**: %d (%.1f%%)\n", len(failures), float64(len(failures))*100/float64(total))
	fmt.Println()

	fmt.Println("## Categories Summary")
	fmt.Println()
	fmt.Println("| Category | Count | % of Failures | Actionable? |")
	fmt.Println("|----------|-------|---------------|-------------|")

	for _, cat := range cats {
		count := catCounts[cat]
		pct := float64(count) * 100 / float64(len(failures))
		action := ""
		switch cat {
		case "ACRONYM":
			action = "Out of scope"
		case "ENGLISH_LOANWORD":
			action = "Out of scope"
		case "V_VS_W":
			action = "Style preference"
		case "LONG_VOWEL_AA":
			action = "Style preference"
		case "LONG_VOWEL_II":
			action = "Style preference"
		case "LONG_VOWEL_UU":
			action = "Style preference"
		case "PH_VS_F":
			action = "Nukta dependent"
		case "EXTRA_SCHWA":
			action = "Rule tuning"
		case "MISSING_SCHWA":
			action = "Rule tuning"
		case "NUKTA_DR":
			action = "Design choice"
		case "NUKTA_HANDLING":
			action = "Nukta detection"
		case "GEMINATION":
			action = "Style preference"
		case "COMPLEX_MULTI":
			action = "Multiple issues"
		default:
			action = "Review needed"
		}
		fmt.Printf("| %s | %d | %.1f%% | %s |\n", cat, count, pct, action)
	}

	fmt.Println()
	fmt.Println("## Detailed Examples by Category")

	for _, cat := range cats {
		examples := catExamples[cat]
		fmt.Printf("\n### %s (%d cases)\n\n", cat, catCounts[cat])

		// Show subcategory breakdown if exists
		if subs := subcatCounts[cat]; len(subs) > 1 {
			fmt.Println("**Subcategories:**")
			var subKeys []string
			for k := range subs {
				subKeys = append(subKeys, k)
			}
			sort.Slice(subKeys, func(i, j int) bool {
				return subs[subKeys[i]] > subs[subKeys[j]]
			})
			for _, sk := range subKeys {
				if sk != "" {
					fmt.Printf("- %s: %d\n", sk, subs[sk])
				}
			}
			fmt.Println()
		}

		fmt.Println("| Hindi | Gomanize | Aksharantar |")
		fmt.Println("|-------|----------|-------------|")
		for _, f := range examples {
			fmt.Printf("| %s | %s | %s |\n", f.Native, f.Got, f.Expected)
		}
	}

	// Summary insights
	fmt.Println()
	fmt.Println("## Key Insights")
	fmt.Println()

	stylePrefs := catCounts["V_VS_W"] + catCounts["LONG_VOWEL_AA"] + catCounts["LONG_VOWEL_II"] +
		catCounts["LONG_VOWEL_UU"] + catCounts["PH_VS_F"] + catCounts["GEMINATION"] + catCounts["CHH_VS_CH"]
	outOfScope := catCounts["ACRONYM"] + catCounts["ENGLISH_LOANWORD"]
	schwaIssues := catCounts["EXTRA_SCHWA"] + catCounts["MISSING_SCHWA"]

	fmt.Printf("- **Style Preferences** (v/w, aa/a, ee/i, oo/u, ph/f): %d (%.1f%%)\n",
		stylePrefs, float64(stylePrefs)*100/float64(len(failures)))
	fmt.Printf("- **Out of Scope** (acronyms, English loanwords): %d (%.1f%%)\n",
		outOfScope, float64(outOfScope)*100/float64(len(failures)))
	fmt.Printf("- **Schwa Rules** (extra/missing): %d (%.1f%%)\n",
		schwaIssues, float64(schwaIssues)*100/float64(len(failures)))
	fmt.Printf("- **Complex/Other**: %d (%.1f%%)\n",
		catCounts["COMPLEX_MULTI"]+catCounts["OTHER"],
		float64(catCounts["COMPLEX_MULTI"]+catCounts["OTHER"])*100/float64(len(failures)))
}
