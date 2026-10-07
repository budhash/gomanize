# Bengali retrain: variant-spelling exclusion (2026-10-06)

**Status:** done. Resolves the variant-spelling leak found in the review of the
B2 training partitions (#117), per the maintainer's decision to disclose it in
the stack and retrain once afterwards.

## Problem

The comparison key (NFC + Unicode Cf removal) does not merge khanda-ta spellings
(ৎ, ত + hasant, word-final bare ত) or a malformed অ + া with আ, although each set
spells one word. In the schema-1 partitions, training pools contained such
variants of held-out types, e.g. উত্সাহ (train) vs উৎসাহ (Dakshina dev),
অর্থাত (train) vs অর্থাৎ (Dakshina dev), and আবার (train) vs অাবার (Dakshina
test). The ta + hasant and malformed-আ cases (15 held-out types, 23 pairs) were
pinned and disclosed during the stack review; the review of this PR found the
word-final bare ত form as well, and the maintainer chose to fold it too.

## Change

Partition schema 2 (`tools/bengali_training_data.py`) applies design B.4.1:
training types whose `collision_key` matches any held-out type are excluded;
held-out sets are never moved. `collision_key` folds ৎ, ত + hasant and
word-final ত to one form and malformed অ + া to আ. `verify` requires zero such
collisions and the lexicon guard rejects any colliding key. The vowel model's
provenance also hashes the shared key and loader modules.

Other orthographic variants (ি/ী and ু/ূ, ন/ণ, জ/য, শ/ষ/স, ং/ঙ্, ৃ/্রি) are
distinct spellings and are **not** merged: design B.4.1 forbids merging
spellings by romanized output. The review measured roughly 430, 130, 27, 145, 16
and 2 training/held-out pairs of these kinds; they are disclosed in the
[training data record](../../training/data/bengali/README.md).

## Results

Every dependent artifact was retrained once from the pinned sources with the
unchanged recorded protocols; held-out was evaluated once after dev selection.

| Artifact | Before (schema 1) | After (schema 2) |
|---|---|---|
| Google train types | 55,121 | 55,097 (−24) |
| Dakshina train types | 23,117 | 23,099 (−18) |
| Google / Dakshina dev and test | 3,184 / 3,023; 2,500 / 2,500 | unchanged (inventory hashes identical) |
| Train/held-out variant pairs (folded classes) | 23 pinned + word-final ত | 0 (asserted) |
| Vowel model training instances | 32,588 | 32,580 |
| Vowel model slot accuracy, Google dev | 1,593 / 1,852 (0.86015) | 1,585 / 1,852 (0.85583) |
| Vowel model slot accuracy, Google test | 1,532 / 1,793 (0.85443) | 1,531 / 1,793 (0.85388) |
| Vowel model on Dakshina (pinned outputs) | dev 895 / 1,545; test 914 / 1,569 | unchanged |
| Spelling lexicon entries | 8,980 | 8,976 (drops `আবার`, `অর্থাত`, `যাবত`, `বিদ্যুৎ`) |
| Native selector (cross-fit) SHA-256 | `37215f8f…d5d2` | `25278160…0961` |
| Cross-fit training pairs / prefer-alternative | 13,954 / 1,263 | 13,946 / 1,259 |
| Cross-fit dev selection | 0.6; 898 / 1,553 / 0.07807792 | unchanged |
| Cross-fit held-out (frozen 0.6) | 916 / 1,576 / 0.07572118; 21 changed | unchanged |
| Runtime reranker: Dakshina dev/test, Aksharantar, BanglaTLit pure | as recorded | unchanged |
| BanglaTLit with lexicon, macro CER (lexicon / model+lexicon / rerank+lexicon) | 0.29401287 / 0.28142436 / 0.28144204 | 0.29402398 / 0.28143547 / 0.28145315 |
| Lyrics pilot (all profiles) | as recorded | unchanged |
| Rejected char reranker (#120), dev | rejected; margin 0 changed 221 | rejected; margin 0 changed 222 (CER 0.08423 → 0.08427) |
| Rejected native selector (#121) | rejected; max B1 leaf 0.487 | rejected; max B1 leaf 0.487 |
| Hindi frozen outputs | — | unchanged |

The removed words reached few learned decisions: Dakshina-scored outputs did not
change for any profile. The vowel model's Google slot accuracy dips slightly
(−8 dev, −1 test slots) because removing training words shifts tree counts; the
lexicon loses four entries, which slightly worsens BanglaTLit lexicon-profile CER
(about +1.1e-5) and in-sample train lexicon matches (B1-lexicon 17,401 → 17,400,
model-lexicon 18,114 → 18,112 match-any).
These are the honest cost of the stricter isolation.

## Reproduction

```sh
python3 tools/bengali_training_data.py --google /path/to/lexicon.tsv.gz
python3 tools/bengali/train_vowels.py --google /path/to/lexicon.tsv.gz --output lang/bengali/vowel_tree.json --report-test
python3 tools/bengali/build_vowel_parity.py lang/bengali/vowel_tree.json lang/bengali/testdata/vowel_features.json
python3 tools/bengali/build_lexicon.py
python3 tools/bengali/train_reranker.py --output /tmp/rr
python3 tools/bengali/train_native_selector.py --output /tmp/ns
(cd tools/bengali && python3 train_crossfit_selector.py --google /path/to/lexicon.tsv.gz --output /tmp/cf --dev \
  && python3 evaluate_crossfit.py --experiment /tmp/cf)
python3 tools/bengali/build_selector_parity.py
python3 tools/bengali/evaluate_external.py /path/to/ben.zip [--lexicon]
python3 tools/bengali/evaluate_runtime.py /path/to/ben.zip --output docs/reviews/2026-10-05-bengali-b2-runtime.json
python3 tools/bengali/evaluate_lyrics.py --output docs/reviews/2026-10-05-bengali-b3-lyrics.json
python3 tools/audit_bengali_training_overlap.py /path/to/ben.zip > training/data/bengali/external-overlap.json
make ci && make npm-test
```

Sources are the pinned Google Bengali lexicon (SHA-256 `1bc2edda…dd61`) and
Aksharantar `ben.zip` (`4ab6edcc…5fe7`). The full chain takes about two minutes.
