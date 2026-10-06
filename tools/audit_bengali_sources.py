#!/usr/bin/env python3
"""Report source revisions, schema, and overlap without printing corpus examples."""
import argparse
import collections
import csv
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import unicodedata


def key(text):
    return unicodedata.normalize('NFC', ''.join(c for c in text if unicodedata.category(c) != 'Cf'))


def audit(root, fixtures):
    report = {'normalization': 'Cf removal + NFC; sentence keys also trim outer whitespace', 'repos': {}}
    for name in ['bntranslit', 'bengali-romanizer', 'BanglaTLit', 'Romabangla']:
        report['repos'][name] = subprocess.check_output(['git', '-C', str(root / name), 'rev-parse', 'HEAD'], text=True).strip()
    dakshina = {}
    for split in ['train', 'dev', 'test']:
        with gzip.open(fixtures / (split + '.csv.gz'), 'rt') as stream:
            dakshina[split] = {key(row['native']) for row in csv.DictReader(stream)}
    lexicon = root / 'bengali-romanizer/bengali_romanizer/data/lexicon.tsv.gz'
    readings = collections.defaultdict(set)
    count = 0
    with gzip.open(lexicon, 'rt') as stream:
        for line in stream:
            if line.startswith('#') or not line.strip():
                continue
            native, phonemes, *_ = line.rstrip('\n').split('\t')
            readings[key(native)].add(phonemes)
            count += 1
    keys = set(readings)
    report['google_lexicon'] = {
        'sha256': hashlib.sha256(lexicon.read_bytes()).hexdigest(), 'rows': count,
        'normalized_words': len(keys), 'multi_pronunciation_words': sum(len(v) > 1 for v in readings.values()),
        'dakshina_overlap': {s: len(keys & words) for s, words in dakshina.items()},
        'after_dev_test_exclusion': len(keys - dakshina['dev'] - dakshina['test']),
    }
    sets = {}
    report['BanglaTLit'] = {}
    for split, filename in [('train', 'BanglaTLit_train.csv'), ('dev', 'BanglaTLiT_val.csv'), ('test', 'BanglaTLiT_test.csv')]:
        path = root / 'BanglaTLit/data' / filename
        with path.open() as stream:
            reader = csv.DictReader(stream)
            columns = reader.fieldnames
            rows = list(reader)
        valid = [r for r in rows if any('\u0980' <= c <= '\u09ff' for c in r['text_bengali'])]
        pairs = {(key(r['text_bengali']).strip(), key(r['text_transliterated']).strip()) for r in valid}
        natives, romans = {n for n, _ in pairs}, {r for _, r in pairs}
        sets[split] = (natives, romans, pairs)
        report['BanglaTLit'][split] = {
            'source_path': 'data/' + filename, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest(),
            'columns': columns, 'rows': len(rows), 'with_bengali': len(valid),
            'empty_bengali': sum(not key(r['text_bengali']).strip() for r in rows),
            'empty_roman': sum(not key(r['text_transliterated']).strip() for r in rows),
            'unique_ids': len({r['id'] for r in rows}), 'unique_native': len(natives),
            'unique_roman': len(romans), 'unique_pairs': len(pairs),
        }
    report['BanglaTLit_overlap'] = {}
    for a, b in [('train', 'dev'), ('train', 'test'), ('dev', 'test')]:
        report['BanglaTLit_overlap'][a + '-' + b] = {
            label: len(sets[a][i] & sets[b][i]) for i, label in enumerate(['native', 'roman', 'pair'])
        }
    return report


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--checkouts', required=True, type=Path)
    parser.add_argument('--dakshina', type=Path, default=Path('benchmark/data/bengali'))
    args = parser.parse_args()
    print(json.dumps(audit(args.checkouts, args.dakshina), ensure_ascii=False, indent=2))
