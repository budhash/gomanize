package hindi2

import (
	"testing"

	oldengine "github.com/budhash/gomanize/engine"
	oldhindi "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/script/brahmic"
)

// TestEngineParityWords tests that both engines produce identical output for common words.
func TestEngineParityWords(t *testing.T) {
	// Old engine
	oldEng := oldengine.New(oldhindi.Hindi{})

	// New engine
	newEng := brahmic.New(Hindi{})

	words := []string{
		// Basic words
		"नमस्ते",
		"भारत",
		"हिंदी",
		"नाम",
		"काम",
		"राम",

		// Schwa deletion patterns
		"जनता",
		"कमला",
		"अपना",
		"समाज",
		"सरकार",

		// Conjuncts
		"मंत्र",
		"चंद्र",
		"स्वागत",
		"ऐश्वर्या",
		"अध्यक्ष",
		"प्रकाश",

		// ज्ञ conjunct
		"ज्ञान",
		"विज्ञान",
		"अज्ञात",

		// Long vowels with final consonant
		"इंसान",
		"मकान",
		"दुकान",

		// ीय suffix
		"केंद्रीय",
		"राष्ट्रीय",
		"भारतीय",

		// Numbers
		"१२३४५६७८९०",

		// Nukta consonants
		"फ़िल्म",
		"ज़िंदगी",

		// Mixed
		"हिंदुस्तान",
		"पाकिस्तान",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			oldResult := oldEng.Transliterate(word)
			newResult := newEng.Transliterate(word)

			if oldResult != newResult {
				t.Errorf("Engine mismatch for %q: old=%q, new=%q", word, oldResult, newResult)
			}
		})
	}
}

// TestEngineParityLongVowels tests that both engines handle long vowels identically.
func TestEngineParityLongVowels(t *testing.T) {
	// Old engine
	oldEng := oldengine.New(oldhindi.Hindi{})

	// New engine
	newEng := brahmic.New(Hindi{})

	opts := oldengine.Options{LongVowels: true}
	brahmicOpts := brahmic.Options{LongVowels: true}

	words := []string{
		"गाना",
		"बनाना",
		"खाना",
		"जाना",
		"आना",
		"काम",
		"नाम",
		"राम",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			oldResult := oldEng.TransliterateWithOptions(word, opts)
			newResult := newEng.TransliterateWithOptions(word, brahmicOpts)

			if oldResult != newResult {
				t.Errorf("Engine mismatch (LongVowels=true) for %q: old=%q, new=%q",
					word, oldResult, newResult)
			}
		})
	}
}
