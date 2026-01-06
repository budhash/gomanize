package gomanize

import (
	"fmt"
	"strings"

	"github.com/budhash/gomanize/internal/lang"
)

// Options configures transliteration behavior.
// Use NewOptions() to get defaults, then modify as needed.
type Options = lang.Options

// NewOptions returns Options with default values.
func NewOptions() Options {
	return lang.DefaultOptions()
}

type Romanizer interface {
	Name() string
	Transliterate(word string) string
	TransliterateWithOptions(word string, opts Options) string
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

func loadRomanizers() map[string]Romanizer {
	romanizers := make(map[string]Romanizer)
	var r Romanizer
	r = &lang.Hindi{}
	romanizers[strings.ToLower(r.Name())] = r
	r = &lang.HindiOrig{}
	romanizers[strings.ToLower(r.Name())] = r
	return romanizers
}
