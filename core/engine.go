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
// Panics if the language returns a nil Script.
func NewEngine(lang Language, scheme Scheme) *Engine {
	script := lang.Script()
	if script == nil {
		panic("core.NewEngine: lang.Script() returned nil")
	}

	// Get language's complete rule catalog
	catalog := lang.Rules()

	// Scheme selects which rules to use
	selectedRules := scheme.SelectRules(catalog)

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
	result, _ := e.transliterateInternal(input, opts)
	return result
}

// TransliterateDebug converts script text and returns debug information.
func (e *Engine) TransliterateDebug(input string, opts Options) (string, *DebugInfo) {
	opts.Debug = true
	return e.transliterateInternal(input, opts)
}

// transliterateInternal is the core transliteration logic.
func (e *Engine) transliterateInternal(input string, opts Options) (string, *DebugInfo) {
	// 1. Parse (Script handles script-specific parsing)
	parser := e.script.NewParser(e.config)
	word := parser.Parse(input, e.symbols)
	word.Options = opts

	// 2. Prepare (Script-specific processing, e.g., IdentifyRuns)
	e.script.PrepareWord(word)

	// Enable debug if requested
	if opts.Debug {
		e.ruleEngine.EnableDebug(true)
		// Set metadata extractor from script
		if metaFn := e.script.DebugMetaExtractor(); metaFn != nil {
			e.ruleEngine.SetDebugMetaExtractor(metaFn)
		}
	}

	// 3. Apply rules (selected by Scheme from Language catalog)
	e.ruleEngine.Apply(word)

	// 4. Render (Script handles script-specific rendering)
	result := e.renderer.Render(word)

	// Collect debug info if enabled
	var debug *DebugInfo
	if opts.Debug {
		debug = e.collectDebugInfo(word, input, result)
		e.ruleEngine.EnableDebug(false) // Reset for next call
	}

	return result, debug
}

// collectDebugInfo gathers debugging information after transliteration.
func (e *Engine) collectDebugInfo(word *Word, input, output string) *DebugInfo {
	info := &DebugInfo{
		Input:  input,
		Output: output,
		Traces: e.ruleEngine.Traces(),
	}

	// Collect unit info
	metaFn := e.script.DebugMetaExtractor()
	for i, unit := range word.Units {
		ud := UnitDebug{
			Index:   i,
			Chars:   string(unit.Runes),
			Type:    unit.Type.String(),
			BaseRom: unit.BaseRom,
			RunePos: unit.Start.Rune,
		}
		if metaFn != nil {
			ud.Metadata = metaFn(unit)
		}
		info.Units = append(info.Units, ud)
	}

	return info
}

// Language returns the engine's language.
func (e *Engine) Language() Language {
	return e.lang
}

// Scheme returns the engine's scheme.
func (e *Engine) Scheme() Scheme {
	return e.scheme
}
