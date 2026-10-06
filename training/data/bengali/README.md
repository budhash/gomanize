# Bengali B2 frozen training partitions

These are **word-key manifests, not models or pronunciation data**. No runtime
output changes. They gate T-0053 training before alignment or feature extraction.
`tools/bengali_training_data.py` verifies the frozen inventories, selects only
training rows, and checks final spelling-lexicon keys. Future trainers must use
`training_rows`, record this manifest's SHA-256 with their artifacts, and call
`assert_lexicon_isolation` on emitted spelling entries.

| Source | Train types | Dev types | Test types |
|---|---:|---:|---:|
| Google pronunciation lexicon | 55,121 | 3,184 | 3,023 |
| Dakshina spelling references | 23,117 | 2,500 | 2,500 |

Comparison keys remove Unicode format characters (`Cf`) then normalize to NFC.
This handles canonical nukta aliases and split matras, without merging words by
Roman spelling. Original source rows remain in the pinned upstream files; the
Google loader returns original spelling, phonemes, optional disambiguation label (unchanged),
and source line number. The Dakshina loader preserves all reference spellings
and their attestation counts. No alignment labels have been generated.

Google starts with 65,037 rows / 64,968 original types / 64,958 normalized types;
15 row spellings change. Excluding 3,630 Dakshina dev/test types leaves 61,328
Google types. Assignment uses SHA-256 of UTF-8 `seed + NUL + key`, first eight
bytes as an unsigned big-endian integer modulo 10,000: below 9,000 train, below
9,500 dev, otherwise test. Seed: `gomanize-bengali-b2-v1`. Every pronunciation of
a normalized word follows the same assignment, independent of source row order.

The union of **both sources' dev/test types** is protected from both training
sources. This also removes 1,883 of Dakshina's 25,000 training types, preventing
its future spelling lexicon from covering Google model evaluation words.
Development references may be used for model selection, never training. Test
references remain final evaluation only. Frozen inventories are checksum-pinned
in the loader; editing a manifest and recomputing its checksum cannot silently
remove protected Google words. Training Google rows additionally requires exact
reconstruction from the pinned raw source. CI checks inventories offline.

### Known limitations (2026-10-05 review)

- **Variant-spelling collisions.** The comparison key does not merge khanda-ta
  (U+09CE) with ta + hasant, nor a malformed অ + া with আ. Each pair spells the
  same word; the engine renders the khanda-ta pairs identically, while the
  malformed অ + া currently renders literally (অাবার → `oabar`, আবার → `abar`). As a result, **15 held-out types** (in 23
  training/held-out pairs) have a variant spelling in a training pool, e.g.
  উত্সাহ (Google and Dakshina train) vs উৎসাহ (Dakshina dev), and অকস্মাত্
  (Google train) vs অকস্মাৎ (Google test). The partitions are frozen and every
  downstream artifact was built on them, so the pairs are pinned and disclosed in
  `KNOWN_VARIANT_COLLISIONS` rather than silently re-split: `verify` fails on any
  new collision, and on removing one without updating the list. A later rebuild
  will drop the colliding training types (exclusion via `collision_key`) and
  retrain dependent artifacts once, reporting old and new numbers.
- **Latin-script Google entries.** Google contains lowercase Latin loan spellings
  with no Bengali-block rune (e.g. `abdomen`): 4,416 train / 264 dev / 226 test
  types, included in the counts above. They are pinned in `NON_BENGALI_TYPES`;
  aligners must filter them explicitly.
- **Manifest counts.** `verify` recomputes the Dakshina counts; the Google counts
  and `unicode_version` require the raw source or record the build environment and
  are informational.

## Sources and attribution

- Google Bengali pronunciation lexicon: Copyright Google 2015/2016,
  [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
  [Original dataset](https://github.com/googlei18n/language-resources/tree/master/bn),
  accessed through [bengali-romanizer at ad6ed111](https://github.com/apakabarlabs/bengali-romanizer/tree/ad6ed11183fa75bc778fb380f5a44cf158dbdeb5).
  Input: `bengali_romanizer/data/lexicon.tsv.gz`; SHA-256 in `manifest.json`.
  These derived manifests normalize/exclude/split its word list; they do not
  copy the wrapper's output convention.
- [Dakshina](https://github.com/google-research-datasets/dakshina), Google, 2020,
  [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
  Derived Dakshina word manifests retain CC BY-SA 4.0. Source version, member
  hashes, attribution, and normalization are in
  [the Bengali benchmark fixtures](../../../benchmark/data/bengali/README.md).
  Fixture hashes are additionally pinned here. All spelling training rows come
  from its train split, filtered by the protected word union.

## Reproduce and verify

From the repository root:

```sh
python3 tools/bengali_training_data.py --google /path/to/lexicon.tsv.gz --output /tmp/bengali-rebuild
make test-bengali-data
python3 tools/audit_bengali_training_overlap.py /path/to/ben.zip
```

The builder writes deterministic gzip key lists and a source/count/hash manifest.
Rebuilding into a fresh directory reproduced all committed partition bytes.
The loader fails on unknown sources, changed source hashes, noncanonical keys,
changed inventories, split overlap, or unauthorized lexicon keys. Negative tests
inject a held-out word and recompute its hash: semantic validation still fails.
Python Unicode version is recorded for reproducibility (13.0.0 for this build).

## External evaluation overlap

`external-overlap.json` records source checksums and exact counting conventions.
Aksharantar's 14,167 test rows contain 7,134 normalized types: 1,098 intersect the
combined training vocabulary and 6,036 are unseen. It contains **all 2,500
Dakshina test words**, so these corpora are not independent benchmarks.

BanglaTLit's 2,500 test sentences contain 4,200 distinct Bengali letter/mark-run
tokens: 2,219 intersect training and 1,981 are unseen. This is a descriptive
word-overlap audit, not a claim about sentence accuracy or the engine tokenizer.
Neither external corpus supplies training answers. B2 evaluations must report
full-corpus and unseen-type performance separately. Bengali lyrics gold does
not exist yet (T-0054): its overlap is unknown, not zero, and must be measured
when imported. Ambiguous-alignment counts and model accuracy belong to T-0053;
this data preparation makes no accuracy claim.
