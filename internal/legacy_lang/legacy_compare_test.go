package legacy_lang

import (
	"testing"
)

func TestComparerBasic(t *testing.T) {
	c := NewComparer()

	tests := []struct {
		input string
	}{
		{"क"},      // single consonant
		{"का"},     // consonant + matra
		{"काम"},    // common word
		{"नमस्ते"}, // namaste
		{"भारत"},   // bharat
		{"ज्ञान"},  // special conjunct
		{"१२३"},    // numbers
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := c.Compare(tt.input)
			t.Logf("Input: %s, Old: %q, New: %q, Match: %v",
				result.Input, result.OldOutput, result.NewOutput, result.Match)
		})
	}
}

func TestComparerBatch(t *testing.T) {
	c := NewComparer()

	inputs := []string{
		"नमस्ते",
		"भारत",
		"हिंदी",
		"राम",
		"काम",
		"ज्ञान",
		"प्रकाश",
	}

	stats := c.BatchCompare(inputs)
	t.Logf("Batch results:\n%s", FormatBatchStats(stats))
}
