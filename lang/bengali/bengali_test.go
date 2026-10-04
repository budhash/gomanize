package bengali

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
	"github.com/budhash/gomanize/script/brahmic"
)

func TestB0CompositionalOutputs(t *testing.T) {
	e := core.NewEngine(b0Language{}, colloquial.Colloquial{})
	// These assert B0's mechanical baseline, not attested B1 pronunciation.
	for input, want := range map[string]string{
		"আমি": "ami", "বাংলা": "bangla", "মন": "mon", "কর্ম": "korm", "মধ্য": "modhj",
		"ক্ষ": "ksh", "জ্ঞ": "jn", "হ্ম": "hm", "হ্ন": "hn", "ফল": "phol",
		"অতি": "oti", "০১২৩৪৫৬৭৮৯": "0123456789", "হঠাৎ": "hothat", "উৎসব": "utshob",
		"ৎ": "t", "ৎই": "ti", "ৎঅ": "to", "ৎা": "ta", "৳": "৳", "Latin": "Latin",
	} {
		if got := e.Transliterate(input); got != want {
			t.Errorf("%s: got %q, want %q", input, got, want)
		}
	}
}

func TestCanonicalEquivalence(t *testing.T) {
	e := core.NewEngine(b0Language{}, colloquial.Colloquial{})
	for _, forms := range [][2]string{{"কো", "কো"}, {"কৌ", "কৌ"}, {"বড়", "বড়"}, {"গাঢ়", "গাঢ়"}, {"হয়", "হয়"}, {"ক্ষ", "ক্\u200dষ"}, {"জ্ঞ", "জ্\u200cঞ"}} {
		if a, b := e.Transliterate(forms[0]), e.Transliterate(forms[1]); a != b {
			t.Errorf("%q/%q: %q != %q", forms[0], forms[1], a, b)
		}
	}
	for _, input := range []string{"কো", "কো", "কৌ", "কৌ"} {
		w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse(input, Symbols)
		if len(w.Units) != 2 || !brahmic.IsMatraUnit(w.Units[1]) {
			t.Errorf("%q: split matra not a single dependent vowel", input)
		}
	}
}

func TestKhandaTaStructuralVowellessness(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules("*"))
	for input, want := range map[string]string{"ৎ": "t", "ৎই": "ti", "উৎস": "utsho", "হঠাৎ": "hothat"} {
		if got := e.Transliterate(input); got != want {
			t.Errorf("all rules disabled %q: %q != %q", input, got, want)
		}
	}
	w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse("উৎস", Symbols)
	brahmic.IdentifyRuns(w)
	if !brahmic.GetBrahmicData(w.Units[1]).NoInherentVowel || !brahmic.IsAfterHalant(w.Units[2]) || brahmic.GetRun(w.Units[1]) != brahmic.GetRun(w.Units[2]) {
		t.Fatal("dead consonant lost cluster/run membership")
	}
	bd := brahmic.GetBrahmicData(w.Units[1])
	bd.Schwa = brahmic.SchwaKeep
	bd.SchwaQuality = brahmic.SchwaRaised
	if got := brahmic.NewRenderer().Render(w); got != "utsho" {
		t.Fatalf("quality/keep resurrected dead vowel: %s", got)
	}
}

func TestB0OptionsAndCatalog(t *testing.T) {
	e := core.NewEngine(b0Language{}, colloquial.Colloquial{})
	if got := e.TransliterateWithOptions("অমন", core.Options{InherentVowelA: true}); got != "aman" {
		t.Fatal(got)
	}
	for _, opts := range []core.Options{{SchwaModel: true}, {Lexicon: true}, {Rerank: true}, {LongVowels: true, SimpleNasals: true, KeepMedialSchwa: true}} {
		if got := e.TransliterateWithOptions("বাংলা", opts); got != e.Transliterate("বাংলা") {
			t.Fatalf("Hindi-only options changed Bengali: %q", got)
		}
	}
	if len(b0Catalog().AllRules()) != 3 {
		t.Fatal("unexpected rules in the B0 baseline")
	}
}

// Retain the historical mechanical tests independently of the production catalog.
type b0Language struct{ Bengali }

func (b0Language) Rules() core.RuleCatalog { return b0Catalog() }
func b0Catalog() core.RuleCatalog {
	all := RuleCatalog().AllRules()
	return core.RuleCatalog{
		Schwa: core.AppendIfFound(core.AppendIfFound(nil, all, "schwa.delete.word-final"), all, "schwa.keep.default"),
		Vowel: core.AppendIfFound(nil, all, "vowel.bengali.bare-a-style"),
	}
}
