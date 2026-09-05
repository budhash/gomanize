package hindi

import (
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// RuleCatalog returns the complete Hindi rule catalog.
// Schemes select which rules to use from this catalog.
func RuleCatalog() core.RuleCatalog {
	return core.RuleCatalog{
		Schwa:     schwaRules(),
		Consonant: consonantRules(),
		Vowel:     vowelRules(),
		Render:    renderRules(),
	}
}

// schwaRules returns all schwa-related rules.
func schwaRules() []core.Rule {
	// Hindi-specific schwa rules, composed with the shared Brahmic schwa rules
	// (script/brahmic). Effective priority ordering is applied by the rule engine,
	// so slice order here does not matter.
	hindiSchwa := []core.Rule{
		// schwa.model.predict (Language:90) — learned decision-tree classifier.
		// When the SchwaModel option is set, this takes over ALL inherent-schwa
		// decisions (delete/keep) from the heuristic rules below. Being Exclusive
		// and highest-priority, it marks acted units so the heuristics don't fire.
		// See lang/hindi/schwa_model.go and docs/reviews/2026-09-04-h3-schwa-classifier.md.
		{
			Name:        "schwa.model.predict",
			Phase:       core.PhaseSchwa,
			Scope:       core.ScopeLanguage,
			Priority:    90,
			Mode:        core.ModeExclusive,
			Conditional: "SchwaModel",
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !w.Options.SchwaModel {
					return false
				}
				if !brahmic.IsConsonantOrConjunct(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
					return false
				}
				_, applies := schwaModelDecision(w, u)
				return applies
			},
			Action: func(u *core.Unit, w *core.Word) {
				del, _ := schwaModelDecision(w, u)
				if del {
					brahmic.SetSchwa(u, brahmic.SchwaDelete)
					if run := brahmic.GetRun(u); run != nil {
						run.DeletedAt = brahmic.GetRunIndex(u)
					}
				} else {
					brahmic.SetSchwa(u, brahmic.SchwaKeep)
				}
			},
		},

		// schwa.keep.gya-final (Language:65)
		// Keep final schwa for ज्ञ conjunct: यज्ञ→yagya
		// ज्ञ is a special Sanskrit conjunct that retains schwa
		// Only applies when there's content before (not for isolated ज्ञ→gy)
		{
			Name:     "schwa.keep.gya-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeLanguage,
			Priority: 65,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !brahmic.IsConsonantOrConjunct(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
					return false
				}
				// Check if this is ज्ञ conjunct at word end
				if !u.IsWordFinal() || u.BaseRom != "gy" {
					return false
				}
				// Only apply if there's content before (not for isolated ज्ञ→gy)
				if u.IsWordInitial() {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
			},
		},

		// schwa.keep.iya-suffix (Language:60)
		// Keep schwa for ीय adjective endings: केंद्रीय→kendriya
		{
			Name:     "schwa.keep.iya-suffix",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeLanguage,
			Priority: 60,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !brahmic.IsConsonantOrConjunct(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
					return false
				}
				if !u.IsWordFinal() || u.BaseRom != "y" {
					return false
				}
				// Check if preceded by ी matra
				if u.Prev != nil && len(u.Prev.Runes) == 1 && u.Prev.Runes[0] == 'ी' {
					return true
				}
				return false
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
			},
		},
	}
	return append(hindiSchwa, brahmic.SchwaRules()...)
}

// consonantRules returns all consonant modification rules.
func consonantRules() []core.Rule {
	return []core.Rule{
		// consonant.pha-to-ph.before-uu (Language:58)
		// फ→ph when followed by ू matra: फूल→phool (not "fool")
		// This preserves the aspirated sound in native words like फूल (flower)
		// while allowing फ→f for loanwords like फिल्म→film
		{
			Name:     "consonant.pha-to-ph.before-uu",
			Phase:    core.PhaseConsonant,
			Scope:    core.ScopeLanguage,
			Priority: 58,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.BaseRom != "f" {
					return false
				}
				// Check if original character is फ (not फ़ which is genuinely 'f')
				if len(u.Runes) != 1 || u.Runes[0] != 'फ' {
					return false
				}
				// Check if followed by ू matra
				next := u.Next
				if next == nil || next.Type != core.UnitVowel {
					return false
				}
				return len(next.Runes) == 1 && next.Runes[0] == 'ू'
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "ph"
			},
		},

		// consonant.va-to-wa.conjunct (Language:55)
		// व→w in conjuncts after स, श, द, ख: स्वागत→swagat, ऐश्वर्या→aishwarya
		{
			Name:     "consonant.va-to-wa.conjunct",
			Phase:    core.PhaseConsonant,
			Scope:    core.ScopeLanguage,
			Priority: 55,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.BaseRom != "v" || !brahmic.IsAfterHalant(u) {
					return false
				}
				if u.Prev == nil {
					return false
				}
				prev := u.Prev.BaseRom
				// Use 'w' after स, श, द, ख (common semivowel conjuncts)
				return prev == "s" || prev == "sh" || prev == "d" || prev == "kh"
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "w"
			},
		},

		// consonant.va-to-wa.wala-suffix (Language:52)
		// Word-initial वाल pattern → 'w': वाली→wali, वाले→wale, वालो→walo, वालों→walon
		// This is the Hindi suffix meaning "one who has/does X"
		{
			Name:     "consonant.va-to-wa.wala-suffix",
			Phase:    core.PhaseConsonant,
			Scope:    core.ScopeLanguage,
			Priority: 52,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.BaseRom != "v" || !u.IsWordInitial() {
					return false
				}
				// Must be followed by ा matra
				next := u.Next
				if next == nil || next.Type != core.UnitVowel {
					return false
				}
				if len(next.Runes) != 1 || next.Runes[0] != 'ा' {
					return false
				}
				// Must be followed by ल (वाल pattern)
				afterVowel := next.Next
				if afterVowel == nil || !brahmic.IsConsonantOrConjunct(afterVowel) {
					return false
				}
				return isLaRune(afterVowel)
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "w"
			},
		},

		// consonant.va-to-wa.with-vowel (Language:50)
		// व + vowel matra (except इ-type) → 'w': दिवाली→diwali, भगवान→bhagwaan, हवा→hawa
		// व + consonant → 'v': अवस्था→avastha, परिवर्तन→parivartan, दिवस→divas
		// व + इ-type matra (ि, ी, े, ै) → 'v': कवि→kavi, गोविंद→govind
		// इ-type matra + व → 'v': विवाद→vivaad, विवाह→vivaah (वि+वा pattern)
		{
			Name:     "consonant.va-to-wa.with-vowel",
			Phase:    core.PhaseConsonant,
			Scope:    core.ScopeLanguage,
			Priority: 50,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.BaseRom != "v" {
					return false
				}
				// Not word-initial
				if u.IsWordInitial() {
					return false
				}
				// Must be followed by a vowel matra
				next := u.Next
				if next == nil || next.Type != core.UnitVowel {
					return false
				}
				// Check the vowel type after व
				if len(next.Runes) != 1 {
					return false
				}
				r := next.Runes[0]
				// इ-type matras after व: ि (i), ी (ii), े (e), ै (ai) → keep as 'v'
				if r == 'ि' || r == 'ी' || r == 'े' || r == 'ै' {
					return false
				}
				// व+व pattern (विवाद, विवाह): व followed by व → keep as 'v'
				// Check if the previous consonant (before the vowel) is also व
				if u.Prev != nil && u.Prev.Type == core.UnitVowel {
					if u.Prev.Prev != nil && isVaRune(u.Prev.Prev) {
						return false
					}
				}
				// Other vowel matras (ा, ु, ू, ो, ौ) → 'w'
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "w"
			},
		},
	}
}

// vowelRules returns all vowel modification rules.
func vowelRules() []core.Rule {
	return []core.Rule{
		// vowel.i-matra.e-glide (Language:70)
		// िए pattern becomes 'iye' not 'ie': किए→kiye, लिए→liye
		// The ए after ि-matra takes a y-glide for natural pronunciation
		{
			Name:     "vowel.i-matra.e-glide",
			Phase:    core.PhaseVowel,
			Scope:    core.ScopeLanguage,
			Priority: 70,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Check if this is ए (independent e vowel)
				if u.Type != core.UnitVowel {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ए' {
					return false
				}
				// Check if preceded by ि-matra
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'ि' {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				// Add y-glide: ए becomes 'ye'
				u.BaseRom = "ye"
			},
		},

		// vowel.long-aa.all (Scheme:60)
		// ा→aa for all positions when LongVowels option is enabled: गाना→gaana, बनाना→banaana
		{
			Name:        "vowel.long-aa.all",
			Phase:       core.PhaseVowel,
			Scope:       core.ScopeScheme,
			Priority:    60,
			Mode:        core.ModeAlways,
			Conditional: "LongVowels",
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !w.Options.LongVowels {
					return false
				}
				if u.Type != core.UnitVowel {
					return false
				}
				// Check if this is ा (aa matra)
				return len(u.Runes) == 1 && u.Runes[0] == 'ा'
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "aa"
			},
		},

		// vowel.long-aa.closed-final (Language:50)
		// ा→aa when in closed syllable at word end: काम→kaam, इंसान→insaan
		// Also handles ा+modifier patterns: दांत→daant, पांच→paanch, मां→maa
		// Exception: NOT applied for ांव pattern (गांव→gaon, handled by render.pattern.aanv-to-aon rule)
		{
			Name:     "vowel.long-aa.closed-final",
			Phase:    core.PhaseVowel,
			Scope:    core.ScopeLanguage,
			Priority: 50,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.Type != core.UnitVowel {
					return false
				}
				// Check if this is ा (aa matra)
				if len(u.Runes) != 1 || u.Runes[0] != 'ा' {
					return false
				}
				// Check what follows the vowel
				next := u.Next
				if next == nil {
					return false
				}
				// Pattern 1: ा + consonant at word end (काम→kaam)
				if brahmic.IsConsonantOrConjunct(next) && next.IsWordFinal() {
					return true
				}
				// Pattern 2: ा + modifier (दांत→daant, मां→maa)
				// The modifier may be followed by consonant or be word-final
				// In either case, ा should be 'aa' (the anusvara-final-silent rule handles the 'n')
				if next.Type == core.UnitModifier {
					// Exception: ांव pattern (गांव→gaon) - don't apply aa rule here
					// This is handled by the aanv-to-aon render rule
					if len(next.Runes) == 1 && next.Runes[0] == 'ं' {
						afterModifier := next.Next
						if afterModifier != nil && brahmic.IsConsonantOrConjunct(afterModifier) {
							if isVaRune(afterModifier) {
								return false // Don't apply - let aanv-to-aon handle it
							}
						}
					}

					// Word-final modifier: मां→maa (anusvara-final-silent removes the 'n')
					if next.IsWordFinal() {
						return true
					}
					// Modifier + consonant at word end: दांत→daant
					afterModifier := next.Next
					if afterModifier != nil && brahmic.IsConsonantOrConjunct(afterModifier) && afterModifier.IsWordFinal() {
						return true
					}
				}
				return false
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "aa"
			},
		},
	}
}

// isLaRune reports whether the unit's source character is ल or ळ (both → "l").
func isLaRune(u *core.Unit) bool {
	return len(u.Runes) >= 1 && (u.Runes[0] == 'ल' || u.Runes[0] == 'ळ')
}

// isVaRune reports whether the unit's source character is व.
func isVaRune(u *core.Unit) bool {
	return len(u.Runes) >= 1 && u.Runes[0] == 'व'
}

// renderRules returns all render phase rules.
func renderRules() []core.Rule {
	return []core.Rule{
		// render.anusvara.final-silent (Language:50)
		// Word-final anusvara after ा-matra in short words is nasalization only, not 'n': मां→maa
		// This applies to monosyllabic words like मां where the ं is just nasalization
		// Longer words like कलाकृतियां keep the 'n': kalakritiyaan
		// Note: anusvara BEFORE consonant still renders as 'n': दांत→daant
		// Note: anusvara after other vowels (में→men) keeps the 'n'
		{
			Name:     "render.anusvara.final-silent",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 50,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Check if this is word-final anusvara
				if u.Type != core.UnitModifier || !u.IsWordFinal() {
					return false
				}
				// Check if it's anusvara (ं) specifically
				if len(u.Runes) != 1 || u.Runes[0] != 'ं' {
					return false
				}
				// Check if preceded by ा-matra (aa vowel) specifically
				// This rule only applies to मां pattern, not में pattern
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				// Must be ा-matra specifically
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'ा' {
					return false
				}
				// Only suppress 'n' in short words (monosyllabic like मां)
				// Count consonants in the word - if more than 1, keep the 'n'
				consonantCount := 0
				for _, unit := range w.Units {
					if unit.Type == core.UnitConsonant {
						consonantCount++
					}
				}
				return consonantCount <= 1
			},
			Action: func(u *core.Unit, w *core.Word) {
				// Suppress the 'n' - nasalization doesn't add a consonant
				u.BaseRom = ""
			},
		},

		// render.chandrabindu.final-silent (Language:45)
		// Word-final chandrabindu after ा-matra is nasalization only: माँ→maa
		// Chandrabindu (ँ) is always pure nasalization, never adds a consonant sound
		{
			Name:     "render.chandrabindu.final-silent",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 45,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Check if this is chandrabindu (ँ)
				if u.Type != core.UnitModifier {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ँ' {
					return false
				}
				// Check if preceded by ा-matra
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'ा' {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				// Suppress the 'n' - chandrabindu is nasalization only
				u.BaseRom = ""
			},
		},

		// render.anusvara.e-matra-final (Language:40)
		// Word-final anusvara after े-matra becomes 'in' not 'en': में→mein, करें→karein
		// Only applies at word end; medial cases like केंद्र stay as kendra
		// Skipped when SimpleNasals is enabled (that mode has its own rule)
		{
			Name:     "render.anusvara.e-matra-final",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 40,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Skip when SimpleNasals is enabled - that has its own handling
				if w.Options.SimpleNasals {
					return false
				}
				// Must be word-final
				if !u.IsWordFinal() {
					return false
				}
				// Check if this is anusvara (ं)
				if u.Type != core.UnitModifier {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ं' {
					return false
				}
				// Check if preceded by े-matra (e vowel)
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'े' {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				// Change 'n' to 'in' for में→mein pattern
				u.BaseRom = "in"
			},
		},

		// render.anusvara.e-matra-final.simple (Scheme:41)
		// When SimpleNasals is enabled, only apply 'ein' to monosyllabic words (में→mein)
		// Multi-syllable words use plain 'en': करें→karen, चलें→chalen
		{
			Name:        "render.anusvara.e-matra-final.simple",
			Phase:       core.PhaseRender,
			Scope:       core.ScopeScheme,
			Priority:    41,
			Mode:        core.ModeAlways,
			Conditional: "SimpleNasals",
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Only applies when SimpleNasals option is enabled
				if !w.Options.SimpleNasals {
					return false
				}
				// Must be word-final
				if !u.IsWordFinal() {
					return false
				}
				// Check if this is anusvara (ं)
				if u.Type != core.UnitModifier {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ं' {
					return false
				}
				// Check if preceded by े-matra (e vowel)
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'े' {
					return false
				}
				// Count consonants before this pattern
				// में = म + ें (1 consonant) → mein
				// करें = क + र + ें (2 consonants) → karen
				consonantCount := 0
				for curr := u.Prev.Prev; curr != nil; curr = curr.Prev {
					if curr.Type == core.UnitConsonant || curr.Type == core.UnitConjunct {
						consonantCount++
					}
				}
				// Only apply 'ein' if single consonant before ें (monosyllabic)
				return consonantCount == 1
			},
			Action: func(u *core.Unit, w *core.Word) {
				// Change 'n' to 'in' for में→mein pattern (monosyllabic only)
				u.BaseRom = "in"
			},
		},

		// render.anusvara.before-labial (Language:35)
		// Anusvara before labial consonants (प, ब, भ, म) becomes 'm': संभव→sambhav
		// This is standard Hindi phonology - nasal assimilates to following consonant
		{
			Name:     "render.anusvara.before-labial",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 35,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Check if this is anusvara (ं)
				if u.Type != core.UnitModifier {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ं' {
					return false
				}
				// Check if followed by a labial consonant (प, ब, भ, म)
				next := u.Next
				if next == nil || !brahmic.IsConsonantOrConjunct(next) {
					return false
				}
				// Check if the consonant is a labial
				baseRom := next.BaseRom
				return baseRom == "p" || baseRom == "b" || baseRom == "bh" || baseRom == "m" || baseRom == "ph"
			},
			Action: func(u *core.Unit, w *core.Word) {
				u.BaseRom = "m"
			},
		},

		// render.pattern.aanv-to-aon (Language:30)
		// ांव pattern becomes 'aon' not 'aanv': गांव→gaon
		// This is a common Hindi spelling convention where the व is semi-silent
		// and the nasalization creates the 'n' sound
		{
			Name:     "render.pattern.aanv-to-aon",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 30,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				// Check if this is anusvara (ं)
				if u.Type != core.UnitModifier {
					return false
				}
				if len(u.Runes) != 1 || u.Runes[0] != 'ं' {
					return false
				}
				// Check if preceded by ा-matra
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				if len(u.Prev.Runes) != 1 || u.Prev.Runes[0] != 'ा' {
					return false
				}
				// Check if followed by व at word end
				next := u.Next
				if next == nil || !brahmic.IsConsonantOrConjunct(next) {
					return false
				}
				return isVaRune(next) && next.IsWordFinal()
			},
			Action: func(u *core.Unit, w *core.Word) {
				// ांव at word end becomes 'aon':
				// - ा stays as 'a' (blocked from becoming 'aa' by long-aa-closed-final exception)
				// - ं becomes 'on' (nasal diphthong)
				// - व becomes silent (we'll suppress it by setting BaseRom to "")
				u.BaseRom = "on"
				// Suppress the following व
				if u.Next != nil {
					u.Next.BaseRom = ""
				}
			},
		},
	}
}
