package brahmic

import (
	"fmt"

	"github.com/budhash/gomanize/core"
)

// Parser converts Brahmic script text into a Word structure.
// Handles halant tracking and nukta combinations.
// Implements core.Parser interface.
type Parser struct {
	vowellessConsonants []rune
	profile             *ScriptProfile
	multiChar           []string
	halant              string
	nukta               string
	canonical           *CanonicalForms
}

// NewParser creates a parser with the given configuration.
// Panics if config is not a brahmic.Config type.
func NewParser(config interface{}) *Parser {
	cfg, ok := config.(Config)
	if !ok {
		panic(fmt.Sprintf("brahmic.NewParser: expected brahmic.Config, got %T", config))
	}
	cfg = cfg.normalize()
	return &Parser{
		profile:             cfg.Profile,
		vowellessConsonants: append([]rune(nil), cfg.VowellessConsonants...),
		halant:              cfg.Halant,
		nukta:               cfg.Nukta,
		multiChar:           cfg.MultiChar,
		canonical:           cfg.Canonical,
	}
}

// SetMultiChar sets the multi-character sequences to match.
func (p *Parser) SetMultiChar(mc []string) {
	p.multiChar = mc
}

// Parse converts input text into a Word with linked Units.
// Implements core.Parser interface.
func (p *Parser) Parse(input string, symbols core.SymbolMap) *core.Word {
	// Canonicalize BEFORE parsing (idempotent; the engine already did it).
	// Format characters in script text carry no phonetic content: emitting them
	// as units corrupts output, and skipping them during the walk would leave
	// Unit.Start.Rune pointing at raw-input positions, breaking rune-indexed
	// schwa rules and the learned models' feature windows. Removing them first
	// also lets multi-char sequences match across them (ज्&#8205;ञ parses as the
	// ज्ञ conjunct). Format characters outside script text (ZWJ in emoji, RTL
	// marks) are kept. Word.Original is the canonical form, so unit indices
	// always align with it.
	runes := []rune(Canonicalize(input, Config{Canonical: p.canonical}, symbols))
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
				if bd := GetBrahmicData(unit); bd != nil {
					bd.Nukta = true
				}
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

	if afterHalant && len(word.Units) > 0 {
		if bd := GetBrahmicData(word.Units[len(word.Units)-1]); bd != nil {
			bd.TrailingHalant = true
		}
	}
	SetWordBrahmicData(word, &WordBrahmicData{Profile: p.profile})
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
	bd.IsMatra = info.Category == CatMatra
	if unitType == core.UnitConsonant && len(runes) == 1 && containsRune(p.vowellessConsonants, runes[0]) {
		bd.NoInherentVowel = true
		bd.Schwa = SchwaDelete
	}

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
