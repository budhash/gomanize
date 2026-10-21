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

// This opt-in model only handles the eligibility inventory shared with
// tools/bengali/vowels.py (vowel_inventory.json). Unsupported words retain the
// B1 rules in full.
//
//go:embed vowel_tree.json
var vowelTreeJSON []byte

//go:embed vowel_inventory.json
var vowelInventoryJSON []byte

type vowelInventory struct {
	Schema       int      `json:"schema"`
	Scope        string   `json:"scope"`
	LegacyScope  string   `json:"legacy_scope"`
	Consonants   []string `json:"consonants"`
	Matras       []string `json:"matras"`
	Independent  []string `json:"independent_vowels"`
	Halant       string   `json:"halant"`
	MaxCluster   int      `json:"max_cluster"`
	consonantSet map[string]bool
	matraSet     map[rune]bool
	vowelSet     map[rune]bool
	halant       rune
}

// inventory is the shared eligibility data; a malformed file disables the model.
var inventory = decodeInventory(vowelInventoryJSON)

func decodeInventory(data []byte) *vowelInventory {
	var inv vowelInventory
	if json.Unmarshal(data, &inv) != nil || inv.Schema != 1 || inv.Scope == "" || inv.LegacyScope == "" || inv.MaxCluster < 2 || len([]rune(inv.Halant)) != 1 {
		return nil
	}
	inv.halant = []rune(inv.Halant)[0]
	inv.consonantSet, inv.matraSet, inv.vowelSet = map[string]bool{}, map[rune]bool{}, map[rune]bool{}
	for _, c := range inv.Consonants {
		inv.consonantSet[c] = true
	}
	for _, sets := range []struct {
		from []string
		to   map[rune]bool
	}{{inv.Matras, inv.matraSet}, {inv.Independent, inv.vowelSet}} {
		for _, s := range sets.from {
			if len([]rune(s)) != 1 {
				return nil
			}
			sets.to[[]rune(s)[0]] = true
		}
	}
	return &inv
}

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
	// maxCluster is the cluster policy implied by Scope: 1 for the legacy
	// simple-word scope, the inventory's max_cluster for its current scope.
	maxCluster int
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
	if inventory == nil || json.Unmarshal(data, &model) != nil || model.Schema != 1 || strings.Join(model.Classes, ",") != "absent,default,raised" || strings.Join(model.Features, ",") != strings.Join(vowelFeatureNames, ",") || !validVowelNode(model.Tree, 0) {
		return nil
	}
	switch model.Scope {
	case inventory.Scope:
		model.maxCluster = inventory.MaxCluster
	case inventory.LegacyScope:
		model.maxCluster = 1
	default:
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

func vowelConsonant(s string) bool { return inventory != nil && inventory.consonantSet[s] }
func vowelMatra(r rune) bool       { return inventory != nil && inventory.matraSet[r] }
func vowelIndependent(r rune) bool { return inventory != nil && inventory.vowelSet[r] }
func vowelHalant(s string) bool    { return inventory != nil && s == inventory.Halant }

type vowelUnit struct {
	index    int
	spelling string
	inherent bool
}

// vowelUnits mirrors tools/bengali/vowels.py units: consonants, halants, matras
// and independent vowels from the shared inventory. A consonant before a halant
// has no inherent slot; clusters have at most maxCluster members (1 = legacy).
func vowelUnits(runes []rune, maxCluster int) ([]vowelUnit, bool) {
	if inventory == nil {
		return nil, false
	}
	var out []vowelUnit
	members := 0
	for i := 0; i < len(runes); {
		ch, size := string(runes[i]), 1
		if i+1 < len(runes) && runes[i] == 'ড' && runes[i+1] == '়' {
			ch, size = "ড়", 2
		}
		afterHalant := len(out) > 0 && vowelHalant(out[len(out)-1].spelling)
		if afterHalant && !vowelConsonant(ch) {
			return nil, false
		}
		slot := false
		switch {
		case vowelConsonant(ch):
			if afterHalant {
				members++
			} else {
				members = 1
			}
			if members > maxCluster {
				return nil, false
			}
			end := i + size
			slot = end == len(runes) || (!vowelMatra(runes[end]) && runes[end] != inventory.halant)
		case runes[i] == inventory.halant:
			if maxCluster < 2 || len(out) == 0 || !vowelConsonant(out[len(out)-1].spelling) || i+1 == len(runes) {
				return nil, false
			}
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

// vowelEligible reports whether the model decides any slot of word (already
// normalized by vowelWordView) under the given cluster policy.
func vowelEligible(word string, maxCluster int) bool {
	units, ok := vowelUnits([]rune(word), maxCluster)
	for _, unit := range units {
		if unit.inherent {
			return ok
		}
	}
	return false
}

func vowelFeatures(word string, index, maxCluster int) (map[string]string, bool) {
	runes := []rune(word)
	units, ok := vowelUnits(runes, maxCluster)
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
	features, ok := vowelFeatures(word, index, learnedVowels.maxCluster)
	if !ok {
		return 0, false
	}
	return predictVowel(learnedVowels, features), true
}
