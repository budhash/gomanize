package brahmic

import "github.com/budhash/gomanize/core"

// Brahmic-specific categories (starting at 100 to avoid collision with core)
const (
	CatHalant       core.Category = 100 + iota // Halant/virama
	CatAnusvara                                // Anusvara (ं)
	CatVisarga                                 // Visarga (ः)
	CatChandrabindu                            // Chandrabindu (ँ)
	CatNukta                                   // Nukta (़)
	CatMatra                                   // Dependent vowel sign
	CatConjunct                                // Multi-character conjunct
)

// CategoryName returns a human-readable name for Brahmic categories.
func CategoryName(c core.Category) string {
	switch c {
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
	case CatMatra:
		return "Matra"
	case CatConjunct:
		return "Conjunct"
	default:
		return c.String()
	}
}

// IsBrahmicCategory returns true if the category is Brahmic-specific.
func IsBrahmicCategory(c core.Category) bool {
	return c >= 100
}

// BrahmicCategories returns all Brahmic-specific categories.
func BrahmicCategories() []core.Category {
	return []core.Category{
		CatHalant,
		CatAnusvara,
		CatVisarga,
		CatChandrabindu,
		CatNukta,
		CatMatra,
		CatConjunct,
	}
}

// CategoryToUnitType converts a category to the appropriate UnitType.
func CategoryToUnitType(cat core.Category) core.UnitType {
	switch cat {
	case core.CatVowel:
		return core.UnitVowel
	case core.CatConsonant:
		return core.UnitConsonant
	case core.CatNumber:
		return core.UnitNumber
	case core.CatSymbol:
		return core.UnitSymbol
	case CatHalant, CatNukta:
		return core.UnitSymbol // Processed but not rendered as separate units
	case CatAnusvara, CatVisarga, CatChandrabindu:
		return core.UnitModifier
	case CatMatra:
		return core.UnitVowel
	case CatConjunct:
		return core.UnitConjunct
	default:
		return core.UnitSymbol
	}
}

// IsVowelCategory returns true if the category represents a vowel.
func IsVowelCategory(cat core.Category) bool {
	return cat == core.CatVowel || cat == CatMatra
}

// IsConsonantCategory returns true if the category represents a consonant.
func IsConsonantCategory(cat core.Category) bool {
	return cat == core.CatConsonant
}

// IsModifierCategory returns true if the category represents a modifier.
func IsModifierCategory(cat core.Category) bool {
	return cat == CatAnusvara || cat == CatVisarga || cat == CatChandrabindu
}
