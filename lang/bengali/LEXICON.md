# Bengali spelling lexicon

`lexicon.tsv` contains 8,980 normalized native keys and attested Roman spellings.
It is generated from the isolated Dakshina Bengali v1.0 training subset by
`tools/bengali/build_lexicon.py`. A spelling needs at least three attestations
and must be the unique highest-vote spelling for its word. Ties are excluded;
Roman case and spelling are preserved. Keys must contain only Bengali letters
and marks after Cf removal and NFC normalization.

Source: [Google Dakshina](https://github.com/google-research-datasets/dakshina)
(Roark et al., 2020). **The derived `lexicon.tsv` data is licensed under
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)**, as is the
source dataset. The repository's MIT code license does not replace this data
license. Changes from the source: normalized keys, frozen held-out exclusions,
attestation filtering, unique-winner selection and sorted TSV serialization.
Source fixture, partition manifest, builder and output hashes, plus counts,
are in `lexicon_manifest.json`.

All Google and Dakshina dev/test word types are excluded, including canonical
aliases. [Frozen source policy](../../training/data/bengali/README.md).
The builder validates final output keys against the approved train vocabulary.
CI reconstructs both the data and provenance byte-for-byte from pinned inputs.

## Lookup policy

`Options{Lexicon: true}` enables lookup. `SchwaModel` may be combined with it:
known words use attested spelling, unknown words use the requested rule/model
pipeline. The default configuration remains rules-only.

The lexicon is **default-style only**. Any of `InherentVowelA`, `LongVowels`,
`SimpleNasals`, or `KeepMedialSchwa` bypasses lookup. The option-aware provider's
miss goes directly to the pipeline, never to an older style-unaware lookup.
Debug behavior matches Hindi's existing contract: a hit short-circuits and
returns nil rule-debug information.

Lookup removes Unicode Cf and normalizes the Bengali block's canonical aliases,
split matras and combining-mark order. Other scripts, digits, punctuation and
empty keys miss. The public sentence API handles whitespace/punctuation outside
lookup and preserves them verbatim. A synthetic 5,750-case fixture checks Go
normalization against Python's Unicode NFC; every entry is also tested through
canonical aliases and the engine.

## Reproduction

```sh
python3 tools/bengali/build_lexicon.py --output /tmp/bengali-lexicon
python3 tools/bengali/build_lexicon.py --verify
python3 tools/bengali/build_lexicon_keys.py /tmp/lexicon_keys.json.gz
make test-bengali-lexicon
```

The word lexicon cannot improve excluded held-out words. Train-set gains measure
memorization/coverage, not generalization. External transfer and unseen-word
results are separated in the [evaluation report](../../docs/reviews/2026-10-04-bengali-b2-lexicon.md).
