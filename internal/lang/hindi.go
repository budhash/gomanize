package lang

import (
	"fmt"
	"unicode/utf8"
)

/*
 * references:
 * - https://en.wikipedia.org/wiki/Devanagari_transliteration
 * - https://en.wikipedia.org/wiki/Hunterian_transliteration
 * - https://en.wikipedia.org/wiki/ISO_15919
 * - https://www.ushuaia.pl/transliterate/?ln=en
 * - http://transliteration.eki.ee/pdf/Hindi-Marathi-Nepali.pdf
 * - https://eclecticgeek.com/dompdf/core_tests/encoding_utf-8.html
 * - https://en.wiktionary.org/wiki/Wiktionary:Hindi_transliteration#Nuqt%C4%81
 */

type alt string

type category int

const (
	unused category = iota
	vowel
	consonant
	number
	symbol
	conjuncts
)

func (c category) String() string {
	return [...]string{"unused", "vowel", "consonant", "number", "symbol", "conjuncts"}[c]
}

type SymbolInfo struct {
	ctg   category // symbol category (vowel, consonant, etc.)
	hun   string   // Hunterian transliteration
	alt   []alt    // alternate transliterations
	nuqta bool     // additional consonants with nuqta (़) diacritic for sounds not in Sanskrit
}

func (si *SymbolInfo) isVowel() bool {
	if nil == si {
		return false
	} else {
		return si.ctg == vowel
	}
}

func (si *SymbolInfo) isConsonant() bool {
	if nil == si {
		return false
	} else {
		return si.ctg == consonant || si.ctg == conjuncts
	}
}

var symbols = map[string]SymbolInfo{
	"ँ": {ctg: vowel, hun: "n"},                       // U+0901
	"ं": {ctg: vowel, hun: "n", alt: []alt{"m"}},      // U+0902
	"ः": {ctg: vowel, hun: "h"},                       // U+0903
	"ऄ": {ctg: vowel, hun: "a"},                       // U+0904
	"अ": {ctg: vowel, hun: "a"},                       // U+0905
	"आ": {ctg: vowel, hun: "aa"},                      // U+0906
	"इ": {ctg: vowel, hun: "i"},                       // U+0907
	"ई": {ctg: vowel, hun: "i", alt: []alt{"ī"}},      // U+0908		// "ee"
	"उ": {ctg: vowel, hun: "u"},                       // U+0909
	"ऊ": {ctg: vowel, hun: "u"},                       // U+090A		// "oo"
	"ऋ": {ctg: vowel, hun: "ri"},                      // U+090B
	"ऌ": {ctg: vowel, hun: "l"},                       // U+090C
	"ऍ": {ctg: vowel, hun: "æ"},                       // U+090D
	"ऎ": {ctg: vowel, hun: "e"},                       // U+090E
	"ए": {ctg: vowel, hun: "e"},                       // U+090F
	"ऐ": {ctg: vowel, hun: "ai"},                      // U+0910
	"ऑ": {ctg: vowel, hun: "o"},                       // U+0911
	"ऒ": {ctg: vowel, hun: "o"},                       // U+0912
	"ओ": {ctg: vowel, hun: "o"},                       // U+0913
	"औ": {ctg: vowel, hun: "au"},                      // U+0914
	"क": {ctg: consonant, hun: "k"},                   // U+0915
	"ख": {ctg: consonant, hun: "kh"},                  // U+0916
	"ग": {ctg: consonant, hun: "g"},                   // U+0917
	"घ": {ctg: consonant, hun: "gh"},                  // U+0918
	"ङ": {ctg: consonant, hun: "n", alt: []alt{"ng"}}, // U+0919
	"च": {ctg: consonant, hun: "ch"},                  // U+091A
	"छ": {ctg: consonant, hun: "chh"},                 // U+091B
	"ज": {ctg: consonant, hun: "j"},                   // U+091C
	"झ": {ctg: consonant, hun: "jh"},                  // U+091D
	"ञ": {ctg: consonant, hun: "ny"},                  // U+091E
	"ट": {ctg: consonant, hun: "t"},                   // U+091F
	"ठ": {ctg: consonant, hun: "th"},                  // U+0920
	"ड": {ctg: consonant, hun: "d"},                   // U+0921
	"ढ": {ctg: consonant, hun: "dh"},                  // U+0922
	"ण": {ctg: consonant, hun: "n"},                   // U+0923
	"त": {ctg: consonant, hun: "t"},                   // U+0924
	"थ": {ctg: consonant, hun: "th"},                  // U+0925
	"द": {ctg: consonant, hun: "d"},                   // U+0926
	"ध": {ctg: consonant, hun: "dh"},                  // U+0927
	"न": {ctg: consonant, hun: "n"},                   // U+0928
	"ऩ": {ctg: consonant, hun: "nh", nuqta: true},     // U+0929
	"प": {ctg: consonant, hun: "p"},                   // U+092A
	"फ": {ctg: consonant, hun: "ph"},                  // U+092B
	"ब": {ctg: consonant, hun: "b"},                   // U+092C
	"भ": {ctg: consonant, hun: "bh"},                  // U+092D
	"म": {ctg: consonant, hun: "m"},                   // U+092E
	"य": {ctg: consonant, hun: "y"},                   // U+092F
	"र": {ctg: consonant, hun: "r"},                   // U+0930
	"ऱ": {ctg: consonant, hun: "rh", nuqta: true},     // U+0931
	"ल": {ctg: consonant, hun: "l"},                   // U+0932
	"ळ": {ctg: consonant, hun: "l"},                   // U+0933
	"ऴ": {ctg: consonant, hun: "lh", nuqta: true},     // U+0934
	// Note: Using 'v' for colloquial Hindi (not Hunterian 'w')
	// In Marathi, [w] except [v] before [i]; Hindi uses [v] in most contexts
	"व": {ctg: consonant, hun: "v"},               // U+0935
	"श": {ctg: consonant, hun: "sh"},              // U+0936
	"ष": {ctg: consonant, hun: "sh"},              // U+0937
	"स": {ctg: consonant, hun: "s"},               // U+0938
	"ह": {ctg: consonant, hun: "h"},               // U+0939
	"ऺ": {ctg: unused, hun: ""},                   // U+093A
	"ऻ": {ctg: unused, hun: ""},                   // U+093B
	"़": {ctg: symbol, hun: ""},                   // U+093C
	"ऽ": {ctg: symbol, hun: ""},                   // U+093D
	"ा": {ctg: vowel, hun: "a"},                   // U+093E
	"ि": {ctg: vowel, hun: "i"},                   // U+093F
	"ी": {ctg: vowel, hun: "i"},                   // U+0940
	"ु": {ctg: vowel, hun: "u"},                   // U+0941
	"ू": {ctg: vowel, hun: "u", alt: []alt{"ū"}},  // U+0942 // "oo" or
	"ृ": {ctg: vowel, hun: "ri"},                  // U+0943
	"ॄ": {ctg: vowel, hun: "ri"},                  // U+0944
	"ॅ": {ctg: vowel, hun: "æ"},                   // U+0945
	"ॆ": {ctg: vowel, hun: "e"},                   // U+0946
	"े": {ctg: vowel, hun: "e"},                   // U+0947
	"ै": {ctg: vowel, hun: "ai"},                  // U+0948
	"ॉ": {ctg: vowel, hun: "o"},                   // U+0949
	"ॊ": {ctg: vowel, hun: "o"},                   // U+094A
	"ो": {ctg: vowel, hun: "o"},                   // U+094B
	"ौ": {ctg: vowel, hun: "au"},                  // U+094C
	"्": {ctg: symbol, hun: ""},                   // U+094D
	"ॎ": {ctg: symbol, hun: ""},                   // U+094E
	"ॏ": {ctg: vowel, hun: "au"},                  // U+094F
	"ॐ": {ctg: consonant, hun: "om"},              // U+0950
	"॑": {ctg: unused, hun: ""},                   // U+0951
	"॒": {ctg: unused, hun: ""},                   // U+0952
	"॓": {ctg: unused, hun: ""},                   // U+0953
	"॔": {ctg: unused, hun: ""},                   // U+0954
	"ॕ": {ctg: unused, hun: ""},                   // U+0955
	"ॖ": {ctg: unused, hun: ""},                   // U+0956
	"ॗ": {ctg: unused, hun: ""},                   // U+0957
	"क़": {ctg: consonant, hun: "q", nuqta: true},  // U+0958
	"ख़": {ctg: consonant, hun: "kh", nuqta: true}, // U+0959
	"ग़": {ctg: consonant, hun: "gh", nuqta: true}, // U+095A
	"ज़": {ctg: consonant, hun: "z", nuqta: true},  // U+095B
	"ड़": {ctg: consonant, hun: "r", nuqta: true},  // U+095C
	"ढ़": {ctg: consonant, hun: "rh", nuqta: true}, // U+095D
	"फ़": {ctg: consonant, hun: "f", nuqta: true},  // U+095E
	"य़": {ctg: consonant, hun: "yh", nuqta: true}, // U+095F
	"ॠ": {ctg: consonant, hun: "ri"},              // U+0960
	"ॡ": {ctg: consonant, hun: "ll"},              // U+0961
	"ॢ": {ctg: vowel, hun: "l"},                   // U+0962
	"ॣ": {ctg: vowel, hun: "ll"},                  // U+0963
	"।": {ctg: unused, hun: ""},                   // U+0964
	"॥": {ctg: unused, hun: ""},                   // U+0965
	"०": {ctg: number, hun: "0"},                  // U+0966
	"१": {ctg: number, hun: "1"},                  // U+0967
	"२": {ctg: number, hun: "2"},                  // U+0968
	"३": {ctg: number, hun: "3"},                  // U+0969
	"४": {ctg: number, hun: "4"},                  // U+096A
	"५": {ctg: number, hun: "5"},                  // U+096B
	"६": {ctg: number, hun: "6"},                  // U+096C
	"७": {ctg: number, hun: "7"},                  // U+096D
	"८": {ctg: number, hun: "8"},                  // U+096E
	"९": {ctg: number, hun: "9"},                  // U+096F
	// conjuncts - '़' ड़
	"क़": {ctg: conjuncts, hun: "q"},
	"ख़": {ctg: conjuncts, hun: "kh"},
	"ग़": {ctg: conjuncts, hun: "gh"},
	"ज़": {ctg: conjuncts, hun: "z"},
	"ड़": {ctg: conjuncts, hun: "r"},
	"ढ़": {ctg: conjuncts, hun: "rh", alt: []alt{"ddh"}},
	"फ़": {ctg: conjuncts, hun: "f"},
	"य़": {ctg: conjuncts, hun: "yh"},
	"झ़": {ctg: conjuncts, hun: "zh"},
	"ऩ": {ctg: conjuncts, hun: "nh"},
	"ऱ": {ctg: conjuncts, hun: "rh"},
	"ऴ": {ctg: conjuncts, hun: "lh"},
	// conjuncts - additional
	// Only ज्ञ needs special handling (ज+ञ would give "jny" but "gy" is correct)
	// क्ष, त्र, श्र work correctly as component-wise (क्ष=ksh, त्र=tr, श्र=shr)
	"ज्ञ": {ctg: conjuncts, hun: "gy"},
}

type Symbols struct {
	runes []rune
	index int
}

func (s *Symbols) len() int {
	return len(s.runes)
}

func (s *Symbols) Next() bool {
	next := s.index + 1
	if next >= s.len() {
		return false
	}

	// Check for conjuncts: consonant + halant + consonant (e.g., ज्ञ, क्ष, त्र)
	// If we just processed a conjunct, skip 3 characters
	if next+1 < s.len() && s.runes[next] == '्' {
		conjunct := string([]rune{s.runes[s.index], s.runes[next], s.runes[next+1]})
		if _, found := symbols[conjunct]; found {
			if s.index+3 >= s.len() {
				return false
			}
			s.index = s.index + 3
			return true
		}
	}

	// Check for nuqta or halant: skip 2 characters
	if s.runes[next] == '़' || s.runes[next] == '्' {
		if s.index+2 >= s.len() {
			return false
		}
		s.index = s.index + 2
	} else {
		s.index = s.index + 1
	}
	return true
}

func (s *Symbols) Item() (string, SymbolInfo, bool) {
	next := s.index + 1
	sym := ""

	// Check for conjuncts: consonant + halant + consonant (e.g., ज्ञ, क्ष, त्र)
	if next+1 < s.len() && s.runes[next] == '्' {
		// Try to match a 3-character conjunct
		conjunct := string([]rune{s.runes[s.index], s.runes[next], s.runes[next+1]})
		if si, found := symbols[conjunct]; found {
			return conjunct, si, true
		}
	}

	// Check for nuqta: consonant + nuqta (e.g., क़, ख़)
	if next < s.len() && s.runes[next] == '़' {
		sym = string([]rune{s.runes[s.index], s.runes[next]})
	} else {
		sym = string(s.runes[s.index])
	}
	si, found := symbols[sym]
	return sym, si, found
}

func (s *Symbols) PeekN(n int) (string, *SymbolInfo, bool) {
	if n < 1 {
		return "", nil, false
	}
	next := s.index + n
	nextToNext := next + 1
	if next >= s.len() {
		return "", nil, false
	} else {
		sym := ""
		hasHalant := false
		if nextToNext < s.len() && (s.runes[nextToNext] == '़' || s.runes[nextToNext] == '्') {
			sym = string([]rune{s.runes[next], s.runes[nextToNext]})
			hasHalant = s.runes[nextToNext] == '्'
		} else {
			sym = string(s.runes[next])
		}
		si, found := symbols[sym]
		// If consonant+halant combined form not in map, use base consonant info
		if !found && hasHalant {
			baseSym := string(s.runes[next])
			if baseSi, baseFound := symbols[baseSym]; baseFound {
				return sym, &baseSi, true
			}
		}
		return sym, &si, found
	}
}

// Options configures transliteration behavior.
type Options struct {
	// LongVowels outputs "aa" for all ा (aa-matra) positions.
	LongVowels bool
}

// DefaultOptions returns the default transliteration options.
func DefaultOptions() Options {
	return Options{LongVowels: false}
}

type Hindi struct {
}

func (l Hindi) Name() string {
	return "hindi"
}

func (l Hindi) Info() {
	for i := 2305; i < 2415; i++ {
		sym := string(rune(i))
		t := symbols[sym]
		fmt.Printf("%s - U+0%X - %s - %s", sym, i, t.hun, t.ctg)
	}
}

// Transliterate converts a word using default options.
func (l Hindi) Transliterate(word string) string {
	return l.TransliterateWithOptions(word, DefaultOptions())
}

// TransliterateWithOptions converts a word using the specified options.
func (l Hindi) TransliterateWithOptions(word string, opts Options) string {
	sb := &Symbols{[]rune(word), -1}
	// ्
	var converted = ""
	for sb.Next() {
		currentRaw, si, mapped := sb.Item()
		if !mapped {
			// this is a catch-all,  when the script is unable to find a valid conversion
			// add it back as is
			converted = converted + currentRaw
		} else {
			switch si.ctg {
			case number:
				converted = converted + si.hun
			case vowel:
				// Special handling for ा (aa-matra)
				// Default: "aa" only when followed by consonant at word end (ा + C + END)
				// With LongVowels: "aa" for all ा positions
				if currentRaw == "ा" {
					if opts.LongVowels {
						// LongVowels mode: always use "aa" for ा
						converted = converted + "aa"
					} else {
						// Default mode: "aa" only in closed final syllables
						// Examples: काम→kaam, इंसान→insaan, but गाना→gana
						nxtRaw, nxtSi, nxtExists := sb.PeekN(1)
						if nxtExists && nxtSi.isConsonant() {
							// Check if that consonant is at word end
							nxtNxtRaw, _, nxtNxtExists := sb.PeekN(1 + utf8.RuneCountInString(nxtRaw))
							if !nxtNxtExists || nxtNxtRaw == "" {
								converted = converted + "aa"
							} else {
								converted = converted + si.hun
							}
						} else {
							converted = converted + si.hun
						}
					}
				} else {
					converted = converted + si.hun
				}
			case consonant, conjuncts:
				nxtIndex := utf8.RuneCountInString(currentRaw)
				nxtRaw, nxtSi, nxtExists := sb.PeekN(nxtIndex)
				rom := si.hun

				// Special handling for व: use 'w' only in specific conjuncts (स्व, श्व, द्व, ख्व)
				// These conjuncts have a semivowel 'w' sound in Hindi
				// Examples: स्वागत→swagat, ऐश्वर्या→aishwarya, द्वार→dwaar, ख्वाब→khwaab
				// But NOT for र्व, त्व, etc. which keep 'v' sound (पर्वत→parvat, तत्व→tatva)
				if currentRaw == "व" && sb.index > 1 && sb.runes[sb.index-1] == '्' {
					prevConsonant := sb.runes[sb.index-2]
					// Use 'w' only after स, श, द, ख (common semivowel conjuncts)
					if prevConsonant == 'स' || prevConsonant == 'श' || prevConsonant == 'द' || prevConsonant == 'ख' {
						rom = "w"
					}
				}

				// Check if this consonant follows a halant (part of a conjunct)
				// Only protect schwa for word-initial conjuncts:
				// - Halant at index 1: C्C pattern (e.g., प्र in प्रकाश)
				// - After independent vowel: अC्C pattern (e.g., अध्य in अध्यक्ष)
				isAfterHalant := sb.index > 0 && sb.runes[sb.index-1] == '्'
				isWordInitialConjunct := false
				if isAfterHalant {
					halantIdx := sb.index - 1
					// Word-initial if halant is at index 1 (C्C)
					// or at index 2 after independent vowel (अC्C)
					switch halantIdx {
					case 1:
						isWordInitialConjunct = true
					case 2:
						// Check if index 0 is independent vowel (अ-औ)
						firstChar := sb.runes[0]
						if firstChar >= 0x0905 && firstChar <= 0x0914 {
							isWordInitialConjunct = true
						}
					}
				}

				// र् = 'र'+'्' (reph) is handled automatically
				// consonant + '्' + 'र' (rakar) needs to be handled
				if nxtExists && nxtSi.isConsonant() {
					// get next to next
					nxtToNxtIndex := nxtIndex + utf8.RuneCountInString(nxtRaw)
					_, nxtToNxtSi, nxtToNxtExists := sb.PeekN(nxtToNxtIndex)
					// Suppress schwa only if: not at start, not word-initial conjunct, and followed by C+V
					if sb.index != 0 && !isWordInitialConjunct && nxtToNxtExists && nxtToNxtSi.isVowel() {
						converted = converted + rom
					} else {
						converted = converted + rom + "a"
					}
				} else {
					// Add schwa if:
					// - followed by anusvara, OR
					// - at word end AND part of final conjunct ending in र, य, or व
					//   (Sanskrit words like mantra, chandra, karya, kshetra, indra retain final 'a')
					// - at word end AND य follows ी (adjective suffix -iya: केंद्रीय→kendriya)
					isWordEnd := !nxtExists || nxtRaw == ""
					isSonorousFinal := currentRaw == "र" || currentRaw == "य" || currentRaw == "व"

					// Check for ीय pattern (adjective ending)
					isIyaEnding := false
					if isWordEnd && currentRaw == "य" && sb.index > 0 {
						prevChar := sb.runes[sb.index-1]
						if prevChar == 'ी' { // long i matra
							isIyaEnding = true
						}
					}

					if nxtRaw == "ं" || (isWordEnd && isAfterHalant && isSonorousFinal) || isIyaEnding {
						converted = converted + rom + "a"
					} else {
						converted = converted + rom
					}
				}
			case symbol:
				// symbols are ignored
			default:
				converted = converted + si.hun
			}
		}
	}
	return converted
}
