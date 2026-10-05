package benchmark

import (
	"os"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

func TestBengaliVowelModel(t *testing.T) {
	splits := []string{"dev"}
	if os.Getenv("BENGALI_MODEL_REPORT_TEST") == "1" {
		splits = append(splits, "test")
	}
	baseline, _ := gomanize.New("bengali")
	model, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true})
	for _, split := range splits {
		var a, b bnCounts
		changed := 0
		for _, word := range loadBengali(t, split)[split] {
			before, after := baseline.Translit(word.native), model.Translit(word.native)
			a.add(before, word.refs)
			b.add(after, word.refs)
			if before != after {
				changed++
			}
		}
		if split == "dev" && (b.any <= a.any || b.strict < a.strict || b.minCER >= a.minCER) {
			t.Fatalf("vowel model failed dev gate: any must improve, strict cannot regress, CER must improve")
		}
		t.Logf("%s words=%d B1 strict=%d any=%d CER=%.8f model strict=%d any=%d CER=%.8f changed=%d", split, a.words, a.strict, a.any, a.minCER/float64(a.words), b.strict, b.any, b.minCER/float64(b.words), changed)
	}
}
