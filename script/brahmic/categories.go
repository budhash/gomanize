package brahmic

// Category classifies characters in Brahmic scripts.
type Category int

const (
	CatVowel        Category = iota // Independent vowels (Devanagari अ, आ, इ, etc.)
	CatMatra                        // Dependent vowel signs (Devanagari ा, ि, ी, etc.)
	CatConsonant                    // Consonants (Devanagari क, ख, ग, etc.)
	CatConjunct                     // Multi-char conjuncts (Devanagari ज्ञ, क्ष, etc.)
	CatNumber                       // Script numerals (Devanagari ०-९)
	CatHalant                       // Virama - suppresses inherent vowel
	CatAnusvara                     // Nasal modifier (Devanagari ं)
	CatVisarga                      // Aspirate modifier (Devanagari ः)
	CatChandrabindu                 // Nasalization (Devanagari ँ)
	CatNukta                        // Nukta - modifies consonant (Devanagari ़)
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

// SymbolInfo holds romanization data for a Brahmic script character.
type SymbolInfo struct {
	Category Category
	BaseRom  string // Default romanization
}

// SymbolMap maps script strings to their romanization info.
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

// IsNumber returns true if the character is a script numeral.
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
