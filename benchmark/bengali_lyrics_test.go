package benchmark

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"testing"

	gomanize "github.com/budhash/gomanize"
)

// Reference agreement is diagnostic only: this pilot is not human-attested gold.
func TestBenchmarkBengaliLyricsPilot(t *testing.T) {
	blob, err := os.ReadFile("data/bengali_lyrics/pilot.csv")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(blob)); got != "44b33280eb0cdd342d12a58e083f008953b83a1379a9e5c724d39a7fedf424ae" {
		t.Fatal("pilot reference snapshot changed; review provenance before updating")
	}
	rows, err := csv.NewReader(strings.NewReader(string(blob))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 83 || strings.Join(rows[0], ",") != "id,song,line,native,roman,reference_status" {
		t.Fatal("unexpected pilot schema or line count")
	}
	for _, tc := range []struct {
		name string
		opts gomanize.Options
	}{
		{"B1", gomanize.Options{}},
		{"model", gomanize.Options{SchwaModel: true}},
		{"lexicon", gomanize.Options{Lexicon: true}},
		{"model_lexicon", gomanize.Options{SchwaModel: true, Lexicon: true}},
		{"rerank", gomanize.Options{SchwaModel: true, Rerank: true}},
		{"rerank_lexicon", gomanize.Options{SchwaModel: true, Rerank: true, Lexicon: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := gomanize.NewWithOptions("bengali", tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			var native, outputs []string
			var totalCER float64
			exact := 0
			for _, row := range rows[1:] {
				if row[5] != "assistant_draft_unreviewed" {
					t.Fatal("pilot status changed")
				}
				out := e.Translit(row[3])
				native = append(native, row[3])
				outputs = append(outputs, out)
				totalCER += cer(out, row[4])
				if out == row[4] {
					exact++
				}
			}
			if e.Translit(strings.Join(native, "\n")) != strings.Join(outputs, "\n") {
				t.Fatal("multiline lyrics changed line boundaries or per-line rendering")
			}
			t.Logf("UNREVIEWED reference agreement: %d/82 exact; macro line CER %.8f (no quality gate)", exact, totalCER/82)
		})
	}
}
