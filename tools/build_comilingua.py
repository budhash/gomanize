#!/usr/bin/env python3
"""Extract a colloquial word-level benchmark from COMI-LINGUA's MT test split.

Source: https://huggingface.co/datasets/LingoIITGN/COMI-LINGUA (CC-BY 4.0;
Sheth et al., Findings of EMNLP 2025). Each row carries, per human annotator,
a Devanagari-Hindi translation (DH) and a Romanized-Hindi translation (RH) of
the same sentence — parallel text showing how people ACTUALLY romanize Hindi.
Machine "Predicted_*" columns are ignored; only human annotator columns are used.

Alignment: for each (DH, RH) sentence pair with equal whitespace token counts,
per-position tokens are paired; a pair is kept when the DH token is pure
Devanagari and the RH token is pure Latin (after stripping punctuation).
Variants are aggregated per DH word with occurrence counts — the counts make the
set naturally frequency-weighted toward the colloquial register.

Output: benchmark/data/comilingua_hi.csv (native,roman,notes count=N).
Words with a single occurrence are kept only if 3+ chars (noise control).

Usage: python3 tools/build_comilingua.py <MT_test.csv>
"""
import csv
import re
import sys
from collections import defaultdict

SRC = sys.argv[1]
OUT = 'benchmark/data/comilingua_hi.csv'
DEVA = re.compile(r'^[ऀ-ॿ]+$')
LATIN = re.compile(r'^[a-z]+$')
STRIP = '।.,!?;:"\'()[]{}“”‘’-–—…'

# Positional alignment pairs some Devanagari words with English TRANSLATIONS
# (annotators code-switch: अंक → "points"). Filter those with an engine-neutral
# consonant-skeleton check: the word's consonants (standard chart) must appear
# in order in the roman. Requires >=70% in-order hits (tolerates variant
# spellings like q/k, w/v via alternates).
CONS = {
 'क': ('k', 'q', 'c'), 'ख': ('kh',), 'ग': ('g',), 'घ': ('gh',), 'च': ('ch',), 'छ': ('chh', 'ch'),
 'ज': ('j', 'z'), 'झ': ('jh',), 'ट': ('t',), 'ठ': ('th',), 'ड': ('d',), 'ढ': ('dh',),
 'ण': ('n',), 'त': ('t',), 'थ': ('th',), 'द': ('d',), 'ध': ('dh',), 'न': ('n',),
 'प': ('p',), 'फ': ('ph', 'f'), 'ब': ('b',), 'भ': ('bh',), 'म': ('m',), 'य': ('y',),
 'र': ('r',), 'ल': ('l',), 'व': ('v', 'w'), 'श': ('sh', 's'), 'ष': ('sh', 's'),
 'स': ('s',), 'ह': ('h',), 'ळ': ('l',),
 'क़': ('q', 'k'), 'ख़': ('kh',), 'ग़': ('g',), 'ज़': ('z', 'j'), 'ड़': ('r', 'd'),
 'ढ़': ('rh', 'dh'), 'फ़': ('f', 'ph'),
}
NUKTA = '़'


def is_transliteration(native, roman):
    skel = []
    chars = list(native)
    i = 0
    while i < len(chars):
        c = chars[i]
        if i + 1 < len(chars) and chars[i + 1] == NUKTA:
            c += NUKTA
            i += 1
        if c in CONS:
            skel.append(CONS[c])
        i += 1
    if not skel:
        return True  # vowel-only word: nothing to check
    pos = hits = 0
    for alts in skel:
        for alt in alts:
            j = roman.find(alt, pos)
            if j >= 0:
                pos = j + len(alt)
                hits += 1
                break
    return hits / len(skel) >= 0.7

pairs = defaultdict(lambda: defaultdict(int))
cols = [('Annotator_1_DH_translation', 'Annotator_1_RH_translation'),
        ('Annotator_2_DH_translation', 'annotator2_RH_translation'),
        ('Annotator_3_DH_translation', 'annotator3_RH_translation')]

csv.field_size_limit(10 ** 7)
sentences = 0
with open(SRC, encoding='utf-8') as f:
    r = csv.DictReader(f)
    for row in r:
        for dh_col, rh_col in cols:
            dh, rh = (row.get(dh_col) or '').strip(), (row.get(rh_col) or '').strip()
            if not dh or not rh:
                continue
            dt, rt = dh.split(), rh.split()
            if len(dt) != len(rt):
                continue
            sentences += 1
            for d, ro in zip(dt, rt):
                d, ro = d.strip(STRIP), ro.strip(STRIP).lower()
                if DEVA.match(d) and LATIN.match(ro) and is_transliteration(d, ro):
                    pairs[d][ro] += 1

rows = []
for native, spellings in pairs.items():
    for roman, n in spellings.items():
        if n >= 2 or len(roman) >= 3:
            rows.append((native, roman, n))
rows.sort(key=lambda x: (x[0], -x[2]))

with open(OUT, 'w', encoding='utf-8') as f:
    f.write("native,roman,notes\n")
    for native, roman, n in rows:
        f.write(f"{native},{roman},count={n}\n")
uniq = len({r[0] for r in rows})
print(f"aligned sentence pairs: {sentences}")
print(f"wrote {OUT}: {len(rows)} rows, {uniq} unique natives")
