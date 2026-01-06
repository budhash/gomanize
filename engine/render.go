package engine

import (
	"strconv"
	"strings"
)

// Renderer converts a parsed Word into romanized output.
type Renderer struct {
	// Future: scheme support, rule engine, etc.
}

// NewRenderer creates a new renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render converts a Word to its romanized string representation.
// Uses SchwaState decisions from the rule engine to determine schwa output.
func (r *Renderer) Render(word *Word) string {
	var sb strings.Builder

	for _, unit := range word.Units {
		sb.WriteString(unit.BaseRom)

		// Schwa handling for consonants/conjuncts
		if unit.Type == UnitConsonant || unit.Type == UnitConjunct {
			// Skip schwa if followed by vowel/matra (vowel provides the sound)
			// Note: UnitModifier (anusvara, visarga, chandrabindu) does NOT suppress schwa
			if unit.Next != nil && unit.Next.Type == UnitVowel {
				continue
			}

			// Skip schwa if next unit is part of a conjunct (came after halant)
			// This means current consonant + halant + next consonant form a cluster
			// Example: प्र in प्रकाश - प has no schwa because ् + र follows
			if unit.Next != nil && unit.Next.AfterHalant {
				continue
			}

			// Use SchwaState decision from rules
			// - SchwaKeep: add "a"
			// - SchwaDelete: no schwa
			// - SchwaPending: treat as Keep (rules didn't run or fallback applies)
			if unit.Schwa != SchwaDelete {
				sb.WriteString("a")
			}
		}
	}

	return sb.String()
}

// RenderDebug returns a detailed debug representation of the word.
func (r *Renderer) RenderDebug(word *Word) string {
	var sb strings.Builder

	sb.WriteString("Word: ")
	sb.WriteString(word.Original)
	sb.WriteString("\n")
	sb.WriteString("Units:\n")

	for i, unit := range word.Units {
		sb.WriteString("  [")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("] ")
		sb.WriteString(string(unit.Runes))
		sb.WriteString(" → ")
		sb.WriteString(unit.BaseRom)
		sb.WriteString(" (")
		sb.WriteString(unit.Type.String())
		sb.WriteString(")")

		if unit.AfterHalant {
			sb.WriteString(" [after-halant]")
		}
		if unit.Run != nil {
			sb.WriteString(" [run:")
			sb.WriteString(strconv.Itoa(unit.RunIndex))
			sb.WriteString("]")
		}
		if unit.Type == UnitConsonant || unit.Type == UnitConjunct {
			sb.WriteString(" schwa=")
			sb.WriteString(unit.Schwa.String())
		}
		sb.WriteString("\n")
	}

	if len(word.Runs) > 0 {
		sb.WriteString("Runs:\n")
		for i, run := range word.Runs {
			sb.WriteString("  Run ")
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(": ")

			for j, u := range run.Units {
				if j > 0 {
					sb.WriteString("-")
				}
				sb.WriteString(string(u.Runes))
			}

			if run.PrevVowel != nil {
				sb.WriteString(" (after ")
				sb.WriteString(string(run.PrevVowel.Runes))
				sb.WriteString(")")
			} else {
				sb.WriteString(" (word-initial)")
			}
			if run.NextVowel != nil {
				sb.WriteString(" (before ")
				sb.WriteString(string(run.NextVowel.Runes))
				sb.WriteString(")")
			} else {
				sb.WriteString(" (word-final)")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("Output: ")
	sb.WriteString(r.Render(word))

	return sb.String()
}
