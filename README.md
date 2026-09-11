# Gomanize

[![CI](https://github.com/budhash/gomanize/actions/workflows/ci.yml/badge.svg)](https://github.com/budhash/gomanize/actions/workflows/ci.yml)

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
# Go library
go get github.com/budhash/gomanize

# JavaScript/TypeScript (Node or browser) — the same engine as WebAssembly
npm install @budhash/gomanize

# CLI from source
git clone https://github.com/budhash/gomanize
cd gomanize
make build
```

## Usage

### CLI

```bash
./gomanize "नमस्ते दुनिया"       # namaste duniya
echo "हिंदी गाना" | ./gomanize   # hindi gana
```

| Flag | Effect | Example |
|------|--------|---------|
| (default) | Colloquial rules | जनता → janta |
| `--keep-medial-schwa` | Retain medial schwa | जनता → janata |
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
COMI-LINGUA (CC-BY 4.0), and Shabd (CC0); see
[docs/RESEARCH.md](docs/RESEARCH.md) for full attribution.
