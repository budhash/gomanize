// Package hindi provides Hindi-specific symbol mappings and rules using the core architecture.
package hindi

import (
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// Symbols is the complete Devanagari symbol map for Hindi.
// Maps Devanagari characters to their romanization info.
var Symbols = core.SymbolMap{
	// Chandrabindu, Anusvara, Visarga
	"ँ": {Category: brahmic.CatChandrabindu, BaseRom: "n"}, // U+0901
	"ं": {Category: brahmic.CatAnusvara, BaseRom: "n"},     // U+0902
	"ः": {Category: brahmic.CatVisarga, BaseRom: "h"},      // U+0903

	// Independent Vowels
	"ऄ": {Category: core.CatVowel, BaseRom: "a"},  // U+0904
	"अ": {Category: core.CatVowel, BaseRom: "a"},  // U+0905
	"आ": {Category: core.CatVowel, BaseRom: "aa"}, // U+0906
	"इ": {Category: core.CatVowel, BaseRom: "i"},  // U+0907
	"ई": {Category: core.CatVowel, BaseRom: "i"},  // U+0908
	"उ": {Category: core.CatVowel, BaseRom: "u"},  // U+0909
	"ऊ": {Category: core.CatVowel, BaseRom: "u"},  // U+090A
	"ऋ": {Category: core.CatVowel, BaseRom: "ri"}, // U+090B
	"ऌ": {Category: core.CatVowel, BaseRom: "l"},  // U+090C
	"ऍ": {Category: core.CatVowel, BaseRom: "æ"},  // U+090D
	"ऎ": {Category: core.CatVowel, BaseRom: "e"},  // U+090E
	"ए": {Category: core.CatVowel, BaseRom: "e"},  // U+090F
	"ऐ": {Category: core.CatVowel, BaseRom: "ai"}, // U+0910
	"ऑ": {Category: core.CatVowel, BaseRom: "o"},  // U+0911
	"ऒ": {Category: core.CatVowel, BaseRom: "o"},  // U+0912
	"ओ": {Category: core.CatVowel, BaseRom: "o"},  // U+0913
	"औ": {Category: core.CatVowel, BaseRom: "au"}, // U+0914

	// Consonants
	"क": {Category: core.CatConsonant, BaseRom: "k"},   // U+0915
	"ख": {Category: core.CatConsonant, BaseRom: "kh"},  // U+0916
	"ग": {Category: core.CatConsonant, BaseRom: "g"},   // U+0917
	"घ": {Category: core.CatConsonant, BaseRom: "gh"},  // U+0918
	"ङ": {Category: core.CatConsonant, BaseRom: "n"},   // U+0919
	"च": {Category: core.CatConsonant, BaseRom: "ch"},  // U+091A
	"छ": {Category: core.CatConsonant, BaseRom: "chh"}, // U+091B
	"ज": {Category: core.CatConsonant, BaseRom: "j"},   // U+091C
	"झ": {Category: core.CatConsonant, BaseRom: "jh"},  // U+091D
	"ञ": {Category: core.CatConsonant, BaseRom: "ny"},  // U+091E
	"ट": {Category: core.CatConsonant, BaseRom: "t"},   // U+091F
	"ठ": {Category: core.CatConsonant, BaseRom: "th"},  // U+0920
	"ड": {Category: core.CatConsonant, BaseRom: "d"},   // U+0921
	"ढ": {Category: core.CatConsonant, BaseRom: "dh"},  // U+0922
	"ण": {Category: core.CatConsonant, BaseRom: "n"},   // U+0923
	"त": {Category: core.CatConsonant, BaseRom: "t"},   // U+0924
	"थ": {Category: core.CatConsonant, BaseRom: "th"},  // U+0925
	"द": {Category: core.CatConsonant, BaseRom: "d"},   // U+0926
	"ध": {Category: core.CatConsonant, BaseRom: "dh"},  // U+0927
	"न": {Category: core.CatConsonant, BaseRom: "n"},   // U+0928
	"ऩ": {Category: core.CatConsonant, BaseRom: "nh"},  // U+0929 (nuqta)
	"प": {Category: core.CatConsonant, BaseRom: "p"},   // U+092A
	"फ": {Category: core.CatConsonant, BaseRom: "ph"},  // U+092B
	"ब": {Category: core.CatConsonant, BaseRom: "b"},   // U+092C
	"भ": {Category: core.CatConsonant, BaseRom: "bh"},  // U+092D
	"म": {Category: core.CatConsonant, BaseRom: "m"},   // U+092E
	"य": {Category: core.CatConsonant, BaseRom: "y"},   // U+092F
	"र": {Category: core.CatConsonant, BaseRom: "r"},   // U+0930
	"ऱ": {Category: core.CatConsonant, BaseRom: "rh"},  // U+0931 (nuqta)
	"ल": {Category: core.CatConsonant, BaseRom: "l"},   // U+0932
	"ळ": {Category: core.CatConsonant, BaseRom: "l"},   // U+0933
	"ऴ": {Category: core.CatConsonant, BaseRom: "lh"},  // U+0934 (nuqta)
	"व": {Category: core.CatConsonant, BaseRom: "v"},   // U+0935
	"श": {Category: core.CatConsonant, BaseRom: "sh"},  // U+0936
	"ष": {Category: core.CatConsonant, BaseRom: "sh"},  // U+0937
	"स": {Category: core.CatConsonant, BaseRom: "s"},   // U+0938
	"ह": {Category: core.CatConsonant, BaseRom: "h"},   // U+0939

	// Nukta and other marks
	"़": {Category: brahmic.CatNukta, BaseRom: ""}, // U+093C
	"ऽ": {Category: core.CatSymbol, BaseRom: ""},   // U+093D (avagraha)

	// Dependent Vowel Signs (Matras)
	"ा": {Category: brahmic.CatMatra, BaseRom: "a"},  // U+093E (aa matra, but colloquial uses 'a')
	"ि": {Category: brahmic.CatMatra, BaseRom: "i"},  // U+093F
	"ी": {Category: brahmic.CatMatra, BaseRom: "i"},  // U+0940
	"ु": {Category: brahmic.CatMatra, BaseRom: "u"},  // U+0941
	"ू": {Category: brahmic.CatMatra, BaseRom: "u"},  // U+0942
	"ृ": {Category: brahmic.CatMatra, BaseRom: "ri"}, // U+0943
	"ॄ": {Category: brahmic.CatMatra, BaseRom: "ri"}, // U+0944
	"ॅ": {Category: brahmic.CatMatra, BaseRom: "æ"},  // U+0945
	"ॆ": {Category: brahmic.CatMatra, BaseRom: "e"},  // U+0946
	"े": {Category: brahmic.CatMatra, BaseRom: "e"},  // U+0947
	"ै": {Category: brahmic.CatMatra, BaseRom: "ai"}, // U+0948
	"ॉ": {Category: brahmic.CatMatra, BaseRom: "o"},  // U+0949
	"ॊ": {Category: brahmic.CatMatra, BaseRom: "o"},  // U+094A
	"ो": {Category: brahmic.CatMatra, BaseRom: "o"},  // U+094B
	"ौ": {Category: brahmic.CatMatra, BaseRom: "au"}, // U+094C

	// Halant (Virama)
	"्": {Category: brahmic.CatHalant, BaseRom: ""}, // U+094D

	// Om
	"ॐ": {Category: core.CatConsonant, BaseRom: "om"}, // U+0950

	// Consonants with Nukta (Urdu/Persian sounds)
	"क़": {Category: core.CatConsonant, BaseRom: "q"},  // U+0958
	"ख़": {Category: core.CatConsonant, BaseRom: "kh"}, // U+0959
	"ग़": {Category: core.CatConsonant, BaseRom: "gh"}, // U+095A
	"ज़": {Category: core.CatConsonant, BaseRom: "z"},  // U+095B
	"ड़": {Category: core.CatConsonant, BaseRom: "r"},  // U+095C
	"ढ़": {Category: core.CatConsonant, BaseRom: "rh"}, // U+095D
	"फ़": {Category: core.CatConsonant, BaseRom: "f"},  // U+095E
	"य़": {Category: core.CatConsonant, BaseRom: "yh"}, // U+095F

	// Additional vowels
	"ॠ": {Category: core.CatVowel, BaseRom: "ri"},    // U+0960
	"ॡ": {Category: core.CatVowel, BaseRom: "ll"},    // U+0961
	"ॢ": {Category: brahmic.CatMatra, BaseRom: "l"},  // U+0962
	"ॣ": {Category: brahmic.CatMatra, BaseRom: "ll"}, // U+0963

	// Devanagari Numerals
	"०": {Category: core.CatNumber, BaseRom: "0"}, // U+0966
	"१": {Category: core.CatNumber, BaseRom: "1"}, // U+0967
	"२": {Category: core.CatNumber, BaseRom: "2"}, // U+0968
	"३": {Category: core.CatNumber, BaseRom: "3"}, // U+0969
	"४": {Category: core.CatNumber, BaseRom: "4"}, // U+096A
	"५": {Category: core.CatNumber, BaseRom: "5"}, // U+096B
	"६": {Category: core.CatNumber, BaseRom: "6"}, // U+096C
	"७": {Category: core.CatNumber, BaseRom: "7"}, // U+096D
	"८": {Category: core.CatNumber, BaseRom: "8"}, // U+096E
	"९": {Category: core.CatNumber, BaseRom: "9"}, // U+096F

	// Special Conjuncts (multi-character)
	// Only ज्ञ needs special handling; others work component-wise
	"ज्ञ": {Category: brahmic.CatConjunct, BaseRom: "gy"}, // ज + ् + ञ → "gy" (not "jny")
}

// MultiChar lists multi-character sequences to try during parsing.
// These are checked before single characters, longest first.
var MultiChar = []string{
	"ज्ञ", // ज + ् + ञ (must be parsed as single unit)
	// Nukta combinations (base + nukta)
	"क़", "ख़", "ग़", "ज़", "ड़", "ढ़", "फ़", "य़",
}

// Halant is the virama character that suppresses inherent vowel.
const Halant = "्"

// Nukta is the diacritic that modifies consonants.
const Nukta = "़"

// brahmicScript is the shared Brahmic script instance.
var brahmicScript = brahmic.New()

// Hindi implements the core.Language interface.
type Hindi struct{}

// Name returns the language identifier.
func (h Hindi) Name() string {
	return "hindi"
}

// Script returns the script used by this language.
func (h Hindi) Script() core.Script {
	return brahmicScript
}

// Symbols returns the Hindi symbol map.
func (h Hindi) Symbols() core.SymbolMap {
	return Symbols
}

// ScriptConfig returns Brahmic-specific configuration.
func (h Hindi) ScriptConfig() interface{} {
	return brahmic.Config{
		Halant:    Halant,
		Nukta:     Nukta,
		MultiChar: MultiChar,
	}
}

// Rules returns the complete rule catalog for Hindi.
// Schemes select which rules to use from this catalog.
func (h Hindi) Rules() core.RuleCatalog {
	return RuleCatalog()
}
