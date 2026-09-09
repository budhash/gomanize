# Structured parser QA — finding systematic bugs without manual sifting

**Date:** 2026-09-08 · **Status:** plan (tracked as F-0010) · **Trigger:** the
`गई → gi` bug found in real lyrics text.

## The bug that started this

`गई` is `ग` (GA) + **`ई`** — the *independent* vowel ई (U+0908), not the matra
`ी` (U+0940). It is two syllables, `ga-ī`, and should romanize `gai`/`gayi`/
`gaee`. gomanize returns `gi`: it parses the independent vowel as if it were the
consonant's matra, dropping ग's inherent `a` (the debug trace even marks the
schwa `[Keep]`, yet the output has no `a`) and mapping the long vowel to short.
The smoking gun: `गी` (matra) and `गई` (independent) produce the **identical**
output `gi`; likewise `की`/`कई` both give `ki`. The class is common in running
text — गई, नई, कई, हुई, गए, लिए — so it matters for the lyrics use case.

## Why every dataset missed it

Scanning the corpora for "a consonant immediately followed by an independent
vowel":

- Curated set: **2 / 1331 (0.2%)**. Full Dakshina: **247 / 53,065 (0.5%)**.
- Those few are almost all acronyms/proper nouns (`आईएसएल→isl`,
  `अय्यदुरई→ayyadurai`, `मुंबई→mumbai`) — not the common grammatical forms.

Two structural reasons the aggregate accuracy never flagged it:

1. **Averages hide rare constructs.** A class that is 0.2% of tokens can be 100%
   wrong while the headline pure number (86.1%) barely moves. The metric measures
   average correctness, not coverage of the orthographic construct space.
2. **Distribution mismatch.** The benchmarks are dictionary **word-lists**; the
   affected forms are *inflectional/verbal* and live in **running lyrics text**,
   which the corpora under-sample. (The lexicon incidentally covers `लिए→liye`
   and `हुई→hui` but not गई/नई/कई — coverage is accidental, not systematic.)

**Conclusion:** more example-scoring will not help — the examples share the same
blind spots. We need **construct-space coverage and structural invariants**, not
more averages.

## Approach — five tiers (cheap → thorough)

Each tier is independent and adds signal; none replaces the existing aggregate
gate (pure ≥ 85% on curated Dakshina), which stays as-is.

### Tier 1 — Per-construct accuracy report *(find)*
Bucket every token in the full corpora by the orthographic features it contains
— independent-vowel-after-consonant, each matra, conjuncts, nukta, anusvara /
chandrabindu / visarga, halant-final, … — and report per-construct accuracy
ranked by `prevalence × error-rate`. One run enumerates every systematic
weakness from data we already have. Deliverable: a `make` target + a benchmark
test that prints the ranked table. This is what *surfaces* the lurking bugs.

### Tier 2 — Invariant / property tests *(catch by contradiction, gold-free)*
Structural properties that must hold for any input, needing no "correct" answer:
- a consonant whose schwa is `[Keep]` **must** contribute a vowel to the output;
- an independent vowel **must** become its own output vowel, never absorbed as a
  matra;
- **differential:** the matra form and the independent form must differ
  (`गी ≠ गई`, `की ≠ कई`) — a strong, convention-free check;
- the count of vowel nuclei out ≈ vowel-bearing units parsed (no dropped/added
  syllables).
The `गी ≠ गई` check alone is a one-screen test that would have caught this on
day one.

### Tier 3 — Combinatorial construct enumeration *(guarantee coverage)*
Machine-generate the cross-product `consonant × {inherent, each matra, each
following independent vowel, halant-conjunct, +nukta, +nasal}` and assert the
**parse** is well-formed (correct unit count and types) and the Tier-2
invariants hold. Thousands of machine-checked cases; no manual sifting. Turns
"hope the dataset covers it" into "every construct is provably exercised."

### Tier 4 — Differential parse vs. a Unicode akshara reference *(deepest)*
Compare gomanize's segmentation into units against Unicode grapheme-cluster /
akshara boundaries (UAX #29). Divergences are parser bugs, independent of
romanization convention. (गई: Unicode treats ग and ई as separate clusters;
gomanize merges them.)

### Tier 5 — Construct-anchored golden set *(fix + guard)*
A small test file whose expected values are fixed by **linguistic rule, not
annotation convention** (e.g. "a consonant + independent vowel keeps the
consonant's inherent vowel"): गई→gai, नई→nayi, कई→kai, गए→gaye, …. Serves as the
TDD anchor for each fix and a permanent regression guard. Complements (does not
replace) growing the lyrics gold set toward the real distribution.

## Sequencing

- **Phase A (find):** Tier 1 analyzer + Tier 2 invariants → a ranked,
  data-driven list of every lurking parser issue.
- **Phase B (fix highest-impact):** repair the consonant + independent-vowel
  class against a Tier-5 golden; measure on Dakshina — pure must not regress.
- **Phase C (durable net):** Tier 3 combinatorial coverage wired into CI.
- **Phase D (deep):** Tier 4 differential vs. the Unicode reference.

Each phase is its own PR, benchmark-gated. Fixes surfaced by Phase A are triaged
and filed individually rather than fixed en masse.

## Success criteria

- `make` target prints per-construct accuracy ranked by `prevalence × error`.
- Invariant tests in CI fail on the `गी == गई` class (and pass once fixed).
- The independent-vowel class is fixed with **no curated-Dakshina pure
  regression** (gate stays ≥ 85%).
- Every `consonant × {matra / independent / conjunct / nasal}` combination is
  machine-exercised.

## Non-goals / guardrails

- **Do not chase convention variance.** Invariants must be *structural* (a
  dropped syllable is a bug; `gai` vs `gayi` vs `gaee` is a convention choice,
  decided in the scheme, not asserted as a universal truth).
- The aggregate gate remains the release gate; construct slices are additional
  diagnostics, not a replacement.
- No neural / heavyweight dependency — consistent with the project's identity
  ([`ROADMAP.md`](../ROADMAP.md) "What to avoid").

## Decision

Adopt the tiered approach. Track as **F-0010**; build incrementally, Phase A
first.
