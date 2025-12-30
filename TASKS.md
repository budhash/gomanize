# Gomanize Tasks

## Current Sprint: Accuracy Improvement

Goal: Improve Dakshina accuracy from 48.8% to 80%+

### In Progress

### Pending

- [ ] **Fix V_VS_W**: Change व from 'w' to 'v'
  - File: `internal/lang/hindi.go:113`
  - Change `hun: "w"` to `hun: "v"`
  - Remove word-initial special case (lines 285-288)
  - Expected: 48.8% → 55%

- [ ] **Fix MISSING_SCHWA**: Protect first-syllable schwa
  - File: `internal/lang/hindi.go:281-307`
  - Track syllable position
  - Never delete schwa in first syllable
  - Expected: 55% → 76%

- [ ] **Fix MISSING_FINAL_A**: Retain schwa for Sanskrit word endings
  - File: `internal/lang/hindi.go:296-307`
  - Detect consonant cluster endings (त्र, द्र, न्य, र्य)
  - Retain final 'a' for these patterns
  - Expected: 76% → 79%

- [ ] **Fix EXTRA_SCHWA**: Improve VC+CV schwa suppression
  - File: `internal/lang/hindi.go:281-295`
  - Refine consonant cluster detection
  - Expected: 79% → 80%+

### Completed

- [x] Initial project setup
- [x] CI/CD infrastructure (GitHub Actions, GoReleaser)
- [x] Pre-commit hooks
- [x] golangci-lint v2 configuration
- [x] v0.1.0 release
- [x] Algorithm improvement plan (ALGORITHM_FIXES.md)

## Backlog

- [ ] Add branch protection rules
- [ ] Dependabot configuration
- [ ] Multiple transliteration schemes (IAST option)
- [ ] Bidirectional transliteration (Roman → Devanagari)
- [ ] Additional languages (Marathi, Nepali)

## Commands

```bash
make test-unit      # Fast unit tests
make test-dakshina  # Accuracy test
make test-analysis  # Failure breakdown
make ci             # Full CI pipeline
```
