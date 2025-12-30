# Gomanize

A Go library and CLI tool for transliterating Devanagari script (Hindi) to Latin/Roman characters. Designed for practical use cases like song lyrics romanization.

## Project Overview

**Purpose**: Romanize Hindi text from Devanagari script into readable Latin characters for singing along, international audiences, or systems without Devanagari support.

**Author**: Budhaditya (budhash@gmail.com)
**License**: MIT (2023)
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
│   │   └── all_high_conf.tsv      # High-confidence pairs (1,923 entries)
│   └── ISSUES.md                  # Documented failure patterns
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

### Test Results (Baseline)

| Dataset | Passed | Total | Accuracy |
|---------|--------|-------|----------|
| Original (hindi-common.txt) | 590 | 1,036 | 56.9% |
| Dakshina (native Hindi) | 914 | 1,872 | **48.8%** |
| Current Threshold | - | - | 45% |
| Target | - | - | **80%+** |

### Failure Pattern Analysis

| Issue | Count | % of Failures | Priority |
|-------|-------|---------------|----------|
| MISSING_SCHWA | 243 | 25.4% | **HIGH** |
| V_VS_W | 59 | 6.2% | MEDIUM |
| MISSING_FINAL_A | 31 | 3.2% | MEDIUM |
| EXTRA_SCHWA | 13 | 1.4% | MEDIUM |
| OTHER | 612 | 63.9% | (compound issues) |

### Key Issues to Fix

1. **First Syllable Schwa** (HIGH) - Never delete schwa in first syllable
   - `प्रकाश → prkash` should be `prakash`
   - `अध्यक्ष → adhyksh` should be `adhyaksh`

2. **व Mapping** (MEDIUM) - Should be 'v' in most cases, not 'w'
   - `देव → dew` should be `dev`
   - `उत्सव → utsaw` should be `utsav`

3. **Word-final Schwa** (MEDIUM) - Retain for Sanskrit-origin words
   - `चंद्र → chandr` should be `chandra`
   - `मंत्र → mantr` should be `mantra`

4. **Number Mapping** - ९ (9) missing from character map

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

### Phase 1: Fix Core Issues (Current)
- [ ] Fix first syllable schwa deletion
- [ ] Fix व → v mapping
- [ ] Fix word-final schwa for Sanskrit words
- [x] Add missing number ९ → 9
- [ ] Target: 80%+ accuracy

### Phase 2: Enhancements
- [ ] Multiple transliteration schemes (IAST option)
- [ ] Bidirectional (Roman → Devanagari)
- [ ] Additional languages (Marathi, Nepali)

### Phase 3: Distribution
- [ ] Web API
- [ ] WASM build for browser
- [ ] npm package via wasm
