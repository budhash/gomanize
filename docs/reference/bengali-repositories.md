# Viable Bengali transliteration references

Inspected 2026-10-04 from user-supplied repositories. All are retained as viable
references for the roles below. Output-style differences do not make their
architecture irrelevant. Code was inspected in temporary clones; no upstream
runtime, pretrained model, or GPL implementation was incorporated.

| Repository / pinned revision | Useful role | License / data boundary |
|---|---|---|
| [sagorbrur/bntranslit](https://github.com/sagorbrur/bntranslit/tree/a109eae4f8cc685616536d8711730d6a3fb6e67a) | Roman → Bengali word candidates; attention/LSTM plumbing and top-k ranking/evaluation reference for the reverse project | MIT code; uses Dakshina, not a new independent dataset. Reported top-k reverse results are not comparable to our forward top-1 scores. |
| [apakabarlabs/bengali-romanizer](https://github.com/apakabarlabs/bengali-romanizer/tree/ad6ed11183fa75bc778fb380f5a44cf158dbdeb5) | **Viable architectural reference:** separates phoneme transcription from spelling output, caches lexicon loading, preserves multiple readings, handles uncertainty explicitly, and separates text scanning/digits from word lookup | MIT code; bundled Google pronunciation lexicon is separately CC BY 4.0. Its learner-oriented ô/æ/retroflex output and Bengali fallback are policy choices; we can learn from the plumbing without adopting those choices. |
| [farhanishmam/BanglaTLit](https://github.com/farhanishmam/BanglaTLit/tree/f780fcdd5176ddfe6364ffe7948acb6a1c2ea58e) | Sentence-level social-text evaluation now; contextual reverse transliteration architecture later | Repository MIT license and [dataset card](https://huggingface.co/datasets/aplycaebous/BanglaTLit/blob/main/README.md) declare MIT. Import only pinned official test data for evaluation; do not infer training isolation from filenames. |
| [translatorswb/Romabangla](https://github.com/translatorswb/Romabangla/tree/1411d6df055aac3b08dd628014d69f016ecf29cd) | Small deterministic Latin → Bengali parser, longest-token/context decisions, and limitations as a reverse baseline | GPL-3.0-or-later in source headers. Retain as a reference/separate comparator; no code copied into this MIT implementation. No substantial paired benchmark found. |

## bengali-romanizer: plumbing worth carrying forward

The [implementation](https://github.com/apakabarlabs/bengali-romanizer/blob/ad6ed11183fa75bc778fb380f5a44cf158dbdeb5/bengali_romanizer/romanizer.py)
turns phoneme tokens into display spelling in a distinct function, builds a
word-to-set-of-readings map, sorts readings deterministically, and caches the
resource load. It emits a dictionary result only when unambiguous; unknown or
ambiguous words remain native script. Its tests distinguish verified words,
text boundaries, and uncertainty behavior.

For gomanize, the useful separation is pronunciation representation → scheme
rendering, with ambiguity preserved rather than silently selecting one reading.
Our existing rules can remain an OOV fallback, and a future API can expose
confidence/candidates without adopting this repository's native-script fallback.
This is directly relevant to Bengali B2 and to reverse candidate ranking.

The bundled lexicon has 65,037 rows and 64,958 normalized native types, including
70 types with multiple distinct phoneme readings. It overlaps Dakshina train by
19,384 types, dev by 1,803, and test by 1,827. Removing dev/test types leaves
61,328 types **before** an external split or any other exclusions. It is not
safe to treat the entire dictionary as a new held-out benchmark or training
pool. T-0057 built the exclusion/split pipeline
([training partitions](../../training/data/bengali/README.md)).
Compressed SHA-256 matches upstream's declared
`1bc2edda15da62bd4ef8576114e391c9a89ad8971f1eef3635e8fee0d7c1dd61`.

## BanglaTLit: useful data, with an audited split caveat

At this pinned revision:

- `BanglaTLit_train.csv`: 245,727 rows; 202,863 empty Bengali targets;
  42,835 rows contain Bengali characters. This looks like a pool, not an already
  isolated supervised training split. We make no claim about the paper's actual
  training run from these file observations alone.
- Official validation: 1,500 paired rows / 1,500 distinct normalized native
  sentences. Official test: 2,500 paired rows / 2,498 distinct native sentences.
- Every validation and test **pair** also occurs in the large `train` CSV.
  Validation/test share four native sentence types, but no exact native/Roman
  pairs. A reverse-training importer must exclude by both normalized native and
  Roman identity, not merely row ID.
- The notebook refers to separately hosted Kaggle train/val/test filenames;
  inspecting that reference is not proof that the checked-in CSVs are disjoint.

`benchmark/data/banglatlit` now contains all 2,500 official test rows, with
source hash, revision, MIT attribution, deterministic importer, and an external
forward-comparison benchmark. No examples from it were used to tune B1. Keep it
out of training, exception mining, and default selection. It is an independent
source for this rules-only engine, not a claim that every token is unseen in
Dakshina. Sentence spelling/spacing/style variation makes exact match a demanding
transfer measure, not the Bengali word-accuracy gate.

The larger pool and underlying source collections remain out of the repo. Any
future training use must audit each source's terms and perform explicit overlap
exclusion, rather than assuming the top-level label resolves all upstream data.

## Reverse-direction roadmap

The user identified Roman → actual Hindi/Bengali as the next goal. Track it as a
separate feature (F-0012). Romanization loses distinctions and admits many spellings;
reverse conversion therefore needs an explicit single-result versus candidate
API, language selection, normalization, unknown-token behavior, and independent
native-script evaluation. bntranslit's top-k interface, bengali-romanizer's
ambiguity handling, BanglaTLit's sentence task, and Romabangla's simple parser
are complementary references for that design. No reverse API has been added yet.

Machine-readable evidence is in [bengali-source-audit.json](bengali-source-audit.json).
Reproduce it from the four pinned checkouts under one parent directory with
`python3 tools/audit_bengali_sources.py --checkouts /path/to/checkouts`.
The audit reads text/metadata only and prints no corpus examples.
