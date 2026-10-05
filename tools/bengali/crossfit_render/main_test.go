package main

import (
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestInjectedDecisionsAndRuleOwnership(t *testing.T) {
	l := &injected{}
	engine := core.NewEngine(l, colloquial.Colloquial{})
	for _, tc := range []struct {
		native string
		labels map[int]int
		want   string
	}{
		{"কম", map[int]int{0: 0, 1: 1}, "kmo"},
		{"কম", map[int]int{0: 1, 1: 0}, "kom"},
		{"কম", map[int]int{0: 2, 1: 0}, "kom"},
		{"কম", nil, "kom"},
		{"আহ", map[int]int{1: 0}, "aho"},        // higher-priority final-h rule owns retention
		{"কাম", map[int]int{0: 0, 2: 0}, "kam"}, // explicit matra is not an inherent slot
	} {
		l.labels = tc.labels
		got := engine.TransliterateWithOptions(tc.native, core.Options{SchwaModel: true})
		if got != tc.want {
			t.Errorf("%s %+v = %q, want %q", tc.native, tc.labels, got, tc.want)
		}
	}
	l.labels = map[int]int{0: 0, 1: 1}
	if got := engine.Transliterate("কম"); got != "kom" {
		t.Fatalf("injection bypassed option: %q", got)
	}
}

func TestValidateRequests(t *testing.T) {
	if err := validate(request{Native: "কম", Variants: []map[int]int{{0: 2}}, Check: map[int]int{1: 0}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []request{
		{}, {Native: "ক"}, {Native: "ক", Variants: make([]map[int]int, 10)},
		{Native: "ক", Variants: []map[int]int{{-1: 0}}},
		{Native: "ক", Variants: []map[int]int{{1: 0}}},
		{Native: "ক", Variants: []map[int]int{{0: 3}}},
		{Native: "ক", Variants: []map[int]int{{}}, Check: map[int]int{0: -1}},
	} {
		if validate(r) == nil {
			t.Errorf("accepted invalid request: %+v", r)
		}
	}
}
