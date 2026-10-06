package gomanize

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// Update only on a deliberately reviewed baseline change, never to make a
// behavior-preserving refactor pass. The initial snapshot predates Part A.
var updateHindiSnapshot = flag.Bool("update-hindi-snapshot", false, "rewrite the frozen Hindi outputs")

const hindiSnapshotDir = "testdata/hindi_snapshot"

type snapshotProfile struct {
	name     string
	opts     Options
	disabled []string
}

func hindiSnapshotProfiles() []snapshotProfile {
	return []snapshotProfile{
		{name: "default"},
		{name: "schwa-model", opts: Options{SchwaModel: true}},
		{name: "lexicon", opts: Options{Lexicon: true}},
		{name: "rerank", opts: Options{Rerank: true}},
		{name: "keep-medial", opts: Options{KeepMedialSchwa: true}},
		{name: "long-vowels", opts: Options{LongVowels: true}},
		{name: "simple-nasals", opts: Options{SimpleNasals: true}},
		{name: "combined-style", opts: Options{KeepMedialSchwa: true, LongVowels: true, SimpleNasals: true}},
		{name: "learned-style", opts: Options{SchwaModel: true, LongVowels: true, SimpleNasals: true}},
		{name: "lexicon-rerank", opts: Options{Lexicon: true, Rerank: true}},
		{name: "no-schwa-rules", disabled: []string{"schwa.*"}},
		{name: "no-final-delete", disabled: []string{"schwa.delete.word-final"}},
		{name: "no-rules", disabled: []string{"*"}},
	}
}

type hindiSnapshotRow struct {
	Input   string   `json:"input"`
	Outputs []string `json:"outputs"`
}

// The manifest pins the frozen input set itself, not the benchmark files it
// was collected from: a gold/notes-only CSV edit must not force a regeneration.
type hindiSnapshotManifest struct {
	Profiles    []string `json:"profiles"`
	InputSHA256 string   `json:"input_sha256"`
	Inputs      int      `json:"inputs"`
}

// hindiSnapshotInputHash hashes length-prefixed inputs in order, so inputs
// containing newlines or tabs cannot collide.
func hindiSnapshotInputHash(inputs []string) string {
	h := sha256.New()
	for _, s := range inputs {
		fmt.Fprintf(h, "%d:%s", len(s), s)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func hindiSnapshotInputs(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob("benchmark/data/*_hi.csv")
	if err != nil || len(paths) == 0 {
		t.Fatalf("snapshot sources: %v (%d files)", err, len(paths))
	}
	seen := map[string]bool{}
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		reader := csv.NewReader(f)
		reader.FieldsPerRecord = -1
		if _, err = reader.Read(); err != nil {
			t.Fatal(err)
		}
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(row) > 0 && row[0] != "" {
				seen[row[0]] = true
			}
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	// Include the invariant suite's construct space, rare independent vowels,
	// nukta/conjunct starts, and public sentence segmentation boundaries.
	for _, c := range invConsonants {
		seen[c] = true
		for _, v := range invVowelForms {
			seen[c+v.matra] = true
			seen[c+v.indep] = true
		}
	}
	for _, gold := range []map[string][]string{chandrabinduNasalGold, chandrabinduLabialGold} {
		for input := range gold {
			seen[input] = true
		}
	}
	for _, s := range []string{"गई", "कऋ", "दरअसल", "ऄ", "कऄ", "ज्ञकरन", "क़करन", "जबरदस्त", "चाँद", "कहाँ", "माँ", "हाँ", "", "\nभारत\tभारत।", "नमस्ते  दुनिया"} {
		seen[s] = true
	}
	inputs := make([]string, 0, len(seen))
	for s := range seen {
		inputs = append(inputs, s)
	}
	sort.Strings(inputs)
	return inputs
}

func TestHindiFrozenSnapshot(t *testing.T) {
	inputs := hindiSnapshotInputs(t)
	profiles := hindiSnapshotProfiles()
	names := make([]string, len(profiles))
	engines := make([]*Gomanize, len(profiles))
	for i, p := range profiles {
		names[i] = p.name
		var err error
		engines[i], err = NewWithOptions("hindi", p.opts, WithDisabledRules(p.disabled...))
		if err != nil {
			t.Fatal(err)
		}
	}
	if *updateHindiSnapshot {
		manifest := hindiSnapshotManifest{Profiles: names, InputSHA256: hindiSnapshotInputHash(inputs), Inputs: len(inputs)}
		writeHindiSnapshot(t, inputs, engines, manifest)
		return
	}
	data, err := os.ReadFile(filepath.Join(hindiSnapshotDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var frozen hindiSnapshotManifest
	if err = json.Unmarshal(data, &frozen); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(frozen.Profiles, names) {
		t.Fatalf("snapshot profiles changed: frozen %v, current %v; review before deliberately regenerating", frozen.Profiles, names)
	}
	paths, err := filepath.Glob(filepath.Join(hindiSnapshotDir, "*.jsonl.gz"))
	if err != nil {
		t.Fatal(err)
	}
	// Replay the frozen rows themselves; the current corpus only feeds the
	// coverage note below, so benchmark edits never invalidate the snapshot.
	frozenInputs := make([]string, 0, frozen.Inputs)
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		zr, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(zr)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			var row hindiSnapshotRow
			if err = json.Unmarshal(scanner.Bytes(), &row); err != nil {
				t.Fatal(err)
			}
			count := len(frozenInputs)
			if len(row.Outputs) != len(profiles) || (count > 0 && row.Input <= frozenInputs[count-1]) {
				t.Fatalf("invalid snapshot row %d in %s", count, path)
			}
			for i, g := range engines {
				if got := g.Translit(row.Input); got != row.Outputs[i] {
					t.Fatalf("%s: %q: got %q, frozen %q", names[i], row.Input, got, row.Outputs[i])
				}
			}
			frozenInputs = append(frozenInputs, row.Input)
		}
		if err = scanner.Err(); err != nil {
			t.Fatal(err)
		}
		if err = zr.Close(); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if len(frozenInputs) != frozen.Inputs || hindiSnapshotInputHash(frozenInputs) != frozen.InputSHA256 {
		t.Fatalf("snapshot rows (%d) do not match manifest (%d inputs, sha256 %s)", len(frozenInputs), frozen.Inputs, frozen.InputSHA256)
	}
	frozenSet := make(map[string]bool, len(frozenInputs))
	for _, s := range frozenInputs {
		frozenSet[s] = true
	}
	uncovered := 0
	for _, s := range inputs {
		if !frozenSet[s] {
			uncovered++
		}
	}
	if uncovered > 0 {
		t.Logf("note: %d current corpus inputs are not in the frozen snapshot (not a failure)", uncovered)
	}
	t.Logf("verified %d inputs across %d profiles (%d exact outputs)", len(frozenInputs), len(profiles), len(frozenInputs)*len(profiles))
}

func writeHindiSnapshot(t *testing.T, inputs []string, engines []*Gomanize, manifest hindiSnapshotManifest) {
	t.Helper()
	if err := os.MkdirAll(hindiSnapshotDir, 0755); err != nil {
		t.Fatal(err)
	}
	old, err := filepath.Glob(filepath.Join(hindiSnapshotDir, "*.jsonl.gz"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range old {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	const chunk = 10000
	for start := 0; start < len(inputs); start += chunk {
		end := start + chunk
		if end > len(inputs) {
			end = len(inputs)
		}
		path := filepath.Join(hindiSnapshotDir, fmt.Sprintf("%04d.jsonl.gz", start/chunk))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		zw := gzip.NewWriter(f)
		enc := json.NewEncoder(zw)
		for _, input := range inputs[start:end] {
			row := hindiSnapshotRow{Input: input, Outputs: make([]string, len(engines))}
			for i, g := range engines {
				row.Outputs[i] = g.Translit(input)
			}
			if err = enc.Encode(row); err != nil {
				t.Fatal(err)
			}
		}
		if err = zw.Close(); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(hindiSnapshotDir, "manifest.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("froze %d inputs across %d profiles", len(inputs), len(engines))
}
