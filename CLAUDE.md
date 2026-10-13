# Gomanize

A Go library and CLI that romanizes Hindi (Devanagari) and, experimentally,
Bengali into Latin script, built for song lyrics and other colloquial text
(pronunciation-based romanization, not strict reversible transliteration; the
public API keeps the `Translit` name, and "transliteration" is the common
colloquial term). Code MIT licensed (2023-2026); embedded data licenses in
`NOTICE.md`. Go 1.21+.

This file covers development workflow and repository conventions. For users:
`README.md`. For research background, datasets, and results: `docs/RESEARCH.md`.
For architecture and the rule system: `docs/DESIGN.md`.

## Quick Start

```bash
make init                      # First-time setup (deps + pre-commit hooks)
make build

./gomanize "नमस्ते दुनिया"       # namaste duniya
echo "हिंदी गाना" | ./gomanize   # hindi gana
# Flags (--lexicon, --rerank, --schwa-model, ...): see README.md

make ci                        # Full pipeline before any PR
make wasm-serve                # Build + serve the browser (WASM) demo locally
make help                      # All commands
```

Live demo: <https://budhash.com/gomanize> — the engine compiled to WebAssembly,
auto-deployed from `web/` by `.github/workflows/pages.yml` on push to `main`.

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
# Bengali (experimental; all run in `make ci`):
make test-bengali            # Fixtures, symbols, rules, pinned outputs
make test-bengali-b1-gate    # Preregistered B1 dev gate
make test-bengali-data       # Frozen training partitions + isolation guards
make test-bengali-vowels / -lexicon / -crossfit / -runtime / -lyrics
make test-bengali-rerank / -native-selector   # Rejected-experiment records

# Code Quality
make fmt / fmt-check / lint / lint-fix

# Utilities
make demo           # Demo with sample words
make bench          # Performance benchmarks
make tasks ARGS=".." # Task tracker passthrough

# Web / WASM demo
make wasm           # Build web/gomanize.wasm + copy wasm_exec.js + vendored loader (gomanize.mjs)
make wasm-serve     # Build + serve the demo at http://localhost:8080

# npm package (@budhash/gomanize — WASM engine + JS wrapper)
make npm            # Assemble npm/ (copies wasm + wasm_exec.js into npm/dist)
make npm-test       # Build + smoke-test the package under Node
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
├── cmd/gomanize-wasm/main.go      # WebAssembly entry point (syscall/js bridge, js&&wasm)
├── core/                          # Engine mechanics (no script knowledge)
│   ├── engine.go                  # Pipeline (canonicalize → lexicon → rerank → parse → rules → render) + provider interfaces
│   ├── interfaces.go              # Language / Script / Canonicalizer interfaces
│   ├── types.go                   # Unit, Word, Options (DefaultStyle)
│   └── rule.go                    # Rule definitions, phases, scopes, modes
├── lang/hindi/                    # Hindi language implementation
│   ├── symbols.go / rules.go      # Symbol map + Hindi-specific rules
│   ├── schwa_model.go + schwa_tree.json    # Learned schwa classifier (embedded, 34 KB)
│   ├── lexicon.go + lexicon.tsv            # 8,367-word attested lexicon (embedded, ~204 KB)
│   └── reranker.go + roman_ngrams.tsv      # Char 4-gram re-ranker (embedded, 224 KB)
├── lang/bengali/                  # Experimental Bengali (rules, opt-in learned components)
│   ├── symbols.go / rules.go      # Symbol map, canonical forms, B1 rules
│   ├── vowel_model.go + vowel_tree.json    # Three-class inherent-vowel model (Google lexicon, CC BY 4.0)
│   ├── lexicon.go + lexicon.tsv + LEXICON.md   # 8,976-entry spelling lexicon (CC BY-SA 4.0)
│   └── reranker.go + selector.json + RERANKER.md # Cross-fitted native selector (CC BY-SA 4.0)
├── scheme/colloquial/             # Colloquial romanization scheme
├── webdemo/                       # Host-testable JS-flag → Options mapping (WASM demo)
├── web/                           # Browser demo: index.html → budhash.com/gomanize
├── npm/                           # npm package @budhash/gomanize (WASM engine + JS/TS wrapper)
├── script/brahmic/                # Brahmic script support (shared by Hindi and Bengali)
│   ├── brahmic.go / parser.go / renderer.go / runs.go / types.go
│   ├── profile.go                 # ScriptProfile (per-script rendering constants)
│   ├── canonical.go               # Canonicalize: per-script NFC + format-character policy
│   └── schwa_rules.go             # Shared Brahmic schwa rules (brahmic.SchwaRules())
├── hindi_snapshot_test.go + testdata/hindi_snapshot/  # Frozen Hindi outputs (264,064 inputs × 13 profiles)
├── invariants_test.go             # Tier 2 parser invariants + construct goldens (F-0010)
├── benchmark/                     # Evaluation + parser-QA suites (accuracy, construct coverage)
│   ├── benchmark_test.go          # Curated/Dakshina/Aksharantar/COMI/lyrics/frequency suites
│   ├── metrics_test.go            # CER / minCER / match-any / reference loaders
│   ├── construct_test.go          # Tier 1: per-construct accuracy analyzer
│   ├── heldout_test.go            # Held-out construct regression set (T-0044)
│   ├── akshara_test.go            # Tier 4: unit-vs-akshara segmentation differential
│   ├── bengali_*_test.go          # Bengali benchmarks, gate, pins and record assertions
│   └── data/                      # Datasets incl. bengali/, banglatlit/, bengali_lyrics/ (licenses: docs/RESEARCH.md §3)
├── training/data/bengali/         # Frozen Bengali training partitions (schema 2) + manifest
├── tools/                         # All dev tooling
│   ├── tasks, tasks.py            # Task tracker CLI over TASKS.md
│   ├── ushuaia                    # Compare against ushuaia.pl schemes
│   ├── regression/                # Corpus-diff transition-matrix harness (make regression-*)
│   ├── mine_constructs.py         # Construct-stratified candidate mining (F-0010)
│   ├── schwa/                     # Schwa classifier training + build_lexicon.py
│   ├── bengali/                   # Bengali trainers/evaluators (vowels, lexicon, selector, records)
│   ├── bengali_*.py / build_bengali.py / build_banglatlit.py / audit_bengali_*.py
│   └── build_freq.py / build_aksharantar_test.py / build_comilingua.py /
│       train_ngram.py / mine_overrides.py
├── datasets/                      # Raw-dataset download/generation scripts
├── docs/
│   ├── RESEARCH.md                # Problem, literature, datasets, methodology, results
│   ├── DESIGN.md                  # Architecture, rule system, scheme, learned components
│   ├── ROADMAP.md                 # Post-1.0 directions with tradeoffs
│   ├── PROCESS.md                 # Task tracking, PR discipline, accuracy reporting
│   ├── reviews/                   # Dated decision records (incl. negative results)
│   ├── archive/                   # Historical docs (2025-era, bannered)
│   └── reference/                 # External reference material (incl. candidate-datasets.md)
├── .claude/                       # Claude Code configuration + hooks
├── .github/workflows/             # ci.yml + release.yml (GoReleaser) + pages.yml (WASM demo) + release-npm.yml (OIDC npm publish)
├── npm/NOTICE.md                  # Code + embedded-data licenses (shipped in npm and release archives)
├── Makefile / TASKS.md / README.md / CLAUDE.md / CHANGELOG.md / LEARNINGS.md
```

`internal/legacy_lang` and `scripts/` were removed in the pre-1.0 reorg (PR
#54); git history preserves them.

## Current Status

**Shipped:** v1.2.1 (tagged; unreleased changes in `CHANGELOG.md`); the browser
(WASM) demo — live at <https://budhash.com/gomanize> (F-0006: `make wasm`,
auto-deployed by `pages.yml`); the npm package `@budhash/gomanize` — the WASM
engine + JS/TS wrapper (F-0009: `make npm`, published by `release-npm.yml`); and
the structured parser-QA program (F-0010, complete) that fixed the गई
independent-vowel and chandrabindu-nasal bugs and added construct-coverage
regression nets (`make test-constructs` / `test-heldout` / `test-combinatorial`
/ `test-akshara`). Remaining F-0006 work is T-0030 (consent-based correction
capture).

**Experimental Bengali** (F-0011; merged 2026-10-05/06, keel-reviewed
2026-10-07): rules plus opt-in vowel model, lexicon and native selector across
the Go API, CLI (`--language=bengali`), npm and web. Held-out match-any 56.52%
rules / 62.76% model / 63.04% model + selector. Lyrics references are unreviewed
pending maintainer attestation (T-0068). Keel follow-ups: F-0013. Records:
[`docs/reviews/2026-10-07-bengali-keel.md`](docs/reviews/2026-10-07-bengali-keel.md).

Full results: [`docs/RESEARCH.md`](docs/RESEARCH.md) §4. The Hindi CI regression
gate is strict top-1 pure ≥85% on curated Dakshina (`make test-dakshina`), plus
the frozen Hindi snapshot; Bengali is gated by the preregistered B1 dev gate and
frozen output pins; overrides
are an exception list, never headline numbers. Hindi's remaining failures are
lexical, not rule-governed — the evidence, including rejected rules, is in RESEARCH
§5. Bengali's are mostly convention- and rule-governed (measured plan:
[`docs/reviews/2026-10-10-bengali-quality-assessment.md`](docs/reviews/2026-10-10-bengali-quality-assessment.md)).

Contamination rules (train-split-only training; benchmarks never mined):
RESEARCH §2. Deliberate divergences from Dakshina/Hunterian conventions:
[`docs/DESIGN.md`](docs/DESIGN.md) §3.

## Ushuaia Comparison Tool

Compare output against [ushuaia.pl](https://www.ushuaia.pl/transliterate/)
schemes:

```bash
./tools/ushuaia "नमस्ते" --compare   # prints Hunterian vs gomanize side by side
```

Other schemes: `--all` and `--iso` (usage is in the script header).

## CI/CD

Pre-commit hooks run on every commit: block commits to main/master, format
(`make fmt`), lint (`make lint`), validate `TASKS.md`, and catch
whitespace/merge-conflict issues. Install via `make init`; run manually with
`pre-commit run --all-files`; skip sparingly with `git commit --no-verify`.

CI on push/PR to main: format check, lint, build, test with coverage, accuracy
benchmarks, the Bengali gates and record assertions, and `make npm-test`. Hosted CI
uses the runner's default Python (3.12 today); run Python suites locally under
3.12 too (`uv run --no-project --python 3.12 python -m unittest ...`).

Releases are automated by GoReleaser on version tags:

```bash
git tag vX.Y.Z && git push origin vX.Y.Z
```

This publishes binaries (Linux/macOS/Windows, amd64+arm64), checksums, and a
changelog; verify with `./gomanize --version`.

## Roadmap

Live backlog: `TASKS.md` via `./tools/tasks tree`. History and reasoning:
`docs/reviews/`. Post-1.0 directions with tradeoffs: [`docs/ROADMAP.md`](docs/ROADMAP.md)
(seeded as features F-0006–F-0009); design constraints in [`docs/DESIGN.md`](docs/DESIGN.md) §6.
