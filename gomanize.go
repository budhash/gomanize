package gomanize

import (
	"fmt"
	"strings"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/internal/lang"
	hindiLang "github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

// Options configures transliteration behavior.
// Use NewOptions() to get defaults, then modify as needed.
type Options = core.Options

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

type Gomanize struct {
	romanizer Romanizer
	options   Options
}

// New creates a Gomanize instance with default options.
func New(language string) (*Gomanize, error) {
	return NewWithOptions(language, NewOptions())
}

// NewWithOptions creates a Gomanize instance with custom options.
func NewWithOptions(language string, opts Options) (*Gomanize, error) {
	converters := loadRomanizers()
	l := strings.ToLower(language)
	g, found := converters[l]
	if found {
		return &Gomanize{romanizer: g, options: opts}, nil
	} else {
		return nil, fmt.Errorf("language not supported : %s", l)
	}
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

// Translit transliterates a sentence using the configured options.
func (g Gomanize) Translit(sentence string) string {
	words := strings.Split(sentence, " ")
	var result []string
	var converter = g.romanizer
	for _, word := range words {
		translated := converter.TransliterateWithOptions(word, g.options)
		result = append(result, translated)
	}
	return strings.Join(result, " ")
}

// TranslitDebug transliterates a word and returns debug information.
func (g Gomanize) TranslitDebug(word string) (string, *DebugInfo) {
	return g.romanizer.TransliterateDebug(word, g.options)
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

// legacyOptionsAdapter adapts core.Options to legacy lang.Options.
type legacyAdapter struct {
	legacy *lang.Hindi
}

func (a *legacyAdapter) Name() string {
	return "hindi-legacy"
}

func (a *legacyAdapter) Transliterate(word string) string {
	return a.legacy.Transliterate(word)
}

func (a *legacyAdapter) TransliterateWithOptions(word string, opts Options) string {
	// Convert core.Options to lang.Options
	legacyOpts := lang.Options{LongVowels: opts.LongVowels}
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

// hindiOrigAdapter adapts HindiOrig to use core.Options.
type hindiOrigAdapter struct {
	legacy *lang.HindiOrig
}

func (a *hindiOrigAdapter) Name() string {
	return a.legacy.Name()
}

func (a *hindiOrigAdapter) Transliterate(word string) string {
	return a.legacy.Transliterate(word)
}

func (a *hindiOrigAdapter) TransliterateWithOptions(word string, opts Options) string {
	// HindiOrig doesn't support LongVowels well, just use plain transliterate
	legacyOpts := lang.Options{LongVowels: opts.LongVowels}
	return a.legacy.TransliterateWithOptions(word, legacyOpts)
}

func (a *hindiOrigAdapter) TransliterateDebug(word string, opts Options) (string, *DebugInfo) {
	// HindiOrig adapter doesn't support debug, return nil debug info
	result := a.TransliterateWithOptions(word, opts)
	return result, nil
}

func (a *hindiOrigAdapter) Info() {
	a.legacy.Info()
}

func loadRomanizers() map[string]Romanizer {
	romanizers := make(map[string]Romanizer)

	// New architecture: Hindi with colloquial scheme
	hindiEngine := core.NewEngine(hindiLang.Hindi{}, colloquial.Colloquial{})
	romanizers["hindi"] = &coreEngineAdapter{
		name:   "hindi",
		engine: hindiEngine,
	}

	// Legacy implementation (for comparison/reference)
	romanizers["hindi-legacy"] = &legacyAdapter{legacy: &lang.Hindi{}}

	// Original legacy implementation
	romanizers["hindiorig"] = &hindiOrigAdapter{legacy: &lang.HindiOrig{}}

	return romanizers
}
