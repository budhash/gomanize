#!/usr/bin/env python3
"""Record T-0121: the Bengali vowel model extended to halant-cluster words.

Preregistered in docs/reviews/2026-10-21-bengali-vowel-conjuncts-prereg.md.
Selection is Dakshina train only: the retrained model profile ships iff, paired
against the previous simple-word model (build tag bengali_vowel_v1), net
match-any > 0 and macro minimum CER is not worse. Wins and losses are split by
words newly eligible and words eligible before. Dakshina dev is confirmation
only (model match-any must rise, minCER must not worsen); test and BanglaTLit
are reported after it. Aksharantar is not run. Nothing here fits or tunes.
"""
import argparse
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DAK_SHA, sha, verify
from audit_bengali_training_overlap import bengali_tokens
from evaluate_external import distance
from evaluate_rule_bundle import BANGLATLIT_SHA, banglatlit_rows
from profiles import PROFILES, SIMPLE_V1, predict, references, score, strict
from vowels import eligible

CRITERION = ('Dakshina train, model profile paired against the previous simple-word model: '
             'net match-any > 0 and macro minCER not worse')
CONFIRMATION = 'Dakshina dev, run once: model match-any rises and macro minCER is not worse'
STRATA = ('newly_eligible', 'previously_eligible', 'ineligible')
MODEL = PROFILES.index('model')


def stratum(native):
    if eligible(native, 1):
        return 'previously_eligible'
    return 'newly_eligible' if eligible(native) else 'ineligible'


def cer(output, refs):
    return min(distance(output, ref)/max(len(ref), 1) for ref in refs)


def paired(rows, before, after, column=MODEL):
    """Paired match-any wins/losses of after against before, overall and per stratum."""
    table = {name: {'words': 0, 'wins': 0, 'losses': 0, 'changed': 0} for name in ('all',) + STRATA}
    for (native, refs), b, a in zip(rows, before, after):
        for name in ('all', stratum(native)):
            row = table[name]
            row['words'] += 1
            row['wins'] += a[column] in refs and b[column] not in refs
            row['losses'] += b[column] in refs and a[column] not in refs
            row['changed'] += a[column] != b[column]
    for row in table.values():
        row['net'] = row['wins'] - row['losses']
    return table


def stratified(rows, values):
    """Strict / match-any / minCER per eligibility stratum for every profile."""
    out = {}
    for name in STRATA:
        keep = [i for i, (native, _) in enumerate(rows) if stratum(native) == name]
        sub = [rows[i] for i in keep]
        vals = [values[i] for i in keep]
        out[name] = {p: strict(sub, vals, PROFILES.index(p)) if sub else {'words': 0} for p in PROFILES}
    return out


def profile_scores(rows, values):
    return {name: strict(rows, values, PROFILES.index(name)) for name in PROFILES}


def unchanged_rules_profiles(before, after, split):
    # B1 and B1 + lexicon do not consult the model; the retrain must not move them.
    if any(b[0] != a[0] or b[2] != a[2] for b, a in zip(before, after)):
        raise ValueError(f'rules-only output changed on {split}')


def train_section():
    rows = references('train')
    before, after = predict(rows, tags=SIMPLE_V1), predict(rows)
    unchanged_rules_profiles(before, after, 'train')
    model_before = strict(rows, before, MODEL)
    model_after = strict(rows, after, MODEL)
    pairs = paired(rows, before, after)
    ships = pairs['all']['net'] > 0 and model_after['CER'] <= model_before['CER']
    return {'paired_model': pairs, 'before': profile_scores(rows, before), 'after': profile_scores(rows, after),
            'stratified_after': stratified(rows, after), 'ships': ships}


def heldout_section(split):
    rows = references(split)
    before, after = predict(rows, tags=SIMPLE_V1), predict(rows)
    unchanged_rules_profiles(before, after, split)
    # Held-out words are excluded from the lexicon, so it must not move them.
    if not all(p[0] == p[2] and p[1] == p[3] for p in before+after):
        raise ValueError(f'lexicon changed a {split} word')
    return {'before': profile_scores(rows, before), 'after': profile_scores(rows, after),
            'paired_model': paired(rows, before, after),
            'stratified_before': stratified(rows, before), 'stratified_after': stratified(rows, after)}


def confirmed(dev):
    b, a = dev['before']['model'], dev['after']['model']
    return a['any'] > b['any'] and a['CER'] <= b['CER']


def banglatlit_section():
    rows = banglatlit_rows()
    groups = verify()
    training = groups['google-train'] | groups['dakshina-train']
    unseen = [i for i, r in enumerate(rows) if (tokens := set(bengali_tokens(r[0]))) and not tokens & training]
    tokens = [t for native, _ in rows for t in bengali_tokens(native)]
    coverage = {name: sum(stratum(t) == name for t in tokens) for name in STRATA}
    coverage['tokens'] = len(tokens)
    coverage['sentences_with_newly_eligible_token'] = sum(
        any(stratum(t) == 'newly_eligible' for t in bengali_tokens(native)) for native, _ in rows)

    def section(values):
        return {'full_sentences': score(rows, values),
                'all_tokens_unseen_sentences': score([rows[i] for i in unseen], [values[i] for i in unseen])}
    before, after = predict(rows, tags=SIMPLE_V1), predict(rows)
    unchanged_rules_profiles(before, after, 'BanglaTLit')
    return {'token_coverage': coverage, 'before': section(before), 'after': section(after)}


def inputs():
    return {**{f'benchmark/data/bengali/{s}.csv.gz': DAK_SHA[s] for s in ('train', 'dev', 'test')},
            'benchmark/data/banglatlit/test.csv.gz': BANGLATLIT_SHA,
            **{f'lang/bengali/{name}': sha((ROOT/'lang/bengali'/name).read_bytes())
               for name in ('rules.go', 'vowel_model.go', 'vowel_inventory.json', 'vowel_tree.json',
                            'testdata/vowel_tree_simple_v1.json', 'lexicon.tsv')}}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=ROOT/'docs/reviews/2026-10-21-bengali-vowel-conjuncts.json')
    parser.add_argument('--stage', choices=('train', 'dev', 'all'), default='all',
                        help='train: selection only; dev: selection then the one dev confirmation; all: everything')
    args = parser.parse_args()
    result = {'criterion': CRITERION, 'confirmation': CONFIRMATION,
              'before_definition': 'same engine built with -tags bengali_vowel_v1: the simple-word model shipped at 46c2967',
              'Dakshina': {'train': train_section()}}
    print(json.dumps(result['Dakshina']['train']['paired_model'], indent=2))
    if not result['Dakshina']['train']['ships'] or args.stage == 'train':
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True)+'\n')
        raise SystemExit(0 if result['Dakshina']['train']['ships'] else 'train criterion failed; stop')
    dev = heldout_section('dev')
    dev['confirmed'] = confirmed(dev)
    result['Dakshina']['dev'] = dev
    print(json.dumps({'dev_model': {'before': dev['before']['model'], 'after': dev['after']['model']}}, indent=2))
    if not dev['confirmed'] or args.stage == 'dev':
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True)+'\n')
        raise SystemExit(0 if dev['confirmed'] else 'dev confirmation failed; stop')
    result['Dakshina']['test'] = heldout_section('test')
    result['BanglaTLit'] = banglatlit_section()
    result['inputs_sha256'] = inputs()
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True)+'\n')


if __name__ == '__main__':
    main()
