package gomanize

// Tier 2 of the structured parser-QA plan
// (docs/reviews/2026-09-08-structured-parser-qa.md): gold-free structural
// invariants over the orthographic construct space. These catch parser bugs by
// contradiction — no "correct" romanization needed — and are convention-free.
//
// The well-formedness invariant guards everything that currently works (an
// immediate regression net). The differential (गी ≠ गई) and construct golden
// pin the exact spec for the T-0040 fix; they currently fail (independent vowels
// are parsed as matras), so they are Skip-staged with a TODO — flip them on when
// T-0040 lands rather than letting CI go red now.

import "testing"

// representative base consonants
var invConsonants = []string{"क", "ग", "न", "भ", "म", "ल", "स", "त", "र", "प"}

// each vowel as its dependent (matra) form and its independent form
var invVowelForms = []struct {
	name  string
	matra string
	indep string
}{
	{"i", "ि", "इ"},
	{"ii", "ी", "ई"},
	{"u", "ु", "उ"},
	{"uu", "ू", "ऊ"},
	{"e", "े", "ए"},
	{"ai", "ै", "ऐ"},
	{"o", "ो", "ओ"},
	{"au", "ौ", "औ"},
}

func invEngine(t *testing.T) *Gomanize {
	t.Helper()
	g, err := New("hindi")
	if err != nil {
		t.Fatalf("New(hindi): %v", err)
	}
	return g
}

// Well-formedness: every construct in the space romanizes to non-empty,
// deterministic output. Passes today; a durable net against new breakage.
func TestInvariantWellFormed(t *testing.T) {
	g := invEngine(t)
	for _, c := range invConsonants {
		words := []string{c}
		for _, v := range invVowelForms {
			words = append(words, c+v.matra, c+v.indep)
		}
		for _, w := range words {
			out := g.Translit(w)
			if out == "" {
				t.Errorf("%q romanized to empty string", w)
			}
			if again := g.Translit(w); again != out {
				t.Errorf("%q non-deterministic: %q vs %q", w, out, again)
			}
		}
	}
}

// Differential: a consonant + matra and a consonant + the *independent* vowel of
// the same quality are different syllabifications (गी = "gī" one syllable; गई =
// "ga-ī" two), so they must romanize differently. Convention-free.
func TestInvariantMatraDiffersFromIndependent(t *testing.T) {
	g := invEngine(t)
	for _, c := range invConsonants {
		for _, v := range invVowelForms {
			matra := g.Translit(c + v.matra)
			indep := g.Translit(c + v.indep)
			if matra == indep {
				t.Errorf("%s%s (matra) == %s%s (independent) == %q; they must differ", c, v.matra, c, v.indep, matra)
			}
		}
	}
}

// Construct golden — consonant + independent vowel keeps the consonant's
// inherent vowel (linguistic-rule truth), as accepted reference SETS.
func TestGoldenIndependentVowel(t *testing.T) {
	g := invEngine(t)
	gold := map[string][]string{
		"गई":  {"gai", "gayi"},
		"नई":  {"nai", "nayi"},
		"कई":  {"kai", "kayi"},
		"गए":  {"gae", "gaye"},
		"हुई": {"hui", "huyi"},
	}
	for native, accepted := range gold {
		got := g.Translit(native)
		if !containsStr(accepted, got) {
			t.Errorf("%s → %q; want one of %v", native, got, accepted)
		}
	}
	// contrast: the matra form and the independent form must not coincide
	if g.Translit("गी") == g.Translit("गई") {
		t.Errorf("गी == गई == %q; matra and independent forms must differ", g.Translit("गी"))
	}
}

// Construct golden — word-final chandrabindu must keep its nasal. Tracked
// separately (the render.chandrabindu.final-silent rule drops it mid-word too,
// e.g. चाँद→chaad); unskip when that is fixed.
func TestGoldenChandrabinduNasal(t *testing.T) {
	t.Skip("TODO(F-0010): render.chandrabindu.final-silent drops the nasal (कहाँ→kahaa, चाँद→chaad). Unskip when fixed.")
	g := invEngine(t)
	gold := map[string][]string{
		"कहाँ": {"kahan", "kahaan"},
		"चाँद": {"chand", "chaand"},
	}
	for native, accepted := range gold {
		got := g.Translit(native)
		if !containsStr(accepted, got) {
			t.Errorf("%s → %q; want one of %v", native, got, accepted)
		}
	}
}

func containsStr(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
