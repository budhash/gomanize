package bengali

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode"

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
		got, ok := vowelFeatures(vowelWordView(tc.Word), tc.Index, model.maxCluster)
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
		{"মত", "mot", "mot", "mat"},
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
	for _, word := range []string{"ক্ষ্ম্য", "জ্ঞান", "ক্কঅ", "ক্", "হও", "হওয়া", "বাংলা", "চাঁদ", "কঅ", "কও", "উৎসব", "কাা", "ঐ", "aক"} {
		if vowelEligible(vowelWordView(word), inventory.MaxCluster) {
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
	_, info := e.TransliterateDebug("কলম", core.Options{SchwaModel: true})
	found := false
	for _, tr := range info.Traces {
		if tr.Rule == "schwa.bengali.model" && tr.Unit == "ল" && strings.Contains(tr.Metadata, "quality=Raised") {
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

// stripFormat mirrors the parser's (and native_key's) removal of Cf runes.
func stripFormat(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}

func TestBengaliVowelEligibilityParity(t *testing.T) {
	data, err := os.ReadFile("testdata/vowel_eligibility.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Word           string
		Eligible       bool
		LegacyEligible bool `json:"legacy_eligible"`
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 100 {
		t.Fatal("incomplete eligibility parity corpus")
	}
	both := map[[2]bool]bool{}
	for _, tc := range cases {
		word := vowelWordView(stripFormat(tc.Word))
		got := [2]bool{vowelEligible(word, inventory.MaxCluster), vowelEligible(word, 1)}
		if got != [2]bool{tc.Eligible, tc.LegacyEligible} {
			t.Errorf("%q: Go eligible/legacy %v, Python %v/%v", tc.Word, got, tc.Eligible, tc.LegacyEligible)
		}
		both[got] = true
	}
	if !both[[2]bool{true, false}] || !both[[2]bool{true, true}] || !both[[2]bool{false, false}] {
		t.Fatal("eligibility fixture must cover new, previous and ineligible words")
	}
}

func TestBengaliVowelClusterUnits(t *testing.T) {
	units, ok := vowelUnits([]rune("গর্ত"), inventory.MaxCluster)
	want := []vowelUnit{{0, "গ", true}, {1, "র", false}, {2, "্", false}, {3, "ত", true}}
	if !ok || !reflect.DeepEqual(units, want) {
		t.Fatalf("units = %v", units)
	}
	// A matra after the last member removes its slot; nukta letters stay one unit.
	units, ok = vowelUnits([]rune("ড়্গা"), inventory.MaxCluster)
	if !ok || len(units) != 4 || units[0].spelling != "ড়" || units[0].inherent || units[2].inherent {
		t.Fatalf("nukta cluster units = %v", units)
	}
	if vowelEligible("স্ত্রী", inventory.MaxCluster) {
		t.Fatal("a word with no inherent slot is not eligible")
	}
	for _, word := range []string{"ক্ষ", "স্ত্র", "গর্ত"} {
		if !vowelEligible(word, inventory.MaxCluster) || vowelEligible(word, 1) {
			t.Errorf("%s: clusters must be eligible only in the cluster scope", word)
		}
	}
	for _, word := range []string{"ক্ষ্ম্য", "্ক", "ক্্ক", "ক্", "ক্অ", "ক্কঅ", "জ্ঞ", "ক্য়", "ক্ৎ", "া্ক"} {
		if vowelEligible(word, inventory.MaxCluster) {
			t.Errorf("%s: ineligible cluster accepted", word)
		}
	}
	got, ok := vowelFeatures("স্মরণ", 2, inventory.MaxCluster)
	if !ok || got["cons"] != "ম" || got["prev"] != "্" || got["first"] != "0" || got["position"] != "2" {
		t.Fatalf("cluster features = %v", got)
	}
	if got, _ := vowelFeatures("গর্ত", 0, inventory.MaxCluster); got["next2"] != "্" {
		t.Fatalf("next2 = %v", got)
	}
	if _, ok := vowelFeatures("গর্ত", 1, inventory.MaxCluster); ok {
		t.Fatal("a member before a halant has no slot")
	}
}

func TestBengaliVowelClusterPrecedence(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	decided := func(word string) map[string]bool {
		_, info := e.TransliterateDebug(word, core.Options{SchwaModel: true})
		units := map[string]bool{}
		for _, tr := range info.Traces {
			if tr.Rule == "schwa.bengali.model" {
				units[tr.Unit] = true
			}
		}
		return units
	}
	// The model decides a cluster's last member (ম in স্মরণ) ...
	if !decided("স্মরণ")["ম"] {
		t.Fatal("model did not decide the cluster's last member")
	}
	// ... but never a member before a halant, and higher-priority rules
	// (schwa.bengali.final-cluster, 80) still own a word-final cluster.
	if units := decided("গর্ত"); units["র"] || units["ত"] || !units["গ"] {
		t.Fatalf("precedence changed: %v", units)
	}
	if e.Transliterate("গর্ত") != e.TransliterateWithOptions("গর্ত", core.Options{SchwaModel: true}) {
		t.Fatal("final-cluster rule no longer overrides the model")
	}
}

func TestBengaliVowelModelScopes(t *testing.T) {
	model := decodeVowels(vowelTreeJSON)
	if model == nil || model.Scope != inventory.Scope || model.maxCluster != inventory.MaxCluster {
		t.Fatal("embedded model must use the inventory's cluster scope")
	}
	legacy, err := os.ReadFile("testdata/vowel_tree_simple_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if old := decodeVowels(legacy); old == nil || old.maxCluster != 1 {
		t.Fatal("the legacy simple-word tree must decode with clusters disabled")
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(vowelTreeJSON, &raw); err != nil {
		t.Fatal(err)
	}
	raw["scope"] = "unknown-scope"
	data, _ := json.Marshal(raw)
	if decodeVowels(data) != nil {
		t.Fatal("accepted an unknown scope")
	}
	for _, bad := range []string{`not json`, `{"schema":2,"scope":"s","legacy_scope":"l","halant":"্","max_cluster":3}`,
		`{"schema":1,"scope":"s","legacy_scope":"l","halant":"্","max_cluster":1}`,
		`{"schema":1,"scope":"s","legacy_scope":"l","halant":"্","max_cluster":3,"matras":["াি"]}`} {
		if decodeInventory([]byte(bad)) != nil {
			t.Errorf("accepted malformed inventory %s", bad)
		}
	}
}
