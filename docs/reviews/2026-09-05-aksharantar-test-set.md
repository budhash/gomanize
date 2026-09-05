# T-0007 — Aksharantar human-annotated test set

**Date:** 2026-09-05
**Task:** F-0002 / T-0007 (deferred until network was available). Adds the field's
newer human benchmark alongside the frozen Dakshina.

## What was added

- `benchmark/data/aksharantar_test_hi.csv` — the **Aksharantar Hindi test set**:
  10,112 human-annotated pairs (native-speaker annotators via the Karya platform,
  2022), 6,424 unique natives, 2,614 with multiple attested variants.
  License **CC-BY 4.0** (committed with attribution; Madhani et al., Findings of
  EMNLP 2023, arXiv:2205.03018). Rebuild: `tools/build_aksharantar_test.py`.
- `TestBenchmarkAksharantarTestSet` — match-any + minCER per slice, rules vs
  rules+lexicon, with a lexicon-never-hurts assertion.

## Results (match-any over attested variants)

| Slice | Words | Rules | +Lexicon | Note |
|---|---|---|---|---|
| AK-Freq (frequent words) | 2,000 | 42.6% | 42.6% | zero lexicon overlap — AK-Freq was deduplicated against Dakshina |
| AK-NEF (foreign names) | 818 | 15.2% | 18.7% | the predicted weak spot |
| AK-NEI (Indian names) | 1,164 | 43.0% | **50.9%** | lexicon +7.9 pts (289 words known) |
| Dakshina (re-included test) | 2,475 | 68.8% | 68.8% | matches our held-out 69.0% — sanity ✓ |

## Reading the numbers honestly

1. **The Dakshina slice agreeing with our own held-out number (68.8% vs 69.0%)
   validates the harness.** Same data, independently ingested, same score.
2. **AK-Freq at 42.6% is a domain-convention gap, not an engine regression.**
   Sampling shows ~29% of failures are pure vowel-length/v-w spelling conventions
   (Aksharantar annotators double aa/ee/oo far more than Dakshina's curated set:
   *atyaachaarapoorn* vs our *atyacharpurn*), and much of the rest is medial-schwa
   variance in the opposite direction (*akharane* where Dakshina style deletes).
   This is the H2 finding writ large: romanization conventions differ *between
   annotation efforts*, not just between annotators. For context, IndicXlit — an
   11M-parameter transformer trained on Aksharantar itself — reports ~52% top-1
   on this direction; a rule engine tuned to Dakshina conventions scoring 42.6%
   against Aksharantar conventions is consistent with that spread.
3. **Named entities confirm the lexicon thesis**: rules alone get 15–43% on
   names; the lexicon adds +7.9 pts on Indian names and +3.5 on foreign ones.
   Names are lexical; more name gold (Aksharantar train NE pairs) would extend it.
4. **Contamination status**: AK-Freq/Dakshina-slice words have zero lexicon
   overlap. The NE slices do overlap the train-derived lexicon (289 + 52 words),
   but Aksharantar's gold was annotated independently of Dakshina train, so
   scoring rules+lexicon here is legitimate; rules-only remains the engine-skill
   number.

## Side finding

The failure sampling initially mis-scored because the CLI splits words on
spaces only — newline-fed stdin defeats word-final rules (a known issue from the
original review). Tracked as a bug task; per-word invocation is correct.

## Status
T-0007 **done**. Follow-ups tracked: COMI-LINGUA sentence benchmark (T-0022),
lyrics gold set (T-0021), and the new whitespace-handling bug task.
