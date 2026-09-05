package hindi_test

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// These are golden regression tests for the Hindi colloquial engine. They lock
// current end-to-end behavior (the lang/hindi package previously had no direct
// unit tests). Vowel-length spellings such as गाना→gana and संगीत→sangit are
// deliberate divergences validated in
// docs/reviews/2026-09-04-h2-vowel-length-experiments.md.

func engine() *core.Engine {
	return core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
}

func TestHindiGolden(t *testing.T) {
	e := engine()
	groups := map[string][][2]string{
		"basic": {
			{"नमस्ते", "namaste"},
			{"भारत", "bharat"},
			{"दुनिया", "duniya"},
		},
		"schwa-deletion": {
			{"जनता", "janta"},     // medial CCV schwa deleted
			{"कहते", "kahte"},     // medial schwa deleted
			{"देशभर", "deshbhar"}, // compound-boundary deletion
		},
		"schwa-retention": {
			{"मंत्र", "mantra"},      // sonorous-final (र) keeps schwa
			{"यज्ञ", "yagya"},        // ज्ञ conjunct keeps schwa
			{"केंद्रीय", "kendriya"}, // ीय suffix keeps schwa
			{"प्रकाश", "prakaash"},   // word-initial conjunct protects schwa
		},
		"conjuncts": {
			{"क्षत्रिय", "kshatriy"},
			{"ज्ञान", "gyaan"},
			{"श्री", "shri"},
			{"क्या", "kya"},
			{"स्वागत", "swagat"}, // स्व conjunct → sw
		},
		"vowel-length-divergence": {
			{"काम", "kaam"},     // closed-final ा → aa
			{"गाना", "gana"},    // medial open ा stays a (deliberate)
			{"संगीत", "sangit"}, // medial ी stays i (deliberate)
			{"फूल", "phul"},     // medial ू stays u (deliberate)
			{"हिंदी", "hindi"},  // word-final ी → i
		},
		"nukta": {
			{"राज़ी", "razi"},
			{"ख़ास", "khaas"},
			{"फ़िल्म", "film"}, // फ़ + short u exception path
		},
		"numbers": {
			{"९", "9"},
		},
		"long-conjunct-words": {
			{"अंतर्राष्ट्रीय", "antarrashtriya"},
			{"प्रधानमंत्री", "pradhanmantri"},
			{"विश्वविद्यालय", "vishwavidyalay"},
		},
	}

	for group, cases := range groups {
		for _, c := range cases {
			in, want := c[0], c[1]
			if got := e.Transliterate(in); got != want {
				t.Errorf("[%s] Transliterate(%q) = %q, want %q", group, in, got, want)
			}
		}
	}
}

func TestHindiOptions(t *testing.T) {
	e := engine()
	cases := []struct {
		name string
		in   string
		opts core.Options
		want string
	}{
		{"keep-medial-schwa/janta", "जनता", core.Options{KeepMedialSchwa: true}, "janata"},
		{"keep-medial-schwa/kahte", "कहते", core.Options{KeepMedialSchwa: true}, "kahate"},
		{"long-vowels/gaana", "गाना", core.Options{LongVowels: true}, "gaanaa"},
		{"long-vowels/kaam", "काम", core.Options{LongVowels: true}, "kaam"},
		{"simple-nasals/karen", "करें", core.Options{SimpleNasals: true}, "karen"},
		{"simple-nasals/main", "मैं", core.Options{SimpleNasals: true}, "main"},
	}
	for _, c := range cases {
		if got := e.TransliterateWithOptions(c.in, c.opts); got != c.want {
			t.Errorf("[%s] = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestSchwaModel exercises the learned schwa classifier (opt-in). It must load
// the embedded tree without panicking and produce correct output on words whose
// schwa handling is unambiguous.
func TestSchwaModel(t *testing.T) {
	e := engine()
	cases := map[string]string{
		"जनता":   "janta",  // medial schwa deleted
		"कमल":    "kamal",  // 3-consonant word retains
		"गरम":    "garam",  // retains
		"समझ":    "samajh", // retains medial
		"भारत":   "bharat",
		"नमस्ते": "namaste",
	}
	for in, want := range cases {
		got := e.TransliterateWithOptions(in, core.Options{SchwaModel: true})
		if got != want {
			t.Errorf("SchwaModel Transliterate(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestHindiDefaultVsKeepMedialSchwa guards the flag actually changes behavior.
func TestHindiDefaultVsKeepMedialSchwa(t *testing.T) {
	e := engine()
	def := e.Transliterate("जनता")
	keep := e.TransliterateWithOptions("जनता", core.Options{KeepMedialSchwa: true})
	if def == keep {
		t.Errorf("KeepMedialSchwa had no effect on जनता: both %q", def)
	}
}

// TestFormatCharInvariance: output must be identical with and without ZWNJ/ZWJ,
// on every engine configuration (these appear routinely in real Devanagari).
func TestFormatCharInvariance(t *testing.T) {
	e := engine()
	pairs := [][2]string{
		{"‌मकसद", "मकसद"},
		{"कमल‌", "कमल"},
		{"ज्‍ञान", "ज्ञान"},
		{"मुख्‍यमंत्री", "मुख्यमंत्री"},
	}
	configs := map[string]core.Options{
		"default":     {},
		"schwa-model": {SchwaModel: true},
		"rerank":      {Rerank: true},
	}
	for name, opts := range configs {
		for _, p := range pairs {
			got := e.TransliterateWithOptions(p[0], opts)
			want := e.TransliterateWithOptions(p[1], opts)
			if got != want {
				t.Errorf("[%s] %q -> %q but clean form -> %q", name, p[0], got, want)
			}
		}
	}
}
