package webdemo

import (
	"testing"

	"github.com/budhash/gomanize"
)

// getters maps each flag to the Options field it must control. Every flag in
// FlagNames must appear here, so a new flag without a wired-up field fails the
// completeness check below rather than silently no-opping in the browser.
var getters = map[string]func(gomanize.Options) bool{
	FlagLongVowels:      func(o gomanize.Options) bool { return o.LongVowels },
	FlagSimpleNasals:    func(o gomanize.Options) bool { return o.SimpleNasals },
	FlagKeepMedialSchwa: func(o gomanize.Options) bool { return o.KeepMedialSchwa },
	FlagSchwaModel:      func(o gomanize.Options) bool { return o.SchwaModel },
	FlagLexicon:         func(o gomanize.Options) bool { return o.Lexicon },
	FlagRerank:          func(o gomanize.Options) bool { return o.Rerank },
}

func TestOptionsDefaultsAllFalse(t *testing.T) {
	o := Options(nil)
	for name, get := range getters {
		if get(o) {
			t.Errorf("empty flags: %q should be false, got true", name)
		}
	}
}

func TestOptionsEachFlagSetsOnlyItsField(t *testing.T) {
	for _, flag := range FlagNames() {
		o := Options(map[string]bool{flag: true})
		for name, get := range getters {
			want := name == flag
			if got := get(o); got != want {
				t.Errorf("setting %q: field %q = %v, want %v", flag, name, got, want)
			}
		}
	}
}

func TestFlagNamesCoversEveryGetter(t *testing.T) {
	if len(FlagNames()) != len(getters) {
		t.Fatalf("FlagNames has %d entries, test wires %d fields — keep them in sync",
			len(FlagNames()), len(getters))
	}
	for _, flag := range FlagNames() {
		if _, ok := getters[flag]; !ok {
			t.Errorf("FlagNames includes %q with no field wired in the test", flag)
		}
	}
}

func TestOptionsIgnoresUnknownFlags(t *testing.T) {
	o := Options(map[string]bool{"notARealFlag": true})
	if o != (Options(nil)) {
		t.Errorf("unknown flag changed options: %+v", o)
	}
}
