# Algorithm Improvement Plan

Goal: Improve Dakshina accuracy from 48.8% to 80%+

## Current State

| Metric | Value |
|--------|-------|
| Native Hindi Accuracy | 48.8% (914/1,872) |
| Original Suite | 56.9% (590/1,036) |
| Target | 80%+ |

## Failure Breakdown

| Issue | Count | % of Failures | Priority | Expected Gain |
|-------|-------|---------------|----------|---------------|
| MISSING_SCHWA | 243 | 25.4% | HIGH | +21% |
| V_VS_W | 59 | 6.2% | MEDIUM | +6% |
| MISSING_FINAL_A | 31 | 3.2% | MEDIUM | +3% |
| EXTRA_SCHWA | 13 | 1.4% | LOW | +1% |
| OTHER | 612 | 63.9% | - | Auto-improve |

---

## Fix #1: V_VS_W (Easiest Win)

### Problem
व is mapped to 'w' but should be 'v' in colloquial Hindi.

**Examples:**
- `देव → dew` (expected: `dev`)
- `उत्सव → utsaw` (expected: `utsav`)
- `कवि → kawi` (expected: `kavi`)

### Root Cause
```go
// hindi.go:113
"व": {ctg: consonant, hun: "w"},  // Currently 'w'
```

Only word-initial position is overridden to 'v' (lines 285-288).

### Fix
1. Change line 113: `hun: "w"` → `hun: "v"`
2. Remove special case for word-initial व (lines 285-288)

### Expected Impact
- Fixes 59 failures immediately
- Accuracy: 48.8% → ~55%

---

## Fix #2: MISSING_SCHWA (Biggest Impact)

### Problem
First-syllable schwa is being deleted when it shouldn't be.

**Examples:**
- `प्रकाश → prkash` (expected: `prakash`)
- `अध्यक्ष → adhyksh` (expected: `adhyaksh`)
- `गर्भ → grbh` (expected: `garbh`)

### Root Cause
The schwa suppression logic (lines 281-307) doesn't protect first-syllable consonants.

```go
// hindi.go:296 - Current logic
if sb.index != 0 && nxtToNxtExists && nxtToNxtSi.isVowel() {
    converted = converted + rom  // Suppress schwa
}
```

When processing `प्रकाश`:
1. `प` → inherent 'a' → `pa` ✗ (should be `pra`)
2. `्र` → `r` (halant + r, no schwa)
3. Result: `prkash` instead of `prakash`

### Hindi Schwa Deletion Rules
1. **Never delete schwa of first syllable** (Critical rule being violated)
2. Delete schwa when consonant cluster + vowel follows
3. Retain schwa at word boundary or before vowels

### Fix Strategy
Track syllable position and protect first-syllable schwa:

```go
// Proposed: Track if we're still in first syllable
inFirstSyllable := true
for each character:
    if isVowelMatra || isIndependentVowel:
        inFirstSyllable = false  // First syllable ends after first vowel

    if isConsonant && shouldSuppressSchwa():
        if inFirstSyllable:
            addSchwa()  // Protect first syllable
        else:
            suppressSchwa()  // Apply normal rules
```

### Expected Impact
- Fixes ~243 failures
- Accuracy: 55% → ~76%

---

## Fix #3: MISSING_FINAL_A (Sanskrit Words)

### Problem
Word-final schwa deleted for Sanskrit-origin words that need it.

**Examples:**
- `अन्य → any` (expected: `anya`)
- `इंद्र → indr` (expected: `indra`)
- `मंत्र → mantr` (expected: `mantra`)
- `चंद्र → chandr` (expected: `chandra`)

### Pattern
Sanskrit words ending in consonant clusters retain final 'a':
- `त्र` (tra): chandra, mantra
- `द्र` (dra): indra
- `न्य` (nya): anya
- `र्य` (rya): karya

### Fix Strategy
Detect word-final consonant clusters and conditionally retain schwa:

```go
// At word end, if previous char was halant (्):
//   → This is a consonant cluster ending
//   → Retain final 'a'
// Else (single consonant):
//   → Delete final 'a' (current behavior)
```

### Expected Impact
- Fixes ~31 failures
- Accuracy: 76% → ~79%

---

## Fix #4: EXTRA_SCHWA

### Problem
Extra schwa added in VC+CV patterns.

**Examples:**
- `अमृतसर → amritasar` (expected: `amritsar`)
- `आजकल → aajakal` (expected: `aajkal`)

### Root Cause
Schwa suppression not triggering correctly in vowel-consonant + consonant-vowel sequences.

### Fix Strategy
Refine consonant cluster detection:
```go
// If current consonant followed by:
//   - Another consonant (cluster) + vowel → suppress schwa
// Ensure lookahead handles matras correctly
```

### Expected Impact
- Fixes ~13 failures
- Accuracy: 79% → 80%+

---

## Implementation Order

| Phase | Fix | Complexity | Files | Accuracy |
|-------|-----|------------|-------|----------|
| 1 | V_VS_W | Low | hindi.go:113, 285-288 | 48.8% → 55% |
| 2 | MISSING_SCHWA | Medium | hindi.go:281-307 | 55% → 76% |
| 3 | MISSING_FINAL_A | Medium | hindi.go:296-307 | 76% → 79% |
| 4 | EXTRA_SCHWA | Low | hindi.go:281-295 | 79% → 80%+ |

---

## Key Code Locations

| Component | File:Line | Description |
|-----------|-----------|-------------|
| व mapping | hindi.go:113 | Change 'w' to 'v' |
| Word-initial व | hindi.go:285-288 | Remove after Fix #1 |
| Schwa logic | hindi.go:281-307 | Main consonant handling |
| Lookahead | hindi.go:220-250 | PeekN() mechanism |

## Test Infrastructure

| Test | Command | Purpose |
|------|---------|---------|
| Unit tests | `make test-unit` | Fast, specific rules |
| Dakshina accuracy | `make test-dakshina` | Overall accuracy |
| Failure analysis | `make test-analysis` | Pattern breakdown |
| Full CI | `make ci` | All checks |

## Success Criteria

- [ ] Dakshina native Hindi accuracy ≥ 80%
- [ ] All existing unit tests pass
- [ ] No regression in original test suite
- [ ] Known bug tests converted to passing tests
