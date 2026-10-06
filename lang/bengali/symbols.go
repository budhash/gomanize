// Package bengali provides experimental rules-only Bengali romanization.
// Parsing remains compositional; scoped B1 rules handle positional behavior.
// Character identities follow the Unicode Bengali block.
package bengali

import (
	"github.com/budhash/gomanize/core"
	"github.com/budhash/gomanize/script/brahmic"
)

// Symbols maps Bengali letters/signs to the initial colloquial spelling.
// Currency, historic fractions, Assamese-only letters and unknown symbols pass
// through unchanged. This baseline makes no claim of phonetic completeness.
var Symbols = core.SymbolMap{
	"ঁ":  {Category: brahmic.CatChandrabindu, BaseRom: "n"},
	"ং":  {Category: brahmic.CatAnusvara, BaseRom: "ng"},
	"ঃ":  {Category: brahmic.CatVisarga, BaseRom: "h"},
	"অ":  {Category: core.CatVowel, BaseRom: "o"},
	"আ":  {Category: core.CatVowel, BaseRom: "a"},
	"ই":  {Category: core.CatVowel, BaseRom: "i"},
	"ঈ":  {Category: core.CatVowel, BaseRom: "i"},
	"উ":  {Category: core.CatVowel, BaseRom: "u"},
	"ঊ":  {Category: core.CatVowel, BaseRom: "u"},
	"ঋ":  {Category: core.CatVowel, BaseRom: "ri"},
	"ঌ":  {Category: core.CatVowel, BaseRom: "li"},
	"এ":  {Category: core.CatVowel, BaseRom: "e"},
	"ঐ":  {Category: core.CatVowel, BaseRom: "oi"},
	"ও":  {Category: core.CatVowel, BaseRom: "o"},
	"ঔ":  {Category: core.CatVowel, BaseRom: "ou"},
	"ক":  {Category: core.CatConsonant, BaseRom: "k"},
	"খ":  {Category: core.CatConsonant, BaseRom: "kh"},
	"গ":  {Category: core.CatConsonant, BaseRom: "g"},
	"ঘ":  {Category: core.CatConsonant, BaseRom: "gh"},
	"ঙ":  {Category: core.CatConsonant, BaseRom: "ng"},
	"চ":  {Category: core.CatConsonant, BaseRom: "ch"},
	"ছ":  {Category: core.CatConsonant, BaseRom: "chh"},
	"জ":  {Category: core.CatConsonant, BaseRom: "j"},
	"ঝ":  {Category: core.CatConsonant, BaseRom: "jh"},
	"ঞ":  {Category: core.CatConsonant, BaseRom: "n"},
	"ট":  {Category: core.CatConsonant, BaseRom: "t"},
	"ঠ":  {Category: core.CatConsonant, BaseRom: "th"},
	"ড":  {Category: core.CatConsonant, BaseRom: "d"},
	"ঢ":  {Category: core.CatConsonant, BaseRom: "dh"},
	"ণ":  {Category: core.CatConsonant, BaseRom: "n"},
	"ত":  {Category: core.CatConsonant, BaseRom: "t"},
	"থ":  {Category: core.CatConsonant, BaseRom: "th"},
	"দ":  {Category: core.CatConsonant, BaseRom: "d"},
	"ধ":  {Category: core.CatConsonant, BaseRom: "dh"},
	"ন":  {Category: core.CatConsonant, BaseRom: "n"},
	"প":  {Category: core.CatConsonant, BaseRom: "p"},
	"ফ":  {Category: core.CatConsonant, BaseRom: "ph"},
	"ব":  {Category: core.CatConsonant, BaseRom: "b"},
	"ভ":  {Category: core.CatConsonant, BaseRom: "bh"},
	"ম":  {Category: core.CatConsonant, BaseRom: "m"},
	"য":  {Category: core.CatConsonant, BaseRom: "j"},
	"র":  {Category: core.CatConsonant, BaseRom: "r"},
	"ল":  {Category: core.CatConsonant, BaseRom: "l"},
	"শ":  {Category: core.CatConsonant, BaseRom: "sh"},
	"ষ":  {Category: core.CatConsonant, BaseRom: "sh"},
	"স":  {Category: core.CatConsonant, BaseRom: "sh"},
	"হ":  {Category: core.CatConsonant, BaseRom: "h"},
	"ৎ":  {Category: core.CatConsonant, BaseRom: "t"},
	"ড়": {Category: core.CatConsonant, BaseRom: "r"},
	"ঢ়": {Category: core.CatConsonant, BaseRom: "rh"},
	"য়": {Category: core.CatConsonant, BaseRom: "y"},
	"ৠ":  {Category: core.CatVowel, BaseRom: "ri"},
	"ৡ":  {Category: core.CatVowel, BaseRom: "li"},
	"া":  {Category: brahmic.CatMatra, BaseRom: "a"},
	"ি":  {Category: brahmic.CatMatra, BaseRom: "i"},
	"ী":  {Category: brahmic.CatMatra, BaseRom: "i"},
	"ু":  {Category: brahmic.CatMatra, BaseRom: "u"},
	"ূ":  {Category: brahmic.CatMatra, BaseRom: "u"},
	"ৃ":  {Category: brahmic.CatMatra, BaseRom: "ri"},
	"ৄ":  {Category: brahmic.CatMatra, BaseRom: "ri"},
	"ে":  {Category: brahmic.CatMatra, BaseRom: "e"},
	"ৈ":  {Category: brahmic.CatMatra, BaseRom: "oi"},
	"ো":  {Category: brahmic.CatMatra, BaseRom: "o"},
	"ৌ":  {Category: brahmic.CatMatra, BaseRom: "ou"},
	"ো": {Category: brahmic.CatMatra, BaseRom: "o"},
	"ৌ": {Category: brahmic.CatMatra, BaseRom: "ou"},
	"ৢ":  {Category: brahmic.CatMatra, BaseRom: "li"},
	"ৣ":  {Category: brahmic.CatMatra, BaseRom: "li"},
	"্":  {Category: brahmic.CatHalant},
	"়":  {Category: brahmic.CatNukta},
	"ঽ":  {Category: core.CatSymbol, BaseRom: "'"},
}

func init() {
	for pre, base := range map[rune]rune{0x09DC: 0x09A1, 0x09DD: 0x09A2, 0x09DF: 0x09AF} {
		Symbols[string(pre)] = Symbols[string(base)+"়"]
	}
	for i := 0; i < 10; i++ {
		Symbols[string(rune(0x09E6+i))] = core.SymbolInfo{Category: core.CatNumber, BaseRom: string(rune('0' + i))}
	}
}

// Bengali implements core.Language with the measured B1 rule catalog.
type Bengali struct{}

func (Bengali) Name() string            { return "bengali" }
func (Bengali) Script() core.Script     { return brahmic.New() }
func (Bengali) Symbols() core.SymbolMap { return Symbols }
func (Bengali) ScriptConfig() interface{} {
	return brahmic.Config{
		Halant: "্", Nukta: "়", MultiChar: []string{"ো", "ৌ"},
		VowellessConsonants: []rune{'ৎ'},
		Profile: &brahmic.ScriptProfile{
			InherentVowel: "o", RaisedVowel: "o", BareVowelRunes: []rune{'অ'},
			AaMatra: 'া', SonorousRunes: []rune{'র', 'য', 'ব'}, IndependentVowelRange: [2]rune{0x0985, 0x0994},
		},
	}
}
func (Bengali) Rules() core.RuleCatalog { return RuleCatalog() }
