#!/usr/bin/env python3
"""Build the Bengali default-style spelling lexicon from isolated Dakshina train."""
import argparse
from collections import Counter, defaultdict
import json
from pathlib import Path
import sys
import unicodedata
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import DEFAULT, ROOT, DAK_SHA, assert_lexicon_isolation, native_key, sha, training_rows

MIN_ATTESTATIONS = 3
OUTPUT = ROOT / 'lang/bengali'


def select(rows):
    words = defaultdict(Counter)
    for row in rows:
        word, roman, votes = native_key(row['native']), row['roman'], int(row['attestations'])
        if not word or not roman or votes <= 0 or any(c in roman for c in '\t\r\n'):
            raise ValueError('invalid source row')
        words[word][roman] += votes
    selected = {}
    counts = Counter({'source_types': len(words), 'below_threshold': 0, 'tied_top': 0, 'nonword': 0})
    for word, refs in sorted(words.items()):
        if any(not ('\u0980' <= c <= '\u09ff' and unicodedata.category(c)[0] in 'LM') for c in word):
            counts['nonword'] += 1
            continue
        top = max(refs.values())
        if top < MIN_ATTESTATIONS:
            counts['below_threshold'] += 1
            continue
        winners = [r for r, votes in refs.items() if votes == top]
        if len(winners) != 1:
            counts['tied_top'] += 1
            continue
        selected[word] = winners[0]
    counts['entries'] = len(selected)
    return selected, dict(sorted(counts.items()))


def build():
    entries, counts = select(training_rows('dakshina'))
    assert_lexicon_isolation(entries)
    payload = ''.join(f'{word}\t{roman}\n' for word, roman in sorted(entries.items())).encode()
    manifest = {
        'schema': 1, 'source': 'Dakshina Bengali v1.0 train; frozen T-0057 exclusions',
        'license': 'CC-BY-SA-4.0', 'style': 'default only; bypass on any alternate-style option',
        'selection': 'at least three attestations for the unique highest-vote spelling; reject ties; Bengali letter/mark keys only',
        'min_attestations': MIN_ATTESTATIONS, 'counts': counts,
        'normalization': 'remove Unicode Cf; NFC Bengali native keys; Roman case/spelling unchanged',
        'source_fixture_sha256': DAK_SHA['train'],
        'split_manifest_sha256': sha((DEFAULT/'manifest.json').read_bytes()),
        'builder_sha256': sha(Path(__file__).read_bytes()), 'lexicon_sha256': sha(payload),
    }
    return payload, (json.dumps(manifest, indent=2)+'\n').encode()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=OUTPUT)
    parser.add_argument('--verify', action='store_true')
    args = parser.parse_args()
    payload, manifest = build()
    if args.verify:
        if ((args.output/'lexicon.tsv').read_bytes() != payload or
                (args.output/'lexicon_manifest.json').read_bytes() != manifest):
            raise SystemExit('lexicon/provenance differs from isolated training reconstruction')
        print('Bengali lexicon/provenance exactly matches isolated training reconstruction')
    else:
        args.output.mkdir(parents=True, exist_ok=True)
        (args.output/'lexicon.tsv').write_bytes(payload)
        (args.output/'lexicon_manifest.json').write_bytes(manifest)
        print(manifest.decode())
