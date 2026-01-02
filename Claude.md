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
```

## Architecture

```
gomanize/
├── gomanize.go                    # Public API (Romanizer interface)
├── cmd/main.go                    # CLI entry point
├── internal/lang/
│   ├── hindi.go                   # Hindi transliterator (main implementation)
│   ├── hindi-orig.go              # Legacy implementation
│   ├── unit_test.go               # Unit tests (fast, targeted)
│   └── integration_test.go        # Integration tests (full datasets)
├── testbed/
│   ├── hindi-common.txt           # Original test data (1,036 pairs)
│   ├── dakshina/                  # Google Dakshina dataset
│   │   ├── native_hindi.tsv       # Native Hindi words (1,335 entries)
│   │   └── english_loanwords.tsv  # English loanwords (588 entries)
│   └── ISSUES.md                  # Documented failure patterns
├── scripts/
│   └── ushuaia                    # Compare with ushuaia.pl Hunterian
├── datasets/                      # Downloaded datasets (gitignored)
│   └── dakshina_dataset_v1.0/     # Full Dakshina dataset
├── .claude/                       # Claude Code configuration
│   ├── settings.json              # Plugins and hooks
│   ├── settings.local.json        # Local permissions
│   └── hooks/                     # Session and safety hooks
├── .github/workflows/
│   ├── ci.yml                     # CI pipeline
│   └── release.yml                # GoReleaser on tags
├── .golangci.yml                  # Linter configuration (v2)
├── .goreleaser.yml                # Release builds configuration
├── .pre-commit-config.yaml        # Pre-commit hooks
├── Makefile                       # Development workflow
└── Claude.md                      # This file
```

## Current Status

### Test Results

| Dataset | Passed | Total | Accuracy |
|---------|--------|-------|----------|
| Original (hindi-common.txt) | 613 | 1,036 | 59.2% |
| Dakshina (native Hindi) | 1,102 | 1,335 | **82.5%** |
| Target | - | - | **80%+** ✓ |

### Remaining Failure Patterns

| Issue | Count | % of Failures | Notes |
|-------|-------|---------------|-------|
| OTHER | 116 | 49.8% | Compound issues |
| MISSING_SCHWA | 66 | 28.3% | Medial schwa variations |
| EXTRA_SCHWA | 30 | 12.9% | Over-retention |
| V_VS_W | 15 | 6.4% | व mapping edge cases |
| MISSING_FINAL_A | 6 | 2.6% | Sanskrit endings |

### Key Differences from Hunterian

Based on comparison with [ushuaia.pl](https://www.ushuaia.pl/transliterate/) Hunterian:

| Word | Hunterian | Gomanize | Issue |
|------|-----------|----------|-------|
| देव | dew | dev | व → w vs v |
| ऐश्वर्या | aishwrya | aishvarya | व → w vs v |
| मंत्र | mantr | mantra | Final schwa |
| चंद्र | chandr | chandra | Final schwa |
| समझना | samjhana | samajhna | Medial schwa |

## Transliteration Standards

This project follows **colloquial/phonetic Hindi romanization** (not scholarly IAST):
- Aggressive schwa deletion to match spoken Hindi
- No diacritics (aa not ā, i not ī)
- Phonetic spelling (kh not ḵẖ)

### Research References
- [Hunterian Transliteration](https://en.wikipedia.org/wiki/Hunterian_transliteration) - Official India standard
- [Schwa Deletion Rules](https://en.wikipedia.org/wiki/Schwa_deletion_in_Indo-Aryan_languages)
- [Google Dakshina Dataset](https://github.com/google-research-datasets/dakshina) - Test data source

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
| `gomanize.go` | Public API: `New()`, `Translit()` |
| `internal/lang/hindi.go` | Main transliteration logic, symbol maps |
| `internal/lang/unit_test.go` | Unit tests for specific rules |
| `internal/lang/integration_test.go` | Full dataset tests |
| `testbed/ISSUES.md` | Detailed failure analysis |
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
- [x] Target: 80%+ accuracy ✓ (81.7%)

### Phase 2: Refinements (Current)
- [ ] Evaluate व → w mapping (Hunterian uses 'w')
- [ ] Fine-tune schwa deletion rules
- [ ] Multiple transliteration schemes (IAST option)

### Phase 3: Enhancements
- [ ] Bidirectional (Roman → Devanagari)
- [ ] Additional languages (Marathi, Nepali)

### Phase 4: Distribution
- [ ] Web API
- [ ] WASM build for browser
- [ ] npm package via wasm
