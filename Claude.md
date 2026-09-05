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
The current roadmap is seeded as F-0001…F-0004 (H0 tooling → H1 measurement →
H2 rules → H3 learned component); reasoning in
[`docs/reviews/2026-09-04-state-of-project-and-path-to-next-level.md`](docs/reviews/2026-09-04-state-of-project-and-path-to-next-level.md).

**Rules:** feature branches only (never commit to `main`); `make ci` before every
PR; report **pure** (no-override) accuracy as the headline; a shortcut is either
fixed in the same PR or tracked via `./tools/tasks new` with rationale.

## Architecture

```
gomanize/
├── gomanize.go                    # Public API (Romanizer interface)
├── cmd/main.go                    # CLI entry point
├── core/                          # Universal transliteration engine
│   ├── engine.go                  # Rule-based engine
│   ├── types.go                   # Core types (Unit, Word, Options)
│   └── rule.go                    # Rule definitions and phases
├── lang/hindi/                    # Hindi language implementation
│   ├── hindi.go                   # Hindi language definition
│   └── rules.go                   # Hindi-specific rules (schwa, consonant, vowel, render)
├── scheme/colloquial/             # Colloquial romanization scheme
│   └── colloquial.go              # Scheme definition
├── script/brahmic/                # Brahmic script support
│   └── brahmic.go                 # Devanagari parsing and utilities
├── benchmark/                     # Accuracy benchmarks
│   ├── benchmark_test.go          # Benchmark tests
│   └── data/                      # Test datasets
│       ├── curated_hi.csv         # Curated Hindi test data (1,335 entries)
│       ├── override_hi.csv        # Manual overrides
│       └── ignore_hi.csv          # Words to skip
├── internal/legacy_lang/          # Legacy implementation (for comparison)
├── scripts/
│   └── ushuaia                    # Compare with ushuaia.pl Hunterian
├── .claude/                       # Claude Code configuration
├── .github/workflows/
│   ├── ci.yml                     # CI pipeline
│   └── release.yml                # GoReleaser on tags
├── Makefile                       # Development workflow
└── CLAUDE.md                      # This file
```

## Current Status

### Test Results

Romanization is many-to-one (जनता = janata / janta / janataa are all valid), so
**match-any-attested-variant** is the honest headline; strict single-reference
under-counts correctness. See `make test-dakshina` and the multi-reference test.

| Metric | Passed | Total | Accuracy |
|--------|--------|-------|----------|
| **Match-any attested variant** (no overrides) | 1,239 | 1,335 | **92.8%** ✓ |
| Strict top-1, pure (single ref) | 1,150 | 1,335 | 86.1% |
| Strict top-1, with overrides | 1,202 | 1,335 | 90.1% |
| Mean minCER (over variants) | - | - | **0.0116** (human floor ≈ 0.054) |

Pure single-ref accuracy is the CI gate (floor 85%); overrides are an exception
lexicon, not engine skill, and are reported only as a secondary line.

### Remaining Failure Patterns

| Issue | Count | % of Failures | Notes |
|-------|-------|---------------|-------|
| OTHER | 70 | 44.3% | Long vowel variations (ee/oo) |
| MISSING_SCHWA | 58 | 36.7% | Medial schwa variations (phonetically correct) |
| EXTRA_SCHWA | 15 | 11.4% | Schwa preservation in compounds |

### Deliberate Divergences

Some romanization choices prioritize phonetic accuracy over matching Dakshina or Hunterian:

| Pattern | Dakshina/Hunterian | Gomanize | Rationale |
|---------|-------------------|----------|-----------|
| जनता | janata | janta | Phonetic: schwa deleted in CCV |
| कहते | kahate | kahte | Same: medial schwa deletion |
| गाना | gaana | gana | Current rule: ा→aa only in ा+C+END |
| मंत्र | mantr | mantra | Final 'a' for readability |

Use `--keep-medial-schwa` flag to get dataset-compatible output (janata, kahate, etc.)

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

### Dakshina Dataset
Human-attested romanizations from Google Research:
- 53K Hindi word pairs total
- Using 1,923 high-confidence pairs (4+ attestations)
- CC BY-SA 4.0 license

```bash
# Download and setup test data
make download-datasets
make setup-testdata
```

### Test Commands

```bash
make test              # Run all tests
make test-unit         # Unit tests only (fast)
make test-cover        # Tests with coverage
make test-dakshina     # Dakshina accuracy test
make test-analysis     # Failure breakdown
make test-original     # Original hindi-common.txt
make bench             # Benchmarks
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
| `gomanize.go` | Public API: `New()`, `Translit()`, `NewWithOptions()` |
| `core/engine.go` | Rule-based transliteration engine |
| `core/types.go` | Options struct (LongVowels, SimpleNasals, KeepMedialSchwa) |
| `lang/hindi/rules.go` | Hindi-specific rules (schwa, consonant, vowel, render) |
| `script/brahmic/brahmic.go` | Devanagari parsing and schwa state management |
| `benchmark/benchmark_test.go` | Accuracy benchmarks against Dakshina dataset |
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

### Phase 1: Core Accuracy ✓ (Complete)
- [x] Fix first syllable schwa deletion
- [x] Fix word-final schwa for Sanskrit words (र, य, व endings)
- [x] Add missing number ९ → 9
- [x] Add long vowel "aa" rule for ा+C+END
- [x] व → w for conjuncts only (स्व, श्व, द्व, ख्व)
- [x] Target: 80%+ accuracy ✓ (82.5%)

### Phase 2: Refinements ✓ (Complete)
- [x] Add `--long-vowels` flag for broader ा→aa (गाना→gaana)
- [x] Add `--simple-nasals` flag for simplified nasal endings (करें→karen)
- [x] Add `--keep-medial-schwa` flag for dataset-compatible output (जनता→janata)
- [x] Add compound word schwa deletion rule (देशभर→deshbhar)
- [x] Add CLI batch testing (`--test=FILE`, `--diff`)
- [x] Add rule inspection (`--list-rules`, `--disable-rule`)
- [x] Add फ→f surgical rule with फू→ph exception (film vs phool)
- [x] Target: 90%+ accuracy ✓ (86.1% pure, 90.1% with overrides)

### Phase 3: Enhancements
- [ ] Multiple transliteration schemes (IAST option)
- [ ] Bidirectional (Roman → Devanagari)
- [ ] Additional languages (Marathi, Nepali)

### Phase 4: Distribution
- [ ] Web API
- [ ] WASM build for browser
- [ ] npm package via wasm
