package core

import (
	"strings"
	"testing"
)

// canonScript canonicalizes by lower-casing, so the test can see whether each
// component received the canonical spelling.
type canonScript struct{ mockScript }

func (s *canonScript) Canonicalize(input string, _ interface{}, _ SymbolMap) string {
	return strings.ToLower(input)
}

// recordingLanguage records the input its lexicon and reranker receive.
type recordingLanguage struct {
	mockLanguage
	lexiconSaw, rerankSaw []string
}

func (l *recordingLanguage) LexiconLookup(word string) (string, bool) {
	l.lexiconSaw = append(l.lexiconSaw, word)
	return "", false
}

func (l *recordingLanguage) RerankRomans(candidates []string) string {
	l.rerankSaw = append(l.rerankSaw, candidates...)
	return candidates[0]
}

// The engine canonicalizes once, before lexicon lookup, candidate ranking and
// parsing (keel K1): every component must see the canonical spelling.
func TestEngineCanonicalizesBeforeEveryComponent(t *testing.T) {
	script := &canonScript{mockScript{name: "canon", parser: &mockParser{}, renderer: &mockRenderer{}}}
	lang := &recordingLanguage{mockLanguage: mockLanguage{name: "rec", script: script, symbols: SymbolMap{}}}
	e := NewEngine(lang, &mockScheme{name: "test"})
	if got := e.TransliterateWithOptions("ABC", Options{Lexicon: true, Rerank: true}); got != "abc" {
		t.Fatalf("parser/renderer saw non-canonical input: %q", got)
	}
	// The reranker renders its candidates through the same pipeline, so the
	// lexicon is consulted once per call; every lookup must be canonical.
	if len(lang.lexiconSaw) == 0 {
		t.Error("lexicon not consulted")
	}
	for _, w := range lang.lexiconSaw {
		if w != "abc" {
			t.Errorf("lexicon saw %q, want canonical \"abc\"", lang.lexiconSaw)
		}
	}
	if len(lang.rerankSaw) != 1 || lang.rerankSaw[0] != "abc" {
		t.Errorf("reranker saw %q, want canonical \"abc\"", lang.rerankSaw)
	}
}
