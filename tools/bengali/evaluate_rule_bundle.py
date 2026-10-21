#!/usr/bin/env python3
"""Record the T-0119 Bengali rule bundle: train-only ablation, then held-out scores.

Selection is preregistered and train-only: a bundle rule ships iff disabling it
alone from the bundle loses net match-any on Dakshina train, with paired losses
at most a quarter of wins, for both the rules-only (B1) and vowel-model
profiles. Dev and test are reported before/after and never select; BanglaTLit is
reported after. Aksharantar is not run.

History since 2026-10-21: the T-0121 vowel-model retrain moved the model
profiles, so this script no longer reproduces its record; the shipped profiles
are asserted against docs/reviews/2026-10-21-bengali-vowel-conjuncts.json.
"""
import argparse
import csv
import gzip
import io
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DAK_SHA, pinned, sha, verify
from audit_bengali_training_overlap import bengali_tokens
from profiles import PROFILES, predict, references, score, strict

BANGLATLIT_SHA = '386721b4221e2d93b16ab73f3761c7c4e4cd295adf1160fe9c95c06cb34c5386'
# Key -> catalog rules it owns (lang/bengali/rules.go); ablated together.
BUNDLE = {
    'ito': ('schwa.bengali.final-ito',),
    'tacons': ('schwa.bengali.final-ta',),
    'coda': ('schwa.bengali.loan-coda',),
    'osuf': ('vowel.bengali.final-o',),
    'yph': ('consonant.bengali.sibilant-ya', 'render.bengali.sibilant-ya'),
}
# Measured once at selection, leave-one-out from the six-rule candidate bundle,
# then removed from the code; history, not recomputable.
REJECTED = {'bph': {
    'mechanism': 'ba-phala after শ/স renders w; medial দ্ব renders db',
    'B1': {'wins': 37, 'losses': 30, 'net': 7, 'changed': 346},
    'model': {'wins': 37, 'losses': 30, 'net': 7, 'changed': 346}}}
ABLATED = ('B1', 'model')
LOSS_RATIO = 0.25
CRITERION = ('train leave-one-out from the full bundle, for both B1 and model: '
             'net match-any > 0 and paired losses <= 0.25 x wins')


def ships(row):
    return all(row[p]['net'] > 0 and row[p]['losses'] <= LOSS_RATIO*row[p]['wins'] for p in ABLATED)


def ablation(rows, full):
    """Paired match-any wins/losses of the full bundle against it minus each rule."""
    table = {}
    for key, rules in BUNDLE.items():
        without = predict(rows, rules)
        row = {'rules': list(rules)}
        for name in ABLATED:
            column = PROFILES.index(name)
            wins = losses = changed = 0
            for (_, refs), f, w in zip(rows, full, without):
                wins += f[column] in refs and w[column] not in refs
                losses += w[column] in refs and f[column] not in refs
                changed += f[column] != w[column]
            row[name] = {'wins': wins, 'losses': losses, 'net': wins-losses, 'changed': changed}
        row['shipped'] = ships(row)
        table[key] = row
    return table


def profile_scores(rows, values, columns=PROFILES):
    return {name: strict(rows, values, PROFILES.index(name)) for name in columns}


def all_rules():
    return [r for rules in BUNDLE.values() for r in rules]


def train_section():
    rows = references('train')
    full = predict(rows)
    before = predict(rows, all_rules())
    return {'ablation': ablation(rows, full),
            'before': profile_scores(rows, before, ABLATED), 'after': profile_scores(rows, full, ABLATED)}


def heldout_section(split):
    rows = references(split)
    before, after = predict(rows, all_rules()), predict(rows)
    # Held-out words are excluded from the lexicon, so it must not move them.
    if not all(p[0] == p[2] and p[1] == p[3] for p in before+after):
        raise ValueError(f'lexicon changed a {split} word')
    return {'before': profile_scores(rows, before), 'after': profile_scores(rows, after)}


def banglatlit_rows():
    blob = pinned(ROOT/'benchmark/data/banglatlit/test.csv.gz', BANGLATLIT_SHA)
    return [(r['native'], [r['roman']]) for r in csv.DictReader(io.StringIO(gzip.decompress(blob).decode()))]


def banglatlit_section():
    rows = banglatlit_rows()
    groups = verify()
    training = groups['google-train'] | groups['dakshina-train']
    unseen = [i for i, r in enumerate(rows) if (tokens := set(bengali_tokens(r[0]))) and not tokens & training]

    def section(values):
        return {'full_sentences': score(rows, values),
                'all_tokens_unseen_sentences': score([rows[i] for i in unseen], [values[i] for i in unseen])}
    return {'before': section(predict(rows, all_rules())), 'after': section(predict(rows))}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=ROOT/'docs/reviews/2026-10-14-bengali-rule-bundle.json')
    args = parser.parse_args()
    train = train_section()
    rejected = [k for k, row in train['ablation'].items() if not row['shipped']]
    if rejected:
        raise SystemExit(f'bundle rules fail the preregistered criterion: {rejected}; remove them from the code')
    result = {
        'criterion': CRITERION,
        'rejected_at_selection': REJECTED,
        'Dakshina': {'train': train, 'dev': heldout_section('dev'), 'test': heldout_section('test')},
        'BanglaTLit': banglatlit_section(),
        'before_definition': 'same engine with every bundle rule disabled (equals the pre-bundle runtime)',
        'inputs_sha256': {
            **{f'benchmark/data/bengali/{s}.csv.gz': DAK_SHA[s] for s in ('train', 'dev', 'test')},
            'benchmark/data/banglatlit/test.csv.gz': BANGLATLIT_SHA,
            **{f'lang/bengali/{name}': sha((ROOT/'lang/bengali'/name).read_bytes())
               for name in ('rules.go', 'vowel_tree.json', 'lexicon.tsv')}},
    }
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True)+'\n')
    print(json.dumps(train['ablation'], indent=2))


if __name__ == '__main__':
    main()
