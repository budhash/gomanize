# Gomanize

[![CI](https://github.com/budhash/gomanize/actions/workflows/ci.yml/badge.svg)](https://github.com/budhash/gomanize/actions/workflows/ci.yml)

A Go library and CLI tool for transliterating Devanagari script (Hindi) to Latin/Roman characters. Designed for practical use cases like song lyrics romanization.

## Features

- Romanize Hindi text from Devanagari script into readable Latin characters
- CLI tool for quick transliteration (whitespace- and punctuation-aware, works on full lyrics)
- Library API for integration into Go projects
- Based on the Hunterian transliteration system (India's national standard)
- Optimized for colloquial/phonetic Hindi (no diacritics)
- Optional embedded learned components — zero runtime dependencies:
  - `--schwa-model`: decision-tree schwa classifier (90.7% held-out per-schwa accuracy)
  - `--lexicon`: 8,367 human-attested spellings for common words & loanwords (अंकल → uncle)
  - `--rerank`: character-LM picks the best candidate — improves every benchmark
- Validated against five independent benchmark suites (see Current Status)

## Installation

```bash
# As a library
go get github.com/budhash/gomanize

# Build CLI from source
git clone https://github.com/budhash/gomanize
cd gomanize
make build
```

## Usage

### CLI

```bash
# Direct argument
./gomanize "नमस्ते भारत"
# Output: namaste bharat

# Pipe input
echo "हिंदी गाना" | ./gomanize
# Output: hindi gana

# Options
./gomanize --long-vowels "गाना"        # gaanaa (aa for all ā positions)
./gomanize --simple-nasals "करें"       # karen (simplified nasal endings)
./gomanize --keep-medial-schwa "जनता"  # janata (retain medial schwa)
./gomanize --schwa-model "जनता"        # janta (learned schwa classifier)
./gomanize --lexicon "अंकल"            # uncle (attested spellings for known words)
./gomanize --rerank "जनता"             # janta (best of rules/model candidates via char LM)
./gomanize --list-rules                # inspect the rule catalog
./gomanize --debug "नमस्ते"            # trace rule applications

# Version
./gomanize --version
```

### Library

```go
package main

import (
    "fmt"
    gomanize "github.com/budhash/gomanize"
)

func main() {
    g, err := gomanize.New("hindi")
    if err != nil {
        panic(err)
    }

    output := g.Translit("नमस्ते दुनिया")
    fmt.Println(output)  // "namaste duniya"
}
```

## Transliteration System

Gomanize is based on the **Hunterian transliteration system**, the national romanization standard of India, with adaptations for modern colloquial Hindi.

### Design Principles

| Principle | Description | Example |
|-----------|-------------|---------|
| **No diacritics** | Uses ASCII-only output | आ → aa (not ā) |
| **Schwa deletion** | Matches spoken Hindi pronunciation | करना → karna (not karanā) |
| **Phonetic spelling** | Readable without linguistic training | ख → kh (not ḵẖ) |
| **Long vowel doubling** | Distinguishes vowel length in key positions | काम → kaam |

### Character Mappings

#### Vowels

| Devanagari | Independent | Matra | Gomanize | Notes |
|------------|-------------|-------|----------|-------|
| अ | अ | (inherent) | a | Short a (schwa) |
| आ | आ | ा | a, aa | "aa" when ा+C at word end |
| इ | इ | ि | i | Short i |
| ई | ई | ी | i | Long i (no distinction) |
| उ | उ | ु | u | Short u |
| ऊ | ऊ | ू | u | Long u (no distinction) |
| ऋ | ऋ | ृ | ri | Vocalic r |
| ए | ए | े | e | |
| ऐ | ऐ | ै | ai | |
| ओ | ओ | ो | o | |
| औ | औ | ौ | au | |

#### Consonants

| Devanagari | Gomanize | Devanagari | Gomanize |
|------------|----------|------------|----------|
| क | k | ट | t |
| ख | kh | ठ | th |
| ग | g | ड | d |
| घ | gh | ढ | dh |
| ङ | n | ण | n |
| च | ch | त | t |
| छ | chh | थ | th |
| ज | j | द | d |
| झ | jh | ध | dh |
| ञ | ny | न | n |
| प | p | य | y |
| फ | ph | र | r |
| ब | b | ल | l |
| भ | bh | व | v |
| म | m | श | sh |
| ष | sh | स | s |
| ह | h | | |

#### Nuqta Consonants (Perso-Arabic loans)

| Devanagari | Gomanize | Used for |
|------------|----------|----------|
| क़ | q | Arabic ق |
| ख़ | kh | Arabic/Persian خ |
| ग़ | gh | Arabic غ |
| ज़ | z | Arabic ز |
| फ़ | f | Arabic/Persian ف |
| ड़ | r | Flapped r |
| ढ़ | rh | Aspirated flapped r |

#### Common Conjuncts

| Devanagari | Gomanize | Example |
|------------|----------|---------|
| क्ष | ksh | क्षमा → kshama |
| त्र | tr | त्रिशूल → trishul |
| ज्ञ | gy | ज्ञान → gyaan |
| श्र | shr | श्री → shri |

### Schwa Deletion Rules

Hindi exhibits **schwa deletion** where the inherent 'a' vowel is not pronounced in certain positions. Gomanize implements these rules:

1. **Word-final deletion**: Final schwa is deleted
   - भारत → bharat (not bharata)

2. **Medial deletion**: Schwa deleted in C+C+V patterns
   - समझना → samajhna (not samajhana); जनता → janta; करना → karna

3. **Preserved positions**:
   - Word-initial conjuncts: प्रकाश → prakaash (not prkaash)
   - Short words where deletion would break the syllable: करम → karam, कमल → kamal
   - Sanskrit word endings with र, य, व: मंत्र → mantra, कार्य → karya

An optional learned classifier (`--schwa-model`, a decision tree distilled from
force-aligned Dakshina data) can take over these decisions: 90.7% per-schwa
accuracy on held-out words.

### Long Vowel "aa" Rule

The aa-matra (ा) outputs "aa" when followed by a consonant at word end:
- काम → kaam (not kam)
- इंसान → insaan
- अभिमान → abhimaan

But remains "a" in other positions:
- कामना → kamna (medial position)
- गाना → gana (word-final open syllable)

## Comparison with Other Standards

| Character | ISO 15919 | UN (1977) | Hunterian | Gomanize |
|-----------|-----------|-----------|-----------|----------|
| आ | ā | ā | ā, a | aa, a |
| ई | ī | ī | ī, i | i |
| ऊ | ū | ū | ū | u |
| च | ca | cha | cha | ch |
| छ | cha | chha | chha | chh |
| व | va | va | wa, v- | v |
| श | śa | sha | sa, sha | sh |
| ज्ञ | jña | jña | gy | gy |
| ं | ṁ | ṁ | n, m | n |

### Key Differences from Hunterian

1. **No diacritics**: We use "aa" instead of "ā", making output ASCII-compatible
2. **व mapping**: We use "v" by default, but "w" in specific conjuncts (स्व, श्व, द्व, ख्व)
3. **Schwa deletion**: We apply aggressive deletion to match spoken Hindi
4. **Sanskrit finals**: We retain final 'a' for readability (मंत्र→mantra, चंद्र→chandra)

### Deliberate Divergences from Dakshina Dataset

Some of our romanization choices prioritize phonetic accuracy over matching the Dakshina dataset:

| Word | Dakshina | Gomanize | Rationale |
|------|----------|----------|-----------|
| स्वागत | swagat | swagat | ✓ Match (conjunct w) |
| विश्व | vishwa | vishwa | ✓ Match (conjunct w) |
| देव | dev | dev | ✓ Match (final v, not w) |
| पर्वत | parvat | parvat | ✓ Match (rv keeps v) |
| मंत्र | mantr | mantra | Final 'a' for readability |
| चंद्र | chandr | chandra | Final 'a' for readability |

## Current Status

Romanization is many-to-one (जनता = *janata* / *janta* / *janataa* are all
human-attested), so accuracy is scored against **all attested variants**:

| Benchmark | Result |
|-----------|--------|
| Curated Dakshina, match-any + `--rerank` | **94.7%** |
| Curated Dakshina, match-any (default rules) | 92.8% (minCER 0.0116) |
| COMI-LINGUA naturally-typed Hindi, token-weighted (+`--lexicon`) | 85.7% |
| Held-out Dakshina test (2,500 unseen words, `--rerank`) | 70.4% |
| **Song lyrics gold set, line CER** | **0.047 — at the human consistency floor (≈0.054)** |

On the flagship use case — romanizing lyrics — gomanize's character-level error
is comparable to how much human romanizers disagree with themselves.

See [CLAUDE.md](CLAUDE.md) for full status and `docs/reviews/` for dated
decision records of every result (including negative ones).

## References

### Standards Documentation

- [ISO 15919:2001](https://www.iso.org/standard/28333.html) - International standard for Indic transliteration
- [UNGEGN Romanization](https://www.eki.ee/wgrs/) - United Nations romanization systems
- [Hunterian Transliteration](https://en.wikipedia.org/wiki/Hunterian_transliteration) - India's national standard
- [ALA-LC Romanization](https://www.loc.gov/catdir/cpso/roman.html) - Library of Congress standard

### Technical References

- [Hindi-Marathi-Nepali Transliteration Tables (PDF)](docs/reference/Hindi-Marathi-Nepali-Transliteration.pdf) - Comprehensive comparison by Thomas T. Pedersen ([online](https://transliteration.eki.ee/pdf/Hindi-Marathi-Nepali.pdf))
- [transliteration.eki.ee](https://transliteration.eki.ee) - Collection of transliteration systems
- [ushuaia.pl/transliterate](https://www.ushuaia.pl/transliterate/) - Online transliteration tool with Hunterian, ISO-15919, and Polish schemes
- [Schwa Deletion in Indo-Aryan Languages](https://en.wikipedia.org/wiki/Schwa_deletion_in_Indo-Aryan_languages)
- [Devanagari Transliteration](https://en.wikipedia.org/wiki/Devanagari_transliteration)

### Test Data

- [Google Dakshina Dataset](https://github.com/google-research-datasets/dakshina) - Human-attested romanization pairs (CC BY-SA 4.0)
- [Aksharantar](https://huggingface.co/datasets/ai4bharat/Aksharantar) (AI4Bharat) - Human-annotated Hindi test set, 10,112 pairs (CC-BY 4.0)
- [COMI-LINGUA](https://huggingface.co/datasets/LingoIITGN/COMI-LINGUA) - Naturally-typed parallel Devanagari/Roman Hindi (CC-BY 4.0)
- [Shabd](https://osf.io/xfbhd/) - Hindi word-frequency database (CC0), used for frequency weighting and lexicon ranking
- Public-domain lyrics gold seed (Kabir, Meera, Tagore, et al.) - in-repo, line-level evaluation

## Development

```bash
# First-time setup
make init

# Development workflow
make dev      # Format, vet, test
make ci       # Full CI pipeline

# Testing
make test           # Run all tests
make test-unit      # Unit tests only (fast)
make test-cover     # Tests with coverage
make test-dakshina  # Dakshina accuracy tests
make test-analysis  # Failure pattern breakdown

# See all commands
make help
```

## License

MIT License - Copyright (c) 2023-2026 Budhaditya

## Author

Budhaditya (budhash@gmail.com)
