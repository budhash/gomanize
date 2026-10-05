package bengali

import (
	"bufio"
	_ "embed"
	"strings"
	"sync"
	"unicode"

	"github.com/budhash/gomanize/core"
)

// Derived Dakshina train data, CC BY-SA 4.0; see LEXICON.md for attribution,
// selection policy, exclusions and the default-style-only lookup contract.
//
//go:embed lexicon.tsv
var lexiconTSV string

var lexiconOnce sync.Once
var spellings map[string]string

// bengaliLexiconKey implements Cf removal and canonical normalization within
// the Bengali block. Keys outside Bengali letters/marks are not in this lexicon.
// The block's nonzero combining classes are nukta=7, virama=9, sandhi=230;
// canonical decompositions are the three nukta aliases and the two split matras.
func bengaliLexiconKey(word string) (string, bool) {
	var runes []rune
	for _, r := range word {
		if unicode.Is(unicode.Cf, r) {
			continue
		}
		if r < 0x0980 || r > 0x09ff || (!unicode.IsLetter(r) && !unicode.IsMark(r)) {
			return "", false
		}
		switch r {
		case 'ড়':
			runes = append(runes, 'ড', '়')
		case 'ঢ়':
			runes = append(runes, 'ঢ', '়')
		case 'য়':
			runes = append(runes, 'য', '়')
		default:
			runes = append(runes, r)
		}
	}
	class := func(r rune) int {
		switch r {
		case '়':
			return 7
		case '্':
			return 9
		case '\u09fe':
			return 230
		}
		return 0
	}
	for i := 1; i < len(runes); i++ {
		if class(runes[i]) == 0 {
			continue
		}
		for j := i; j > 0 && class(runes[j-1]) > class(runes[j]); j-- {
			runes[j], runes[j-1] = runes[j-1], runes[j]
		}
	}
	key := strings.NewReplacer("ো", "ো", "ৌ", "ৌ").Replace(string(runes))
	return key, key != ""
}

func loadLexicon() map[string]string {
	lexiconOnce.Do(func() {
		entries := map[string]string{}
		scanner := bufio.NewScanner(strings.NewReader(lexiconTSV))
		for scanner.Scan() {
			native, roman, ok := strings.Cut(scanner.Text(), "\t")
			if !ok || native == "" || roman == "" || strings.ContainsRune(roman, '\t') {
				spellings = map[string]string{}
				return
			}
			if _, exists := entries[native]; exists {
				spellings = map[string]string{}
				return
			}
			key, ok := bengaliLexiconKey(native)
			if !ok || key != native {
				spellings = map[string]string{}
				return
			}
			entries[native] = roman
		}
		if scanner.Err() != nil {
			spellings = map[string]string{}
			return
		}
		spellings = entries
	})
	return spellings
}

// LexiconLookupWithOptions returns an attested spelling only for the default
// rendering style. SchwaModel changes fallback decisions, not lexicon spellings.
func (Bengali) LexiconLookupWithOptions(word string, opts core.Options) (string, bool) {
	if opts.InherentVowelA || opts.LongVowels || opts.SimpleNasals || opts.KeepMedialSchwa {
		return "", false
	}
	key, ok := bengaliLexiconKey(word)
	if !ok {
		return "", false
	}
	roman, found := loadLexicon()[key]
	return roman, found
}

// LexiconSize is the number of isolated, unambiguous training spellings.
func LexiconSize() int { return len(loadLexicon()) }
