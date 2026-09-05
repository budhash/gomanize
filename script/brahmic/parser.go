package brahmic

import (
	"unicode"

	"fmt"

	"github.com/budhash/gomanize/core"
)

// Parser converts Brahmic script text into a Word structure.
// Handles halant tracking and nukta combinations.
// Implements core.Parser interface.
type Parser struct {
	multiChar []string
	halant    string
	nukta     string
}

// NewParser creates a parser with the given configuration.
// Panics if config is not a brahmic.Config type.
func NewParser(config interface{}) *Parser {
	cfg, ok := config.(Config)
	if !ok {
		panic(fmt.Sprintf("brahmic.NewParser: expected brahmic.Config, got %T", config))
	}
	return &Parser{
		halant:    cfg.Halant,
		nukta:     cfg.Nukta,
		multiChar: cfg.MultiChar,
	}
}

// SetMultiChar sets the multi-character sequences to match.
func (p *Parser) SetMultiChar(mc []string) {
	p.multiChar = mc
}

// Parse converts input text into a Word with linked Units.
// Implements core.Parser interface.
func (p *Parser) Parse(input string, symbols core.SymbolMap) *core.Word {
	// Strip format characters (ZWNJ/ZWJ etc., Unicode category Cf) BEFORE
	// parsing. They control conjunct rendering but carry no phonetic content;
	// emitting them as units corrupts output, and merely skipping them during
	// the walk leaves Unit.Start.Rune pointing at raw-input positions, which
	// breaks rune-indexed schwa rules and the schwa model's feature window.
	// Stripping first also lets multi-char sequences match across them
	// (ज्&#8205;ञ still parses as the ज्ञ conjunct). Word.Original is the
	// stripped form so unit indices always align with it.
	runes := make([]rune, 0, len(input))
	for _, r := range input {
		if !unicode.Is(unicode.Cf, r) {
			runes = append(runes, r)
		}
	}
	word := core.NewWord(string(runes))
	pos := 0
	runeIdx := 0

	// Track if previous character was halant
	afterHalant := false

	for pos < len(runes) {
		// Try multi-char sequences first (first match in MultiChar order)
		matched := false
		for _, mc := range p.multiChar {
			mcRunes := []rune(mc)
			if pos+len(mcRunes) <= len(runes) && string(runes[pos:pos+len(mcRunes)]) == mc {
				// Found a multi-char match
				unit := p.createUnit(mc, mcRunes, runeIdx, afterHalant, symbols)
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
			if info, ok := symbols[combined]; ok {
				unit := p.createUnitWithInfo([]rune{runes[pos], runes[pos+1]}, runeIdx, afterHalant, info)
				word.AddUnit(unit)
				pos += 2
				runeIdx += 2
				afterHalant = false
				continue
			}
		}

		// Single character lookup
		if info, ok := symbols[char]; ok {
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
			bd := NewBrahmicData()
			bd.AfterHalant = afterHalant
			bd.Schwa = SchwaKeep // Unknown symbols don't have schwa decisions

			unit := &core.Unit{
				Runes:   []rune{runes[pos]},
				Start:   core.Position{Rune: runeIdx},
				End:     core.Position{Rune: runeIdx + 1},
				Type:    core.UnitSymbol,
				BaseRom: char,
			}
			SetBrahmicData(unit, bd)
			word.AddUnit(unit)
			afterHalant = false
		}

		pos++
		runeIdx++
	}

	return word
}

// createUnit creates a Unit from a character string.
func (p *Parser) createUnit(char string, runes []rune, runeIdx int, afterHalant bool, symbols core.SymbolMap) *core.Unit {
	info := symbols[char]
	return p.createUnitWithInfo(runes, runeIdx, afterHalant, info)
}

// createUnitWithInfo creates a Unit with a pre-looked-up SymbolInfo.
func (p *Parser) createUnitWithInfo(runes []rune, runeIdx int, afterHalant bool, info core.SymbolInfo) *core.Unit {
	unitType := CategoryToUnitType(info.Category)

	bd := NewBrahmicData()
	bd.AfterHalant = afterHalant

	// Only consonants and conjuncts need schwa tracking
	if unitType != core.UnitConsonant && unitType != core.UnitConjunct {
		bd.Schwa = SchwaKeep // Vowels/numbers/symbols don't have schwa decisions
	}

	unit := &core.Unit{
		Runes:   runes,
		Start:   core.Position{Rune: runeIdx},
		End:     core.Position{Rune: runeIdx + len(runes)},
		Type:    unitType,
		BaseRom: info.BaseRom,
	}
	SetBrahmicData(unit, bd)

	return unit
}
