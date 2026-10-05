# Bengali B2 native-conditioned selector: no improvement at dev

T-0064 tested whether Bengali input features can improve selection between the
frozen B1 and vowel-model candidates. The fixed experiment produced no output
changes at any tested threshold, so it fails the improvement gate. Nothing is
installed in runtime and no test/external scoring was performed for this model.
The existing vowel-model held-out match-any remains 62.76%; that is the earlier
experiment's result, not a new estimate for this selector.

## Protocol fixed before dev

- Candidates: B1 and the existing vowel model, both with lexicon disabled.
- Training: normalized native types from isolated Dakshina train, additionally
  excluding **every Google training type**. Candidate-model training words must
  not make selector training labels artificially favor the vowel model.
- Keep differing candidate pairs only; one training vote per native type.
  Compare candidates by match-any, strict top-attested spelling (lexical ties),
  then minimum CER; a tied preference stays with the vowel model.
- Features: first/last one and two native code points, capped native length,
  halant/i/u/o/nasal flags, and the candidates' differing fragments, adjacent
  Roman characters, bounded length difference, and edit boundary flags.
  Features never receive references, attestations, or split labels.
- Deterministic categorical CART: Gini splits, maximum depth five, minimum leaf
  30; sorted feature/value tie-breaks. Leaf B1 fraction is an empirical training
  frequency, **not a calibrated probability**.
- B1 is chosen only strictly above one of six predeclared thresholds:
  0.5, 0.6, 0.7, 0.8, 0.9, 0.95. Identical candidates stay unchanged.
- Dev acceptance: match-any strictly higher, strict accuracy nondecreasing, and
  macro minimum CER strictly lower than the vowel-model baseline. Eligible ties
  prefer match-any, CER, strict, then the higher threshold.
- No post-dev feature, depth, minimum-leaf, or threshold tuning. Held-out
  evaluation is conditional on passing this gate, and was not reached.

## Results and interpretation

The isolated Dakshina pool contains 23,117 types. Excluding 17,501 Google training
types leaves 5,616 selector-eligible words. Only 701 have different B1/model
candidates; labels prefer B1 for 129 (18.40%) and the model for 572, including
quality ties. The resulting tree has 10 leaves. Its largest B1 fraction is
19/39 = 48.72%, below every tested threshold. This explains the all-model result
without consulting test outputs.

| Dev setting | Strict / 2,500 | Match-any / 2,500 | Mean minCER | Changed |
|---|---:|---:|---:|---:|
| Vowel model baseline | 894 (35.76%) | 1,544 (61.76%) | 0.07851746 | 0 |
| Each of 0.5, 0.6, 0.7, 0.8, 0.9, 0.95 | 894 (35.76%) | 1,544 (61.76%) | 0.07851746 | 0 |

[Aggregate report](2026-10-04-bengali-b2-native-selector.json) records all trials,
training counts, and hashes. The small [learned tree](2026-10-04-bengali-b2-native-selector-tree.json)
is retained for inspection as a research artifact; Go does not load it.

This is a negative result for this conservative feature/model/data setup. It
neither proves that source-aware ranking cannot help nor establishes a ceiling
for Bengali. Avoiding candidate-model training overlap sharply reduces the
selector sample. Future ranking work should consider generating out-of-fold
candidate predictions from the isolated training pool and expanding candidates,
with its protocol specified before dev evaluation. T-0065 tracks that work;
T-0053 remains open. B3 lyrics evaluation (T-0054) is also still outstanding.
Repeated selection on the same dev set creates selection risk; these two ranking
experiments are recorded separately and no held-out estimate is used to rescue
a failed candidate.

## Attribution, reconstruction and checks

The derived research tree is CC BY-SA 4.0, from the Google Dakshina dataset.
See the [data record](../../training/data/bengali/README.md) for source attribution,
license, pinned source hashes and the frozen normalization/exclusion policy.
The code retains the repository license. No new data dependency was introduced.

```sh
GOCACHE=/tmp/gomanize-review-go-cache python3 tools/bengali/train_native_selector.py --output /tmp/bengali-native
cmp /tmp/bengali-native/report.json docs/reviews/2026-10-04-bengali-b2-native-selector.json
cmp /tmp/bengali-native/native_selector.json docs/reviews/2026-10-04-bengali-b2-native-selector-tree.json
make test-bengali-native-selector
make ci
```

Six tests cover every frozen split exclusion, injected unauthorized rows,
Google-training exclusions, canonical normalization, native/edit features,
reference preferences, deterministic learning, minimum leaves, thresholds,
candidate cardinality and rejection of partial gains/no-ops. Negative controls
remove Google-training exclusion and allow no-op selection; the corresponding
tests must fail. Full local CI covers the existing Bengali gates and unchanged
Hindi regression snapshot. No runtime accuracy improvement is claimed.
