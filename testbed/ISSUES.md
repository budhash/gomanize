# Gomanize - Issue Analysis

## Baseline (Dakshina Dataset)

- **Total tested**: 1,881 native Hindi words (excluding obvious English loanwords)
- **Passed**: 914 (48.6%)
- **Failed**: 967 (51.4%)

## Failure Categories (Priority Order)

### 1. OTHER - 591 failures (61.1%)

Complex issues that don't fit simple patterns. Examples:
- `अगरतला → agaratla` (expected: agartala) - schwa placement
- `अनिवार्य → aniwary` (expected: anivarya) - v/w + schwa
- `अर्थव्यवस्था → arthwyawstha` (expected: arthvyavastha) - multiple issues

**Root cause**: Combination of v/w issues + schwa placement + other factors.

---

### 2. MISSING_SCHWA - 243 failures (25.1%)

Schwa is missing where it should be present. Examples:
- `अंजना → anjna` (expected: anjana)
- `अगले → agle` (expected: agale)
- `अध्यक्ष → adhyksh` (expected: adhyaksh)
- `अप्रकाशित → aprkashit` (expected: aprakashit)

**Root cause**: First syllable schwa being deleted incorrectly.

**Rule violated**: "Schwa of the first syllable cannot be deleted"

**Fix priority**: HIGH - affects 25% of failures

---

### 3. V_VS_W - 59 failures (6.1%)

व is being transliterated as 'w' when 'v' is expected. Examples:
- `उत्सव → utsaw` (expected: utsav)
- `कवि → kawi` (expected: kavi)
- `देव → dew` (expected: dev)
- `पूर्व → purw` (expected: purv)

**Root cause**: Current code maps व→w always (except word-initial).

**Expected behavior**: व→v in most positions (especially before vowels and at word end).

**Fix priority**: MEDIUM - affects 6% of failures but is a simple mapping change

---

### 4. MISSING_FINAL_A - 31 failures (3.2%)

Word-final schwa is missing for Sanskrit-origin words. Examples:
- `अन्य → any` (expected: anya)
- `इंद्र → indr` (expected: indra)
- `कार्य → kary` (expected: karya)
- `चंद्र → chandr` (expected: chandra)
- `मंत्र → mantr` (expected: mantra)

**Root cause**: Word-final schwa being deleted for words that should retain it.

**Pattern**: Words ending in consonant clusters (त्र, द्र, न्य, र्य) often retain final 'a'.

**Fix priority**: MEDIUM - affects 3% but important for proper names

---

### 5. ENGLISH_LOANWORD - 27 failures (2.8%)

Hindi words that are transliterations of English use English spelling. Examples:
- `आईडी → aaidi` (expected: id)
- `आईपीएल → aaipiel` (expected: ipl)
- `इंटर → intar` (expected: inter)

**Root cause**: These are English words written in Devanagari - users expect English spelling.

**Fix priority**: LOW - out of scope for Hindi transliteration (could be a separate feature)

---

### 6. EXTRA_SCHWA - 13 failures (1.3%)

Extra schwa is being added where it shouldn't be. Examples:
- `अपमानजनक → apmanajanak` (expected: apmanjanak)
- `अमृतसर → amritasar` (expected: amritsar)
- `आजकल → aajakal` (expected: aajkal)

**Root cause**: VC_CV schwa deletion rule not being applied correctly.

**Fix priority**: MEDIUM

---

### 7. AA_VS_A_START - 3 failures (0.3%)

Words starting with आ where expected output uses single 'a'. Examples:
- `आयुर्वेद → aayurwed` (expected: ayurved)

**Root cause**: Dataset inconsistency (आ should be 'aa' per Hunterian).

**Fix priority**: SKIP - our output may be more correct

---

## Recommended Fix Order

1. **MISSING_SCHWA** (25.1%) - Fix first syllable schwa rule
2. **V_VS_W** (6.1%) - Change व mapping to 'v' (except specific contexts)
3. **MISSING_FINAL_A** (3.2%) - Handle Sanskrit word endings
4. **EXTRA_SCHWA** (1.3%) - Improve VC_CV rule
5. **OTHER** (61.1%) - Will reduce as other fixes are applied

## Expected Improvement

If we fix issues 1-4, we expect:
- Current: 48.6% accuracy
- After MISSING_SCHWA fix: ~70% accuracy
- After V_VS_W fix: ~76% accuracy
- After MISSING_FINAL_A + EXTRA_SCHWA: ~80% accuracy

Target: **80%+ accuracy** on native Hindi words
