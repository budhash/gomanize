# Bengali vowel model on conjunct-bearing words: preregistration (T-0121)

**Status:** preregistered on 2026-10-21. This file was committed before any code
change, training run or Dakshina run for this work. Results go in a separate
record. This file is not edited after its first commit, except to link to that
record.

**Record:** [2026-10-21-bengali-vowel-conjuncts.md](2026-10-21-bengali-vowel-conjuncts.md).

This is plan item 6 of the [quality assessment](2026-10-10-bengali-quality-assessment.md).
It follows the [suffix/coda negative result](2026-10-20-bengali-suffix-coda.md):
the conjunct-bearing words that rule tried to reach are given to the vowel model
itself. There is no rule-table, symbol, lexicon or Hindi change, and no স/ছ
change (T-0117/T-0118).

## Source data

The pinned Google lexicon `.gz` (`GOOGLE_SHA`) is no longer available locally.
Upstream `googlei18n/language-resources` `bn/data/lexicon.tsv` (last commit for
that path: `49f1dbd`) has uncompressed SHA-256
`c9db2307a004c5b34022ad0d1a6429900e71251f2fe1ef967e8889e5806df268`. Before this
file was written, it was checked to:

- reproduce all six frozen partition digests in `training/data/bengali/manifest.json`
  (row count 65,037);
- retrain the shipped `lang/bengali/vowel_tree.json` **byte-identically** with
  the current code.

Only the gzip bytes differ. The loader will also accept the raw `.tsv` verified
by that content digest. The frozen partitions and manifest are not changed.

## The single variant (fixed now)

1. **Scope.** A word becomes eligible if it has halant clusters (two or three
   consonants joined by ্). Every cluster member must be one of the model's
   existing consonant letters. This includes reph (র্), ya-, ba- and ma-phala,
   ক্ষ and জ্ঞ. Everything currently excluded stays excluded: ং ঁ ঃ ৎ য় ঋ/ৃ ঐ/ৈ
   ঔ/ৌ, other letters, and an inherent slot before an independent vowel.
2. **Units.** A consonant followed by ্ has no inherent slot. A cluster's last
   member has a slot unless a matra follows, the same as a simple consonant.
3. **Features.** The same eight features with the same definitions; `prev` and
   `next` may now be ্. The same CART is used, with depth 10 and min-leaf 20, and
   no hyperparameter search.
4. **Aligner.** Alternatives change **consonant phones only**.
   - A non-first cluster member may match its own phone(s), the previous member's
     phone (gemination), or nothing.
   - ষ after ক may also match `kh`, and ঞ after জ may also match `g`.
   - Vowel handling, the ambiguity rejection and whole-type rejection are
     unchanged.
   - Further consonant-only alternatives may be added only from Google **train**
     alignment failures, before any Dakshina run. They must be listed in the
     record.
5. **One shared eligibility inventory.** A single data file defines the letters,
   matras, independent vowels and cluster policy. It is embedded by
   `vowel_model.go` and read by `vowels.py`. The parity fixture is regenerated,
   and a test asserts that Go and Python agree on eligibility.
6. **Precedence is unchanged.** `schwa.bengali.model` stays at priority 70. Rules
   that override the model today still do. Rules-only (B1) output, the B1 dev
   gate and the Hindi snapshot must be byte-identical.

## Criterion

**Selection (Dakshina train only).** The model profile ships only if, against
the current `main` model profile:

- net paired match-any is > 0, **and**
- macro minCER is not worse.

Wins and losses are reported, split by words newly eligible and words eligible
before. There is no loss-ratio gate: this is a retrained model, not a rule.

Google dev/test classifier accuracy is reported as in the B2 record and never
selects. If the train criterion fails, the work stops with a negative-result
record.

## Held-out use

- **Dakshina dev: confirmation only, run once.** Model-profile match-any must
  rise and minCER must not get worse. Otherwise the work stops and is reported.
  This is the **seventh** use of Bengali dev.
- **Dakshina test and BanglaTLit:** run once, after dev confirmation.
  Aksharantar is not run.
- **Lyrics record:** regenerated and reported as agreement with unreviewed drafts
  (T-0068).
- **Coverage stratification:** the headline is reported next to accuracy split
  into eligible and ineligible words.
