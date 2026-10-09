# BanglaTLit external sentence evaluation

Source: [BanglaTLit](https://github.com/farhanishmam/BanglaTLit), revision
`f780fcdd5176ddfe6364ffe7948acb6a1c2ea58e`, official `data/BanglaTLiT_test.csv`.
Attribution: Fahim et al., [BanglaTLit: A Benchmark Dataset for Back-Transliteration
of Romanized Bangla](https://aclanthology.org/2024.findings-emnlp.859/), Findings
of EMNLP 2024. Upstream repository and dataset card declare MIT; the upstream
license is preserved in LICENSE (Copyright 2024 Aplycaebous).

All 2,500 official test rows are retained. `test.csv.gz` has columns
`id,native,roman`; native text removes Unicode Cf and normalizes NFC, while IDs,
Roman text, and row order remain unchanged. There are 2,498 distinct normalized
native sentences. Original/derived hashes and transformation counts are in
`manifest.json`. No train/dev data or pretraining corpus is imported.

Rebuild using a checkout of the pinned revision:

```sh
python3 tools/build_banglatlit.py --checkout /path/to/BanglaTLit --output /tmp/banglatlit-rebuilt
go test ./benchmark -run '^TestBenchmarkBengaliBanglaTLit$' -count=1 -v
```

This is evaluation-only: never mine its examples for rules, lexicon entries,
reranking, or default selection. The B1 candidate was frozen before this dataset
was inspected/imported. The benchmark reports exact whole-sentence match and
macro character error against the supplied single Roman reference, comparing
B0 and B1. It is a transfer score with spelling, spacing, and punctuation effects,
not the preregistered Bengali word-accuracy gate.

See [repository review](../../../docs/reference/bengali-repositories.md) for the
important upstream overlap finding: every official dev/test pair appears in the
CSV named `train`. No upstream trained model is evaluated or incorporated here.
