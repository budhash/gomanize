package bengali

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestBengaliVowelFeatureParity(t *testing.T) {
	data, err := os.ReadFile("testdata/vowel_features.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Word     string
		Index    int
		Features map[string]string
		Label    int
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	model := decodeVowels(vowelTreeJSON)
	if model == nil {
		t.Fatal("invalid embedded model")
	}
	if len(cases) < 250 {
		t.Fatal("incomplete feature parity corpus")
	}
	labels := map[int]bool{}
	for _, tc := range cases {
		// Runtime normalizes through vowelWordView first; the fixture includes non-NFC words.
		got, ok := vowelFeatures(vowelWordView(tc.Word), tc.Index)
		if !ok || !reflect.DeepEqual(got, tc.Features) {
			t.Fatalf("%s@%d: %v != %v", tc.Word, tc.Index, got, tc.Features)
		}
		if label := predictVowel(model, got); label != tc.Label {
			t.Fatalf("prediction %s@%d: %d != %d", tc.Word, tc.Index, label, tc.Label)
		}
		labels[tc.Label] = true
	}
	if len(labels) != 3 {
		t.Fatal("parity fixture must cover all output classes")
	}
}

func TestBengaliVowelModelScopeAndStyle(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	disabled := core.NewEngine(Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules("schwa.bengali.model"))
	for _, tc := range []struct{ input, baseline, model, style string }{
		{"মত", "mot", "moto", "mato"},
		{"কোড়ক", "korok", "kork", "kork"},
		{"কোড়ক", "korok", "kork", "kork"},
		{"কোড়ক", "korok", "kork", "kork"},
		{"মন", "mon", "mon", "man"},
	} {
		if got := e.Transliterate(tc.input); got != tc.baseline {
			t.Errorf("baseline %s = %s", tc.input, got)
		}
		if got := e.TransliterateWithOptions(tc.input, core.Options{SchwaModel: true}); got != tc.model {
			t.Errorf("model %s = %s, want %s", tc.input, got, tc.model)
		}
		if got := disabled.TransliterateWithOptions(tc.input, core.Options{SchwaModel: true}); got != tc.baseline {
			t.Errorf("disable model %s = %s", tc.input, got)
		}
		if got := e.TransliterateWithOptions(tc.input, core.Options{SchwaModel: true, InherentVowelA: true}); got != tc.style {
			t.Errorf("style %s = %s", tc.input, got)
		}
	}
	for _, word := range []string{"মধ্য", "হও", "হওয়া", "ক্ষতি", "বাংলা", "চাঁদ", "কঅ", "কও", "উৎসব", "কাা", "ঐ", "aক"} {
		if _, ok := vowelUnits([]rune(vowelWordView(word))); ok {
			t.Errorf("unsupported word accepted: %s", word)
		}
		before := e.Transliterate(word)
		after, info := e.TransliterateDebug(word, core.Options{SchwaModel: true})
		if before != after {
			t.Errorf("fallback %s: %s != %s", word, after, before)
		}
		for _, tr := range info.Traces {
			if tr.Rule == "schwa.bengali.model" {
				t.Errorf("model acted on unsupported %s", word)
			}
		}
	}
	if e.TransliterateWithOptions("কো\u200dড়ক", core.Options{SchwaModel: true}) != "kork" {
		t.Fatal("format marks shifted model positions")
	}
	_, info := e.TransliterateDebug("মত", core.Options{SchwaModel: true})
	found := false
	for _, tr := range info.Traces {
		if tr.Rule == "schwa.bengali.model" && tr.Unit == "ত" && strings.Contains(tr.Metadata, "quality=Raised") {
			found = true
		}
	}
	if !found {
		t.Fatal("missing raised-vowel model trace")
	}
}

func TestBengaliVowelModelRejectsMalformedTrees(t *testing.T) {
	var original map[string]interface{}
	if err := json.Unmarshal(vowelTreeJSON, &original); err != nil {
		t.Fatal(err)
	}
	for _, tree := range []interface{}{nil, map[string]interface{}{"leaf": 3}, map[string]interface{}{"f": "unknown", "v": "x", "yes": map[string]int{"leaf": 0}, "no": map[string]int{"leaf": 1}}, map[string]interface{}{"f": "cons", "v": "ক", "yes": map[string]int{"leaf": 0}}, map[string]interface{}{"leaf": 0, "yes": map[string]int{"leaf": 1}}} {
		original["tree"] = tree
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		if decodeVowels(data) != nil {
			t.Errorf("accepted malformed model %s", data)
		}
	}
	if err := json.Unmarshal(vowelTreeJSON, &original); err != nil {
		t.Fatal(err)
	}
	original["classes"] = []string{"raised", "default", "absent"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if decodeVowels(data) != nil {
		t.Fatal("accepted reordered output labels")
	}
	if decodeVowels([]byte("not json")) != nil {
		t.Fatal("accepted invalid JSON")
	}
}
