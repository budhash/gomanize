#!/usr/bin/env python3
"""Fit an isolated Bengali character LM and select a fixed-grid margin on dev."""
import argparse
from collections import Counter, defaultdict
import json
import math
from pathlib import Path
import subprocess
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DEFAULT, DAK_SHA, training_rows, dak_rows, sha
from evaluate_external import distance

MARGINS = (0.0, 0.05, 0.1, 0.2, 0.4, 0.8, 1.6)


def fit(rows):
    counts = Counter()
    accepted = skipped = mass = 0
    for row in rows:
        roman = row['roman'].lower()
        if not roman or not all('a' <= c <= 'z' for c in roman):
            skipped += 1
            continue
        weight = int(row['attestations'])
        if weight <= 0:
            raise ValueError('nonpositive attestation')
        accepted += 1
        mass += weight
        word = '^' + roman + '$'
        for order in range(1, 5):
            for i in range(len(word)-order+1):
                counts[word[i:i+order]] += weight
    kept = {g: n for g, n in counts.items() if n >= 3}
    return kept, {'source_rows': len(rows), 'accepted_rows': accepted, 'skipped_non_ascii_letter_rows': skipped,
                  'attestation_mass': mass, 'grams_before_pruning': len(counts), 'grams': len(kept)}


def score(counts, roman):
    roman = roman.lower()
    if not roman or not all('a' <= c <= 'z' for c in roman):
        return -math.inf
    word = '^' + roman + '$'
    total = sum(n for gram, n in counts.items() if len(gram) == 1)
    value = 0.0
    for i in range(1, len(word)):
        penalty = 1.0
        found = False
        for order in range(4, 0, -1):
            start = i-order+1
            if start < 0:
                continue
            denominator = total if order == 1 else counts.get(word[start:i], 0)
            numerator = counts.get(word[start:i+1], 0)
            if numerator and denominator:
                value += math.log(penalty * numerator / denominator)
                found = True
                break
            penalty *= 0.4
        if not found:
            value += math.log(penalty / (total+1))
    return value/(len(word)-1)


def references(split):
    grouped = defaultdict(dict)
    for row in dak_rows(split):
        grouped[row['native']][row['roman']] = int(row['attestations'])
    return sorted(grouped.items())


def candidates(rows):
    proc = subprocess.run(['go', 'run', './tools/bengali/evaluate'], cwd=ROOT,
                          input=''.join(json.dumps(word)+'\n' for word, _ in rows),
                          text=True, capture_output=True, check=True)
    values = [json.loads(line)[:2] for line in proc.stdout.splitlines()]
    if len(values) != len(rows):
        raise ValueError('candidate count mismatch')
    return values


def metrics(rows, predictions):
    if not rows or len(rows) != len(predictions):
        raise ValueError('metrics require nonempty, equally sized references and predictions')
    strict = any_match = 0
    cer = 0.0
    for (_, refs), prediction in zip(rows, predictions):
        best = min(refs, key=lambda r: (-refs[r], r))
        strict += prediction == best
        any_match += prediction in refs
        cer += min(distance(prediction, ref)/max(len(ref), 1) for ref in refs)
    return {'words': len(rows), 'strict': strict, 'any': any_match, 'CER': cer/len(rows)}


def choose(b1, model, scores, margin):
    return b1 if b1 != model and scores[b1] > scores[model]+margin else model


def select_trial(baseline, trials):
    passing = [r for r in trials if r['any'] > baseline['any']
               and r['strict'] >= baseline['strict'] and r['CER'] < baseline['CER']]
    return min(passing, key=lambda r: (-r['any'], r['CER'], -r['strict'], r['margin'])) if passing else None


def serialize(counts):
    return ''.join(f'{g}\t{counts[g]}\n' for g in sorted(counts)).encode()


def dev_results(counts):
    """Dev-only trial grid for fitted counts: (baseline, trials, selected)."""
    rows = references('dev')
    choices = candidates(rows)
    # Only differing candidates need LM scores. Compute each string once.
    strings = {s for pair in choices if pair[0] != pair[1] for s in pair}
    scores = {s: score(counts, s) for s in sorted(strings)}
    baseline = metrics(rows, [p[1] for p in choices])
    trials = []
    for margin in MARGINS:
        outputs = [choose(b1, model, scores, margin) for b1, model in choices]
        result = metrics(rows, outputs)
        result.update({'margin': margin, 'changed_from_model': sum(out != pair[1] for out, pair in zip(outputs, choices))})
        trials.append(result)
    return baseline, trials, select_trial(baseline, trials)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    counts, stats = fit(training_rows('dakshina'))
    baseline, trials, selected = dev_results(counts)
    payload = serialize(counts)
    report = {'schema': 1, 'license': 'CC-BY-SA-4.0', 'source_fixture_sha256': DAK_SHA['train'],
              'split_manifest_sha256': sha((DEFAULT/'manifest.json').read_bytes()),
              'dev_fixture_sha256': DAK_SHA['dev'],
              'metric_source_sha256': sha((ROOT/'tools/bengali/evaluate_external.py').read_bytes()),
              'training_loader_sha256': sha((ROOT/'tools/bengali_training_data.py').read_bytes()),
              'vowel_model_sha256': sha((ROOT/'lang/bengali/vowel_tree.json').read_bytes()),
              'candidate_source_sha256': {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in sorted(
                  [ROOT/'gomanize.go'] + [p for folder in ('core', 'lang/bengali', 'script/brahmic', 'scheme', 'tools/bengali/evaluate')
                                         for p in (ROOT/folder).rglob('*.go') if not p.name.endswith('_test.go')])},
              'selection_policy': 'match-any strictly higher; strict nondecreasing; CER strictly lower; ties: any, CER, strict, smallest margin',
              'evaluation_split': 'dev only; no held-out evaluation after rejection',
              'trainer_sha256': sha(Path(__file__).read_bytes()), 'ngrams_sha256': sha(payload),
              'training': stats, 'policy': 'weighted 1..4 grams; lowercase ASCII letters; minimum count 3; backoff 0.4; mean log score; model first; B1 requires strict margin improvement',
              'dev_model': baseline, 'margin_grid': MARGINS, 'dev_trials': trials, 'selected': selected}
    args.output.mkdir(parents=True, exist_ok=True)
    (args.output/'roman_ngrams.tsv').write_bytes(payload)
    (args.output/'reranker_manifest.json').write_text(json.dumps(report, indent=2)+'\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
