#!/usr/bin/env python3
"""Train a deterministic three-class categorical CART; never fit dev/test data."""
import argparse
import hashlib
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import DEFAULT, GOOGLE_SHA, google_rows, training_rows, verify
from vowels import FEATURES, dataset


def impurity(counts):
    n = sum(counts)
    return n - sum(c*c for c in counts) / n if n else 0


def tree(rows, depth, min_leaf, level=0):
    counts = [sum(r[1] == i for r in rows) for i in range(3)]
    majority = max(range(3), key=lambda i: counts[i])
    node = {'leaf': majority, 'counts': counts}
    if level == depth or len(rows) < 2*min_leaf or counts[majority] == len(rows):
        return node
    best = None
    best_loss = impurity(counts) - 1e-9
    for feature in FEATURES:
        buckets = {}
        for row in rows:
            buckets.setdefault(row[0][feature], [0, 0, 0])[row[1]] += 1
        for value, left in sorted(buckets.items()):
            right = [a-b for a, b in zip(counts, left)]
            if min(sum(left), sum(right)) < min_leaf:
                continue
            loss = impurity(left) + impurity(right)
            if loss < best_loss - 1e-9:
                best_loss = loss
                best = feature, value
    if best is None:
        return node
    feature, value = best
    yes = [r for r in rows if r[0][feature] == value]
    no = [r for r in rows if r[0][feature] != value]
    return {'f': feature, 'v': value, 'yes': tree(yes, depth, min_leaf, level+1),
            'no': tree(no, depth, min_leaf, level+1)}


def predict(model, features):
    node = model
    while 'leaf' not in node:
        node = node['yes' if features[node['f']] == node['v'] else 'no']
    return node['leaf']


def score(model, rows):
    confusion = [[0]*3 for _ in range(3)]
    for features, actual, *_ in rows:
        confusion[actual][predict(model, features)] += 1
    correct = sum(confusion[i][i] for i in range(3))
    return {'instances': len(rows), 'correct': correct, 'accuracy': correct / max(len(rows), 1),
            'confusion_actual_rows': confusion}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--google', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--depth', type=int, default=10)
    parser.add_argument('--min-leaf', type=int, default=20)
    parser.add_argument('--report-test', action='store_true', help='final reporting only, after dev selection')
    args = parser.parse_args()
    if args.depth < 1 or args.min_leaf < 1:
        parser.error('depth and min-leaf must be positive')
    groups = verify()
    train, train_stats = dataset(training_rows('google', args.google))
    # Selection source is explicit; held-out rows never enter tree().
    source = google_rows(args.google)
    dev, dev_stats = dataset([r for r in source if r[0] in groups['google-dev']])
    learned = tree(train, args.depth, args.min_leaf)
    artifact = {'schema': 1, 'classes': ['absent', 'default', 'raised'], 'features': FEATURES,
                'scope': 'simple-whole-words-v1',
                'provenance': {'google_sha256': GOOGLE_SHA,
                               'split_manifest_sha256': hashlib.sha256((DEFAULT/'manifest.json').read_bytes()).hexdigest(),
                               'training_scripts_sha256': {name: hashlib.sha256((Path(__file__).parent/name).read_bytes()).hexdigest()
                                                           for name in ('vowels.py', 'train_vowels.py')},
                               'depth': args.depth, 'min_leaf': args.min_leaf},
                'alignment': {'train': train_stats, 'dev': dev_stats},
                'metrics': {'train': score(learned, train), 'dev': score(learned, dev)}, 'tree': learned}
    if args.report_test:
        test, stats = dataset([r for r in source if r[0] in groups['google-test']])
        artifact['alignment']['test'] = stats
        artifact['metrics']['test'] = score(learned, test)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(artifact, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({k: artifact[k] for k in ('alignment', 'metrics')}, indent=2))


if __name__ == '__main__':
    main()
