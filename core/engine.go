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

// engineConfig holds configuration for engine creation.
type engineConfig struct {
	disableRules []string
	enableRules  []string
}

// EngineOption configures engine creation.
type EngineOption func(*engineConfig)

// WithDisabledRules disables rules matching the given patterns at engine creation.
// Patterns can be exact names or glob patterns (e.g., "schwa.*", "vowel.long-aa.*").
func WithDisabledRules(patterns ...string) EngineOption {
	return func(cfg *engineConfig) {
		cfg.disableRules = append(cfg.disableRules, patterns...)
	}
}

// WithEnabledRules enables rules matching the given patterns at engine creation.
// Useful for enabling rules that are disabled by default.
// Patterns can be exact names or glob patterns (e.g., "vowel.long-aa.all").
func WithEnabledRules(patterns ...string) EngineOption {
	return func(cfg *engineConfig) {
		cfg.enableRules = append(cfg.enableRules, patterns...)
	}
}

// NewEngine creates an engine for a language + scheme combination.
// Panics if the language returns a nil Script.
func NewEngine(lang Language, scheme Scheme, opts ...EngineOption) *Engine {
	script := lang.Script()
	if script == nil {
		panic("core.NewEngine: lang.Script() returned nil")
	}

	// Apply options
	cfg := &engineConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get language's complete rule catalog
	catalog := lang.Rules()

	// Scheme selects which rules to use
	selectedRules := scheme.SelectRules(catalog)

	// Create rule engine
	ruleEngine := NewRuleEngine(selectedRules)

	// Apply rule overrides from options
	for _, pattern := range cfg.disableRules {
		ruleEngine.DisableRule(pattern)
	}
	for _, pattern := range cfg.enableRules {
		ruleEngine.EnableRule(pattern)
	}

	return &Engine{
		lang:       lang,
		scheme:     scheme,
		script:     script,
		symbols:    lang.Symbols(),
		config:     lang.ScriptConfig(),
		ruleEngine: ruleEngine,
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

// LexiconProvider is an optional interface a Language may implement to supply a
// word-level romanization lexicon. When Options.Lexicon is set, the engine
// consults it before the rule pipeline; a hit short-circuits transliteration.
type LexiconProvider interface {
	LexiconLookup(word string) (string, bool)
}

// OptionsLexiconProvider is an optional, style-aware lexicon interface. It takes
// precedence over LexiconProvider, including on a miss: a provider may decline
// lookup for the requested style without falling through to a legacy lookup.
// Languages implementing only LexiconProvider retain their existing behavior.
type OptionsLexiconProvider interface {
	LexiconLookupWithOptions(word string, opts Options) (string, bool)
}

// Reranker is an optional interface a Language may implement to choose the most
// natural romanization among candidates produced under different rule
// configurations. Used when Options.Rerank is set.
type Reranker interface {
	RerankRomans(candidates []string) string
}

// CandidateRenderer runs a candidate language with the caller's scheme and rule
// enable/disable state. Lexicon and recursive reranking are disabled; candidate
// debug traces are not returned. Providers must use it synchronously.
type CandidateRenderer func(Language, Options) string

// NativeReranker optionally generates and ranks candidates using the source word.
// It takes precedence over Reranker, including when it declines. A false result
// resumes the original pipeline with the original options. Lexicon hits win first.
type NativeReranker interface {
	RerankNative(input string, opts Options, render CandidateRenderer) (string, bool)
}

// transliterateInternal is the core transliteration logic.
func (e *Engine) transliterateInternal(input string, opts Options) (string, *DebugInfo) {
	// 0. Lexicon lookup (optional): known words get their attested spelling.
	if opts.Lexicon {
		if lp, ok := e.lang.(OptionsLexiconProvider); ok {
			if roman, found := lp.LexiconLookupWithOptions(input, opts); found {
				return roman, nil
			}
		} else if lp, ok := e.lang.(LexiconProvider); ok {
			if roman, found := lp.LexiconLookup(input); found {
				return roman, nil
			}
		}
	}

	// 0b. Optional candidate selection: source-aware providers own generation
	// and ranking; legacy providers score the engine-generated variants below.
	// A native decline preserves the caller's original pipeline.
	if opts.Rerank {
		if rr, ok := e.lang.(NativeReranker); ok {
			render := func(lang Language, candidate Options) string {
				candidate.Rerank, candidate.Lexicon, candidate.Debug = false, false, false
				var overrides []EngineOption
				for _, rule := range e.ruleEngine.ListRules("") {
					if rule.Enabled {
						overrides = append(overrides, WithEnabledRules(rule.Name))
					} else {
						overrides = append(overrides, WithDisabledRules(rule.Name))
					}
				}
				return NewEngine(lang, e.scheme, overrides...).TransliterateWithOptions(input, candidate)
			}
			if out, handled := rr.RerankNative(input, opts, render); handled {
				return out, nil
			}
		} else if rr, ok := e.lang.(Reranker); ok {
			base := opts
			base.Rerank = false
			variants := []Options{base}
			alt := base
			alt.SchwaModel = true
			variants = append(variants, alt)

			seen := make(map[string]bool)
			var cands []string
			for _, v := range variants {
				out, _ := e.transliterateInternal(input, v)
				if !seen[out] {
					seen[out] = true
					cands = append(cands, out)
				}
			}
			return rr.RerankRomans(cands), nil
		}
	}

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

// RuleEngine returns the engine's rule engine for direct manipulation.
// Use this to enable/disable rules after engine creation.
func (e *Engine) RuleEngine() *RuleEngine {
	return e.ruleEngine
}
