package core

import (
	"testing"
)

// mockScript implements Script interface for testing
type mockScript struct {
	name       string
	parser     Parser
	renderer   Renderer
	categories []Category
	metaFn     func(*Unit) string
}

func (m *mockScript) Name() string                           { return m.name }
func (m *mockScript) NewParser(config interface{}) Parser    { return m.parser }
func (m *mockScript) NewRenderer() Renderer                  { return m.renderer }
func (m *mockScript) PrepareWord(word *Word)                 {}
func (m *mockScript) Categories() []Category                 { return m.categories }
func (m *mockScript) DebugMetaExtractor() func(*Unit) string { return m.metaFn }

// mockParser implements Parser interface for testing
type mockParser struct {
	parseResult *Word
}

func (p *mockParser) Parse(input string, symbols SymbolMap) *Word {
	if p.parseResult != nil {
		return p.parseResult
	}
	word := NewWord(input)
	for _, r := range input {
		word.AddUnit(&Unit{
			Runes:   []rune{r},
			Type:    UnitConsonant,
			BaseRom: string(r),
		})
	}
	return word
}

// mockRenderer implements Renderer interface for testing
type mockRenderer struct{}

func (r *mockRenderer) Render(word *Word) string {
	result := ""
	for _, u := range word.Units {
		result += u.BaseRom
	}
	return result
}

// mockLanguage implements Language interface for testing
type mockLanguage struct {
	name    string
	script  Script
	symbols SymbolMap
	config  interface{}
	catalog RuleCatalog
}

func (l *mockLanguage) Name() string              { return l.name }
func (l *mockLanguage) Script() Script            { return l.script }
func (l *mockLanguage) Symbols() SymbolMap        { return l.symbols }
func (l *mockLanguage) ScriptConfig() interface{} { return l.config }
func (l *mockLanguage) Rules() RuleCatalog        { return l.catalog }

// mockScheme implements Scheme interface for testing
type mockScheme struct {
	name  string
	rules []Rule
}

func (s *mockScheme) Name() string                           { return s.name }
func (s *mockScheme) SelectRules(catalog RuleCatalog) []Rule { return s.rules }

func TestNewEnginePanicsOnNilScript(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewEngine should panic when lang.Script() returns nil")
		}
	}()

	lang := &mockLanguage{
		name:   "test",
		script: nil, // This should cause a panic
	}
	scheme := &mockScheme{name: "test"}

	NewEngine(lang, scheme)
}

func TestNewEngineSuccess(t *testing.T) {
	script := &mockScript{
		name:     "test-script",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}
	lang := &mockLanguage{
		name:    "test-lang",
		script:  script,
		symbols: make(SymbolMap),
		config:  nil,
		catalog: RuleCatalog{},
	}
	scheme := &mockScheme{name: "test-scheme"}

	engine := NewEngine(lang, scheme)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if engine.Language().Name() != "test-lang" {
		t.Errorf("Language().Name() = %q, want %q", engine.Language().Name(), "test-lang")
	}
	if engine.Scheme().Name() != "test-scheme" {
		t.Errorf("Scheme().Name() = %q, want %q", engine.Scheme().Name(), "test-scheme")
	}
}

func TestEngineTransliterate(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}
	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test"}

	engine := NewEngine(lang, scheme)
	result := engine.Transliterate("abc")

	if result != "abc" {
		t.Errorf("Transliterate() = %q, want %q", result, "abc")
	}
}

func TestEngineTransliterateWithOptions(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}
	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test"}

	engine := NewEngine(lang, scheme)
	opts := Options{LongVowels: true}
	result := engine.TransliterateWithOptions("abc", opts)

	if result != "abc" {
		t.Errorf("TransliterateWithOptions() = %q, want %q", result, "abc")
	}
}

func TestEngineTransliterateDebug(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
		metaFn:   func(u *Unit) string { return "meta-" + u.BaseRom },
	}
	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test"}

	engine := NewEngine(lang, scheme)
	result, debugInfo := engine.TransliterateDebug("ab", DefaultOptions())

	if result != "ab" {
		t.Errorf("TransliterateDebug() result = %q, want %q", result, "ab")
	}
	if debugInfo == nil {
		t.Fatal("TransliterateDebug() debugInfo is nil")
	}
	if debugInfo.Input != "ab" {
		t.Errorf("DebugInfo.Input = %q, want %q", debugInfo.Input, "ab")
	}
	if debugInfo.Output != "ab" {
		t.Errorf("DebugInfo.Output = %q, want %q", debugInfo.Output, "ab")
	}
	if len(debugInfo.Units) != 2 {
		t.Errorf("len(DebugInfo.Units) = %d, want 2", len(debugInfo.Units))
	}

	// Check metadata was extracted
	if debugInfo.Units[0].Metadata != "meta-a" {
		t.Errorf("Units[0].Metadata = %q, want %q", debugInfo.Units[0].Metadata, "meta-a")
	}
}

func TestEngineTransliterateDebugWithRules(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	rules := []Rule{
		{
			Name:      "uppercase",
			Phase:     PhaseRender,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action: func(u *Unit, w *Word) {
				if len(u.BaseRom) > 0 {
					u.BaseRom = string(u.BaseRom[0] - 32) // uppercase ASCII
				}
			},
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	engine := NewEngine(lang, scheme)
	result, debugInfo := engine.TransliterateDebug("ab", DefaultOptions())

	if result != "AB" {
		t.Errorf("TransliterateDebug() result = %q, want %q", result, "AB")
	}
	if len(debugInfo.Traces) != 2 {
		t.Errorf("len(DebugInfo.Traces) = %d, want 2", len(debugInfo.Traces))
	}
	if debugInfo.Traces[0].Rule != "uppercase" {
		t.Errorf("Traces[0].Rule = %q, want %q", debugInfo.Traces[0].Rule, "uppercase")
	}
}

func TestEngineWithSchemeSelectedRules(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	// Scheme selects specific rules
	selectedRules := []Rule{
		{
			Name:      "double-a",
			Phase:     PhaseVowel,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return u.BaseRom == "a" },
			Action:    func(u *Unit, w *Word) { u.BaseRom = "aa" },
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
		catalog: RuleCatalog{
			Vowel: selectedRules,
		},
	}
	scheme := &mockScheme{name: "test", rules: selectedRules}

	engine := NewEngine(lang, scheme)
	result := engine.Transliterate("ab")

	if result != "aab" {
		t.Errorf("Transliterate() = %q, want %q", result, "aab")
	}
}

// =============================================================================
// EngineOption Tests
// =============================================================================

func TestEngineWithDisabledRules(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	var rule1Ran, rule2Ran bool

	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  50,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { rule1Ran = true },
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Scope:     ScopeLanguage,
			Priority:  40,
			Mode:      ModeAlways,
			Condition: func(u *Unit, w *Word) bool { return true },
			Action:    func(u *Unit, w *Word) { rule2Ran = true },
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	// Create engine with one rule disabled
	engine := NewEngine(lang, scheme, WithDisabledRules("schwa.delete.ccv"))

	// Verify the rule is disabled
	if !engine.RuleEngine().IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be disabled")
	}
	if engine.RuleEngine().IsDisabled("schwa.delete.word-final") {
		t.Error("schwa.delete.word-final should NOT be disabled")
	}

	// Run transliteration to verify rule doesn't execute
	engine.Transliterate("a")

	if rule1Ran {
		t.Error("Disabled rule schwa.delete.ccv should not have run")
	}
	if !rule2Ran {
		t.Error("Enabled rule schwa.delete.word-final should have run")
	}
}

func TestEngineWithDisabledRulesGlob(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "schwa.delete.word-final",
			Phase:     PhaseSchwa,
			Priority:  40,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:      "consonant.va-to-wa.conjunct",
			Phase:     PhaseConsonant,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	// Create engine with all schwa.delete.* rules disabled
	engine := NewEngine(lang, scheme, WithDisabledRules("schwa.delete.*"))

	// Both schwa.delete rules should be disabled
	if !engine.RuleEngine().IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be disabled")
	}
	if !engine.RuleEngine().IsDisabled("schwa.delete.word-final") {
		t.Error("schwa.delete.word-final should be disabled")
	}
	// Consonant rule should not be disabled
	if engine.RuleEngine().IsDisabled("consonant.va-to-wa.conjunct") {
		t.Error("consonant.va-to-wa.conjunct should NOT be disabled")
	}
}

func TestEngineWithEnabledRules(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	var ruleRan bool

	rules := []Rule{
		{
			Name:            "vowel.long-aa.all",
			Phase:           PhaseVowel,
			Scope:           ScopeLanguage,
			Priority:        50,
			Mode:            ModeAlways,
			DisabledDefault: true, // Disabled by default
			Condition:       func(u *Unit, w *Word) bool { return true },
			Action:          func(u *Unit, w *Word) { ruleRan = true },
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	// Create engine with the rule enabled
	engine := NewEngine(lang, scheme, WithEnabledRules("vowel.long-aa.all"))

	// Verify the rule is enabled
	if engine.RuleEngine().IsDisabled("vowel.long-aa.all") {
		t.Error("vowel.long-aa.all should be enabled after WithEnabledRules")
	}

	// Run transliteration to verify rule executes
	engine.Transliterate("a")

	if !ruleRan {
		t.Error("Enabled rule vowel.long-aa.all should have run")
	}
}

func TestEngineWithMultipleOptions(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	rules := []Rule{
		{
			Name:      "schwa.delete.ccv",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
		{
			Name:            "vowel.long-aa.all",
			Phase:           PhaseVowel,
			Priority:        50,
			DisabledDefault: true,
			Condition:       func(*Unit, *Word) bool { return true },
			Action:          func(*Unit, *Word) {},
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	// Create engine with both options
	engine := NewEngine(lang, scheme,
		WithDisabledRules("schwa.delete.ccv"),
		WithEnabledRules("vowel.long-aa.all"),
	)

	// Verify both options took effect
	if !engine.RuleEngine().IsDisabled("schwa.delete.ccv") {
		t.Error("schwa.delete.ccv should be disabled")
	}
	if engine.RuleEngine().IsDisabled("vowel.long-aa.all") {
		t.Error("vowel.long-aa.all should be enabled")
	}
}

func TestEngineRuleEngine(t *testing.T) {
	script := &mockScript{
		name:     "test",
		parser:   &mockParser{},
		renderer: &mockRenderer{},
	}

	rules := []Rule{
		{
			Name:      "test-rule",
			Phase:     PhaseSchwa,
			Priority:  50,
			Condition: func(*Unit, *Word) bool { return true },
			Action:    func(*Unit, *Word) {},
		},
	}

	lang := &mockLanguage{
		name:    "test",
		script:  script,
		symbols: make(SymbolMap),
	}
	scheme := &mockScheme{name: "test", rules: rules}

	engine := NewEngine(lang, scheme)

	// RuleEngine() should return the internal rule engine
	re := engine.RuleEngine()
	if re == nil {
		t.Fatal("RuleEngine() returned nil")
	}

	// Should be able to manipulate rules through it
	re.DisableRule("test-rule")
	if !re.IsDisabled("test-rule") {
		t.Error("test-rule should be disabled after calling DisableRule on RuleEngine()")
	}
}
