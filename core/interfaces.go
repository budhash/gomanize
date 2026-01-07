package core

// Parser converts script input into Units.
type Parser interface {
	// Parse converts input text into a Word structure.
	Parse(input string, symbols SymbolMap) *Word
}

// Renderer converts Units into romanized output.
type Renderer interface {
	// Render converts a Word into a romanized string.
	Render(word *Word) string
}

// Script defines a script family (Brahmic, Arabic, Latin, etc.).
// Scripts provide parsing and rendering capabilities.
type Script interface {
	// Name returns the script identifier (e.g., "brahmic", "arabic").
	Name() string

	// NewParser creates a parser for this script.
	// The config parameter is script-specific (e.g., brahmic.Config).
	NewParser(config interface{}) Parser

	// NewRenderer creates a renderer for this script.
	NewRenderer() Renderer

	// PrepareWord performs script-specific processing after parsing.
	// For example, Brahmic scripts identify consonant runs here.
	PrepareWord(word *Word)

	// Categories returns script-specific categories.
	Categories() []Category

	// DebugMetaExtractor returns a function that extracts script-specific
	// metadata for debugging (e.g., schwa state for Brahmic).
	// Returns nil if no metadata is available.
	DebugMetaExtractor() func(*Unit) string
}

// Language defines language-specific behavior within a script.
type Language interface {
	// Name returns the language identifier (e.g., "hindi", "marathi").
	Name() string

	// Script returns the script used by this language.
	Script() Script

	// Symbols returns the symbol map for this language.
	Symbols() SymbolMap

	// ScriptConfig returns script-specific configuration.
	// For Brahmic: brahmic.Config{Halant: "्", Nukta: "़"}
	ScriptConfig() interface{}

	// Rules returns the complete rule catalog for this language.
	// This includes ALL possible rules; schemes select from this catalog.
	Rules() RuleCatalog
}

// Scheme selects which rules to use from a language's catalog.
// Schemes are language-agnostic - they work with any language.
type Scheme interface {
	// Name returns the scheme identifier (e.g., "colloquial", "iast").
	Name() string

	// SelectRules selects rules from the language's catalog.
	// Returns the rules to apply for this scheme.
	SelectRules(catalog RuleCatalog) []Rule
}
