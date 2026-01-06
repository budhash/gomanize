package hindi

import "github.com/budhash/gomanize/engine"

// Rules returns the Hindi-specific transliteration rules.
func Rules() []engine.Rule {
	return []engine.Rule{
		// === PhaseSchwa Rules ===
		// Rules are ordered by effective priority (highest first within each scope)

		// schwa-keep-initial-conjunct (Script:80)
		// Keep schwa for word-initial conjuncts: प्रकाश→prakaash (not prkaash)
		{
			Name:     "schwa-keep-initial-conjunct",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeScript,
			Priority: 80,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				return isConsonantOrConjunct(u) &&
					u.Schwa == engine.SchwaPending &&
					u.AfterHalant &&
					u.RunIndex <= 1
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaKeep
			},
		},

		// schwa-keep-sonorous-final (Script:70)
		// Keep final schwa after halant for र, य, व: मंत्र→mantra
		{
			Name:     "schwa-keep-sonorous-final",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeScript,
			Priority: 70,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				return isConsonantOrConjunct(u) &&
					u.Schwa == engine.SchwaPending &&
					u.IsWordFinal() &&
					u.AfterHalant &&
					isSonorousConsonant(u.BaseRom)
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaKeep
			},
		},

		// schwa-keep-iya-suffix (Language:60)
		// Keep schwa for ीय adjective endings: केंद्रीय→kendriya
		{
			Name:     "schwa-keep-iya-suffix",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeLanguage,
			Priority: 60,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if !isConsonantOrConjunct(u) || u.Schwa != engine.SchwaPending {
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
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaKeep
			},
		},

		// schwa-delete-ccv (Script:50)
		// Delete medial schwa in C+C+V pattern: जनता→janta, कमला→kamla
		{
			Name:     "schwa-delete-ccv",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeScript,
			Priority: 50,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if !isConsonantOrConjunct(u) || u.Schwa != engine.SchwaPending {
					return false
				}
				// Must be in a run and not word-initial
				if u.Run == nil || u.RunIndex == 0 {
					return false
				}
				// Only one deletion per run
				if u.Run.HasDeletion() {
					return false
				}
				// Must have a following consonant
				next := u.Next
				if next == nil || !isConsonantOrConjunct(next) {
					return false
				}
				// That consonant must be followed by a vowel
				afterNext := next.Next
				if afterNext == nil || afterNext.Type != engine.UnitVowel {
					return false
				}
				return true
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaDelete
				if u.Run != nil {
					u.Run.DeletedAt = u.RunIndex
				}
			},
		},

		// schwa-delete-ccc-final (Script:45)
		// Delete schwa in C+C+C+END pattern (consonant-ending words)
		// Examples: मकसद→maksad, झटपट→jhatpat, सरगम→sargam
		// Only applies at index 1 to avoid cascading deletions
		{
			Name:     "schwa-delete-ccc-final",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeScript,
			Priority: 45,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if !isConsonantOrConjunct(u) || u.Schwa != engine.SchwaPending {
					return false
				}
				// Only at index 1 (second consonant in word)
				if u.Run == nil || u.RunIndex != 1 {
					return false
				}
				// Only one deletion per run
				if u.Run.HasDeletion() {
					return false
				}
				// Must have a following consonant
				next := u.Next
				if next == nil || !isConsonantOrConjunct(next) {
					return false
				}
				// Check if word ends in consonants (no trailing vowel)
				hasTrailingVowel := false
				for cur := next.Next; cur != nil; cur = cur.Next {
					if cur.Type == engine.UnitVowel {
						hasTrailingVowel = true
						break
					}
				}
				return !hasTrailingVowel
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaDelete
				if u.Run != nil {
					u.Run.DeletedAt = u.RunIndex
				}
			},
		},

		// schwa-keep-before-anusvara (Script:40)
		// Keep schwa when followed by anusvara: सुमन→suman (not sumn)
		{
			Name:     "schwa-keep-before-anusvara",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeScript,
			Priority: 40,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if !isConsonantOrConjunct(u) || u.Schwa != engine.SchwaPending {
					return false
				}
				// Check if next unit is anusvara (ं)
				if u.Next != nil && len(u.Next.Runes) == 1 && u.Next.Runes[0] == 'ं' {
					return true
				}
				return false
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaKeep
			},
		},

		// schwa-delete-word-final (Universal:10)
		// Delete schwa at word end (unless protected by higher rules)
		{
			Name:     "schwa-delete-word-final",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeUniversal,
			Priority: 10,
			Mode:     engine.ModeExclusive,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				return isConsonantOrConjunct(u) &&
					u.Schwa == engine.SchwaPending &&
					u.IsWordFinal() &&
					!u.AfterHalant // Not protected by sonorous-final rule
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaDelete
			},
		},

		// schwa-keep-default (Universal:0, Fallback)
		// Default: keep schwa if no other rule decided
		{
			Name:     "schwa-keep-default",
			Phase:    engine.PhaseSchwa,
			Scope:    engine.ScopeUniversal,
			Priority: 0,
			Mode:     engine.ModeFallback,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				return isConsonantOrConjunct(u) && u.Schwa == engine.SchwaPending
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.Schwa = engine.SchwaKeep
			},
		},

		// === PhaseConsonant Rules ===

		// va-to-wa-conjunct (Language:50)
		// व→w after स, श, द, ख: स्वागत→swagat, ऐश्वर्या→aishwarya
		{
			Name:     "va-to-wa-conjunct",
			Phase:    engine.PhaseConsonant,
			Scope:    engine.ScopeLanguage,
			Priority: 50,
			Mode:     engine.ModeAlways,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if u.BaseRom != "v" || !u.AfterHalant {
					return false
				}
				if u.Prev == nil {
					return false
				}
				prev := u.Prev.BaseRom
				// Use 'w' only after स, श, द, ख (common semivowel conjuncts)
				return prev == "s" || prev == "sh" || prev == "d" || prev == "kh"
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.BaseRom = "w"
			},
		},

		// === PhaseVowel Rules ===

		// long-aa-closed-final (Language:50)
		// ा→aa when followed by consonant at word end: काम→kaam, इंसान→insaan
		{
			Name:     "long-aa-closed-final",
			Phase:    engine.PhaseVowel,
			Scope:    engine.ScopeLanguage,
			Priority: 50,
			Mode:     engine.ModeAlways,
			Condition: func(u *engine.Unit, w *engine.Word) bool {
				if u.Type != engine.UnitVowel {
					return false
				}
				// Check if this is ा (aa matra)
				if len(u.Runes) != 1 || u.Runes[0] != 'ा' {
					return false
				}
				// Must be followed by consonant at word end
				next := u.Next
				if next == nil || !isConsonantOrConjunct(next) {
					return false
				}
				// That consonant must be at word end
				return next.IsWordFinal()
			},
			Action: func(u *engine.Unit, w *engine.Word) {
				u.BaseRom = "aa"
			},
		},
	}
}

// isConsonantOrConjunct returns true if the unit is a consonant or conjunct.
func isConsonantOrConjunct(u *engine.Unit) bool {
	return u.Type == engine.UnitConsonant || u.Type == engine.UnitConjunct
}

// isSonorousConsonant returns true for र, य, व which retain schwa in Sanskrit words.
func isSonorousConsonant(baseRom string) bool {
	return baseRom == "r" || baseRom == "y" || baseRom == "v"
}
