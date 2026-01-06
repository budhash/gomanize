// Package brahmic provides a romanization engine for Brahmic script family.
//
// The engine follows a Parse → Prepare → Rules → Render pipeline:
//  1. Parse: Convert Brahmic script input into linked Unit structures
//  2. Prepare: Identify consonant runs for schwa deletion decisions
//  3. Rules: Apply schwa deletion and other transformation rules
//  4. Render: Convert Units into romanized output
//
// This package implements core.Script interface.
package brahmic

import "github.com/budhash/gomanize/core"

// Script implements the core.Script interface for Brahmic scripts.
type Script struct{}

// New creates a new Brahmic script instance.
func New() *Script {
	return &Script{}
}

// Name returns the script identifier.
func (s *Script) Name() string {
	return "brahmic"
}

// NewParser creates a parser for Brahmic scripts.
// The config parameter must be brahmic.Config.
func (s *Script) NewParser(config interface{}) core.Parser {
	return NewParser(config)
}

// NewRenderer creates a renderer for Brahmic scripts.
func (s *Script) NewRenderer() core.Renderer {
	return NewRenderer()
}

// PrepareWord performs Brahmic-specific processing after parsing.
// This identifies consonant runs for coordinated schwa deletion.
func (s *Script) PrepareWord(word *core.Word) {
	IdentifyRuns(word)
}

// Categories returns Brahmic-specific categories.
func (s *Script) Categories() []core.Category {
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
