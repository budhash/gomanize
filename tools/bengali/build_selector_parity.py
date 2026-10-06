#!/usr/bin/env python3
"""Synthetic native-selector features and candidate parity; no corpus references."""
import gzip
import io
import json
from pathlib import Path
from train_crossfit_selector import ROOT, render, choose
from train_native_selector import features


def build():
    tree = json.loads((ROOT/'lang/bengali/selector.json').read_text())['tree']
    words = {'কমল', 'আহ', 'কাম', 'ড়ক', 'কোড়', 'ক্ষ', 'কাা', 'কঅ', 'হওয়া', 'চাঁদ', 'কমলকমলকমলকমলকমল'}
    for c in 'কখগচজটডতদনপফবমরলসহ':
        for matra in ('', 'া', 'ি', 'ু', 'ে', 'ো'):
            for ending in ('ম', 'ল', 'রের', 'লের', 'নের'):
                words.add(c+matra+ending)
    for prefix in ('জন', 'সম', 'অব'):
        for ending in ('ক', 'ল', 'ম', 'কর', 'কল', 'ত', 'তা', 'রক', 'কাল', 'কু', 'কক'):
            words.add(prefix+ending)
    rows = [(word, {}) for word in sorted(words)]
    candidates = render(rows, [], False)
    cases = [{'word': word, 'candidates': options, 'output': choose(tree, word, options, .6),
              'features': [features(word, alt, options[0]) for alt in options[1:]]}
             for (word, _), options in zip(rows, candidates)]
    # GzipFile writes the same header on every Python version; gzip.compress
    # switched to zlib in 3.11 and emits a different OS byte.
    buffer = io.BytesIO()
    with gzip.GzipFile(fileobj=buffer, mode='wb', mtime=0) as out:
        out.write((json.dumps(cases, ensure_ascii=False, separators=(',', ':'))+'\n').encode())
    return buffer.getvalue()


if __name__ == '__main__':
    (ROOT/'lang/bengali/testdata/selector_parity.json.gz').write_bytes(build())
