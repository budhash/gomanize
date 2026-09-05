#!/usr/bin/env python3
"""Build a high-confidence romanization lexicon from Dakshina TRAIN.

Sources ONLY the train split — dev/test natives are disjoint and must stay out
of the lexicon so the held-out evaluations remain uncontaminated.

Selection policy (T-0020, frequency-ranked expansion):
  A word enters the lexicon with its most-attested Roman spelling when either
  a) its best spelling has >= HIGH_ATT total attestations (the original
     high-confidence core, kept regardless of frequency), or
  b) it is in the top-N frequency list (benchmark/data/freq_hi.csv, Shabd CC0)
     AND its best spelling has >= MIN_ATT attestations AND is a strict winner
     (no tie with another spelling — ties are ambiguous, skip them).

Emits lang/hindi/lexicon.tsv. This is a production feature; it is NOT used in
the curated accuracy benchmark (that would be circular — see
docs/reviews/2026-09-04-h3-lexicon-layer.md and 2026-09-05-track-c-*.md).
"""
import csv
import sys
from collections import defaultdict

HIGH_ATT = 4
MIN_ATT = int(sys.argv[1]) if len(sys.argv) > 1 else 2
SRC = 'benchmark/data/dakshina_hi.csv'
FREQ = 'benchmark/data/freq_hi.csv'
OUT = 'lang/hindi/lexicon.tsv'

freq = {}
try:
    with open(FREQ, encoding='utf-8') as f:
        r = csv.reader(f)
        next(r, None)
        for row in r:
            if len(row) >= 2:
                freq[row[0]] = int(row[1])
except FileNotFoundError:
    print(f"note: {FREQ} not found; building high-attestation core only")

d = defaultdict(lambda: defaultdict(int))
with open(SRC, encoding='utf-8') as f:
    r = csv.reader(f)
    next(r, None)
    for row in r:
        if len(row) < 3 or 'split=train' not in row[2]:
            continue
        att = next((int(t.split('=')[1]) for t in row[2].split()
                    if t.startswith('attestations=')), 1)
        d[row[0]][row[1]] += att

rows = []
core = expanded = 0
for native, spellings in d.items():
    best_rom = max(spellings, key=spellings.get)
    best = spellings[best_rom]
    if best >= HIGH_ATT:
        rows.append((native, best_rom))
        core += 1
        continue
    if native in freq and best >= MIN_ATT:
        # strict winner: no other spelling ties the best attestation count
        if sum(1 for v in spellings.values() if v == best) == 1:
            rows.append((native, best_rom))
            expanded += 1
rows.sort()

with open(OUT, 'w', encoding='utf-8') as f:
    for native, roman in rows:
        f.write(f"{native}\t{roman}\n")
print(f"wrote {OUT}: {len(rows)} entries "
      f"({core} high-attestation core, {expanded} frequency-ranked, min_att={MIN_ATT})")
