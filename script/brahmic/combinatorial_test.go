package brahmic_test

// Tier 3 of the structured parser-QA plan
// (docs/reviews/2026-09-08-structured-parser-qa.md): combinatorial construct
// enumeration. Tier 2 (invariants_test.go, package gomanize) samples a
// representative handful of consonants/vowels; this tier machine-generates the
// FULL cross-product from the language's own symbol map and asserts each parse
// is well-formed — correct unit count and types. It turns "hope the datasets
// cover it" into "every construct is provably exercised", with no manual
// sifting and no romanization-convention assumptions (structural checks only).
//
// The enumeration is driven off hindi.Hindi{}.Symbols() so it automatically
// covers new symbols as they are added — nothing here hardcodes the inventory.

import (
	"sort"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/hindi"
	"github.com/budhash/gomanize/script/brahmic"
)

// symbolInventory pulls the concrete characters of each category out of the
// Hindi symbol map, so the cross-product tracks the language definition.
type symbolInventory struct {
	consonants []string // CatConsonant, single (no nukta precomposed)
	matras     []string // CatMatra
	indepVowel []string // CatVowel (independent)
	anusvara   string
	chandra    string
	visarga    string
	halant     string
	nukta      string
}

func buildInventory(t *testing.T) symbolInventory {
	t.Helper()
	inv := symbolInventory{}
	for ch, info := range (hindi.Hindi{}).Symbols() {
		switch info.Category {
		case core.CatConsonant:
			// only single-rune base consonants; precomposed nukta consonants
			// (क़, ज़, …) are exercised via the +nukta axis below.
			if len([]rune(ch)) == 1 {
				inv.consonants = append(inv.consonants, ch)
			}
		case core.CatVowel:
			inv.indepVowel = append(inv.indepVowel, ch)
		case brahmic.CatMatra:
			inv.matras = append(inv.matras, ch)
		case brahmic.CatAnusvara:
			inv.anusvara = ch
		case brahmic.CatChandrabindu:
			inv.chandra = ch
		case brahmic.CatVisarga:
			inv.visarga = ch
		case brahmic.CatHalant:
			inv.halant = ch
		case brahmic.CatNukta:
			inv.nukta = ch
		}
	}
	// Deterministic order (map iteration is random) so failures are stable.
	sort.Strings(inv.consonants)
	sort.Strings(inv.matras)
	sort.Strings(inv.indepVowel)
	if len(inv.consonants) == 0 || len(inv.matras) == 0 || len(inv.indepVowel) == 0 {
		t.Fatalf("inventory incomplete: %d consonants, %d matras, %d indep vowels",
			len(inv.consonants), len(inv.matras), len(inv.indepVowel))
	}
	return inv
}

func parseWord(input string) *core.Word {
	lang := hindi.Hindi{}
	s := brahmic.New()
	p := s.NewParser(lang.ScriptConfig())
	w := p.Parse(input, lang.Symbols())
	s.PrepareWord(w)
	return w
}

// assertWellFormed checks structural invariants true of every parse, regardless
// of construct: no halant leaks into a unit, rune positions are monotonic and
// non-overlapping, and no zero-width units. Positions may have gaps — a halant
// is consumed without emitting a unit (TestHalantConsumedNotEmitted), so the
// second consonant of a conjunct legitimately starts one rune past the first
// unit's end — so this asserts Start >= prevEnd, not strict contiguity.
func assertWellFormed(t *testing.T, input string, w *core.Word, halant string) {
	t.Helper()
	hr := []rune(halant)[0]
	prevEnd := 0
	for i, u := range w.Units {
		if len(u.Runes) == 0 {
			t.Errorf("%q unit[%d] has no runes", input, i)
		}
		for _, r := range u.Runes {
			if r == hr {
				t.Errorf("%q: halant leaked into unit[%d]", input, i)
			}
		}
		if u.Start.Rune < prevEnd {
			t.Errorf("%q unit[%d] starts at rune %d, overlaps previous end %d",
				input, i, u.Start.Rune, prevEnd)
		}
		if u.End.Rune <= u.Start.Rune {
			t.Errorf("%q unit[%d] end %d <= start %d", input, i, u.End.Rune, u.Start.Rune)
		}
		prevEnd = u.End.Rune
	}
}

// TestCombinatorialParse enumerates the full construct cross-product and asserts
// each parse has the expected unit count and types.
func TestCombinatorialParse(t *testing.T) {
	inv := buildInventory(t)

	// A representative second consonant for conjuncts (needs a consonant that
	// is not the same rune, but any works structurally).
	c2 := inv.consonants[0]

	cases := 0

	for _, c := range inv.consonants {
		// (1) bare consonant: 1 unit, consonant
		w := parseWord(c)
		assertWellFormed(t, c, w, inv.halant)
		if len(w.Units) != 1 || w.Units[0].Type != core.UnitConsonant {
			t.Errorf("%q: got %d units (types %v), want 1 consonant", c, len(w.Units), unitTypes(w))
		}
		cases++

		// (2) consonant + each matra: 2 units [consonant, vowel]; the vowel
		//     unit must be flagged IsMatra.
		for _, m := range inv.matras {
			in := c + m
			w := parseWord(in)
			assertWellFormed(t, in, w, inv.halant)
			if len(w.Units) != 2 {
				t.Errorf("%q: got %d units %v, want 2 (C+matra)", in, len(w.Units), unitTypes(w))
				continue
			}
			if w.Units[0].Type != core.UnitConsonant || w.Units[1].Type != core.UnitVowel {
				t.Errorf("%q: types %v, want [consonant vowel]", in, unitTypes(w))
			}
			if !brahmic.IsMatraUnit(w.Units[1]) {
				t.Errorf("%q: matra vowel not flagged IsMatra", in)
			}
			cases++
		}

		// (3) consonant + each INDEPENDENT vowel: 2 units [consonant, vowel];
		//     the vowel starts its own syllable, so it must NOT be a matra
		//     (this is the T-0040 गई-class bug, guarded structurally here).
		for _, v := range inv.indepVowel {
			in := c + v
			w := parseWord(in)
			assertWellFormed(t, in, w, inv.halant)
			if len(w.Units) != 2 {
				t.Errorf("%q: got %d units %v, want 2 (C+indep-vowel)", in, len(w.Units), unitTypes(w))
				continue
			}
			if w.Units[0].Type != core.UnitConsonant || w.Units[1].Type != core.UnitVowel {
				t.Errorf("%q: types %v, want [consonant vowel]", in, unitTypes(w))
			}
			if brahmic.IsMatraUnit(w.Units[1]) {
				t.Errorf("%q: independent vowel wrongly flagged IsMatra (matra/independent collapse)", in)
			}
			cases++
		}

		// (4) halant conjunct C + halant + C2: 2 units, the second after-halant.
		in := c + inv.halant + c2
		w = parseWord(in)
		assertWellFormed(t, in, w, inv.halant)
		if len(w.Units) != 2 {
			t.Errorf("%q: got %d units %v, want 2 (conjunct)", in, len(w.Units), unitTypes(w))
		} else if !brahmic.IsAfterHalant(w.Units[1]) {
			t.Errorf("%q: second unit not flagged after-halant", in)
		}
		cases++

		// (5) consonant + matra + nasal (anusvara / chandrabindu): 3 units,
		//     last is a modifier.
		for _, nasal := range []string{inv.anusvara, inv.chandra, inv.visarga} {
			in := c + inv.matras[0] + nasal
			w := parseWord(in)
			assertWellFormed(t, in, w, inv.halant)
			if len(w.Units) != 3 {
				t.Errorf("%q: got %d units %v, want 3 (C+matra+modifier)", in, len(w.Units), unitTypes(w))
				continue
			}
			if w.Units[2].Type != core.UnitModifier {
				t.Errorf("%q: last unit type %v, want modifier", in, w.Units[2].Type)
			}
			cases++
		}
	}

	// (6) +nukta: for the consonants whose base+nukta has a precomposed mapping
	//     (the Perso-Arabic set क़/ज़/ड़/…), the two runes must combine into ONE
	//     unit. Bases with no such mapping legitimately keep the nukta as a
	//     separate symbol, so only the mapped bases are asserted here.
	symbols := (hindi.Hindi{}).Symbols()
	nuktaCombined := 0
	for _, c := range inv.consonants {
		if _, ok := symbols[c+inv.nukta]; !ok {
			continue // no precomposed nukta form for this base
		}
		in := c + inv.nukta
		w := parseWord(in)
		assertWellFormed(t, in, w, inv.halant)
		if len(w.Units) != 1 {
			t.Errorf("%q: got %d units %v, want 1 (precomposed nukta must combine)", in, len(w.Units), unitTypes(w))
		}
		nuktaCombined++
		cases++
	}
	if nuktaCombined == 0 {
		t.Errorf("no nukta-combining consonants exercised; inventory or map changed")
	}

	t.Logf("combinatorial coverage: %d constructs exercised (%d consonants × {bare, %d matras, %d indep vowels, conjunct, 3 nasals, nukta})",
		cases, len(inv.consonants), len(inv.matras), len(inv.indepVowel))
}

// TestCombinatorialDeterministic re-parses every C+matra and C+indep construct
// and asserts identical unit structure (parsing has no hidden state).
func TestCombinatorialDeterministic(t *testing.T) {
	inv := buildInventory(t)
	for _, c := range inv.consonants {
		forms := []string{c}
		for _, m := range inv.matras {
			forms = append(forms, c+m)
		}
		for _, v := range inv.indepVowel {
			forms = append(forms, c+v)
		}
		for _, in := range forms {
			a, b := parseWord(in), parseWord(in)
			if len(a.Units) != len(b.Units) {
				t.Errorf("%q non-deterministic unit count: %d vs %d", in, len(a.Units), len(b.Units))
				continue
			}
			for i := range a.Units {
				if a.Units[i].Type != b.Units[i].Type || string(a.Units[i].Runes) != string(b.Units[i].Runes) {
					t.Errorf("%q unit[%d] differs across parses", in, i)
				}
			}
		}
	}
}

func unitTypes(w *core.Word) []string {
	out := make([]string, len(w.Units))
	for i, u := range w.Units {
		out[i] = u.Type.String()
	}
	return out
}
