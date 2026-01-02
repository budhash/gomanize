package lang

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// =============================================================================
// INTEGRATION TESTS
// These test against full datasets to measure overall accuracy.
// Run with: go test -run "^TestIntegration" or make test-integration
// =============================================================================

// -----------------------------------------------------------------------------
// Test Data Paths
// -----------------------------------------------------------------------------

func getTestDataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "..", "..", "testbed", filename)
}

func getDakshinaPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "..", "..", "testbed", "dakshina", filename)
}

// -----------------------------------------------------------------------------
// Original Test Suite (hindi-common.txt)
// -----------------------------------------------------------------------------

func TestIntegrationOriginalTestSuite(t *testing.T) {
	filePath := getTestDataPath("hindi-common.txt")
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Original test file not found: %v", err)
		return
	}
	defer file.Close()

	h := Hindi{}
	pass, fail := 0, 0
	var failures []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		input := strings.TrimSpace(parts[0])
		expected := strings.TrimSpace(parts[1])
		// Remove any comments after expected
		if idx := strings.Index(expected, " "); idx > 0 {
			expected = expected[:idx]
		}

		result := h.Transliterate(input)
		if result == expected {
			pass++
		} else {
			fail++
			if len(failures) < 10 {
				failures = append(failures, input+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== Original Test Suite (hindi-common.txt) ===")
	t.Logf("Passed: %d / %d (%.1f%%)", pass, total, pct)

	if len(failures) > 0 {
		t.Logf("Sample failures:")
		for _, f := range failures {
			t.Logf("  %s", f)
		}
	}

	// Current baseline - update as we fix issues
	if pct < 50 {
		t.Errorf("Accuracy %.1f%% is below 50%% baseline", pct)
	}
}

// -----------------------------------------------------------------------------
// Dakshina Dataset Tests (using pre-split test files)
// -----------------------------------------------------------------------------

// TestIntegrationNativeHindi tests against native Hindi words only
func TestIntegrationNativeHindi(t *testing.T) {
	filePath := getDakshinaPath("native_hindi.tsv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Native Hindi test file not found: %v (run split_loanwords.go first)", err)
		return
	}
	defer file.Close()

	h := Hindi{}
	pass, fail := 0, 0
	var failures []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		hindi := parts[0]
		expected := parts[1]
		result := h.Transliterate(hindi)

		if result == expected {
			pass++
		} else {
			fail++
			if len(failures) < 10 {
				failures = append(failures, hindi+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== Native Hindi Results ===")
	t.Logf("Passed: %d / %d (%.1f%%)", pass, total, pct)

	if len(failures) > 0 {
		t.Logf("Sample failures:")
		for _, f := range failures {
			t.Logf("  %s", f)
		}
	}

	// Target: 80% accuracy on native Hindi
	// Current: 82.5% (after व→w conjuncts and long vowel rules)
	if pct < 82 {
		t.Errorf("Native Hindi accuracy %.1f%% is below 82%% threshold", pct)
	}
}

// TestIntegrationEnglishLoanwords tests against English loanwords (for info only)
func TestIntegrationEnglishLoanwords(t *testing.T) {
	filePath := getDakshinaPath("english_loanwords.tsv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("English loanwords test file not found: %v (run split_loanwords.go first)", err)
		return
	}
	defer file.Close()

	h := Hindi{}
	pass, fail := 0, 0
	var failures []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		hindi := parts[0]
		expected := parts[1]
		result := h.Transliterate(hindi)

		if result == expected {
			pass++
		} else {
			fail++
			if len(failures) < 5 {
				failures = append(failures, hindi+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== English Loanwords Results (informational) ===")
	t.Logf("Passed: %d / %d (%.1f%%)", pass, total, pct)
	t.Logf("Note: English loanwords are OUT OF SCOPE for phonetic transliteration")

	if len(failures) > 0 {
		t.Logf("Sample failures (expected - these require English spelling):")
		for _, f := range failures {
			t.Logf("  %s", f)
		}
	}
	// No threshold check - English loanwords are out of scope
}

// TestIntegrationDakshinaAccuracy tests combined accuracy for comparison
func TestIntegrationDakshinaAccuracy(t *testing.T) {
	filePath := getDakshinaPath("all_high_conf.tsv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Dakshina test file not found: %v (run 'make setup-testdata' first)", err)
		return
	}
	defer file.Close()

	h := Hindi{}
	pass, fail := 0, 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		hindi := parts[0]
		expected := parts[1]
		result := h.Transliterate(hindi)

		if result == expected {
			pass++
		} else {
			fail++
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== Combined Dakshina Results ===")
	t.Logf("Passed: %d / %d (%.1f%%)", pass, total, pct)
	t.Logf("Note: Includes English loanwords which skew results lower")
}

// -----------------------------------------------------------------------------
// Failure Pattern Analysis
// -----------------------------------------------------------------------------

type FailureInfo struct {
	Hindi    string
	Got      string
	Expected string
}

func categorizeFailure(hindi, got, expected string) string {
	// v vs w issues
	gotNoW := strings.ReplaceAll(got, "w", "v")
	expectedNoW := strings.ReplaceAll(expected, "w", "v")
	if gotNoW == expectedNoW {
		return "V_VS_W"
	}

	// Double vowel issues
	gotNorm := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(got, "aa", "A"), "ee", "E"), "oo", "O")
	expNorm := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(expected, "aa", "A"), "ee", "E"), "oo", "O")
	if gotNorm == expNorm {
		return "VOWEL_LENGTH"
	}

	// Word-final 'a' missing
	if got+"a" == expected {
		return "MISSING_FINAL_A"
	}

	// Extra schwa
	if len(got) > len(expected) {
		gotNoA := strings.ReplaceAll(got, "a", "")
		expNoA := strings.ReplaceAll(expected, "a", "")
		if gotNoA == expNoA {
			return "EXTRA_SCHWA"
		}
	}

	// Missing schwa
	if len(got) < len(expected) {
		gotNoA := strings.ReplaceAll(got, "a", "")
		expNoA := strings.ReplaceAll(expected, "a", "")
		if gotNoA == expNoA {
			return "MISSING_SCHWA"
		}
	}

	return "OTHER"
}

func TestIntegrationFailureAnalysis(t *testing.T) {
	filePath := getDakshinaPath("native_hindi.tsv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Native Hindi test file not found: %v (run split_loanwords.go first)", err)
		return
	}
	defer file.Close()

	h := Hindi{}
	patterns := make(map[string][]FailureInfo)
	passCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		hindi := parts[0]
		expected := parts[1]

		result := h.Transliterate(hindi)
		if result == expected {
			passCount++
		} else {
			cat := categorizeFailure(hindi, result, expected)
			patterns[cat] = append(patterns[cat], FailureInfo{hindi, result, expected})
		}
	}

	// Sort by count
	var cats []string
	for c := range patterns {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		return len(patterns[cats[i]]) > len(patterns[cats[j]])
	})

	totalFail := 0
	for _, c := range cats {
		totalFail += len(patterns[c])
	}
	total := passCount + totalFail

	t.Logf("")
	t.Logf("========================================")
	t.Logf("FAILURE PATTERN ANALYSIS")
	t.Logf("========================================")
	t.Logf("Total tested: %d", total)
	t.Logf("Passed: %d (%.1f%%)", passCount, float64(passCount)*100/float64(total))
	t.Logf("Failed: %d (%.1f%%)", totalFail, float64(totalFail)*100/float64(total))
	t.Logf("")

	for _, cat := range cats {
		failures := patterns[cat]
		pct := float64(len(failures)) * 100 / float64(totalFail)
		t.Logf("## %s: %d failures (%.1f%%)", cat, len(failures), pct)

		// Show examples
		for i, f := range failures {
			if i >= 3 {
				break
			}
			t.Logf("   %s → %s (expected: %s)", f.Hindi, f.Got, f.Expected)
		}
		t.Logf("")
	}
}

// -----------------------------------------------------------------------------
// Benchmark
// -----------------------------------------------------------------------------

func BenchmarkTransliterate(b *testing.B) {
	h := Hindi{}
	words := []string{
		"नमस्ते",
		"भारत",
		"अंतर्राष्ट्रीय",
		"प्रधानमंत्री",
		"विश्वविद्यालय",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range words {
			h.Transliterate(w)
		}
	}
}

func BenchmarkTransliterateLong(b *testing.B) {
	h := Hindi{}
	sentence := "यह एक लंबा वाक्य है जिसमें कई शब्द हैं और इसे अनुवाद करना होगा"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		words := strings.Split(sentence, " ")
		for _, w := range words {
			h.Transliterate(w)
		}
	}
}
