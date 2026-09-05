package hindi

import (
	_ "embed"
	"math"
	"strings"
	"sync"
)

// Character 4-gram re-ranker (T-0018). Candidates produced by different rule
// configurations are scored against a language model of how romanized Hindi
// actually looks, trained on Dakshina TRAIN romanizations only (held-out safe;
// tools/train_ngram.py). Scoring is stupid backoff (4→3→2→1-gram, factor 0.4),
// normalized per character so length differences don't dominate.

//go:embed roman_ngrams.tsv
var romanNgramsTSV string

const backoffFactor = 0.4

var (
	ngramOnce   sync.Once
	ngramCounts map[string]float64
	ngramTotal  float64 // total unigram mass, for the base case
)

func loadNgrams() {
	ngramOnce.Do(func() {
		ngramCounts = make(map[string]float64)
		for _, line := range strings.Split(romanNgramsTSV, "\n") {
			tab := strings.IndexByte(line, '\t')
			if tab <= 0 {
				continue
			}
			var n float64
			for _, c := range line[tab+1:] {
				if c < '0' || c > '9' {
					n = 0
					break
				}
				n = n*10 + float64(c-'0')
			}
			if n > 0 {
				ngramCounts[line[:tab]] = n
				if tab == 1 { // unigram
					ngramTotal += n
				}
			}
		}
	})
}

// charLogProb returns the stupid-backoff log-probability of the character at
// position i given up to 3 preceding characters.
func charLogProb(s string, i int) float64 {
	penalty := 1.0
	for order := 4; order >= 1; order-- {
		start := i - (order - 1)
		if start < 0 {
			continue
		}
		ctx, gram := s[start:i], s[start:i+1]
		var denom float64
		if order == 1 {
			denom = ngramTotal
		} else {
			denom = ngramCounts[ctx]
		}
		if denom > 0 {
			if num := ngramCounts[gram]; num > 0 {
				return math.Log(penalty * num / denom)
			}
		}
		penalty *= backoffFactor
	}
	return math.Log(penalty / (ngramTotal + 1)) // unseen even as unigram
}

// scoreRoman returns the mean per-character log-probability of a romanization.
func scoreRoman(s string) float64 {
	loadNgrams()
	w := "^" + strings.ToLower(s) + "$"
	var sum float64
	for i := 1; i < len(w); i++ {
		sum += charLogProb(w, i)
	}
	return sum / float64(len(w)-1)
}

// RerankRomans implements core.Reranker: given candidate romanizations of one
// word, return the one the character LM finds most like real romanized Hindi.
// Ties (and single candidates) keep the first — candidate order encodes the
// default preference, so the LM must strictly win to override it.
func (h Hindi) RerankRomans(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	best, bestScore := candidates[0], scoreRoman(candidates[0])
	for _, c := range candidates[1:] {
		if s := scoreRoman(c); s > bestScore {
			best, bestScore = c, s
		}
	}
	return best
}
