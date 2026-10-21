# Bengali vowel model on conjunct-bearing words (T-0121, 2026-10-21)

**Status:** shipped (experimental Bengali). Plan item 6 of the
[quality assessment](2026-10-10-bengali-quality-assessment.md), run as
[preregistered](2026-10-21-bengali-vowel-conjuncts-prereg.md) with the six
deviations listed below. No rule-table, symbol, lexicon or Hindi change, and no স/ছ
change.

[Machine-readable record](2026-10-21-bengali-vowel-conjuncts.json), written by
`tools/bengali/evaluate_vowel_conjuncts.py` and asserted by
`tools/bengali_profiles_test.py` (`make test-bengali-profiles`). The test checks
the train selection numbers too. This record supersedes the
[rule-bundle record](2026-10-14-bengali-rule-bundle.md) as the asserted profile
record.

## What changed

- **Shared inventory.** `lang/bengali/vowel_inventory.json` defines the
  consonants, matras, independent vowels, the halant and the cluster policy
  (`max_cluster` 3, scope `halant-clusters-v2`). `vowel_model.go` embeds it and
  `tools/bengali/vowels.py` reads it.
  - The aligner's phone table must list exactly the inventory's letters.
  - A Go test and a Python test assert that both sides agree on eligibility
    (`lang/bengali/testdata/vowel_eligibility.json`; constructed cases plus some
    real Dakshina words, see the fixture note below).
  - Before the Dakshina run, Go and Python were also checked on all 60,708
    Google and Dakshina **train** types: zero eligibility disagreements, and
    zero label disagreements on 69,757 slots. This was a one-off check; it is
    not in the JSON and no test asserts it.
- **Units.** A consonant followed by ্ has no inherent slot. A cluster's last
  member has a slot unless a matra follows. Ineligible words: clusters of four
  or more, a halant that is unterminated, leading or doubled, any member outside
  the inventory, and an inherent slot before an independent vowel.
- **Features.** The same eight features with the same definitions. `prev` and
  `next2` may now be ্. A slot consonant is never followed by ্, so `next`
  never is.
- **Aligner** (consonant phones only).
  - A non-first cluster member may match its own phone(s), the previous member's
    phone(s) (gemination), or nothing.
  - ষ after ক may also match `kh`, and ঞ after জ may also match `g`.
  - One alternative was added from Google train failures (see below).
  - Vowel handling, the ambiguity rejection and whole-type rejection are
    unchanged.
- **Model.** Retrained with the same CART at depth 10 and min-leaf 20
  (`lang/bengali/vowel_tree.json`, scope `halant-clusters-v2`). The decoder
  accepts the legacy scope `simple-whole-words-v1` with clusters disabled. This
  is used only by the evaluation build tag below.
- **Precedence is unchanged.** `schwa.bengali.model` stays at priority 70.
  Higher-priority rules still win, for example `schwa.bengali.final-cluster`
  on a word-final cluster (unit-tested on গর্ত).
- **Source.** `tools/bengali_training_data.py` also accepts the raw upstream
  `lexicon.tsv`, checked against its content digest `GOOGLE_TSV_SHA`
  (`c9db2307…f268`). `GOOGLE_SHA` still pins the original `.gz`. The manifest
  and partitions did not change and `verify()` passes. With the raw `.tsv`, the
  previous code reproduced the partitions byte-identically and the shipped tree
  identically, except for the script-hash provenance.

**"Before" is main's model, reproduced.** `go run -tags bengali_vowel_v1`
swaps in the previous tree (`lang/bengali/testdata/vowel_tree_simple_v1.json`,
from 46c2967). This evaluation-only build matched main's own runtime
(a worktree at 46c2967) on every output of all four profiles:

- Dakshina train, dev and test;
- BanglaTLit;
- the lyrics pilot.

## Alignment (Google pronunciation lexicon)

Reported as before; this never selects.

| Split | Accepted types | Slot instances | Unsupported rows | Unmatched rows | Ambiguous | Conflicting types |
|---|---:|---:|---:|---:|---:|---:|
| Train | 19,658 → 33,298 | 32,580 → 59,485 | 29,424 → 11,980 | 74 → 1,944 | 0 → 0 | 29 → 30 |
| Dev | 1,140 → 1,961 | 1,852 → 3,459 | 1,722 → 678 | 7 → 137 | 0 → 0 | 0 → 0 |
| Test | 1,100 → 1,848 | 1,793 → 3,312 | 1,591 → 636 | 7 → 112 | 0 → 0 | 2 → 3 |

Train class counts (absent / O / o) went from 20,733 / 5,462 / 6,385 to
31,318 / 11,254 / 16,913.

Of the 1,944 unmatched train rows, 1,794 contain ya-phala. Google mostly writes
ya-phala + আ, and often a ya-phala inherent vowel, as `E`
(ব্যাগ `b E g`). That is a vowel phone, so by design those types stay out of
training. The model still decides them at inference.

**Alternative added from Google train failures (before any Dakshina run):**

- The first member ক of ক্ষ may be silent, so ক্ষ is read as a single `kh`
  (ক্ষমা `kh O . m a`).
- Effect on train: unmatched rows 2,109 → 1,944, accepted types 33,144 → 33,298.

**Seen but not adopted:** an aspirated first member read as its plain phone
before a geminating member (মাধ্যম `m a d . dh O m`). It was examined only after
the Dakshina train run had started, so it was not added; see the deviations.
Allowing the plain phone for any aspirate would align at most about 270 more
train rows.

## Google classifier accuracy

Reported as before; this never selects.

| Split | Before | After | After, slots eligible before | After, newly eligible slots |
|---|---:|---:|---:|---:|
| Train | 28,117 / 32,580 (0.86301) | 49,949 / 59,485 (0.83969) | | |
| Dev | 1,585 / 1,852 (0.85583) | 2,883 / 3,459 (0.83348) | 1,552 / 1,852 (0.83801) | 1,331 / 1,607 (0.82825) |
| Test | 1,531 / 1,793 (0.85388) | 2,758 / 3,312 (0.83273) | 1,509 / 1,793 (0.84161) | 1,249 / 1,519 (0.82225) |

The two stratified "after" columns were computed once and are not in the JSON
or asserted; the other figures match `vowel_tree.json` metadata.

Sharing one tree with conjunct words costs some accuracy on the slots that were
eligible before (dev −33, test −22). The Dakshina strata below show the same
trade.

## Selection (Dakshina train only, preregistered)

The model profile is paired against main's model profile, on all 25,000 Dakshina
train words, as in the rule-bundle record. The criterion is net match-any > 0
**and** macro minCER not worse. A win is a word the new model matches and the
old one does not.

| Words | Count | Wins | Losses | Net | Changed |
|---|---:|---:|---:|---:|---:|
| All | 25,000 | 676 | 486 | **+190** | 2,149 |
| Newly eligible | 8,817 | 626 | 274 | +352 | 1,751 |
| Eligible before | 8,995 | 50 | 212 | −162 | 398 |
| Ineligible | 7,188 | 0 | 0 | 0 | 0 |

Results on train:

- Model match-any: 15,656 → 15,846 (62.62% → 63.38%).
- Model strict: 9,238 → 9,292.
- Model minCER: 0.07095 → **0.06923**.
- **The criterion passes.** Model + lexicon (in-sample): 18,373 → 18,500
  match-any. Rules-only and rules + lexicon are byte-identical.

Train examples (old model → new model, top reference):

- **Wins:**
  - প্রধানমন্ত্রী prodhanomontri → prodhanmontri;
  - আবেগপ্রবণ abegoprobon → abegprobon;
  - সিস্টেমটি shistemoti → shistemti.
- **Losses, newly eligible:**
  - শ্যামলী shyamoli → shyamli;
  - আত্মকথা attokotha → attoktha;
  - উপলব্ধি upolobdhi → uplobdhi.
- **Losses, eligible before:**
  - অপহরণ opohoron → ophoron;
  - ততটা totota → totta.

## Held-out results

Dakshina dev was run **once**, after the model was frozen in commit c2a91a6. It
was confirmation only, and is the seventh use of Bengali dev. Test and
BanglaTLit were run once after it. Aksharantar was not run.

Strict is the top-attested reference; minCER is the macro minimum CER.
Rules-only (B1) and rules + lexicon are unchanged on every split. The lexicon
does not move held-out words.

| Split | Profile | Strict | Match-any | minCER |
|---|---|---:|---:|---:|
| Dev (confirmation) | B1 | 791 | 1,405 (56.20%) | 0.08766 |
| Dev (confirmation) | Model | 909 → 914 | 1,577 → 1,590 (63.08% → 63.60%) | 0.07478 → 0.07269 |
| Test | B1 | 828 | 1,467 (58.68%) | 0.08368 |
| Test | Model | 937 → 938 | 1,615 → 1,619 (**64.60% → 64.76%**) | 0.07278 → **0.07029** |

The model + lexicon profile equals the model profile on dev and test. **Dev
confirmed:** match-any rose and minCER did not get worse.

### Coverage stratification

Model profile, before → after. Paired results are wins / losses / net.

| Split | Words | Count | B1 match-any | Model match-any | Model minCER | Paired |
|---|---|---:|---:|---:|---:|---:|
| Dev | Newly eligible | 762 | 400 | 400 → 425 | 0.09039 → 0.08112 | 50 / 25 / +25 |
| Dev | Eligible before | 1,019 | 525 | 697 → 685 | 0.06464 → 0.06644 | 10 / 22 / −12 |
| Dev | Ineligible | 719 | 480 | 480 | 0.07261 | 0 |
| Test | Newly eligible | 770 | 416 (54.03%) | 416 → 434 (56.36%) | 0.08722 → 0.07632 | 40 / 22 / +18 |
| Test | Eligible before | 987 | 532 (53.90%) | 680 → 666 (68.90% → 67.48%) | 0.06625 → 0.06846 | 6 / 20 / −14 |
| Test | Ineligible | 743 | 519 (69.85%) | 519 | 0.06648 | 0 |

The model is now eligible on 70.3% of test words, up from 39.5%.

- The gain comes from the newly eligible words.
- On the words it already covered, the retrained tree is slightly worse,
  matching the train and Google figures.
- The headline net is small: +4 test words, +13 dev words. Most of the
  improvement is in minCER.

### BanglaTLit (2,500 sentences, one reference each)

| Profile | Full sentences: match / macro CER | All-tokens-unseen (38): match / macro CER |
|---|---:|---:|
| B1 | 8 / 0.32727 (unchanged) | 1 / 0.51708 (unchanged) |
| Model | 13 → 10 / 0.30871 → **0.30938** | 2 → 2 / 0.50624 → 0.49194 |
| Lexicon | 13 / 0.29065 (unchanged) | 1 / 0.51708 (unchanged) |
| Model + lexicon | 18 → 14 / 0.27820 → 0.27835 | 2 → 2 / 0.50624 → 0.49194 |

BanglaTLit moves the other way from Dakshina. Sentence CER is slightly
**worse** for both model profiles on the full set (+0.00068 model) and better on
the 38 all-unseen sentences.

- 3,373 of the 25,680 tokens (13.1%) are newly eligible.
- 1,626 sentences contain at least one of them.

These are single informal references for whole sentences, so they cannot show
token-level accuracy.

### Lyrics pilot (agreement with unreviewed drafts, T-0068)

The [lyrics record](2026-10-05-bengali-b3-lyrics.json) was regenerated. The
references are assistant drafts, not gold, and agreement with them is not
accuracy.

| Lines | Exact (of 82) before | Exact (of 82) after | Macro line CER before | Macro line CER after |
|---|---:|---:|---:|---:|
| Model | 55 | 58 | 0.016324 | 0.015461 |
| Model + lexicon | 49 | 52 | 0.024400 | 0.023068 |

B1 (47) and lexicon (40) do not change.

## Pins

Pins moved deliberately:

- **`bnPinnedModel`** (`benchmark/bengali_model_test.go`):
  - train 9,238 / 15,656 / 3,219 changed → **9,292 / 15,846 / 5,051**;
  - dev 909 / 1,577 / 393 → **914 / 1,590 / 567**;
  - both digests are new.
- **`bnPinnedLexiconTrain`** (`benchmark/bengali_lexicon_test.go`):
  - model-lexicon 13,159 / 18,373 → **13,199 / 18,500**, with a new digest;
  - B1-lexicon 12,739 / 17,718 and 8,976 hits are unchanged.
- **Fixtures:**
  - `lang/bengali/testdata/vowel_features.json` was regenerated (525 slots, now
    including clusters);
  - `vowel_eligibility.json` is new.
  - Fixture note: both fixtures contain some real Dakshina words, mostly
    train (গর্ত, দ্বার, মধ্য, ক্ষতি, স্মরণ) and a few short held-out ones (ফল,
    আপ and উল from dev, ওক from test, the last three predating this PR).
    Fixtures train nothing, so this is not contamination. The first commit's
    fixtures used কর্ম, a Dakshina **test** word; the second commit replaced it
    with গর্ত.

These are byte-identical and their tests pass unmodified:

- `bnPinnedRules` (B1 train/dev);
- the B0 evidence digests;
- the B1 dev gate (791 / 1,405, wins 216 / losses 21);
- the Hindi frozen snapshot (`TestHindiFrozenSnapshot`, `testdata/` untouched);
- B1 and rules + lexicon outputs on every Dakshina split, BanglaTLit and the
  lyrics. The evaluator raises an error if any of them moves.

The Aksharantar/BanglaTLit figures in the
[B2 vowel external record](2026-10-04-bengali-b2-vowels-external.json) describe
the previous tree. Aksharantar was not run (preregistered). The model-digest
assertion now points to this record.

## Deviations from the preregistration

1. **জ্ঞ stays ineligible.** The prereg lists জ্ঞ as included but also requires
   every cluster member to be an existing model consonant, and ঞ is not one. The
   conservative reading keeps the letter inventory unchanged. The ঞ → `g`
   alternative is implemented but cannot fire.
2. **One aligner alternative was added** from Google train failures before any
   Dakshina run: silent ক in ক্ষ (above).
3. **The aspirate → plain first-member alternative was not adopted.** I first
   looked at this pattern closely while the Dakshina train selection run was
   already in progress, before seeing its results. A rule added then would no
   longer be "before any Dakshina run", so I left it out.
4. **"Dakshina train" means the full 25,000-word train file**, not the
   23,099-type training partition. This is the same definition the rule-bundle
   record used.
5. **The "before" profile is an evaluation build tag** (`bengali_vowel_v1`), not
   a second checkout. It was verified identical to main's runtime on every
   evaluated set.
6. **`prev` and `next2` (not `prev` and `next`) may be ্.** The prereg's item 3
   named `next`. A slot consonant is never followed by ্, so `next` cannot be;
   `next2` can. The feature definitions themselves are unchanged.

## Reproduction

```sh
python3 tools/bengali/train_vowels.py --google /path/to/lexicon.tsv --output lang/bengali/vowel_tree.json --report-test
(cd tools/bengali && python3 build_vowel_parity.py ../../lang/bengali/vowel_tree.json \
  ../../lang/bengali/testdata/vowel_features.json ../../lang/bengali/testdata/vowel_eligibility.json)
python3 tools/bengali/evaluate_vowel_conjuncts.py --stage train   # selection only
python3 tools/bengali/evaluate_vowel_conjuncts.py                 # dev once, then test and BanglaTLit
python3 tools/bengali/evaluate_lyrics.py --output docs/reviews/2026-10-05-bengali-b3-lyrics.json
make ci
```
