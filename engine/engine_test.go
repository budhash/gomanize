package engine_test

import (
	"testing"

	"github.com/budhash/gomanize/engine"
	"github.com/budhash/gomanize/lang/hindi"
)

func TestParseSimpleWord(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	tests := []struct {
		name     string
		input    string
		wantLen  int // number of units
		wantRuns int // number of consonant runs
	}{
		{
			name:     "single consonant",
			input:    "क",
			wantLen:  1,
			wantRuns: 1,
		},
		{
			name:     "consonant with matra",
			input:    "का",
			wantLen:  2, // क + ा
			wantRuns: 1,
		},
		{
			name:     "two consonants",
			input:    "कम",
			wantLen:  2, // क + म
			wantRuns: 1, // both in same run
		},
		{
			name:     "namaste",
			input:    "नमस्ते",
			wantLen:  5, // न + म + स + त (after halant) + े
			wantRuns: 1,
		},
		{
			name:     "vowel then consonant",
			input:    "अब",
			wantLen:  2, // अ + ब
			wantRuns: 1, // ब starts new run after vowel
		},
		{
			name:     "special conjunct",
			input:    "ज्ञान",
			wantLen:  3, // ज्ञ (conjunct) + ा + न
			wantRuns: 2, // ज्ञ run, then न run after ा vowel
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := eng.ParseOnly(tt.input)
			if len(word.Units) != tt.wantLen {
				t.Errorf("ParseOnly(%q) got %d units, want %d", tt.input, len(word.Units), tt.wantLen)
				for i, u := range word.Units {
					t.Logf("  unit[%d]: %s (%s)", i, string(u.Runes), u.Type)
				}
			}
			if len(word.Runs) != tt.wantRuns {
				t.Errorf("ParseOnly(%q) got %d runs, want %d", tt.input, len(word.Runs), tt.wantRuns)
			}
		})
	}
}

func TestUnitLinks(t *testing.T) {
	eng := engine.New(hindi.Hindi{})
	word := eng.ParseOnly("काम") // क + ा + म

	if len(word.Units) != 3 {
		t.Fatalf("expected 3 units, got %d", len(word.Units))
	}

	// Check bidirectional links
	ka := word.Units[0] // क
	aa := word.Units[1] // ा
	ma := word.Units[2] // म

	if ka.Prev != nil {
		t.Error("first unit should have nil Prev")
	}
	if ka.Next != aa {
		t.Error("first unit's Next should be second unit")
	}
	if aa.Prev != ka {
		t.Error("second unit's Prev should be first unit")
	}
	if aa.Next != ma {
		t.Error("second unit's Next should be third unit")
	}
	if ma.Prev != aa {
		t.Error("third unit's Prev should be second unit")
	}
	if ma.Next != nil {
		t.Error("last unit should have nil Next")
	}
}

func TestConsonantRunIdentification(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	// कमला: क + म + ल + ा (three consonants, then matra)
	// Note: ा is a dependent vowel (matra), so all 3 consonants are in one run
	word := eng.ParseOnly("कमला")

	// All consonants before the final matra are in one run
	if len(word.Runs) != 1 {
		t.Fatalf("expected 1 run (consonants before matra), got %d", len(word.Runs))
	}

	run0 := word.Runs[0]
	if len(run0.Units) != 3 {
		t.Errorf("run should have 3 consonants (क, म, ल), got %d", len(run0.Units))
	}
	if run0.PrevVowel != nil {
		t.Error("run should have nil PrevVowel (word-initial)")
	}
	if run0.NextVowel == nil {
		t.Error("run should have NextVowel (the ा matra)")
	}
}

func TestAfterHalantTracking(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	// स्त = स + ् + त (त comes after halant)
	word := eng.ParseOnly("स्त")

	// Parser should create: स, त (halant is consumed, not a unit)
	// त should have AfterHalant = true
	if len(word.Units) != 2 {
		t.Fatalf("expected 2 units for स्त, got %d", len(word.Units))
	}

	sa := word.Units[0] // स
	ta := word.Units[1] // त

	if sa.AfterHalant {
		t.Error("स should not have AfterHalant")
	}
	if !ta.AfterHalant {
		t.Error("त should have AfterHalant (it follows ्)")
	}
}

func TestSpecialConjunct(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	// ज्ञ should be parsed as a single unit
	word := eng.ParseOnly("ज्ञ")

	if len(word.Units) != 1 {
		t.Fatalf("expected 1 unit for ज्ञ, got %d", len(word.Units))
	}

	unit := word.Units[0]
	if unit.Type != engine.UnitConjunct {
		t.Errorf("ज्ञ should be UnitConjunct, got %s", unit.Type)
	}
	if unit.BaseRom != "gy" {
		t.Errorf("ज्ञ should have BaseRom 'gy', got %q", unit.BaseRom)
	}
}

func TestBasicRender(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	tests := []struct {
		input string
		want  string
	}{
		// Simple consonants (word-final schwa deleted)
		{"क", "k"},
		{"म", "m"},

		// Consonant + matra
		{"का", "ka"}, // क + ा → k + a (matra replaces inherent)
		{"की", "ki"},
		{"कु", "ku"},
		{"के", "ke"},
		{"को", "ko"},

		// Numbers
		{"१२३", "123"},

		// Independent vowels
		{"अ", "a"},
		{"आ", "aa"},
		{"इ", "i"},

		// Conjuncts (word-final schwa deleted)
		{"ज्ञ", "gy"}, // ज्ञ as unit, word-final schwa deleted
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := eng.Transliterate(tt.input)
			if got != tt.want {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunHelpers(t *testing.T) {
	eng := engine.New(hindi.Hindi{})
	word := eng.ParseOnly("कमला") // क + म + ल + ा

	// Find the क, म, and ल consonants
	var ka, ma, la *engine.Unit
	for _, u := range word.Units {
		switch string(u.Runes) {
		case "क":
			ka = u
		case "म":
			ma = u
		case "ल":
			la = u
		}
	}

	if ka == nil || ma == nil || la == nil {
		t.Fatal("could not find क, म, or ल units")
	}

	// Test run position helpers
	if !ka.IsRunInitial() {
		t.Error("क should be run-initial")
	}
	if ka.IsRunFinal() {
		t.Error("क should not be run-final")
	}

	if ma.IsRunInitial() {
		t.Error("म should not be run-initial")
	}
	if ma.IsRunFinal() {
		t.Error("म should not be run-final")
	}

	if la.IsRunInitial() {
		t.Error("ल should not be run-initial")
	}
	if !la.IsRunFinal() {
		t.Error("ल should be run-final")
	}

	// Test navigation
	if ka.NextInRun() != ma {
		t.Error("क.NextInRun() should be म")
	}
	if ma.PrevInRun() != ka {
		t.Error("म.PrevInRun() should be क")
	}
	if ma.NextInRun() != la {
		t.Error("म.NextInRun() should be ल")
	}
	if la.PrevInRun() != ma {
		t.Error("ल.PrevInRun() should be म")
	}
	if ka.PrevInRun() != nil {
		t.Error("क.PrevInRun() should be nil")
	}
	if la.NextInRun() != nil {
		t.Error("ल.NextInRun() should be nil")
	}
}

func TestLongVowelsOption(t *testing.T) {
	eng := engine.New(hindi.Hindi{})

	tests := []struct {
		input      string
		wantNormal string
		wantLong   string
	}{
		// Words with medial aa-matra: default uses "a", --long-vowels uses "aa" everywhere
		{"गाना", "gana", "gaanaa"},      // ga-naa vs gaa-naa
		{"बनाना", "banana", "banaanaa"}, // ba-na-naa vs ba-naa-naa
		{"खाना", "khana", "khaanaa"},    // kha-naa vs khaa-naa
		{"जाना", "jana", "jaanaa"},      // ja-naa vs jaa-naa

		// Words with final aa-matra in closed syllable: both use "aa"
		{"काम", "kaam", "kaam"},
		{"राम", "raam", "raam"},

		// Words without aa-matra: no difference
		{"नमस्ते", "namaste", "namaste"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Test with default options (no long vowels)
			gotNormal := eng.Transliterate(tt.input)
			if gotNormal != tt.wantNormal {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, gotNormal, tt.wantNormal)
			}

			// Test with long vowels option
			opts := engine.Options{LongVowels: true}
			gotLong := eng.TransliterateWithOptions(tt.input, opts)
			if gotLong != tt.wantLong {
				t.Errorf("TransliterateWithOptions(%q, LongVowels=true) = %q, want %q", tt.input, gotLong, tt.wantLong)
			}
		})
	}
}
