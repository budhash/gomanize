# H3 — Lexicon layer (honest coverage, not a benchmark number)

**Date:** 2026-09-04
**Task:** F-0004 / T-0016. A high-confidence romanization lexicon: known words
return the attested human spelling; unknown words fall through to the rule
engine. The standard TTS/G2P architecture (lexicon + fallback).

## The honesty constraint (stated up front)

Dakshina's train/dev/test native words are **fully disjoint** (0 overlap — see
the H3 schwa record). A lexicon built from train therefore covers **~0% of the
held-out test split**, and the curated benchmark set is drawn from the *same*
train pool — so scoring the curated set *with* the lexicon would be circular
(measuring memorization of the answer key).

Consequently this layer is **deliberately excluded from the accuracy benchmark**,
and this record reports coverage honestly instead of a manufactured accuracy jump.

## What was built

- `tools/schwa/build_lexicon.py` → `lang/hindi/lexicon.tsv`: for each Dakshina
  TRAIN native, the most-attested Roman spelling, kept only when that spelling has
  **≥4 attestations** (multiple annotators agreeing). **1,896 entries (~46 KB).**
- `lang/hindi/lexicon.go`: `go:embed` the TSV; `LexiconLookup(word)` on `Hindi`.
- `core.LexiconProvider` interface + `Options.Lexicon` (`--lexicon`): the engine
  consults the lexicon before the rule pipeline; a hit short-circuits, a miss
  falls through unchanged. `core` stays language-agnostic (interface check only).

## What it does — and doesn't

The lexicon captures exactly the classes rules provably cannot (H2/H3 findings):
loanwords and names with conventional spellings.

| Word | Rules | Lexicon |
|---|---|---|
| अंकल | ankal | **uncle** ✓ |
| अंडरकवर | (rule output) | **undercover** ✓ |
| अंग्रेजी | angreji | angreji (agree) |
| क्षमाशीलता (OOV) | kshamashilta | kshamashilta (lossless fallthrough) |

**Held-out Dakshina TEST split (2,500 words):**

| Metric | Value |
|---|---|
| Lexicon size | 1,896 |
| Test words covered | **0 / 2,500 (0.0%)** — expected (disjoint splits) |
| Match-any: rules | 69.0% |
| Match-any: rules + lexicon | 69.0% (unchanged) |

The lexicon does **not** change held-out *type* accuracy — by construction. Its
value is production *token* coverage: real text (song lyrics, common vocabulary)
reuses frequent words, and for those the tool now emits the attested human
spelling. Measuring that gain needs a running-text corpus we don't have offline;
we do not fake it with a circular in-domain number.

## Why this is still worth shipping

1. It fixes the one class rules and the schwa model both fail — loanwords/names —
   for the common vocabulary that dominates real usage.
2. It is lossless: OOV words are byte-identical to rules (asserted in tests), so
   enabling it can never regress output.
3. It generalizes and replaces the ad-hoc `override_hi.csv` mechanism with a
   principled, attestation-weighted, data-generated lexicon.

## Status
- Opt-in via `--lexicon` / `core.Options{Lexicon: true}`. Default byte-identical.
- T-0016 **done**. Natural follow-ups: T-0018 (per-word candidate re-ranking to
  combine lexicon + model + rules), and a frequency-weighted corpus to quantify
  token coverage. Rebuild: `python3 tools/schwa/build_lexicon.py [min_attestations]`.
