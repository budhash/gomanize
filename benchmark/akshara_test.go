package benchmark

// Tier 4 of the structured parser-QA plan
// (docs/reviews/2026-09-08-structured-parser-qa.md): differential of gomanize's
// unit segmentation against an independent Devanagari akshara-boundary
// reference (UAX #29-style, implemented in-repo — the module is intentionally
// dependency-free, so no external grapheme-segmentation library).
//
// The reference marks where a new *akshara* (orthographic syllable) begins.
// gomanize splits FINER than aksharas (a matra and each conjunct member get
// their own unit), so the two are not equal — the correctness relation is
// REFINEMENT: every akshara boundary MUST also be a gomanize unit boundary.
// gomanize may add boundaries (matra/conjunct splits); it must never MERGE
// across an akshara boundary. A missing boundary = two syllables fused into one
// unit span with no split between them — the segmentation-level shape of the
// गई-class bug (T-0040). This check is convention-free: it is about *where the
// cuts are*, not how anything romanizes.

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/script/brahmic"
)

// aksharaStarts returns the set of rune indices at which a new akshara begins.
// A rune continues the current akshara (no boundary) when it attaches to the
// left — a matra, nukta, virama, or a nasal/visarga modifier — or when it
// immediately follows a virama (a conjunct's next consonant). Every other rune
// starts a new akshara.
func aksharaStarts(runes []rune) map[int]bool {
	starts := make(map[int]bool)
	for i, c := range runes {
		if i == 0 {
			starts[0] = true
			continue
		}
		switch {
		case isMatra(c), c == cNukta, c == cHalant,
			c == cAnusvara, c == cChandrabindu, c == cVisarga:
			// attaches to the preceding akshara
		case runes[i-1] == cHalant:
			// consonant/vowel right after a virama continues the conjunct
		default:
			starts[i] = true
		}
	}
	return starts
}

func aksharaParse(input string) *core.Word {
	lang := hindi.Hindi{}
	s := brahmic.New()
	p := s.NewParser(lang.ScriptConfig())
	w := p.Parse(input, lang.Symbols())
	s.PrepareWord(w)
	return w
}

// gomanizeStarts returns the set of rune indices at which gomanize begins a new
// unit, computed over w.Original (the parser strips ZWJ/ZWNJ and aligns indices
// to Original, so the akshara reference must use the same string).
func gomanizeStarts(w *core.Word) map[int]bool {
	starts := make(map[int]bool)
	for _, u := range w.Units {
		starts[u.Start.Rune] = true
	}
	return starts
}

// checkRefinement asserts akshara boundaries ⊆ gomanize unit boundaries for one
// word. Returns false (and logs) on a missing boundary.
func checkRefinement(t *testing.T, native string) bool {
	t.Helper()
	w := aksharaParse(native)
	orig := []rune(w.Original)
	ref := aksharaStarts(orig)
	got := gomanizeStarts(w)
	ok := true
	for i := range ref {
		if !got[i] {
			t.Errorf("%q: akshara boundary at rune %d (%q) is not a gomanize unit boundary "+
				"(syllables merged — segmentation bug)", native, i, string(orig[i]))
			ok = false
		}
	}
	return ok
}

// TestAksharaDifferentialSynthetic covers the sharp construct shapes directly,
// including the गई class (consonant + independent vowel must not merge).
func TestAksharaDifferentialSynthetic(t *testing.T) {
	cases := []string{
		"गई", "नई", "कई", "गए", "हुई", "लिए", // consonant + independent vowel
		"गी", "की", "नी", // matra (contrast — must still refine)
		"काम", "नमस्ते", "क्षत्रिय", "ज़रा", "विद्या", // conjuncts, nukta, matra
		"दरअसल", "बताएं", "आईएसएल", // medial independent vowels
		"माँ", "साँप", "कहाँ", "संस्कृत", // nasals + conjuncts
	}
	for _, c := range cases {
		checkRefinement(t, c)
	}
}

// TestAksharaDifferentialCorpus runs the refinement check across the full
// Dakshina native inventory — thousands of real words — so any systematic
// segmentation divergence surfaces without manual sifting.
func TestAksharaDifferentialCorpus(t *testing.T) {
	entries, err := loadCSV(getTestDataPath("dakshina_hi.csv"))
	if err != nil {
		t.Skipf("dakshina data not found: %v", err)
	}

	seen := make(map[string]bool)
	checked, failed := 0, 0
	for _, e := range entries {
		if e.Native == "" || seen[e.Native] {
			continue
		}
		seen[e.Native] = true
		checked++
		if !checkRefinement(t, e.Native) {
			failed++
		}
	}
	t.Logf("akshara refinement: %d unique natives checked, %d with a missing boundary", checked, failed)
}
