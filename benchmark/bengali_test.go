package benchmark

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

type bnReference struct {
	roman string
	votes int
}
type bnWord struct {
	native string
	refs   []bnReference
}
type bnCounts struct {
	words, strict, any int
	minCER             float64
}

func (c *bnCounts) add(got string, refs []bnReference) {
	c.words++
	best := refs[0]
	romans := make([]string, len(refs))
	for i, r := range refs {
		romans[i] = r.roman
		if r.votes > best.votes || (r.votes == best.votes && r.roman < best.roman) {
			best = r
		}
	}
	if got == best.roman {
		c.strict++
	}
	if matchesAny(got, romans) {
		c.any++
	}
	c.minCER += minCER(got, romans)
}

func bnBucket(n int) string {
	if n >= 4 {
		return "4+"
	}
	return strconv.Itoa(n)
}

// Candidate support is conditional on this deliberately naive B0 engine. A
// neither-match does not tell us which inherent-vowel spelling humans prefer.
func bnStyleSupport(o, a string, refs []bnReference) string {
	if o == a {
		return "same-candidate"
	}
	hasO, hasA := false, false
	for _, r := range refs {
		hasO = hasO || r.roman == o
		hasA = hasA || r.roman == a
	}
	if hasO && hasA {
		return "both"
	}
	if hasO {
		return "o-only"
	}
	if hasA {
		return "a-only"
	}
	return "neither"
}

func loadBengali(t *testing.T, splits ...string) map[string][]bnWord {
	t.Helper()
	dir := getTestDataPath("bengali")
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Splits map[string]struct {
			Hash  string `json:"fixture_sha256"`
			Words int    `json:"normalized_words"`
			Rows  int    `json:"normalized_rows"`
		} `json:"splits"`
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(splits) == 0 {
		splits = []string{"train", "dev", "test"}
	}
	result := map[string][]bnWord{}
	seen := map[string]string{}
	for _, split := range splits {
		meta, ok := manifest.Splits[split]
		if !ok {
			t.Fatalf("missing %s manifest", split)
		}
		packed, err := os.ReadFile(filepath.Join(dir, split+".csv.gz"))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(packed)) != meta.Hash {
			t.Fatalf("%s fixture hash mismatch", split)
		}
		zr, err := gzip.NewReader(bytes.NewReader(packed))
		if err != nil {
			t.Fatal(err)
		}
		cr := csv.NewReader(zr)
		header, err := cr.Read()
		if err != nil || len(header) != 3 || header[0] != "native" || header[1] != "roman" || header[2] != "attestations" {
			t.Fatalf("%s bad header: %v %v", split, header, err)
		}
		words := map[string][]bnReference{}
		rows := 0
		for {
			row, err := cr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			n, err := strconv.Atoi(row[2])
			if err != nil || n < 1 || row[0] == "" || row[1] == "" {
				t.Fatalf("%s invalid row %d", split, rows)
			}
			if prior, ok := seen[row[0]]; ok && prior != split {
				t.Fatalf("native word overlaps %s and %s", prior, split)
			}
			for _, r := range words[row[0]] {
				if r.roman == row[1] {
					t.Fatal("duplicate reference")
				}
			}
			seen[row[0]] = split
			words[row[0]] = append(words[row[0]], bnReference{row[1], n})
			rows++
		}
		if err = zr.Close(); err != nil {
			t.Fatal(err)
		}
		if len(words) != meta.Words || rows != meta.Rows {
			t.Fatalf("%s counts differ from manifest", split)
		}
		keys := make([]string, 0, len(words))
		for native := range words {
			keys = append(keys, native)
		}
		sort.Strings(keys)
		for _, native := range keys {
			result[split] = append(result[split], bnWord{native, words[native]})
		}
	}
	return result
}

// bnPinnedRules pins the rules-only (no learned components) train/dev outputs.
// Scores alone can hide compensating changes, so a digest over every
// default-o and alternate-a output is pinned too. A PR that changes Bengali
// rules output must update these deliberately and report the before/after
// delta. Test-split scores stay log-only: they never gate or guide changes.
var bnPinnedRules = map[string]struct {
	strict, any int
	digest      string
}{
	"train": {7044, 12080, "eadbc03efc070d1ab69fdf134e0738e80a21d4a8352bff7d94825fdc12b6e9dd"},
	"dev":   {698, 1210, "9361f262872291e3ac960837a40246338a367e6287b15948ac0d448586171a49"},
}

func TestBenchmarkBengali(t *testing.T) {
	datasets := loadBengali(t)
	o, err := gomanize.New("bengali")
	if err != nil {
		t.Fatal(err)
	}
	a, err := gomanize.NewWithOptions("bengali", gomanize.Options{InherentVowelA: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, split := range []string{"train", "dev", "test"} {
		t.Run(split, func(t *testing.T) {
			var defaultScore, altScore bnCounts
			refHist, voteHist, topVoteHist, styles := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
			byRefs, byVotes := map[string]*bnCounts{}, map[string]*bnCounts{}
			refsTotal := 0
			digest := sha256.New()
			for _, w := range datasets[split] {
				out, alt := o.Translit(w.native), a.Translit(w.native)
				if out == "" {
					t.Fatalf("empty output for %q", w.native)
				}
				fmt.Fprintf(digest, "%s\t%s\t%s\n", w.native, out, alt)
				defaultScore.add(out, w.refs)
				altScore.add(alt, w.refs)
				styles[bnStyleSupport(out, alt, w.refs)]++
				refBucket := bnBucket(len(w.refs))
				refHist[refBucket]++
				refsTotal += len(w.refs)
				topVotes := 0
				for _, r := range w.refs {
					voteHist[bnBucket(r.votes)]++
					if r.votes > topVotes {
						topVotes = r.votes
					}
				}
				voteBucket := bnBucket(topVotes)
				topVoteHist[voteBucket]++
				if byRefs[refBucket] == nil {
					byRefs[refBucket] = &bnCounts{}
				}
				byRefs[refBucket].add(out, w.refs)
				if byVotes[voteBucket] == nil {
					byVotes[voteBucket] = &bnCounts{}
				}
				byVotes[voteBucket].add(out, w.refs)
			}
			logScore := func(name string, c bnCounts) {
				t.Logf("%s: words=%d strict=%d (%.2f%%) match-any=%d (%.2f%%) macro-minCER=%.5f", name, c.words, c.strict, 100*float64(c.strict)/float64(c.words), c.any, 100*float64(c.any)/float64(c.words), c.minCER/float64(c.words))
			}
			logScore("Current default-o", defaultScore)
			logScore("Current alternate-a", altScore)
			t.Logf("references=%d mean-refs/word=%.4f", refsTotal, float64(refsTotal)/float64(defaultScore.words))
			for _, bucket := range []string{"1", "2", "3", "4+"} {
				t.Logf("bucket=%s ref-count-words=%d variant-attestations=%d best-attestation-words=%d", bucket, refHist[bucket], voteHist[bucket], topVoteHist[bucket])
				if c := byRefs[bucket]; c != nil {
					logScore("reference-count="+bucket, *c)
				}
				if c := byVotes[bucket]; c != nil {
					logScore("best-attestation="+bucket, *c)
				}
			}
			if pin, ok := bnPinnedRules[split]; ok {
				got := fmt.Sprintf("%x", digest.Sum(nil))
				if defaultScore.strict != pin.strict || defaultScore.any != pin.any || got != pin.digest {
					t.Errorf("%s rules output changed: strict=%d match-any=%d digest=%s; pinned strict=%d match-any=%d digest=%s (update bnPinnedRules deliberately and report the delta)", split, defaultScore.strict, defaultScore.any, got, pin.strict, pin.any, pin.digest)
				}
			}
			if split != "test" {
				for _, label := range []string{"o-only", "a-only", "both", "neither", "same-candidate"} {
					t.Logf("train/dev candidate-support %s=%d", label, styles[label])
				}
			}
			// Test split is evaluation-only: no examples or tuning histograms are mined.
		})
	}
}

func TestBengaliMetricDefinitions(t *testing.T) {
	refs := []bnReference{{"b", 2}, {"a", 2}, {"aa", 1}}
	var c bnCounts
	c.add("b", refs)
	if c.words != 1 || c.strict != 0 || c.any != 1 || c.minCER != 0 {
		t.Fatalf("top-1 tie must pick lexically first reference: %+v", c)
	}
	for _, tt := range []struct{ o, a, want string }{{"b", "a", "both"}, {"b", "x", "o-only"}, {"x", "a", "a-only"}, {"x", "y", "neither"}, {"b", "b", "same-candidate"}} {
		if got := bnStyleSupport(tt.o, tt.a, refs); got != tt.want {
			t.Fatalf("%+v: %s", tt, got)
		}
	}
}
