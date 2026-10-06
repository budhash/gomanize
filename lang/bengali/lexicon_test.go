package bengali

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/scheme/colloquial"
)

func TestBengaliLexiconCanonicalKeys(t *testing.T) {
	f, err := os.Open("testdata/lexicon_keys.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var cases [][2]string
	if err = json.NewDecoder(z).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 4000 {
		t.Fatal("incomplete Unicode parity fixture")
	}
	for _, tc := range cases {
		if got, ok := bengaliLexiconKey(tc[0]); !ok || got != tc[1] {
			t.Fatalf("NFC parity %q: %q != %q", tc[0], got, tc[1])
		}
	}
	for _, bad := range []string{"", "\u200d", "বাংলা!", "বাংলা hindi", "২"} {
		if _, ok := bengaliLexiconKey(bad); ok {
			t.Fatalf("accepted nonword %q", bad)
		}
	}
}

func TestBengaliLexiconEntriesAndStyles(t *testing.T) {
	entries := loadLexicon()
	if LexiconSize() != 8979 {
		t.Fatalf("unexpected lexicon size %d", LexiconSize())
	}
	e := core.NewEngine(Bengali{}, colloquial.Colloquial{})
	hits := 0
	for word, roman := range entries {
		if got, ok := (Bengali{}).LexiconLookupWithOptions(word, core.Options{}); !ok || got != roman {
			t.Fatalf("lookup mismatch for %s", word)
		}
		for _, alias := range []string{strings.NewReplacer("ড়", "ড়", "ঢ়", "ঢ়", "য়", "য়", "ো", "ো", "ৌ", "ৌ").Replace(word), "\u200d" + word} {
			if got, ok := (Bengali{}).LexiconLookupWithOptions(alias, core.Options{}); !ok || got != roman {
				t.Fatalf("canonical lookup mismatch for %s", word)
			}
		}
		if e.TransliterateWithOptions(word, core.Options{Lexicon: true, SchwaModel: true}) != roman {
			t.Fatalf("engine did not prioritize lexicon for %s", word)
		}
		hits++
	}
	if hits == 0 {
		t.Fatal("no lexicon entries exercised")
	}
	for _, word := range []string{"অডিও", "অগাস্ট", "অটো"} {
		if e.Transliterate(word) == entries[word] {
			t.Fatalf("style probe %s does not distinguish lookup from rules", word)
		}
		for _, opts := range []core.Options{{InherentVowelA: true}, {LongVowels: true}, {SimpleNasals: true}, {KeepMedialSchwa: true}, {InherentVowelA: true, SchwaModel: true}} {
			expected := e.TransliterateWithOptions(word, opts)
			opts.Lexicon = true
			if got, found := (Bengali{}).LexiconLookupWithOptions(word, opts); found || got != "" {
				t.Fatalf("style lookup was not bypassed for %s", word)
			}
			if got := e.TransliterateWithOptions(word, opts); got != expected {
				t.Fatalf("style bypass %s: %s != %s", word, got, expected)
			}
		}
	}
	if _, found := (Bengali{}).LexiconLookupWithOptions("ককককককককককককক", core.Options{}); found {
		t.Fatal("unexpected OOV hit")
	}
	got, debug := e.TransliterateDebug("অডিও", core.Options{Lexicon: true})
	if got != "audio" || debug != nil {
		t.Fatal("lexicon debug short-circuit contract changed")
	}
}
