package benchmark

import (
	"math"
	"testing"
)

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"gana", "gaana", 1},
		{"janta", "janata", 1},
	}
	for _, c := range cases {
		if got := levenshtein([]rune(c.a), []rune(c.b)); got != c.want {
			t.Errorf("levenshtein(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCER(t *testing.T) {
	cases := []struct {
		got, exp string
		want     float64
	}{
		{"abc", "abc", 0.0},
		{"gana", "gaana", 1.0 / 5.0}, // one insertion over 5-char expected
		{"janta", "janata", 1.0 / 6.0},
		{"", "abc", 1.0},
	}
	for _, c := range cases {
		if got := cer(c.got, c.exp); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("cer(%q,%q) = %f, want %f", c.got, c.exp, got, c.want)
		}
	}
}

func TestMinCERAndMatchAny(t *testing.T) {
	refs := []string{"janata", "janataa", "janta"}
	// "janta" is one of the attested variants -> exact match, CER 0.
	if !matchesAny("janta", refs) {
		t.Errorf("matchesAny(janta) = false, want true")
	}
	if got := minCER("janta", refs); got != 0.0 {
		t.Errorf("minCER(janta) = %f, want 0.0", got)
	}
	// A non-attested spelling should not match but should still find nearest CER.
	if matchesAny("jantaa", refs) {
		t.Errorf("matchesAny(jantaa) = true, want false")
	}
	if got := minCER("jantaa", refs); got <= 0 || got > 1 {
		t.Errorf("minCER(jantaa) = %f, want in (0,1]", got)
	}
	// Empty reference set = full miss.
	if got := minCER("x", nil); got != 1.0 {
		t.Errorf("minCER with no refs = %f, want 1.0", got)
	}
}
