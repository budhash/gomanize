package brahmic

// IdentifyRuns groups consecutive consonants/conjuncts between vowels into ConsonantRuns.
// This enables coordinated schwa deletion decisions within each run.
func IdentifyRuns(word *Word) {
	var currentRun *ConsonantRun
	var prevVowel *Unit

	for _, unit := range word.Units {
		switch unit.Type {
		case UnitConsonant, UnitConjunct:
			// Start new run if not in one
			if currentRun == nil {
				currentRun = NewConsonantRun()
				currentRun.PrevVowel = prevVowel
				word.Runs = append(word.Runs, currentRun)
			}

			// Add consonant to run
			unit.Run = currentRun
			unit.RunIndex = len(currentRun.Units)
			currentRun.Units = append(currentRun.Units, unit)

		case UnitVowel, UnitModifier:
			// Vowels and modifiers (anusvara, visarga, chandrabindu) close runs
			// Close current run if open
			if currentRun != nil {
				currentRun.NextVowel = unit
				currentRun = nil
			}
			prevVowel = unit

		default:
			// Numbers and symbols don't affect runs
			// But they do close any open run (conservative approach)
			if currentRun != nil {
				currentRun = nil
			}
		}
	}
	// Note: word-final runs have NextVowel = nil (handled correctly)
}

// RunStats returns statistics about consonant runs in the word.
type RunStats struct {
	TotalRuns        int
	WordInitialRuns  int // Runs with PrevVowel == nil
	WordFinalRuns    int // Runs with NextVowel == nil
	MaxRunLength     int
	SingleConsonants int // Runs with exactly one consonant
}

// GetRunStats calculates statistics about consonant runs.
func GetRunStats(word *Word) RunStats {
	stats := RunStats{
		TotalRuns: len(word.Runs),
	}

	for _, run := range word.Runs {
		if run.PrevVowel == nil {
			stats.WordInitialRuns++
		}
		if run.NextVowel == nil {
			stats.WordFinalRuns++
		}
		if len(run.Units) > stats.MaxRunLength {
			stats.MaxRunLength = len(run.Units)
		}
		if len(run.Units) == 1 {
			stats.SingleConsonants++
		}
	}

	return stats
}
