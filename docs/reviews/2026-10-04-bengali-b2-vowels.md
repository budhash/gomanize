# Bengali B2 vowel model (T-0061)

The opt-in `SchwaModel` option now selects a Bengali three-class inherent-vowel
model on supported simple words. B1 remains the default. No lexicon or reranker
is used in these results; both remain T-0053 work. This PR builds on T-0057's
frozen source exclusions and is intended to remain draft for review.

## Word-level results

These are pure outputs with no override or spelling-lexicon hits. Strict accuracy
uses the highest-attestation Roman reference (lexical tie-break); match-any
accepts any supplied reference. CER is mean per-word minimum-reference character
error. Dataset denominators and scoring conventions differ from Hindi's curated
benchmark; these percentages are not a direct Hindi/Bengali quality comparison.

| Frozen Dakshina split | B1 strict | Model strict | B1 match-any | Model match-any | B1 CER | Model CER |
|---|---:|---:|---:|---:|---:|---:|
| Dev, 2,500 words | 774 (30.96%) | 894 (35.76%) | 1,370 (54.80%) | 1,544 (61.76%) | 0.09149976 | 0.07851746 |
| Test, 2,500 words | 800 (32.00%) | 914 (36.56%) | 1,413 (56.52%) | 1,569 (62.76%) | 0.08772388 | 0.07642207 |

The model changes 408 outputs in each split. Test match-any gains 6.24 percentage
points, reducing remaining match-any failures by 14.35%. Test CER falls 12.88%.
The dev acceptance test requires match-any improvement, no strict regression,
and lower CER. It is a relative acceptance guard, not a preregistered target.

## Alignment and training

The aligner consumes the entire Google phoneme sequence exactly, allowing only
explicit consonant/vowel realizations and inherent-vowel choices: absent, `O`
(/ɔ/), or `o` (/o/). Syllable separators are ignored; glide-marked independent
vowels retain explicit alternatives. There are no skipped unmatched phonemes,
fuzzy edit paths, or Latin spelling-derived labels. Source spelling, pronunciation,
disambiguation label and line number remain available through T-0057's loader.

Scope is intentionally conservative: the supported consonants plus common
monophthong vowels/matras, canonical ড়, and canonical split ও-matra. Words with
conjuncts, modifiers, silent glides, diphthongs, uncommon letters, malformed vowel
attachments, or consonants directly before independent vowels use B1 throughout.
Existing higher-priority B1 keep rules remain authoritative.
The last exclusion respects the renderer's ownership of pre-independent vowels,
including the existing হও rule. Runtime and training share this eligibility
contract; synthetic fixtures check all feature values and predictions in Go.

All pronunciation variants of a word must align and agree on every vowel label.
A conflicting or failed variant rejects the whole type; duplicate pronunciations
or POS tags never multiply a word's training weight. Counts are embedded in
`lang/bengali/vowel_tree.json`:

| Measure | Train | Dev | Test |
|---|---:|---:|---:|
| Source types | 55,121 | 3,184 | 3,023 |
| Accepted types | 19,662 | 1,140 | 1,100 |
| Aligned rows before variant filtering | 19,727 | 1,140 | 1,105 |
| Unsupported rows | 29,444 | 1,722 | 1,591 |
| Unmatched rows | 74 | 7 | 7 |
| Rows with no inherent slot | 5,936 | 317 | 324 |
| Ambiguous alignment rows | 0 | 0 | 0 |
| Conflicting-variant types | 29 | 0 | 2 |
| Vowel instances retained | 32,588 | 1,852 | 1,793 |
| Three-class instance accuracy | 86.30% | 86.02% | 85.44% |

The last line measures only successfully aligned subsets, **not whole-corpus
phonological accuracy** (always predicting absent would score 64.19% on this
test subset). On Google test, absent/O/o recalls are 96.87%, 77.89%,
and 53.39% respectively; raised-vowel recognition remains substantially weaker.
Both retained classes render as `o` in the default style. Under `InherentVowelA`,
default quality renders `a`, raised quality stays `o`, and scoped B1 Open stays
`a`. The classifier never predicts Open or uses it to imitate Latin `a` names.

The deterministic categorical CART uses depth 10 and minimum leaf size 20,
eight rune-context features, and train rows only. No hyperparameter sweep was
performed. An initial dev experiment included pre-independent-vowel slots;
code review removed those slots from both training and inference to preserve
renderer semantics. The revised candidate was frozen before test reporting.
Neither test examples nor external evaluation examples were inspected or mined
for changes. The embedded provenance includes Google source, frozen partition
manifest, and training script SHA-256 hashes. Rebuilding the final model and its
synthetic parity fixture reproduced their bytes exactly.

## External transfer

See [machine-readable report](2026-10-04-bengali-b2-vowels-external.json).
Aksharantar scores aggregate all references per normalized native word; its
14,167 source rows yield 7,134 types. Unseen means absent from the union of both
frozen Google and Dakshina training vocabularies, a conservative definition that
also excludes words not successfully aligned for this particular model.

| Evaluation | B1 match-any/exact | Model match-any/exact | B1 CER | Model CER |
|---|---:|---:|---:|---:|
| Aksharantar, all 7,134 types | 1,967 (27.57%) | 2,200 (30.84%) | 0.22691465 | 0.21640005 |
| Aksharantar, 6,036 unseen types | 1,835 (30.40%) | 2,052 (34.00%) | 0.21595810 | 0.20464678 |
| BanglaTLit, 2,500 sentences | 8 (0.32%) | 12 (0.48%) | 0.33149376 | 0.31286250 |
| BanglaTLit, 38 sentences whose Bengali tokens are all unseen | 1 | 2 | 0.53180069 | 0.52095747 |

Aksharantar contains all Dakshina test words, so these are not independent
confirmations. BanglaTLit has sentence-level references: isolated unseen-token
accuracy cannot be inferred from them. Its 38-sentence subset is too small for
strong conclusions. Neither external corpus supplies training targets. Lyrics
gold does not exist yet (T-0054), and no lyrics-domain improvement is claimed.

## Reproduction and validation

```sh
python3 tools/bengali/train_vowels.py --google /path/to/lexicon.tsv.gz --output /tmp/vowel_tree.json --report-test
python3 tools/bengali/build_vowel_parity.py lang/bengali/vowel_tree.json /tmp/vowel_features.json
make test-bengali-vowels
BENGALI_MODEL_REPORT_TEST=1 go test ./benchmark -run '^TestBengaliVowelModel$' -count=1 -v
python3 tools/bengali/evaluate_external.py /path/to/ben.zip
make ci
```

`--report-test` is final reporting only. It cannot change tree fitting, which
always receives the isolated Google train rows. CI verifies alignment rejection,
variant conflicts, model/source provenance, 331 synthetic Go/Python feature and
prediction cases, three classes, canonical rune offsets, fallback, style,
metadata traces, and rule-disable controls. Malformed models fail back to B1.
A temporary Go overlay that disables the model makes both the output/style
tests and the dev accuracy gate fail. The overlay never changes repository
sources. Bengali owns its tree loader and feature extraction; the Hindi loader
and model artifacts are unchanged.

The Google pronunciation lexicon is Copyright Google 2015/2016, CC BY 4.0;
its standard Bangladeshi Bengali pronunciations do not establish coverage of
all regional varieties. Source attribution and normalization/split changes are documented in
[the training manifests](../../training/data/bengali/README.md). `lang/bengali/vowel_tree.json` is
derived model data attributed to that CC BY 4.0 source; no upstream implementation code was copied.

`make npm-test` passes (WASM build plus ESM/CommonJS smoke tests). Hindi curated
strict remains 1,147/1,330 (86.2%) before and after this change.

Full `make ci` passed, including the frozen Hindi corpus replay, accuracy
benchmarks, B1 gate, source-isolation tests and model checks. Hindi curated
match-any remains 1,236/1,330 (92.9%). Repository commit hooks pass.

## Review notes (2026-10-05)

Independent stack review. The tables above record the model as submitted.

- **Rules changed underneath.** Two B1 correctness fixes landed during review
  (canonical nukta letters in phalas; explicit final hasant). The model artifact
  is unchanged, but rule-composed outputs move by one dev word: dev B1 775 strict /
  1,371 match-any / CER 0.09137532, model 895 / 1,545 / 0.07839302 (408 changed).
  Test is unchanged. The external results JSON was regenerated: Aksharantar full
  B1/model match-any 1,968/2,201 (was 1,967/2,200), unseen 1,836/2,053; BanglaTLit
  CER moves in the sixth decimal.
- **Model outputs are now pinned.** The dev guard is relative, so a degraded model
  that still beat B1 would pass. `bnPinnedModel` pins train and dev strict,
  match-any, words changed, and a digest of every model output; a mutation that
  disables the model on longer words now fails (dev 895/1,545 → 831/1,450).
- **Parity covers the feature caps.** The fixture previously truncated the
  hand-picked edge words away, so `length` (cap 12) and `position` (cap 6) were
  never exercised; changing either cap in Go passed every test. Edge words are now
  always included, with non-NFC inputs routed through `vowelWordView` as at runtime;
  both cap mutations fail. 331 parity cases.
- **Latin-script Google rows** (4,419 train rows / 4,416 types) are rejected by the
  aligner as unsupported and are included in the 29,444 unsupported train rows
  above; none reach the tree.
- Resolved in the 2026-10-06 retrain: `training_scripts_sha256` now also covers
  the shared key/loader modules. Still tracked (T-0074): the runtime evaluates the model twice per
  pending consonant when enabled.

**Superseded figures (2026-10-06).** Training partitions were rebuilt (schema 2) and this
experiment's committed record regenerated; see the
[retrain record](2026-10-06-bengali-retrain.md) for current counts.
