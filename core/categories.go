package core

// Category classifies symbols for parsing.
// Base categories are defined here; scripts can extend with their own.
type Category int

const (
	CatUnknown   Category = iota // Unknown/unrecognized
	CatVowel                     // Independent vowels
	CatConsonant                 // Consonants
	CatNumber                    // Numerals
	CatSymbol                    // Punctuation, other symbols
	// Script-specific categories start at 100+
	// e.g., Brahmic: CatHalant = 100, CatAnusvara = 101, etc.
)

func (c Category) String() string {
	switch c {
	case CatUnknown:
		return "Unknown"
	case CatVowel:
		return "Vowel"
	case CatConsonant:
		return "Consonant"
	case CatNumber:
		return "Number"
	case CatSymbol:
		return "Symbol"
	default:
		// Script-specific categories
		return "Script-specific"
	}
}

// SymbolInfo holds romanization info for a symbol.
type SymbolInfo struct {
	Category Category
	BaseRom  string
}

// SymbolMap maps script characters to their romanization info.
type SymbolMap map[string]SymbolInfo
