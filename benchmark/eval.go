package benchmark

// Evaluation metrics for transliteration accuracy.
//
// Romanization is many-to-one: a single Devanagari word has several valid Roman
// spellings (जनता → janata / janta / janataa). Scoring against one gold string
// under-counts correctness. These helpers support:
//   - CER (character error rate): normalized edit distance, credits near-misses.
//   - Multi-reference: match against ANY attested human variant (minCER / any-hit).
//
// Reference variants come from the full Dakshina lexicon (multiple attested rows
// per native word, each with an attestation count in its notes).

import (
	"os"
)

// levenshtein returns the edit distance between two rune slices.
func levenshtein(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// cer returns the character error rate of got against a single expected string:
// edit distance normalized by the length of expected (in runes). 0.0 == exact.
func cer(got, expected string) float64 {
	g, e := []rune(got), []rune(expected)
	denom := len(e)
	if denom == 0 {
		if len(g) == 0 {
			return 0
		}
		denom = len(g)
	}
	return float64(levenshtein(g, e)) / float64(denom)
}

// minCER returns the smallest CER of got over a set of reference spellings.
// With an empty reference set it returns 1.0 (treated as a full miss).
func minCER(got string, refs []string) float64 {
	if len(refs) == 0 {
		return 1.0
	}
	best := cer(got, refs[0])
	for _, r := range refs[1:] {
		if c := cer(got, r); c < best {
			best = c
		}
	}
	return best
}

// matchesAny reports whether got exactly equals any reference spelling.
func matchesAny(got string, refs []string) bool {
	for _, r := range refs {
		if got == r {
			return true
		}
	}
	return false
}

// loadReferenceSets builds native -> deduped list of attested Roman spellings
// from a lexicon CSV that may contain multiple rows per native word (Dakshina).
// Order is preserved by first appearance so results are deterministic.
func loadReferenceSets(path string) (map[string][]string, error) {
	entries, err := loadCSV(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]string), nil
		}
		return nil, err
	}
	refs := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	for _, e := range entries {
		if e.Roman == "" {
			continue
		}
		if seen[e.Native] == nil {
			seen[e.Native] = make(map[string]bool)
		}
		if !seen[e.Native][e.Roman] {
			seen[e.Native][e.Roman] = true
			refs[e.Native] = append(refs[e.Native], e.Roman)
		}
	}
	return refs, nil
}
