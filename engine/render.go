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
// For Phase 1, this is a simple concatenation of BaseRom values.
// Schwa handling will be added in Phase 2 with rules.
func (r *Renderer) Render(word *Word) string {
	var sb strings.Builder

	for _, unit := range word.Units {
		sb.WriteString(unit.BaseRom)

		// Basic schwa insertion for consonants/conjuncts
		// Phase 1: Simple inherent vowel 'a' after each consonant
		// This will be refined by schwa deletion rules in Phase 2
		if (unit.Type == UnitConsonant || unit.Type == UnitConjunct) && !unit.AfterHalant {
			// Only add schwa if not suppressed
			if unit.Schwa != SchwaDelete {
				// Check if followed by a matra (dependent vowel)
				// If so, don't add schwa
				if unit.Next == nil || unit.Next.Type != UnitVowel {
					sb.WriteString("a")
				}
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
