# Bengali suffix boundary and loan-coda entries: preregistration (T-0120, T-0127)

**Status:** preregistered 2026-10-20, committed before any code change, train
run or dev run for this work. The results record will be a separate document;
this file is not edited after the first commit except to link to it.

Plan item 5 of the [quality assessment](2026-10-10-bengali-quality-assessment.md)
(suffix-boundary deletion) and the per-entry loan-coda follow-up from the
[rule bundle](2026-10-14-bengali-rule-bundle.md) review. No স/ছ change
(T-0117/T-0118), no shared CCV (T-0123), no vowel-model, lexicon or Hindi change.

## Candidates (fixed now)

**Suffix table (`suf`).** A rule deletes the inherent vowel of a stem-final bare
consonant (no matra, no hasant, not the first akshara of the word) that is
immediately followed by one of these word-final nominal suffixes. The longest
suffix matches first.

কে টি টা টার টির টাকে টিকে গুলো গুলি গুলোর গুলির গুলোকে গুলিকে তে দের দেরকে রা খানা খানি

Example: অঞ্চলকে `oncholoke` → `oncholke`.

**Loan-coda table (`coda`).** The candidates are the eight shipped entries plus
every other word-final simple cluster (C + hasant + C) on which the engine
currently keeps the final vowel, and which occurs in at least 5 distinct Dakshina
**train** word types. A tool enumerates this list from train only and writes it
into the record. The older `loanCoda` entries (র্ট র্ড স্ট) are B1-gated and are
not candidates.

## Criterion (train only)

Profiles: rules only (B1) and the vowel model. The data is Dakshina train
(25,000 words). The metric is paired match-any wins and losses. An entry or rule
**passes** if both of these hold for **both** profiles:

- net > 0
- losses ≤ 0.25 × wins

1. **Per entry.** Each candidate entry is ablated leave-one-out from its full
   candidate table. The rest of the engine is current `main` plus the table.
   Failing entries are removed. This repeats until every remaining entry passes,
   up to three rounds. An entry that still fails after three rounds is removed.
2. **Per rule.** With the surviving tables, each of `suf` and `coda` is ablated
   leave-one-out from the full engine. A failing rule is removed from the code.
   The five 2026-10-14 bundle rules are re-ablated and reported. If one of them
   now fails, it is reported, but it is not removed in this PR.
3. **Removal.** Removed entries and rules are deleted from the code, not left
   disabled.

## Held-out use

- **Dev is confirmation only, run once after freezing.** The shipped change must
  raise dev match-any over current `main` for both profiles. If it does not, the
  work stops and is reported. Nothing is iterated on dev.
- **Disclosure.** This is the **sixth** use of Bengali dev. The 2026-10-10 review
  already measured a `suf` prototype on dev (+21 model).
- **Test and BanglaTLit** are run once, after dev confirmation.
- **Aksharantar** is not run.
- **Gates.** The B1 dev gate and the Hindi snapshot must pass unchanged.
- **Unit-test examples** come from train or are constructed. They avoid dev and
  test words.
