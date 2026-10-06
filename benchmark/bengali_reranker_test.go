package benchmark

import (
	"testing"

	gomanize "github.com/budhash/gomanize"
)

func TestBengaliRerankerFrozenMetrics(t *testing.T) {
	model, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true})
	rank, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true, Rerank: true})
	lex, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true, Rerank: true, Lexicon: true})
	for _, tc := range []struct {
		split       string
		strict, any int
	}{{"dev", 898, 1553}, {"test", 916, 1576}} {
		var before, after bnCounts
		for _, w := range loadBengali(t, tc.split)[tc.split] {
			a, b := model.Translit(w.native), rank.Translit(w.native)
			before.add(a, w.refs)
			after.add(b, w.refs)
			if lex.Translit(w.native) != b {
				t.Fatal("lexicon changed excluded word")
			}
		}
		if after.strict != tc.strict || after.any != tc.any || after.minCER >= before.minCER {
			t.Fatalf("frozen selector parity %s: strict=%d any=%d CER=%f", tc.split, after.strict, after.any, after.minCER/float64(after.words))
		}
	}
}

func TestBengaliRerankerSentenceAPI(t *testing.T) {
	for _, useLexicon := range []bool{false, true} {
		e, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true, Rerank: true, Lexicon: useLexicon})
		input := "জনক, সমতা।\nঅডিও\tEnglish 123"
		want := e.Translit("জনক") + ", " + e.Translit("সমতা") + "।\n" + e.Translit("অডিও") + "\tEnglish 123"
		if got := e.Translit(input); got != want {
			t.Fatalf("boundaries: %q != %q", got, want)
		}
	}
}
