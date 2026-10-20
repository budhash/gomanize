package bengali

import (
	"strings"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/internal/bengalitables"
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
	// T-0119 bundle: each rule ships on its own train-only leave-one-out ablation
	// (docs/reviews/2026-10-14-bengali-rule-bundle.md).
	add("schwa.bengali.final-ito", core.PhaseSchwa, 81, func(u *core.Unit, w *core.Word) bool {
		return finalBareTa(u) && matraOf(u.Prev, "িীৃূ")
	}, keep)
	add("schwa.bengali.final-ta", core.PhaseSchwa, 82, func(u *core.Unit, w *core.Word) bool {
		return finalBareTa(u) && consonant(u.Prev) && !brahmic.IsAfterHalant(u.Prev) && oneOf(u.Prev, "গরহনশষ")
	}, keep)
	add("schwa.bengali.loan-coda", core.PhaseSchwa, 83, func(u *core.Unit, w *core.Word) bool {
		return u.IsWordFinal() && simpleClusterRight(u) && bengalitables.LoanCodas[string(u.Prev.Runes)+string(u.Runes)]
	}, func(u *core.Unit, w *core.Word) { brahmic.SetSchwa(u, brahmic.SchwaDelete) })
	// T-0120: a stem-final bare consonant before a word-final nominal suffix
	// loses its vowel (অঞ্চলকে oncholke). Entries and the rule are selected on
	// train (docs/reviews/2026-10-20-bengali-suffix-coda.md).
	add("schwa.bengali.suffix-boundary", core.PhaseSchwa, 84, suffixStem, func(u *core.Unit, w *core.Word) { brahmic.SetSchwa(u, brahmic.SchwaDelete) })
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
	add("consonant.bengali.sibilant-ya", core.PhaseConsonant, 61, sibilantYa, func(u *core.Unit, w *core.Word) { u.BaseRom = "y" })
	add("render.bengali.sibilant-ya", core.PhaseRender, 61, sibilantYa, noGemination)
	add("render.bengali.medial-phala", core.PhaseRender, 60, func(u *core.Unit, w *core.Word) bool { return phala(u) && !u.Prev.IsWordInitial() }, func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.RepeatPrevious })
	add("consonant.bengali.visarga-suppress", core.PhaseConsonant, 50, func(u *core.Unit, w *core.Word) bool { return source(u, "ঃ") && visargaTarget(u.Next) }, func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
	add("render.bengali.visarga-gemination", core.PhaseRender, 50, func(u *core.Unit, w *core.Word) bool { return visargaTarget(u) }, func(u *core.Unit, w *core.Word) { brahmic.GetBrahmicData(u).Gemination = brahmic.GeminateSelf })
	// এখনও: the final ও after a bare consonant repeats the inherent vowel already
	// rendered, but only when that vowel renders as o (not a-style or open: কও kao).
	add("vowel.bengali.final-o", core.PhaseVowel, 41, func(u *core.Unit, w *core.Word) bool {
		return source(u, "ও") && u.IsWordFinal() && consonant(u.Prev) && !brahmic.IsAfterHalant(u) && w.Original != "হও" &&
			brahmic.InherentVowelRoman(u.Prev, w) == "o"
	}, func(u *core.Unit, w *core.Word) { u.BaseRom = "" })
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
func matraOf(u *core.Unit, s string) bool { return brahmic.IsMatraUnit(u) && oneOf(u, s) }
func noGemination(u *core.Unit, w *core.Word) {
	brahmic.GetBrahmicData(u).Gemination = brahmic.GeminationNone
}

// finalBareTa is a word-final ত with its own inherent vowel (not a cluster
// member and not spelled dead), as in the participle ending -িত.
func finalBareTa(u *core.Unit) bool {
	return source(u, "ত") && u.IsWordFinal() && !brahmic.IsAfterHalant(u) && !brahmic.GetBrahmicData(u).TrailingHalant
}

// sibilantYa is a medial ya-phala after শ/ষ/স: দৃশ্যের reads drishyer, not drishsher.
func sibilantYa(u *core.Unit, w *core.Word) bool {
	return source(u, "য") && simpleClusterRight(u) && oneOf(u.Prev, "শষস") && !u.Prev.IsWordInitial()
}

func howa(w *core.Word) bool { return w.Original == "হওয়া" || w.Original == "হওয়া" }

// These final clusters in Bengali loans retain a closed coda (art/card/test),
// unlike the native dental clusters in অন্ত and শব্দ. No lexical lookup occurs.
func loanCoda(u *core.Unit) bool {
	return (source(u.Prev, "র") && oneOf(u, "টড")) || (source(u.Prev, "স") && source(u, "ট"))
}

// The loan-coda entries (bengalitables.LoanCodas) extend loanCoda with final
// clusters that Dakshina train mostly writes closed (ন্ট in পেইন্ট, ক্স in বক্স).

// suffixStem is a bare consonant (no matra, no hasant, not in the word's first
// akshara) immediately followed by the longest word-final suffix in
// bengalitables.Suffixes. Suffixes are matched on unit boundaries.
func suffixStem(u *core.Unit, w *core.Word) bool {
	if !consonant(u) || u.Next == nil || !consonant(u.Next) || brahmic.IsAfterHalant(u.Next) || brahmic.GetBrahmicData(u).NoInherentVowel {
		return false
	}
	first := u
	for brahmic.IsAfterHalant(first) && first.Prev != nil {
		first = first.Prev
	}
	if first.IsWordInitial() {
		return false
	}
	var tail strings.Builder
	for x := u.Next; x != nil; x = x.Next {
		tail.WriteString(string(x.Runes))
	}
	if !bengalitables.Suffixes[tail.String()] {
		return false
	}
	// Longest match: no longer table suffix may start at or before u.
	longer := tail.String()
	for x := u; x != nil; x = x.Prev {
		longer = string(x.Runes) + longer
		if bengalitables.Suffixes[longer] {
			return false
		}
	}
	return true
}
