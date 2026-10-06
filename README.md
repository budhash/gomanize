# Gomanize

[![CI](https://github.com/budhash/gomanize/actions/workflows/ci.yml/badge.svg)](https://github.com/budhash/gomanize/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/@budhash/gomanize?logo=npm)](https://www.npmjs.com/package/@budhash/gomanize)
[![Go Reference](https://pkg.go.dev/badge/github.com/budhash/gomanize.svg)](https://pkg.go.dev/github.com/budhash/gomanize)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Go library and CLI that romanizes Devanagari (Hindi) into readable Latin
script, built for song lyrics and other colloquial text. Rule-based engine with
optional embedded learned components; no runtime dependencies.

```
नमस्ते दुनिया  →  namaste duniya
```

**Try it in your browser → [budhash.com/gomanize](https://budhash.com/gomanize)** —
the full engine, compiled to WebAssembly, runs entirely client-side (no server;
no text leaves your machine).

**Use it from JavaScript/TypeScript → [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize) on npm** —
the same engine as a WebAssembly package for Node and the browser (not a
reimplementation, so output is byte-identical to the CLI).

It does *romanization* — spelling Hindi the way it sounds (नमस्ते → *namaste*) —
not the strict, reversible *transliteration* of IAST or ISO 15919. There is no
single correct answer (जनता is validly *janata*, *janta*, or *janataa*), so it
is measured against all human-attested variants; on verse, its error rate is at
the level where human romanizers disagree. Background, methodology, and
limitations: [docs/RESEARCH.md](docs/RESEARCH.md).

## Install

```bash
# CLI via Homebrew (macOS/Linux)
brew install budhash/tools/gomanize

# Go library
go get github.com/budhash/gomanize

# JavaScript/TypeScript (Node or browser) — the same engine as WebAssembly
npm install @budhash/gomanize

# CLI from source
git clone https://github.com/budhash/gomanize
cd gomanize
make build
```

Prebuilt binaries for Linux/macOS/Windows (amd64+arm64) are also attached to each
[GitHub release](https://github.com/budhash/gomanize/releases).

## Usage

### CLI

```bash
gomanize "नमस्ते दुनिया"       # namaste duniya
echo "हिंदी गाना" | gomanize   # hindi gana

# Batch a file — one word/phrase per line (# comments and blank lines skipped)
gomanize --input=lyrics.txt
gomanize --input=words.txt --lexicon --rerank   # flags apply
gomanize < lyrics.txt                           # stdin/pipe also works
```

| Flag | Effect | Example |
|------|--------|---------|
| (default) | Colloquial rules | जनता → janta |
| `--keep-medial-schwa` | Retain medial schwa | जनता → janata |
| `--language=NAME` | `hindi` (default) or experimental `bengali` | `--language bengali` also works |
| `--long-vowels` | aa for every ā | गाना → gaanaa |
| `--simple-nasals` | Simplified nasal endings | करें → karen |
| `--schwa-model` | Learned schwa classifier | जनता → janta |
| `--lexicon` | Attested spellings for 8,367 known words | अंकल → uncle |
| `--rerank` | Character-LM picks best of rules/schwa-model outputs | see Accuracy below |
| `--list-rules`, `--debug` | Inspect and trace the rule engine | |

### Library (Go)

```go
import gomanize "github.com/budhash/gomanize"

g, err := gomanize.New("hindi")
if err != nil {
    panic(err)
}
fmt.Println(g.Translit("नमस्ते दुनिया")) // "namaste duniya"
```

Options mirror the CLI flags via `gomanize.NewWithOptions`.

### Library (JavaScript/TypeScript)

The [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize) package
is the same engine compiled to WebAssembly — not a reimplementation, so output is
byte-identical to the Go CLI. Works in Node and the browser:

```js
import { load } from "@budhash/gomanize";

const g = await load();
g.translit("नमस्ते दुनिया");              // "namaste duniya"
g.translit("गाना", { longVowels: true }); // "gaanaa"
```

The same flags are accepted as options (`longVowels`, `simpleNasals`,
`keepMedialSchwa`, `schwaModel`, `lexicon`, `rerank`).

## Accuracy

Scored against all human-attested romanization variants (see
[docs/RESEARCH.md](docs/RESEARCH.md) for methodology, datasets, and the full
result set including negative results):

| Benchmark | Result |
|-----------|--------|
| Curated Dakshina | 86.2% exact / 92.9% any attested variant (94.8% with `--rerank`) |
| Naturally-typed Hindi (COMI-LINGUA), token-weighted | 79.9% (86.6% with `--lexicon`) |
| Held-out unseen words | 69.3% (70.7% with `--rerank`) |
| Song lyrics, line-level character error | 0.049, or 0.039 with `--lexicon` (human agreement floor is about 0.054) |

Known limitations (vowel-length spelling, named entities, cross-convention
scoring): [docs/RESEARCH.md §4](docs/RESEARCH.md).

## Documentation

| Document | Contents |
|----------|----------|
| [CHANGELOG.md](CHANGELOG.md) | Release history and notable changes |
| [docs/RESEARCH.md](docs/RESEARCH.md) | The problem, literature, datasets and licenses, evaluation methodology, all results including negatives |
| [docs/DESIGN.md](docs/DESIGN.md) | Engine architecture, rule system, character mappings, learned components, future directions |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Post-1.0 directions (convention schemes, more languages) with tradeoffs |
| [web/README.md](web/README.md) | The browser demo ([budhash.com/gomanize](https://budhash.com/gomanize)): `make wasm` / `make wasm-serve` and the Pages deploy |
| [CLAUDE.md](CLAUDE.md) | Development workflow, commands, repository conventions |
| [docs/PROCESS.md](docs/PROCESS.md) | Task tracking, PR discipline, accuracy reporting rules |
| [docs/reviews/](docs/reviews/) | Dated decision records for every result, including failures |

## Development

```bash
make init     # First-time setup (deps + pre-commit hooks)
make ci       # Full pipeline: format check, lint, build, tests, benchmarks
make help     # All commands
```

Contributions follow [docs/PROCESS.md](docs/PROCESS.md): feature branches,
`make ci` before PRs, accuracy changes must show before/after on the benchmark
suite.

## License

MIT. Copyright (c) 2023-2026 Budhaditya (budhash@gmail.com).

Benchmark data derives from Dakshina (CC BY-SA 4.0), Aksharantar (CC-BY 4.0),
COMI-LINGUA (CC-BY 4.0), and Shabd (CC0). Embedded data files are licensed
separately from the MIT code: the spelling lexicons (`lang/hindi/lexicon.tsv`,
`lang/bengali/lexicon.tsv`) and the Bengali selector (`lang/bengali/selector.json`,
which also uses Google's Bengali pronunciation data, CC BY 4.0) derive from
Dakshina and are CC BY-SA 4.0; the Bengali vowel model
(`lang/bengali/vowel_tree.json`) derives from Google's Bengali pronunciation
lexicon (CC BY 4.0). Experimental Bengali training data also
derives from Google's Bengali pronunciation lexicon (CC BY 4.0); see
[docs/RESEARCH.md](docs/RESEARCH.md) for full attribution.

## Experimental Bengali support

The Go library accepts `gomanize.New("bengali")`. Experimental B1 adds scoped
phalas, positional conjuncts, final-cluster vowels, and হও/হওয়া handling to the
B0 symbols and Unicode aliases. It reaches 56.52% match-any on the held-out
Dakshina word set; pronunciation ambiguities remain. The CLI accepts `--language=bengali` (or
`--language bengali`); omitting it preserves Hindi. npm accepts `{ language: "bengali" }`, and the browser has an explicit language
selector. Both continue to default to Hindi. See the [B1 results and
limitations](docs/reviews/2026-10-04-bengali-b1-results.md).


`SchwaModel: true` enables the [vowel model](docs/reviews/2026-10-04-bengali-b2-vowels.md),
which raises held-out match-any to 62.76%. `Lexicon: true` adds 8,979 attested
training spellings; it helps known words and leaves the excluded held-out score
unchanged. Both are opt-in and can be combined. Alternate-style flags bypass the
lexicon so the requested style is preserved. See the [lexicon evaluation](docs/reviews/2026-10-04-bengali-b2-lexicon.md).

Add `Rerank: true` alongside `SchwaModel: true` to enable the experimental
[native selector](docs/reviews/2026-10-05-bengali-b2-runtime.md): held-out match-any
is 63.04%. It preserves alternate styles by bypassing selection, and lexicon
hits still win first. External word results improve slightly; sentence character
error worsens slightly. It remains opt-in, pending independent lyrics validation.

```go
g, err := gomanize.New("bengali")
if err != nil {
    panic(err)
}
fmt.Println(g.Translit("আমি বাংলা")) // ami bangla
```

```sh
gomanize --language=bengali "আমি বাংলা"  # ami bangla
echo "আমি বাংলা" | gomanize --language bengali --schwa-model --rerank
gomanize --language=bengali --input=lyrics.txt
gomanize --language=bengali --list-rules
```

Language selection also applies to `--test`, `--debug`, and rule overrides.
Only `hindi` and `bengali` are accepted (case-insensitive); unsupported or empty
languages fail with a nonzero exit. Bengali reranking needs `--schwa-model`;
Hindi style flags do not implement Bengali rendering styles; supplying them
bypasses Bengali lexicon lookup and reranking. The browser hides those Hindi
style controls while Bengali is selected.

```js
const g = await load();
g.translit("আমি বাংলা", { language: "bengali" }); // ami bangla
g.translit("नमस्ते दुनिया"); // namaste duniya — default remains Hindi
```
