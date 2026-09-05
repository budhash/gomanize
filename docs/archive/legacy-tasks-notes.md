> **HISTORICAL DOCUMENT (pre-tracker free-form task notes, archived when the ./tools/tasks tracker was adopted).** Statistics and plans below reflect the
> project as of 2025 and are NOT current. For today's status see `CLAUDE.md`;
> for the live backlog see `/TASKS.md` (via `./tools/tasks tree`); for every
> subsequent result and decision see `docs/reviews/`.

# Gomanize Tasks

## Current Status

**Accuracy: 82.5%** (Target: 80%+ ✓)

## Phase 2: Refinements (Current)

### To Investigate

- [ ] **Broader ा→aa rule**: गाना→gaana pattern
  - Currently: ा→aa only in ा+C+END (काम→kaam)
  - Proposed: ा→aa in more positions (गाना→gaana, बनाना→banaana)
  - Impact: Would affect ~45% of words - needs careful analysis

- [ ] **Medial schwa fine-tuning**
  - Current failures: समझना→samajhna (expected: samjhana)
  - Complex rules needed for consonant clusters

- [ ] **Multiple transliteration schemes**
  - Add IAST option for scholarly use
  - Reference: docs/reference/Hindi-Marathi-Nepali-Transliteration.pdf

### Remaining Failure Patterns

| Issue | Count | % of Failures | Notes |
|-------|-------|---------------|-------|
| OTHER | 116 | 49.8% | Compound issues |
| MISSING_SCHWA | 66 | 28.3% | Medial schwa variations |
| EXTRA_SCHWA | 30 | 12.9% | Over-retention |
| V_VS_W | 15 | 6.4% | Edge cases (e.g., गुवाहाटी) |
| MISSING_FINAL_A | 6 | 2.6% | Sanskrit endings |

## Completed (Phase 1)

- [x] Fix first syllable schwa deletion (प्रकाश→prakash)
- [x] Fix word-final schwa for Sanskrit words (मंत्र→mantra)
- [x] Add missing number ९ → 9
- [x] Add long vowel "aa" rule for ा+C+END (काम→kaam)
- [x] व→w for conjuncts only (स्व→sw, श्व→shw, द्व→dw, ख्व→khw)
- [x] Ushuaia comparison tool (scripts/ushuaia)
- [x] Reference documentation (docs/reference/)

## Backlog (Phase 3+)

- [ ] Bidirectional transliteration (Roman → Devanagari)
- [ ] Additional languages (Marathi, Nepali)
- [ ] Web API
- [ ] WASM build for browser
- [ ] npm package via wasm

## Commands

```bash
make test-unit      # Fast unit tests
make test-dakshina  # Accuracy test
make test-analysis  # Failure breakdown
make ci             # Full CI pipeline
./scripts/ushuaia "word" --compare  # Compare with Hunterian
```
