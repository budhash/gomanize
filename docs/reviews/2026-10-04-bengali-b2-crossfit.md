# Bengali B2 expanded candidates and cross-fitting

T-0065 protocol, fixed before dev evaluation:

- Use only the frozen, isolated Google and Dakshina training partitions.
- Assign normalized native word types to five folds using the first eight bytes
  of SHA-256(`gomanize-bengali-crossfit-v1` + NUL + native), big endian, modulo 5.
  The same word shares a fold across sources and pronunciation variants.
- Fit a Google vowel tree excluding the entire prediction fold, with existing
  alignment, maximum depth 10 and minimum leaf 20. Each Dakshina training word
  receives candidate-model predictions from the tree excluding its fold.
- Candidates, in order: fold model, B1, single-slot absent/retained flips at the
  first eight inherent-vowel slots. Keep higher-priority B1 rules authoritative,
  use default colloquial rendering, deduplicate strings, and disable lexicon.
  Unsupported words receive no slot expansions. No Latin character substitution,
  length-style variant, or reference-derived spelling is added.
- Train one binary preference tree over distinct alternative/model pairs:
  maximum depth eight, minimum leaf 30, native/edit features from T-0064.
  Each distinct pair gets one vote; words with more candidates contribute more
  pairs. Labels compare match-any, strict, then minimum CER; ties prefer model.
- Freeze the training artifacts before evaluating dev with the existing full
  vowel model. Choose the alternative with largest leaf preference fraction
  strictly above one of 0.5/0.6/0.7/0.8/0.9/0.95; ties retain candidate order.
- Accept only strictly higher dev match-any, nondecreasing strict, and strictly
  lower macro minimum CER. Passing ties prefer match-any, CER, strict, then
  higher threshold. No post-dev parameter or feature tuning in this experiment.
- Held-out evaluation is conditional on passing dev. Candidate coverage/oracle
  statistics describe an upper bound, not actual prediction accuracy.

## Result: dev-selected candidate improves held-out metrics

The frozen threshold is **0.6**, selected on dev before any held-out scoring.
The offline selector improves all three aggregate held-out metrics. This is a
small point-estimate gain, not evidence of a statistically established advantage.
There was no test-set threshold search or parameter update.

| Split / method | Strict | Match-any | Mean minCER | Changed |
|---|---:|---:|---:|---:|
| Dev vowel model | 894/2,500 (35.76%) | 1,544/2,500 (61.76%) | 0.07851746 | — |
| Dev selected 0.6 | 897/2,500 (35.88%) | 1,552/2,500 (62.08%) | 0.07820237 | 21 |
| Test vowel model | 914/2,500 (36.56%) | 1,569/2,500 (62.76%) | 0.07642207 | — |
| Test frozen 0.6 | 916/2,500 (36.64%) | 1,576/2,500 (63.04%) | 0.07572118 | 21 |

The dev grid remains fully visible:

| Threshold | Strict | Match-any | Mean minCER | Changed | Pass dev gate |
|---|---:|---:|---:|---:|---|
| 0.5 | 896 | 1,545 | 0.07852696 | 61 | No: CER worsens |
| 0.6 | 897 | 1,552 | 0.07820237 | 21 | Yes: selected |
| 0.7 | 895 | 1,549 | 0.07840237 | 14 | Yes |
| 0.8 | 896 | 1,548 | 0.07830665 | 4 | Yes |
| 0.9 | 894 | 1,544 | 0.07851746 | 0 | No: no improvement |
| 0.95 | 894 | 1,544 | 0.07851746 | 0 | No: no improvement |

All 23,117 isolated Dakshina training types are eligible again. Cross-fitted
candidate generation yields 8,243 words with alternatives and 13,954 distinct
alternative/model pairs; labels favor the alternative in 1,263 pairs. Fold
models fit 25,964–26,206 Google vowel slots apiece. Each fold's fitted word set
is checked against its prediction word set; all pronunciation variants share
the normalized word's fold.

Candidate match-any coverage is 1,636/2,500 on dev and 1,662/2,500 on test. These
are reference-assisted oracle bounds (65.44% and 66.48%), not achieved accuracy.
The selector reaches 1,552 and 1,576 respectively. This comparison describes the
remaining choice problem; it is not a reason to tune against test.

## Scope and limitations

The [selector artifact](2026-10-04-bengali-b2-crossfit-selector.json),
[training/dev report](2026-10-04-bengali-b2-crossfit.json), and
[frozen held-out report](2026-10-04-bengali-b2-crossfit-heldout.json) are research
artifacts only. The reports include source, model, split, fold-key and artifact
hashes. **The runtime still uses the existing 62.76% vowel-model configuration.**
No new selector is installed and Hindi behavior is unchanged.

The word-level helper injects only supplied vowel decisions into the existing
Bengali model rule. It preserves rule priority, pending-slot ownership and
option checks. Every rendered training/dev/test word also undergoes a full-model
parity check against the existing public API, independently of reference scores.
The helper is an offline command, not a public model-loading API.

The fold models use 80% of aligned Google training words while dev/test use the
full model, so some training/deployment distribution shift remains. Leaf fractions
are empirical preferences, not calibrated probabilities. Pair weighting gives
words with more alternatives more influence. The experiment handles the existing
simple-word model scope and at most eight editable slots; it does not expand
conjunct coverage or validate lyrics/sentences. Repeated work on the same dev set
also carries selection risk; preceding rejected experiments remain recorded.

T-0065 closes the fixed cross-fitting experiment. T-0066 tracks native-aware
runtime integration, synthetic parity, style/fallback behavior, and external
validation before promotion. Overall B2 (T-0053) and B3 lyrics evaluation
(T-0054) remain open. Reverse transliteration remains separate future work.

## Attribution and reproduction

The selector is derived from Google Dakshina (CC BY-SA 4.0) and Google Bengali
pronunciation data (CC BY 4.0). Credit both datasets; the derived selector is
CC BY-SA 4.0. Source attribution, license links and pinned hashes are in the
[training data record](../../training/data/bengali/README.md). Code retains the
repository license. The five intermediate vowel trees are reproducible local
artifacts, not additional embedded runtime models.

```sh
python3 tools/bengali/train_crossfit_selector.py \
  --google /path/to/pinned/lexicon.tsv.gz --output /tmp/bengali-crossfit --dev
python3 tools/bengali/evaluate_crossfit.py \
  --experiment /tmp/bengali-crossfit
cmp /tmp/bengali-crossfit/selector.json docs/reviews/2026-10-04-bengali-b2-crossfit-selector.json
cmp /tmp/bengali-crossfit/report.json docs/reviews/2026-10-04-bengali-b2-crossfit.json
cmp /tmp/bengali-crossfit/heldout.json docs/reviews/2026-10-04-bengali-b2-crossfit-heldout.json
make test-bengali-crossfit
make ci
```

Training writes frozen artifacts before dev evaluation. The held-out evaluator
requires a passing, unchanged dev selection and validates artifact/source hashes;
it exposes no test-time threshold parameter. A separate reconstruction reproduced
the selector, five fold trees and dev report byte for byte at the time (see the
2026-10-05 review notes for the regenerated dev report).

Five Python tests cover fold/source isolation, bounded one-slot expansion,
training preferences, stable candidate selection and the frozen held-out gate.
Two Go tests cover supplied labels, B1 fallback, option gating, explicit matras,
higher-priority final-h retention and malformed requests. Negative controls
remove fold exclusion, bypass the artifact gate, and elevate injection above the
final-h rule: all three intended guards must fail. Full local CI includes the
Hindi frozen snapshot and prior Bengali gates.

## Review notes (2026-10-05)

Independent stack review. The tables above record the experiment as submitted.

- **Rebuilt on the current engine.** After the B1 review fixes (canonical nukta
  letters; explicit final hasant) the selector, all five fold trees, training
  counts and alignment are **byte-identical** (SHA-256 `37215f8f…d5d2`, the
  artifact #123 ships). Dev rows each move by one word: model 895 / 1,545 / CER
  0.07839302; selected 0.6 898 / 1,553 / 0.07807792 (21 changed); 0.5 897 / 1,546
  (still fails on CER); 0.7 896 / 1,550; 0.8 897 / 1,549; 0.9 and 0.95 unchanged
  from the model. Oracle coverage on dev 1,637 (65.48%). **0.6 is still selected and
  the held-out result is unchanged** (916 strict / 1,576 match-any / 0.07572118,
  21 changed, 8 wins / 1 loss). The dev report and held-out record were regenerated.
- **Records are now asserted in CI.** `test_committed_records_reproduce` re-evaluates
  the committed selector on dev and held-out and requires both committed records to
  match (assert only; it never re-selects); the stale dev record failed it. A full
  retrain needs the raw Google lexicon and remains an offline reproduction step.
- **Protocol timing.** Protocol, code, selector, dev and held-out records landed in
  one commit, so git cannot show the grid, edit cap and held-out-once rule came
  first. The evaluator exposes no threshold parameter and requires the frozen dev
  choice and artifact hash.
- **Variant spellings.** Folds use `native_key`, so 37 train words share a fold
  model with a khanda-ta/অা variant from Google train; rerunning with folds keyed
  by `collision_key` gives an identical selector and results. Selector features
  are not invariant to ৎ vs ত্, with no observed choice divergence. None of the 21
  changed dev or test outputs is a disclosed collision word.

- **Retrained 2026-10-06** after excluding variant spellings of held-out words from training: new selector SHA-256 `25278160…0961`; dev selection (0.6) and held-out results unchanged. See the [retrain record](2026-10-06-bengali-retrain.md).
