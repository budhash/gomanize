package bengali

import (
	"strings"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// RuleCatalog composes Bengali rules without importing Hindi heuristics or
// learned models. Each action owns its unit; source identities survive rewrites.
func RuleCatalog() core.RuleCatalog {
	shared := brahmic.SchwaRules()
	schwa := core.AppendIfFound(nil, shared, "schwa.delete.word-final")
	schwa = core.AppendIfFound(schwa, shared, "schwa.keep.default")
	c := core.RuleCatalog{Schwa: schwa}
	add := func(name string, phase core.RulePhase, priority int, cond func(*core.Unit, *core.Word) bool, action func(*core.Unit, *core.Word)) {
		r := core.Rule{Name: name, Phase: phase, Scope: core.ScopeLanguage, Priority: priority, Mode: core.ModeExclusive, Condition: cond, Action: action}
		switch phase {
		case core.PhaseSchwa:
			c.Schwa = append(c.Schwa, r)
		case core.PhaseConsonant:
			c.Consonant = append(c.Consonant, r)
		case core.PhaseVowel:
			c.Vowel = append(c.Vowel, r)
		case core.PhaseRender:
			c.Render = append(c.Render, r)
		}
	}
	add("schwa.bengali.model", core.PhaseSchwa, 70, func(u *core.Unit, w *core.Word) bool {
		_, ok := vowelModelDecision(w, u)
		return ok
	}, func(u *core.Unit, w *core.Word) {
		label, _ := vowelModelDecision(w, u)
		if label == 0 {
			brahmic.SetSchwa(u, brahmic.SchwaDelete)
		} else {
			brahmic.SetSchwa(u, brahmic.SchwaKeep)
			if label == 2 {
				brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaRaised
			}
		}
	})
	c.Schwa[len(c.Schwa)-1].Conditional = "SchwaModel"
	keep := func(u *core.Unit, w *core.Word) { brahmic.SetSchwa(u, brahmic.SchwaKeep) }
	// An explicit final hasant spells "no vowel"; keep rules must not override it.
	spelledDead := func(u *core.Unit) bool { return brahmic.GetBrahmicData(u).TrailingHalant }
	add("schwa.bengali.final-cluster", core.PhaseSchwa, 80, func(u *core.Unit, w *core.Word) bool {
		return consonant(u) && u.IsWordFinal() && brahmic.IsAfterHalant(u) && !brahmic.GetBrahmicData(u).NoInherentVowel && !loanCoda(u) && !spelledDead(u)
	}, keep)
	add("schwa.bengali.final-h", core.PhaseSchwa, 79, func(u *core.Unit, w *core.Word) bool {
		return source(u, "হ") && u.Prev != nil && u.IsWordFinal() && !spelledDead(u)
	}, keep)
	add("schwa.bengali.post-visarga", core.PhaseSchwa, 78, func(u *core.Unit, w *core.Word) bool { return visargaTarget(u) && !spelledDead(u) }, keep)
	add("consonant.bengali.f-spelling", core.PhaseConsonant, 95, func(u *core.Unit, w *core.Word) bool { return source(u, "ফ") }, func(u *core.Unit, w *core.Word) { u.BaseRom = "f" })
	add("consonant.bengali.h-nasal", core.PhaseConsonant, 94, func(u *core.Unit, w *core.Word) bool { return pairLeft(u, "হ", "ম") || pairLeft(u, "হ", "ন") }, func(u *core.Unit, w *core.Word) {
		if source(u.Next, "ম") {
			u.BaseRom = "m"
		} else {
			u.BaseRom = "n"
		}
	})
	add("consonant.bengali.ksha-left", core.PhaseConsonant, 90, func(u *core.Unit, w *core.Word) bool { return pairLeft(u, "ক", "ষ") }, func(u *core.Unit, w *core.Word) {
		if u.IsWordInitial() {
			u.BaseRom = ""
		} else {
			u.BaseRom = "k"
		}
	})
	add("consonant.bengali.ksha-right", core.PhaseConsonant, 89, func(u *core.Unit, w *core.Word) bool { return pairRight(u, "ক", "ষ") }, func(u *core.Unit, w *core.Word) { u.BaseRom = "kh" })
	add("consonant.bengali.jna-left", core.PhaseConsonant, 88, func(u *core.Unit, w *core.Word) bool { return pairLeft(u, "জ", "ঞ") }, func(u *core.Unit, w *core.Word) {
		if u.IsWordInitial() {
			u.BaseRom = ""
		} else {
			u.BaseRom = "g"
		}
	})
	add("consonant.bengali.jna-right", core.PhaseConsonant, 87, func(u *core.Unit, w *core.Word) bool { return pairRight(u, "জ", "ঞ") }, func(u *core.Unit, w *core.Word) {
		if u.Prev.IsWordInitial() {
			u.BaseRom = "gy"
		} else {
			u.BaseRom = "g"
		}
	})
	add("consonant.bengali.s-cluster", core.PhaseConsonant, 70, func(u *core.Unit, w *core.Word) bool {
		return source(u, "স") && u.Next != nil && brahmic.IsAfterHalant(u.Next) && oneOf(u.Next, "তথটঠনপফকখ")
	}, func(u *core.Unit, w *core.Word) { u.BaseRom = "s" })
	add("consonant.bengali.initial-phala", core.PhaseConsonant, 60, func(u *core.Unit, w *core.Word) bool { return phala(u) && u.Prev.IsWordInitial() }, func(u *core.Unit, w *core.Word) {
		u.BaseRom = ""
		if source(u, "য") {
			u.BaseRom = "y"
			brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaOpen
		}
	})
	add("render.bengali.medial-phala", core.PhaseRender, 60, func(u *core.Unit, w *core.Word) bool { return phala(u) && !u.Prev.IsWordInitial() }, func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.RepeatPrevious })
	add("consonant.bengali.visarga-suppress", core.PhaseConsonant, 50, func(u *core.Unit, w *core.Word) bool { return source(u, "ঃ") && visargaTarget(u.Next) }, func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
	add("render.bengali.visarga-gemination", core.PhaseRender, 50, func(u *core.Unit, w *core.Word) bool { return visargaTarget(u) }, func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.GeminateSelf })
	add("vowel.bengali.hao", core.PhaseVowel, 40, func(u *core.Unit, w *core.Word) bool { return w.Original == "হও" && source(u, "হ") }, func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaOpen })
	add("vowel.bengali.howa-w", core.PhaseVowel, 39, func(u *core.Unit, w *core.Word) bool { return howa(w) && source(u, "ও") }, func(u *core.Unit, w *core.Word) { u.BaseRom = "w" })
	add("consonant.bengali.howa-y", core.PhaseConsonant, 39, func(u *core.Unit, w *core.Word) bool { return howa(w) && (source(u, "য়") || source(u, "য়")) }, func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
	add("vowel.bengali.bare-a-style", core.PhaseVowel, 10, func(u *core.Unit, w *core.Word) bool { return w.Options.InherentVowelA && source(u, "অ") }, func(u *core.Unit, w *core.Word) { u.BaseRom = "a" })
	c.Vowel[len(c.Vowel)-1].Conditional = "InherentVowelA"
	return c
}

func source(u *core.Unit, s string) bool { return u != nil && string(u.Runes) == s }
func consonant(u *core.Unit) bool {
	return u != nil && u.Type == core.UnitConsonant && brahmic.GetBrahmicData(u) != nil
}
func oneOf(u *core.Unit, s string) bool {
	return u != nil && len(u.Runes) == 1 && strings.ContainsRune(s, u.Runes[0])
}
func simpleClusterRight(u *core.Unit) bool {
	return consonant(u) && brahmic.IsSingleLetter(u) && brahmic.GetBrahmicData(u).AfterHalant && consonant(u.Prev) && brahmic.IsSingleLetter(u.Prev) && !brahmic.IsAfterHalant(u.Prev) && !brahmic.GetBrahmicData(u.Prev).NoInherentVowel && (u.Next == nil || !brahmic.IsAfterHalant(u.Next))
}
func pairRight(u *core.Unit, left, right string) bool {
	return simpleClusterRight(u) && source(u.Prev, left) && source(u, right)
}
func pairLeft(u *core.Unit, left, right string) bool {
	return u != nil && pairRight(u.Next, left, right)
}
func phala(u *core.Unit) bool {
	if !simpleClusterRight(u) {
		return false
	}
	switch string(u.Runes) {
	case "য":
		return !oneOf(u.Prev, "রহ")
	case "ব":
		return !oneOf(u.Prev, "রমনণঙহব")
	case "ম":
		return oneOf(u.Prev, "তথদধশষস")
	}
	return false
}
func visargaTarget(u *core.Unit) bool {
	// Only the attested dukh family is contracted here. A blanket medial rule
	// would incorrectly rewrite contexts such as অতঃপর and বহিঃপ্রকাশ.
	if !consonant(u) || !source(u, "খ") || !source(u.Prev, "ঃ") || brahmic.IsAfterHalant(u) || (u.Next != nil && brahmic.IsAfterHalant(u.Next)) {
		return false
	}
	vowel := u.Prev.Prev
	return source(vowel, "ু") && source(vowel.Prev, "দ") && vowel.Prev.IsWordInitial()
}
func howa(w *core.Word) bool { return w.Original == "হওয়া" || w.Original == "হওয়া" }

// These final clusters in Bengali loans retain a closed coda (art/card/test),
// unlike the native dental clusters in অন্ত and শব্দ. No lexical lookup occurs.
func loanCoda(u *core.Unit) bool {
	return (source(u.Prev, "র") && oneOf(u, "টড")) || (source(u.Prev, "স") && source(u, "ট"))
}
