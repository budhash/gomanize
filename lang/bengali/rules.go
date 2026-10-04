package bengali

import (
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// RuleCatalog is deliberately naive: retain medial vowels, delete word-final
// vowels. Bengali final-cluster retention, phalas, and positional conjuncts
// belong to B1. Never import Hindi's trained models or language rules here.
func RuleCatalog() core.RuleCatalog {
	shared := brahmic.SchwaRules()
	rules := core.AppendIfFound(nil, shared, "schwa.delete.word-final")
	rules = core.AppendIfFound(rules, shared, "schwa.keep.default")
	return core.RuleCatalog{Schwa: rules, Vowel: []core.Rule{{
		Name: "vowel.bengali.bare-a-style", Phase: core.PhaseVowel, Scope: core.ScopeLanguage,
		Priority: 10, Mode: core.ModeExclusive, Conditional: "InherentVowelA",
		Condition: func(u *core.Unit, w *core.Word) bool {
			return w.Options.InherentVowelA && len(u.Runes) == 1 && u.Runes[0] == 'অ'
		},
		Action: func(u *core.Unit, w *core.Word) { u.BaseRom = "a" },
	}}}
}
