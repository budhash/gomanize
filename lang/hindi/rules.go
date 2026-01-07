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
	return []core.Rule{
		// schwa-keep-sonorous-final (Script:70)
		// Keep final schwa after halant for र, य, व: मंत्र→mantra
		{
			Name:     "schwa-keep-sonorous-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeScript,
			Priority: 70,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return brahmic.IsConsonantOrConjunct(u) &&
					brahmic.GetSchwa(u) == brahmic.SchwaPending &&
					u.IsWordFinal() &&
					brahmic.IsAfterHalant(u) &&
					isSonorousConsonant(u.BaseRom)
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
			},
		},

		// schwa-keep-iya-suffix (Language:60)
		// Keep schwa for ीय adjective endings: केंद्रीय→kendriya
		{
			Name:     "schwa-keep-iya-suffix",
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

		// schwa-delete-ccv (Script:50)
		// Delete medial schwa in C+C+V pattern: जनता→janta, कमला→kamla, अपना→apna
		// Also applies for C+C+Modifier (anusvara/chandrabindu): झारखंड→jharkhand
		// Does NOT apply to word-initial conjuncts: प्रकाश→prakaash (not "prkaash")
		{
			Name:     "schwa-delete-ccv",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeScript,
			Priority: 50,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !brahmic.IsConsonantOrConjunct(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
					return false
				}
				// Must not be at absolute word start (first character)
				if u.IsWordInitial() {
					return false
				}
				// Do not delete schwa for word-initial conjuncts
				// Word-initial conjuncts should keep their schwa (प्रकाश→prakaash)
				if brahmic.IsAfterHalant(u) && isWordInitialConjunct(u, w) {
					return false
				}
				// Must be in a run
				run := brahmic.GetRun(u)
				if run == nil {
					return false
				}
				// Only one deletion per run
				if run.HasDeletion() {
					return false
				}
				// Must have a following consonant
				next := u.Next
				if next == nil || !brahmic.IsConsonantOrConjunct(next) {
					return false
				}
				// That consonant must be followed by a vowel or modifier (anusvara, etc.)
				afterNext := next.Next
				if afterNext == nil {
					return false
				}
				if afterNext.Type != core.UnitVowel && afterNext.Type != core.UnitModifier {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaDelete)
				run := brahmic.GetRun(u)
				if run != nil {
					run.DeletedAt = brahmic.GetRunIndex(u)
				}
			},
		},

		// schwa-delete-cccc-final (Script:45)
		// Delete schwa in C+C+C+C+END pattern (4+ consonant words ending in consonants)
		// Examples: मकसद→maksad, झटपट→jhatpat
		// Only applies at RUNE index 1 (second character in original string)
		// and requires at least 2 more consonants after
		// This does NOT apply to 3-consonant words like कमल→kamal, गरम→garam
		// Does NOT apply to word-initial conjuncts: प्रथम→pratham (not "prtham")
		{
			Name:     "schwa-delete-cccc-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeScript,
			Priority: 45,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if !brahmic.IsConsonantOrConjunct(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
					return false
				}
				// Do not delete schwa for word-initial conjuncts
				if brahmic.IsAfterHalant(u) && isWordInitialConjunct(u, w) {
					return false
				}
				// Only at RUNE index 1 (second character in original string)
				// This matches the old engine's sb.index == 1 check
				if u.Start.Rune != 1 {
					return false
				}
				// Only one deletion per run
				run := brahmic.GetRun(u)
				if run != nil && run.HasDeletion() {
					return false
				}
				// Must have TWO following consonants (CCCC pattern, not CCC)
				next := u.Next
				if next == nil || !brahmic.IsConsonantOrConjunct(next) {
					return false
				}
				afterNext := next.Next
				if afterNext == nil || !brahmic.IsConsonantOrConjunct(afterNext) {
					return false
				}
				// Check if word ends in consonants (no trailing vowel)
				hasTrailingVowel := false
				for cur := afterNext.Next; cur != nil; cur = cur.Next {
					if cur.Type == core.UnitVowel {
						hasTrailingVowel = true
						break
					}
				}
				return !hasTrailingVowel
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaDelete)
				run := brahmic.GetRun(u)
				if run != nil {
					run.DeletedAt = brahmic.GetRunIndex(u)
				}
			},
		},

		// schwa-delete-word-final (Universal:10)
		// Delete schwa at word end (unless protected by higher rules)
		// Note: schwa-keep-sonorous-final (Script:70) runs first to protect र, य, व
		{
			Name:     "schwa-delete-word-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeUniversal,
			Priority: 10,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return brahmic.IsConsonantOrConjunct(u) &&
					brahmic.GetSchwa(u) == brahmic.SchwaPending &&
					u.IsWordFinal()
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaDelete)
			},
		},

		// schwa-keep-default (Universal:0, Fallback)
		// Default: keep schwa if no other rule decided
		{
			Name:     "schwa-keep-default",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeUniversal,
			Priority: 0,
			Mode:     core.ModeFallback,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return brahmic.IsConsonantOrConjunct(u) && brahmic.GetSchwa(u) == brahmic.SchwaPending
			},
			Action: func(u *core.Unit, w *core.Word) {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
			},
		},
	}
}

// consonantRules returns all consonant modification rules.
func consonantRules() []core.Rule {
	return []core.Rule{
		// va-to-wa-conjunct (Language:50)
		// व→w after स, श, द, ख: स्वागत→swagat, ऐश्वर्या→aishwarya
		{
			Name:     "va-to-wa-conjunct",
			Phase:    core.PhaseConsonant,
			Scope:    core.ScopeLanguage,
			Priority: 50,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
				if u.BaseRom != "v" || !brahmic.IsAfterHalant(u) {
					return false
				}
				if u.Prev == nil {
					return false
				}
				prev := u.Prev.BaseRom
				// Use 'w' only after स, श, द, ख (common semivowel conjuncts)
				return prev == "s" || prev == "sh" || prev == "d" || prev == "kh"
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
		// long-aa-all (Scheme:60)
		// ा→aa for all positions when LongVowels option is enabled: गाना→gaana, बनाना→banaana
		{
			Name:     "long-aa-all",
			Phase:    core.PhaseVowel,
			Scope:    core.ScopeScheme,
			Priority: 60,
			Mode:     core.ModeAlways,
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

		// long-aa-closed-final (Language:50)
		// ा→aa when in closed syllable at word end: काम→kaam, इंसान→insaan
		// Also handles ा+modifier patterns: दांत→daant, पांच→paanch, मां→maa
		{
			Name:     "long-aa-closed-final",
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

// isSonorousConsonant returns true for र, य, व which retain schwa in Sanskrit words.
func isSonorousConsonant(baseRom string) bool {
	return baseRom == "r" || baseRom == "y" || baseRom == "v"
}

// isWordInitialConjunct returns true if the unit is part of a word-initial conjunct.
// Word-initial conjunct means: halant at position 1 (C्C) or position 2 after independent vowel (अC्C).
// This determines whether schwa should be protected after the conjunct.
// Examples:
//   - प्रकाश: प्र is word-initial conjunct (halant at index 1) → schwa protected → "prakaash"
//   - अध्यक्ष: ध्य is word-initial conjunct (halant at index 2 after अ) → schwa protected → "adhyaksh"
//   - कर्मकांड: र्म is NOT word-initial (halant at index 2, but after क not vowel) → schwa deleted → "karmkand"
func isWordInitialConjunct(u *core.Unit, w *core.Word) bool {
	// Find the unit's position in the word
	unitIdx := -1
	for i, unit := range w.Units {
		if unit == u {
			unitIdx = i
			break
		}
	}
	if unitIdx < 0 {
		return false
	}

	// The halant should be immediately before this unit
	// In the parsed structure, halant is absorbed into the conjunct parsing
	// So we need to check the original rune positions

	// For word-initial conjunct:
	// Case 1: Unit at index 1, prev unit is consonant (C्C pattern)
	// Case 2: Unit at index 2, unit[0] is independent vowel, unit[1] is consonant (अC्C pattern)
	switch unitIdx {
	case 1:
		// Check if first unit is a consonant (C्C pattern)
		if len(w.Units) > 0 && w.Units[0].Type == core.UnitConsonant {
			return true
		}
	case 2:
		// Check if first unit is independent vowel and second is consonant (अC्C pattern)
		if len(w.Units) >= 2 {
			firstUnit := w.Units[0]
			secondUnit := w.Units[1]
			// Independent vowels in Devanagari: अ-औ (U+0905 to U+0914)
			if firstUnit.Type == core.UnitVowel && secondUnit.Type == core.UnitConsonant {
				// Check if it's a standalone vowel (not a matra)
				if len(firstUnit.Runes) == 1 {
					r := firstUnit.Runes[0]
					if r >= 0x0905 && r <= 0x0914 {
						return true
					}
				}
			}
		}
	}

	return false
}

// renderRules returns all render phase rules.
func renderRules() []core.Rule {
	return []core.Rule{
		// anusvara-final-silent (Language:50)
		// Word-final anusvara after ा-matra in short words is nasalization only, not 'n': मां→maa
		// This applies to monosyllabic words like मां where the ं is just nasalization
		// Longer words like कलाकृतियां keep the 'n': kalakritiyaan
		// Note: anusvara BEFORE consonant still renders as 'n': दांत→daant
		// Note: anusvara after other vowels (में→men) keeps the 'n'
		{
			Name:     "anusvara-final-silent",
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

		// chandrabindu-final-silent (Language:45)
		// Word-final chandrabindu after ा-matra is nasalization only: माँ→maa
		// Chandrabindu (ँ) is always pure nasalization, never adds a consonant sound
		{
			Name:     "chandrabindu-final-silent",
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

		// anusvara-after-e-matra-final (Language:40)
		// Word-final anusvara after े-matra becomes 'in' not 'en': में→mein
		// Only applies at word end; medial cases like केंद्र stay as kendra
		{
			Name:     "anusvara-after-e-matra-final",
			Phase:    core.PhaseRender,
			Scope:    core.ScopeLanguage,
			Priority: 40,
			Mode:     core.ModeAlways,
			Condition: func(u *core.Unit, w *core.Word) bool {
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
	}
}
