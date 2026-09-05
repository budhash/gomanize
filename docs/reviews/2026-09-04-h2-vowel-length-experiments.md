# H2 Decision Record — Vowel-length rules are lexical, not rule-governed

**Date:** 2026-09-04
**Context:** H2 (F-0003) aimed to push the rule ceiling. The review estimated the
medial **ee/oo** rule as the biggest tractable win (~+2 pts) and a narrow
**ā→aa** extension as a secondary gain. Now that H1 gave us an honest
multi-reference metric, both were implemented and measured before shipping.
**Outcome: both are net-negative and were not shipped.**

## What was tried

Baseline (no change): strict 86.1% · **match-any 92.8%** · mean minCER **0.0116**.

### 1. Medial ee/oo (T-0010)
Rule: long-i ी/ई → `ee` and long-u ू/ऊ → `oo` when word-**medial**; keep i/u
word-final (data supports this: word-final ी attests i/u 168:10, ू 12:3).
Examples produced correctly: संगीत→sangeet, फूल→phool, हिंदी→hindi, नहीं→nahin.

| Variant | Strict | Match-any | Mean minCER |
|---|---|---|---|
| Baseline | 86.1% | **92.8%** | 0.0116 |
| ee + oo | 82.5% | 89.0% | 0.0224 |
| ee only (ी) | 84.9% | 90.6% | 0.0175 |
| oo only (ू) | 83.7% | 91.2% | 0.0165 |

Every variant loses. The data explains why: medial ी is a near-even split
(27 words attest `ee`, 26 attest only `i`); medial ू is 30 vs 21. Switching to
ee/oo breaks as many currently-correct i/u spellings as it fixes.

### 2. Narrow ā→aa, medial open syllable (T-0011)
Rule: ा→`aa` when followed by a non-word-final consonant (गाना→gaana pattern).

| Variant | Strict | Match-any | Mean minCER |
|---|---|---|---|
| Baseline | 86.1% | **92.8%** | 0.0116 |
| medial-open ā→aa | 68.8% | **80.0%** | 0.0329 |

A large loss — e.g. भारत→bhaarat (want bharat). Medial ा in open syllables is
dominated by single-`a` cases. (Consistent with `--long-vowels` scoring 60.4%.)

## Conclusion

The remaining vowel-length gap is **lexical, not rule-governed**: surface form
does not predict whether a medial long vowel romanizes as ee/oo or i/u, or
whether ा doubles. No positional rule can satisfy the split without breaking the
other half. **The rule-based ceiling is reached** for vowel length, exactly as
the state-of-project review predicted (rule systems top out ~94–95% per-schwa;
the residual is loanwords/compounds/lexical variation).

This is a positive result, not a dead end: H1's honest metric prevented a
benchmark-fitting change that would have "fixed" 29 single-reference failures
while regressing real quality by ~3–13 points. It redirects effort to **H3**
(F-0004): a frequency-ranked lexicon and/or a distilled learned component, which
is the only approach the literature shows can close a lexical gap.

## Disposition
- T-0010, T-0011: **skipped** (investigated, rejected on evidence). Reproduce via
  the experiment rules in git history for this branch, or re-run against
  `TestBenchmarkMultiReference`.
- H2 ships its remaining value: unit-test coverage for `lang/hindi` and
  `script/brahmic` (T-0014) — the gap the review flagged (those packages had zero
  direct tests). The golden set includes these vowel-length spellings as
  deliberate, now-evidence-backed divergences.
