#!/usr/bin/env python3
"""Generate synthetic feature/prediction parity cases, without corpus examples."""
import argparse
import hashlib
import json
from pathlib import Path
from vowels import CONS, MATRAS, VOWELS, units, features
from train_vowels import predict


def cases(model):
    words = {'কমল', 'কাম', 'ফল', 'মন', 'ড়ক', 'কড়', 'কোড়', 'কামালমালকামাল'}
    for consonant in CONS:
        for vowel in MATRAS:
            words.add(consonant + vowel + 'কর')
        for vowel in VOWELS:
            words.add(vowel + consonant)
    words = sorted(words, key=lambda w: hashlib.sha256(w.encode()).digest())[:180]
    rows = []
    for word in words:
        for index, _, slot in units(word):
            if slot:
                f = features(word, index)
                rows.append({'word': word, 'index': index, 'features': f, 'label': predict(model, f)})
    return rows


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('model', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    model = json.loads(args.model.read_text())['tree']
    args.output.write_text(json.dumps(cases(model), ensure_ascii=False, indent=2) + '\n')
