package core

// Engine orchestrates the transliteration pipeline.
type Engine struct {
	lang       Language
	scheme     Scheme
	script     Script
	symbols    SymbolMap
	config     interface{}
	ruleEngine *RuleEngine
	renderer   Renderer
}

// NewEngine creates an engine for a language + scheme combination.
func NewEngine(lang Language, scheme Scheme) *Engine {
	// Get language's complete rule catalog
	catalog := lang.Rules()

	// Scheme selects which rules to use
	selectedRules := scheme.SelectRules(catalog)

	script := lang.Script()

	return &Engine{
		lang:       lang,
		scheme:     scheme,
		script:     script,
		symbols:    lang.Symbols(),
		config:     lang.ScriptConfig(),
		ruleEngine: NewRuleEngine(selectedRules),
		renderer:   script.NewRenderer(),
	}
}

// Transliterate converts script text to romanized form using default options.
func (e *Engine) Transliterate(input string) string {
	return e.TransliterateWithOptions(input, DefaultOptions())
}

// TransliterateWithOptions converts script text to romanized form with custom options.
func (e *Engine) TransliterateWithOptions(input string, opts Options) string {
	// 1. Parse (Script handles script-specific parsing)
	parser := e.script.NewParser(e.config)
	word := parser.Parse(input, e.symbols)
	word.Options = opts

	// 2. Prepare (Script-specific processing, e.g., IdentifyRuns)
	e.script.PrepareWord(word)

	// 3. Apply rules (selected by Scheme from Language catalog)
	e.ruleEngine.Apply(word)

	// 4. Render (Script handles script-specific rendering)
	return e.renderer.Render(word)
}

// Language returns the engine's language.
func (e *Engine) Language() Language {
	return e.lang
}

// Scheme returns the engine's scheme.
func (e *Engine) Scheme() Scheme {
	return e.scheme
}
