package brahmic_test

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/script/brahmic"
)

// Direct structural tests for the Brahmic parser/preparer (previously untested).
// Uses the Hindi symbol map + config as the concrete Brahmic instance.

func parse(t *testing.T, input string) *core.Word {
	t.Helper()
	lang := hindi.Hindi{}
	s := brahmic.New()
	p := s.NewParser(lang.ScriptConfig())
	w := p.Parse(input, lang.Symbols())
	s.PrepareWord(w)
	return w
}

func baseRoms(w *core.Word) []string {
	out := make([]string, len(w.Units))
	for i, u := range w.Units {
		out[i] = u.BaseRom
	}
	return out
}

func TestParseBasicUnits(t *testing.T) {
	// काम = क + ा(matra) + म : 3 units.
	w := parse(t, "काम")
	if len(w.Units) != 3 {
		t.Fatalf("काम: got %d units %v, want 3", len(w.Units), baseRoms(w))
	}
	if w.Units[0].BaseRom != "k" || w.Units[2].BaseRom != "m" {
		t.Errorf("काम: base romanizations %v, want first k / last m", baseRoms(w))
	}
}

func TestHalantConsumedNotEmitted(t *testing.T) {
	// नमस्ते: the halant (्) must NOT produce its own unit; the following
	// consonant (त) must be flagged as after-halant (part of the स्त conjunct).
	w := parse(t, "नमस्ते")

	for _, u := range w.Units {
		for _, r := range u.Runes {
			if r == '्' {
				t.Fatalf("halant leaked into a unit: %v", baseRoms(w))
			}
		}
	}
	afterHalant := 0
	for _, u := range w.Units {
		if brahmic.IsAfterHalant(u) {
			afterHalant++
		}
	}
	if afterHalant != 1 {
		t.Errorf("नमस्ते: after-halant units = %d, want 1 (%v)", afterHalant, baseRoms(w))
	}
}

func TestNuktaCombines(t *testing.T) {
	// ज़ (ja + nukta) should combine into a single unit, not two.
	w := parse(t, "ज़")
	if len(w.Units) != 1 {
		t.Errorf("ज़: got %d units %v, want 1 (nukta should combine)", len(w.Units), baseRoms(w))
	}
}

func TestNumberParsed(t *testing.T) {
	w := parse(t, "९")
	if len(w.Units) != 1 || w.Units[0].BaseRom != "9" {
		t.Errorf("९: got %v, want single unit '9'", baseRoms(w))
	}
}

func TestIdentifyRunsProducesRun(t *testing.T) {
	// जनता has a consonant run between vowels; at least one unit should belong
	// to a run after PrepareWord.
	w := parse(t, "जनता")
	hasRun := false
	for _, u := range w.Units {
		if brahmic.GetRun(u) != nil {
			hasRun = true
			break
		}
	}
	if !hasRun {
		t.Errorf("जनता: expected at least one unit assigned to a consonant run")
	}
}

func TestParseLinksPrevNext(t *testing.T) {
	// Doubly-linked list integrity: first has no Prev, last no Next, and the
	// chain length matches Units.
	w := parse(t, "भारत")
	if len(w.Units) == 0 {
		t.Fatal("भारत produced no units")
	}
	if w.Units[0].Prev != nil {
		t.Errorf("first unit should have nil Prev")
	}
	last := w.Units[len(w.Units)-1]
	if last.Next != nil {
		t.Errorf("last unit should have nil Next")
	}
	n := 0
	for u := w.Units[0]; u != nil; u = u.Next {
		n++
	}
	if n != len(w.Units) {
		t.Errorf("linked chain length %d != Units length %d", n, len(w.Units))
	}
}

func TestSchwaRulesShared(t *testing.T) {
	rules := brahmic.SchwaRules()
	if len(rules) != 6 {
		t.Fatalf("SchwaRules() returned %d rules, want 6", len(rules))
	}
	want := map[string]bool{
		"schwa.keep.sonorous-final": true,
		"schwa.delete.ccv":          true,
		"schwa.delete.cccc-final":   true,
		"schwa.delete.before-cc":    true,
		"schwa.delete.word-final":   true,
		"schwa.keep.default":        true,
	}
	for _, r := range rules {
		if !want[r.Name] {
			t.Errorf("unexpected shared rule %q", r.Name)
		}
		delete(want, r.Name)
	}
	for name := range want {
		t.Errorf("missing shared rule %q", name)
	}
}
