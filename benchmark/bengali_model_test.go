package benchmark

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

// bnPinnedModel pins the opt-in vowel model's train/dev outputs (strict,
// match-any, words changed vs B1, digest of every model output). The relative
// dev guard below cannot see a degraded model that still beats B1. A change to
// the model or the rules it composes with must update this deliberately.
var bnPinnedModel = map[string]struct {
	strict, any, changed int
	digest               string
}{
	"train": {9026, 15262, 3372, "0784bb6674748c5f6f0944e5a675f550fa0a65a487c4b083c7a12dcaa61d140e"},
	"dev":   {895, 1545, 408, "a365907e0142de00a59e2c0d7ee9451b182ea27e8c113b380377a819e18ac47e"},
}

func TestBengaliVowelModel(t *testing.T) {
	splits := []string{"train", "dev"}
	if os.Getenv("BENGALI_MODEL_REPORT_TEST") == "1" {
		splits = append(splits, "test")
	}
	baseline, _ := gomanize.New("bengali")
	model, _ := gomanize.NewWithOptions("bengali", gomanize.Options{SchwaModel: true})
	for _, split := range splits {
		var a, b bnCounts
		changed := 0
		digest := sha256.New()
		for _, word := range loadBengali(t, split)[split] {
			before, after := baseline.Translit(word.native), model.Translit(word.native)
			fmt.Fprintf(digest, "%s\t%s\n", word.native, after)
			a.add(before, word.refs)
			b.add(after, word.refs)
			if before != after {
				changed++
			}
		}
		if split == "dev" && (b.any <= a.any || b.strict < a.strict || b.minCER >= a.minCER) {
			t.Fatalf("vowel model failed dev gate: any must improve, strict cannot regress, CER must improve")
		}
		if pin, ok := bnPinnedModel[split]; ok {
			got := fmt.Sprintf("%x", digest.Sum(nil))
			if b.strict != pin.strict || b.any != pin.any || changed != pin.changed || got != pin.digest {
				t.Errorf("%s vowel model output changed: strict=%d any=%d changed=%d digest=%s; pinned %d/%d/%d %s (update bnPinnedModel deliberately and report the delta)", split, b.strict, b.any, changed, got, pin.strict, pin.any, pin.changed, pin.digest)
			}
		}
		t.Logf("%s words=%d B1 strict=%d any=%d CER=%.8f model strict=%d any=%d CER=%.8f changed=%d", split, a.words, a.strict, a.any, a.minCER/float64(a.words), b.strict, b.any, b.minCER/float64(b.words), changed)
	}
}
