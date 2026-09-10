package gomanize

// Tier 2 of the structured parser-QA plan
// (docs/reviews/2026-09-08-structured-parser-qa.md): gold-free structural
// invariants over the orthographic construct space. These catch parser bugs by
// contradiction — no "correct" romanization needed — and are convention-free.
//
// The well-formedness invariant guards everything that works; the differential
// (गी ≠ गई, default and under the schwa model) and the independent-vowel golden
// lock the T-0040 fix. The chandrabindu golden locks the T-0045 fix: chandrabindu
// always keeps its nasal (चाँद→chaand, कहाँ→kahaan, माँ→maan, हाँ→haan).

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
	{"ri", "ृ", "ऋ"},
	{"rri", "ॄ", "ॠ"},
}

func invEngine(t *testing.T) *Gomanize {
	t.Helper()
	g, err := New("hindi")
	if err != nil {
		t.Fatalf("New(hindi): %v", err)
	}
	return g
}

// invEngineSchwaModel returns an engine using the learned schwa classifier — the
// mode most likely to re-collapse the matra/independent distinction if the fix
// leaned on the default schwa rules (Codex review).
func invEngineSchwaModel(t *testing.T) *Gomanize {
	t.Helper()
	g := invEngine(t)
	o := NewOptions()
	o.SchwaModel = true
	g.SetOptions(o)
	return g
}

// diffMatraVsIndependent asserts C+matra ≠ C+independent across the vowel forms.
func diffMatraVsIndependent(t *testing.T, g *Gomanize, mode string) {
	t.Helper()
	for _, c := range invConsonants {
		for _, v := range invVowelForms {
			matra := g.Translit(c + v.matra)
			indep := g.Translit(c + v.indep)
			if matra == indep {
				t.Errorf("[%s] %s%s (matra) == %s%s (independent) == %q; they must differ",
					mode, c, v.matra, c, v.indep, matra)
			}
		}
	}
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
	diffMatraVsIndependent(t, invEngine(t), "default")
}

// Same differential under the learned schwa model — the model must not
// re-collapse the distinction by deleting a consonant's schwa before an
// independent vowel (Codex review: कऋ vs कृ, कॠ vs कॄ).
func TestInvariantMatraDiffersFromIndependentSchwaModel(t *testing.T) {
	diffMatraVsIndependent(t, invEngineSchwaModel(t), "schwa-model")
}

// Construct golden — consonant + independent vowel keeps the consonant's
// inherent vowel (linguistic-rule truth), as accepted reference SETS.
func TestGoldenIndependentVowel(t *testing.T) {
	g := invEngine(t)
	gold := map[string][]string{
		"गई":    {"gai", "gayi"},
		"नई":    {"nai", "nayi"},
		"कई":    {"kai", "kayi"},
		"गए":    {"gae", "gaye"},
		"हुई":   {"hui", "huyi"},
		"दरअसल": {"darasal"}, // independent अ coalesces with the schwa (not "daraasal")
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

// Construct golden — chandrabindu keeps its nasal 'n' everywhere except the
// lexical allowlist (माँ→maa). Locks T-0045: the render.chandrabindu.final-silent
// rule used to drop the nasal for any ा+ँ (चाँद→chaad, पाँच→paach) and a
// word-final+monosyllabic guard still silenced हाँ→haa; it is now a whole-word
// allowlist so माँ→maa but हाँ→haan, चाँद→chaand, कहाँ→kahaan.
func TestGoldenChandrabinduNasal(t *testing.T) {
	g := invEngine(t)
	gold := map[string][]string{
		"कहाँ": {"kahan", "kahaan"},
		"चाँद": {"chand", "chaand"},
		"साँस": {"sans", "saans"},
		"पाँच": {"panch", "paanch"},
		"माँ":  {"maa"},  // lexical exception (allowlist): the iconic form, never "maan"
		"हाँ":  {"haan"}, // same shape as माँ but NOT silenced — never "haa"
	}
	for native, accepted := range gold {
		got := g.Translit(native)
		if !containsStr(accepted, got) {
			t.Errorf("%s → %q; want one of %v", native, got, accepted)
		}
	}
}

// Construct golden — chandrabindu before a labial (प/फ/ब/भ/म) keeps its nasal
// 'n'; it does NOT assimilate to 'm' the way anusvara does (संबंध→sambandh).
// Locks the T-0046 decision (docs/reviews/2026-09-10-chandrabindu-before-labial.md):
// ँ is vowel nasalization, ं is a homorganic nasal, so the marks differ. The
// dominant native class is 'n' (काँपना, साँप, हाँफना); m-cases (ताँबा, सँभाल)
// are lexical, not rule-governed. A blanket ँ→m rule would trip this test.
func TestGoldenChandrabinduLabial(t *testing.T) {
	g := invEngine(t)
	gold := map[string][]string{
		"काँप": {"kanp", "kaanp"},
		"साँप": {"sanp", "saanp"},
		"हाँफ": {"hanf", "haanf"},
		"धाँप": {"dhanp", "dhaanp"},
	}
	for native, accepted := range gold {
		got := g.Translit(native)
		if !containsStr(accepted, got) {
			t.Errorf("%s → %q; want one of %v (chandrabindu+labial must stay n, not m)", native, got, accepted)
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
