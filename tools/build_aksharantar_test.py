#!/usr/bin/env python3
"""Convert the Aksharantar Hindi TEST set to benchmark CSV format.

Source: https://huggingface.co/datasets/ai4bharat/Aksharantar (hin.zip →
hin_test.json). License: CC-BY 4.0 (manually annotated by native speakers via
the Karya platform; Madhani et al., Findings of EMNLP 2023, arXiv:2205.03018).

Output: benchmark/data/aksharantar_test_hi.csv with notes carrying the slice
label (source=AK-Freq | AK-NEF | AK-NEI | Dakshina). Multiple rows per native
word are preserved — they are the attested variant set for match-any scoring.

Usage: python3 tools/build_aksharantar_test.py <hin_test.json>
"""
import json
import sys

SRC = sys.argv[1]
OUT = 'benchmark/data/aksharantar_test_hi.csv'

rows = []
with open(SRC, encoding='utf-8') as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        r = json.loads(line)
        rows.append((r['native word'], r['english word'], r['source']))

rows.sort()
with open(OUT, 'w', encoding='utf-8') as f:
    f.write("native,roman,notes\n")
    for native, roman, source in rows:
        f.write(f"{native},{roman},source={source}\n")
print(f"wrote {OUT}: {len(rows)} rows")
