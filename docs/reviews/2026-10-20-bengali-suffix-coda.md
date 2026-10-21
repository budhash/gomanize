# Bengali suffix boundary and loan-coda entries: negative result (T-0120, T-0127)

**Status:** not shipped. The single dev confirmation lowered match-any for both
profiles, so work stopped as the
[preregistration](2026-10-20-bengali-suffix-coda-prereg.md) requires. `main` is
unchanged. Test and BanglaTLit were not run. The attempt (rule, module-internal
table package, ablation tooling) is preserved at tag `archive/t0120-t0127-attempt`
(commit `2244a3d`), not merged.

[Machine-readable record](2026-10-20-bengali-suffix-coda.json): candidates,
per-entry rounds, rule ablation, train before/after and the dev confirmation.

## Selection (Dakshina train, as preregistered)

**Loan-coda.** There were 70 candidates: the 8 entries on `main` plus 62 final
clusters found in at least 5 train types.
- **Rounds.** Round 1 removed 60. Round 2 kept all 10 that remained.
- **Survivors.** Wins/losses were identical in B1 and model:

  | Entry | Wins / losses |
  |---|---:|
  | ন্ট | 31/0 |
  | ল্ড | 9/0 |
  | ন্স | 9/0 |
  | র্ক | 9/2 |
  | স্ক | 7/1 |
  | ক্ট | 6/0 |
  | র্চ | 6/0 |
  | র্স | 5/0 |
  | ল্ট | 4/0 |
  | ম্প | 4/1 |

- **Shipped entries that failed.** Three entries already on `main` failed the loss
  bound despite a positive net:

  | Entry | Wins / losses |
  |---|---:|
  | ন্ড | 15/9 |
  | ঞ্জ | 9/3 |
  | ক্স | 3/1 |

- **Common native clusters.** Closing them loses heavily: ক্ত 0/54, ন্ত 0/45,
  ত্র 0/41.

**Suffixes.** There were 19 candidates. Round 1 removed 16, and the 3 that
remained passed round 2.

| Suffix | B1 wins / losses | Model wins / losses | Outcome |
|---|---:|---:|---|
| খানা | 4/0 | 1/0 | kept |
| টাকে | 2/0 | 1/0 | kept |
| টিকে | 17/1 | 6/1 | kept |
| কে | 174/47 | 65/40 | failed |
| দের | 103/28 | 54/18 | failed |
| টি | 110/37 | 38/32 | failed |
| রা | 61/13 | 24/11 | failed |
| তে | 50/10 | 8/9 | failed |
| গুলো | 40/16 | 11/16 | failed |

Most failures broke the loss bound. 13 of the 16 broke it under the model and
10 under B1, so the B1 bound alone would have removed কে, দের, টি and গুলো. Five
(খানি, গুলোকে, টার, তে, রা) failed only under the model.

**Rules.** Both rules passed leave-one-out with the surviving tables:

| Rule | B1 wins / losses | Model wins / losses |
|---|---:|---:|
| suffix-boundary | 23/1 | 8/1 |
| loan-coda | 90/4 | 90/4 |

The five 2026-10-14 bundle rules all still passed.

## Results

| Split | Profile | Strict | Match-any | minCER |
|---|---|---:|---:|---:|
| Train (selection) | B1 | 8,268 → 8,294 | 14,370 → 14,403 (+33) | 0.08053 → 0.08060 |
| Train (selection) | Model | 9,238 → 9,252 | 15,656 → 15,674 (+18) | 0.07095 → 0.07110 |
| **Dev (once)** | B1 | 791 → 792 | 1,405 → **1,401 (−4)** | 0.08766 → 0.08810 |
| **Dev (once)** | Model | 909 → 908 | 1,577 → **1,571 (−6)** | 0.07478 → 0.07539 |

"Before" reproduces the current dev pins exactly. The train gain was already small,
and minCER got slightly worse even on train.

## Why it failed (hypotheses, not tested)

Testing either hypothesis would take a seventh use of dev, after the failure had
been seen. That is exactly the forking path the preregistration exists to prevent,
so neither is pursued.

1. **The suffix rule's scope was wrong in the preregistration.** The 2026-10-10
   assessment scoped the layer to *conjunct-bearing stems*, where the vowel model
   does not run. The preregistration let it fire on every stem, so it overrode
   the vowel model on words the model already handles. This is consistent with
   the model's much worse loss-to-win ratio (কে: 40/65 under the model
   against 47/174 under B1).
2. **The per-entry loss bound was stricter than the rule-level one.** It removed
   three net-positive `main` entries (ন্ড, ঞ্জ, ক্স) whose wins may generalize
   better than their train losses suggest.

## Other findings

- **Quadratic cost of the suffix condition (from the security review).** For each
  consonant, the prototype condition rebuilt the word tail and then prepended
  units walking backwards. On one long unspaced Bengali run, that is superlinear
  work per word. A future suffix rule should match a bounded-length table against
  the word ending once.
- **The mechanism is sound.** Moving tables into a module-internal package
  (`internal/bengalitables`) allowed per-entry ablation without a public-API
  knob. Reuse it if a table-driven rule returns (plan item 11).

## Data discipline

- This was the **sixth** use of Bengali dev. Dev figures for any later work on
  these words remain optimistic.
- The preregistration was committed and pushed before any code or run.
- Test was never run for this work.

## Next

T-0120 and T-0127 are closed as tried. The words they targeted (suffixed and
conjunct-bearing) are the main target of T-0121, which extends the vowel model to
conjunct-bearing words. That addresses the same gap with the component that
already wins on eligible words.
