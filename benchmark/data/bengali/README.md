# Bengali Dakshina lexicon fixtures

Source: [Google Dakshina v1.0](https://github.com/google-research-datasets/dakshina),
by Roark et al., *Processing South Asian Languages Written in the Latin Script*
(LREC 2020). The data is licensed under
[CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), separately from
this repository's MIT-licensed code. These derived fixtures retain that license.

The three compressed CSV files preserve the original train/dev/test boundaries
and every attested spelling/count. Columns: `native,roman,attestations`.
Native keys remove Unicode format characters and normalize to NFC; duplicate
normalized native/roman pairs would have their votes summed. Roman strings are
preserved, including case. This release required no native normalization changes
or duplicate aggregation. No curation threshold has been applied.

| Split | Native word types | Reference spellings |
|---|---:|---:|
| train | 25,000 | 94,546 |
| dev | 2,500 | 9,279 |
| test | 2,500 | 9,228 |

`manifest.json` records member paths, original TSV and compressed fixture SHA-256
hashes, normalization/version metadata, and word-key hashes. Normalized word sets
are pairwise disjoint. The benchmark checks fixture hashes, counts, and split
isolation; missing or damaged data fails instead of skipping the evaluation.

Rebuild from the original archive (default `datasets/dakshina/dakshina.tar`):

```sh
python3 tools/build_bengali.py
python3 -m unittest discover -s tools -p 'build_bengali_test.py'
make test-bengali
```

Train/dev may inform spelling-style and attestation-threshold choices. Test is
reserved for aggregate evaluation, never example mining or rule tuning. These
B0 fixtures are benchmark data, not a runtime lexicon or model. B2 must add the
external-training exclusions specified in design B.4.1 (T-0057).

Strict top-1 selects the highest-attestation reference, breaking ties by lexical
order. Match-any accepts any reference, and macro minCER averages each word's
best normalized edit distance. Reports separate reference-count strata from
attestation-count strata. The o/a histogram measures whole-word support for two
B0 candidates; it does not align or identify each human-spelled inherent vowel.
The `neither` and `same-candidate` groups must be reported with the preferences.
