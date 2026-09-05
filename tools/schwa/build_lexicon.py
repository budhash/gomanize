#!/usr/bin/env python3
"""Build a high-confidence romanization lexicon from Dakshina TRAIN.

For each native word, pick the most-attested Roman spelling; keep only words
whose best spelling has >= MIN_ATT total attestations (high inter-annotator
agreement = common, confident vocabulary). Emits lang/hindi/lexicon.tsv.

This is a production feature (known-word -> attested human spelling, rules as OOV
fallback). It is trained on TRAIN only; it is NOT used in the accuracy benchmark
(that would be circular — see docs/reviews/2026-09-04-h3-lexicon-layer.md).
"""
import csv, sys
from collections import defaultdict

MIN_ATT = int(sys.argv[1]) if len(sys.argv) > 1 else 4
src = 'benchmark/data/dakshina_hi.csv'
out = 'lang/hindi/lexicon.tsv'

d = defaultdict(lambda: defaultdict(int))
with open(src, encoding='utf-8') as f:
    r = csv.reader(f); next(r, None)
    for row in r:
        if len(row) < 3 or 'split=train' not in row[2]:
            continue
        att = next((int(t.split('=')[1]) for t in row[2].split() if t.startswith('attestations=')), 1)
        d[row[0]][row[1]] += att

rows = []
for native, spellings in d.items():
    best = max(spellings, key=spellings.get)
    if spellings[best] >= MIN_ATT:
        rows.append((native, best))
rows.sort()

with open(out, 'w', encoding='utf-8') as f:
    for native, roman in rows:
        f.write(f"{native}\t{roman}\n")
print(f"wrote {out}: {len(rows)} entries (min_attestations={MIN_ATT})")
