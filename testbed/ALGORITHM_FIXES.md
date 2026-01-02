# Algorithm Improvement Plan

## Current State ✓

| Metric | Value |
|--------|-------|
| Native Hindi Accuracy | **82.5%** (1,102/1,335) |
| Original Suite | 59.2% (613/1,036) |
| Target | 80%+ ✓ |

## Completed Fixes (Phase 1)

### Fix #1: V_VS_W ✓
- Changed default व mapping from 'w' to 'v'
- Added व→w for specific conjuncts only (स्व, श्व, द्व, ख्व)
- Examples: स्वागत→swagat, देव→dev, पर्वत→parvat

### Fix #2: MISSING_SCHWA ✓
- Protected first-syllable schwa from deletion
- Track first vowel matra to determine first syllable boundary
- Examples: प्रकाश→prakash, अध्यक्ष→adhyaksh

### Fix #3: MISSING_FINAL_A ✓
- Retain schwa for Sanskrit word-final consonant clusters
- Patterns: त्र, द्र, न्य, र्य, र्व
- Examples: मंत्र→mantra, चंद्र→chandra

### Fix #4: Long Vowel "aa" ✓
- ा outputs "aa" when followed by consonant at word end
- Examples: काम→kaam, इंसान→insaan

---

## Remaining Failure Patterns

| Issue | Count | % of Failures | Notes |
|-------|-------|---------------|-------|
| OTHER | 116 | 49.8% | Compound issues |
| MISSING_SCHWA | 66 | 28.3% | Medial schwa variations |
| EXTRA_SCHWA | 30 | 12.9% | Over-retention |
| V_VS_W | 15 | 6.4% | Edge cases (गुवाहाटी) |
| MISSING_FINAL_A | 6 | 2.6% | Sanskrit endings |

---

## Phase 2: To Investigate

### Broader ा→aa Rule
**Problem:** Words like गाना output "gana" but Dakshina expects "gaana"

**Current rule:** ा→aa only in ा+C+END pattern
- काम → kaam ✓
- इंसान → insaan ✓

**Proposed:** ा→aa in more positions
- गाना → gaana (currently: gana)
- बनाना → banaana (currently: banana)

**Impact:** Would affect ~45% of words - needs careful analysis

### Medial Schwa Fine-tuning
**Problem:** Complex medial schwa patterns

**Examples:**
- समझना → samajhna (expected: samjhana)
- आजकल → aajakal (expected: aajkal)

**Analysis needed:** Determine rules for VC+CV patterns

---

## Key Code Locations

| Component | File:Line | Description |
|-----------|-----------|-------------|
| व→w conjuncts | hindi.go:~320 | Check for स, श, द, ख before व |
| Long vowel aa | hindi.go:~350 | ा+C+END pattern |
| Schwa logic | hindi.go:281-307 | Main consonant handling |
| Final conjunct | hindi.go:~310 | Sanskrit word endings |

## Test Infrastructure

| Test | Command | Purpose |
|------|---------|---------|
| Unit tests | `make test-unit` | Fast, specific rules |
| Dakshina accuracy | `make test-dakshina` | Overall accuracy |
| Failure analysis | `make test-analysis` | Pattern breakdown |
| Full CI | `make ci` | All checks |
| Hunterian compare | `./scripts/ushuaia "word" --compare` | Compare with standard |

## Success Criteria

- [x] Dakshina native Hindi accuracy ≥ 80% (achieved: 82.5%)
- [x] All existing unit tests pass
- [x] No regression in original test suite
- [x] Known bug tests converted to passing tests
