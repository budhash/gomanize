#!/usr/bin/env python3
"""Evaluate only a frozen dev-selected crossfit selector on the held-out test."""
import argparse
import json
from pathlib import Path
from train_crossfit_selector import ROOT, DEFAULT, sha, dumps, evaluate, select, THRESHOLDS


def frozen_selection(artifact, report):
    if sha(dumps(artifact)) != report['artifact_sha256']:
        raise ValueError('selector artifact changed after dev selection')
    if report.get('selected') is None:
        raise ValueError('dev did not select a candidate')
    dev = report['dev']
    selected = select(dev['baseline'], dev['trials'])
    if selected != report['selected'] or selected['threshold'] not in THRESHOLDS:
        raise ValueError('selection differs from the dev gate')
    return selected['threshold']


def verify_sources(report):
    provenance = report['provenance']
    for name, expected in provenance['sources_sha256'].items():
        if sha((ROOT/name).read_bytes()) != expected:
            raise ValueError('experiment source changed: '+name)
    for path, expected in ((ROOT/'lang/bengali/vowel_tree.json', provenance['vowel_model_sha256']),
                           (DEFAULT/'manifest.json', provenance['split_manifest_sha256'])):
        if sha(path.read_bytes()) != expected:
            raise ValueError('model or split changed')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--experiment', type=Path, required=True)
    args = parser.parse_args()
    artifact = json.loads((args.experiment/'selector.json').read_text())
    report = json.loads((args.experiment/'report.json').read_text())
    threshold = frozen_selection(artifact, report)
    verify_sources(report)
    result = {'artifact_sha256': report['artifact_sha256'], 'dev_report_sha256': sha(dumps(report)),
              'evaluator_sha256': sha(Path(__file__).read_bytes()),
              'frozen_threshold': threshold, 'test': evaluate(artifact, 'test', threshold)}
    (args.experiment/'heldout.json').write_bytes(dumps(result))
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
