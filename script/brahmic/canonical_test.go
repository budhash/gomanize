package brahmic

import (
	"testing"

	"github.com/budhash/gomanize/core"
)

func TestAlreadyCanonicalFastPath(t *testing.T) {
	forms := &CanonicalForms{
		Decompose: map[rune][]rune{0x0958: {0x0915, 0x093C}},
		Compose:   map[[2]rune]rune{{0x0928, 0x093C}: 0x0929},
		Class:     map[rune]int{0x093C: 7, 0x094D: 9},
	}
	symbols := core.SymbolMap{"क": {}, "न": {}}
	for input, want := range map[string]bool{
		"क्क": true, "क़्क": true, "नमस्ते": true, "": true,
		"क़": false, "ऩ": false, "क़्": false, "क‍क": false,
	} {
		if got := alreadyCanonical(input, forms); got != want {
			t.Errorf("alreadyCanonical(%q) = %v, want %v", input, got, want)
		}
		if want && Canonicalize(input, Config{Canonical: forms}, symbols) != input {
			t.Errorf("fast path disagrees with Canonicalize for %q", input)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { Canonicalize("नमस्ते", Config{Canonical: forms}, symbols) }); allocs != 0 {
		t.Errorf("canonical input allocated %v times", allocs)
	}
}
