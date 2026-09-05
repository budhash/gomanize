package benchmark

import (
	"encoding/csv"
	"fmt"
	"os"

	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	gomanize "github.com/budhash/gomanize"

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

	// Mean CER against the single curated reference (credits near-misses).
	var cerSum float64
	var cerN int
	for _, entry := range entries {
		if entry.Ignored {
			continue
		}
		cerSum += cer(engine.Transliterate(entry.Native), entry.Roman)
		cerN++
	}
	if cerN > 0 {
		t.Logf("Mean CER (single-ref):  %.4f", cerSum/float64(cerN))
	}

	// Gate on PURE accuracy — overrides are an exception lexicon, not engine skill,
	// so the honest headline number is what must not regress. See docs/PROCESS.md.
	const pureThreshold = 85.0
	if purePct < pureThreshold {
		t.Errorf("Pure accuracy %.1f%% is below %.1f%% threshold", purePct, pureThreshold)
	}
}

// -----------------------------------------------------------------------------
// Multi-reference evaluation (benchmark/data/curated_hi.csv scored against the
// full set of attested Dakshina romanizations, not a single gold string).
// -----------------------------------------------------------------------------

// TestBenchmarkMultiReference scores the curated set against ALL attested human
// romanizations per word (from dakshina_hi.csv), reflecting that romanization is
// many-to-one. Reports strict top-1 (single ref), match-any-attested-variant,
// and mean minCER. This is the honest measure of correctness; the gap between
// strict and any-hit is the benchmark artifact, not engine error.
func TestBenchmarkMultiReference(t *testing.T) {
	curatedPath := getTestDataPath("curated_hi.csv")
	refPath := getTestDataPath("dakshina_hi.csv")
	ignorePath := getTestDataPath("ignore_hi.csv")

	entries, err := loadCSV(curatedPath)
	if err != nil {
		t.Skipf("Curated Hindi test file not found: %v", err)
		return
	}
	refs, err := loadReferenceSets(refPath)
	if err != nil {
		t.Skipf("Dakshina reference file not found: %v", err)
		return
	}
	ignores, err := loadIgnores(ignorePath)
	if err != nil {
		t.Fatalf("loading ignores: %v", err)
	}

	engine := newEngine()
	strictPass, anyPass, total := 0, 0, 0
	withRefs, multiRef := 0, 0
	var minCERSum float64
	var recovered []string // matched a variant but not the single curated gold

	for _, entry := range entries {
		if ignores[entry.Native] {
			continue
		}
		total++

		// Reference set: the curated gold plus every attested Dakshina variant.
		refSet := []string{entry.Roman}
		seen := map[string]bool{entry.Roman: true}
		for _, r := range refs[entry.Native] {
			if !seen[r] {
				seen[r] = true
				refSet = append(refSet, r)
			}
		}
		if len(refs[entry.Native]) > 0 {
			withRefs++
		}
		if len(refSet) > 1 {
			multiRef++
		}

		result := engine.Transliterate(entry.Native)
		strict := result == entry.Roman
		if strict {
			strictPass++
		}
		if matchesAny(result, refSet) {
			anyPass++
			if !strict && len(recovered) < 10 {
				recovered = append(recovered, entry.Native+" → "+result+" (curated gold: "+entry.Roman+")")
			}
		}
		minCERSum += minCER(result, refSet)
	}

	if total == 0 {
		t.Skip("no entries to evaluate")
	}
	strictPct := float64(strictPass) * 100 / float64(total)
	anyPct := float64(anyPass) * 100 / float64(total)

	t.Logf("=== Multi-reference Results (curated set) ===")
	t.Logf("Words evaluated:            %d (%d have Dakshina refs, %d have >1 valid spelling)", total, withRefs, multiRef)
	t.Logf("Strict top-1 (single ref):  %d / %d (%.1f%%)", strictPass, total, strictPct)
	t.Logf("Match-any attested variant: %d / %d (%.1f%%)", anyPass, total, anyPct)
	t.Logf("Mean minCER (over variants):%.4f", minCERSum/float64(total))
	t.Logf("Benchmark-artifact gap:     +%.1f pts recovered by multi-reference", anyPct-strictPct)

	if len(recovered) > 0 {
		t.Logf("")
		t.Logf("Sample words scored wrong by single-ref but matching a valid variant:")
		for _, r := range recovered {
			t.Logf("  %s", r)
		}
	}

	// Sanity: multi-reference can only help, never hurt.
	if anyPass < strictPass {
		t.Errorf("match-any (%d) < strict (%d): impossible, reference set bug", anyPass, strictPass)
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
// Held-out schwa-model evaluation (Dakshina test split, disjoint from train).
// -----------------------------------------------------------------------------

// referenceSetsForSplit builds native -> attested variants for rows whose notes
// contain the given "split=<name>" marker.
func referenceSetsForSplit(path, split string) (map[string][]string, error) {
	entries, err := loadCSV(path)
	if err != nil {
		return nil, err
	}
	refs := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	marker := "split=" + split
	for _, e := range entries {
		if e.Roman == "" || !strings.Contains(e.Notes, marker) {
			continue
		}
		if seen[e.Native] == nil {
			seen[e.Native] = make(map[string]bool)
		}
		if !seen[e.Native][e.Roman] {
			seen[e.Native][e.Roman] = true
			refs[e.Native] = append(refs[e.Native], e.Roman)
		}
	}
	return refs, nil
}

// TestBenchmarkSchwaModelHeldout compares the heuristic schwa rules against the
// learned schwa model on the Dakshina TEST split — whose native words are
// disjoint from the TRAIN split the model was trained on. This is a genuine
// generalization test (unlike the curated set, which is all training data).
func TestBenchmarkSchwaModelHeldout(t *testing.T) {
	refs, err := referenceSetsForSplit(getTestDataPath("dakshina_hi.csv"), "test")
	if err != nil {
		t.Skipf("Dakshina file not found: %v", err)
		return
	}
	if len(refs) == 0 {
		t.Skip("no test-split rows")
	}

	engine := newEngine()
	rulesAny, modelAny, total := 0, 0, 0
	var rulesCER, modelCER float64
	var wins, losses []string

	for native, variants := range refs {
		total++
		rulesOut := engine.Transliterate(native)
		modelOut := engine.TransliterateWithOptions(native, core.Options{SchwaModel: true})

		rHit := matchesAny(rulesOut, variants)
		mHit := matchesAny(modelOut, variants)
		if rHit {
			rulesAny++
		}
		if mHit {
			modelAny++
		}
		rulesCER += minCER(rulesOut, variants)
		modelCER += minCER(modelOut, variants)

		if mHit && !rHit && len(wins) < 10 {
			wins = append(wins, native+": rules="+rulesOut+" model="+modelOut+" ✓")
		}
		if rHit && !mHit && len(losses) < 10 {
			losses = append(losses, native+": rules="+rulesOut+" ✓ model="+modelOut)
		}
	}

	rp := float64(rulesAny) * 100 / float64(total)
	mp := float64(modelAny) * 100 / float64(total)
	t.Logf("=== Schwa model vs rules — Dakshina TEST split (held-out, %d words) ===", total)
	t.Logf("Match-any:  rules %.1f%%  |  model %.1f%%  (delta %+.1f pts)", rp, mp, mp-rp)
	t.Logf("Mean minCER: rules %.4f  |  model %.4f", rulesCER/float64(total), modelCER/float64(total))
	t.Logf("Model fixes %d words rules got wrong; regresses %d words rules got right", len(wins), len(losses))
	for _, w := range wins {
		t.Logf("  + %s", w)
	}
	for _, l := range losses {
		t.Logf("  - %s", l)
	}
}

// TestBenchmarkLexiconCoverage reports the lexicon's coverage HONESTLY. Because
// Dakshina's train/test native words are disjoint, a train-built lexicon covers
// ~0% of the held-out test split — so it cannot (and must not) inflate the
// held-out accuracy number. Its real value is production token coverage of common
// vocabulary, which a type-disjoint benchmark structurally cannot credit. This
// test asserts that truth rather than a misleading accuracy jump.
func TestBenchmarkLexiconCoverage(t *testing.T) {
	testRefs, err := referenceSetsForSplit(getTestDataPath("dakshina_hi.csv"), "test")
	if err != nil {
		t.Skipf("Dakshina file not found: %v", err)
		return
	}
	engine := newEngine()

	covered, total := 0, 0
	rulesAny, lexAny := 0, 0
	for native, variants := range testRefs {
		total++
		if _, ok := (hindi.Hindi{}).LexiconLookup(native); ok {
			covered++
		}
		if matchesAny(engine.Transliterate(native), variants) {
			rulesAny++
		}
		if matchesAny(engine.TransliterateWithOptions(native, core.Options{Lexicon: true}), variants) {
			lexAny++
		}
	}

	t.Logf("=== Lexicon coverage — Dakshina TEST split (held-out) ===")
	t.Logf("Lexicon size: %d entries (built from TRAIN)", hindi.LexiconSize())
	t.Logf("Test words covered by lexicon: %d / %d (%.1f%%) — near-zero is EXPECTED (disjoint splits)",
		covered, total, float64(covered)*100/float64(total))
	t.Logf("Match-any: rules %.1f%%  |  rules+lexicon %.1f%%", float64(rulesAny)*100/float64(total), float64(lexAny)*100/float64(total))
	t.Logf("Takeaway: the lexicon does not change held-out TYPE accuracy (by construction);")
	t.Logf("its value is production TOKEN coverage of common words, not measurable here.")

	// The lexicon must never hurt: rules+lexicon >= rules on any set.
	if lexAny < rulesAny {
		t.Errorf("lexicon regressed held-out match-any: rules=%d lexicon=%d", rulesAny, lexAny)
	}
}

// -----------------------------------------------------------------------------
// Aksharantar human-annotated test set (benchmark/data/aksharantar_test_hi.csv)
// -----------------------------------------------------------------------------

// TestBenchmarkAksharantarTestSet scores the Aksharantar Hindi TEST set —
// 10,112 human-annotated pairs (native-speaker annotators via Karya, 2022,
// CC-BY 4.0) — per slice: AK-Freq (frequent words, closest to the song-lyrics
// register), AK-NEF/AK-NEI (foreign/Indian named entities, the known weak spot),
// and a re-included Dakshina test slice. Match-any over the attested variants.
//
// The gold here was annotated independently of Dakshina train, so scoring
// rules+lexicon is legitimate even where the lexicon knows a word; rules-only
// remains the measure of engine skill.
func TestBenchmarkAksharantarTestSet(t *testing.T) {
	entries, err := loadCSV(getTestDataPath("aksharantar_test_hi.csv"))
	if err != nil {
		t.Skipf("Aksharantar test set not found: %v (run tools/build_aksharantar_test.py)", err)
		return
	}

	// Group attested variants per (slice, native).
	type key struct{ slice, native string }
	variants := make(map[key][]string)
	for _, e := range entries {
		slice := strings.TrimPrefix(e.Notes, "source=")
		k := key{slice, e.Native}
		variants[k] = append(variants[k], e.Roman)
	}

	engine := newEngine()
	type stat struct {
		total, rulesHit, lexHit int
		rulesCER, lexCER        float64
		inLexicon               int
	}
	stats := make(map[string]*stat)
	for k, refs := range variants {
		s := stats[k.slice]
		if s == nil {
			s = &stat{}
			stats[k.slice] = s
		}
		s.total++
		rulesOut := engine.Transliterate(k.native)
		lexOut := engine.TransliterateWithOptions(k.native, core.Options{Lexicon: true})
		if matchesAny(rulesOut, refs) {
			s.rulesHit++
		}
		if matchesAny(lexOut, refs) {
			s.lexHit++
		}
		s.rulesCER += minCER(rulesOut, refs)
		s.lexCER += minCER(lexOut, refs)
		if _, ok := (hindi.Hindi{}).LexiconLookup(k.native); ok {
			s.inLexicon++
		}
	}

	slices := make([]string, 0, len(stats))
	for s := range stats {
		slices = append(slices, s)
	}
	sort.Strings(slices)

	t.Logf("=== Aksharantar Hindi test set (human-annotated, match-any) ===")
	for _, name := range slices {
		s := stats[name]
		t.Logf("%-9s %4d words: rules %5.1f%% (minCER %.3f) | +lexicon %5.1f%% (minCER %.3f) | %d in lexicon",
			name, s.total,
			float64(s.rulesHit)*100/float64(s.total), s.rulesCER/float64(s.total),
			float64(s.lexHit)*100/float64(s.total), s.lexCER/float64(s.total),
			s.inLexicon)
	}

	// Lexicon must never hurt on any slice.
	for name, s := range stats {
		if s.lexHit < s.rulesHit {
			t.Errorf("slice %s: lexicon regressed match-any (rules=%d, lexicon=%d)", name, s.rulesHit, s.lexHit)
		}
	}
}

// -----------------------------------------------------------------------------
// Lyrics gold seed (benchmark/data/lyrics_gold_hi.csv) — the flagship use case.
// -----------------------------------------------------------------------------

// TestBenchmarkLyricsGold evaluates LINE-LEVEL romanization on a small
// public-domain lyrics/verse gold set (Kabir, Rahim, Meera, Vande Mataram,
// Sarfaroshi, Raghupati, Jana Gana Mana — all PD under Indian copyright,
// life+60). Romanizations are maintainer-attested single references, so CER is
// the primary metric (exact-line match is a strict secondary). Uses the public
// sentence API (gomanize.Translit) — this exercises the whitespace/punctuation
// segmentation, not just per-word transliteration.
func TestBenchmarkLyricsGold(t *testing.T) {
	entries, err := loadCSV(getTestDataPath("lyrics_gold_hi.csv"))
	if err != nil {
		t.Skipf("lyrics gold not found: %v", err)
		return
	}

	g, err := gomanize.New("hindi")
	if err != nil {
		t.Fatalf("gomanize.New: %v", err)
	}
	gl, err := gomanize.NewWithOptions("hindi", core.Options{Lexicon: true})
	if err != nil {
		t.Fatalf("gomanize.NewWithOptions: %v", err)
	}

	var lines, rulesExact, lexExact int
	var wordsTotal, rulesWordHit, lexWordHit int
	var rulesCER, lexCER float64
	var worst []string

	for _, e := range entries {
		lines++
		rulesOut := strings.ToLower(g.Translit(e.Native))
		lexOut := strings.ToLower(gl.Translit(e.Native))
		want := strings.ToLower(e.Roman)

		rc, lc := cer(rulesOut, want), cer(lexOut, want)
		rulesCER += rc
		lexCER += lc
		if rulesOut == want {
			rulesExact++
		}
		if lexOut == want {
			lexExact++
		}

		wr, ww := strings.Fields(rulesOut), strings.Fields(want)
		wl := strings.Fields(lexOut)
		for i, w := range ww {
			wordsTotal++
			if i < len(wr) && wr[i] == w {
				rulesWordHit++
			}
			if i < len(wl) && wl[i] == w {
				lexWordHit++
			}
		}
		if rc > 0.12 && len(worst) < 5 {
			worst = append(worst, e.Native+" → "+rulesOut+" (want "+want+")")
		}
	}

	t.Logf("=== Lyrics gold seed (%d public-domain lines, single maintainer-attested ref) ===", lines)
	t.Logf("Line CER (primary):    rules %.4f | +lexicon %.4f", rulesCER/float64(lines), lexCER/float64(lines))
	t.Logf("Word accuracy:         rules %5.1f%% | +lexicon %5.1f%%",
		float64(rulesWordHit)*100/float64(wordsTotal), float64(lexWordHit)*100/float64(wordsTotal))
	t.Logf("Exact lines (strict):  rules %d/%d | +lexicon %d/%d", rulesExact, lines, lexExact, lines)
	for _, w := range worst {
		t.Logf("  worst: %s", w)
	}

	// Regression floor: mean line CER must stay under 0.15 against this seed.
	if meanCER := rulesCER / float64(lines); meanCER > 0.15 {
		t.Errorf("lyrics mean CER %.4f exceeds 0.15 floor", meanCER)
	}
}

// -----------------------------------------------------------------------------
// COMI-LINGUA colloquial benchmark (benchmark/data/comilingua_hi.csv)
// -----------------------------------------------------------------------------

// TestBenchmarkComiLingua scores against word pairs extracted from COMI-LINGUA's
// human-annotated MT split (CC-BY 4.0) — parallel Devanagari/Romanized-Hindi
// sentences showing how people ACTUALLY type Hindi in Latin script. Variants
// carry occurrence counts, so scoring is naturally token-weighted toward the
// colloquial register. This is the project's closest proxy to the song-lyrics
// use case until a lyrics gold set exists (T-0021).
func TestBenchmarkComiLingua(t *testing.T) {
	entries, err := loadCSV(getTestDataPath("comilingua_hi.csv"))
	if err != nil {
		t.Skipf("COMI-LINGUA file not found: %v (run tools/build_comilingua.py)", err)
		return
	}

	type variant struct {
		roman string
		count int64
	}
	byNative := make(map[string][]variant)
	for _, e := range entries {
		var n int64
		if _, err := fmt.Sscanf(e.Notes, "count=%d", &n); err != nil || n <= 0 {
			n = 1
		}
		byNative[e.Native] = append(byNative[e.Native], variant{e.Roman, n})
	}

	engine := newEngine()
	var types, rulesTypeHit, lexTypeHit int
	var tokens, rulesTokHit, lexTokHit int64
	var rulesCER, lexCER float64

	for native, vars := range byNative {
		refs := make([]string, len(vars))
		var weight int64
		for i, v := range vars {
			refs[i] = v.roman
			weight += v.count
		}
		types++
		tokens += weight

		rulesOut := engine.Transliterate(native)
		lexOut := engine.TransliterateWithOptions(native, core.Options{Lexicon: true})
		if matchesAny(rulesOut, refs) {
			rulesTypeHit++
			rulesTokHit += weight
		}
		if matchesAny(lexOut, refs) {
			lexTypeHit++
			lexTokHit += weight
		}
		rulesCER += minCER(rulesOut, refs)
		lexCER += minCER(lexOut, refs)
	}

	t.Logf("=== COMI-LINGUA colloquial benchmark (%d word types, %d token occurrences) ===", types, tokens)
	t.Logf("Type-level match-any:  rules %5.1f%% | +lexicon %5.1f%%",
		float64(rulesTypeHit)*100/float64(types), float64(lexTypeHit)*100/float64(types))
	t.Logf("Token-weighted:        rules %5.1f%% | +lexicon %5.1f%%",
		float64(rulesTokHit)*100/float64(tokens), float64(lexTokHit)*100/float64(tokens))
	t.Logf("Mean minCER:           rules %.4f | +lexicon %.4f",
		rulesCER/float64(types), lexCER/float64(types))

	if lexTokHit < rulesTokHit {
		t.Errorf("lexicon regressed token-weighted match-any: rules=%d lexicon=%d", rulesTokHit, lexTokHit)
	}
}

// -----------------------------------------------------------------------------
// Frequency-weighted real-world evaluation (Track C).
// -----------------------------------------------------------------------------

// TestBenchmarkFrequencyWeighted answers the real-world question the disjoint
// Dakshina splits cannot: "how good is gomanize on the words people actually
// use?" Words are weighted by corpus frequency (Shabd, CC0), so common words
// count more than rare ones. It scores match-any against attested Dakshina
// spellings on the frequency∩gold intersection, and reports the lexicon's
// TOKEN coverage of the frequency distribution (its true production value).
func TestBenchmarkFrequencyWeighted(t *testing.T) {
	freq, err := loadFrequencies(getTestDataPath("freq_hi.csv"))
	if err != nil {
		t.Skipf("freq_hi.csv not found: %v (run tools/build_freq.py)", err)
		return
	}
	refs, err := loadReferenceSets(getTestDataPath("dakshina_hi.csv"))
	if err != nil {
		t.Skipf("Dakshina file not found: %v", err)
		return
	}
	engine := newEngine()

	// (1) Lexicon token coverage over the whole frequency distribution.
	var totalTokens, lexTokens int64
	for w, f := range freq {
		totalTokens += f
		if _, ok := (hindi.Hindi{}).LexiconLookup(w); ok {
			lexTokens += f
		}
	}

	// (2) Frequency-weighted accuracy on words that have attested gold.
	var wTotal, wRulesHit, wLexHit int64
	scored := 0
	for w, f := range freq {
		variants, ok := refs[w]
		if !ok {
			continue // no gold to score against
		}
		scored++
		wTotal += f
		if matchesAny(engine.Transliterate(w), variants) {
			wRulesHit += f
		}
		if matchesAny(engine.TransliterateWithOptions(w, core.Options{Lexicon: true}), variants) {
			wLexHit += f
		}
	}

	t.Logf("=== Frequency-weighted real-world evaluation (Shabd top-%d) ===", len(freq))
	t.Logf("Lexicon token coverage: %.1f%% of running-text tokens (%d/%d words in lexicon by frequency mass)",
		float64(lexTokens)*100/float64(totalTokens), countInLexicon(freq), len(freq))
	if wTotal > 0 {
		t.Logf("Frequency-weighted match-any on %d gold-covered words:", scored)
		t.Logf("  rules:          %.1f%%", float64(wRulesHit)*100/float64(wTotal))
		t.Logf("  rules+lexicon:  %.1f%%", float64(wLexHit)*100/float64(wTotal))
		t.Logf("(Type-level accuracy under-weights common words; this weights by real usage.)")
	}

	if wLexHit < wRulesHit {
		t.Errorf("lexicon regressed frequency-weighted accuracy: rules=%d lexicon=%d", wRulesHit, wLexHit)
	}
}

func countInLexicon(freq map[string]int64) int {
	n := 0
	for w := range freq {
		if _, ok := (hindi.Hindi{}).LexiconLookup(w); ok {
			n++
		}
	}
	return n
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
