#!/usr/bin/env python3
"""Measure external evaluation overlap without importing answers into training."""
import argparse
import csv
import gzip
import io
import json
import unicodedata
import zipfile

from bengali_training_data import ROOT, native_key, pinned, sha, verify


def bengali_tokens(text):
    """Maximal Bengali letter/mark runs; punctuation/digits/other scripts delimit."""
    current = ''
    for char in native_key(text) + ' ':
        if '\u0980' <= char <= '\u09ff' and unicodedata.category(char)[0] in 'LM':
            current += char
        elif current:
            yield current
            current = ''


def overlap(keys, groups):
    training = groups['google-train'] | groups['dakshina-train']
    return {'types': len(keys), 'by_partition': {k: len(keys & v) for k, v in groups.items()},
            'training_union': len(keys & training), 'unseen_training_types': len(keys - training)}


def audit(archive):
    groups = verify()
    archive_data = pinned(archive, '4ab6edcc6ab556040d8f43ee75eec73cc25af0cf72d09e4650dab26b196d5fe7')
    with zipfile.ZipFile(io.BytesIO(archive_data)) as z:
        raw = z.read('ben_test.json')
    rows = [json.loads(line) for line in raw.decode().splitlines() if line]
    keys = {native_key(r['native word']) for r in rows}
    fixture = pinned(ROOT / 'benchmark/data/banglatlit/test.csv.gz',
                     '386721b4221e2d93b16ab73f3761c7c4e4cd295adf1160fe9c95c06cb34c5386')
    sentences = list(csv.DictReader(io.StringIO(gzip.decompress(fixture).decode())))
    tokens = {word for r in sentences for word in bengali_tokens(r['native'])}
    return {
        'policy': 'evaluation overlap only; neither corpus supplies training rows or answers',
        'Aksharantar': {'url': 'https://huggingface.co/datasets/ai4bharat/Aksharantar/resolve/main/ben.zip',
                       'license': 'CC-BY-4.0', 'archive_sha256': sha(archive_data),
                       'test_member_sha256': sha(raw), 'rows': len(rows), **overlap(keys, groups)},
        'BanglaTLit': {'fixture_sha256': sha(fixture), 'sentences': len(sentences),
                      'tokenization': 'Cf removal + NFC; maximal Bengali Unicode block letter/mark runs',
                      **overlap(tokens, groups)},
        'lyrics_gold': {'status': 'not yet built (T-0054); overlap unknown, not zero'},
        'accuracy': 'no model exists yet; full and unseen-type accuracy required when evaluating B2 artifacts',
    }


if __name__ == '__main__':
    from pathlib import Path
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('archive', type=Path)
    args = parser.parse_args()
    print(json.dumps(audit(args.archive), indent=2))
