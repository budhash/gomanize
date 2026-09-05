# Gomanize

A Go library and CLI for transliterating Devanagari (Hindi) into Latin script,
built for song lyrics and other colloquial text. MIT licensed (2023-2026);
Go 1.21+.

This file covers development workflow and repository conventions. For users:
`README.md`. For research background, datasets, and results: `docs/RESEARCH.md`.
For architecture and the rule system: `docs/DESIGN.md`.

## Quick Start

```bash
make init                      # First-time setup (deps + pre-commit hooks)
make build

./gomanize "नमस्ते भारत"        # namaste bharat
echo "हिंदी गाना" | ./gomanize   # hindi gana
# Flags (--lexicon, --rerank, --schwa-model, ...): see README.md

make ci                        # Full pipeline before any PR
make help                      # All commands
```

## Development Workflow

```bash
# Setup
make init           # First-time setup (deps + pre-commit hooks)
make hooks          # Install pre-commit hooks
make hooks-update   # Update pre-commit hook versions

# Development
make dev            # Full workflow: format, vet, test
make ci             # CI pipeline: fmt-check, lint, build, test, benchmarks

# Testing
make test           # Run all tests
make test-unit      # Unit tests only (fast; all packages except benchmark)
make test-cover     # Tests with coverage
make test-dakshina  # Curated Dakshina accuracy (pure + overrides + CER)
make test-integration # Full Dakshina + Aksharantar bulk runs
make test-analysis  # Failure pattern breakdown
# Individual suites: go test ./benchmark/... -run TestBenchmark<Name> -v
#   MultiReference | SchwaModelHeldout | LexiconCoverage | FrequencyWeighted |
#   AksharantarTestSet | ComiLingua | LyricsGold

# Code Quality
make fmt / fmt-check / lint / lint-fix

# Utilities
make demo           # Demo with sample words
make bench          # Performance benchmarks
make tasks ARGS=".." # Task tracker passthrough
```

## Task Tracking & Process

`TASKS.md` is the single source of truth for features and tasks, managed by a
vendored CLI (`tools/tasks.py`, canonical: <https://github.com/budhash/tasks>).
**Edit it only through `./tools/tasks`** — never by hand. Common verbs: `tree`,
`next`, `new`, `start`, `done`, `validate`. Full process, PR discipline, and
accuracy-reporting rules: [`docs/PROCESS.md`](docs/PROCESS.md).

## Architecture

```
gomanize/
├── gomanize.go                    # Public API (New, Translit — whitespace/punct-aware)
├── cmd/gomanize/main.go           # CLI entry point (all flags)
├── core/                          # Engine mechanics (no script knowledge)
│   ├── engine.go                  # Pipeline + LexiconProvider/Reranker interfaces
│   ├── types.go                   # Unit, Word, Options
│   └── rule.go                    # Rule definitions, phases, scopes, modes
├── lang/hindi/                    # Hindi language implementation
│   ├── symbols.go / rules.go      # Symbol map + Hindi-specific rules
│   ├── schwa_model.go + schwa_tree.json    # Learned schwa classifier (embedded, 34 KB)
│   ├── lexicon.go + lexicon.tsv            # 8,367-word attested lexicon (embedded, ~204 KB)
│   └── reranker.go + roman_ngrams.tsv      # Char 4-gram re-ranker (embedded, 224 KB)
├── scheme/colloquial/             # Colloquial romanization scheme
├── script/brahmic/                # Brahmic script support (shared by future languages)
│   ├── brahmic.go / parser.go / renderer.go / runs.go
│   └── schwa_rules.go             # Shared Brahmic schwa rules (brahmic.SchwaRules())
├── benchmark/                     # Five evaluation suites
│   ├── benchmark_test.go          # All benchmark tests
│   ├── metrics_test.go            # CER / minCER / match-any / reference loaders
│   └── data/                      # Datasets (licenses: docs/RESEARCH.md §3)
├── tools/                         # All dev tooling
│   ├── tasks, tasks.py            # Task tracker CLI over TASKS.md
│   ├── ushuaia                    # Compare against ushuaia.pl schemes
│   ├── schwa/                     # Schwa classifier training + build_lexicon.py
│   └── build_freq.py / build_aksharantar_test.py / build_comilingua.py /
│       train_ngram.py / mine_overrides.py
├── datasets/                      # Raw-dataset download/generation scripts
├── docs/
│   ├── RESEARCH.md                # Problem, literature, datasets, methodology, results
│   ├── DESIGN.md                  # Architecture, rule system, scheme, learned components
│   ├── PROCESS.md                 # Task tracking, PR discipline, accuracy reporting
│   ├── reviews/                   # Dated decision records (incl. negative results)
│   ├── archive/                   # Historical docs (2025-era, bannered)
│   └── reference/                 # External reference material
├── .claude/                       # Claude Code configuration + hooks
├── .github/workflows/             # ci.yml + release.yml (GoReleaser on tags)
├── Makefile / TASKS.md / README.md / CLAUDE.md
```

`internal/legacy_lang` and `scripts/` were removed in the pre-1.0 reorg (PR
#54); git history preserves them.

## Current Status

Full results: [`docs/RESEARCH.md`](docs/RESEARCH.md) §4. The CI regression gate
is strict top-1 pure ≥85% on curated Dakshina (`make test-dakshina`); overrides
are an exception list, never headline numbers. Remaining failures are lexical,
not rule-governed — the evidence, including rejected rules, is in RESEARCH §5.

Contamination rules (train-split-only training; benchmarks never mined):
RESEARCH §2. Deliberate divergences from Dakshina/Hunterian conventions:
[`docs/DESIGN.md`](docs/DESIGN.md) §3.

## Ushuaia Comparison Tool

Compare output against [ushuaia.pl](https://www.ushuaia.pl/transliterate/)
schemes:

```bash
./tools/ushuaia "नमस्ते" --compare   # prints Hunterian vs gomanize side by side
```

Run `./tools/ushuaia --help` for other schemes (`--all`, `--iso`).

## CI/CD

Pre-commit hooks run on every commit: block commits to main/master, format
(`make fmt`), lint (`make lint`), validate `TASKS.md`, and catch
whitespace/merge-conflict issues. Install via `make init`; run manually with
`pre-commit run --all-files`; skip sparingly with `git commit --no-verify`.

CI on push/PR to main: format check, lint, build, test with coverage,
accuracy benchmarks.

Releases are automated by GoReleaser on version tags:

```bash
git tag v1.0.0 && git push origin v1.0.0
```

This publishes binaries (Linux/macOS/Windows, amd64+arm64), checksums, and a
changelog; verify with `./gomanize --version`.

## Roadmap

Live backlog: `TASKS.md` via `./tools/tasks tree`. History and reasoning:
`docs/reviews/`. Post-1.0 directions with tradeoffs: [`docs/ROADMAP.md`](docs/ROADMAP.md)
(seeded as features F-0006–F-0008); design constraints in [`docs/DESIGN.md`](docs/DESIGN.md) §6.
