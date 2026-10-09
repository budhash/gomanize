package brahmic

import (
	"sort"
	"unicode"

	"github.com/budhash/gomanize/core"
)

// CanonicalForms describes the script's Unicode canonical equivalences, so a
// script word can be brought to its NFC form without external dependencies.
// Every component that reads script text (lexicon, learned models, parser)
// sees the same canonical spelling.
type CanonicalForms struct {
	// Decompose maps composition-excluded letters to their canonical sequence
	// (e.g. क़ U+0958 -> क + nukta; NFC keeps these decomposed).
	Decompose map[rune][]rune
	// Compose maps a base + combining pair to its composed letter
	// (e.g. न + nukta -> ऩ, Bengali ে + া -> ো; NFC composes these).
	Compose map[[2]rune]rune
	// Class gives the canonical combining class of combining marks that can
	// occur in non-canonical order (nukta 7, virama 9). Unlisted runes are 0.
	Class map[rune]int
}

// Canonicalize returns the canonical spelling of a word for the script:
// format characters (Unicode Cf: ZWJ, ZWNJ, BOM, soft hyphen) adjacent to a
// script character are removed (they control rendering, not sound); composition-
// excluded letters are decomposed; combining marks are put in canonical order;
// canonical pairs are composed. Format characters not touching script text
// (e.g. the ZWJs inside an emoji sequence) are kept. Idempotent.
func Canonicalize(input string, cfg Config, symbols core.SymbolMap) string {
	in := []rune(input)
	isScript := func(r rune) bool {
		_, ok := symbols[string(r)]
		return ok
	}
	// Nearest non-format neighbors decide whether a format character sits in
	// script text. Two linear passes record them, so long runs of format
	// characters cost O(n), not O(n^2).
	prevScript := make([]bool, len(in))
	nextScript := make([]bool, len(in))
	last := false
	for i, r := range in {
		prevScript[i] = last
		if !unicode.Is(unicode.Cf, r) {
			last = isScript(r)
		}
	}
	last = false
	for i := len(in) - 1; i >= 0; i-- {
		nextScript[i] = last
		if !unicode.Is(unicode.Cf, in[i]) {
			last = isScript(in[i])
		}
	}
	out := make([]rune, 0, len(in))
	for i, r := range in {
		if unicode.Is(unicode.Cf, r) && (prevScript[i] || nextScript[i]) {
			continue
		}
		out = append(out, r)
	}
	forms := cfg.Canonical
	if forms == nil {
		return string(out)
	}
	if len(forms.Decompose) > 0 {
		expanded := out[:0:0]
		for _, r := range out {
			if seq, ok := forms.Decompose[r]; ok {
				expanded = append(expanded, seq...)
			} else {
				expanded = append(expanded, r)
			}
		}
		out = expanded
	}
	// Canonical ordering: stable sort of each maximal run of nonzero-class
	// marks (O(k log k) per run, so long mark runs cannot cost O(k^2)).
	for i := 0; i < len(out); {
		if forms.Class[out[i]] == 0 {
			i++
			continue
		}
		j := i
		for j < len(out) && forms.Class[out[j]] != 0 {
			j++
		}
		run := out[i:j]
		sort.SliceStable(run, func(a, b int) bool { return forms.Class[run[a]] < forms.Class[run[b]] })
		i = j
	}
	if len(forms.Compose) > 0 {
		composed := out[:0:0]
		for _, r := range out {
			if n := len(composed); n > 0 {
				if c, ok := forms.Compose[[2]rune{composed[n-1], r}]; ok {
					composed[n-1] = c
					continue
				}
			}
			composed = append(composed, r)
		}
		out = composed
	}
	return string(out)
}

// Canonicalize implements core.Canonicalizer for Brahmic scripts.
func (s *Script) Canonicalize(input string, config interface{}, symbols core.SymbolMap) string {
	cfg, ok := config.(Config)
	if !ok {
		return input
	}
	return Canonicalize(input, cfg, symbols)
}
