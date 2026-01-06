package hindi2

import (
	"testing"

	"github.com/budhash/gomanize/script/brahmic"
)

func TestHindiBasic(t *testing.T) {
	engine := brahmic.New(Hindi{})

	tests := []struct {
		input    string
		expected string
	}{
		// Basic words
		{"नमस्ते", "namaste"},
		{"भारत", "bharat"},
		{"हिंदी", "hindi"},

		// Schwa deletion at word end
		{"काम", "kaam"},
		{"नाम", "naam"},
		{"राम", "raam"},

		// Medial schwa deletion (CCV pattern)
		{"जनता", "janta"},
		{"कमला", "kamla"},
		{"अपना", "apna"},

		// Conjuncts with preserved schwa
		{"मंत्र", "mantra"},
		{"चंद्र", "chandra"},

		// व→w in conjuncts
		{"स्वागत", "swagat"},

		// ज्ञ conjunct
		{"ज्ञान", "gyaan"},

		// Numbers
		{"१२३", "123"},

		// ीय suffix
		{"केंद्रीय", "kendriya"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := engine.Transliterate(tt.input)
			if got != tt.expected {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLongVowelsOption(t *testing.T) {
	engine := brahmic.New(Hindi{})

	tests := []struct {
		input      string
		longVowels bool
		expected   string
	}{
		// Without long vowels: only ा+C+END gets aa
		{"गाना", false, "gana"},
		{"बनाना", false, "banana"},
		{"काम", false, "kaam"},

		// With long vowels: all ा gets aa
		{"गाना", true, "gaanaa"},
		{"बनाना", true, "banaanaa"},
		{"काम", true, "kaam"},
	}

	for _, tt := range tests {
		name := tt.input
		if tt.longVowels {
			name += "-long"
		}
		t.Run(name, func(t *testing.T) {
			opts := brahmic.DefaultOptions()
			opts.LongVowels = tt.longVowels
			got := engine.TransliterateWithOptions(tt.input, opts)
			if got != tt.expected {
				t.Errorf("TransliterateWithOptions(%q, LongVowels=%v) = %q, want %q",
					tt.input, tt.longVowels, got, tt.expected)
			}
		})
	}
}

func TestSymbolMap(t *testing.T) {
	h := Hindi{}
	symbols := h.Symbols()

	// Test some key entries
	if info, ok := symbols["क"]; !ok || info.BaseRom != "k" {
		t.Errorf("Expected क→k")
	}
	if info, ok := symbols["ज्ञ"]; !ok || info.BaseRom != "gy" {
		t.Errorf("Expected ज्ञ→gy")
	}
	if info, ok := symbols["्"]; !ok || info.Category != brahmic.CatHalant {
		t.Errorf("Expected ् to be CatHalant")
	}
}

func TestMultiChar(t *testing.T) {
	h := Hindi{}
	mc := h.MultiChar()

	// Should include ज्ञ and nukta combinations
	found := false
	for _, s := range mc {
		if s == "ज्ञ" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MultiChar should include ज्ञ")
	}
}

func TestLanguageInterface(t *testing.T) {
	h := Hindi{}

	if h.Name() != "hindi" {
		t.Errorf("Name() = %q, want %q", h.Name(), "hindi")
	}
	if h.Halant() != "्" {
		t.Errorf("Halant() = %q, want %q", h.Halant(), "्")
	}
	if h.Nukta() != "़" {
		t.Errorf("Nukta() = %q, want %q", h.Nukta(), "़")
	}
}
