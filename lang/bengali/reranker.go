package bengali

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// Frozen cross-fitted selector, derived from Dakshina and Google Bengali data.
// CC BY-SA 4.0; see RERANKER.md. No dev/test training or runtime retraining.
//
//go:embed selector.json
var selectorJSON []byte

const selectorThreshold = 0.6
const selectorEdits = 8

type selectorNode struct {
	Counts  []int         `json:"counts"`
	Feature string        `json:"feature"`
	Value   string        `json:"value"`
	Yes     *selectorNode `json:"yes"`
	No      *selectorNode `json:"no"`
}

var selectorOnce sync.Once
var learnedSelector *selectorNode

func validSelector(n *selectorNode, depth int) bool {
	if n == nil || depth > 8 || len(n.Counts) != 2 || n.Counts[0] < 0 || n.Counts[1] < 0 || n.Counts[0]+n.Counts[1] <= 0 {
		return false
	}
	if n.Feature == "" {
		return n.Yes == nil && n.No == nil
	}
	if _, ok := selectorFeatures("ক", "k", "ko")[n.Feature]; !ok {
		return false
	}
	if !validSelector(n.Yes, depth+1) || !validSelector(n.No, depth+1) {
		return false
	}
	return n.Counts[0] == n.Yes.Counts[0]+n.No.Counts[0] && n.Counts[1] == n.Yes.Counts[1]+n.No.Counts[1]
}

func decodeSelector(data []byte) *selectorNode {
	var artifact struct {
		Schema int           `json:"schema"`
		Tree   *selectorNode `json:"tree"`
	}
	if json.Unmarshal(data, &artifact) != nil || artifact.Schema != 1 || !validSelector(artifact.Tree, 0) {
		return nil
	}
	return artifact.Tree
}

func selectorFeatures(native, alternative, model string) map[string]string {
	n, a, b := []rune(native), []rune(alternative), []rune(model)
	prefix := 0
	for prefix < min(len(a), len(b)) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < min(len(a), len(b))-prefix && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	left, right := a[prefix:len(a)-suffix], b[prefix:len(b)-suffix]
	before, after := "", ""
	if prefix > 0 {
		before = string(a[prefix-1])
	}
	if suffix > 0 {
		after = string(a[len(a)-suffix])
	}
	flag := func(value bool) string {
		if value {
			return "True"
		}
		return "False"
	}
	return map[string]string{
		"first": string(n[:min(1, len(n))]), "first2": string(n[:min(2, len(n))]),
		"last": string(n[max(0, len(n)-1):]), "last2": string(n[max(0, len(n)-2):]),
		"length": strconv.Itoa(min(len(n), 12)), "has_halant": flag(strings.ContainsRune(native, '্')),
		"has_i": flag(strings.ContainsAny(native, "ইঈিী")), "has_u": flag(strings.ContainsAny(native, "উঊুূ")),
		"has_o": flag(strings.ContainsAny(native, "ওো")), "has_nasal": flag(strings.ContainsAny(native, "ংঁ")),
		"edit_b1": string(left[:min(2, len(left))]), "edit_model": string(right[:min(2, len(right))]),
		"delta": strconv.Itoa(max(-3, min(3, len(a)-len(b)))), "edit_at_start": flag(prefix == 0), "edit_at_end": flag(suffix == 0),
		"before": before, "after": after,
	}
}

func selectorProbability(n *selectorNode, features map[string]string) float64 {
	for n.Feature != "" {
		if features[n.Feature] == n.Value {
			n = n.Yes
		} else {
			n = n.No
		}
	}
	return float64(n.Counts[1]) / float64(n.Counts[0]+n.Counts[1])
}

// Each variant owns its rule closures; no model data or shared word is mutated.
type vowelVariant struct {
	Bengali
	flip int
}

func (v vowelVariant) Rules() core.RuleCatalog {
	catalog := v.Bengali.Rules()
	for i := range catalog.Schwa {
		r := &catalog.Schwa[i]
		if r.Name != "schwa.bengali.model" {
			continue
		}
		r.Action = func(u *core.Unit, w *core.Word) {
			label, _ := vowelModelDecision(w, u)
			raw := w.Runes()
			index := len([]rune(vowelWordView(string(raw[:u.Start.Rune]))))
			if index == v.flip {
				if label == 0 {
					label = 1
				} else {
					label = 0
				}
			}
			if label == 0 {
				brahmic.SetSchwa(u, brahmic.SchwaDelete)
			} else {
				brahmic.SetSchwa(u, brahmic.SchwaKeep)
				if label == 2 {
					brahmic.GetBrahmicData(u).SchwaQuality = brahmic.SchwaRaised
				}
			}
		}
	}
	return catalog
}

// RerankNative is opt-in through both SchwaModel and Rerank, and limited to the
// evaluated default style. Unsupported words/options retain the caller's output.
func (Bengali) RerankNative(input string, opts core.Options, render core.CandidateRenderer) (string, bool) {
	if !opts.Rerank || !opts.SchwaModel || !opts.DefaultStyle() || utf8.RuneCountInString(input) > core.MaxLearnedWordRunes {
		return "", false
	}
	word, ok := bengaliLexiconKey(input)
	if !ok {
		return "", false
	}
	units, ok := vowelUnits([]rune(word))
	if !ok {
		return "", false
	}
	selectorOnce.Do(func() { learnedSelector = decodeSelector(selectorJSON) })
	vowelOnce.Do(func() { learnedVowels = decodeVowels(vowelTreeJSON) })
	if learnedSelector == nil || learnedVowels == nil {
		return "", false
	}
	model := render(Bengali{}, opts)
	best, bestScore := model, selectorThreshold
	seen := map[string]bool{model: true}
	consider := func(alt string) {
		if seen[alt] {
			return
		}
		seen[alt] = true
		score := selectorProbability(learnedSelector, selectorFeatures(word, alt, model))
		if score > bestScore {
			best, bestScore = alt, score
		}
	}
	b1 := opts
	b1.SchwaModel = false
	consider(render(Bengali{}, b1))
	edits := 0
	for _, unit := range units {
		if !unit.inherent {
			continue
		}
		if edits == selectorEdits {
			break
		}
		edits++
		consider(render(vowelVariant{flip: unit.index}, opts))
	}
	return best, true
}
