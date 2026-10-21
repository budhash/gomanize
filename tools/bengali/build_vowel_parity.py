#!/usr/bin/env python3
"""Generate synthetic feature/prediction and eligibility parity cases, without corpus examples."""
import argparse
import hashlib
import json
from pathlib import Path
from vowels import CONS, HALANT, MATRAS, VOWELS, eligible, units, features
from train_vowels import predict


# Always included: feature caps (length 12, position 6) and nukta/split-matra
# edges. Truncation applies only to the generated grid, so these never drop out.
EDGE_WORDS = ['কমল', 'কাম', 'ফল', 'মন', 'ড়ক', 'কড়', 'কোড়', 'কামালমালকামাল',
              'কমলকমলকমলকমলকমল', 'মনমনমনমনমনমনমন', 'কামালমালকামালমালকামাল',
              'ড\u09bcকমলকমলকমলকমল', 'কে\u09beকমলকমলকমলক',
              # Halant clusters (T-0121): reph, phalas, ক্ষ, three members, nukta member.
              'গর্ত', 'মধ্য', 'ক্ষতি', 'পক্ষ', 'দ্বাল', 'স্মরণ', 'স্ত্রী', 'বস্ত্র', 'ড়্গ', 'কন্যা']
# Eligibility edges: unterminated, leading or doubled halant; four members;
# non-inventory members (ঞ, য়, ৎ); a slot before an independent vowel.
ELIGIBILITY_EDGES = ['ক্', '্ক', 'ক্্ক', 'ক্\u200dষ', 'ক্ষ্ম্য', 'স্ত্র্য', 'জ্ঞান', 'ক্য়', 'ক্ৎ', 'ক্অ',
                     'ক্কঅ', 'ক্কাঅ', 'া্ক', 'অ্ক', 'ক্ষ', 'ক্ক', 'ক', 'কা', 'ক্কা', 'ঙ্ক্ষ']


def cluster_words():
    words = set()
    for first in CONS:
        for second in CONS:
            words.add('ক' + first + HALANT + second)
            words.add(first + HALANT + second + 'া' + 'র')
    return sorted(words, key=lambda w: hashlib.sha256(w.encode()).digest())[:120]


def cases(model):
    words = set()
    for consonant in CONS:
        for vowel in MATRAS:
            words.add(consonant + vowel + 'কর')
        for vowel in VOWELS:
            words.add(vowel + consonant)
    words -= set(EDGE_WORDS)
    words = EDGE_WORDS + sorted(words, key=lambda w: hashlib.sha256(w.encode()).digest())[:180] + cluster_words()
    rows = []
    for word in words:
        for index, _, slot in units(word):
            if slot:
                f = features(word, index)
                rows.append({'word': word, 'index': index, 'features': f, 'label': predict(model, f)})
    return rows


def eligibility_cases():
    words = list(dict.fromkeys(EDGE_WORDS + ELIGIBILITY_EDGES + cluster_words()))
    return [{'word': w, 'eligible': eligible(w), 'legacy_eligible': eligible(w, 1)} for w in words]


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('model', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('eligibility', type=Path, nargs='?')
    args = parser.parse_args()
    model = json.loads(args.model.read_text())['tree']
    args.output.write_text(json.dumps(cases(model), ensure_ascii=False, indent=2) + '\n')
    if args.eligibility:
        args.eligibility.write_text(json.dumps(eligibility_cases(), ensure_ascii=False, indent=2) + '\n')
