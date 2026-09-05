# T-0022 — COMI-LINGUA colloquial benchmark

**Date:** 2026-09-05
**Task:** F-0005 / T-0022. The first benchmark of gomanize against Hindi as
people *actually type it* in Latin script — the closest available proxy to the
song-lyrics use case until a lyrics gold set exists (T-0021).

## Source & extraction

**COMI-LINGUA** (LingoIITGN, Findings of EMNLP 2025, **CC-BY 4.0**): the MT test
split carries, per human annotator (3 per sentence), parallel Devanagari-Hindi
and Romanized-Hindi translations of the same sentence — naturally-typed
romanization, expert-annotated.

`tools/build_comilingua.py` extracts word pairs:
- Only human annotator columns (machine `Predicted_*` columns ignored).
- Sentences with equal token counts → per-position (Devanagari, Latin) pairs.
- **Transliteration filter**: annotators sometimes translate (अंक → "points");
  an engine-neutral consonant-skeleton check (word's consonants must appear in
  order in the roman, ≥70%) removes translations while keeping variant spellings.
- Variants aggregated per word with occurrence counts → naturally
  token-weighted toward the colloquial register.

Result: `benchmark/data/comilingua_hi.csv` — 12,296 rows, **9,606 unique words,
152,665 token occurrences** (452 KB, CC-BY with attribution). The variant
distributions are the colloquial ground truth in miniature: में → *mein* (5,054)
/ *me* (186); नहीं → *nahin* (224) / *nahi* (197); करें → *karen* (41) /
*karein* (25).

## Results (match-any over attested variants)

| Metric | Rules | +Lexicon |
|---|---|---|
| **Token-weighted** (the real-world number) | **78.9%** | **85.7%** |
| Type-level (unweighted) | 48.1% | 55.8% |
| Mean minCER (type-level) | 0.182 | 0.156 |

## Reading it honestly

- **Token-weighted is the number that matters** for running text: on the words
  people actually use, weighted by how often they use them, gomanize+lexicon
  romanizes **85.7%** of tokens to a form some human annotator actually typed.
- The type-level tail (48%) is dominated by rare words, residual alignment noise
  (single-count pairs can carry annotator typos or loose alignments), and
  English loanwords typed as English — treat it as a lower bound, not a grade.
- The lexicon again earns +6.8 token-weighted points on an independently
  annotated corpus — consistent with the Aksharantar NE result and the
  frequency-weighted eval (T-0020).
- CER caveat: mean minCER is type-level (unweighted) and inflated by the tail.

## Status
T-0022 **done**. Remaining in F-0005: the curated lyrics gold set (T-0021) and
the CLI whitespace fix (T-0023) which would enable sentence-level scoring here.
