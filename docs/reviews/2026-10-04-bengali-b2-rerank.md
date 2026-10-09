# Bengali B2 character reranker: rejected at dev

T-0063 evaluated a character-only selector between B1 and the opt-in vowel
model. It fails the dev gate. No reranker is installed, no Go behavior changes,
and no held-out or external evaluation was run for this candidate. T-0053
remains open; native-conditioned candidate selection is tracked separately.
The planned held-out validation is conditional on passing dev, and was therefore
not reached. This task closes the experiment, not the overall reranking work.

## Method and provenance

Train only on the frozen, isolated Dakshina training pool: 87,391 spelling rows,
122,215 attestations. Lowercase ASCII letters; count boundary-marked 1–4 grams,
weight by attestations, prune counts below three. All source rows pass the
alphabet filter. The resulting 36,975 grams use backoff 0.4 and mean log score,
following the existing Hindi scorer's approach. The model candidate wins ties;
B1 replaces it only if its score exceeds the model score plus the margin.
Lexicon lookup is disabled for both candidates.

The seven-margin grid below was fixed before running the experiment. Selection
requires strictly higher match-any, nondecreasing strict accuracy, and strictly
lower macro minimum CER on 2,500 dev words. Eligible candidates are ordered by
match-any, CER, strict, then smallest margin. Dev is for selection; its results
are not a new generalization estimate. No finer grid was tried after rejection.

[Machine-readable report](2026-10-04-bengali-b2-rerank.json) records the grid,
counts, source/model hashes, trainer hash, and rejected n-gram artifact hash.
Source partitioning and normalized-word exclusions are enforced by the existing
[training loader](../../tools/bengali_training_data.py). Dakshina-derived n-grams
retain CC BY-SA 4.0 attribution to the Google Dakshina dataset; see the
[training data record](../../training/data/bengali/README.md). The rejected
n-gram TSV is reproducible locally and is not embedded or checked in.

## Dev results

Strict compares the highest-attestation reference (lexical tie-break); match-any
accepts any supplied spelling. CER is the mean, over words, of minimum edit
distance divided by reference length across references. Counts are out of 2,500.

| Selection | Strict | Match-any | Mean minCER | Changed from model |
|---|---:|---:|---:|---:|
| Vowel model baseline | 894 | 1,544 (61.76%) | 0.07851746 | 0 |
| Margin 0 | 842 | 1,474 (58.96%) | 0.08435596 | 221 |
| Margin 0.05 | 858 | 1,494 | 0.08310636 | 190 |
| Margin 0.1 | 863 | 1,502 | 0.08269083 | 164 |
| Margin 0.2 | 874 | 1,518 | 0.08143273 | 122 |
| Margin 0.4 | 889 | 1,536 | 0.07965509 | 61 |
| Margin 0.8 | 893 | 1,545 (61.80%) | 0.07852540 | 13 |
| Margin 1.6 | 894 | 1,544 | 0.07851746 | 0 |

Margin 0.8's single match-any gain is not an accepted improvement: strict loses
one word and CER worsens. Margin 1.6 makes no changes. Character plausibility
alone is insufficient for this candidate pair under the tested setup; this does
not rule out native-conditioned ranking or other candidate generation methods.
Those need a separate, train-only experiment with a fixed dev selection policy.

The previous held-out model result remains **62.76% match-any**, versus B1's
56.52%. These numbers come from the earlier vowel-model experiment, not a fresh
reranker test. Hindi remains unchanged, including its frozen output snapshot.

## Reproduction and checks

```sh
python3 tools/bengali/train_reranker.py --output /tmp/bengali-rerank
cmp /tmp/bengali-rerank/reranker_manifest.json docs/reviews/2026-10-04-bengali-b2-rerank.json
make test-bengali-rerank
make ci
```

Tests cover weighted counts, boundaries, pruning, case handling, hand-computed
backoff scoring, candidate ties and margins, train-only artifact reconstruction,
and rejection of no-op or partial metric gains. The negative control removes
the strict/CER requirements: the rejection tests must fail, including selection
of the recorded margin 0.8 trial. No runtime accuracy increase is claimed.

## Review notes (2026-10-05)

Independent stack review. The table above records the experiment as submitted.

- **Rerun after the B1 review fixes.** Rule-composed candidates moved by one dev
  word (canonical nukta letters; explicit final hasant). The regenerated record
  shifts every row by exactly +1: model baseline 895 strict / 1,545 match-any / CER
  0.07839302; margin 0 843 / 1,475; margin 0.8 894 / 1,546 / 0.07840095 (still +1
  match-any for −1 strict and worse CER). `changed_from_model` counts and the
  n-gram table are unchanged. **The rejection holds**; `selected` stays null.
- **The record is now asserted.** `tools/bengali_reranker_test.py` reruns the dev
  grid against the current engine and requires the committed baseline and trials to
  match (source hashes are provenance and not compared); the stale record failed it.
- **Protocol timing.** The margin grid and guard landed in the same commit as the
  results, so git cannot show the protocol came first; the guard is the same
  relative dev guard the vowel model used, and no trial passed it.
- The disclosed variant-spelling collisions include four dakshina-train/dev pairs;
  B1 and the model agree on all four dev words, so the reranker never touched them.

**Superseded figures (2026-10-06).** Training partitions were rebuilt (schema 2) and this
experiment's committed record regenerated; see the
[retrain record](2026-10-06-bengali-retrain.md) for current counts.
