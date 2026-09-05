#!/usr/bin/env python3
"""Build benchmark/data/freq_hi.csv from the Shabd psycholinguistic database.

Shabd (CC0 1.0): https://osf.io/xfbhd/ — a Hindi word-frequency list from a
1.4B-token news corpus. We keep the top-N Devanagari word types by frequency for
a frequency-WEIGHTED real-world evaluation (common words matter more than rare).

Source file: shabd96k.csv (Word, Frequency, ...). Output: native,freq (CC0).
"""
import csv, sys, re
SRC = sys.argv[1]
OUT = 'benchmark/data/freq_hi.csv'
TOPN = int(sys.argv[2]) if len(sys.argv) > 2 else 15000
DEVA = re.compile(r'^[ऀ-ॿ]+$')
rows = []
with open(SRC, encoding='utf-8') as f:
    r = csv.reader(f); next(r, None)
    for row in r:
        if len(row) < 2: continue
        w = row[0].strip()
        try: freq = int(float(row[1]))
        except ValueError: continue
        if DEVA.match(w):
            rows.append((w, freq))
rows.sort(key=lambda x: -x[1])
rows = rows[:TOPN]
with open(OUT, 'w', encoding='utf-8') as f:
    f.write("native,freq\n")
    for w, freq in rows:
        f.write(f"{w},{freq}\n")
print(f"wrote {OUT}: {len(rows)} words (top {TOPN} Devanagari by frequency)")
print("top 5:", rows[:5])
