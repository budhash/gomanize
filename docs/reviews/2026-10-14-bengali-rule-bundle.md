# Bengali convention-neutral rule bundle (T-0119, 2026-10-14)

**Status:** shipped (experimental Bengali). Item 4 of the
[quality assessment](2026-10-10-bengali-quality-assessment.md) plan, without the
স/ছ spelling change (T-0117/T-0118, pending a maintainer decision), the suffix
layer (T-0120) or shared CCV (T-0123). No vowel-model, lexicon or Hindi change.

[Machine-readable record](2026-10-14-bengali-rule-bundle.json), written by
`tools/bengali/evaluate_rule_bundle.py` and asserted by
`tools/bengali_profiles_test.py` (`make test-bengali-profiles`).

## What shipped

Five named rules in `lang/bengali/rules.go`, each disableable with
`WithDisabledRules` / `--disable-rule`:

| Key | Rule(s) | Mechanism | Example (before → after) |
|---|---|---|---|
| ito | `schwa.bengali.final-ito` | Final bare ত after ি/ী/ৃ/ূ keeps its vowel | চিহ্নিত chinnit → chinnito; মৃত mrit → mrito |
| tacons | `schwa.bengali.final-ta` | Final bare ত after a bare গ র হ ন শ ষ keeps its vowel | গত got → goto; ভারত bharot → bharoto |
| coda | `schwa.bengali.loan-coda` | Final ন্ট ন্ড ঞ্জ ন্স ক্স র্স র্ক ল্ড close (র্ট/র্ড/স্ট already did) | পেইন্ট peinto → peint; গোল্ড goldo → gold |
| osuf | `vowel.bengali.final-o` | Final ও after a bare consonant adds no second `o` (হও excepted) | এখনও ekhonoo → ekhono; কও koo → ko |
| yph | `consonant.bengali.sibilant-ya`, `render.bengali.sibilant-ya` | Medial ya-phala after শ/ষ/স is `y`, not a geminate | দৃশ্যের drishsher → drishyer |

Unchanged contexts (unit-tested): মত/যত `mot`/`jot`, রাত `rat`, চিত্ত `chitto`,
শরৎ `shorot`, explicit-hasant হিত্ `hit`, অন্ত `onto`, মন্ত্র `montro`, আর্ট `art`,
যাও `jao`, তবুও `tobuo`, ও `o`, initial শ্যাম `shyam`. The B1 construct examples
(kormo, moddho, dukkho, khoti, kokkho, gyan, chinno, hao, howa) are unchanged.

**Rejected:** `bph` (ba-phala after শ/স as `w`, medial দ্ব as `db`; বিশ্বাস
bishwas) failed the criterion and was removed from the code.

## Preregistered criterion (train only)

Fixed before dev was run: a rule ships iff, on Dakshina **train** (25,000 words),
disabling it alone from the full bundle loses net match-any **and** its paired
losses are at most 25% of its wins, for **both** the rules-only (B1) and
vowel-model profiles. Failing rules are removed, not left disabled. Dev is
confirmation only (stop if the shipped bundle does not improve it for both
profiles); test and BanglaTLit are reported once after freezing; Aksharantar is
not run. Ablation uses the engine's disable mechanism (`render_profiles -disable`),
not code edits.

### Selection ablation (six candidates)

Wins: words the full bundle matches and the bundle minus the rule does not.

| Rule | B1 wins / losses / net | Model wins / losses / net | Ships |
|---|---:|---:|---|
| ito | 275 / 27 / +248 | 206 / 11 / +195 | yes |
| tacons | 55 / 10 / +45 | 29 / 4 / +25 | yes |
| coda | 90 / 15 / +75 | 90 / 15 / +75 | yes |
| osuf | 77 / 7 / +70 | 77 / 7 / +70 | yes |
| yph | 35 / 7 / +28 | 35 / 7 / +28 | yes |
| bph | 37 / 30 / +7 | 37 / 30 / +7 | **no** (30 > 9.25) |

### Shipped-bundle ablation (asserted)

Recomputed against the five-rule bundle; only `osuf` moves (one more win,
an interaction with the removed ba-phala rule). Every rule still passes.

| Rule | B1 wins / losses / net (changed) | Model wins / losses / net (changed) |
|---|---:|---:|
| ito | 275 / 27 / +248 (482) | 206 / 11 / +195 (356) |
| tacons | 55 / 10 / +45 (117) | 29 / 4 / +25 (59) |
| coda | 90 / 15 / +75 (348) | 90 / 15 / +75 (348) |
| osuf | 78 / 7 / +71 (130) | 78 / 7 / +71 (130) |
| yph | 35 / 7 / +28 (118) | 35 / 7 / +28 (118) |

## Results

"Before" is the same engine with every bundle rule disabled; it reproduces the
previous pins and the 2026-10-05 runtime record exactly. Strict is the
top-attested reference; minCER is the macro minimum CER.

| Split | Profile | Strict | Match-any | minCER |
|---|---|---:|---:|---:|
| Train (selection) | B1 | 8,016 → 8,268 | 13,903 → 14,370 (55.61% → 57.48%) | 0.08470 → 0.08053 |
| Train (selection) | Model | 9,026 → 9,238 | 15,262 → 15,656 (61.05% → 62.62%) | 0.07471 → 0.07095 |
| Dev (confirmation) | B1 | 775 → 791 | 1,371 → 1,405 (54.84% → 56.20%) | 0.09138 → 0.08766 |
| Dev (confirmation) | Model | 895 → 909 | 1,545 → 1,577 (61.80% → 63.08%) | 0.07839 → 0.07478 |
| **Test (once)** | B1 | 800 → 828 (32.00% → 33.12%) | 1,413 → 1,467 (**56.52% → 58.68%**) | 0.08772 → 0.08368 |
| **Test (once)** | Model | 914 → 937 (36.56% → 37.48%) | 1,569 → 1,615 (**62.76% → 64.60%**) | 0.07642 → 0.07278 |

The lexicon and model + lexicon profiles equal B1 and model on dev and test
(held-out words are excluded from the lexicon; the evaluator fails otherwise).

Dev confirmed the bundle for both profiles, so nothing was iterated. Test was
run once, after the bundle was frozen.

**BanglaTLit** (2,500 sentences; before from the 2026-10-05 runtime record):

| Profile | Full: exact, macro CER | All-tokens-unseen (38): exact, macro CER |
|---|---|---|
| B1 | 8 → 8, 0.33149 → 0.32727 | 1 → 1, 0.53180 → 0.51708 |
| Model | 12 → 13, 0.31286 → 0.30871 | 2 → 2, 0.52096 → 0.50624 |
| Lexicon | 13 → 13, 0.29402 → 0.29065 | 1 → 1, 0.53180 → 0.51708 |
| Model + lexicon | 17 → 18, 0.28144 → 0.27820 | 2 → 2, 0.52096 → 0.50624 |

## Pins and records moved (deliberately)

| Pin | Before | After |
|---|---|---|
| `bnPinnedRules` train (strict / match-any) | 8,016 / 13,903 | 8,268 / 14,370 |
| `bnPinnedRules` dev | 775 / 1,371 | 791 / 1,405 |
| `bnPinnedModel` train (strict / match-any / changed vs B1) | 9,026 / 15,262 / 3,372 | 9,238 / 15,656 / 3,219 |
| `bnPinnedModel` dev | 895 / 1,545 / 408 | 909 / 1,577 / 393 |
| `bnPinnedLexiconTrain` B1-lexicon (strict / match-any) | 12,602 / 17,400 | 12,739 / 17,718 |
| `bnPinnedLexiconTrain` model-lexicon | 13,045 / 18,112 | 13,159 / 18,373 |

All output digests changed with them. The preregistered B1 dev gate
(`make test-bengali-b1-gate`) passes unchanged. The Hindi snapshot is unchanged
(the rules are Bengali-only). `tools/bengali_profiles_test.py` now asserts this
record (ablation table, train/dev/test before and after, BanglaTLit) instead of
the 2026-10-05 runtime record, which stays as history.

**Lyrics pilot** (82 lines; regenerated with `tools/bengali/evaluate_lyrics.py`,
references unchanged). Exact lines / macro line CER against the **unreviewed
assistant drafts** (T-0068), so this is agreement, not a quality claim:

| Profile | Before | After |
|---|---|---|
| B1 | 43, 0.021638 | 47, 0.019476 |
| Model | 52, 0.017664 | 55, 0.016324 |
| Lexicon | 37, 0.029390 | 40, 0.027664 |
| Model + lexicon | 47, 0.025304 | 49, 0.024400 |

## Data discipline

- Selection used Dakshina train only, with the criterion above fixed first. The
  scratch reference implementation's single-rule train deltas (from the
  2026-10-10 review) informed the candidate list, not the decision.
- Dev was previously consulted for these candidates by the 2026-10-10 review
  (the **fifth** use of Bengali dev); here it was only a confirmation step and
  changed nothing. Dev figures remain mildly optimistic.
- Test and BanglaTLit were run once, after freezing. Unit-test examples avoid
  dev/test words except দৃশ্যের and বিশ্বাস, the spec's own examples.
