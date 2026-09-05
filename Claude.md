# Gomanize

A Go library and CLI tool for transliterating Devanagari script (Hindi) to Latin/Roman characters. Designed for practical use cases like song lyrics romanization.

## Project Overview

**Purpose**: Romanize Hindi text from Devanagari script into readable Latin characters for singing along, international audiences, or systems without Devanagari support.

**Author**: Budhaditya (budhash@gmail.com)
**License**: MIT (2023-2025)
**Go Version**: 1.21+

## Quick Start

```bash
# First-time setup
make init

# Build
make build

# Run
./gomanize "नमस्ते भारत"
# Output: namste bharat

echo "हिंदी गाना" | ./gomanize
# Output: hindi gana

# Options
./gomanize --long-vowels "गाना"     # gaana (aa for all ā)
./gomanize --simple-nasals "करें"   # karen (simplified nasals)
./gomanize --keep-medial-schwa "जनता" # janata (dataset-compatible)
./gomanize --schwa-model "जनता"     # learned schwa classifier (see docs/reviews H3)
./gomanize --lexicon "अंकल"          # uncle (known words → attested spelling, rules for OOV)
./gomanize --rerank "जनता"          # best of rules/schwa-model candidates via char LM

# Run CI (format, lint, build, test)
make ci

# See all commands
make help
```

## Development Workflow

```bash
# Setup
make init           # First-time setup (deps + pre-commit hooks)
make hooks          # Install pre-commit hooks
make hooks-update   # Update pre-commit hook versions

# Development
make help           # Show all commands
make dev            # Full workflow: format, vet, test
make ci             # CI pipeline: fmt-check, lint, build, test

# Testing
make test           # Run all tests
make test-unit      # Run unit tests only (fast)
make test-cover     # Run tests with coverage
make test-dakshina  # Run Dakshina accuracy tests
make test-analysis  # Show failure pattern breakdown

# Code Quality
make fmt            # Format code
make fmt-check      # Check formatting (CI)
make lint           # Run golangci-lint
make lint-fix       # Run linter with auto-fix

# Utilities
make demo           # Demo with sample words
make status         # Quick status check
make bench          # Run benchmarks
make tasks ARGS=".." # Task tracker passthrough (see below)
```

## Task Tracking & Process

`TASKS.md` is the single source of truth for features and tasks, managed by a
vendored zero-dependency CLI (`tools/tasks.py`, canonical:
<https://github.com/budhash/tasks>). **Edit it only through the CLI** — never by
hand. Full process: [`docs/PROCESS.md`](docs/PROCESS.md).

```bash
./tools/tasks tree            # Everything, grouped by feature
./tools/tasks next            # Next actionable task (respects deps)
./tools/tasks show F-2 --full # Feature + design notes
./tools/tasks start T-10      # → doing (single WIP)
./tools/tasks done  T-10      # → done (auto-stamps @done)
./tools/tasks validate        # Sanity-check (run in CI)
```

Schema: `- [ ] (ID) [PRIO] [STATUS] Title @tags...` — features `F-####`, tasks
`T-####`, priorities `P0..P3`, status `[todo] [doing] [done] [deferred] [skipped]`.
Features F-0001…F-0005 (tooling → measurement → rules → learned components →
real-world validation) are all **done** as of 2026-09-05; the founding review is
[`docs/reviews/2026-09-04-state-of-project-and-path-to-next-level.md`](docs/reviews/2026-09-04-state-of-project-and-path-to-next-level.md)
and each subsequent result (including negatives) has a dated decision record in
`docs/reviews/`.

**Rules:** feature branches only (never commit to `main`); `make ci` before every
PR; report **pure** (no-override) accuracy as the headline; a shortcut is either
fixed in the same PR or tracked via `./tools/tasks new` with rationale.

## Architecture

```
gomanize/
├── gomanize.go                    # Public API (New, Translit — whitespace/punct-aware)
├── cmd/main.go                    # CLI entry point (all flags incl. --schwa-model/--lexicon/--rerank)
├── core/                          # Universal transliteration engine (no script knowledge)
│   ├── engine.go                  # Pipeline + LexiconProvider/Reranker optional interfaces
│   ├── types.go                   # Core types (Unit, Word, Options)
│   └── rule.go                    # Rule definitions, phases, scopes, modes
├── lang/hindi/                    # Hindi language implementation
│   ├── symbols.go / rules.go      # Symbol map + Hindi-specific rules
│   ├── schwa_model.go + schwa_tree.json    # Learned schwa classifier (embedded, 34KB)
│   ├── lexicon.go + lexicon.tsv            # 8.4k-word attested lexicon (embedded, 260KB)
│   └── reranker.go + roman_ngrams.tsv      # Char 4-gram re-ranker (embedded, 224KB)
├── scheme/colloquial/             # Colloquial romanization scheme
├── script/brahmic/                # Brahmic script support (shared by future languages)
│   ├── brahmic.go / parser.go / renderer.go / runs.go
│   └── schwa_rules.go             # Shared Brahmic schwa rules (brahmic.SchwaRules())
├── benchmark/                     # Accuracy benchmarks (5 evaluation suites)
│   ├── benchmark_test.go          # All benchmark tests
│   ├── metrics_test.go            # CER / minCER / match-any / reference loaders
│   └── data/
│       ├── curated_hi.csv         # Curated Dakshina subset (1,335 entries)
│       ├── dakshina_hi.csv        # Full Dakshina lexicon w/ splits + attestations
│       ├── aksharantar_test_hi.csv# Aksharantar human test set (10,112 pairs, CC-BY)
│       ├── comilingua_hi.csv      # COMI-LINGUA colloquial word pairs (CC-BY)
│       ├── freq_hi.csv            # Shabd top-15k frequency list (CC0)
│       ├── lyrics_gold_hi.csv     # Public-domain lyrics gold seed (43 lines)
│       └── override_hi.csv / ignore_hi.csv / aksharantar_hi.csv
├── tools/                         # Dev tooling (vendored task tracker + data pipelines)
│   ├── tasks, tasks.py            # Task tracker CLI over TASKS.md
│   ├── schwa/                     # Schwa classifier training (align/features/train)
│   ├── build_lexicon.py / build_freq.py / build_aksharantar_test.py
│   ├── build_comilingua.py / train_ngram.py / mine_overrides.py
├── internal/legacy_lang/          # Legacy implementation (for comparison)
├── scripts/ushuaia                # Compare with ushuaia.pl Hunterian
├── docs/
│   ├── PROCESS.md                 # Development process (tasks, PRs, accuracy discipline)
│   └── reviews/                   # Decision records (dated; every result incl. negatives)
├── .claude/                       # Claude Code configuration + hooks
├── .github/workflows/             # ci.yml + release.yml (GoReleaser on tags)
├── Makefile                       # Development workflow
├── TASKS.md                       # Live task tracker (edit ONLY via ./tools/tasks)
└── CLAUDE.md                      # This file
```

## Current Status

### Test Results

Romanization is many-to-one (जनता = janata / janta / janataa are all valid), so
**match-any-attested-variant** is the honest headline; strict single-reference
under-counts correctness. See `make test-dakshina` and the multi-reference test.

**Curated Dakshina (1,335 words, 2026-09-05):**

| Metric | Accuracy |
|--------|----------|
| **Match-any + `--rerank`** | **94.7%** |
| Match-any attested variant (default rules, no overrides) | 92.8% |
| Strict top-1, pure (single ref) — the CI gate, floor 85% | 86.1% |
| Mean minCER (over variants) | **0.0116** (human floor ≈ 0.054) |

**Real-world benchmarks (all independent of the curated set):**

| Benchmark | Default rules | Best config |
|-----------|--------------|-------------|
| Held-out Dakshina test (2,500 unseen words) | 69.0% | 70.4% (`--rerank`) |
| COMI-LINGUA naturally-typed Hindi (token-weighted) | 78.9% | 85.7% (`--lexicon`) |
| Frequency-weighted (Shabd top-15k ∩ gold) | 82.8% | 97.4% (`--lexicon`) |
| Lyrics gold seed (line CER, lower=better) | 0.0492 | 0.0465 (`--rerank`) |
| Aksharantar test (per-slice; NE slices) | 15–69% | lexicon +3.5 to +7.9 pts |

Overrides (`override_hi.csv`) are an exception list, not engine skill — reported
only as a secondary line and excluded from headline numbers.

### Where the remaining errors live

The rule ceiling is **reached and proven** (see `docs/reviews/`): remaining
failures are lexical, not rule-governed — vowel-length spelling conventions
(ee/oo vs i/u is a ~50/50 attested split), medial-schwa variance between
annotators, and loanwords/named entities. The path past them is data (lexicon
growth with human review), not more rules. Both candidate vowel rules were
implemented, measured net-negative, and rejected on evidence
(`docs/reviews/2026-09-04-h2-vowel-length-experiments.md`).

### Deliberate Divergences

Some romanization choices prioritize phonetic accuracy over matching Dakshina or Hunterian:

| Pattern | Dakshina/Hunterian | Gomanize | Rationale |
|---------|-------------------|----------|-----------|
| जनता | janata | janta | Phonetic: schwa deleted in CCV (janta IS also attested) |
| कहते | kahate | kahte | Same: medial schwa deletion |
| गाना | gaana | gana | ा→aa only in closed final syllable; broader rule measured net-negative |
| मंत्र | mantr | mantra | Final 'a' for readability |

Flags: `--keep-medial-schwa` (janata-style), `--schwa-model` (learned classifier),
`--lexicon` (attested spellings for 8.4k known words), `--rerank` (char-LM picks
best of rules/schwa-model candidates — improves every benchmark).

## Transliteration Standards

This project follows **colloquial/phonetic Hindi romanization** (not scholarly IAST):
- Aggressive schwa deletion to match spoken Hindi
- No diacritics (aa not ā, i not ī)
- Phonetic spelling (kh not ḵẖ)

### Research References
- [Hunterian Transliteration](https://en.wikipedia.org/wiki/Hunterian_transliteration) - Official India standard
- [Schwa Deletion Rules](https://en.wikipedia.org/wiki/Schwa_deletion_in_Indo-Aryan_languages)
- [Google Dakshina Dataset](https://github.com/google-research-datasets/dakshina) - Test data source
- [Aksharantar Dataset](https://huggingface.co/datasets/ai4bharat/Aksharantar) - Largest Indic transliteration dataset (26M pairs)
- [IndicXlit Paper](https://arxiv.org/abs/2205.03018) - "Aksharantar: Towards Building Open Transliteration Tools for the Next Billion Users"
- [IndicXlit Model](https://github.com/AI4Bharat/IndicXlit) - Neural transliteration model (11M params, 21 languages)
- [AI4Bharat Tools](https://ai4bharat.iitm.ac.in/tools) - Online transliteration demo and other NLP tools
- [LDCIL](https://ldcil.org/) - Linguistic Data Consortium for Indian Languages
- [LDCIL Data Portal](https://data.ldcil.org/) - Indian language datasets and corpora
- [Anuvadika](https://anuvadika.ciil.org/index.php) - CIIL transliteration tool

## Test Data

Five evaluation suites, all committed (licenses documented per file / in
`docs/reviews/`):

| Dataset | Size | License | Role |
|---------|------|---------|------|
| Dakshina (Google, frozen/archived 2026) | 53K rows, splits + attestation counts | CC BY-SA 4.0 | Curated benchmark + held-out test + training data for learned components |
| Aksharantar test (AI4Bharat, human-annotated 2022) | 10,112 pairs, 4 slices | CC-BY 4.0 | Independent human benchmark; named-entity slices |
| COMI-LINGUA word pairs (extracted from MT split) | 9,606 words / 152K tokens | CC-BY 4.0 | Naturally-typed colloquial Hindi (closest proxy to lyrics) |
| Shabd frequency list (top-15k Devanagari) | 15,000 words | CC0 | Frequency weighting + lexicon ranking |
| Lyrics gold seed (Kabir/Meera/Tagore etc.) | 43 PD lines, maintainer-attested | Public domain | Line-level flagship-use-case eval (CER floor 0.15 gate) |

**Contamination discipline:** learned components (schwa tree, lexicon, n-gram LM)
train on the Dakshina TRAIN split only; dev/test natives are disjoint and never
enter the lexicon. COMI-LINGUA is a benchmark, never a training/mining source.

```bash
# Bulk raw datasets (optional, for regeneration)
make download-datasets
make setup-testdata
# Data pipelines: tools/build_*.py, tools/train_ngram.py, tools/schwa/train.py
```

### Test Commands

```bash
make test              # Run all tests
make test-unit         # Unit tests only (fast; all packages except benchmark)
make test-cover        # Tests with coverage
make test-dakshina     # Curated Dakshina accuracy (pure + overrides + CER)
make test-integration  # Full Dakshina + Aksharantar bulk runs
make test-analysis     # Failure breakdown
make bench             # Performance benchmarks
# Individual suites: go test ./benchmark/... -run TestBenchmark<Name> -v
#   MultiReference | SchwaModelHeldout | LexiconCoverage | FrequencyWeighted |
#   AksharantarTestSet | ComiLingua | LyricsGold
```

### Ushuaia Comparison Tool

Compare gomanize output against [ushuaia.pl](https://www.ushuaia.pl/transliterate/) Hunterian transliteration:

```bash
# Compare single word
./scripts/ushuaia "नमस्ते" --compare
# Input:     नमस्ते
# Hunterian: namste
# Gomanize:  namaste
# Status:    ✗ Different

# Show all schemes (Hunterian, ISO-15919, Polish)
./scripts/ushuaia "ऐश्वर्या" --all
# Input:     ऐश्वर्या
# Hunterian: aishwrya
# ISO-15919: aiśvaryā
# Polish:    ajśwrja

# Get Hunterian only
./scripts/ushuaia "काम"
# kām

# Get ISO-15919 only
./scripts/ushuaia "काम" --iso
# kāma
```

Available language codes (for reference):
- `devanagari_hunt_transcribe` - Hunterian transcription
- `devanagari_iso_transliterate` - ISO-15919 transliteration
- `devanagari_hindi_pl_transcribe` - Polish transcription
- `devanagari_iast_transliterate` - Sanskrit IAST

## Key Files

| File | Purpose |
|------|---------|
| `gomanize.go` | Public API: `New()`, `Translit()` (whitespace/punct-aware), `NewWithOptions()` |
| `core/engine.go` | Pipeline + `LexiconProvider`/`Reranker` optional interfaces |
| `core/types.go` | Options (LongVowels, SimpleNasals, KeepMedialSchwa, SchwaModel, Lexicon, Rerank) |
| `lang/hindi/rules.go` | Hindi-specific rules; composes `brahmic.SchwaRules()` |
| `lang/hindi/schwa_model.go` | Learned schwa classifier (embedded decision tree, 90.7% held-out per-schwa) |
| `lang/hindi/lexicon.go` | 8,367-word attested lexicon (Dakshina-train only), rules as OOV fallback |
| `lang/hindi/reranker.go` | Char 4-gram LM re-ranker over {rules, schwa-model} candidates |
| `script/brahmic/schwa_rules.go` | Shared Brahmic schwa rules (reused by future languages) |
| `script/brahmic/brahmic.go` | Devanagari parsing and schwa state management |
| `benchmark/benchmark_test.go` | All 5 evaluation suites |
| `benchmark/metrics_test.go` | CER / minCER / match-any metric helpers |
| `tools/tasks` | Task tracker CLI (the ONLY way to edit TASKS.md) |
| `docs/reviews/` | Dated decision records — every result, including negatives |
| `Makefile` | Development commands |

## API Usage

```go
import gomanize "github.com/budhash/gomanize"

// Create transliterator
g, err := gomanize.New("hindi")
if err != nil {
    panic(err)
}

// Transliterate
output := g.Translit("नमस्ते दुनिया")
fmt.Println(output)  // "namste duniya"
```

## CI/CD

### Pre-commit Hooks
Pre-commit hooks run automatically on `git commit`:
- Prevent commits to main/master
- Format code (`make fmt`)
- Lint code (`make lint`)
- Check for common issues (trailing whitespace, merge conflicts, etc.)

```bash
# Install hooks (done by make init)
make hooks

# Run manually
pre-commit run --all-files

# Skip hooks (use sparingly)
git commit --no-verify
```

### GitHub Actions
CI runs on push/PR to main:
- Format check
- Lint (golangci-lint)
- Build
- Test with coverage

### Releases
Releases are automated via GoReleaser on version tags:

```bash
# Create a release
git tag v1.0.0
git push origin v1.0.0
```

This creates a GitHub Release with:
- Binaries for Linux, macOS, Windows (amd64 + arm64)
- Checksums
- Auto-generated changelog

CLI supports version flag:
```bash
./gomanize --version
# gomanize v1.0.0 (commit: abc1234, built: 2024-01-01)
```

## Roadmap

**Live backlog lives in `TASKS.md`** (view via `./tools/tasks tree`). History and
reasoning live in `docs/reviews/`. Status as of 2026-09-05:

### Completed (2025 — Phases 1–2: rule engine + accuracy)
Core rule engine, schwa deletion, CLI flags (`--long-vowels`, `--simple-nasals`,
`--keep-medial-schwa`), rule inspection, batch testing. Reached the rule-based
ceiling (86.1% pure).

### Completed (2026-09 — measurement, learned components, real-world validation)
- **Honest evaluation**: multi-reference match-any + CER/minCER; CI gates pure ≥85%
- **Rule ceiling proven**: candidate vowel rules measured net-negative, rejected
- **Learned components** (all embedded, zero runtime deps, opt-in):
  `--schwa-model` (decision tree, 90.7% held-out per-schwa), `--lexicon`
  (8,367 attested words, 71% token coverage), `--rerank` (char-LM candidate
  arbitration — improves every benchmark; curated match-any 94.7%)
- **Five benchmark suites** incl. naturally-typed Hindi (COMI-LINGUA) and a
  public-domain lyrics gold seed (line CER at the human consistency floor)
- **Architecture hardened**: source-char rule matching; shared Brahmic schwa
  rules extracted for multi-language reuse; whitespace/punctuation segmentation
- **Process**: task tracker, PR template, decision records, honest CI gates

### Open (candidate future work — not currently scheduled)
- Additional languages (Marathi, Nepali) — `brahmic.SchwaRules()` makes this
  tractable; needs per-language symbol maps + language rules
- Multiple schemes (IAST) — needs per-scheme symbol maps (interface change)
- Lexicon growth past the 78.2% train-gold ceiling — needs human review of mined
  candidates (`tools/mine_overrides.py`; unreviewed precision is only ~43%)
- Expanded lyrics gold set (more PD verse; LyricsTranslate curation)
- Bidirectional (Roman → Devanagari) — effectively a separate engine
- Distribution: Web API, WASM build, npm package
