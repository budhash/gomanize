package gomanize

import "testing"

func TestTranslitWhitespace(t *testing.T) {
	g, err := New("hindi")
	if err != nil {
		t.Fatalf("New(hindi): %v", err)
	}

	cases := []struct{ in, want string }{
		{"नमस्ते भारत", "namaste bharat"},
		// Newlines are word boundaries and are preserved verbatim (T-0023):
		// previously "भारत\nभारत" was treated as ONE word, defeating word-final
		// rules on the first भारत.
		{"भारत\nभारत", "bharat\nbharat"},
		{"भारत\tभारत", "bharat\tbharat"},
		// Multiple spaces preserved exactly.
		{"नमस्ते  भारत", "namaste  bharat"},
		// Leading/trailing whitespace preserved.
		{"\nभारत ", "\nbharat "},
		{"", ""},
	}
	for _, c := range cases {
		if got := g.Translit(c.in); got != c.want {
			t.Errorf("Translit(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestTranslitWordFinalAcrossLines is the sanity-revert guard for T-0023: the
// word before a newline must get word-final schwa deletion, identical to the
// same word alone.
func TestTranslitWordFinalAcrossLines(t *testing.T) {
	g, err := New("hindi")
	if err != nil {
		t.Fatalf("New(hindi): %v", err)
	}
	alone := g.Translit("अनजान")
	multi := g.Translit("अनजान\nअनजान")
	if want := alone + "\n" + alone; multi != want {
		t.Errorf("multi-line = %q, want %q", multi, want)
	}
}

func TestPublicAPISurface(t *testing.T) {
	g, err := New("hindi")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Unsupported language errors.
	if _, err := New("klingon"); err == nil {
		t.Error("New(klingon) should error")
	}

	// Options round-trip.
	opts := NewOptions()
	opts.KeepMedialSchwa = true
	g.SetOptions(opts)
	if !g.GetOptions().KeepMedialSchwa {
		t.Error("SetOptions/GetOptions did not round-trip")
	}
	if g.Translit("जनता") != "janata" {
		t.Errorf("options not applied through Translit: got %q", g.Translit("जनता"))
	}
	g.SetOptions(NewOptions())

	// TranslitDebug returns output and traces for a plain word.
	out, dbg := g.TranslitDebug("नमस्ते")
	if out != "namaste" {
		t.Errorf("TranslitDebug output = %q, want namaste", out)
	}
	if dbg == nil || len(dbg.Units) == 0 {
		t.Error("TranslitDebug returned no debug info for plain word")
	}

	// Rule management: listing, disabling a real rule changes output, re-enabling restores.
	rules := g.ListRules("")
	if len(rules) == 0 {
		t.Fatal("ListRules returned no rules")
	}
	if n := g.ListRules("schwa.*"); len(n) == 0 || len(n) >= len(rules) {
		t.Errorf("ListRules(schwa.*) returned %d of %d — pattern filter broken", len(n), len(rules))
	}
	before := g.Translit("जनता") // janta via schwa.delete.ccv
	if got := g.DisableRule("schwa.delete.ccv"); got != 1 {
		t.Fatalf("DisableRule matched %d rules, want 1", got)
	}
	if after := g.Translit("जनता"); after == before {
		t.Errorf("disabling schwa.delete.ccv had no effect (still %q)", after)
	}
	if got := g.EnableRule("schwa.delete.ccv"); got != 1 {
		t.Fatalf("EnableRule matched %d rules, want 1", got)
	}
	if restored := g.Translit("जनता"); restored != before {
		t.Errorf("re-enabling did not restore output: %q vs %q", restored, before)
	}
	// Garbage pattern matches nothing.
	if got := g.DisableRule("no.such.rule.*"); got != 0 {
		t.Errorf("DisableRule(garbage) matched %d rules, want 0", got)
	}
}

func TestNewWithOptionsAndEngineOpts(t *testing.T) {
	// Engine-option plumbing: constructing with a disabled rule changes output.
	g, err := NewWithOptions("hindi", NewOptions(), WithDisabledRules("schwa.delete.ccv"))
	if err != nil {
		t.Fatalf("NewWithOptions: %v", err)
	}
	if got := g.Translit("जनता"); got == "janta" {
		t.Errorf("WithDisabledRules had no effect: got %q", got)
	}
}
