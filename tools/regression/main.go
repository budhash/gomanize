// Command regression is the before/after transition-matrix diff for gomanize
// engine changes (F-0010 / T-0043, per the Codex design review in
// docs/reviews/2026-09-08-structured-parser-qa.md).
//
// Aggregate accuracy is an average and can hide a small-but-severe structural
// regression behind many low-value wins. This tool classifies every token whose
// output CHANGED between a baseline snapshot and the current engine, against the
// multi-reference sets, so a fix can be shown net-positive with every exact
// regression individually inspected.
//
//	make regression-baseline   # snapshot current (pre-change) outputs
//	# … make the engine change …
//	make regression-diff       # classify current vs the snapshot
//
// Evaluation is rules-only (default options): the lexicon/schwa-model/reranker
// are Dakshina-train-derived, so scoring them here would be circular — a parser
// fix is judged on the rules path and checked in the other modes separately.
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	gomanize "github.com/budhash/gomanize"
)

const (
	dataPath   = "benchmark/data/dakshina_hi.csv"
	cerEps     = 0.001 // ignore trivial minCER wobble
	lexicalCER = 0.5   // both-miss with best minCER above this ⇒ likely loanword/acronym
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: regression (dump|diff) <path>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "dump":
		dump(os.Args[2])
	case "diff":
		os.Exit(diff(os.Args[2]))
	default:
		fmt.Fprintln(os.Stderr, "unknown subcommand:", os.Args[1])
		os.Exit(2)
	}
}

func newEngine() *gomanize.Gomanize {
	g, err := gomanize.New("hindi") // default options = rules only
	if err != nil {
		fmt.Fprintln(os.Stderr, "engine:", err)
		os.Exit(1)
	}
	return g
}

// loadRefs returns native -> deduped roman variants from the Dakshina CSV.
func loadRefs(path string) map[string][]string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open data:", err)
		os.Exit(1)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "read data:", err)
		os.Exit(1)
	}
	refs := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	for i, row := range rows {
		if i == 0 || len(row) < 2 || row[0] == "" || row[1] == "" {
			continue
		}
		n, rom := row[0], row[1]
		if seen[n] == nil {
			seen[n] = map[string]bool{}
		}
		if !seen[n][rom] {
			seen[n][rom] = true
			refs[n] = append(refs[n], rom)
		}
	}
	return refs
}

func sortedNatives(refs map[string][]string) []string {
	ns := make([]string, 0, len(refs))
	for n := range refs {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	return ns
}

func dump(out string) {
	refs := loadRefs(dataPath)
	g := newEngine()
	f, err := os.Create(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create:", err)
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	n := 0
	for _, native := range sortedNatives(refs) {
		fmt.Fprintf(w, "%s\t%s\n", native, g.Translit(native))
		n++
	}
	fmt.Printf("wrote %d rows to %s\n", n, out)
}

func loadTSV(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open baseline:", err, "(run `make regression-baseline` first)")
		os.Exit(1)
	}
	defer f.Close()
	m := make(map[string]string)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func matchesAny(got string, refs []string) bool {
	for _, r := range refs {
		if got == r {
			return true
		}
	}
	return false
}

func minCER(got string, refs []string) float64 {
	g := []rune(got)
	best := 1.0
	first := true
	for _, ref := range refs {
		rr := []rune(ref)
		denom := len(rr)
		if denom == 0 {
			denom = 1
		}
		c := float64(lev(g, rr)) / float64(denom)
		if first || c < best {
			best, first = c, false
		}
	}
	return best
}

func lev(a, b []rune) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func diff(base string) int {
	refs := loadRefs(dataPath)
	before := loadTSV(base)
	g := newEngine()

	order := []string{"improved_exact", "regressed_exact", "variant_drift", "improved_distance", "regressed_distance", "lexical_unresolved", "neutral_distance"}
	counts := map[string]int{}
	var regressions, drifts []string
	changed := 0

	for _, native := range sortedNatives(refs) {
		bOut, ok := before[native]
		if !ok {
			continue // token not in the baseline snapshot
		}
		aOut := g.Translit(native)
		if aOut == bOut {
			continue // unchanged — the vast majority
		}
		changed++
		rs := refs[native]
		bHit, aHit := matchesAny(bOut, rs), matchesAny(aOut, rs)
		switch {
		case !bHit && aHit:
			counts["improved_exact"]++
		case bHit && !aHit:
			counts["regressed_exact"]++
			regressions = append(regressions, fmt.Sprintf("%s: %q → %q (was a hit; refs %v)", native, bOut, aOut, capRefs(rs)))
		case bHit && aHit:
			counts["variant_drift"]++
			if len(drifts) < 15 {
				drifts = append(drifts, fmt.Sprintf("%s: %q → %q (both attested)", native, bOut, aOut))
			}
		default: // both miss
			bC, aC := minCER(bOut, rs), minCER(aOut, rs)
			switch {
			case bC > lexicalCER && aC > lexicalCER:
				counts["lexical_unresolved"]++
			case aC < bC-cerEps:
				counts["improved_distance"]++
			case aC > bC+cerEps:
				counts["regressed_distance"]++
			default:
				counts["neutral_distance"]++
			}
		}
	}

	fmt.Println("========================================")
	fmt.Println("PARSER REGRESSION DIFF (rules-only, vs baseline)")
	fmt.Println("========================================")
	fmt.Printf("baseline: %s  |  changed tokens: %d\n\n", base, changed)
	for _, k := range order {
		fmt.Printf("  %-20s %6d\n", k, counts[k])
	}
	net := counts["improved_exact"] - counts["regressed_exact"]
	fmt.Printf("\n  net exact (improved - regressed): %+d\n", net)

	if len(drifts) > 0 {
		fmt.Printf("\n--- variant_drift (hit→hit, output changed; review) ---\n")
		for _, d := range drifts {
			fmt.Println("  " + d)
		}
	}
	if len(regressions) > 0 {
		fmt.Printf("\n--- regressed_exact (MUST inspect) ---\n")
		for _, r := range regressions {
			fmt.Println("  " + r)
		}
		fmt.Printf("\nFAIL: %d exact regression(s). Inspect each; ship only if intended.\n", counts["regressed_exact"])
		return 1
	}
	fmt.Printf("\nOK: no exact regressions.\n")
	return 0
}

func capRefs(rs []string) []string {
	if len(rs) > 4 {
		return rs[:4]
	}
	return rs
}
