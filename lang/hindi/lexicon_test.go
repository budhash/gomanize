package hindi_test

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestLexiconLookup(t *testing.T) {
	h := hindi.Hindi{}
	// Known loanword: lexicon has the conventional English spelling.
	if got, ok := h.LexiconLookup("अंकल"); !ok || got != "uncle" {
		t.Errorf("LexiconLookup(अंकल) = %q, %v; want uncle, true", got, ok)
	}
	// Out-of-vocabulary: not found.
	if got, ok := h.LexiconLookup("क्षमाशीलता"); ok {
		t.Errorf("LexiconLookup(OOV) = %q, %v; want _, false", got, ok)
	}
	if hindi.LexiconSize() < 1000 {
		t.Errorf("lexicon size = %d, want >= 1000", hindi.LexiconSize())
	}
}

func TestLexiconOptionRouting(t *testing.T) {
	e := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})

	// With the lexicon, a known loanword returns the attested spelling.
	if got := e.TransliterateWithOptions("अंकल", core.Options{Lexicon: true}); got != "uncle" {
		t.Errorf("with lexicon: अंकल = %q, want uncle", got)
	}
	// OOV must be byte-identical to the pure rule output (lossless fallthrough).
	oov := "क्षमाशीलता"
	rules := e.Transliterate(oov)
	lex := e.TransliterateWithOptions(oov, core.Options{Lexicon: true})
	if rules != lex {
		t.Errorf("OOV fallthrough not lossless: rules=%q lexicon=%q", rules, lex)
	}
	// Default (no option) must ignore the lexicon entirely.
	if def := e.Transliterate("अंकल"); def == "uncle" {
		t.Errorf("lexicon leaked into default path: अंकल=%q", def)
	}
}
