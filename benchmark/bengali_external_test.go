package benchmark

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

// This is a descriptive external transfer score, not the B1 acceptance gate.
// Human Roman sentences can differ in spelling, spacing and punctuation; do not
// mine this test split for pronunciation rules or word exceptions.
func TestBenchmarkBengaliBanglaTLit(t *testing.T) {
	raw, err := os.ReadFile(getTestDataPath("banglatlit/manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Hash string `json:"fixture_sha256"`
		Rows int    `json:"rows"`
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	blob, err := os.ReadFile(getTestDataPath("banglatlit/test.csv.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(blob)) != manifest.Hash {
		t.Fatal("BanglaTLit fixture hash mismatch")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
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
	if len(rows) != manifest.Rows+1 || len(rows[0]) != 3 || rows[0][0] != "id" || rows[0][1] != "native" || rows[0][2] != "roman" {
		t.Fatal("bad external fixture schema/count")
	}
	before, err := gomanize.NewWithOptions("bengali", gomanize.Options{}, gomanize.WithDisabledRules("schwa.bengali.*", "consonant.bengali.*", "vowel.bengali.*", "render.bengali.*"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := gomanize.New("bengali")
	if err != nil {
		t.Fatal(err)
	}
	if before.Translit("মধ্য") != "modhj" || after.Translit("মধ্য") != "moddho" {
		t.Fatal("external comparison did not select distinct B0/B1 catalogs")
	}
	seen := map[string]bool{}
	natives := map[string]bool{}
	var baseline, current bnCounts
	for _, row := range rows[1:] {
		if len(row) != 3 || seen[row[0]] || row[1] == "" || row[2] == "" {
			t.Fatal("bad or duplicate external row")
		}
		seen[row[0]] = true
		natives[row[1]] = true
		refs := []bnReference{{roman: row[2], votes: 1}}
		baseline.add(before.Translit(row[1]), refs)
		current.add(after.Translit(row[1]), refs)
	}
	t.Logf("BanglaTLit official test: %d rows, %d normalized native sentences; single Roman reference per row; no tuning", current.words, len(natives))
	t.Logf("B0 exact=%d/%d macro-CER=%.8f; B1 exact=%d/%d macro-CER=%.8f", baseline.any, baseline.words, baseline.minCER/float64(baseline.words), current.any, current.words, current.minCER/float64(current.words))
}
