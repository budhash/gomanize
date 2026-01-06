package engine

// Category classifies Devanagari characters.
type Category int

const (
	CatVowel        Category = iota // Independent vowels (अ, आ, इ, etc.)
	CatMatra                        // Dependent vowel signs (ा, ि, ी, etc.)
	CatConsonant                    // Consonants (क, ख, ग, etc.)
	CatConjunct                     // Multi-char conjuncts (ज्ञ, क्ष, etc.)
	CatNumber                       // Devanagari numerals (०-९)
	CatHalant                       // Virama (्)
	CatAnusvara                     // Anusvara (ं)
	CatVisarga                      // Visarga (ः)
	CatChandrabindu                 // Chandrabindu (ँ)
	CatNukta                        // Nukta (़)
	CatSymbol                       // Other symbols
)

func (c Category) String() string {
	switch c {
	case CatVowel:
		return "Vowel"
	case CatMatra:
		return "Matra"
	case CatConsonant:
		return "Consonant"
	case CatConjunct:
		return "Conjunct"
	case CatNumber:
		return "Number"
	case CatHalant:
		return "Halant"
	case CatAnusvara:
		return "Anusvara"
	case CatVisarga:
		return "Visarga"
	case CatChandrabindu:
		return "Chandrabindu"
	case CatNukta:
		return "Nukta"
	case CatSymbol:
		return "Symbol"
	default:
		return "Unknown"
	}
}

// SymbolInfo holds romanization data for a Devanagari character.
type SymbolInfo struct {
	Category Category
	BaseRom  string // Default romanization
}

// SymbolMap maps Devanagari strings to their romanization info.
// The key is a string to support multi-character symbols (conjuncts).
type SymbolMap map[string]SymbolInfo

// Lookup returns the SymbolInfo for a character, and whether it was found.
func (m SymbolMap) Lookup(s string) (SymbolInfo, bool) {
	info, ok := m[s]
	return info, ok
}

// IsVowel returns true if the character is a vowel (independent or matra).
func (m SymbolMap) IsVowel(s string) bool {
	info, ok := m[s]
	if !ok {
		return false
	}
	return info.Category == CatVowel || info.Category == CatMatra
}

// IsConsonant returns true if the character is a consonant.
func (m SymbolMap) IsConsonant(s string) bool {
	info, ok := m[s]
	if !ok {
		return false
	}
	return info.Category == CatConsonant
}

// IsConjunct returns true if the character is a multi-char conjunct.
func (m SymbolMap) IsConjunct(s string) bool {
	info, ok := m[s]
	if !ok {
		return false
	}
	return info.Category == CatConjunct
}

// IsHalant returns true if the character is a halant (virama).
func (m SymbolMap) IsHalant(s string) bool {
	info, ok := m[s]
	if !ok {
		return false
	}
	return info.Category == CatHalant
}

// IsNumber returns true if the character is a Devanagari numeral.
func (m SymbolMap) IsNumber(s string) bool {
	info, ok := m[s]
	if !ok {
		return false
	}
	return info.Category == CatNumber
}

// CategoryToUnitType converts a symbol Category to a UnitType.
func CategoryToUnitType(cat Category) UnitType {
	switch cat {
	case CatVowel, CatMatra:
		return UnitVowel
	case CatAnusvara, CatVisarga, CatChandrabindu:
		return UnitModifier // Modifiers don't suppress preceding schwa
	case CatConsonant:
		return UnitConsonant
	case CatConjunct:
		return UnitConjunct
	case CatNumber:
		return UnitNumber
	default:
		return UnitSymbol
	}
}
