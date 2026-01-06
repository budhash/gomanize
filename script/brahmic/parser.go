package brahmic

// Parser converts Brahmic script text into a Word structure.
// Handles halant tracking and nukta combinations.
type Parser struct {
	symbols   SymbolMap
	multiChar []string
	halant    string
	nukta     string
}

// NewParser creates a parser with the given symbol map and configuration.
func NewParser(symbols SymbolMap, multiChar []string, halant, nukta string) *Parser {
	return &Parser{
		symbols:   symbols,
		multiChar: multiChar,
		halant:    halant,
		nukta:     nukta,
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
		if p.nukta != "" && pos+1 < len(runes) && string(runes[pos+1]) == p.nukta {
			combined := char + p.nukta
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
				Runes:       []rune{runes[pos]},
				Start:       runeIdx,
				End:         runeIdx + 1,
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
		Runes:       runes,
		Start:       runeIdx,
		End:         runeIdx + len(runes),
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
