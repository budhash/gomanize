package benchmark

import (
	"bufio"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	gomanize "github.com/budhash/gomanize"
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/bengali"
)

func TestBengaliLexiconHeldoutIsolation(t *testing.T) {
	for _, name := range []string{"dakshina-dev", "dakshina-test", "google-dev", "google-test"} {
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("..", "training", "data", "bengali", name+".txt.gz"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			z, err := gzip.NewReader(f)
			if err != nil {
				t.Fatal(err)
			}
			defer z.Close()
			scanner := bufio.NewScanner(z)
			count := 0
			for scanner.Scan() {
				count++
				if roman, ok := (bengali.Bengali{}).LexiconLookupWithOptions(scanner.Text(), core.Options{}); ok {
					t.Fatalf("held-out lookup %s -> %s", scanner.Text(), roman)
				}
			}
			if err = scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if count == 0 {
				t.Fatal("empty isolation set")
			}
			t.Logf("%s: %d types, zero lexicon hits", name, count)
		})
	}
}

func TestBenchmarkBengaliLexicon(t *testing.T) {
	options := []gomanize.Options{{}, {SchwaModel: true}, {Lexicon: true}, {SchwaModel: true, Lexicon: true}}
	names := []string{"B1-pure", "model-pure", "B1-lexicon", "model-lexicon"}
	engines := make([]*gomanize.Gomanize, len(options))
	for i, opts := range options {
		var err error
		engines[i], err = gomanize.NewWithOptions("bengali", opts)
		if err != nil {
			t.Fatal(err)
		}
	}
	for split, words := range loadBengali(t) {
		t.Run(split, func(t *testing.T) {
			counts := make([]bnCounts, len(engines))
			hits := 0
			for _, word := range words {
				out := make([]string, len(engines))
				for i, e := range engines {
					out[i] = e.Translit(word.native)
					counts[i].add(out[i], word.refs)
				}
				if _, found := (bengali.Bengali{}).LexiconLookupWithOptions(word.native, core.Options{}); found {
					hits++
				}
				if split != "train" && (out[0] != out[2] || out[1] != out[3]) {
					t.Fatalf("lexicon changed held-out output for %s", word.native)
				}
			}
			if split != "train" && hits != 0 {
				t.Fatal("held-out lexicon coverage")
			}
			for i, c := range counts {
				t.Logf("%s %s words=%d strict=%d any=%d CER=%.8f lexicon-hits=%d (train scores are in-sample, not generalization)", split, names[i], c.words, c.strict, c.any, c.minCER/float64(c.words), hits)
			}
		})
	}
}

func TestBengaliLexiconSentenceAPI(t *testing.T) {
	e, err := gomanize.NewWithOptions("bengali", gomanize.Options{Lexicon: true, SchwaModel: true})
	if err != nil {
		t.Fatal(err)
	}
	input := "অডিও, অটো।\nঅগাস্ট"
	if got := e.Translit(input); got != "audio, auto।\naugust" {
		t.Fatal(got)
	}
	plain, err := gomanize.NewWithOptions("bengali", gomanize.Options{InherentVowelA: true, SchwaModel: true})
	if err != nil {
		t.Fatal(err)
	}
	styled, err := gomanize.NewWithOptions("bengali", gomanize.Options{Lexicon: true, InherentVowelA: true, SchwaModel: true})
	if err != nil {
		t.Fatal(err)
	}
	if styled.Translit(input) != plain.Translit(input) {
		t.Fatal("sentence lexicon overrode alternate style")
	}
}
