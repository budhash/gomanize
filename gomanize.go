package gomanize

import (
	"fmt"
	"strings"

	"github.com/budhash/gomanize/internal/lang"
)

type Romanizer interface {
	Name() string
	Transliterate(word string) string
	Info()
}

type Gomanize struct {
	romanizer Romanizer
}

func New(language string) (*Gomanize, error) {
	converters := loadRomanizers()
	lang := strings.ToLower(language)
	g, found := converters[lang]
	if found {
		return &Gomanize{romanizer: g}, nil
	} else {
		return nil, fmt.Errorf("language not supported : %s", lang)
	}
}

func (g Gomanize) Test() {
	g.romanizer.Info()
}

func (g Gomanize) Translit(sentence string) string {
	words := strings.Split(sentence, " ")
	var result []string
	var converter = g.romanizer
	for _, word := range words {
		translated := converter.Transliterate(word)
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
