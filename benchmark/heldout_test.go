package benchmark

// Held-out construct regression set (F-0010, T-0044).
//
// A small, human-validated gold set mined from external corpora (never from any
// benchmark or the Dakshina train split — see docs/reference/candidate-datasets.md)
// and stratified by rare orthographic construct. Its purpose is the opposite of
// curated_hi.csv: the curated/frequency-sampled benchmarks barely cover rare
// constructs (the गई family was ~0.2% of curated), so systematic parser bugs in
// them stay invisible to aggregate accuracy — this is how गई→"gi" (T-0040) and
// चाँद→"chaad" (T-0045) survived. This set gives each target construct explicit
// coverage so those regressions fail loudly here.
//
// Scored match-any (a native may list several accepted spellings as repeated
// rows). Floors are set below current pass rates with headroom; a real
// regression (e.g. the chandrabindu nasal breaking again) collapses a whole
// construct and trips the floor. Known misses (vowel-length aa, schwa retention)
// are recorded, not hidden — they are honest future work, logged below.

import (
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
)

func TestBenchmarkHeldoutConstructs(t *testing.T) {
	path := getTestDataPath("heldout_constructs_hi.csv")

	// native -> accepted Roman variants (match-any)
	refs, err := loadReferenceSets(path)
	if err != nil {
		t.Fatalf("loading held-out set: %v", err)
	}
	if len(refs) == 0 {
		t.Skipf("held-out construct set not found or empty: %s", path)
	}

	// native -> construct class (from the notes column)
	entries, err := loadCSV(path)
	if err != nil {
		t.Fatalf("loading held-out set (classes): %v", err)
	}
	class := make(map[string]string)
	for _, e := range entries {
		for _, tok := range strings.Fields(e.Notes) {
			if v, ok := strings.CutPrefix(tok, "class="); ok {
				class[e.Native] = v
			}
		}
	}

	engine := newEngine()
	best := core.Options{Lexicon: true, Rerank: true}

	type stat struct{ total, defHit, bestHit int }
	per := make(map[string]*stat)
	var defHit, bestHit, total int
	var misses []string

	for native, variants := range refs {
		c := class[native]
		if per[c] == nil {
			per[c] = &stat{}
		}
		per[c].total++
		total++

		dOut := engine.Transliterate(native)
		bOut := engine.TransliterateWithOptions(native, best)
		if matchesAny(dOut, variants) {
			per[c].defHit++
			defHit++
		} else {
			misses = append(misses, native+" -> "+dOut+" (want "+strings.Join(variants, "|")+")")
		}
		if matchesAny(bOut, variants) {
			per[c].bestHit++
			bestHit++
		}
	}

	pct := func(n, d int) float64 {
		if d == 0 {
			return 0
		}
		return 100 * float64(n) / float64(d)
	}

	t.Logf("Held-out constructs: %d words", total)
	t.Logf("  default match-any: %d/%d (%.1f%%)", defHit, total, pct(defHit, total))
	t.Logf("  best    match-any: %d/%d (%.1f%%)", bestHit, total, pct(bestHit, total))
	for c, s := range per {
		t.Logf("  %-20s default %d/%d (%.0f%%)  best %d/%d (%.0f%%)",
			c, s.defHit, s.total, pct(s.defHit, s.total), s.bestHit, s.total, pct(s.bestHit, s.total))
	}
	for _, m := range misses {
		t.Logf("  MISS (recorded gap): %s", m)
	}

	// Regression floors — below current rates (default 97.7%, best 98.4%,
	// chandrabindu 100%; after the T-0048/T-0049 gold corrections) with
	// headroom, so intended small changes pass but a construct-level regression
	// trips them. Remaining misses are 3 honest schwa-on-rare-word gaps (T-0048).
	const overallFloor = 95.0
	if p := pct(defHit, total); p < overallFloor {
		t.Errorf("default match-any %.1f%% below floor %.1f%%", p, overallFloor)
	}
	if p := pct(bestHit, total); p < overallFloor {
		t.Errorf("best match-any %.1f%% below floor %.1f%%", p, overallFloor)
	}
	// Lock the T-0045 chandrabindu fix specifically: if the nasal starts
	// dropping again (चाँद→chaad, साँस→saas) this construct collapses.
	if cb := per["chandrabindu"]; cb != nil {
		const chandrabinduFloor = 95.0
		if p := pct(cb.defHit, cb.total); p < chandrabinduFloor {
			t.Errorf("chandrabindu default match-any %.1f%% below floor %.1f%% "+
				"(T-0045 regression?)", p, chandrabinduFloor)
		}
	}
}
