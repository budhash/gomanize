package hindi_test

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestRerankRomans(t *testing.T) {
	h := hindi.Hindi{}
	// Real romanized Hindi must outscore keyboard mash.
	if got := h.RerankRomans([]string{"xqzt", "namaste"}); got != "namaste" {
		t.Errorf("RerankRomans preferred %q over namaste", got)
	}
	// Single candidate returned as-is; empty slice is safe.
	if got := h.RerankRomans([]string{"bharat"}); got != "bharat" {
		t.Errorf("single candidate: got %q", got)
	}
	if got := h.RerankRomans(nil); got != "" {
		t.Errorf("nil candidates: got %q", got)
	}
	// Identical candidates: first kept (ties must not override the default).
	if got := h.RerankRomans([]string{"janta", "janta"}); got != "janta" {
		t.Errorf("tie: got %q", got)
	}
}

func TestRerankOptionRouting(t *testing.T) {
	e := core.NewEngine(hindi.Hindi{}, colloquial.Colloquial{})
	// Rerank output must always be one of the candidate systems' outputs.
	for _, w := range []string{"जनता", "नमस्ते", "भारत", "कहते"} {
		rules := e.Transliterate(w)
		model := e.TransliterateWithOptions(w, core.Options{SchwaModel: true})
		got := e.TransliterateWithOptions(w, core.Options{Rerank: true})
		if got != rules && got != model {
			t.Errorf("Rerank(%s) = %q, not among candidates {%q, %q}", w, got, rules, model)
		}
	}
}
