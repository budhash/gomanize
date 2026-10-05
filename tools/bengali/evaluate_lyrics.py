#!/usr/bin/env python3
"""Report B3 pilot agreement with unreviewed references; never fit or set a gate."""
import argparse
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_lyrics import ROOT, DATA, verify, sha, STATUS
from bengali_training_data import verify as training_groups
from audit_bengali_training_overlap import bengali_tokens, overlap
from evaluate_external import distance
from evaluate_runtime import PROFILES, predict


def score(rows, predictions):
    if len(rows) != len(predictions) or any(len(p) != len(PROFILES) or
            any(not isinstance(v, str) for v in p) for p in predictions):
        raise ValueError('prediction shape mismatch')
    if any(not r['roman'] for r in rows):
        raise ValueError('empty reference')
    result = {'lines': len(rows)}
    errors = [[distance(v, r['roman']) for v in p] for r, p in zip(rows, predictions)]
    chars = sum(len(r['roman']) for r in rows)
    for i, name in enumerate(PROFILES):
        edits = sum(e[i] for e in errors)
        exact = sum(e[i] == 0 for e in errors)
        result[name] = {'exact_lines': exact, 'exact_fraction': exact/len(rows) if rows else None,
                        'macro_line_CER': sum(e[i]/len(r['roman']) for r, e in zip(rows, errors))/len(rows) if rows else None,
                        'micro_CER': edits/chars if chars else None, 'edits': edits, 'reference_characters': chars}
    for a, b, name in ((1, 4, 'rerank_vs_model'), (3, 5, 'rerank_lexicon_vs_model_lexicon')):
        result[name] = {'changed': sum(p[a] != p[b] for p in predictions),
                        'CER_better': sum(e[b] < e[a] for e in errors),
                        'CER_worse': sum(e[b] > e[a] for e in errors),
                        'CER_equal': sum(e[b] == e[a] for e in errors),
                        'exact_wins': sum(e[b] == 0 < e[a] for e in errors),
                        'exact_losses': sum(e[a] == 0 < e[b] for e in errors)}
    return result


def evaluate(rows, predictions, groups):
    # Audit word exposure, not the unknowable source exposure of the assistant.
    training = groups['google-train'] | groups['dakshina-train']
    tokens = [list(bengali_tokens(r['native'])) for r in rows]
    unique = list(dict.fromkeys(r['native'] for r in rows))
    first = {r['native']: i for i, r in reversed(list(enumerate(rows)))}
    subsets = {'all_lines': list(range(len(rows))), 'unique_native_lines': [first[w] for w in unique],
               'all_tokens_unseen_lines': [i for i, words in enumerate(tokens) if words and not set(words) & training]}
    subsets.update({f'song_{s}': [i for i, r in enumerate(rows) if r['song'] == s] for s in sorted({r['song'] for r in rows})})
    scores = {name: score([rows[i] for i in indexes], [predictions[i] for i in indexes]) for name, indexes in subsets.items()}
    songs = [v for k, v in scores.items() if k.startswith('song_')]
    return {'scores': scores, 'equal_weight_song_macro_CER': {
        p: sum(s[p]['macro_line_CER'] for s in songs)/len(songs) if songs else None for p in PROFILES},
        'overlap': {**overlap(set(w for words in tokens for w in words), groups),
                    'token_occurrences': sum(map(len, tokens)),
                    'training_token_occurrences': sum(w in training for words in tokens for w in words),
                    'all_tokens_unseen_lines': len(subsets['all_tokens_unseen_lines']),
                    'definition': 'NFC/Cf normalized Bengali letter/mark runs; absent from union of frozen Google/Dakshina train vocabularies'},
        'predictions': [{'id': r['id'], 'native': r['native'], 'reference': r['roman'],
                         'training_tokens': sorted(set(words) & training),
                         'profiles': dict(zip(PROFILES, p))} for r, words, p in zip(rows, tokens, predictions)]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    rows, manifest = verify()
    predictions = predict([(r['native'], [r['roman']]) for r in rows])
    result = evaluate(rows, predictions, training_groups())
    sources = [ROOT/name for name in ('gomanize.go', 'tools/bengali/evaluate_lyrics.py',
        'tools/bengali_lyrics.py', 'tools/bengali/evaluate_runtime.py',
        'tools/bengali/evaluate_external.py', 'tools/audit_bengali_training_overlap.py',
        'tools/bengali_training_data.py', 'tools/build_bengali.py',
        'training/data/bengali/manifest.json')] + [p for folder in (
        'core', 'lang/bengali', 'script/brahmic', 'scheme', 'tools/bengali/evaluate_runtime')
        for p in (ROOT/folder).rglob('*.go') if not p.name.endswith('_test.go')]
    result.update({'reference_status': STATUS, 'fixture_sha256': manifest['fixture_sha256'],
        'manifest_sha256': sha((DATA/'manifest.json').read_bytes()),
        'sources_sha256': {str(p.relative_to(ROOT)): sha(p.read_bytes()) for p in sorted(sources)},
        'artifacts_sha256': {name: sha((ROOT/'lang/bengali'/name).read_bytes()) for name in ('vowel_tree.json','selector.json','lexicon.tsv')},
        'limitations': ['Assistant-authored single references, not human-attested gold or an accuracy gate.',
            'Four consecutive Tagore songs only; repeated refrains count in all_lines.',
            'Exact Unicode character scoring includes punctuation and spaces; no spelling-variant acceptance.',
            'Training word overlap audited; full-line independence and assistant pretraining exposure are not established.',
            'Evaluation only: no model, rule, lexicon, reference or threshold selected using these predictions.']})
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({'reference_status': STATUS, 'all_lines': result['scores']['all_lines'], 'overlap': result['overlap']}, indent=2))


if __name__ == '__main__':
    main()
