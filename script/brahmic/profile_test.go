package brahmic_test

import (
	"reflect"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
	"github.com/budhash/gomanize/script/brahmic"
)

// Synthetic symbols test script independence without introducing a second
// language before the Hindi-preserving refactor is complete.
func profileSymbols() core.SymbolMap {
	return core.SymbolMap{
		"k": {Category: core.CatConsonant, BaseRom: "k"},
		"t": {Category: core.CatConsonant, BaseRom: "t"},
		"r": {Category: core.CatConsonant, BaseRom: "r"},
		"A": {Category: core.CatVowel, BaseRom: "o"},
		"E": {Category: core.CatVowel, BaseRom: "e"},
		"~": {Category: brahmic.CatMatra, BaseRom: "a"},
		"^": {Category: brahmic.CatHalant},
	}
}

func customProfile() *brahmic.ScriptProfile {
	return &brahmic.ScriptProfile{
		InherentVowel: "o", RaisedVowel: "u", BareVowelRunes: []rune{'A'},
		AaMatra: '~', SonorousRunes: []rune{'r'}, IndependentVowelRange: [2]rune{'A', 'E'},
	}
}

func TestProfileSurvivesPreparation(t *testing.T) {
	supplied := customProfile()
	p := brahmic.NewParser(brahmic.Config{Profile: supplied})
	// A caller changing configuration after parser construction must not change
	// any subsequently parsed word's resolved constants.
	supplied.InherentVowel = "x"
	supplied.BareVowelRunes[0] = 'X'
	supplied.SonorousRunes[0] = 'X'
	for _, input := range []string{"kA", "A", "!"} {
		w := p.Parse(input, profileSymbols())
		before := brahmic.GetWordBrahmicData(w)
		if before == nil || !reflect.DeepEqual(before.Profile, customProfile()) {
			t.Fatalf("%q: unresolved/aliased profile: %+v", input, before)
		}
		brahmic.IdentifyRuns(w)
		if brahmic.GetWordBrahmicData(w) != before {
			t.Fatalf("%q: Prepare replaced word metadata", input)
		}
		n := len(before.Runs)
		brahmic.IdentifyRuns(w)
		if len(before.Runs) != n {
			t.Fatalf("%q: repeated Prepare accumulated runs", input)
		}
	}
	empty := p.Parse("", profileSymbols())
	brahmic.IdentifyRuns(empty)
	if got := brahmic.NewRenderer().Render(empty); got != "" {
		t.Fatal(got)
	}
}

func TestProfileRendering(t *testing.T) {
	cases := []struct {
		name, input string
		state       brahmic.SchwaState
		quality     brahmic.SchwaQuality
		style       bool
		want        string
	}{
		{"pending", "k", brahmic.SchwaPending, brahmic.SchwaDefault, false, "ko"},
		{"keep", "k", brahmic.SchwaKeep, brahmic.SchwaDefault, false, "ko"},
		{"delete", "k", brahmic.SchwaDelete, brahmic.SchwaRaised, false, "k"},
		{"raised", "k", brahmic.SchwaKeep, brahmic.SchwaRaised, false, "ku"},
		{"open", "k", brahmic.SchwaKeep, brahmic.SchwaOpen, false, "ka"},
		{"style", "k", brahmic.SchwaPending, brahmic.SchwaDefault, true, "ka"},
		{"raised-style", "k", brahmic.SchwaKeep, brahmic.SchwaRaised, true, "ku"},
		{"bare-coalesces", "kA", brahmic.SchwaKeep, brahmic.SchwaDefault, false, "ko"},
		{"independent-ignores-delete", "kE", brahmic.SchwaDelete, brahmic.SchwaDefault, false, "koe"},
		{"independent-quality", "kE", brahmic.SchwaDelete, brahmic.SchwaOpen, false, "kae"},
		{"independent-raised", "kE", brahmic.SchwaPending, brahmic.SchwaRaised, false, "kue"},
		{"independent-style", "kE", brahmic.SchwaPending, brahmic.SchwaDefault, true, "kae"},
		{"matra", "k~", brahmic.SchwaKeep, brahmic.SchwaRaised, false, "ka"},
		{"halant", "k^t", brahmic.SchwaKeep, brahmic.SchwaRaised, false, "kto"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := brahmic.NewParser(brahmic.Config{Profile: customProfile()}).Parse(tt.input, profileSymbols())
			brahmic.IdentifyRuns(w)
			w.Options.InherentVowelA = tt.style
			bd := brahmic.GetBrahmicData(w.Units[0])
			bd.Schwa = tt.state
			bd.SchwaQuality = tt.quality
			if got := brahmic.NewRenderer().Render(w); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfileRuleConstants(t *testing.T) {
	cases := []struct {
		name, input string
		profile     *brahmic.ScriptProfile
		want        bool
	}{
		{"schwa.keep.sonorous-final", "k^r", customProfile(), true},
		{"schwa.keep.sonorous-final", "k^t", customProfile(), false},
		{"schwa.delete.before-cc", "k~ktr", customProfile(), false},
		{"schwa.delete.before-cc", "kEktr", customProfile(), true},
		{"schwa.delete.cccc-final", "Ak^ttr", customProfile(), false},
	}
	outsideRange := customProfile()
	outsideRange.IndependentVowelRange = [2]rune{'X', 'Z'}
	cases = append(cases, struct {
		name, input string
		profile     *brahmic.ScriptProfile
		want        bool
	}{"schwa.delete.cccc-final", "Ak^ttr", outsideRange, true})
	for _, tt := range cases {
		t.Run(tt.name+"/"+tt.input, func(t *testing.T) {
			w := brahmic.NewParser(brahmic.Config{Profile: tt.profile}).Parse(tt.input, profileSymbols())
			brahmic.IdentifyRuns(w)
			index := 2
			if tt.name == "schwa.keep.sonorous-final" {
				index = 1
			}
			// Isolate profile-based initial-cluster detection from the unchanged
			// rune-index guard in cccc-final. The synthetic position is only
			// for this condition-level test; parsing itself is not modified.
			if tt.name == "schwa.delete.cccc-final" {
				w.Units[index].Start.Rune = 1
			}

			for _, r := range brahmic.SchwaRules() {
				if r.Name == tt.name {
					if got := r.Condition(w.Units[index], w); got != tt.want {
						t.Fatalf("got %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

// Legacy exported Config literals and manually built words must remain usable.
func TestLegacyProfileDefaults(t *testing.T) {
	renderer := brahmic.NewRenderer()
	for _, cfg := range []brahmic.Config{{}, {Halant: hindi.Halant, Nukta: hindi.Nukta}, {Profile: &brahmic.ScriptProfile{}}} {
		for input, want := range map[string]string{"क": "ka", "कऋ": "kari", "कअ": "ka", "कऄ": "ka", "का": "ka"} {
			w := brahmic.NewParser(cfg).Parse(input, hindi.Symbols)
			brahmic.IdentifyRuns(w)
			if got := renderer.Render(w); got != want {
				t.Fatalf("%q: got %q, want %q", input, got, want)
			}
			brahmic.GetWordBrahmicData(w).Profile = nil
			if got := renderer.Render(w); got != want {
				t.Fatalf("nil profile %q: got %q", input, got)
			}
		}
	}
	w := core.NewWord("क")
	w.AddUnit(&core.Unit{Runes: []rune{'क'}, Type: core.UnitConsonant, BaseRom: "k"})
	if got := renderer.Render(w); got != "ka" {
		t.Fatalf("manual word: %q", got)
	}
	// Hindi's explicit profile and every legacy config use the same full rules.
	for _, cfg := range []brahmic.Config{{}, {Halant: hindi.Halant, Nukta: hindi.Nukta, MultiChar: hindi.MultiChar}} {
		for _, input := range []string{"नमस्ते", "मंत्र", "दरअसल", "गई", "जनता"} {
			w := brahmic.NewParser(cfg).Parse(input, hindi.Symbols)
			brahmic.IdentifyRuns(w)
			core.NewRuleEngine(hindi.RuleCatalog().AllRules()).Apply(w)
			want := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{}).Transliterate(input)
			if got := renderer.Render(w); got != want {
				t.Fatalf("legacy %q: %q != %q", input, got, want)
			}
		}
	}
}

func TestEmptyMembershipAndPartialDefaults(t *testing.T) {
	p := &brahmic.ScriptProfile{InherentVowel: "o", BareVowelRunes: []rune{}, SonorousRunes: []rune{}}
	w := brahmic.NewParser(brahmic.Config{Profile: p}).Parse("कअ", hindi.Symbols)
	if got := brahmic.NewRenderer().Render(w); got != "koa" {
		t.Fatalf("empty bare set: %q", got)
	}
	wd := brahmic.GetWordBrahmicData(w)
	if wd.Profile.AaMatra != 'ा' || wd.Profile.RaisedVowel != "o" || len(wd.Profile.SonorousRunes) != 0 {
		t.Fatalf("partial defaults: %+v", wd.Profile)
	}
}

func TestHindiCatalogPrioritiesAndStyle(t *testing.T) {
	seen := map[[2]int]string{}
	for _, r := range hindi.RuleCatalog().AllRules() {
		key := [2]int{int(r.Phase), r.EffectivePriority()}
		if prev, ok := seen[key]; ok {
			t.Fatalf("Hindi priority conflict: %s and %s", prev, r.Name)
		}
		seen[key] = r.Name
	}
	e := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
	for _, input := range []string{"मंत्र", "गई", "कऋ", "दरअसल", "ऄ", "जनता", "क़करन", "ज्ञानी"} {
		for _, opts := range []core.Options{{}, {SchwaModel: true}, {Lexicon: true}, {Rerank: true}, {KeepMedialSchwa: true, LongVowels: true, SimpleNasals: true}} {
			want := e.TransliterateWithOptions(input, opts)
			opts.InherentVowelA = true
			if got := e.TransliterateWithOptions(input, opts); got != want {
				t.Fatalf("Hindi style changed %q: %q != %q", input, got, want)
			}
		}
	}
}
