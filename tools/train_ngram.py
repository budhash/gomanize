#!/usr/bin/env python3
"""Train a character 4-gram LM on Dakshina TRAIN romanizations (held-out safe).

Counts are attestation-weighted. Stupid-backoff scoring happens in Go
(lang/hindi/reranker.go); this emits lang/hindi/roman_ngrams.tsv:
  ngram<TAB>count      (1..4-grams over [a-z] with ^ start / $ end markers)
Grams with count < 3 are pruned. Pure stdlib.
"""
import csv
from collections import Counter

counts = Counter()
with open('benchmark/data/dakshina_hi.csv', encoding='utf-8') as f:
    r = csv.reader(f)
    next(r, None)
    for row in r:
        if len(row) < 3 or 'split=train' not in row[2]:
            continue
        att = next((int(t.split('=')[1]) for t in row[2].split()
                    if t.startswith('attestations=')), 1)
        s = '^' + row[1] + '$'
        if not all(c.isascii() and (c.isalpha() or c in '^$') for c in s):
            continue
        for n in range(1, 5):
            for i in range(len(s) - n + 1):
                counts[s[i:i + n]] += att

kept = {g: c for g, c in counts.items() if c >= 3}
with open('lang/hindi/roman_ngrams.tsv', 'w', encoding='utf-8') as f:
    for g in sorted(kept):
        f.write(f'{g}\t{kept[g]}\n')
print(f'ngrams kept: {len(kept)} (pruned {len(counts)-len(kept)})')
