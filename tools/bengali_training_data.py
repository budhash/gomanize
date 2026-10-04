#!/usr/bin/env python3
"""Frozen Bengali B2 word partitions; no alignment or training is performed."""
import argparse
import csv
import gzip
import hashlib
import io
import json
from pathlib import Path
import unicodedata

from build_bengali import native_key

ROOT = Path(__file__).resolve().parents[1]
DEFAULT = ROOT / 'training/data/bengali'
SEED = 'gomanize-bengali-b2-v1'
GOOGLE_SHA = '1bc2edda15da62bd4ef8576114e391c9a89ad8971f1eef3635e8fee0d7c1dd61'
GOOGLE_KEYS_SHA = {
    'train': 'c14efe6f8597256a951052607273d35cc5e9f66b943caced62da8eab6d70f414',
    'dev': '892f87edb11483b313bc637d06d1cb98bcb807fd7f5755cb38257055daa06fe6',
    'test': '035678309bcb0668e575a8a0d83dd826d7b51278ccb005be6b97181ec7fe985f',
}
DAK_SHA = dict(zip(('train', 'dev', 'test'), (
    '44b554fa749c717386bd9f68daaf837788d6c5ff10197d7f23d2bcefc28cf378',
    '8e83784102a527a88af0ad6dfba0fa41c11f57633a34fe2a9abe12acb54aa51a',
    '9cb15941b0893acdcc2891130ff98ba403645fd9170b955372af62f64bd33ca0')))


def sha(data):
    return hashlib.sha256(data).hexdigest()


def pinned(path, expected):
    data = Path(path).read_bytes()
    if sha(data) != expected:
        raise ValueError(f'source checksum mismatch: {path}')
    return data


def partition(key):
    bucket = int.from_bytes(hashlib.sha256((SEED + '\0' + key).encode()).digest()[:8], 'big') % 10000
    return 'train' if bucket < 9000 else 'dev' if bucket < 9500 else 'test'


def google_rows(path):
    text = gzip.decompress(pinned(path, GOOGLE_SHA)).decode('utf-8')
    rows = []
    for number, line in enumerate(text.splitlines(), 1):
        if not line or line.startswith('#'):
            continue
        fields = line.split('\t')
        if len(fields) not in (2, 3) or not fields[0] or not fields[1]:
            raise ValueError(f'invalid lexicon row {number}')
        # Original spelling, pronunciation, optional label and line number survive.
        rows.append((native_key(fields[0]), number, *fields))
    return rows


def dak_rows(split):
    path = ROOT / f'benchmark/data/bengali/{split}.csv.gz'
    data = pinned(path, DAK_SHA[split])
    return list(csv.DictReader(io.StringIO(gzip.decompress(data).decode())))


def split_keys(google, dak):
    protected = dak['dev'] | dak['test']
    groups = {f'google-{s}': set() for s in ('train', 'dev', 'test')}
    for key in google - protected:
        groups['google-' + partition(key)].add(key)
    protected |= groups['google-dev'] | groups['google-test']
    groups.update({'dakshina-' + s: keys - protected if s == 'train' else keys
                   for s, keys in dak.items()})
    validate_groups(groups)
    return groups


def validate_groups(groups):
    names = {f'{source}-{split}' for source in ('google', 'dakshina')
             for split in ('train', 'dev', 'test')}
    if set(groups) != names:
        raise ValueError('unexpected partition names')
    for name, keys in groups.items():
        if any(not key or native_key(key) != key for key in keys):
            raise ValueError(f'noncanonical key in {name}')
    heldout = set().union(*(v for k, v in groups.items() if not k.endswith('-train')))
    for name in ('google-train', 'dakshina-train'):
        if groups[name] & heldout:
            raise ValueError(f'held-out training leakage: {name}')
    for source in ('google', 'dakshina'):
        if groups[source + '-dev'] & groups[source + '-test']:
            raise ValueError(f'dev/test overlap: {source}')
    for split in ('train', 'dev', 'test'):
        if any(partition(k) != split for k in groups['google-' + split]):
            raise ValueError('Google split assignment changed')
    if (groups['google-dev'] | groups['google-test']) & (groups['dakshina-dev'] | groups['dakshina-test']):
        raise ValueError('Google pool contains Dakshina held-out types')


def key_bytes(keys):
    return ''.join(k + '\n' for k in sorted(keys)).encode()


def build(google_path, output):
    rows = google_rows(google_path)
    google = {r[0] for r in rows}
    dak = {s: {native_key(r['native']) for r in dak_rows(s)} for s in DAK_SHA}
    groups = split_keys(google, dak)
    output.mkdir(parents=True, exist_ok=True)
    manifest = {
        'schema': 1, 'seed': SEED,
        'assignment': 'SHA256(seed + NUL + NFC/Cf-free native); first 8 bytes big-endian modulo 10000; train <9000, dev <9500, else test',
        'unicode_version': unicodedata.unidata_version,
        'sources': {'google_sha256': GOOGLE_SHA, 'dakshina_sha256': DAK_SHA},
        'counts': {'google_rows': len(rows), 'google_original_types': len({r[2] for r in rows}),
                   'google_normalized_types': len(google),
                   'google_changed_rows': sum(r[0] != r[2] for r in rows),
                   'google_dakshina_overlap': {s: len(google & keys) for s, keys in dak.items()},
                   'dakshina_original_types': {s: len(keys) for s, keys in dak.items()},
                   'google_excluded_types': len(google - set().union(*(groups['google-' + s] for s in DAK_SHA))),
                   'dakshina_train_excluded_types': len(dak['train'] - groups['dakshina-train'])},
        'alignment': 'not run; ambiguous alignment counts belong to B2 model artifacts',
        'partitions': {},
    }
    for name, keys in sorted(groups.items()):
        data = key_bytes(keys)
        # Hash the canonical payload, independent of gzip implementation versions.
        (output / (name + '.txt.gz')).write_bytes(gzip.compress(data, mtime=0))
        manifest['partitions'][name] = {'words': len(keys), 'sha256': sha(data)}
    (output / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    verify(output)
    return manifest


def verify(directory=DEFAULT):
    manifest = json.loads((directory / 'manifest.json').read_text())
    if (manifest['schema'] != 1 or manifest['seed'] != SEED or
            manifest['sources'] != {'google_sha256': GOOGLE_SHA, 'dakshina_sha256': DAK_SHA}):
        raise ValueError('provenance or split policy changed')
    groups = {}
    for name, entry in manifest['partitions'].items():
        if name not in {f'{a}-{b}' for a in ('google', 'dakshina') for b in DAK_SHA}:
            raise ValueError('unexpected partition name')
        data = gzip.decompress((directory / (name + '.txt.gz')).read_bytes())
        keys = set(data.decode().splitlines())
        if sha(data) != entry['sha256'] or len(keys) != entry['words'] or data != key_bytes(keys):
            raise ValueError(f'partition integrity failure: {name}')
        groups[name] = keys
    validate_groups(groups)
    dak = {s: {native_key(r['native']) for r in dak_rows(s)} for s in DAK_SHA}
    if any(groups['dakshina-' + s] != dak[s] for s in ('dev', 'test')):
        raise ValueError('held-out Dakshina manifest changed')
    heldout = set().union(*(v for k, v in groups.items() if not k.endswith('-train')))
    if groups['dakshina-train'] != dak['train'] - heldout:
        raise ValueError('Dakshina training manifest changed')
    for split, expected in GOOGLE_KEYS_SHA.items():
        if sha(key_bytes(groups['google-' + split])) != expected:
            raise ValueError('frozen Google word inventory changed')
    return groups


def training_rows(source, google_path=None, directory=DEFAULT):
    """Only train rows reach future aligners, lexicon builders and rerankers."""
    groups = verify(directory)
    if source == 'google':
        rows = google_rows(google_path)
        # Verify completeness against the pinned raw source before yielding anything.
        dak = {s: {native_key(r['native']) for r in dak_rows(s)} for s in DAK_SHA}
        if split_keys({r[0] for r in rows}, dak) != groups:
            raise ValueError('frozen manifest differs from pinned sources')
        return [r for r in rows if r[0] in groups['google-train']]
    if source == 'dakshina':
        return [r for r in dak_rows('train') if native_key(r['native']) in groups['dakshina-train']]
    raise ValueError('training source must be google or dakshina')


def assert_lexicon_isolation(keys, directory=DEFAULT):
    groups = verify(directory)
    allowed = groups['dakshina-train']
    if {native_key(k) for k in keys} - allowed:
        raise ValueError('lexicon contains a held-out or unauthorized source type')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--google', type=Path, help='pinned lexicon.tsv.gz, required for build')
    parser.add_argument('--output', type=Path, default=DEFAULT)
    parser.add_argument('--verify', action='store_true')
    args = parser.parse_args()
    if args.verify:
        result = verify(args.output)
        print(json.dumps({k: len(v) for k, v in result.items()}, indent=2))
    elif args.google:
        print(json.dumps(build(args.google, args.output), indent=2))
    else:
        parser.error('--google is required to build')
