package brahmic

import "github.com/budhash/gomanize/core"

// SchwaRules returns the script-level (Brahmic-general) schwa rules shared by all
// Brahmic languages: the sonorous-final keep, the CCV / CCCC / before-CC medial
// deletions, word-final deletion, and the keep-by-default fallback. Language
// packages compose these with their own language-specific schwa rules.
//
// These were previously defined inside lang/hindi; they contain no Hindi-only
// logic (only Devanagari-general character checks), so they live here to be
// reused by Marathi, Nepali, etc.
func SchwaRules() []core.Rule {
	return []core.Rule{
		// schwa.keep.sonorous-final (Script:70)
		// Keep final schwa after halant for र, य, व: मंत्र→mantra
		{
			Name:     "schwa.keep.sonorous-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeScript,
			Priority: 70,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return IsConsonantOrConjunct(u) &&
					GetSchwa(u) == SchwaPending &&
					u.IsWordFinal() &&
					IsAfterHalant(u) &&
					isSonorousRune(u)
			},
			Action: func(u *core.Unit, w *core.Word) {
				SetSchwa(u, SchwaKeep)
			},
		},

		// schwa.delete.ccv (Script:50)
		// Delete medial schwa in C+C+V pattern: जनता→janta, कमला→kamla, अपना→apna
		// Also applies for C+C+Modifier (anusvara/chandrabindu): झारखंड→jharkhand
		// Does NOT apply to word-initial conjuncts: प्रकाश→prakaash (not "prkaash")
		// Does NOT apply after halant (conjuncts): पार्वती→parvati (not "parvti")
		// Disabled when KeepMedialSchwa option is enabled
		{
			Name:        "schwa.delete.ccv",
			Phase:       core.PhaseSchwa,
			Scope:       core.ScopeScript,
			Priority:    50,
			Mode:        core.ModeExclusive,
			Conditional: "!KeepMedialSchwa",
			Condition: func(u *core.Unit, w *core.Word) bool {
				if w.Options.KeepMedialSchwa {
					return false
				}
				if !IsConsonantOrConjunct(u) || GetSchwa(u) != SchwaPending {
					return false
				}
				// Must not be at absolute word start (first character)
				if u.IsWordInitial() {
					return false
				}
				// Do not delete schwa after halant (part of conjunct)
				if IsAfterHalant(u) {
					return false
				}
				run := GetRun(u)
				if run == nil {
					return false
				}
				if run.HasDeletion() {
					return false
				}
				next := u.Next
				if next == nil || !IsConsonantOrConjunct(next) {
					return false
				}
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
				SetSchwa(u, SchwaDelete)
				run := GetRun(u)
				if run != nil {
					run.DeletedAt = GetRunIndex(u)
				}
			},
		},

		// schwa.delete.cccc-final (Script:45)
		// Delete schwa in C+C+C+C+END pattern (4+ consonant words ending in consonants)
		// Examples: मकसद→maksad, झटपट→jhatpat. Only at RUNE index 1.
		// Does NOT apply to word-initial conjuncts: प्रथम→pratham. Disabled by KeepMedialSchwa.
		{
			Name:        "schwa.delete.cccc-final",
			Phase:       core.PhaseSchwa,
			Scope:       core.ScopeScript,
			Priority:    45,
			Mode:        core.ModeExclusive,
			Conditional: "!KeepMedialSchwa",
			Condition: func(u *core.Unit, w *core.Word) bool {
				if w.Options.KeepMedialSchwa {
					return false
				}
				if !IsConsonantOrConjunct(u) || GetSchwa(u) != SchwaPending {
					return false
				}
				if IsAfterHalant(u) && isWordInitialConjunct(u, w) {
					return false
				}
				// Only at RUNE index 1 (second character in original string)
				if u.Start.Rune != 1 {
					return false
				}
				run := GetRun(u)
				if run != nil && run.HasDeletion() {
					return false
				}
				next := u.Next
				if next == nil || !IsConsonantOrConjunct(next) {
					return false
				}
				afterNext := next.Next
				if afterNext == nil || !IsConsonantOrConjunct(afterNext) {
					return false
				}
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
				SetSchwa(u, SchwaDelete)
				run := GetRun(u)
				if run != nil {
					run.DeletedAt = GetRunIndex(u)
				}
			},
		},

		// schwa.delete.before-cc (Script:42)
		// Delete schwa when consonant is followed by 2 separate consonants ending at
		// word-final: देशभर→deshbhar, अमृतसर→amritsar, मेहनत→mehnat. Not for conjuncts,
		// not after ा (Sanskrit heuristic). Disabled by KeepMedialSchwa.
		{
			Name:        "schwa.delete.before-cc",
			Phase:       core.PhaseSchwa,
			Scope:       core.ScopeScript,
			Priority:    42,
			Mode:        core.ModeExclusive,
			Conditional: "!KeepMedialSchwa",
			Condition: func(u *core.Unit, w *core.Word) bool {
				if w.Options.KeepMedialSchwa {
					return false
				}
				if !IsConsonantOrConjunct(u) || GetSchwa(u) != SchwaPending {
					return false
				}
				if u.IsWordInitial() {
					return false
				}
				if u.Prev == nil || u.Prev.Type != core.UnitVowel {
					return false
				}
				// Don't delete after ा (aa-matra): often Sanskrit words retaining schwa.
				if len(u.Prev.Runes) == 1 && u.Prev.Runes[0] == 'ा' {
					return false
				}
				run := GetRun(u)
				if run != nil && run.HasDeletion() {
					return false
				}
				next := u.Next
				if next == nil || !IsConsonantOrConjunct(next) {
					return false
				}
				afterNext := next.Next
				if afterNext == nil || !IsConsonantOrConjunct(afterNext) {
					return false
				}
				if IsAfterHalant(afterNext) {
					return false
				}
				if !afterNext.IsWordFinal() {
					return false
				}
				return true
			},
			Action: func(u *core.Unit, w *core.Word) {
				SetSchwa(u, SchwaDelete)
				run := GetRun(u)
				if run != nil {
					run.DeletedAt = GetRunIndex(u)
				}
			},
		},

		// schwa.delete.word-final (Universal:10)
		// Delete schwa at word end (unless protected by higher rules).
		{
			Name:     "schwa.delete.word-final",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeUniversal,
			Priority: 10,
			Mode:     core.ModeExclusive,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return IsConsonantOrConjunct(u) &&
					GetSchwa(u) == SchwaPending &&
					u.IsWordFinal()
			},
			Action: func(u *core.Unit, w *core.Word) {
				SetSchwa(u, SchwaDelete)
			},
		},

		// schwa.keep.default (Universal:0, Fallback)
		// Default: keep schwa if no other rule decided.
		{
			Name:     "schwa.keep.default",
			Phase:    core.PhaseSchwa,
			Scope:    core.ScopeUniversal,
			Priority: 0,
			Mode:     core.ModeFallback,
			Condition: func(u *core.Unit, w *core.Word) bool {
				return IsConsonantOrConjunct(u) && GetSchwa(u) == SchwaPending
			},
			Action: func(u *core.Unit, w *core.Word) {
				SetSchwa(u, SchwaKeep)
			},
		},
	}
}

// isSonorousRune reports whether the unit's source character is र, य, or व — the
// sonorous consonants that retain a word-final schwa in Sanskrit-derived words.
func isSonorousRune(u *core.Unit) bool {
	if len(u.Runes) == 0 {
		return false
	}
	r := u.Runes[0]
	return r == 'र' || r == 'य' || r == 'व'
}

// isWordInitialConjunct reports whether the unit is part of a word-initial
// conjunct: a halant at index 1 (C्C) or index 2 after an independent vowel
// (अC्C). Word-initial conjuncts protect their schwa (प्रकाश→prakaash).
func isWordInitialConjunct(u *core.Unit, w *core.Word) bool {
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
	switch unitIdx {
	case 1:
		if len(w.Units) > 0 && w.Units[0].Type == core.UnitConsonant {
			return true
		}
	case 2:
		if len(w.Units) >= 2 {
			firstUnit := w.Units[0]
			secondUnit := w.Units[1]
			if firstUnit.Type == core.UnitVowel && secondUnit.Type == core.UnitConsonant {
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
