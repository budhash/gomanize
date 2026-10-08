package bengali

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// This opt-in model only handles the simple-word inventory shared with
// tools/bengali/vowels.py. Unsupported words retain the B1 rules in full.
//
//go:embed vowel_tree.json
var vowelTreeJSON []byte

type vowelNode struct {
	Feature string     `json:"f"`
	Value   string     `json:"v"`
	Yes     *vowelNode `json:"yes"`
	No      *vowelNode `json:"no"`
	Leaf    *int       `json:"leaf"`
}

type vowelTree struct {
	Classes  []string   `json:"classes"`
	Schema   int        `json:"schema"`
	Scope    string     `json:"scope"`
	Features []string   `json:"features"`
	Tree     *vowelNode `json:"tree"`
}

var vowelOnce sync.Once
var learnedVowels *vowelTree
var vowelFeatureNames = []string{"cons", "prev", "next", "next2", "first", "last", "length", "position"}

func validVowelNode(n *vowelNode, depth int) bool {
	if n == nil || depth > 32 {
		return false
	}
	if n.Leaf != nil {
		return *n.Leaf >= 0 && *n.Leaf <= 2 && n.Yes == nil && n.No == nil && n.Feature == ""
	}
	valid := false
	for _, name := range vowelFeatureNames {
		if n.Feature == name {
			valid = true
		}
	}
	return valid && n.Value != "" && validVowelNode(n.Yes, depth+1) && validVowelNode(n.No, depth+1)
}

func decodeVowels(data []byte) *vowelTree {
	var model vowelTree
	if json.Unmarshal(data, &model) != nil || model.Schema != 1 || model.Scope != "simple-whole-words-v1" || strings.Join(model.Classes, ",") != "absent,default,raised" || strings.Join(model.Features, ",") != strings.Join(vowelFeatureNames, ",") || !validVowelNode(model.Tree, 0) {
		return nil
	}
	return &model
}

// Normalize only equivalences in the supported inventory. The parser has
// already removed Cf; this view must not change shared parser rune indices.
func vowelWordView(s string) string {
	s = strings.ReplaceAll(s, "ড়", "ড়")
	return strings.ReplaceAll(s, "ো", "ো")
}

func vowelConsonant(s string) bool {
	return s == "ড়" || (len([]rune(s)) == 1 && strings.Contains("কখগঘঙচছজঝটঠডঢণতথদধনপফবভমযরলশষসহ", s))
}
func vowelMatra(r rune) bool       { return strings.ContainsRune("ািীুূেো", r) }
func vowelIndependent(r rune) bool { return strings.ContainsRune("অআইঈউঊএও", r) }

type vowelUnit struct {
	index    int
	spelling string
	inherent bool
}

func vowelUnits(runes []rune) ([]vowelUnit, bool) {
	var out []vowelUnit
	for i := 0; i < len(runes); {
		ch, size := string(runes[i]), 1
		if i+1 < len(runes) && runes[i] == 'ড' && runes[i+1] == '়' {
			ch, size = "ড়", 2
		}
		slot := false
		switch {
		case vowelConsonant(ch):
			slot = i+size == len(runes) || !vowelMatra(runes[i+size])
		case vowelIndependent(runes[i]):
		case vowelMatra(runes[i]):
			if len(out) == 0 || !vowelConsonant(out[len(out)-1].spelling) || out[len(out)-1].inherent {
				return nil, false
			}
		default:
			return nil, false
		}
		out = append(out, vowelUnit{i, ch, slot})
		i += size
	}
	for n, unit := range out {
		if unit.inherent && n+1 < len(out) && vowelIndependent([]rune(out[n+1].spelling)[0]) {
			return nil, false
		}
	}
	return out, len(out) > 0
}

func vowelFeatures(word string, index int) (map[string]string, bool) {
	runes := []rune(word)
	units, ok := vowelUnits(runes)
	if !ok {
		return nil, false
	}
	first, last, cons := -1, -1, ""
	for _, unit := range units {
		if vowelConsonant(unit.spelling) {
			if first < 0 {
				first = unit.index
			}
			last = unit.index
		}
		if unit.index == index && unit.inherent {
			cons = unit.spelling
		}
	}
	if cons == "" {
		return nil, false
	}
	previous, next, next2 := "^", "$", "$"
	if index > 0 {
		previous = string(runes[index-1])
	}
	if index+1 < len(runes) {
		next = string(runes[index+1])
	}
	if index+2 < len(runes) {
		next2 = string(runes[index+2])
	}
	isFirst, isLast := "0", "0"
	if index == first {
		isFirst = "1"
	}
	if index == last {
		isLast = "1"
	}
	length, position := len(runes), index
	if length > 12 {
		length = 12
	}
	if position > 6 {
		position = 6
	}
	return map[string]string{"cons": cons, "prev": previous, "next": next, "next2": next2,
		"first": isFirst, "last": isLast, "length": strconv.Itoa(length), "position": strconv.Itoa(position)}, true
}

func predictVowel(model *vowelTree, features map[string]string) int {
	n := model.Tree
	for n.Leaf == nil {
		if features[n.Feature] == n.Value {
			n = n.Yes
		} else {
			n = n.No
		}
	}
	return *n.Leaf
}

func vowelModelDecision(w *core.Word, u *core.Unit) (int, bool) {
	if !w.Options.SchwaModel || !consonant(u) || brahmic.GetSchwa(u) != brahmic.SchwaPending {
		return 0, false
	}
	vowelOnce.Do(func() { learnedVowels = decodeVowels(vowelTreeJSON) })
	if learnedVowels == nil {
		return 0, false
	}
	raw := w.Runes()
	if len(raw) > core.MaxLearnedWordRunes || u.Start.Rune < 0 || u.Start.Rune >= len(raw) {
		return 0, false
	}
	word := vowelWordView(w.Original)
	index := len([]rune(vowelWordView(string(raw[:u.Start.Rune]))))
	features, ok := vowelFeatures(word, index)
	if !ok {
		return 0, false
	}
	return predictVowel(learnedVowels, features), true
}
