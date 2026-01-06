package engine

import (
	"unicode/utf8"
)

// Parser converts Devanagari text into a Word structure.
type Parser struct {
	symbols   SymbolMap
	multiChar []string
	halant    string
}

// NewParser creates a parser with the given symbol map and multi-char sequences.
func NewParser(symbols SymbolMap, multiChar []string, halant string) *Parser {
	return &Parser{
		symbols:   symbols,
		multiChar: multiChar,
		halant:    halant,
	}
}

// Parse converts input text into a Word with linked Units.
func (p *Parser) Parse(input string) *Word {
	word := NewWord(input)
	runes := []rune(input)
	pos := 0
	runeIdx := 0

	// Track if previous character was halant
	afterHalant := false

	for pos < len(runes) {
		// Try multi-char sequences first (longest match)
		matched := false
		for _, mc := range p.multiChar {
			mcRunes := []rune(mc)
			if pos+len(mcRunes) <= len(runes) && string(runes[pos:pos+len(mcRunes)]) == mc {
				// Found a multi-char match
				unit := p.createUnit(mc, mcRunes, runeIdx, afterHalant)
				word.AddUnit(unit)

				pos += len(mcRunes)
				runeIdx += len(mcRunes)
				afterHalant = false
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		// Try single character with optional nukta
		char := string(runes[pos])

		// Check for character + nukta combination
		if pos+1 < len(runes) && string(runes[pos+1]) == "़" {
			combined := char + "़"
			if info, ok := p.symbols.Lookup(combined); ok {
				unit := p.createUnitWithInfo([]rune{runes[pos], runes[pos+1]}, runeIdx, afterHalant, info)
				word.AddUnit(unit)
				pos += 2
				runeIdx += 2
				afterHalant = false
				continue
			}
		}

		// Single character lookup
		if info, ok := p.symbols.Lookup(char); ok {
			// Handle halant specially - don't create a unit, just track it
			if info.Category == CatHalant {
				afterHalant = true
				pos++
				runeIdx++
				continue
			}

			unit := p.createUnitWithInfo([]rune{runes[pos]}, runeIdx, afterHalant, info)
			word.AddUnit(unit)
			afterHalant = false
		} else {
			// Unknown character - create a symbol unit
			unit := &Unit{
				Runes: []rune{runes[pos]},
				Start: Position{
					Offset: byteOffset(input, runeIdx),
					Rune:   runeIdx,
				},
				End: Position{
					Offset: byteOffset(input, runeIdx+1),
					Rune:   runeIdx + 1,
				},
				Type:        UnitSymbol,
				BaseRom:     char,
				AfterHalant: afterHalant,
				Schwa:       SchwaPending,
			}
			word.AddUnit(unit)
			afterHalant = false
		}

		pos++
		runeIdx++
	}

	return word
}

// createUnit creates a Unit from a character string.
func (p *Parser) createUnit(char string, runes []rune, runeIdx int, afterHalant bool) *Unit {
	info, _ := p.symbols.Lookup(char)
	return p.createUnitWithInfo(runes, runeIdx, afterHalant, info)
}

// createUnitWithInfo creates a Unit with a pre-looked-up SymbolInfo.
func (p *Parser) createUnitWithInfo(runes []rune, runeIdx int, afterHalant bool, info SymbolInfo) *Unit {
	unitType := CategoryToUnitType(info.Category)

	unit := &Unit{
		Runes: runes,
		Start: Position{
			Offset: -1, // Not calculated for known symbols (Rune index suffices)
			Rune:   runeIdx,
		},
		End: Position{
			Offset: -1, // Not calculated for known symbols
			Rune:   runeIdx + len(runes),
		},
		Type:        unitType,
		BaseRom:     info.BaseRom,
		AfterHalant: afterHalant,
		Schwa:       SchwaPending,
	}

	// Only consonants and conjuncts need schwa tracking
	if unitType != UnitConsonant && unitType != UnitConjunct {
		unit.Schwa = SchwaKeep // Vowels/numbers/symbols don't have schwa decisions
	}

	return unit
}

// byteOffset calculates the byte offset for a rune index in a string.
func byteOffset(s string, runeIdx int) int {
	offset := 0
	for i := 0; i < runeIdx && offset < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[offset:])
		offset += size
	}
	return offset
}
