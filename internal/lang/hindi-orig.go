package lang

import "fmt"

/*
 * references:
 * - https://www.ushuaia.pl/transliterate/?ln=en
 * - http://transliteration.eki.ee/pdf/Hindi-Marathi-Nepali.pdf
 * - https://eclecticgeek.com/dompdf/core_tests/encoding_utf-8.html
 * - https://en.wiktionary.org/wiki/Wiktionary:Hindi_transliteration#Nuqt%C4%81
 */

var ignored = map[rune]int{
	'़': 1, // https://graphemica.com/093C
	'ऽ': 1, // https://graphemica.com/093D
	'्': 1, // https://graphemica.com/094D - https://en.wiktionary.org/wiki/%E0%A5%8D
	'ॎ': 1, // https://graphemica.com/094E - TODO
}

var vowels = map[string]string{
	"ँ": "n",  // https://graphemica.com/0901
	"ं": "n",  // https://graphemica.com/0902
	"ः": "a",  // https://graphemica.com/0903
	"ऄ": "a",  // https://graphemica.com/0904
	"अ": "a",  // https://graphemica.com/0905
	"आ": "aa", // https://graphemica.com/0906
	"इ": "i",  // https://graphemica.com/0907
	"ई": "ee", // https://graphemica.com/0908
	"उ": "u",  // https://graphemica.com/0909
	"ऊ": "oo", // https://graphemica.com/090A
	"ऋ": "ri", // https://graphemica.com/090B
	"ऌ": "l",  // https://graphemica.com/090C - https://en.wikipedia.org/wiki/%E1%B8%B6_(Indic)
	"ऍ": "æ",  // https://graphemica.com/090D - https://en.wiktionary.org/wiki/%E0%A4%8D
	"ऎ": "e",  // https://graphemica.com/090E - https://en.wiktionary.org/wiki/%E0%A4%8E
	"ए": "e",  // https://graphemica.com/090F
	"ऐ": "ai", // https://graphemica.com/0910
	"ऑ": "o",  // https://graphemica.com/0911 - https://en.wiktionary.org/wiki/%E0%A4%91
	"ऒ": "o",  // https://graphemica.com/0912 - https://en.wiktionary.org/wiki/%E0%A4%92
	"ओ": "o",  // https://graphemica.com/0913
	"औ": "au", // https://graphemica.com/0914

	"ा": "a",  // https://graphemica.com/093E
	"ि": "i",  // https://graphemica.com/093F
	"ी": "i",  // https://graphemica.com/0940
	"ु": "u",  // https://graphemica.com/0941
	"ू": "oo", // https://graphemica.com/0942
	"ृ": "ri", // https://graphemica.com/0943
	"ॄ": "ri", // https://graphemica.com/0944
	"ॅ": "æ",  // https://graphemica.com/0945
	"ॆ": "e",  // https://graphemica.com/0946
	"े": "e",  // https://graphemica.com/0947
	"ै": "ai", // https://graphemica.com/0948
	"ॉ": "o",  // https://graphemica.com/0949
	"ॊ": "o",  // https://graphemica.com/094A
	"ो": "o",  // https://graphemica.com/094B
	"ौ": "au", // https://graphemica.com/094C
	// 094E ^
	"ॏ": "au", // https://graphemica.com/094F
	//
	"ॢ": "l",  // https://graphemica.com/0962
	"ॣ": "ll", // https://graphemica.com/0963
}

var consonants = map[string]string{
	"क": "k",  // https://graphemica.com/0915
	"ख": "kh", // https://graphemica.com/0916
	"ग": "g",  // https://graphemica.com/0917
	"घ": "gh", // https://graphemica.com/0918
	"ङ": "n",  // https://graphemica.com/0919 or "ng"

	"च": "ch",  // https://graphemica.com/091A
	"छ": "chh", // https://graphemica.com/091B
	"ज": "j",   // https://graphemica.com/091C
	"झ": "jh",  // https://graphemica.com/091D
	"ञ": "ny",  // https://graphemica.com/091E

	"ट": "t",  // https://graphemica.com/091F
	"ठ": "th", // https://graphemica.com/0920
	"ड": "d",  // https://graphemica.com/0921
	"ढ": "dh", // https://graphemica.com/0922
	"ण": "n",  // https://graphemica.com/0923

	"त": "t",  // https://graphemica.com/0924
	"थ": "th", // https://graphemica.com/0925
	"द": "d",  // https://graphemica.com/0926
	"ध": "dh", // https://graphemica.com/0927
	"न": "n",  // https://graphemica.com/0928
	// nuqta
	"ऩ": "nh", // https://graphemica.com/0929 - https://en.wiktionary.org/wiki/%E0%A4%A9

	"प": "p",  // https://graphemica.com/092A
	"फ": "ph", // https://graphemica.com/092B
	"ब": "b",  // https://graphemica.com/092C
	"भ": "bh", // https://graphemica.com/092D
	"म": "m",  // https://graphemica.com/092E

	"य": "y", // https://graphemica.com/092F
	"र": "r", // https://graphemica.com/0930
	// nuqta
	"ऱ": "rh", // https://graphemica.com/0931
	"ल": "l",  // https://graphemica.com/0932
	"ळ": "l",  // https://graphemica.com/0933
	// nukta
	"ऴ": "lh", // https://graphemica.com/0934
	// TODO: In Marathi, [w], except [v] before [i]; [v], [ʋ], [w] allophony in Hindi
	"व": "v", // https://graphemica.com/0935

	"श": "sh", // https://graphemica.com/0936
	"ष": "sh", // https://graphemica.com/0937
	"स": "s",  // https://graphemica.com/0938
	"ह": "h",  // https://graphemica.com/0939

	// additional consonants with nuqta (़) diacritic
	// The nuqta diacritic describes sounds not found in Sanskrit, and therefore not represented by standard Devanagari.
	// Most of these sounds are found in Persian and English loanwords
	"क़": "q",  // https://graphemica.com/0958
	"ख़": "kh", // https://graphemica.com/0959 -  or "x"
	"ग़": "gh", // https://graphemica.com/095A
	"ज़": "z",  // https://graphemica.com/095B
	"ड़": "r",  // https://graphemica.com/095C - or "ddh"
	"ढ़": "rh", // https://graphemica.com/095D
	"फ़": "f",  // https://graphemica.com/095E
	"य़": "yh", // https://graphemica.com/095F

	// additional
	"ॠ": "ri", // https://graphemica.com/0960
	"ॡ": "ll", // https://graphemica.com/0961

	// conjuncts - '़'
	"क़": "q",
	"ख़": "kh",
	"ग़": "gh",
	"ज़": "z",
	"ङ़": "r",
	"ढ़": "rh",
	"फ़": "f",
	"य़": "yh",
	"झ़": "zh",
	"ऩ": "nh",
	"ऱ": "rh",
	"ऴ": "lh",

	"क्ष": "ksh",
	"त्र": "tr",
	"ज्ञ": "gy",
	"ॐ":   "om", // https://graphemica.com/094F
}

var numbers = map[string]string{
	"०": "0",
	"१": "1",
	"२": "2",
	"३": "3",
	"४": "4",
	"५": "5",
	"६": "6",
	"७": "7",
	"८": "8",
	"९": "9",
}

type HindiOrig struct {
}

func (l HindiOrig) Name() string {
	return "hindi-orig"
}

func (l HindiOrig) Info() {
	for i := 2305; i < 2415; i++ {
		r := rune(i)
		s := string(r)
		t := "unknown"
		m := ""
		if v, vFound := vowels[s]; vFound {
			t = "vowel"
			m = v
		} else if c, cFound := consonants[s]; cFound {
			t = "consonant"
			m = c
		} else if n, nFound := numbers[s]; nFound {
			t = "number"
			m = n
		} else if _, iFound := ignored[r]; iFound {
			t = "ignored"
			m = ""
		}
		fmt.Printf("\"%s\":  {\"%s\", %s}, // U+0%X\n", s, m, t, i)
	}
}

// Transliterate converts a word using default options.
func (l HindiOrig) Transliterate(word string) string {
	return l.TransliterateWithOptions(word, DefaultOptions())
}

// TransliterateWithOptions converts a word (options ignored in legacy implementation).
func (l HindiOrig) TransliterateWithOptions(word string, _ Options) string {
	runes := []rune(word)

	var converted []rune
	var next rune
	var k = ""
	var kLen = 0
	for i := 0; i < len(runes); i++ {
		current := runes[i]
		next = rune(0) // stores the next valid rune - if

		// '़' is the `nuqta` diacritic used in the devanagari, for sounds not present
		// in the original scripts. This is a special mark added to a letter to indicate
		// a different pronunciation
		if i+1 < len(runes) && runes[i+1] == '़' {
			k = string([]rune{runes[i], runes[i+1]})
			kLen = 2
		} else {
			k = string([]rune{runes[i]})
			kLen = 1
		}

		// calculate the "next" full rune
		if i+kLen < len(runes) {
			next = runes[i+kLen]
		}

		if v, found := vowels[k]; found {
			converted = append(converted, []rune(v)...)
		} else if c, found := consonants[k]; found {
			converted = append(converted, []rune(c)...)
			if isConsonant(next) {
				// add 'a' after 'jh', only if झ appears in the starting of the word
				if current == 'झ' && i == 0 {
					converted = append(converted, 'a')
					continue
				}
				// add 'a' if the next rune is not a vowel
				if !isVowel(next) {
					converted = append(converted, 'a')
					continue
				}
			}
		} else if n, found := numbers[k]; found {
			converted = append(converted, []rune(n)...)
		} else {
			// this is a catch-all,  when the script is unable to find a valid conversion
			// add it back as is
			if !isIgnored(current) {
				converted = append(converted, current)
			}
		}
	}
	return string(converted)
}

func isVowel(r rune) bool {
	_, found := vowels[string(r)]
	return found
}

func isConsonant(r rune) bool {
	_, found := consonants[string(r)]
	return found
}

func isIgnored(r rune) bool {
	_, found := ignored[r]
	return found
}
