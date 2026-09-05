package hindi

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/budhash/gomanize/core"
)

// Learned schwa-deletion classifier. A shallow CART decision tree trained on
// force-aligned Dakshina data (tools/schwa/train.py) and embedded here as JSON.
// Held-out per-schwa accuracy ~90.7% on the disjoint Dakshina test split; see
// docs/reviews/2026-09-04-h3-schwa-classifier.md.
//
// IMPORTANT: feature extraction here MUST stay identical to tools/schwa/features.py.
// Both compute features from the raw Devanagari rune sequence so training and
// inference agree exactly.

//go:embed schwa_tree.json
var schwaTreeJSON []byte

const (
	nuktaRune  = '़'
	halantRune = '्'
)

// consonantBaseRunes mirrors the CONS keys in tools/schwa/{align,train}.py
// (base characters, nukta handled separately).
var consonantBaseRunes = map[rune]bool{
	'क': true, 'ख': true, 'ग': true, 'घ': true, 'ङ': true, 'च': true, 'छ': true, 'ज': true, 'झ': true, 'ञ': true,
	'ट': true, 'ठ': true, 'ड': true, 'ढ': true, 'ण': true, 'त': true, 'थ': true, 'द': true, 'ध': true, 'न': true,
	'प': true, 'फ': true, 'ब': true, 'भ': true, 'म': true, 'य': true, 'र': true, 'ल': true, 'व': true, 'श': true,
	'ष': true, 'स': true, 'ह': true, 'ळ': true,
}

// matraRunes mirrors the MATRA set in the Python tools (dependent vowel signs +
// nasal/visarga that suppress or replace the inherent schwa).
var matraRunes = map[rune]bool{
	'ा': true, 'ि': true, 'ी': true, 'ु': true, 'ू': true, 'ृ': true,
	'े': true, 'ै': true, 'ो': true, 'ौ': true, 'ं': true, 'ः': true, 'ँ': true, 'ॉ': true,
}

func isConsonantRune(r rune) bool { return consonantBaseRunes[r] }

// schwaModelDecision reports, for a consonant unit u, whether the learned model
// applies (u carries an inherent schwa in the training sense) and if so whether
// that schwa should be deleted. Features are computed from the raw input runes
// exactly as tools/schwa/features.py does.
func schwaModelDecision(w *core.Word, u *core.Unit) (delete bool, applies bool) {
	runes := []rune(w.Original)
	i := u.Start.Rune
	if i < 0 || i >= len(runes) || !isConsonantRune(runes[i]) {
		return false, false
	}
	// Determine the rune after the consonant (skipping its own nukta).
	next := i + 1
	if next < len(runes) && runes[next] == nuktaRune {
		next++
	}
	// Inherent schwa only: not a conjunct (halant) and not followed by a matra.
	if next < len(runes) {
		if runes[next] == halantRune || matraRunes[runes[next]] {
			return false, false
		}
	}
	isFirst := true
	for k := 0; k < i; k++ {
		if isConsonantRune(runes[k]) {
			isFirst = false
			break
		}
	}
	isLast := true
	for k := next; k < len(runes); k++ {
		if isConsonantRune(runes[k]) {
			isLast = false
			break
		}
	}
	return predictSchwaDelete(runes, i, isFirst, isLast), true
}

// treeNode is either an internal split (F/V/Yes/No) or a leaf (Leaf set, IsLeaf).
type treeNode struct {
	F      string    `json:"f"`
	V      string    `json:"v"`
	Yes    *treeNode `json:"yes"`
	No     *treeNode `json:"no"`
	Leaf   *int      `json:"leaf"`
	IsLeaf bool      `json:"-"`
}

type schwaTree struct {
	Features []string  `json:"features"`
	Tree     *treeNode `json:"tree"`
}

var (
	schwaModelOnce sync.Once
	schwaModel     *schwaTree
	schwaModelErr  error
)

func loadSchwaModel() (*schwaTree, error) {
	schwaModelOnce.Do(func() {
		var m schwaTree
		if err := json.Unmarshal(schwaTreeJSON, &m); err != nil {
			schwaModelErr = err
			return
		}
		markLeaves(m.Tree)
		schwaModel = &m
	})
	return schwaModel, schwaModelErr
}

func markLeaves(n *treeNode) {
	if n == nil {
		return
	}
	if n.Leaf != nil {
		n.IsLeaf = true
		return
	}
	markLeaves(n.Yes)
	markLeaves(n.No)
}

// schwaFeatures computes the feature map for the inherent-schwa consonant whose
// consonant character begins at rune index i in runes. Mirrors features.py.
func schwaFeatures(runes []rune, i int, isFirst, isLast bool) map[string]string {
	cons := string(runes[i])
	next := i + 1
	if next < len(runes) && runes[next] == nuktaRune {
		cons += string(runes[next])
		next++
	}
	prev := "^"
	if i > 0 {
		prev = string(runes[i-1])
	}
	nx := "$"
	if next < len(runes) {
		nx = string(runes[next])
	}
	nx2 := "$"
	if next+1 < len(runes) {
		nx2 = string(runes[next+1])
	}
	first, last := "0", "0"
	if isFirst {
		first = "1"
	}
	if isLast {
		last = "1"
	}
	return map[string]string{
		"cons": cons, "prev": prev, "next": nx, "next2": nx2,
		"first": first, "last": last,
	}
}

// predictSchwaDelete returns true if the model predicts the inherent schwa at
// rune index i should be deleted. Falls back to true (delete) on load error,
// matching the majority class.
func predictSchwaDelete(runes []rune, i int, isFirst, isLast bool) bool {
	m, err := loadSchwaModel()
	if err != nil || m == nil {
		return true
	}
	feats := schwaFeatures(runes, i, isFirst, isLast)
	n := m.Tree
	for n != nil && !n.IsLeaf {
		if feats[n.F] == n.V {
			n = n.Yes
		} else {
			n = n.No
		}
	}
	if n == nil || n.Leaf == nil {
		return true
	}
	return *n.Leaf == 1
}
