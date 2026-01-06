// Package engine provides a multi-pass romanization engine for Indic scripts.
//
// The engine follows a Parse → Rules → Render pipeline:
//  1. Parse: Convert Devanagari input into linked Unit structures
//  2. Rules: Apply schwa deletion and other transformation rules (Phase 2)
//  3. Render: Convert Units into romanized output
//
// This separation enables rules to see full word context before making decisions.
package engine

// Language defines the interface for language-specific romanization.
type Language interface {
	// Name returns the language identifier (e.g., "hindi", "marathi").
	Name() string

	// Symbols returns the symbol map for this language.
	Symbols() SymbolMap

	// MultiChar returns multi-character sequences to match first.
	MultiChar() []string

	// Halant returns the halant (virama) character for this script.
	Halant() string
}

// Engine is the main romanization engine.
type Engine struct {
	language Language
	parser   *Parser
	renderer *Renderer
}

// New creates a new Engine for the given language.
func New(lang Language) *Engine {
	return &Engine{
		language: lang,
		parser:   NewParser(lang.Symbols(), lang.MultiChar(), lang.Halant()),
		renderer: NewRenderer(),
	}
}

// Transliterate converts Devanagari text to romanized form.
func (e *Engine) Transliterate(input string) string {
	word := e.parser.Parse(input)
	IdentifyRuns(word)
	// Phase 2 will add: ApplyRules(word, scheme)
	return e.renderer.Render(word)
}

// TransliterateDebug returns detailed debug info along with the output.
func (e *Engine) TransliterateDebug(input string) string {
	word := e.parser.Parse(input)
	IdentifyRuns(word)
	return e.renderer.RenderDebug(word)
}

// ParseOnly parses the input without rendering (useful for testing).
func (e *Engine) ParseOnly(input string) *Word {
	word := e.parser.Parse(input)
	IdentifyRuns(word)
	return word
}
