package gomanize

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/budhash/gomanize/core"
	legacyLang "github.com/budhash/gomanize/internal/legacy_lang"
	hindiLang "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// Options configures transliteration behavior.
// Use NewOptions() to get defaults, then modify as needed.
type Options = core.Options

// EngineOption configures engine creation.
type EngineOption = core.EngineOption

// WithDisabledRules returns an option that disables rules matching the given patterns.
// Patterns can be exact names or glob patterns (e.g., "schwa.*", "vowel.long-aa.*").
var WithDisabledRules = core.WithDisabledRules

// WithEnabledRules returns an option that enables rules matching the given patterns.
// Useful for enabling rules that are disabled by default.
var WithEnabledRules = core.WithEnabledRules

// RuleStatus represents a rule and its current enabled/disabled state.
type RuleStatus = core.RuleStatus

// NewOptions returns Options with default values.
func NewOptions() Options {
	return core.DefaultOptions()
}

// DebugInfo contains debugging information from transliteration.
type DebugInfo = core.DebugInfo

type Romanizer interface {
	Name() string
	Transliterate(word string) string
	TransliterateWithOptions(word string, opts Options) string
	TransliterateDebug(word string, opts Options) (string, *DebugInfo)
	Info()
}

// ExtendedRomanizer extends Romanizer with rule management capabilities.
type ExtendedRomanizer interface {
	Romanizer
	// ListRules returns all rules with their enabled/disabled status.
	// If pattern is empty, returns all rules.
	ListRules(pattern string) []RuleStatus
	// DisableRule disables rules matching the given pattern.
	DisableRule(pattern string) int
	// EnableRule enables rules matching the given pattern.
	EnableRule(pattern string) int
}

type Gomanize struct {
	romanizer Romanizer
	options   Options
}

// New creates a Gomanize instance with default options.
func New(language string) (*Gomanize, error) {
	return NewWithOptions(language, NewOptions())
}

// NewWithOptions creates a Gomanize instance with custom options.
func NewWithOptions(language string, opts Options, engineOpts ...EngineOption) (*Gomanize, error) {
	l := strings.ToLower(language)

	// Create engine with options
	var romanizer Romanizer
	switch l {
	case "hindi":
		engine := core.NewEngine(hindiLang.Hindi{}, colloquial.Colloquial{}, engineOpts...)
		romanizer = &coreEngineAdapter{name: "hindi", engine: engine}
	case "hindi-legacy":
		romanizer = &legacyAdapter{legacy: &legacyLang.Hindi{}}
	default:
		return nil, fmt.Errorf("language not supported : %s", l)
	}

	return &Gomanize{romanizer: romanizer, options: opts}, nil
}

// SetOptions updates the options for this instance.
func (g *Gomanize) SetOptions(opts Options) {
	g.options = opts
}

// GetOptions returns the current options.
func (g *Gomanize) GetOptions() Options {
	return g.options
}

func (g Gomanize) Test() {
	g.romanizer.Info()
}

// Translit transliterates text using the configured options. Words are
// segmented on ALL whitespace (spaces, tabs, newlines) and the original
// whitespace is preserved verbatim, so multi-line input romanizes line by line
// instead of leaking separators into words (which broke word-final rules).
func (g Gomanize) Translit(sentence string) string {
	var sb strings.Builder
	start := -1
	flush := func(end int) {
		if start >= 0 {
			sb.WriteString(g.romanizer.TransliterateWithOptions(sentence[start:end], g.options))
			start = -1
		}
	}
	for i, r := range sentence {
		if unicode.IsSpace(r) {
			flush(i)
			sb.WriteRune(r)
		} else if start < 0 {
			start = i
		}
	}
	flush(len(sentence))
	return sb.String()
}

// TranslitDebug transliterates a word and returns debug information.
func (g Gomanize) TranslitDebug(word string) (string, *DebugInfo) {
	return g.romanizer.TransliterateDebug(word, g.options)
}

// ListRules returns all rules with their enabled/disabled status.
// If pattern is empty, returns all rules. Supports glob patterns (e.g., "schwa.*").
// Returns nil if the romanizer doesn't support rule listing.
func (g *Gomanize) ListRules(pattern string) []RuleStatus {
	if ext, ok := g.romanizer.(ExtendedRomanizer); ok {
		return ext.ListRules(pattern)
	}
	return nil
}

// DisableRule disables rules matching the given pattern.
// Returns the number of rules disabled, or 0 if not supported.
func (g *Gomanize) DisableRule(pattern string) int {
	if ext, ok := g.romanizer.(ExtendedRomanizer); ok {
		return ext.DisableRule(pattern)
	}
	return 0
}

// EnableRule enables rules matching the given pattern.
// Returns the number of rules enabled, or 0 if not supported.
func (g *Gomanize) EnableRule(pattern string) int {
	if ext, ok := g.romanizer.(ExtendedRomanizer); ok {
		return ext.EnableRule(pattern)
	}
	return 0
}

// coreEngineAdapter adapts a core.Engine to the Romanizer interface.
type coreEngineAdapter struct {
	name   string
	engine *core.Engine
}

func (a *coreEngineAdapter) Name() string {
	return a.name
}

func (a *coreEngineAdapter) Transliterate(word string) string {
	return a.engine.Transliterate(word)
}

func (a *coreEngineAdapter) TransliterateWithOptions(word string, opts Options) string {
	return a.engine.TransliterateWithOptions(word, opts)
}

func (a *coreEngineAdapter) TransliterateDebug(word string, opts Options) (string, *DebugInfo) {
	return a.engine.TransliterateDebug(word, opts)
}

func (a *coreEngineAdapter) Info() {
	fmt.Printf("Romanizer: %s (new architecture)\n", a.name)
	fmt.Printf("Language: %s\n", a.engine.Language().Name())
	fmt.Printf("Scheme: %s\n", a.engine.Scheme().Name())
}

func (a *coreEngineAdapter) ListRules(pattern string) []RuleStatus {
	return a.engine.RuleEngine().ListRules(pattern)
}

func (a *coreEngineAdapter) DisableRule(pattern string) int {
	return a.engine.RuleEngine().DisableRule(pattern)
}

func (a *coreEngineAdapter) EnableRule(pattern string) int {
	return a.engine.RuleEngine().EnableRule(pattern)
}

// legacyAdapter adapts the legacy engine to the Romanizer interface.
type legacyAdapter struct {
	legacy *legacyLang.Hindi
}

func (a *legacyAdapter) Name() string {
	return "hindi-legacy"
}

func (a *legacyAdapter) Transliterate(word string) string {
	return a.legacy.Transliterate(word)
}

func (a *legacyAdapter) TransliterateWithOptions(word string, opts Options) string {
	legacyOpts := legacyLang.Options{LongVowels: opts.LongVowels}
	return a.legacy.TransliterateWithOptions(word, legacyOpts)
}

func (a *legacyAdapter) TransliterateDebug(word string, opts Options) (string, *DebugInfo) {
	// Legacy adapter doesn't support debug, return nil debug info
	result := a.TransliterateWithOptions(word, opts)
	return result, nil
}

func (a *legacyAdapter) Info() {
	a.legacy.Info()
}
