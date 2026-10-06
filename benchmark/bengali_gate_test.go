package benchmark

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

// Immutable B0 development outputs captured at 6a82b6a before B1 tuning.
// Never regenerate this artifact to accommodate changed pronunciation rules.
const bnB0DevSHA = "20f8a790aebc41fbf7f2af6a58bce3ba3f6770f97e910833f703c57c2d571a3f"
const bnGateDevFixtureSHA = "8e83784102a527a88af0ad6dfba0fa41c11f57633a34fe2a9abe12acb54aa51a"

func bnCurated(w bnWord) bool {
	for _, r := range w.refs {
		if r.votes >= 3 {
			return true
		}
	}
	return false
}

func bnGateFailures(base, candidate bnCounts) []string {
	var failures []string
	if base.words == 0 || candidate.words != base.words {
		return []string{"invalid/mismatched word count"}
	}
	required := base.any + int(math.Ceil(float64(base.words-base.any)*0.10))
	if candidate.any < required {
		failures = append(failures, fmt.Sprintf("match-any %d < %d", candidate.any, required))
	}
	if candidate.strict < base.strict {
		failures = append(failures, fmt.Sprintf("strict %d < %d", candidate.strict, base.strict))
	}
	if candidate.minCER > base.minCER*0.95+1e-12 {
		failures = append(failures, fmt.Sprintf("macro minCER %.8f > %.8f", candidate.minCER/float64(base.words), base.minCER*0.95/float64(base.words)))
	}
	return failures
}

func TestBengaliGateDefinitions(t *testing.T) {
	base := bnCounts{words: 2500, strict: 698, any: 1210, minCER: 250}
	pass := bnCounts{words: 2500, strict: 698, any: 1339, minCER: 237.5}
	if f := bnGateFailures(base, pass); len(f) != 0 {
		t.Fatal(f)
	}
	for _, bad := range []bnCounts{base, {words: 2500, strict: 697, any: 1339, minCER: 237.5}, {words: 2500, strict: 698, any: 1338, minCER: 237.5}, {words: 2500, strict: 698, any: 1339, minCER: 237.5001}, {words: 2499, strict: 698, any: 1339, minCER: 200}} {
		if len(bnGateFailures(base, bad)) == 0 {
			t.Fatalf("accepted failing candidate: %+v", bad)
		}
	}
	// Round the secondary subset's 44.7-word requirement upward.
	if f := bnGateFailures(bnCounts{words: 992, strict: 424, any: 545, minCER: 100}, bnCounts{words: 992, strict: 424, any: 589, minCER: 95}); len(f) == 0 {
		t.Fatal("fractional target rounded down")
	}
	for _, tc := range []struct {
		refs []bnReference
		want bool
	}{
		{[]bnReference{{"a", 2}, {"b", 2}}, false},
		{[]bnReference{{"a", 3}, {"b", 1}}, true},
	} {
		if got := bnCurated(bnWord{refs: tc.refs}); got != tc.want {
			t.Fatal("curation must use maximum reference votes, not total")
		}
	}
}

// Capture is an explicit one-time maintenance path, never invoked by CI. Run
// only at the B0 commit recorded above; the destination must not already exist.
func bnCaptureDev(t *testing.T, path string, words []bnWord, engine *gomanize.Gomanize) {
	t.Helper()
	var buf bytes.Buffer
	z := gzip.NewWriter(&buf)
	c := csv.NewWriter(z)
	rows := [][]string{{"native", "b0_default"}}
	for _, w := range words {
		rows = append(rows, []string{w.native, engine.Translit(w.native)})
	}
	if err := c.WriteAll(rows); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())) != bnB0DevSHA {
		t.Fatal("refusing to capture changed outputs as the frozen B0 baseline; use the recorded B0 commit")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("captured %d dev outputs sha256=%x", len(words), sha256.Sum256(buf.Bytes()))
}

func TestBengaliDevGate(t *testing.T) {
	fixture, err := os.ReadFile(getTestDataPath("bengali/dev.csv.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(fixture)) != bnGateDevFixtureSHA {
		t.Fatal("gate dev references changed")
	}
	words := loadBengali(t, "dev")["dev"]
	engine, err := gomanize.New("bengali")
	if err != nil {
		t.Fatal(err)
	}
	if path := os.Getenv("BENGALI_CAPTURE_DEV"); path != "" {
		bnCaptureDev(t, path, words, engine)
		return
	}
	packed, err := os.ReadFile(getTestDataPath("bengali/b0-dev.csv.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(packed)); got != bnB0DevSHA {
		t.Fatalf("B0 dev snapshot hash changed: %s", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(z).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(words)+1 || strings.Join(rows[0], ",") != "native,b0_default" {
		t.Fatal("bad B0 snapshot dimensions/header")
	}
	var base, candidate [2]bnCounts
	var wins, losses [2]int
	for i, w := range words {
		if len(rows[i+1]) != 2 || rows[i+1][0] != w.native {
			t.Fatal("B0 keys differ from frozen dev fixture")
		}
		before, after := rows[i+1][1], engine.Translit(w.native)
		for j := 0; j < 2; j++ {
			if j == 1 && !bnCurated(w) {
				continue
			}
			var b, c bnCounts
			b.add(before, w.refs)
			c.add(after, w.refs)
			base[j].add(before, w.refs)
			candidate[j].add(after, w.refs)
			if b.any == 0 && c.any == 1 {
				wins[j]++
			}
			if b.any == 1 && c.any == 0 {
				losses[j]++
			}
		}
	}
	for i, label := range []string{"full-dev", "curated-dev-max-votes>=3"} {
		b, c := base[i], candidate[i]
		t.Logf("%s words=%d baseline strict=%d any=%d CER=%.10f current strict=%d any=%d CER=%.10f wins=%d losses=%d", label, b.words, b.strict, b.any, b.minCER/float64(b.words), c.strict, c.any, c.minCER/float64(c.words), wins[i], losses[i])
		f := bnGateFailures(b, c)
		if len(f) > 0 {
			if os.Getenv("BENGALI_REQUIRE_B1") != "" {
				t.Errorf("B1 gate %s: %s", label, strings.Join(f, "; "))
			} else {
				t.Logf("B1 NOT READY (%s): %s; enforce with make test-bengali-b1-gate", label, strings.Join(f, "; "))
			}
		} else {
			t.Logf("B1 numerical gate passes (%s); construct/trace and Hindi checks remain mandatory", label)
		}
	}
}
