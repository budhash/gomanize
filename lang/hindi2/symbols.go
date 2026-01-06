// Package hindi2 provides Hindi-specific symbol mappings and rules using the brahmic script layer.
package hindi2

import "github.com/budhash/gomanize/script/brahmic"

// Symbols is the complete Devanagari symbol map for Hindi.
// Maps Devanagari characters to their romanization info.
var Symbols = brahmic.SymbolMap{
	// Chandrabindu, Anusvara, Visarga
	"ँ": {Category: brahmic.CatChandrabindu, BaseRom: "n"}, // U+0901
	"ं": {Category: brahmic.CatAnusvara, BaseRom: "n"},     // U+0902
	"ः": {Category: brahmic.CatVisarga, BaseRom: "h"},      // U+0903

	// Independent Vowels
	"ऄ": {Category: brahmic.CatVowel, BaseRom: "a"},  // U+0904
	"अ": {Category: brahmic.CatVowel, BaseRom: "a"},  // U+0905
	"आ": {Category: brahmic.CatVowel, BaseRom: "aa"}, // U+0906
	"इ": {Category: brahmic.CatVowel, BaseRom: "i"},  // U+0907
	"ई": {Category: brahmic.CatVowel, BaseRom: "i"},  // U+0908
	"उ": {Category: brahmic.CatVowel, BaseRom: "u"},  // U+0909
	"ऊ": {Category: brahmic.CatVowel, BaseRom: "u"},  // U+090A
	"ऋ": {Category: brahmic.CatVowel, BaseRom: "ri"}, // U+090B
	"ऌ": {Category: brahmic.CatVowel, BaseRom: "l"},  // U+090C
	"ऍ": {Category: brahmic.CatVowel, BaseRom: "æ"},  // U+090D
	"ऎ": {Category: brahmic.CatVowel, BaseRom: "e"},  // U+090E
	"ए": {Category: brahmic.CatVowel, BaseRom: "e"},  // U+090F
	"ऐ": {Category: brahmic.CatVowel, BaseRom: "ai"}, // U+0910
	"ऑ": {Category: brahmic.CatVowel, BaseRom: "o"},  // U+0911
	"ऒ": {Category: brahmic.CatVowel, BaseRom: "o"},  // U+0912
	"ओ": {Category: brahmic.CatVowel, BaseRom: "o"},  // U+0913
	"औ": {Category: brahmic.CatVowel, BaseRom: "au"}, // U+0914

	// Consonants
	"क": {Category: brahmic.CatConsonant, BaseRom: "k"},   // U+0915
	"ख": {Category: brahmic.CatConsonant, BaseRom: "kh"},  // U+0916
	"ग": {Category: brahmic.CatConsonant, BaseRom: "g"},   // U+0917
	"घ": {Category: brahmic.CatConsonant, BaseRom: "gh"},  // U+0918
	"ङ": {Category: brahmic.CatConsonant, BaseRom: "n"},   // U+0919
	"च": {Category: brahmic.CatConsonant, BaseRom: "ch"},  // U+091A
	"छ": {Category: brahmic.CatConsonant, BaseRom: "chh"}, // U+091B
	"ज": {Category: brahmic.CatConsonant, BaseRom: "j"},   // U+091C
	"झ": {Category: brahmic.CatConsonant, BaseRom: "jh"},  // U+091D
	"ञ": {Category: brahmic.CatConsonant, BaseRom: "ny"},  // U+091E
	"ट": {Category: brahmic.CatConsonant, BaseRom: "t"},   // U+091F
	"ठ": {Category: brahmic.CatConsonant, BaseRom: "th"},  // U+0920
	"ड": {Category: brahmic.CatConsonant, BaseRom: "d"},   // U+0921
	"ढ": {Category: brahmic.CatConsonant, BaseRom: "dh"},  // U+0922
	"ण": {Category: brahmic.CatConsonant, BaseRom: "n"},   // U+0923
	"त": {Category: brahmic.CatConsonant, BaseRom: "t"},   // U+0924
	"थ": {Category: brahmic.CatConsonant, BaseRom: "th"},  // U+0925
	"द": {Category: brahmic.CatConsonant, BaseRom: "d"},   // U+0926
	"ध": {Category: brahmic.CatConsonant, BaseRom: "dh"},  // U+0927
	"न": {Category: brahmic.CatConsonant, BaseRom: "n"},   // U+0928
	"ऩ": {Category: brahmic.CatConsonant, BaseRom: "nh"},  // U+0929 (nuqta)
	"प": {Category: brahmic.CatConsonant, BaseRom: "p"},   // U+092A
	"फ": {Category: brahmic.CatConsonant, BaseRom: "ph"},  // U+092B
	"ब": {Category: brahmic.CatConsonant, BaseRom: "b"},   // U+092C
	"भ": {Category: brahmic.CatConsonant, BaseRom: "bh"},  // U+092D
	"म": {Category: brahmic.CatConsonant, BaseRom: "m"},   // U+092E
	"य": {Category: brahmic.CatConsonant, BaseRom: "y"},   // U+092F
	"र": {Category: brahmic.CatConsonant, BaseRom: "r"},   // U+0930
	"ऱ": {Category: brahmic.CatConsonant, BaseRom: "rh"},  // U+0931 (nuqta)
	"ल": {Category: brahmic.CatConsonant, BaseRom: "l"},   // U+0932
	"ळ": {Category: brahmic.CatConsonant, BaseRom: "l"},   // U+0933
	"ऴ": {Category: brahmic.CatConsonant, BaseRom: "lh"},  // U+0934 (nuqta)
	"व": {Category: brahmic.CatConsonant, BaseRom: "v"},   // U+0935
	"श": {Category: brahmic.CatConsonant, BaseRom: "sh"},  // U+0936
	"ष": {Category: brahmic.CatConsonant, BaseRom: "sh"},  // U+0937
	"स": {Category: brahmic.CatConsonant, BaseRom: "s"},   // U+0938
	"ह": {Category: brahmic.CatConsonant, BaseRom: "h"},   // U+0939

	// Nukta and other marks
	"़": {Category: brahmic.CatNukta, BaseRom: ""},  // U+093C
	"ऽ": {Category: brahmic.CatSymbol, BaseRom: ""}, // U+093D (avagraha)

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
	"ॐ": {Category: brahmic.CatConsonant, BaseRom: "om"}, // U+0950

	// Consonants with Nukta (Urdu/Persian sounds)
	"क़": {Category: brahmic.CatConsonant, BaseRom: "q"},  // U+0958
	"ख़": {Category: brahmic.CatConsonant, BaseRom: "kh"}, // U+0959
	"ग़": {Category: brahmic.CatConsonant, BaseRom: "gh"}, // U+095A
	"ज़": {Category: brahmic.CatConsonant, BaseRom: "z"},  // U+095B
	"ड़": {Category: brahmic.CatConsonant, BaseRom: "r"},  // U+095C
	"ढ़": {Category: brahmic.CatConsonant, BaseRom: "rh"}, // U+095D
	"फ़": {Category: brahmic.CatConsonant, BaseRom: "f"},  // U+095E
	"य़": {Category: brahmic.CatConsonant, BaseRom: "yh"}, // U+095F

	// Additional vowels
	"ॠ": {Category: brahmic.CatVowel, BaseRom: "ri"}, // U+0960
	"ॡ": {Category: brahmic.CatVowel, BaseRom: "ll"}, // U+0961
	"ॢ": {Category: brahmic.CatMatra, BaseRom: "l"},  // U+0962
	"ॣ": {Category: brahmic.CatMatra, BaseRom: "ll"}, // U+0963

	// Devanagari Numerals
	"०": {Category: brahmic.CatNumber, BaseRom: "0"}, // U+0966
	"१": {Category: brahmic.CatNumber, BaseRom: "1"}, // U+0967
	"२": {Category: brahmic.CatNumber, BaseRom: "2"}, // U+0968
	"३": {Category: brahmic.CatNumber, BaseRom: "3"}, // U+0969
	"४": {Category: brahmic.CatNumber, BaseRom: "4"}, // U+096A
	"५": {Category: brahmic.CatNumber, BaseRom: "5"}, // U+096B
	"६": {Category: brahmic.CatNumber, BaseRom: "6"}, // U+096C
	"७": {Category: brahmic.CatNumber, BaseRom: "7"}, // U+096D
	"८": {Category: brahmic.CatNumber, BaseRom: "8"}, // U+096E
	"९": {Category: brahmic.CatNumber, BaseRom: "9"}, // U+096F

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

// Hindi implements the brahmic.Language interface.
type Hindi struct{}

// Name returns the language identifier.
func (h Hindi) Name() string {
	return "hindi"
}

// Symbols returns the Hindi symbol map.
func (h Hindi) Symbols() brahmic.SymbolMap {
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

// Nukta returns the Hindi nukta character.
func (h Hindi) Nukta() string {
	return Nukta
}

// Rules returns the Hindi transliteration rules.
// This implements brahmic.RuleProvider.
func (h Hindi) Rules() []brahmic.Rule {
	return Rules()
}
