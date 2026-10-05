#!/usr/bin/env python3
"""Synthetic Bengali NFC parity cases, covering the block's letter/mark pairs."""
import argparse
import gzip
import itertools
import json
from pathlib import Path
import unicodedata
from build_lexicon import native_key


def cases():
    alphabet = [chr(i) for i in range(0x980, 0xa00) if unicodedata.category(chr(i))[0] in 'LM']
    inputs = set(alphabet)
    inputs.update(a+b for a in alphabet for b in alphabet)
    for marks in itertools.product('়্\u09fe', repeat=3):
        inputs.add('ক'+''.join(marks))
    for word in ('বড়', 'বঢ়', 'হয়', 'কো', 'কৌ', 'কোড়ক', 'দৌড়'):
        inputs.update((word, unicodedata.normalize('NFD', word), '\u200d'+word, '\u200c'.join(word)))
    return [[word, native_key(word)] for word in sorted(inputs)]


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    args.output.write_bytes(gzip.compress(json.dumps(cases(), ensure_ascii=False).encode(), mtime=0))
