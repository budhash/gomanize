# T-0021 — Lyrics gold seed (public-domain, line-level)

**Date:** 2026-09-05
**Task:** F-0005 / T-0021. The flagship use case — romanizing song lyrics —
finally has a direct, committable benchmark. The research confirmed no aligned
Devanagari↔Roman lyrics dataset exists anywhere; modern film lyrics are
copyrighted, so the repo must never redistribute them.

## Approach: public-domain seed + line-level harness

`benchmark/data/lyrics_gold_hi.csv` — **43 lines of verse/lyrics that are
safely public domain** under Indian copyright (life + 60): Kabir, Rahim and
Meera (15th–16th c.), Vande Mataram (Bankim Chandra Chatterjee, d. 1894),
Sarfaroshi ki Tamanna (Ram Prasad Bismil, d. 1927), Raghupati Raghava
(traditional), Jana Gana Mana (Tagore, d. 1941). Romanizations are
**maintainer-attested single references** — clearly labeled as such in the data
(`attested=maintainer`); they are natural colloquial spellings, not
engine-echoes, so convention divergences (ee/oo) score as errors by design.

`TestBenchmarkLyricsGold` evaluates **line-level** through the public
`gomanize.Translit` sentence API — exercising whitespace/punctuation
segmentation, not just per-word calls. CER is the primary metric (single
reference); word accuracy and exact-line are secondary. A 0.15 mean-CER floor
gates regressions.

## Enabling fix shipped in the same change

`Translit` treated punctuation as part of words, so the danda broke word-final
schwa (मन के जीते **जीत।** → *jita।*). Segmentation now splits on punctuation as
well as whitespace (both preserved verbatim); Devanagari combining marks are
Unicode marks, not punctuation, so words never split internally. जीत। → *jit।*.

## Results

| Metric (43 lines) | Rules | +Lexicon |
|---|---|---|
| **Mean line CER** | **0.0492** | **0.0394** |
| Word accuracy | 83.2% | 86.5% |
| Exact lines (strict single-ref) | 22/43 | 25/43 |

Mean CER on real verse is **at the human round-trip consistency floor (~0.054)**
— on the flagship use case, gomanize's character-level error is comparable to
how much human romanizers disagree with themselves. Residual errors are the
documented convention divergences (dheere/dhire, phool/phul, Sanskrit schwas in
सुजलाम्).

## Expansion path (future)

- Bulk Devanagari lyrics: Giitaayan (v9y/giit, ITRANS, converts
  deterministically) via a local fetch script — held back because bulk text
  without human Roman gold adds little eval value.
- Human gold at scale: LyricsTranslate per-song human transliterations —
  requires case-by-case curation and stays out of the repo (copyright).
- More PD verse (Surdas, Tulsidas, more Kabir) can grow the committed seed.

## Status
T-0021 **done** (seed + harness). F-0005 complete. Remaining backlog: P3 stretch
items T-0017/T-0018 only.
