#!/usr/bin/env python3
"""Mine lexicon-candidate entries for frequent words lacking Dakshina-train gold.

T-0017. Pipeline (human-in-the-loop per the hybrid statistical+rule principle:
machine proposes, evidence recorded, humans/benchmarks keep control):

  pool  = top-frequency words (Shabd) with NO Dakshina-train gold,
          EXCLUDING every held-out native (Dakshina dev/test, Aksharantar test)
  mine  = attested spellings from Aksharantar train (local aksharantar_hi.csv);
          optionally an external model's output (--model-output FILE: TSV
          native<TAB>roman, e.g. from IndicXlit ai4bharat-transliteration)
  keep  = strict-winner single spelling + consonant-skeleton check (>=70%)

Outputs mined_candidates_hi.csv (native,roman,evidence) for
review, for HUMAN review only. Measured precision of unreviewed candidates against the
COMI-LINGUA overlap is ~43% (57% with an engine-distance filter) — far below
the lexicon's >=4-human-attestation bar, so nothing is auto-promoted; see
docs/reviews/2026-09-05-p3-mining-and-reranker.md.

Validation without contamination: COMI-LINGUA is a benchmark, so it is NEVER a
mining source; the AK∩COMI overlap is used only to REPORT estimated precision
(how often the mined spelling matches an independently-typed COMI variant).
"""
import csv
import sys
from collections import defaultdict

sys.path.insert(0, 'tools')
from build_comilingua import is_transliteration  # engine-neutral skeleton check

MODEL_OUT = None
if '--model-output' in sys.argv:
    MODEL_OUT = sys.argv[sys.argv.index('--model-output') + 1]


def rows(p):
    with open(p, encoding='utf-8') as f:
        r = csv.reader(f)
        next(r, None)
        for row in r:
            if len(row) >= 2:
                yield row


freq = {w: int(n) for w, n, *_ in rows('benchmark/data/freq_hi.csv')}
dak_train, held = set(), set()
for w, ro, *rest in rows('benchmark/data/dakshina_hi.csv'):
    (dak_train if rest and 'split=train' in rest[0] else held).add(w)
held |= {w for w, *_ in rows('benchmark/data/aksharantar_test_hi.csv')}
comi = defaultdict(set)
for w, ro, *_ in rows('benchmark/data/comilingua_hi.csv'):
    comi[w].add(ro)
ak = defaultdict(set)
for w, ro, *_ in rows('benchmark/data/aksharantar_hi.csv'):
    ak[w].add(ro)
model = {}
if MODEL_OUT:
    with open(MODEL_OUT, encoding='utf-8') as f:
        for line in f:
            parts = line.rstrip('\n').split('\t')
            if len(parts) == 2:
                model[parts[0]] = parts[1]

pool = [w for w in freq if w not in dak_train and w not in held]
existing = {w for w, *_ in (r.split('\t') for r in open('lang/hindi/lexicon.tsv', encoding='utf-8').read().splitlines())}

cands = []
prec_hit = prec_tot = 0
for w in pool:
    if w in existing:
        continue
    spellings = set(ak.get(w, set()))
    if w in model:
        spellings.add(model[w])
    valid = sorted(s for s in spellings if s.isascii() and s.isalpha()
                   and is_transliteration(w, s.lower()))
    if len(valid) != 1:  # strict winner only; ambiguous words stay unmined
        continue
    roman = valid[0].lower()
    ev = 'src=aksharantar' + ('+model' if w in model else '')
    cands.append((w, roman, f'{ev} freq={freq[w]}'))
    if w in comi:  # precision estimate only — never a promotion criterion
        prec_tot += 1
        prec_hit += roman in comi[w]

cands.sort(key=lambda x: -int(x[2].rsplit('=', 1)[1]))
with open('mined_candidates_hi.csv', 'w', encoding='utf-8') as f:
    f.write('native,roman,notes\n')
    for w, ro, ev in cands:
        f.write(f'{w},{ro},{ev}\n')
print(f'pool={len(pool)} candidates={len(cands)}')
if prec_tot:
    print(f'estimated precision vs COMI-LINGUA overlap: {prec_hit}/{prec_tot} '
          f'({100*prec_hit/prec_tot:.1f}%)')
