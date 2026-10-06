#!/usr/bin/env python3
"""Fit a native-conditioned candidate selector on isolated train, then gate on dev."""
import argparse
from collections import defaultdict
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DEFAULT, DAK_SHA, native_key, training_rows, verify, sha
from train_reranker import candidates, references, metrics
from evaluate_external import distance

# Fixed before the first dev evaluation. No parameter search beyond this grid.
THRESHOLDS = (.5, .6, .7, .8, .9, .95)
DEPTH = 5
MIN_LEAF = 30


def features(native, b1, model):
    native = native_key(native)
    prefix = 0
    while prefix < min(len(b1), len(model)) and b1[prefix] == model[prefix]:
        prefix += 1
    suffix = 0
    while suffix < min(len(b1), len(model))-prefix and b1[-1-suffix] == model[-1-suffix]:
        suffix += 1
    left = b1[prefix:len(b1)-suffix if suffix else len(b1)]
    right = model[prefix:len(model)-suffix if suffix else len(model)]
    return {'first': native[:1], 'first2': native[:2], 'last': native[-1:], 'last2': native[-2:],
            'length': str(min(len(native), 12)), 'has_halant': str('্' in native),
            'has_i': str(any(c in native for c in 'ইঈিী')), 'has_u': str(any(c in native for c in 'উঊুূ')),
            'has_o': str(any(c in native for c in 'ওো')), 'has_nasal': str(any(c in native for c in 'ংঁ')),
            'edit_b1': left[:2], 'edit_model': right[:2], 'delta': str(max(-3, min(3, len(b1)-len(model)))),
            'edit_at_start': str(prefix == 0), 'edit_at_end': str(suffix == 0),
            'before': b1[max(0, prefix-1):prefix], 'after': b1[len(b1)-suffix:][:1] if suffix else ''}


def training_references(rows, groups):
    """Also exclude Google train so candidate-model training cannot bias labels."""
    grouped = defaultdict(lambda: defaultdict(int))
    for row in rows:
        word = native_key(row['native'])
        if word not in groups['dakshina-train']:
            raise ValueError('unauthorized selector training word')
        if word in groups['google-train']:
            continue
        votes = int(row['attestations'])
        if votes <= 0 or not row['roman']:
            raise ValueError('invalid training reference')
        grouped[word][row['roman']] += votes
    return sorted((word, dict(refs)) for word, refs in grouped.items())


def target(refs, b1, model):
    best = min(refs, key=lambda r: (-refs[r], r))
    def quality(candidate):
        return (candidate in refs, candidate == best,
                -min(distance(candidate, ref)/max(len(ref), 1) for ref in refs))
    return int(quality(b1) > quality(model))


def impurity(counts):
    n = sum(counts)
    return n - sum(c*c for c in counts)/n if n else 0


def fit(rows, depth=DEPTH, min_leaf=MIN_LEAF):
    if not rows:
        raise ValueError('empty selector training set')
    counts = [sum(y == i for _, y in rows) for i in (0, 1)]
    node = {'counts': counts}
    if depth == 0 or len(rows) < 2*min_leaf or 0 in counts:
        return node
    best = None
    loss = impurity(counts) - 1e-9
    for feature in sorted(rows[0][0]):
        buckets = defaultdict(lambda: [0, 0])
        for x, y in rows:
            buckets[x[feature]][y] += 1
        for value, left in sorted(buckets.items()):
            right = [a-b for a, b in zip(counts, left)]
            if min(sum(left), sum(right)) < min_leaf:
                continue
            candidate = impurity(left) + impurity(right)
            if candidate < loss - 1e-9:
                best, loss = (feature, value), candidate
    if best is None:
        return node
    feature, value = best
    yes = [r for r in rows if r[0][feature] == value]
    no = [r for r in rows if r[0][feature] != value]
    return dict(node, feature=feature, value=value,
                yes=fit(yes, depth-1, min_leaf), no=fit(no, depth-1, min_leaf))


def probability(tree, x):
    while 'feature' in tree:
        tree = tree['yes' if x[tree['feature']] == tree['value'] else 'no']
    return tree['counts'][1]/sum(tree['counts'])


def predictions(tree, rows, choices, threshold):
    if len(rows) != len(choices):
        raise ValueError('candidate count mismatch')
    return [b1 if b1 != model and probability(tree, features(word, b1, model)) > threshold else model
            for (word, _), (b1, model) in zip(rows, choices)]


def select(baseline, trials):
    passing = [r for r in trials if r['any'] > baseline['any'] and r['strict'] >= baseline['strict']
               and r['CER'] < baseline['CER']]
    return min(passing, key=lambda r: (-r['any'], r['CER'], -r['strict'], -r['threshold'])) if passing else None


def build():
    groups = verify()
    source = training_rows('dakshina')
    rows = training_references(source, groups)
    choices = candidates(rows)
    examples = [(features(word, b1, model), target(refs, b1, model))
                for (word, refs), (b1, model) in zip(rows, choices) if b1 != model]
    tree = fit(examples)
    artifact = {'schema': 1, 'license': 'CC-BY-SA-4.0', 'depth': DEPTH, 'min_leaf': MIN_LEAF,
                'classes': ['model', 'B1'], 'tree': tree}
    payload = (json.dumps(artifact, ensure_ascii=False, sort_keys=True, indent=2)+'\n').encode()
    provenance = {'train_fixture_sha256': DAK_SHA['train'], 'dev_fixture_sha256': DAK_SHA['dev'],
                  'split_manifest_sha256': sha((DEFAULT/'manifest.json').read_bytes()),
                  'vowel_model_sha256': sha((ROOT/'lang/bengali/vowel_tree.json').read_bytes()),
                  'scripts_sha256': {name: sha((ROOT/name).read_bytes()) for name in
                      ('tools/bengali/train_native_selector.py', 'tools/bengali/train_reranker.py',
                       'tools/bengali/evaluate_external.py', 'tools/bengali_training_data.py')},
                  'candidate_sources_sha256': {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in sorted(
                      [ROOT/'gomanize.go'] + [p for folder in ('core', 'lang/bengali', 'script/brahmic', 'scheme', 'tools/bengali/evaluate')
                                             for p in (ROOT/folder).rglob('*.go') if not p.name.endswith('_test.go')])}}
    report = {'schema': 1, 'provenance': provenance, 'artifact_sha256': sha(payload),
              'training': {'isolated_dakshina_types': len(groups['dakshina-train']),
                           'excluded_google_training_types': len(groups['dakshina-train'] & groups['google-train']),
                           'eligible_types': len(rows), 'differing_candidates': len(examples),
                           'prefer_b1': sum(y for _, y in examples)},
              'policy': {'depth': DEPTH, 'min_leaf': MIN_LEAF, 'thresholds': THRESHOLDS,
                         'target': 'per-word match-any, strict, then minCER; ties prefer model; one vote per native type',
                         'selection': 'any strictly improves, strict nondecreasing, CER strictly improves; ties: any, CER, strict, higher threshold'}}
    dev = references('dev')
    choices = candidates(dev)
    baseline = metrics(dev, [p[1] for p in choices])
    trials = []
    for threshold in THRESHOLDS:
        outputs = predictions(tree, dev, choices, threshold)
        trials.append(dict(metrics(dev, outputs), threshold=threshold,
                           changed=sum(out != pair[1] for out, pair in zip(outputs, choices))))
    report.update(dev_model=baseline, dev_trials=trials, selected=select(baseline, trials))
    return payload, report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    payload, report = build()
    args.output.mkdir(parents=True, exist_ok=True)
    (args.output/'native_selector.json').write_bytes(payload)
    (args.output/'report.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
