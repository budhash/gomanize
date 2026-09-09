package benchmark

// Tier 1 of the structured parser-QA plan (docs/reviews/2026-09-08-structured-parser-qa.md).
//
// Aggregate accuracy is an average, so it hides a construct that is rare in the
// corpus even when the engine gets that whole construct wrong (this is exactly
// how गई→gi slipped through — the "consonant + independent vowel" construct is
// ~0.5% of Dakshina). This analyzer slices the full corpus by the orthographic
// features each token contains and reports per-construct match-any accuracy with
// the default rule engine (no lexicon/rerank — we are diagnosing the parser and
// rules, which the lexicon would otherwise mask), ranked by failure count.
//
//   go test ./benchmark/... -run TestBenchmarkConstructAnalysis -v
//   make test-constructs

import (
	"os"
	"sort"
	"strings"
	"testing"
)

// Devanagari code-point classes (see Unicode block U+0900–U+097F).
func isIndepVowel(c rune) bool { return c >= 0x0905 && c <= 0x0914 } // अ..औ
func isConsonant(c rune) bool {
	return (c >= 0x0915 && c <= 0x0939) || (c >= 0x0958 && c <= 0x095F) // क..ह + precomposed nukta
}
func isMatra(c rune) bool { return c >= 0x093E && c <= 0x094C } // ा..ौ (dependent vowel signs)

const (
	cHalant       = 0x094D // ्  virama / conjunct former
	cNukta        = 0x093C // ़
	cAnusvara     = 0x0902 // ं
	cChandrabindu = 0x0901 // ँ
	cVisarga      = 0x0903 // ः
	cAvagraha     = 0x093D // ऽ
)

// classifyConstructs returns the set of orthographic constructs present in a
// native token. A token may contribute to several buckets.
func classifyConstructs(w string) []string {
	r := []rune(w)
	set := map[string]bool{}
	mark := func(k string) { set[k] = true }

	for i, c := range r {
		switch {
		case isIndepVowel(c):
			mark("independent_vowel")
			if i > 0 {
				p := r[i-1]
				if isConsonant(p) {
					mark("cons_then_indep_vowel") // गई, नई, कई, गए — the sharp case
				}
				if isConsonant(p) || isMatra(p) || p == cAnusvara || p == cVisarga || p == cChandrabindu || p == cHalant {
					mark("indep_vowel_medial") // any independent vowel fused to a preceding syllable
				}
			}
		case c == cHalant:
			mark("conjunct")
		case c == cNukta:
			mark("nukta")
		case c == cAnusvara:
			mark("anusvara")
		case c == cChandrabindu:
			mark("chandrabindu")
		case c == cVisarga:
			mark("visarga")
		case c == cAvagraha:
			mark("avagraha")
		case c >= 0x0966 && c <= 0x096F:
			mark("devanagari_digit")
		}
		if c >= 0x0958 && c <= 0x095F {
			mark("nukta") // precomposed nukta consonant
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestBenchmarkConstructAnalysis(t *testing.T) {
	path := getTestDataPath("dakshina_hi.csv")
	refs, err := loadReferenceSets(path)
	if os.IsNotExist(err) {
		t.Skipf("dakshina data not found: %v", err)
		return
	} else if err != nil {
		t.Fatalf("loading references: %v", err)
	}

	engine := newEngine()

	type stat struct {
		total, fail int
		samples     []string // up to a few "native → got (refs…)" for triage
	}
	byC := map[string]*stat{}
	overallTotal, overallFail := 0, 0

	natives := make([]string, 0, len(refs))
	for n := range refs {
		natives = append(natives, n)
	}
	sort.Strings(natives) // deterministic

	for _, n := range natives {
		got := engine.Transliterate(n)
		ok := matchesAny(got, refs[n])
		overallTotal++
		if !ok {
			overallFail++
		}
		for _, c := range classifyConstructs(n) {
			s := byC[c]
			if s == nil {
				s = &stat{}
				byC[c] = s
			}
			s.total++
			if !ok {
				s.fail++
				if len(s.samples) < 6 {
					refShown := refs[n]
					if len(refShown) > 3 {
						refShown = refShown[:3]
					}
					s.samples = append(s.samples, n+" → "+got+"  (want: "+strings.Join(refShown, ", ")+")")
				}
			}
		}
	}

	// rank constructs by failure count (absolute impact), then by error rate
	labels := make([]string, 0, len(byC))
	for k := range byC {
		labels = append(labels, k)
	}
	sort.Slice(labels, func(i, j int) bool {
		a, b := byC[labels[i]], byC[labels[j]]
		if a.fail != b.fail {
			return a.fail > b.fail
		}
		return errRate(a.fail, a.total) > errRate(b.fail, b.total)
	})

	overallErr := errRate(overallFail, overallTotal)
	t.Logf("")
	t.Logf("========================================")
	t.Logf("PER-CONSTRUCT ACCURACY (default rules, match-any, full Dakshina)")
	t.Logf("========================================")
	t.Logf("overall: %d tokens, %d fail (%.1f%% error / %.1f%% match-any)",
		overallTotal, overallFail, overallErr, 100-overallErr)
	t.Logf("%-24s %8s %8s %8s %8s", "construct", "tokens", "fails", "err%", "share%")
	for _, k := range labels {
		s := byC[k]
		t.Logf("%-24s %8d %8d %7.1f%% %7.1f%%",
			k, s.total, s.fail, errRate(s.fail, s.total), 100*float64(s.total)/float64(overallTotal))
	}

	// actionable samples for the constructs with the worst error rates (min N)
	t.Logf("")
	t.Logf("--- sample failures (worst constructs by error rate, N>=20) ---")
	byErr := append([]string(nil), labels...)
	sort.Slice(byErr, func(i, j int) bool {
		return errRate(byC[byErr[i]].fail, byC[byErr[i]].total) > errRate(byC[byErr[j]].fail, byC[byErr[j]].total)
	})
	shown := 0
	for _, k := range byErr {
		s := byC[k]
		if s.total < 20 || shown >= 5 {
			continue
		}
		shown++
		t.Logf("[%s] %.1f%% error (%d/%d):", k, errRate(s.fail, s.total), s.fail, s.total)
		for _, ex := range s.samples {
			t.Logf("    %s", ex)
		}
	}
}

func errRate(fail, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(fail) / float64(total)
}
