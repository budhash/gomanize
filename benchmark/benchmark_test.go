package benchmark

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// =============================================================================
// BENCHMARK TESTS
// These test against datasets to measure transliteration accuracy.
// Run with: go test -v ./benchmark/...
// =============================================================================

// -----------------------------------------------------------------------------
// Test Data Loading
// -----------------------------------------------------------------------------

func getTestDataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	return filepath.Join(dir, "data", filename)
}

// TestEntry represents a word pair from test data
type TestEntry struct {
	Native   string
	Roman    string
	Notes    string
	Override string // Applied from override file
	Ignored  bool   // True if in ignore file
}

// loadCSV loads a CSV file with format: native,roman,notes
func loadCSV(path string) ([]TestEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []TestEntry
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(record) < 2 {
			continue
		}
		entry := TestEntry{
			Native: record[0],
			Roman:  record[1],
		}
		if len(record) >= 3 {
			entry.Notes = record[2]
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// loadOverrides loads override file and returns a map of native -> roman
func loadOverrides(path string) (map[string]string, error) {
	entries, err := loadCSV(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	overrides := make(map[string]string)
	for _, e := range entries {
		overrides[e.Native] = e.Roman
	}
	return overrides, nil
}

// loadIgnores loads ignore file and returns a set of native words
func loadIgnores(path string) (map[string]bool, error) {
	entries, err := loadCSV(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]bool), nil
		}
		return nil, err
	}

	ignores := make(map[string]bool)
	for _, e := range entries {
		ignores[e.Native] = true
	}
	return ignores, nil
}

// loadTestData loads test data with overrides and ignores applied
func loadTestData(dataPath, overridePath, ignorePath string) ([]TestEntry, error) {
	entries, err := loadCSV(dataPath)
	if err != nil {
		return nil, err
	}

	overrides, err := loadOverrides(overridePath)
	if err != nil {
		return nil, err
	}

	ignores, err := loadIgnores(ignorePath)
	if err != nil {
		return nil, err
	}

	// Apply overrides and ignores
	for i := range entries {
		if override, ok := overrides[entries[i].Native]; ok {
			entries[i].Override = override
		}
		if ignores[entries[i].Native] {
			entries[i].Ignored = true
		}
	}

	return entries, nil
}

// newEngine returns the Hindi engine (lang/hindi with core architecture).
func newEngine() *core.Engine {
	return core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
}

// -----------------------------------------------------------------------------
// Curated Hindi Dataset Test (benchmark/data/curated_hi.csv)
// -----------------------------------------------------------------------------

// TestBenchmarkCuratedHindi tests against curated native Hindi words.
// Uses override_hi.csv and ignore_hi.csv for customization.
func TestBenchmarkCuratedHindi(t *testing.T) {
	dataPath := getTestDataPath("curated_hi.csv")
	overridePath := getTestDataPath("override_hi.csv")
	ignorePath := getTestDataPath("ignore_hi.csv")

	entries, err := loadTestData(dataPath, overridePath, ignorePath)
	if err != nil {
		t.Skipf("Curated Hindi test file not found: %v", err)
		return
	}

	engine := newEngine()
	// Track both modes
	purePass, pureFail := 0, 0
	overridePass, overrideFail := 0, 0
	overrideCount, ignoreCount := 0, 0
	var pureFailures []string
	var overrideFailures []string

	for _, entry := range entries {
		if entry.Ignored {
			ignoreCount++
			continue
		}

		result := engine.Transliterate(entry.Native)

		// Check against pure (original expected)
		if result == entry.Roman {
			purePass++
		} else {
			pureFail++
			if len(pureFailures) < 10 {
				pureFailures = append(pureFailures, entry.Native+" → "+result+" (expected: "+entry.Roman+")")
			}
		}

		// Check with overrides applied
		expected := entry.Roman
		if entry.Override != "" {
			expected = entry.Override
			overrideCount++
		}
		if result == expected {
			overridePass++
		} else {
			overrideFail++
			if len(overrideFailures) < 10 {
				overrideFailures = append(overrideFailures, entry.Native+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	pureTotal := purePass + pureFail
	purePct := float64(purePass) * 100 / float64(pureTotal)

	overrideTotal := overridePass + overrideFail
	overridePct := float64(overridePass) * 100 / float64(overrideTotal)

	t.Logf("=== Curated Hindi Results ===")
	t.Logf("")
	t.Logf("Pure (no overrides):  %d / %d (%.1f%%)", purePass, pureTotal, purePct)
	t.Logf("With overrides:       %d / %d (%.1f%%) [%d overrides, %d ignored]", overridePass, overrideTotal, overridePct, overrideCount, ignoreCount)

	if len(overrideFailures) > 0 {
		t.Logf("")
		t.Logf("Sample failures:")
		for _, f := range overrideFailures {
			t.Logf("  %s", f)
		}
	}

	// Target: 82% accuracy with overrides
	if overridePct < 82 {
		t.Errorf("Accuracy %.1f%% is below 82%% threshold", overridePct)
	}
}

// -----------------------------------------------------------------------------
// Dakshina Dataset Test (benchmark/data/dakshina_hi.csv)
// -----------------------------------------------------------------------------

// TestBenchmarkDakshinaHindi tests against the full Dakshina dataset.
// Uses override_hi.csv and ignore_hi.csv for customization.
func TestBenchmarkDakshinaHindi(t *testing.T) {
	dataPath := getTestDataPath("dakshina_hi.csv")
	overridePath := getTestDataPath("override_hi.csv")
	ignorePath := getTestDataPath("ignore_hi.csv")

	entries, err := loadTestData(dataPath, overridePath, ignorePath)
	if err != nil {
		t.Skipf("Dakshina Hindi test file not found: %v (run datasets/dakshina/generate_dataset.sh hi)", err)
		return
	}

	engine := newEngine()
	pass, fail, ignored := 0, 0, 0
	var failures []string

	for _, entry := range entries {
		if entry.Ignored {
			ignored++
			continue
		}

		result := engine.Transliterate(entry.Native)
		expected := entry.Roman
		if entry.Override != "" {
			expected = entry.Override
		}

		if result == expected {
			pass++
		} else {
			fail++
			if len(failures) < 10 {
				failures = append(failures, entry.Native+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== Dakshina Hindi Results ===")
	t.Logf("Passed: %d / %d (%.1f%%) [%d ignored]", pass, total, pct, ignored)

	if len(failures) > 0 {
		t.Logf("")
		t.Logf("Sample failures:")
		for _, f := range failures {
			t.Logf("  %s", f)
		}
	}
}

// -----------------------------------------------------------------------------
// Aksharantar Dataset Test (testbed/aksharantar_hi.csv)
// -----------------------------------------------------------------------------

// TestBenchmarkAksharantarHindi tests against the Aksharantar dataset.
// Uses override_hi.csv and ignore_hi.csv for customization.
func TestBenchmarkAksharantarHindi(t *testing.T) {
	dataPath := getTestDataPath("aksharantar_hi.csv")
	overridePath := getTestDataPath("override_hi.csv")
	ignorePath := getTestDataPath("ignore_hi.csv")

	entries, err := loadTestData(dataPath, overridePath, ignorePath)
	if err != nil {
		t.Skipf("Aksharantar Hindi test file not found: %v (run datasets/aksharantar/generate_dataset.sh hi)", err)
		return
	}

	engine := newEngine()
	pass, fail, ignored := 0, 0, 0
	var failures []string

	for _, entry := range entries {
		if entry.Ignored {
			ignored++
			continue
		}

		result := engine.Transliterate(entry.Native)
		expected := entry.Roman
		if entry.Override != "" {
			expected = entry.Override
		}

		if result == expected {
			pass++
		} else {
			fail++
			if len(failures) < 10 {
				failures = append(failures, entry.Native+" → "+result+" (expected: "+expected+")")
			}
		}
	}

	total := pass + fail
	pct := float64(pass) * 100 / float64(total)

	t.Logf("=== Aksharantar Hindi Results ===")
	t.Logf("Passed: %d / %d (%.1f%%) [%d ignored]", pass, total, pct, ignored)

	if len(failures) > 0 {
		t.Logf("")
		t.Logf("Sample failures:")
		for _, f := range failures {
			t.Logf("  %s", f)
		}
	}
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

func TestBenchmarkFailureAnalysis(t *testing.T) {
	dataPath := getTestDataPath("curated_hi.csv")
	overridePath := getTestDataPath("override_hi.csv")
	ignorePath := getTestDataPath("ignore_hi.csv")

	entries, err := loadTestData(dataPath, overridePath, ignorePath)
	if err != nil {
		t.Skipf("Curated Hindi test file not found: %v", err)
		return
	}

	engine := newEngine()
	patterns := make(map[string][]FailureInfo)
	passCount := 0

	for _, entry := range entries {
		if entry.Ignored {
			continue
		}

		expected := entry.Roman
		if entry.Override != "" {
			expected = entry.Override
		}

		result := engine.Transliterate(entry.Native)
		if result == expected {
			passCount++
		} else {
			cat := categorizeFailure(entry.Native, result, expected)
			patterns[cat] = append(patterns[cat], FailureInfo{entry.Native, result, expected})
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
	engine := newEngine()
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
			engine.Transliterate(w)
		}
	}
}

func BenchmarkTransliterateLong(b *testing.B) {
	engine := newEngine()
	sentence := "यह एक लंबा वाक्य है जिसमें कई शब्द हैं और इसे अनुवाद करना होगा"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		words := strings.Split(sentence, " ")
		for _, w := range words {
			engine.Transliterate(w)
		}
	}
}
