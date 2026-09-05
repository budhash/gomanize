package legacy_lang

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
func (c *Comparer) CompareWithDebug(input string) (CompareResult, string) {
	oldOutput := c.oldEngine.Transliterate(input)
	newOutput, debugInfo := c.newEngine.TransliterateDebug(input, core.DefaultOptions())

	result := CompareResult{
		Input:     input,
		OldOutput: oldOutput,
		NewOutput: newOutput,
		Match:     oldOutput == newOutput,
	}

	// Format debug info
	debugStr := formatDebugInfo(debugInfo)
	return result, debugStr
}

// formatDebugInfo converts DebugInfo to a readable string.
func formatDebugInfo(info *core.DebugInfo) string {
	if info == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Parsed Units:\n")
	for _, u := range info.Units {
		meta := ""
		if u.Metadata != "" {
			meta = fmt.Sprintf(" [%s]", u.Metadata)
		}
		fmt.Fprintf(&sb, "  [%d] %s → %s (%s @ rune %d)%s\n",
			u.Index, u.Chars, u.BaseRom, u.Type, u.RunePos, meta)
	}

	if len(info.Traces) > 0 {
		sb.WriteString("\nRule Applications:\n")
		for _, t := range info.Traces {
			change := ""
			if t.Before != t.After {
				change = fmt.Sprintf(" %s→%s", t.Before, t.After)
			}
			meta := ""
			if t.Metadata != "" {
				meta = fmt.Sprintf(" [%s]", t.Metadata)
			}
			fmt.Fprintf(&sb, "  %s: %s on %s (idx %d)%s%s\n",
				t.Phase, t.Rule, t.Unit, t.UnitIdx, change, meta)
		}
	}

	return sb.String()
}

// FormatResult formats a CompareResult for display.
func FormatResult(r CompareResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Input:  %s\n", r.Input)
	fmt.Fprintf(&sb, "Old:    %s\n", r.OldOutput)
	fmt.Fprintf(&sb, "New:    %s\n", r.NewOutput)
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

	fmt.Fprintf(&sb, "Total: %d, Matches: %d (%.1f%%), Mismatches: %d\n",
		stats.Total,
		stats.Matches,
		matchPct,
		stats.Mismatches)

	if stats.Mismatches > 0 {
		sb.WriteString("\nMismatches:\n")
		for _, r := range stats.Results {
			if !r.Match {
				fmt.Fprintf(&sb, "  %s: old=%q new=%q\n", r.Input, r.OldOutput, r.NewOutput)
			}
		}
	}

	return sb.String()
}
