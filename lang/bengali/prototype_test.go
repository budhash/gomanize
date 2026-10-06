package bengali

import (
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// prototypeRules deliberately stays out of RuleCatalog. Exact design examples
// prove mechanics; they do not establish a general Bengali pronunciation rule.
func prototypeRules() []core.Rule {
	rules := b0Catalog().AllRules()
	add := func(name string, phase core.RulePhase, priority int, cond func(*core.Unit, *core.Word) bool, action func(*core.Unit, *core.Word)) {
		rules = append(rules, core.Rule{Name: name, Phase: phase, Scope: core.ScopeLanguage, Priority: priority, Mode: core.ModeAlways, Condition: cond, Action: action})
	}
	is := func(word, unit string) func(*core.Unit, *core.Word) bool {
		return func(u *core.Unit, w *core.Word) bool { return w.Original == word && string(u.Runes) == unit }
	}
	add("prototype.phala", core.PhaseRender, 10, is("মধ্য", "য"), func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.RepeatPrevious })
	add("prototype.visarga", core.PhaseConsonant, 11, is("দুঃখ", "ঃ"), func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
	add("prototype.self", core.PhaseRender, 12, is("দুঃখ", "খ"), func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.GeminateSelf })
	// The keep decisions are explicit and independent of render-time modes.
	add("prototype.keep", core.PhaseSchwa, 99, func(u *core.Unit, w *core.Word) bool {
		return is("মধ্য", "য")(u, w) || is("দুঃখ", "খ")(u, w)
	}, func(u *core.Unit, w *core.Word) { brahmic.SetSchwa(u, brahmic.SchwaKeep) })
	add("prototype.hao", core.PhaseVowel, 20, is("হও", "হ"), func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaOpen })
	howa := func(w *core.Word) bool { return w.Original == "হওয়া" || w.Original == "হওয়া" }
	add("prototype.howa-w", core.PhaseVowel, 21, func(u *core.Unit, w *core.Word) bool { return howa(w) && string(u.Runes) == "ও" }, func(u *core.Unit, w *core.Word) { u.BaseRom = "w" })
	add("prototype.howa-y", core.PhaseConsonant, 22, func(u *core.Unit, w *core.Word) bool {
		return howa(w) && (string(u.Runes) == "য়" || string(u.Runes) == "য়")
	}, func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
	return rules
}

func TestB1PrototypeContextsAndTraces(t *testing.T) {
	for _, tc := range []struct {
		input, want, alternate string
		traces                 []string
	}{
		{"মধ্য", "moddho", "maddha", []string{"prototype.phala", "prototype.keep"}},
		{"দুঃখ", "dukkho", "dukkha", []string{"prototype.visarga", "prototype.self", "prototype.keep"}},
		{"হও", "hao", "hao", []string{"prototype.hao"}},
		{"হওয়া", "howa", "hawa", []string{"prototype.howa-w", "prototype.howa-y"}},
		{"হওয়া", "howa", "hawa", []string{"prototype.howa-w", "prototype.howa-y"}},
		{"ও", "o", "o", nil}, {"কও", "koo", "kao", nil}, {"য়া", "ya", "ya", nil},
	} {
		for _, alt := range []bool{false, true} {
			w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse(tc.input, Symbols)
			w.Options = core.Options{InherentVowelA: alt}
			brahmic.IdentifyRuns(w)
			e := core.NewRuleEngine(prototypeRules())
			e.SetDebugMetaExtractor(brahmic.New().DebugMetaExtractor())
			e.EnableDebug(true)
			e.Apply(w)
			want := tc.want
			if alt {
				want = tc.alternate
			}
			if got := brahmic.NewRenderer().Render(w); got != want {
				t.Errorf("%s alt=%v: %s != %s", tc.input, alt, got, want)
			}
			for _, name := range tc.traces {
				found := false
				for _, trace := range e.Traces() {
					if trace.Rule == name {
						found = true
						if name == "prototype.phala" && !strings.Contains(trace.Metadata, "gemination=RepeatPrevious") {
							t.Error(trace)
						}
						if name == "prototype.self" && !strings.Contains(trace.Metadata, "gemination=GeminateSelf") {
							t.Error(trace)
						}
						if name == "prototype.hao" && !strings.Contains(trace.Metadata, "quality=Open") {
							t.Error(trace)
						}
					}
				}
				if !found {
					t.Errorf("%s missing trace %s", tc.input, name)
				}
			}
			if len(tc.traces) == 0 {
				for _, trace := range e.Traces() {
					if strings.HasPrefix(trace.Rule, "prototype.") {
						t.Errorf("negative context matched: %+v", trace)
					}
				}
			}
		}
	}
}

func TestB1GeminationVowelOwnership(t *testing.T) {
	for _, tc := range []struct {
		input       string
		index       int
		mode        brahmic.GeminationMode
		onset, want string
		del         bool
	}{
		{"মধ্য", 2, brahmic.RepeatPrevious, "", "moddho", false},
		{"মধ্য", 2, brahmic.RepeatPrevious, "", "moddh", true},
		{"মধ্যা", 2, brahmic.RepeatPrevious, "", "moddha", true},
		{"মত্য", 2, brahmic.RepeatPrevious, "", "motto", false},
		{"মধ্য", 2, brahmic.RepeatPrevious, "kh", "mokkho", false},
		{"মধ্য", 2, brahmic.RepeatPrevious, "f", "moffo", false},
		{"দুঃখ", 3, brahmic.GeminateSelf, "", "dukkho", false},
		{"দুঃখ", 3, brahmic.GeminateSelf, "", "dukkh", true},
		{"দুঃখা", 3, brahmic.GeminateSelf, "", "dukkha", true},
		{"দুঃখ", 3, brahmic.GeminateSelf, "t", "dutto", false},
		// Invalid/ambiguous modes preserve the original compositional onset.
		{"মন্দ্য", 3, brahmic.RepeatPrevious, "", "mondjo", false},
		{"মধ্য্র", 2, brahmic.RepeatPrevious, "", "modhjro", false},
		{"মধয", 2, brahmic.RepeatPrevious, "", "modhojo", false},
		{"মধ্য", 2, brahmic.RepeatPrevious, "ksh", "mokshjo", false},
	} {
		w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse(tc.input, Symbols)
		u := w.Units[tc.index]
		bd := brahmic.GetBrahmicData(u)
		bd.Gemination = tc.mode
		if tc.del {
			bd.Schwa = brahmic.SchwaDelete
		}
		if tc.mode == brahmic.GeminateSelf {
			w.Units[2].BaseRom = ""
			if tc.onset != "" {
				u.BaseRom = tc.onset
			}
		} else if tc.onset != "" {
			u.Prev.BaseRom = tc.onset
		}
		before := make([]string, len(w.Units))
		for i, x := range w.Units {
			before[i] = x.BaseRom
		}
		if got := brahmic.NewRenderer().Render(w); got != tc.want {
			t.Errorf("%s (%v/%s/delete=%v): %s != %s", tc.input, tc.mode, tc.onset, tc.del, got, tc.want)
		}
		for i, x := range w.Units {
			if x.BaseRom != before[i] {
				t.Fatal("renderer mutated onset")
			}
		}
	}
}

func TestB1GeminationRejectsAtomicAndDeadOnsets(t *testing.T) {
	for _, change := range []string{"atomic", "multi-rune", "dead", "previous-mode", "self-cluster"} {
		w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse("মধ্য", Symbols)
		u := w.Units[2]
		brahmic.GetBrahmicData(u).Gemination = brahmic.RepeatPrevious
		want := "modhjo"
		switch change {
		case "atomic":
			u.Prev.Type = core.UnitConjunct
		case "multi-rune":
			u.Prev.Runes = []rune("দ্ধ")
		case "dead":
			brahmic.GetBrahmicData(u.Prev).NoInherentVowel = true
		case "previous-mode":
			brahmic.GetBrahmicData(u.Prev).Gemination = brahmic.GeminateSelf
		case "self-cluster":
			brahmic.GetBrahmicData(u).Gemination = brahmic.GeminateSelf
		}
		if got := brahmic.NewRenderer().Render(w); got != want {
			t.Errorf("%s: %s != %s", change, got, want)
		}
	}
}
