package core

import "testing"

type legacyTestLexicon struct {
	*mockLanguage
	calls int
}

func (l *legacyTestLexicon) LexiconLookup(string) (string, bool) { l.calls++; return "legacy", true }

type optionsTestLexicon struct {
	*legacyTestLexicon
	seen []Options
}

func (l *optionsTestLexicon) LexiconLookupWithOptions(_ string, opts Options) (string, bool) {
	l.seen = append(l.seen, opts)
	if opts.InherentVowelA {
		return "", false
	}
	return "styled", true
}

func TestOptionsLexiconPrecedenceAndLegacyCompatibility(t *testing.T) {
	language := &mockLanguage{script: &mockScript{parser: &mockParser{}, renderer: &mockRenderer{}}}
	legacy := &legacyTestLexicon{mockLanguage: language}
	e := NewEngine(legacy, &mockScheme{})
	if got := e.TransliterateWithOptions("word", Options{Lexicon: true, InherentVowelA: true}); got != "legacy" {
		t.Fatal(got)
	}
	if legacy.calls != 1 {
		t.Fatal("legacy provider was not used")
	}
	styled := &optionsTestLexicon{legacyTestLexicon: legacy}
	e = NewEngine(styled, &mockScheme{})
	opts := Options{Lexicon: true, SchwaModel: true, Debug: true}
	if got := e.TransliterateWithOptions("word", opts); got != "styled" {
		t.Fatal(got)
	}
	if len(styled.seen) != 1 || styled.seen[0] != opts {
		t.Fatal("options not forwarded")
	}
	if got := e.TransliterateWithOptions("word", Options{Lexicon: true, InherentVowelA: true}); got != "word" {
		t.Fatalf("style miss fell through to legacy lookup: %s", got)
	}
	if legacy.calls != 1 {
		t.Fatal("options-aware miss invoked legacy lookup")
	}
	calls := len(styled.seen)
	if got := e.Transliterate("word"); got != "word" || len(styled.seen) != calls {
		t.Fatal("lexicon consulted while disabled")
	}
}
