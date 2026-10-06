#!/usr/bin/env python3
"""Validate frozen-selector runtime parity and external scores; never fit or select."""
import argparse
import csv
import gzip
import io
import json
from pathlib import Path
import subprocess
import zipfile
from train_crossfit_selector import ROOT, render, choose, dumps
from train_reranker import references, metrics
from evaluate_external import distance
from bengali_training_data import pinned, sha, verify, native_key
from audit_bengali_training_overlap import bengali_tokens

PROFILES = ('B1', 'model', 'lexicon', 'model_lexicon', 'rerank', 'rerank_lexicon')


def predict(rows):
    proc = subprocess.run(['go', 'run', './tools/bengali/evaluate_runtime'], cwd=ROOT,
                          input=''.join(json.dumps(w, ensure_ascii=False)+'\n' for w, _ in rows),
                          text=True, capture_output=True, check=True)
    values = [json.loads(line) for line in proc.stdout.splitlines()]
    if len(values) != len(rows) or any(len(p) != len(PROFILES) for p in values):
        raise ValueError('prediction shape mismatch')
    return values


def score(rows, predictions):
    if len(rows) != len(predictions):
        raise ValueError('prediction count mismatch')
    result = {'items': len(rows)}
    for column, name in enumerate(PROFILES):
        matches = sum(p[column] in refs for (_, refs), p in zip(rows, predictions))
        cer = sum(min(distance(p[column], ref)/max(len(ref), 1) for ref in refs) for (_, refs), p in zip(rows, predictions))
        result[name] = {'match_any': matches, 'accuracy': matches/len(rows) if rows else None,
                        'macro_minCER': cer/len(rows) if rows else None}
    for a, b, name in ((1, 4, 'rerank_vs_model'), (3, 5, 'rerank_lexicon_vs_model_lexicon')):
        result[name] = {'wins': sum(p[b] in refs and p[a] not in refs for (_, refs), p in zip(rows, predictions)),
                        'losses': sum(p[a] in refs and p[b] not in refs for (_, refs), p in zip(rows, predictions)),
                        'changed': sum(p[a] != p[b] for p in predictions)}
    return result


FROZEN = ROOT/'docs/reviews/2026-10-04-bengali-b2-crossfit-selector.json'
EMBEDDED = ROOT/'lang/bengali/selector.json'
BANGLATLIT_SHA = '386721b4221e2d93b16ab73f3761c7c4e4cd295adf1160fe9c95c06cb34c5386'


def dakshina_section():
    """Dakshina dev/test with word-for-word offline parity; no external data."""
    if EMBEDDED.read_bytes() != FROZEN.read_bytes():
        raise ValueError('runtime selector differs from frozen experiment')
    tree = json.loads(FROZEN.read_text())['tree']
    dak = {}
    for split in ('dev', 'test'):
        rows = references(split)
        values = predict(rows)
        choices = render(rows, [], False)
        expected = [choose(tree, word, options, .6) for (word, _), options in zip(rows, choices)]
        if [p[4] for p in values] != expected:
            bad = [(word, actual[4], want) for (word, _), actual, want in zip(rows, values, expected) if actual[4] != want]
            raise ValueError('wordwise offline/runtime parity failure: '+str(bad[:5]))
        if any(p[1] != p[3] or p[4] != p[5] for p in values):
            raise ValueError('lexicon changed held-out output')
        dak[split] = dict(score(rows, values), strict_profiles={name: metrics(rows, [p[i] for p in values]) for i, name in enumerate(PROFILES)},
                          offline_wordwise_mismatches=0)
    return dak


def banglatlit_rows():
    blob = pinned(ROOT/'benchmark/data/banglatlit/test.csv.gz', BANGLATLIT_SHA)
    return blob, [(r['native'], [r['roman']]) for r in csv.DictReader(io.StringIO(gzip.decompress(blob).decode()))]


def banglatlit_section(bangla, values, training):
    unseen = [(r, p) for r, p in zip(bangla, values)
              if (tokens := set(bengali_tokens(r[0]))) and not tokens & training]
    return {'full_sentences': score(bangla, values),
            'all_tokens_unseen_sentences': score([r for r, _ in unseen], [p for _, p in unseen])}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('aksharantar_zip', type=Path)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    embedded = EMBEDDED
    dak = dakshina_section()
    raw = pinned(args.aksharantar_zip, '4ab6edcc6ab556040d8f43ee75eec73cc25af0cf72d09e4650dab26b196d5fe7')
    with zipfile.ZipFile(io.BytesIO(raw)) as archive:
        member = archive.read('ben_test.json')
    words = {}
    for line in member.decode().splitlines():
        row = json.loads(line)
        words.setdefault(native_key(row['native word']), set()).add(row['english word'])
    ak = [(word, sorted(refs)) for word, refs in sorted(words.items())]
    blob, bangla = banglatlit_rows()
    values = predict(ak+bangla)
    groups = verify()
    training = groups['google-train'] | groups['dakshina-train']
    unseen_ak = [(r, p) for r, p in zip(ak, values[:len(ak)]) if r[0] not in training]
    if any(p[1] != p[3] or p[4] != p[5] for _, p in unseen_ak):
        raise ValueError('lexicon changed unseen external word')
    sources = [ROOT/'gomanize.go', ROOT/'tools/bengali/evaluate_runtime.py'] + [p for folder in (
        'core', 'lang/bengali', 'script/brahmic', 'scheme', 'tools/bengali/evaluate_runtime')
        for p in (ROOT/folder).rglob('*.go') if not p.name.endswith('_test.go')]
    result = {'selector_sha256': sha(embedded.read_bytes()), 'threshold': .6,
              'vowel_model_sha256': sha((ROOT/'lang/bengali/vowel_tree.json').read_bytes()),
              'lexicon_sha256': sha((ROOT/'lang/bengali/lexicon.tsv').read_bytes()),
              'sources_sha256': {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in sorted(sources)},
              'aksharantar_archive_sha256': sha(raw), 'aksharantar_test_sha256': sha(member),
              'banglatlit_fixture_sha256': sha(blob), 'Dakshina': dak,
              'Aksharantar': {'full_types': score(ak, values[:len(ak)]),
                              'unseen_types': score([r for r, _ in unseen_ak], [p for _, p in unseen_ak])},
              'BanglaTLit': banglatlit_section(bangla, values[len(ak):], training),
              'unseen_definition': 'absent from union of frozen Google and Dakshina training vocabularies',
              'limitations': ['Aksharantar overlaps Dakshina test; not independent',
                             'sentence references cannot establish token-level accuracy', 'lyrics gold remains T-0054']}
    args.output.write_bytes(dumps(result))
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
