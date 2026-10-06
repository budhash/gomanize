# Bengali retrain: variant-spelling exclusion (2026-10-06)

**Status:** done. Resolves the variant-spelling leak found in the review of the
B2 training partitions (#117), per the maintainer's decision to disclose it in
the stack and retrain once afterwards.

## Problem

The comparison key (NFC + Unicode Cf removal) does not merge khanda-ta (U+09CE)
with ta + hasant, nor a malformed অ + া with আ, although each pair spells one
word. In the schema-1 partitions, 15 held-out types (23 training/held-out pairs)
had such a variant in a training pool, e.g. উত্সাহ (Google and Dakshina train)
vs উৎসাহ (Dakshina dev), অকস্মাত্ (Google train) vs অকস্মাৎ (Google test), and
আবার (train) vs অাবার (Dakshina test). The pairs were pinned and disclosed while
the stack was reviewed, so every downstream artifact kept consistent provenance.

## Change

Partition schema 2 (`tools/bengali_training_data.py`) applies design B.4.1:
training types whose `collision_key` matches any held-out type are excluded;
held-out sets are never moved. `verify` now requires zero variant collisions and
the lexicon guard rejects any colliding key. The vowel model's provenance also
hashes the shared key and loader modules.

## Results

Every dependent artifact was retrained once from the pinned sources. Dev selection
used the unchanged recorded protocols; held-out was evaluated once afterwards.

| Artifact | Before | After |
|---|---|---|
| Google train types | 55,121 | 55,108 (−13) |
| Dakshina train types | 23,117 | 23,107 (−10) |
| Google / Dakshina dev and test | 3,184 / 3,023; 2,500 / 2,500 | unchanged (inventory hashes identical) |
| Train/held-out variant pairs | 23 (pinned) | 0 (asserted) |
| Vowel model training instances | 32,588 | 32,587 |
| Vowel model tree | — | one leaf count (10,094 → 10,093); every prediction unchanged |
| Vowel model slot accuracy, dev / test | 0.86015 / 0.85443 | unchanged |
| Spelling lexicon entries | 8,980 | 8,979 (drops `আবার → abar`) |
| Native selector (cross-fit) SHA-256 | `37215f8f…d5d2` | `dfe16670…3bc5` |
| Native selector tree | — | same structure; one path's counts −1 |
| Cross-fit dev selection | 0.6; 898 / 1,553 / 0.07807792 | unchanged |
| Cross-fit held-out (frozen 0.6) | 916 / 1,576 / 0.07572118; 21 changed | unchanged |
| Rejected char reranker (#120), dev | rejected; margin 0 changed 221 | rejected; margin 0 changed 222 (CER 0.08423 → 0.08427) |
| Rejected native selector (#121) | rejected | rejected; tree byte-identical |

Engine outputs did not change in any pinned profile: rules, vowel model,
lexicon train modes, runtime reranker (Dakshina dev/test, Aksharantar,
BanglaTLit) and the lyrics pilot all reproduce their committed records; only
artifact and source hashes in those records were refreshed. Hindi frozen outputs
are unchanged.

The leak was therefore real but inconsequential for every reported number: the
affected words either never reached a model (unsupported by the vowel aligner or
identical candidates for the selector) or were overridden by rules with the same
output.

## Reproduction

```sh
python3 tools/bengali_training_data.py --google /path/to/lexicon.tsv.gz
python3 tools/bengali/train_vowels.py --google /path/to/lexicon.tsv.gz --output lang/bengali/vowel_tree.json --report-test
python3 tools/bengali/build_vowel_parity.py lang/bengali/vowel_tree.json lang/bengali/testdata/vowel_features.json
python3 tools/bengali/build_lexicon.py
python3 tools/bengali/train_reranker.py --output /tmp/rr
python3 tools/bengali/train_native_selector.py --output /tmp/ns
python3 tools/bengali/train_crossfit_selector.py --google /path/to/lexicon.tsv.gz --output /tmp/cf --dev
python3 tools/bengali/evaluate_crossfit.py --experiment /tmp/cf
python3 tools/bengali/evaluate_external.py /path/to/ben.zip [--lexicon]
python3 tools/bengali/evaluate_runtime.py /path/to/ben.zip --output docs/reviews/2026-10-05-bengali-b2-runtime.json
python3 tools/bengali/evaluate_lyrics.py --output docs/reviews/2026-10-05-bengali-b3-lyrics.json
make ci && make npm-test
```

Sources are the pinned Google Bengali lexicon (SHA-256 `1bc2edda…dd61`) and
Aksharantar `ben.zip` (`4ab6edcc…5fe7`).
