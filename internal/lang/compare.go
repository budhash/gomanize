package lang

import (
	"fmt"
	"strings"

	"github.com/budhash/gomanize/core"
	hindiLang "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// CompareResult holds the output of comparing old and new engines.
type CompareResult struct {
	Input     string
	OldOutput string
	NewOutput string
	Match     bool
}

// Comparer runs both old and new engines for comparison.
type Comparer struct {
	oldEngine Hindi        // existing implementation (internal/lang)
	newEngine *core.Engine // new core engine
}

// NewComparer creates a new Comparer instance.
func NewComparer() *Comparer {
	return &Comparer{
		oldEngine: Hindi{},
		newEngine: core.NewEngine(hindiLang.Hindi{}, colloquial.Colloquial{}),
	}
}

// Compare runs both engines on the input and returns the results.
func (c *Comparer) Compare(input string) CompareResult {
	oldOutput := c.oldEngine.Transliterate(input)
	newOutput := c.newEngine.Transliterate(input)

	return CompareResult{
		Input:     input,
		OldOutput: oldOutput,
		NewOutput: newOutput,
		Match:     oldOutput == newOutput,
	}
}

// CompareWithDebug returns comparison with debug info from new engine.
// Note: Debug output is no longer available in core.Engine - returns empty string.
func (c *Comparer) CompareWithDebug(input string) (CompareResult, string) {
	result := c.Compare(input)
	// Debug output is not available in core.Engine
	return result, ""
}

// FormatResult formats a CompareResult for display.
func FormatResult(r CompareResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Input:  %s\n", r.Input))
	sb.WriteString(fmt.Sprintf("Old:    %s\n", r.OldOutput))
	sb.WriteString(fmt.Sprintf("New:    %s\n", r.NewOutput))
	if r.Match {
		sb.WriteString("Status: ✓ Match\n")
	} else {
		sb.WriteString("Status: ✗ Different\n")
	}
	return sb.String()
}

// BatchCompare compares multiple inputs and returns statistics.
type BatchStats struct {
	Total      int
	Matches    int
	Mismatches int
	Results    []CompareResult
}

// BatchCompare compares multiple inputs.
func (c *Comparer) BatchCompare(inputs []string) BatchStats {
	stats := BatchStats{
		Total:   len(inputs),
		Results: make([]CompareResult, 0, len(inputs)),
	}

	for _, input := range inputs {
		result := c.Compare(input)
		stats.Results = append(stats.Results, result)
		if result.Match {
			stats.Matches++
		} else {
			stats.Mismatches++
		}
	}

	return stats
}

// FormatBatchStats formats batch comparison statistics.
func FormatBatchStats(stats BatchStats) string {
	var sb strings.Builder

	// Guard against division by zero for empty input
	var matchPct float64
	if stats.Total > 0 {
		matchPct = float64(stats.Matches) / float64(stats.Total) * 100
	}

	sb.WriteString(fmt.Sprintf("Total: %d, Matches: %d (%.1f%%), Mismatches: %d\n",
		stats.Total,
		stats.Matches,
		matchPct,
		stats.Mismatches))

	if stats.Mismatches > 0 {
		sb.WriteString("\nMismatches:\n")
		for _, r := range stats.Results {
			if !r.Match {
				sb.WriteString(fmt.Sprintf("  %s: old=%q new=%q\n", r.Input, r.OldOutput, r.NewOutput))
			}
		}
	}

	return sb.String()
}
