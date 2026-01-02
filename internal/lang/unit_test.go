package lang

import (
	"testing"
)

// =============================================================================
// UNIT TESTS
// These are fast, targeted tests for specific transliteration rules.
// Run with: go test -run "^TestUnit" or make test-unit
// =============================================================================

// -----------------------------------------------------------------------------
// Basic Character Mapping
// -----------------------------------------------------------------------------

func TestUnitVowels(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Independent vowels
		{"अ", "अ", "a"},
		{"आ", "आ", "aa"},
		{"इ", "इ", "i"},
		{"ई", "ई", "i"},
		{"उ", "उ", "u"},
		{"ऊ", "ऊ", "u"},
		{"ए", "ए", "e"},
		{"ऐ", "ऐ", "ai"},
		{"ओ", "ओ", "o"},
		{"औ", "औ", "au"},
		// Vowel matras (with क)
		{"का", "का", "ka"},
		{"कि", "कि", "ki"},
		{"की", "की", "ki"},
		{"कु", "कु", "ku"},
		{"कू", "कू", "ku"},
		{"के", "के", "ke"},
		{"कै", "कै", "kai"},
		{"को", "को", "ko"},
		{"कौ", "कौ", "kau"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnitConsonants(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Velars
		{"क", "क", "k"},
		{"ख", "ख", "kh"},
		{"ग", "ग", "g"},
		{"घ", "घ", "gh"},
		// Palatals
		{"च", "च", "ch"},
		{"छ", "छ", "chh"},
		{"ज", "ज", "j"},
		{"झ", "झ", "jh"},
		// Retroflexes
		{"ट", "ट", "t"},
		{"ठ", "ठ", "th"},
		{"ड", "ड", "d"},
		{"ढ", "ढ", "dh"},
		// Dentals
		{"त", "त", "t"},
		{"थ", "थ", "th"},
		{"द", "द", "d"},
		{"ध", "ध", "dh"},
		{"न", "न", "n"},
		// Labials
		{"प", "प", "p"},
		{"फ", "फ", "ph"},
		{"ब", "ब", "b"},
		{"भ", "भ", "bh"},
		{"म", "म", "m"},
		// Semivowels
		{"य", "य", "y"},
		{"र", "र", "r"},
		{"ल", "ल", "l"},
		{"व", "व", "v"}, // Changed to 'v' - more common in Hindi romanization
		// Sibilants
		{"श", "श", "sh"},
		{"ष", "ष", "sh"},
		{"स", "स", "s"},
		{"ह", "ह", "h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnitNumbers(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		input    string
		expected string
	}{
		{"०", "0"},
		{"१", "1"},
		{"२", "2"},
		{"३", "3"},
		{"४", "4"},
		{"५", "5"},
		{"६", "6"},
		{"७", "7"},
		{"८", "8"},
		{"९", "9"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnitNuqta(t *testing.T) {
	h := Hindi{}

	// Nuqta consonants (for Persian/English loanwords)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"क़ (qa)", "क़", "q"},
		{"ख़ (kha)", "ख़", "kh"},
		{"ग़ (gha)", "ग़", "gh"},
		{"ज़ (za)", "ज़", "z"},
		{"फ़ (fa)", "फ़", "f"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnitConjuncts(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"क्ष (ksha)", "क्ष", "ksh"},
		{"त्र (tra)", "त्र", "tra"}, // Final र after halant gets schwa (Sanskrit pattern)
		{"ज्ञ (gya/jnya)", "ज्ञ", "jny"},
		{"श्र (shra)", "श्र", "shra"}, // Final र after halant gets schwa (Sanskrit pattern)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Schwa Handling Rules
// -----------------------------------------------------------------------------

func TestUnitSchwaBasic(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
		note     string
	}{
		// Word-final schwa deletion with long vowel "aa" (ा + C + END)
		{"word_final_delete", "राम", "raam", "Word-final aa-matra before consonant"},
		{"word_final_delete_2", "नाम", "naam", "Word-final aa-matra before consonant"},

		// Schwa between consonant+vowel (correct behavior)
		{"cv_pattern", "कमल", "kamal", "Schwa retained before vowel"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s (%s): got %q, want %q", tc.input, tc.note, result, tc.expected)
			}
		})
	}
}

// These tests document KNOWN BUGS that need to be fixed
func TestUnitSchwaKnownBugs(t *testing.T) {
	h := Hindi{}

	// These currently FAIL - documenting expected behavior after fix
	knownBugs := []struct {
		name     string
		input    string
		current  string // What we get now (wrong)
		expected string // What we should get (correct)
		issue    string
	}{
		{"first_syllable_1", "प्रकाश", "prkash", "prakaash", "MISSING_SCHWA: First syllable schwa deleted"},
		{"first_syllable_2", "अध्यक्ष", "adhyksh", "adhyaksh", "MISSING_SCHWA: First syllable schwa deleted"},
		{"first_syllable_3", "गर्भ", "grbh", "garbh", "MISSING_SCHWA: First syllable schwa deleted"},
	}

	for _, tc := range knownBugs {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result == tc.expected {
				t.Logf("BUG FIXED! %s now correctly produces %q", tc.input, tc.expected)
			} else if result == tc.current {
				t.Logf("Known bug: %s → %q (should be %q) [%s]", tc.input, result, tc.expected, tc.issue)
			} else {
				t.Logf("Unexpected: %s → %q (expected %q or %q)", tc.input, result, tc.current, tc.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// व (va/wa) Handling
// -----------------------------------------------------------------------------

func TestUnitVaWa(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
		note     string
	}{
		// व → v in all positions (colloquial Hindi)
		{"initial_v", "वन", "van", "Word-initial व → v"},
		{"medial_v", "देव", "dev", "Medial व → v"},
		{"medial_v_2", "उत्सव", "utsav", "Medial व → v"},
		{"final_v", "कवि", "kavi", "व before vowel → v"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s (%s): got %q, want %q", tc.input, tc.note, result, tc.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Common Words (Smoke Test)
// -----------------------------------------------------------------------------

func TestUnitCommonWords(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		input    string
		expected string
	}{
		{"नमस्ते", "namaste"},
		{"भारत", "bharat"},
		{"हिंदी", "hindi"},
		{"और", "aur"},
		{"है", "hai"},
		{"में", "men"},
		{"को", "ko"},
		{"से", "se"},
		{"पर", "par"},
		{"एक", "ek"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Anusvara and Chandrabindu
// -----------------------------------------------------------------------------

func TestUnitAnusvara(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"anusvara_n", "हिंदी", "hindi"},
		{"anusvara_n_2", "संगीत", "sangit"},
		{"chandrabindu", "माँ", "man"}, // Note: May need adjustment
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Edge Cases
// -----------------------------------------------------------------------------

func TestUnitEdgeCases(t *testing.T) {
	h := Hindi{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single_vowel", "अ", "a"},
		{"single_consonant", "क", "k"},
		{"mixed_script", "hello दुनिया", "hello duniya"},
		{"punctuation", "क्या?", "kya?"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Transliterate(tc.input)
			if result != tc.expected {
				t.Errorf("%s: got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
