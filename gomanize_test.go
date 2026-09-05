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
