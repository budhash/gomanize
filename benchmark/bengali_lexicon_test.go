package benchmark

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
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

// bnPinnedLexiconTrain pins in-sample train scores for the lexicon modes
// (strict, match-any, hits) and a digest of their outputs. Held-out invariance
// is asserted below; this catches drift in what the lexicon does fire on.
var bnPinnedLexiconTrain = struct {
	b1Strict, b1Any, modelStrict, modelAny, hits int
	digest                                       string
}{12602, 17400, 13045, 18112, 8976, "d0e161d0d49edcb2a931c8c8f4cf242a8db0720eef9932bf54de25050937e309"}

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
			digest := sha256.New()
			for _, word := range words {
				out := make([]string, len(engines))
				for i, e := range engines {
					out[i] = e.Translit(word.native)
					counts[i].add(out[i], word.refs)
				}
				fmt.Fprintf(digest, "%s\t%s\t%s\n", word.native, out[2], out[3])
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
			if split == "train" {
				p := bnPinnedLexiconTrain
				got := fmt.Sprintf("%x", digest.Sum(nil))
				if counts[2].strict != p.b1Strict || counts[2].any != p.b1Any || counts[3].strict != p.modelStrict || counts[3].any != p.modelAny || hits != p.hits || got != p.digest {
					t.Errorf("train lexicon output changed: B1-lexicon %d/%d model-lexicon %d/%d hits=%d digest=%s (update bnPinnedLexiconTrain deliberately and report the delta)", counts[2].strict, counts[2].any, counts[3].strict, counts[3].any, hits, got)
				}
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
