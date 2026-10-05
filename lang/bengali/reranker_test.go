package bengali

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
)

type selectorCase struct {
	Word       string              `json:"word"`
	Candidates []string            `json:"candidates"`
	Output     string              `json:"output"`
	Features   []map[string]string `json:"features"`
}

func selectorCases(t *testing.T) []selectorCase {
	t.Helper()
	f, err := os.Open("testdata/selector_parity.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var cases []selectorCase
	if err = json.NewDecoder(z).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 500 {
		t.Fatal("incomplete parity fixture")
	}
	return cases
}

func TestBengaliRerankerParityAndFallbacks(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	opts := core.Options{SchwaModel: true, Rerank: true}
	changed := 0
	for _, tc := range selectorCases(t) {
		for i, alt := range tc.Candidates[1:] {
			if got := selectorFeatures(tc.Word, alt, tc.Candidates[0]); !reflect.DeepEqual(got, tc.Features[i]) {
				t.Fatalf("feature parity %s: %+v != %+v", tc.Word, got, tc.Features[i])
			}
		}
		if got := e.TransliterateWithOptions(tc.Word, opts); got != tc.Output {
			t.Fatalf("output parity %s: %s != %s", tc.Word, got, tc.Output)
		}
		if tc.Output != tc.Candidates[0] {
			changed++
		}
		var actual []string
		_, handled := (Bengali{}).RerankNative(tc.Word, opts, func(lang core.Language, o core.Options) string {
			o.Rerank = false
			o.Lexicon = false
			out := core.NewEngine(lang, colloquial.Colloquial{}).TransliterateWithOptions(tc.Word, o)
			for _, seen := range actual {
				if seen == out {
					return out
				}
			}
			actual = append(actual, out)
			return out
		})
		if handled && !reflect.DeepEqual(actual, tc.Candidates) {
			t.Fatalf("candidate parity %s: %+v != %+v", tc.Word, actual, tc.Candidates)
		}
		alias := strings.NewReplacer("ড়", "ড়", "ো", "ো").Replace(tc.Word)
		if got := e.TransliterateWithOptions("\u200d"+alias, opts); got != tc.Output {
			t.Fatalf("canonical parity %s: %s != %s", alias, got, tc.Output)
		}
		for _, styled := range []core.Options{{SchwaModel: true, InherentVowelA: true}, {SchwaModel: true, LongVowels: true}, {SchwaModel: true, SimpleNasals: true}, {SchwaModel: true, KeepMedialSchwa: true}, {}} {
			before := e.TransliterateWithOptions(tc.Word, styled)
			styled.Rerank = true
			if after := e.TransliterateWithOptions(tc.Word, styled); after != before {
				t.Fatalf("option fallback %s %+v: %s != %s", tc.Word, styled, after, before)
			}
		}
	}
	if changed == 0 {
		t.Fatal("fixture does not exercise selected alternatives")
	}
	if got, info := e.TransliterateDebug("কমল", opts); got == "" || info != nil {
		t.Fatal("handled rerank must return output without misleading single-path traces")
	}
	if _, info := e.TransliterateDebug("ক্ষ", opts); info == nil {
		t.Fatal("unsupported fallback lost debug traces")
	}
	if got := e.TransliterateWithOptions("অডিও", core.Options{Lexicon: true, SchwaModel: true, Rerank: true}); got != "audio" {
		t.Fatal(got)
	}
}

func TestBengaliRerankerArtifactAndRuleControls(t *testing.T) {
	if decodeSelector(selectorJSON) == nil {
		t.Fatal("embedded selector invalid")
	}
	for _, bad := range []string{`{}`, `{"schema":1,"tree":{"counts":[0,0]}}`, `{"schema":1,"tree":{"counts":[-1,3]}}`, `{"schema":1,"tree":{"counts":[1,2],"feature":"unknown"}}`, `{"schema":1,"tree":{"counts":[1,2],"yes":{"counts":[1,2]}}}`} {
		if decodeSelector([]byte(bad)) != nil {
			t.Fatalf("accepted malformed selector %s", bad)
		}
	}
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules("schwa.bengali.model"))
	for _, word := range []string{"কমল", "কমলের", "আহ", "কাম"} {
		before := e.TransliterateWithOptions(word, core.Options{SchwaModel: true})
		after := e.TransliterateWithOptions(word, core.Options{SchwaModel: true, Rerank: true})
		if before != after {
			t.Fatalf("disabled model was re-enabled: %s", word)
		}
	}
	e.RuleEngine().EnableRule("schwa.bengali.model")
	e.RuleEngine().DisableRule("schwa.bengali.final-h")
	base := e.TransliterateWithOptions("আহ", core.Options{SchwaModel: true})
	if out := e.TransliterateWithOptions("আহ", core.Options{SchwaModel: true, Rerank: true}); out != base {
		t.Fatal("runtime rule disable lost")
	}
}

func TestBengaliRerankerConcurrentCalls(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	opts := core.Options{SchwaModel: true, Rerank: true}
	words := []string{"কমলের", "ফলের", "আহ", "কাম", "ড়ক"}
	expected := make([]string, len(words))
	for i, w := range words {
		expected[i] = e.TransliterateWithOptions(w, opts)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j, w := range words {
				if e.TransliterateWithOptions(w, opts) != expected[j] {
					t.Error("shared candidate state")
				}
			}
		}()
	}
	wg.Wait()
}
