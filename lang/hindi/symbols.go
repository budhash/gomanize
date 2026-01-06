// Package hindi provides Hindi-specific symbol mappings and rules.
package hindi

import "github.com/budhash/gomanize/engine"

// Symbols is the complete Devanagari symbol map for Hindi.
// Maps Devanagari characters to their romanization info.
var Symbols = engine.SymbolMap{
	// Chandrabindu, Anusvara, Visarga
	"ँ": {Category: engine.CatChandrabindu, BaseRom: "n"}, // U+0901
	"ं": {Category: engine.CatAnusvara, BaseRom: "n"},     // U+0902
	"ः": {Category: engine.CatVisarga, BaseRom: "h"},      // U+0903

	// Independent Vowels
	"ऄ": {Category: engine.CatVowel, BaseRom: "a"},  // U+0904
	"अ": {Category: engine.CatVowel, BaseRom: "a"},  // U+0905
	"आ": {Category: engine.CatVowel, BaseRom: "aa"}, // U+0906
	"इ": {Category: engine.CatVowel, BaseRom: "i"},  // U+0907
	"ई": {Category: engine.CatVowel, BaseRom: "i"},  // U+0908
	"उ": {Category: engine.CatVowel, BaseRom: "u"},  // U+0909
	"ऊ": {Category: engine.CatVowel, BaseRom: "u"},  // U+090A
	"ऋ": {Category: engine.CatVowel, BaseRom: "ri"}, // U+090B
	"ऌ": {Category: engine.CatVowel, BaseRom: "l"},  // U+090C
	"ऍ": {Category: engine.CatVowel, BaseRom: "æ"},  // U+090D
	"ऎ": {Category: engine.CatVowel, BaseRom: "e"},  // U+090E
	"ए": {Category: engine.CatVowel, BaseRom: "e"},  // U+090F
	"ऐ": {Category: engine.CatVowel, BaseRom: "ai"}, // U+0910
	"ऑ": {Category: engine.CatVowel, BaseRom: "o"},  // U+0911
	"ऒ": {Category: engine.CatVowel, BaseRom: "o"},  // U+0912
	"ओ": {Category: engine.CatVowel, BaseRom: "o"},  // U+0913
	"औ": {Category: engine.CatVowel, BaseRom: "au"}, // U+0914

	// Consonants
	"क": {Category: engine.CatConsonant, BaseRom: "k"},   // U+0915
	"ख": {Category: engine.CatConsonant, BaseRom: "kh"},  // U+0916
	"ग": {Category: engine.CatConsonant, BaseRom: "g"},   // U+0917
	"घ": {Category: engine.CatConsonant, BaseRom: "gh"},  // U+0918
	"ङ": {Category: engine.CatConsonant, BaseRom: "n"},   // U+0919
	"च": {Category: engine.CatConsonant, BaseRom: "ch"},  // U+091A
	"छ": {Category: engine.CatConsonant, BaseRom: "chh"}, // U+091B
	"ज": {Category: engine.CatConsonant, BaseRom: "j"},   // U+091C
	"झ": {Category: engine.CatConsonant, BaseRom: "jh"},  // U+091D
	"ञ": {Category: engine.CatConsonant, BaseRom: "ny"},  // U+091E
	"ट": {Category: engine.CatConsonant, BaseRom: "t"},   // U+091F
	"ठ": {Category: engine.CatConsonant, BaseRom: "th"},  // U+0920
	"ड": {Category: engine.CatConsonant, BaseRom: "d"},   // U+0921
	"ढ": {Category: engine.CatConsonant, BaseRom: "dh"},  // U+0922
	"ण": {Category: engine.CatConsonant, BaseRom: "n"},   // U+0923
	"त": {Category: engine.CatConsonant, BaseRom: "t"},   // U+0924
	"थ": {Category: engine.CatConsonant, BaseRom: "th"},  // U+0925
	"द": {Category: engine.CatConsonant, BaseRom: "d"},   // U+0926
	"ध": {Category: engine.CatConsonant, BaseRom: "dh"},  // U+0927
	"न": {Category: engine.CatConsonant, BaseRom: "n"},   // U+0928
	"ऩ": {Category: engine.CatConsonant, BaseRom: "nh"},  // U+0929 (nuqta)
	"प": {Category: engine.CatConsonant, BaseRom: "p"},   // U+092A
	"फ": {Category: engine.CatConsonant, BaseRom: "ph"},  // U+092B
	"ब": {Category: engine.CatConsonant, BaseRom: "b"},   // U+092C
	"भ": {Category: engine.CatConsonant, BaseRom: "bh"},  // U+092D
	"म": {Category: engine.CatConsonant, BaseRom: "m"},   // U+092E
	"य": {Category: engine.CatConsonant, BaseRom: "y"},   // U+092F
	"र": {Category: engine.CatConsonant, BaseRom: "r"},   // U+0930
	"ऱ": {Category: engine.CatConsonant, BaseRom: "rh"},  // U+0931 (nuqta)
	"ल": {Category: engine.CatConsonant, BaseRom: "l"},   // U+0932
	"ळ": {Category: engine.CatConsonant, BaseRom: "l"},   // U+0933
	"ऴ": {Category: engine.CatConsonant, BaseRom: "lh"},  // U+0934 (nuqta)
	"व": {Category: engine.CatConsonant, BaseRom: "v"},   // U+0935
	"श": {Category: engine.CatConsonant, BaseRom: "sh"},  // U+0936
	"ष": {Category: engine.CatConsonant, BaseRom: "sh"},  // U+0937
	"स": {Category: engine.CatConsonant, BaseRom: "s"},   // U+0938
	"ह": {Category: engine.CatConsonant, BaseRom: "h"},   // U+0939

	// Nukta and other marks
	"़": {Category: engine.CatNukta, BaseRom: ""},  // U+093C
	"ऽ": {Category: engine.CatSymbol, BaseRom: ""}, // U+093D (avagraha)

	// Dependent Vowel Signs (Matras)
	"ा": {Category: engine.CatMatra, BaseRom: "a"},  // U+093E (aa matra, but colloquial uses 'a')
	"ि": {Category: engine.CatMatra, BaseRom: "i"},  // U+093F
	"ी": {Category: engine.CatMatra, BaseRom: "i"},  // U+0940
	"ु": {Category: engine.CatMatra, BaseRom: "u"},  // U+0941
	"ू": {Category: engine.CatMatra, BaseRom: "u"},  // U+0942
	"ृ": {Category: engine.CatMatra, BaseRom: "ri"}, // U+0943
	"ॄ": {Category: engine.CatMatra, BaseRom: "ri"}, // U+0944
	"ॅ": {Category: engine.CatMatra, BaseRom: "æ"},  // U+0945
	"ॆ": {Category: engine.CatMatra, BaseRom: "e"},  // U+0946
	"े": {Category: engine.CatMatra, BaseRom: "e"},  // U+0947
	"ै": {Category: engine.CatMatra, BaseRom: "ai"}, // U+0948
	"ॉ": {Category: engine.CatMatra, BaseRom: "o"},  // U+0949
	"ॊ": {Category: engine.CatMatra, BaseRom: "o"},  // U+094A
	"ो": {Category: engine.CatMatra, BaseRom: "o"},  // U+094B
	"ौ": {Category: engine.CatMatra, BaseRom: "au"}, // U+094C

	// Halant (Virama)
	"्": {Category: engine.CatHalant, BaseRom: ""}, // U+094D

	// Om
	"ॐ": {Category: engine.CatConsonant, BaseRom: "om"}, // U+0950

	// Consonants with Nukta (Urdu/Persian sounds)
	"क़": {Category: engine.CatConsonant, BaseRom: "q"},  // U+0958
	"ख़": {Category: engine.CatConsonant, BaseRom: "kh"}, // U+0959
	"ग़": {Category: engine.CatConsonant, BaseRom: "gh"}, // U+095A
	"ज़": {Category: engine.CatConsonant, BaseRom: "z"},  // U+095B
	"ड़": {Category: engine.CatConsonant, BaseRom: "r"},  // U+095C
	"ढ़": {Category: engine.CatConsonant, BaseRom: "rh"}, // U+095D
	"फ़": {Category: engine.CatConsonant, BaseRom: "f"},  // U+095E
	"य़": {Category: engine.CatConsonant, BaseRom: "yh"}, // U+095F

	// Additional vowels
	"ॠ": {Category: engine.CatVowel, BaseRom: "ri"}, // U+0960
	"ॡ": {Category: engine.CatVowel, BaseRom: "ll"}, // U+0961
	"ॢ": {Category: engine.CatMatra, BaseRom: "l"},  // U+0962
	"ॣ": {Category: engine.CatMatra, BaseRom: "ll"}, // U+0963

	// Devanagari Numerals
	"०": {Category: engine.CatNumber, BaseRom: "0"}, // U+0966
	"१": {Category: engine.CatNumber, BaseRom: "1"}, // U+0967
	"२": {Category: engine.CatNumber, BaseRom: "2"}, // U+0968
	"३": {Category: engine.CatNumber, BaseRom: "3"}, // U+0969
	"४": {Category: engine.CatNumber, BaseRom: "4"}, // U+096A
	"५": {Category: engine.CatNumber, BaseRom: "5"}, // U+096B
	"६": {Category: engine.CatNumber, BaseRom: "6"}, // U+096C
	"७": {Category: engine.CatNumber, BaseRom: "7"}, // U+096D
	"८": {Category: engine.CatNumber, BaseRom: "8"}, // U+096E
	"९": {Category: engine.CatNumber, BaseRom: "9"}, // U+096F

	// Special Conjuncts (multi-character)
	// Only ज्ञ needs special handling; others work component-wise
	"ज्ञ": {Category: engine.CatConjunct, BaseRom: "gy"}, // ज + ् + ञ → "gy" (not "jny")
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

// Hindi implements the engine.Language interface.
type Hindi struct{}

// Name returns the language identifier.
func (h Hindi) Name() string {
	return "hindi"
}

// Symbols returns the Hindi symbol map.
func (h Hindi) Symbols() engine.SymbolMap {
	return Symbols
}

// MultiChar returns multi-character sequences for Hindi.
func (h Hindi) MultiChar() []string {
	return MultiChar
}

// Halant returns the Hindi halant character.
func (h Hindi) Halant() string {
	return Halant
}
