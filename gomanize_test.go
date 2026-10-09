package gomanize

import (
	"strings"
	"testing"
	"time"
)

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

func TestBengaliPublicAPI(t *testing.T) {
	g, err := New("Bengali")
	if err != nil {
		t.Fatal(err)
	}
	if got := g.Translit("আমি\tবাংলা\nমন। ১২!"); got != "ami\tbangla\nmon। 12!" {
		t.Fatal(got)
	}
	if g.DisableRule("schwa.delete.word-final") != 1 {
		t.Fatal("missing B0 final-deletion rule")
	}
	if got := g.Translit("মন"); got != "mono" {
		t.Fatal(got)
	}
	if got, info := g.TranslitDebug("ৎ"); got != "t" || info == nil || !strings.Contains(info.Units[0].Metadata, "no-inherent-vowel") {
		t.Fatalf("Bengali debug: %q, %+v", got, info)
	}
}

// Text without any character of the selected script must pass through
// byte-for-byte; parsing strips format characters such as the ZWJ in emoji
// sequences, flag tag characters, soft hyphens and RTL marks.
func TestNonScriptTextPassesThrough(t *testing.T) {
	for _, language := range []string{"hindi", "bengali"} {
		g, err := New(language)
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range []string{
			"👨‍👩‍👧", "🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f",
			"abc­xyz", "‏שלום", "x‌y", "Latin 123",
		} {
			if got := g.Translit(input); got != input {
				t.Errorf("%s: %q -> %q", language, input, got)
			}
		}
	}
}

// Debug tracing must not mutate shared engine state: concurrent Translit with
// Options.Debug, TranslitDebug and plain Translit on one instance must be
// race-free and match serial output (run with -race).
func TestConcurrentDebugCalls(t *testing.T) {
	for _, tc := range []struct{ language, word string }{{"hindi", "नमस्ते"}, {"bengali", "সোনার"}} {
		debug, err := NewWithOptions(tc.language, Options{Debug: true, SchwaModel: true})
		if err != nil {
			t.Fatal(err)
		}
		plain, _ := NewWithOptions(tc.language, Options{SchwaModel: true})
		want := plain.Translit(tc.word)
		wantTraces := 0
		if _, info := plain.TranslitDebug(tc.word); info != nil {
			wantTraces = len(info.Traces)
		}
		done := make(chan string, 48)
		for i := 0; i < 16; i++ {
			go func() { done <- debug.Translit(tc.word) }()
			go func() { done <- plain.Translit(tc.word) }()
			go func() {
				out, info := plain.TranslitDebug(tc.word)
				if info != nil && len(info.Traces) != wantTraces {
					out = "trace count changed"
				}
				done <- out
			}()
		}
		for i := 0; i < 48; i++ {
			if got := <-done; got != want {
				t.Fatalf("%s: concurrent output %q, want %q", tc.language, got, want)
			}
		}
	}
}

// Learned per-word components are bounded (core.MaxLearnedWordRunes), so a
// pathological unspaced token cannot cost quadratic time. Before the bound, a
// 2,000-rune Bengali word with the vowel model and selector took ~30 s in WASM.
func TestLongTokenCostIsBounded(t *testing.T) {
	for _, tc := range []struct {
		language, unit string
		opts           Options
	}{
		{"bengali", "কলকাতা", Options{SchwaModel: true, Rerank: true, Lexicon: true}},
		{"hindi", "नमस्ते", Options{SchwaModel: true, Rerank: true, Lexicon: true}},
	} {
		g, err := NewWithOptions(tc.language, tc.opts)
		if err != nil {
			t.Fatal(err)
		}
		word := strings.Repeat(tc.unit, 2000) // ~12,000 runes, no spaces
		start := time.Now()
		out := g.Translit(word)
		// Generous ceiling for race/coverage CI runners (~5 s there, 0.1 s
		// locally); the unbounded quadratic path took over 5 minutes.
		if elapsed := time.Since(start); elapsed > 30*time.Second {
			t.Errorf("%s: %d-rune token took %v", tc.language, len([]rune(word)), elapsed)
		}
		if out == "" {
			t.Errorf("%s: empty output", tc.language)
		}
	}
}

// Canonically equivalent spellings must give identical output for every
// profile: precomposed vs decomposed nukta letters, format characters inside
// script text, and nukta/virama order (core canonicalization, keel K1).
func TestCanonicalEquivalenceAllProfiles(t *testing.T) {
	profiles := []Options{{}, {SchwaModel: true}, {Lexicon: true}, {Rerank: true},
		{SchwaModel: true, Lexicon: true, Rerank: true}, {LongVowels: true, SimpleNasals: true, KeepMedialSchwa: true}}
	cases := map[string][][2]string{
		"hindi": {
			{"पढ़ाई", "पढ़ाई"}, {"अरोड़ा", "अरोड़ा"}, {"ज़्यादा", "ज़्यादा"},
			{"अक्टूबर", "अक्‍टूबर"}, {"मकसद", "‌मकसद"}, {"उज़्ज़ा", "उज़्ज़्ा"},
			{"ऩ", "ऩ"},
		},
		"bengali": {
			{"রয়্যালটি", "রয়্যালটি"}, {"বড়", "বড়"}, {"কোথায়", "কোথায়"},
			{"ওয়্যার", "ওয়্যার"}, {"নমস্কার", "‍নমস্কার"},
		},
	}
	for language, pairs := range cases {
		for _, opts := range profiles {
			g, err := NewWithOptions(language, opts)
			if err != nil {
				t.Fatal(err)
			}
			for _, pair := range pairs {
				if a, b := g.Translit(pair[0]), g.Translit(pair[1]); a != b {
					t.Errorf("%s %+v: %q -> %q but %q -> %q", language, opts, pair[0], a, pair[1], b)
				}
			}
		}
	}
	// A token mixing an emoji sequence and script text keeps the emoji's ZWJ.
	g, _ := New("hindi")
	if got := g.Translit("👨‍👩नमस्ते"); got != "👨‍👩namaste" {
		t.Errorf("mixed token: %q", got)
	}
}
