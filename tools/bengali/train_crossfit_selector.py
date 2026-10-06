#!/usr/bin/env python3
"""Five-fold training with expanded vowel candidates; dev only after freezing fit."""
import argparse
from collections import defaultdict
import json
from pathlib import Path
import subprocess
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DEFAULT, DAK_SHA, GOOGLE_SHA, native_key, training_rows, verify, sha
from vowels import dataset, units, features as vowel_features
from train_vowels import tree as vowel_tree, predict
from train_native_selector import features, target, fit, probability, select
from train_reranker import references, metrics

SEED = 'gomanize-bengali-crossfit-v1'
FOLDS = 5
THRESHOLDS = (.5, .6, .7, .8, .9, .95)
MAX_EDITS = 8
DEPTH = 8
MIN_LEAF = 30


def fold(word):
    return int(sha((SEED+'\0'+native_key(word)).encode())[:16], 16) % FOLDS


def dumps(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2)+'\n').encode()


def grouped_training(rows, allowed):
    grouped = defaultdict(lambda: defaultdict(int))
    for row in rows:
        word = native_key(row['native'])
        if word not in allowed:
            raise ValueError('unauthorized crossfit training word')
        votes = int(row['attestations'])
        if votes <= 0 or not row['roman']:
            raise ValueError('invalid reference')
        grouped[word][row['roman']] += votes
    return sorted((word, dict(refs)) for word, refs in grouped.items())


def fold_examples(examples, held_fold):
    return [r for r in examples if fold(r[2]) != held_fold]


def decisions(word, tree):
    return {index: predict(tree, vowel_features(word, index))
            for index, _, slot in (units(word) or []) if slot}


def variants(labels):
    # Keep the current decision first. Each alternative flips just one slot;
    # default/raised both render o in the fixed colloquial style.
    out = [dict(labels)]
    for index in sorted(labels)[:MAX_EDITS]:
        changed = dict(labels)
        changed[index] = 1 if labels[index] == 0 else 0
        out.append(changed)
    return out


def render(rows, trees, use_folds):
    full = json.loads((ROOT/'lang/bengali/vowel_tree.json').read_text())['tree']
    requests = []
    for word, _ in rows:
        native = native_key(word)
        model = trees[fold(native)] if use_folds else full
        requests.append({'native': native, 'variants': variants(decisions(native, model)),
                         'check': decisions(native, full)})
    proc = subprocess.run(['go', 'run', './tools/bengali/crossfit_render'], cwd=ROOT,
                          input=''.join(json.dumps(r, ensure_ascii=False)+'\n' for r in requests),
                          text=True, capture_output=True, check=True)
    output = [json.loads(line) for line in proc.stdout.splitlines()]
    if len(output) != len(rows):
        raise ValueError('renderer count mismatch')
    choices = []
    for (word, _), request, result in zip(rows, requests, output):
        if result['check'] != result['model']:
            raise ValueError('full-model renderer parity failure: '+word)
        if len(result['variants']) != len(request['variants']):
            raise ValueError('variant count mismatch')
        model = result['variants'][0]
        if not use_folds and model != result['model']:
            raise ValueError('evaluation baseline parity failure')
        # Stable order: model, B1, then single-slot variants; remove duplicates.
        choices.append(list(dict.fromkeys([model, result['b1']] + result['variants'][1:])))
    return choices


def pair_examples(rows, choices):
    if len(rows) != len(choices):
        raise ValueError('candidate count mismatch')
    return [(features(word, alt, options[0]), target(refs, alt, options[0]))
            for (word, refs), options in zip(rows, choices) for alt in options[1:]]


def choose(tree, word, options, threshold):
    if not options:
        raise ValueError('empty candidates')
    best, best_score = options[0], threshold
    for alt in options[1:]:
        score = probability(tree, features(word, alt, options[0]))
        if score > best_score:
            best, best_score = alt, score
    return best


def candidate_stats(rows, choices):
    return {'words': len(rows), 'pairs': sum(len(c)-1 for c in choices),
            'words_with_alternatives': sum(len(c) > 1 for c in choices),
            'baseline_any': sum(c[0] in refs for (_, refs), c in zip(rows, choices)),
            'oracle_any': sum(any(s in refs for s in c) for (_, refs), c in zip(rows, choices))}


def train(google, output):
    groups = verify()
    rows = grouped_training(training_rows('dakshina'), groups['dakshina-train'])
    aligned, alignment = dataset(training_rows('google', google))
    trees, folds = [], []
    for held_fold in range(FOLDS):
        examples = fold_examples(aligned, held_fold)
        prediction_words = {w for w, _ in rows if fold(w) == held_fold}
        fitted_words = {r[2] for r in examples}
        if fitted_words & prediction_words:
            raise ValueError('candidate model sees prediction fold')
        learned = vowel_tree(examples, 10, 20)
        trees.append(learned)
        folds.append({'fold': held_fold, 'fit_types': len(fitted_words), 'fit_slots': len(examples),
                      'prediction_types': len(prediction_words),
                      'fit_keys_sha256': sha(('\n'.join(sorted(fitted_words))+'\n').encode()),
                      'prediction_keys_sha256': sha(('\n'.join(sorted(prediction_words))+'\n').encode()),
                      'tree_sha256': sha(dumps(learned))})
        print(f'fit fold {held_fold}: {len(examples)} slots', flush=True)
    choices = render(rows, trees, True)
    examples = pair_examples(rows, choices)
    selector = fit(examples, depth=DEPTH, min_leaf=MIN_LEAF)
    artifact = {'schema': 1, 'license': 'CC-BY-SA-4.0', 'tree': selector,
                'policy': {'seed': SEED, 'folds': FOLDS, 'max_single_slot_edits': MAX_EDITS,
                           'depth': DEPTH, 'min_leaf': MIN_LEAF, 'thresholds': THRESHOLDS,
                           'vowel_depth': 10, 'vowel_min_leaf': 20}}
    source_paths = [ROOT/'gomanize.go'] + [p for folder in ('core', 'lang/bengali', 'script/brahmic', 'scheme', 'tools/bengali/crossfit_render')
                                        for p in (ROOT/folder).rglob('*.go') if not p.name.endswith('_test.go')]
    source_paths += [ROOT/'tools/bengali_training_data.py'] + [ROOT/'tools/bengali'/name for name in (
        'train_crossfit_selector.py', 'train_native_selector.py', 'train_reranker.py', 'train_vowels.py', 'vowels.py', 'evaluate_external.py')]
    report = {'schema': 1, 'provenance': {'google_sha256': GOOGLE_SHA, 'dakshina_train_sha256': DAK_SHA['train'],
              'split_manifest_sha256': sha((DEFAULT/'manifest.json').read_bytes()),
              'vowel_model_sha256': sha((ROOT/'lang/bengali/vowel_tree.json').read_bytes()),
              'sources_sha256': {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in sorted(source_paths)}},
              'alignment': alignment, 'folds': folds, 'training': dict(candidate_stats(rows, choices),
              prefer_alternative_pairs=sum(y for _, y in examples)), 'artifact_sha256': sha(dumps(artifact))}
    output.mkdir(parents=True, exist_ok=True)
    (output/'selector.json').write_bytes(dumps(artifact))
    # Reproducible intermediate folds are local artifacts, not embedded models.
    (output/'fold_trees.json').write_bytes(dumps(trees))
    (output/'training_report.json').write_bytes(dumps(report))
    return artifact, report


def evaluate(artifact, split, threshold=None):
    rows = references(split)
    choices = render(rows, [], False)
    baseline = metrics(rows, [c[0] for c in choices])
    trials = []
    for cutoff in THRESHOLDS if threshold is None else (threshold,):
        predictions = [choose(artifact['tree'], word, options, cutoff) for (word, _), options in zip(rows, choices)]
        trials.append(dict(metrics(rows, predictions), threshold=cutoff,
                           changed=sum(out != options[0] for out, options in zip(predictions, choices))))
    return {'fixture_sha256': DAK_SHA[split], 'baseline': baseline,
            'candidates': candidate_stats(rows, choices), 'trials': trials}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--google', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--dev', action='store_true', help='evaluate fixed grid after freezing training artifacts')
    args = parser.parse_args()
    artifact, report = train(args.google, args.output)
    if args.dev:
        result = evaluate(artifact, 'dev')
        report['dev'] = result
        report['selected'] = select(result['baseline'], result['trials'])
        (args.output/'report.json').write_bytes(dumps(report))
        print(json.dumps({'training': report['training'], 'dev': result, 'selected': report['selected']}, indent=2))


if __name__ == '__main__':
    main()
