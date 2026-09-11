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
  *Shipped (T-0041):* `script/brahmic/combinatorial_test.go` enumerates the full
  cross-product off the Hindi symbol map (~1,900 constructs) and asserts parse
  well-formedness — unit count/types, IsMatra vs independent-vowel, after-halant
  conjuncts, precomposed-nukta combination, monotonic non-overlapping unit
  spans. `make test-combinatorial`; runs in CI via `test-cover`. Two initial
  assertions were relaxed to documented behaviour (nukta combines only where a
  precomposed mapping exists; halant is consumed, so conjunct unit spans have a
  one-rune gap) — no parser bugs surfaced.
- **Phase D (deep):** Tier 4 differential vs. the Unicode reference.
  *Shipped (T-0042):* `benchmark/akshara_test.go` checks gomanize's unit
  boundaries against an in-repo Devanagari akshara-boundary reference
  (dependency-free — no external UAX #29 library). gomanize splits finer than
  aksharas, so the relation is **refinement**: every akshara boundary must be a
  gomanize unit boundary (a *missing* one = the segmentation shape of the गई
  merge). Run over 30,000 Dakshina natives + synthetic construct cases:
  **0 missing boundaries**. `make test-akshara`; runs in CI via `test-cover`.
  Honest scope note: T-0040 was a *rendering* bug (schwa folded), not a
  segmentation merge — the parser always kept गई as two units — so this tier
  guards segmentation correctness rather than re-catching T-0040; a negative
  control confirms the reference *would* flag a real merge.

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

## Codex review — refinements (2026-09-08)

A design review by Codex (gpt-5.5) confirmed the shape and sharpened it. Key
points folded into the plan:

**Root causes located in code (targets for the T-0040 fix):**
- `script/brahmic/categories.go` maps *both* `CatVowel` (independent) and
  `CatMatra` to `UnitVowel`, and `script/brahmic/renderer.go` suppresses the
  inherent schwa before *any* `UnitVowel` — that is exactly why `गी` (matra) and
  `गई` (independent) collapse to the same output. The fix must let the renderer
  distinguish an independent vowel from a matra.
- `lang/hindi/rules.go` `render.chandrabindu.final-silent` does **not** require
  `u.IsWordFinal()`, so it fires mid-word too: `चाँद→chaad`, `चाँदी→chadi`, not
  just word-final `कहाँ→kahaa`. Scope the rule to word-final.

**Regression-safety upgrades:**
- "Net-positive" is too weak alone — stratify by construct × dataset × split ×
  option mode × reference quality, or one high-severity structural regression
  gets washed out by many low-value wins.
- The corpus diff must be a **transition matrix**, not a changed-token list:
  per native × reference set, classify `improved_exact` / `regressed_exact` /
  `improved_distance` / `regressed_distance` (minCER beyond epsilon) /
  `variant_drift` (hit→hit but output changed) / `neutral`. **Tag lexical /
  acronym rows** (`आईएसएल→isl`, `अंकल→uncle`) so CER wins on unromanizable
  loanwords don't masquerade as parser improvements. (T-0043)
- Evaluate the parser diff **rules-only** — the lexicon is built from Dakshina
  train and curated is drawn from the same pool, so lexicon-mode scoring on
  curated is circular; report lexicon/schwa-model/rerank modes separately, and
  weight the independent Aksharantar / COMI-LINGUA evidence more. But a parser
  fix must still be **checked in the rerank and schwa-model paths**, not only
  rules-only.
- `match-any` ignores attestation weight, so a shift from the dominant variant to
  a rare one still scores neutral — hence `variant_drift` reporting weighted by
  Dakshina attestation counts.
- The curated **pure ≥ 85% gate stays as a continuity floor but is not the
  parser safety net** — it has 0–2 examples of the target constructs.
- The **construct-anchored golden must use accepted reference *sets***
  (`गई→{gai,gayi}`, `नई→{nai,nayi}`, `कई→{kai,kayi}`, `गए→{gae,gaye}`,
  `कहाँ→{kahan}`), plus a `गी ≠ गई` contrast; low-N constructs are guarded by
  golden coverage rather than statistical floors.
- Beyond `गई`, the fix should cover related shapes: consonant + `ए/ओ`,
  matra + independent vowel, modifier + independent vowel, and non-final
  chandrabindu.

**Task deltas:** T-0039 (this) adds the gold-free invariants incl. the
`गी ≠ गई` differential (staged as skipped until T-0040 fixes it, to keep CI
green while documenting the bug); T-0040 targets the `categories.go`/`renderer.go`
root cause + the chandrabindu word-final scope, with reference-set goldens and
multi-mode checks; T-0041 becomes a *gated* analyzer (floors/golden-coverage, not
just logging). New: **T-0043** (transition-matrix corpus diff) and **T-0044**
(held-out running-text regression set).
