package bengali

import (
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
	"github.com/budhash/gomanize/script/brahmic"
)

func TestB1ProductionOutputs(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	for input, want := range map[string]string{
		"অন্ত": "onto", "শব্দ": "shobdo", "কর্ম": "kormo", "গ্রহ": "groho", "মন": "mon",
		"মধ্য": "moddho", "মধ্যা": "moddha", "পদ্য": "poddo", "গদ্য": "goddo", "জন্য": "jonno",
		"ব্যবহার": "byabohar", "ব্যথা": "byatha", "স্বামী": "shami", "ত্যাগ": "tyag", "জ্বর": "jor", "উত্থান": "utthan", "স্বাধীন": "shadhin", "তত্ব": "totto",
		"পদ্ম": "poddo", "স্মরণ": "shoron", "জন্ম": "jonmo", "গর্ব": "gorbo", "কম্বল": "kombol",
		"ক্ষতি": "khoti", "ক্ষমা": "khoma", "কক্ষ": "kokkho", "কক্ষে": "kokkhe",
		"জ্ঞান": "gyan", "জ্ঞাত": "gyat", "বিজ্ঞ": "biggo", "অজ্ঞান": "oggan",
		"ব্রহ্ম": "brommo", "চিহ্ন": "chinno", "চিহ্নে": "chinne",
		"দুঃখ": "dukkho", "দুঃখে": "dukkhe", "অতঃপর": "otohpor", "উঃ": "uh",
		"হও": "hao", "হওয়া": "howa", "হওয়া": "howa", "ও": "o", "কও": "ko", "য়া": "ya",
		"স্তর": "stor", "স্থল": "sthol", "আস্তে": "aste", "স্বর": "shor", "আর্ট": "art", "কর্ড": "kord", "টেস্ট": "test",
		"ফল": "fol", "ফুল": "ful", "বাংলা": "bangla", "চাঁদ": "chand", "চাঁ": "chan", "হঠাৎ": "hothat", "উৎসব": "utshob",
	} {
		if got := e.Transliterate(input); got != want {
			t.Errorf("%s: %s != %s", input, got, want)
		}
	}
	for input, want := range map[string]string{"মধ্য": "maddha", "দুঃখ": "dukkha", "হও": "hao", "হওয়া": "hawa", "ব্যবহার": "byabahar", "অতি": "ati"} {
		if got := e.TransliterateWithOptions(input, core.Options{InherentVowelA: true}); got != want {
			t.Errorf("a style %s: %s != %s", input, got, want)
		}
	}
}

func TestB1ProductionRuleDisableControls(t *testing.T) {
	tests := []struct{ input, rule, want, without string }{
		{"কর্ম", "schwa.bengali.final-cluster", "kormo", "korm"},
		{"মোহ", "schwa.bengali.final-h", "moho", "moh"},
		{"দুঃখ", "schwa.bengali.post-visarga", "dukkho", "dukkh"},
		{"ফল", "consonant.bengali.f-spelling", "fol", "phol"},
		{"চিহ্ন", "consonant.bengali.h-nasal", "chinno", "chihno"},
		{"ক্ষতি", "consonant.bengali.ksha-left", "khoti", "kkhoti"},
		{"ক্ষতি", "consonant.bengali.ksha-right", "khoti", "shoti"},
		{"জ্ঞান", "consonant.bengali.jna-left", "gyan", "jgyan"},
		{"জ্ঞান", "consonant.bengali.jna-right", "gyan", "nan"},
		{"স্তর", "consonant.bengali.s-cluster", "stor", "shtor"},
		{"স্বামী", "consonant.bengali.initial-phala", "shami", "shbami"},
		{"মধ্য", "render.bengali.medial-phala", "moddho", "modhjo"},
		{"দুঃখ", "consonant.bengali.visarga-suppress", "dukkho", "duhkkho"},
		{"দুঃখ", "render.bengali.visarga-gemination", "dukkho", "dukho"},
		{"হও", "vowel.bengali.hao", "hao", "hoo"},
		{"হওয়া", "vowel.bengali.howa-w", "howa", "hooa"},
		{"হওয়া", "consonant.bengali.howa-y", "howa", "howya"},
	}
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	for _, tc := range tests {
		t.Run(tc.rule, func(t *testing.T) {
			if got := e.Transliterate(tc.input); got != tc.want {
				t.Fatalf("enabled: %s != %s", got, tc.want)
			}
			off := core.NewEngine(Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules(tc.rule))
			if got := off.Transliterate(tc.input); got != tc.without {
				t.Fatalf("disabled: %s != %s", got, tc.without)
			}
		})
	}
}

func TestB1ProductionMetadataTraces(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	for _, tc := range []struct{ input, rule, unit, metadata string }{
		{"মধ্য", "render.bengali.medial-phala", "য", "gemination=RepeatPrevious"},
		{"দুঃখ", "render.bengali.visarga-gemination", "খ", "gemination=GeminateSelf"},
		{"হও", "vowel.bengali.hao", "হ", "quality=Open"},
	} {
		_, info := e.TransliterateDebug(tc.input, core.Options{})
		found := false
		for _, tr := range info.Traces {
			if tr.Rule == tc.rule && tr.Unit == tc.unit && strings.Contains(tr.Metadata, tc.metadata) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing production metadata trace for %+v", tc)
		}
	}
}

func TestB1RulesOwnTheirUnits(t *testing.T) {
	// NewRuleEngine also rejects all composed priority ties, including disabled rules.
	_ = core.NewRuleEngine(RuleCatalog().AllRules())
	for _, r := range RuleCatalog().AllRules() {
		for _, input := range []string{"অন্ত", "মোহ", "দুঃখ", "ক্ষতি", "কক্ষ", "জ্ঞান", "বিজ্ঞ", "চিহ্ন", "ফল", "স্তর", "স্বামী", "ব্যবহার", "মধ্য", "হও", "হওয়া", "চিহ্নিত", "গত", "পেইন্ট", "এখনও", "কও", "দৃশ্যের", "শ্যাম"} {
			w := brahmic.NewParser(Bengali{}.ScriptConfig()).Parse(input, Symbols)
			w.Options.InherentVowelA = true
			w.Options.SchwaModel = true
			brahmic.IdentifyRuns(w)
			for _, u := range w.Units {
				if r.Condition(u, w) {
					rom := make([]string, len(w.Units))
					data := make([]brahmic.BrahmicData, len(w.Units))
					identity := string(u.Runes)
					for i, x := range w.Units {
						rom[i] = x.BaseRom
						data[i] = *brahmic.GetBrahmicData(x)
					}
					r.Action(u, w)
					if string(u.Runes) != identity {
						t.Fatalf("%s changed source identity", r.Name)
					}
					for i, x := range w.Units {
						if x != u && (rom[i] != x.BaseRom || data[i] != *brahmic.GetBrahmicData(x)) {
							t.Fatalf("%s mutated neighbor", r.Name)
						}
					}
				}
			}
		}
	}
}

func TestRuleBundleProductionOutputs(t *testing.T) {
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	for input, want := range map[string]string{
		"চিহ্নিত": "chinnito", "নির্ধারিত": "nirdharito", "গীত": "gito", "মৃত": "mrito", "দূত": "duto",
		"গত": "goto", "শত": "shoto", "ভারত": "bharoto", "নত": "noto",
		"পেইন্ট": "peint", "গোল্ড": "gold", "মার্ক": "mark", "বক্স": "boksh", "পঞ্জ": "ponj",
		"এখনও": "ekhono", "কখনও": "kokhono", "কও": "ko",
		"দৃশ্যের": "drishyer", "আলস্য": "aloshyo", "অবশ্য": "oboshyo",
		// Negative contexts: none of the bundle rules fire.
		"যত": "jot", "রাত": "rat", "হাত": "hat", "চিত্ত": "chitto", "শরৎ": "shorot", "হিত্": "hit",
		"অন্ত": "onto", "মন্ত্র": "montro", "আর্ট": "art", "যাও": "jao", "তবুও": "tobuo", "ও": "o",
		"শ্যাম": "shyam", "মধ্য": "moddho", "বিশ্বাস": "bishshash",
	} {
		if got := e.Transliterate(input); got != want {
			t.Errorf("%s: %s != %s", input, got, want)
		}
	}
	for input, want := range map[string]string{"চিহ্নিত": "chinnita", "গত": "gata", "এখনও": "ekhana", "দৃশ্যের": "drishyer"} {
		if got := e.TransliterateWithOptions(input, core.Options{InherentVowelA: true}); got != want {
			t.Errorf("a style %s: %s != %s", input, got, want)
		}
	}
	// The bundle overrides the vowel model's final decision as it does B1's.
	for input, want := range map[string]string{"চিহ্নিত": "chinnito", "গত": "goto", "পেইন্ট": "peint", "এখনও": "ekhono"} {
		if got := e.TransliterateWithOptions(input, core.Options{SchwaModel: true}); got != want {
			t.Errorf("model %s: %s != %s", input, got, want)
		}
	}
}

func TestRuleBundleDisableControls(t *testing.T) {
	tests := []struct {
		input, want, without string
		rules                []string
	}{
		{"চিহ্নিত", "chinnito", "chinnit", []string{"schwa.bengali.final-ito"}},
		{"গত", "goto", "got", []string{"schwa.bengali.final-ta"}},
		{"পেইন্ট", "peint", "peinto", []string{"schwa.bengali.loan-coda"}},
		{"এখনও", "ekhono", "ekhonoo", []string{"vowel.bengali.final-o"}},
		{"দৃশ্যের", "drishyer", "drishsher", []string{"consonant.bengali.sibilant-ya", "render.bengali.sibilant-ya"}},
		// The pair splits spelling from gemination: without the render half the
		// medial-phala copy of the sibilant hides the y spelling.
		{"দৃশ্যের", "drishyer", "drishjer", []string{"consonant.bengali.sibilant-ya"}},
		{"দৃশ্যের", "drishyer", "drishsher", []string{"render.bengali.sibilant-ya"}},
	}
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	for _, tc := range tests {
		t.Run(strings.Join(tc.rules, "+"), func(t *testing.T) {
			if got := e.Transliterate(tc.input); got != tc.want {
				t.Fatalf("enabled: %s != %s", got, tc.want)
			}
			off := core.NewEngine(Bengali{}, colloquial.Colloquial{}, core.WithDisabledRules(tc.rules...))
			if got := off.Transliterate(tc.input); got != tc.without {
				t.Fatalf("disabled: %s != %s", got, tc.without)
			}
		})
	}
}
