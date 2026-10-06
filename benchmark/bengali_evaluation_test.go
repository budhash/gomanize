package benchmark

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/lang/bengali"
	"github.com/budhash/gomanize/script/brahmic"
)

// bnVowelPattern marks only vowels emitted by the B0 renderer: ~ is a retained
// inherent vowel and @ is independent অ. Explicit matras/ও remain literal.
// This is a conditional spelling measurement, not a pronunciation alignment.
func bnVowelPattern(native string) string {
	w := brahmic.NewParser(bengali.Bengali{}.ScriptConfig()).Parse(native, bengali.Symbols)
	brahmic.IdentifyRuns(w)
	all := bengali.RuleCatalog().AllRules()
	baseline := core.AppendIfFound(core.AppendIfFound(nil, all, "schwa.delete.word-final"), all, "schwa.keep.default")
	core.NewRuleEngine(baseline).Apply(w)
	wd := brahmic.GetWordBrahmicData(w)
	if wd == nil {
		return ""
	}
	profile := *wd.Profile
	profile.InherentVowel = "~"
	wd.Profile = &profile
	for _, u := range w.Units {
		if string(u.Runes) == "অ" {
			u.BaseRom = "@"
		}
	}
	return brahmic.NewRenderer().Render(w)
}

// bnAlignSlots returns per-position alternatives across ALL exact-anchor
// alignments: 1=o, 2=a, 4=omitted. Multi-bit masks are ambiguous and must not be
// resolved with an arbitrary edit-distance tie break. nil means no alignment.
func bnAlignSlots(pattern, roman string) []int {
	p, r := []rune(pattern), []rune(roman)
	type pos struct{ i, j int }
	memo := map[pos]bool{}
	known := map[pos]bool{}
	slot := func(c rune) bool { return c == '~' || c == '@' }
	var valid func(int, int) bool
	valid = func(i, j int) bool {
		q := pos{i, j}
		if known[q] {
			return memo[q]
		}
		known[q] = true
		ok := false
		if i == len(p) {
			ok = j == len(r)
		} else if slot(p[i]) {
			ok = valid(i+1, j)
			if j < len(r) && (r[j] == 'o' || r[j] == 'a') {
				ok = valid(i+1, j+1) || ok
			}
		} else if j < len(r) && p[i] == r[j] {
			ok = valid(i+1, j+1)
		}
		memo[q] = ok
		return ok
	}
	if !valid(0, 0) {
		return nil
	}
	masks := make([]int, len(p))
	visited := map[pos]bool{}
	var walk func(int, int)
	walk = func(i, j int) {
		q := pos{i, j}
		if visited[q] || i == len(p) {
			return
		}
		visited[q] = true
		if slot(p[i]) {
			if valid(i+1, j) {
				masks[i] |= 4
				walk(i+1, j)
			}
			if j < len(r) && (r[j] == 'o' || r[j] == 'a') && valid(i+1, j+1) {
				if r[j] == 'o' {
					masks[i] |= 1
				} else {
					masks[i] |= 2
				}
				walk(i+1, j+1)
			}
		} else if j < len(r) && p[i] == r[j] && valid(i+1, j+1) {
			walk(i+1, j+1)
		}
	}
	walk(0, 0)
	return masks
}

func TestBengaliAlignmentDefinitions(t *testing.T) {
	for _, tc := range []struct {
		pattern, roman string
		want           string
	}{
		{"k~m~l", "komal", "[0 1 0 2 0]"},
		{"k~m~l", "kml", "[0 4 0 4 0]"},
		{"~@", "o", "[5 5]"},     // either position could own the o
		{"k~a", "ka", "[0 4 0]"}, // a matra must not become an inferred inherent a
		{"k~", "ko", "[0 1]"},
		{"k~", "kho", "[]"}, // mismatched consonant anchors are excluded
	} {
		if got := fmt.Sprint(bnAlignSlots(tc.pattern, tc.roman)); got != tc.want {
			t.Errorf("%+v: %s", tc, got)
		}
	}
	for input, want := range map[string]string{"কমল": "k~m~l", "অমন": "@m~n", "কো": "ko", "কা": "ka", "মধ্য": "m~dhj", "ৎই": "ti"} {
		if got := bnVowelPattern(input); got != want {
			t.Errorf("%s: %s != %s", input, got, want)
		}
	}
}

// bnEvidenceDigest pins the B0 decision evidence recorded in
// docs/reviews/2026-10-04-bengali-evaluation-gate.md (every logged line, per
// split). The harness uses the fixed B0 rule subset, so later rule work must
// not move it; a change means the recorded evidence no longer reproduces.
var bnEvidenceDigest = map[string]string{
	"train": "e63a5323384390b43f5d2573c64f6119fd5b02a430623993d018d8af1c36cb10",
	"dev":   "2a8d073d9900b5efae3e034e46f2cb83470c3fe17ed1b0d839300d9140b7fcee",
}

func TestBengaliVowelEvidence(t *testing.T) {
	// Test is deliberately neither opened nor enumerated here.
	datasets := loadBengali(t, "train", "dev")
	for _, split := range []string{"train", "dev"} {
		slots, coveredWords, withSlots, refs, alignedRefs := 0, 0, 0, 0, 0
		var inherent, independent [8]int
		preferO, preferA, ties, noOA := 0, 0, 0, 0
		thresholds := map[int]int{}
		curatedStyles := map[int][2]bnCounts{}
		for _, w := range datasets[split] {
			p := bnVowelPattern(w.native)
			n := strings.Count(p, "~") + strings.Count(p, "@")
			if n > 0 {
				withSlots++
			}
			slots += n
			best := 0
			wordO, wordA := 0, 0
			covered := false
			for _, ref := range w.refs {
				refs++
				if ref.votes > best {
					best = ref.votes
				}
				masks := bnAlignSlots(p, ref.roman)
				if masks == nil {
					continue
				}
				alignedRefs++
				if n > 0 {
					covered = true
				}
				for i, c := range []rune(p) {
					if c != '~' && c != '@' {
						continue
					}
					mask := masks[i]
					if c == '~' {
						inherent[mask] += ref.votes
					} else {
						independent[mask] += ref.votes
					}
					if mask == 1 {
						wordO += ref.votes
					}
					if mask == 2 {
						wordA += ref.votes
					}
				}
			}
			if covered {
				coveredWords++
			}
			if n > 0 {
				switch {
				case wordO == 0 && wordA == 0:
					noOA++
				case wordO > wordA:
					preferO++
				case wordA > wordO:
					preferA++
				default:
					ties++
				}
			}
			for _, threshold := range []int{1, 2, 3, 4} {
				if best >= threshold {
					thresholds[threshold]++
					scores := curatedStyles[threshold]
					scores[0].add(strings.NewReplacer("~", "o", "@", "o").Replace(p), w.refs)
					scores[1].add(strings.NewReplacer("~", "a", "@", "a").Replace(p), w.refs)
					curatedStyles[threshold] = scores
				}
			}
		}
		lines := []string{
			fmt.Sprintf("%s words=%d with-slots=%d slots=%d covered-words=%d refs=%d aligned-refs=%d", split, len(datasets[split]), withSlots, slots, coveredWords, refs, alignedRefs),
			fmt.Sprintf("%s vote-weighted slot masks [0,o,a,o|a,empty,o|empty,a|empty,all]: inherent=%v independent=%v", split, inherent, independent),
			fmt.Sprintf("%s equal-word preference o=%d a=%d tie=%d no-unambiguous-oa=%d", split, preferO, preferA, ties, noOA),
		}
		for _, threshold := range []int{1, 2, 3, 4} {
			s := curatedStyles[threshold]
			lines = append(lines, fmt.Sprintf("%s threshold=%d words=%d default strict=%d any=%d CER=%.8f alternate strict=%d any=%d CER=%.8f", split, threshold, thresholds[threshold], s[0].strict, s[0].any, s[0].minCER/float64(s[0].words), s[1].strict, s[1].any, s[1].minCER/float64(s[1].words)))
		}
		for _, line := range lines {
			t.Log(line)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines, "\n")))); got != bnEvidenceDigest[split] {
			t.Errorf("%s B0 evidence changed: digest %s, pinned %s", split, got, bnEvidenceDigest[split])
		}
	}
}
