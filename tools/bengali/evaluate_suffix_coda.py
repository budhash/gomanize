#!/usr/bin/env python3
"""Record T-0120/T-0127: suffix-boundary deletion and per-entry loan-coda selection.

Implements docs/reviews/2026-10-20-bengali-suffix-coda-prereg.md. Selection is
train-only: each candidate table entry is ablated leave-one-out from its table
(up to three rounds, failing entries removed), then each of the two rules is
ablated from the full engine, and the five 2026-10-14 bundle rules are
re-ablated for the report. An entry or rule passes iff net match-any > 0 and
paired losses <= 0.25 x wins on Dakshina train for both the rules-only (B1)
and vowel-model profiles. Dev confirms once after freezing; test and BanglaTLit
are reported once after dev. Aksharantar is not run.

Stages: --stage train prints the selection (no held-out data); --stage dev
prints the dev confirmation; --stage all also runs test and BanglaTLit and
writes the record. Every stage requires the compiled tables to equal the
selection's survivors once the code is frozen (--stage train reports it).
"""
import argparse
import json
from pathlib import Path
import subprocess
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, DAK_SHA, sha
import evaluate_rule_bundle as bundle
from profiles import PROFILES, predict, references, strict, tables

PREREG = 'docs/reviews/2026-10-20-bengali-suffix-coda-prereg.md'
RECORD = ROOT/'docs/reviews/2026-10-20-bengali-suffix-coda.json'
HASANT = '্'
NUKTA = '়'
# The preregistered suffix candidates, verbatim.
SUF_CANDIDATES = ('কে', 'টি', 'টা', 'টার', 'টির', 'টাকে', 'টিকে', 'গুলো', 'গুলি', 'গুলোর', 'গুলির',
                  'গুলোকে', 'গুলিকে', 'তে', 'দের', 'দেরকে', 'রা', 'খানা', 'খানি')
# The loan-coda entries on main before this change (2026-10-14 bundle).
MAIN_CODA = ('নট', 'নড', 'ঞজ', 'নস', 'কস', 'রস', 'রক', 'লড')
MIN_TYPES = 5
ROUNDS = 3
RULES = {'suf': ('schwa.bengali.suffix-boundary',), 'coda': ('schwa.bengali.loan-coda',)}
ABLATED = ('B1', 'model')
LOSS_RATIO = 0.25
CRITERION = ('train, for both B1 and model: net match-any > 0 and paired losses <= 0.25 x wins; '
             'entries leave-one-out from their candidate table (up to three rounds), then each rule '
             'leave-one-out from the full engine')
CODA_DEFINITION = ('word-final C + hasant + C (each a single letter, the left not itself after a hasant) on '
                   'which main keeps the final vowel in both B1 and model, in at least 5 distinct '
                   'Dakshina train word types; plus the eight entries on main')


def passes(row):
    return all(row[p]['net'] > 0 and row[p]['losses'] <= LOSS_RATIO*row[p]['wins'] for p in ABLATED)


def letters(key):
    """Split a joined key into letters, each a base rune with an optional nukta."""
    out = []
    for ch in key:
        if ch == NUKTA and out:
            out[-1] += ch
        else:
            out.append(ch)
    return out


def entry_text(table, entry):
    """The written word ending an entry matches (coda keys gain their hasant)."""
    if table == 'coda':
        left, right = letters(entry)
        return left+HASANT+right
    return entry


def configure(coda, suf, disable=()):
    """predict() arguments that turn the compiled tables into exactly coda/suf."""
    compiled = tables()
    want = {'coda': set(coda), 'suf': set(suf)}
    drop = [f'{t}:{e}' for t in ('coda', 'suf') for e in sorted(compiled[t]-want[t])]
    add = [f'{t}:{e}' for t in ('coda', 'suf') for e in sorted(want[t]-compiled[t])]
    return {'disable': tuple(disable), 'drop': drop, 'add': add}


def run(rows, coda, suf, disable=()):
    return predict(rows, **configure(coda, suf, disable)) if rows else []


def paired(rows, full, without):
    row = {}
    for name in ABLATED:
        column = PROFILES.index(name)
        wins = losses = changed = 0
        for (_, refs), f, w in zip(rows, full, without):
            wins += f[column] in refs and w[column] not in refs
            losses += w[column] in refs and f[column] not in refs
            changed += f[column] != w[column]
        row[name] = {'wins': wins, 'losses': losses, 'net': wins-losses, 'changed': changed}
    row['passes'] = passes(row)
    return row


def coda_candidates(rows):
    """Enumerate loan-coda candidates from train word types under main's engine."""
    proc = subprocess.run(['go', 'run', './tools/bengali/final_clusters', '-coda', ','.join(MAIN_CODA)], cwd=ROOT,
                          input=''.join(json.dumps(w, ensure_ascii=False)+'\n' for w, _ in rows),
                          text=True, capture_output=True, check=True)
    found = {}
    for line in proc.stdout.splitlines():
        key, b1, model = json.loads(line)
        if key:
            stats = found.setdefault(key, {'types': 0, 'keep_types': 0})
            stats['types'] += 1
            stats['keep_types'] += b1 == 'Keep' and model == 'Keep'
    if len(proc.stdout.splitlines()) != len(rows):
        raise ValueError('final_clusters shape mismatch')
    enumerated = {k: v for k, v in sorted(found.items()) if v['keep_types'] >= MIN_TYPES and k not in MAIN_CODA}
    return {'definition': CODA_DEFINITION, 'min_types': MIN_TYPES, 'main': sorted(MAIN_CODA),
            'enumerated': enumerated, 'candidates': sorted(set(MAIN_CODA) | set(enumerated))}


def table_config(table, entries):
    """The engine an entry table is selected in: main plus that table."""
    return (entries, ()) if table == 'coda' else (MAIN_CODA, entries)


def select_entries(rows, table, candidates):
    """Leave-one-out entry rounds; returns (rounds, survivors)."""
    current = sorted(candidates)
    rounds = []
    for _ in range(ROUNDS):
        coda, suf = table_config(table, current)
        full = run(rows, coda, suf)
        touched = set()
        result = {}
        for entry in current:
            text = entry_text(table, entry)
            # Only words ending in the entry's text can change when it is dropped.
            subset = [i for i, (native, _) in enumerate(rows) if native.endswith(text)]
            touched.update(subset)
            rest = [e for e in current if e != entry]
            without = run([rows[i] for i in subset], *table_config(table, rest))
            result[entry] = paired([rows[i] for i in subset], [full[i] for i in subset], without)
        # Soundness of the subset shortcut: the empty table changes no other word.
        empty = run(rows, *table_config(table, ()))
        if any(f != e for i, (f, e) in enumerate(zip(full, empty)) if i not in touched):
            raise ValueError(f'{table}: an untouched word changed with the table')
        rounds.append(result)
        failing = [e for e, row in result.items() if not row['passes']]
        current = [e for e in current if e not in failing]
        if not failing:
            break
    return rounds, current


def selection(rows):
    coda = coda_candidates(rows)
    out = {'candidates': {'coda': coda, 'suf': list(SUF_CANDIDATES)}}
    for table, candidates in (('coda', coda['candidates']), ('suf', SUF_CANDIDATES)):
        rounds, survivors = select_entries(rows, table, candidates)
        out[table] = {'rounds': rounds, 'survivors': survivors,
                      'removed': sorted(set(candidates)-set(survivors))}
    return out


def rule_ablation(rows):
    """Per-rule leave-one-out from the shipped engine (compiled tables)."""
    full = predict(rows)
    rules = {key: {'rules': list(names), **paired(rows, full, predict(rows, names))} for key, names in RULES.items()}
    reablation = {key: {'rules': list(names), **paired(rows, full, predict(rows, names))}
                  for key, names in bundle.BUNDLE.items()}
    return full, rules, reablation


def main_config():
    """The engine on main: the eight loan-coda entries and no suffix entries."""
    return configure(MAIN_CODA, ())


def profile_scores(rows, values, columns=PROFILES):
    return {name: strict(rows, values, PROFILES.index(name)) for name in columns}


def frozen(sel):
    compiled = tables()
    return {t: sorted(compiled[t]) == sorted(sel[t]['survivors']) for t in ('coda', 'suf')}


def train_section():
    rows = references('train')
    sel = selection(rows)
    full, rules, reablation = rule_ablation(rows)
    before = predict(rows, **main_config())
    return {'selection': sel, 'rule_ablation': rules, 'bundle_reablation': reablation,
            'before': profile_scores(rows, before, ABLATED), 'after': profile_scores(rows, full, ABLATED)}


def heldout_section(split):
    rows = references(split)
    before, after = predict(rows, **main_config()), predict(rows)
    # Held-out words are excluded from the lexicon, so it must not move them.
    if not all(p[0] == p[2] and p[1] == p[3] for p in before+after):
        raise ValueError(f'lexicon changed a {split} word')
    return {'before': profile_scores(rows, before), 'after': profile_scores(rows, after)}


def banglatlit_section():
    rows = bundle.banglatlit_rows()
    groups = bundle.verify()
    training = groups['google-train'] | groups['dakshina-train']
    unseen = [i for i, r in enumerate(rows) if (tokens := set(bundle.bengali_tokens(r[0]))) and not tokens & training]

    def section(values):
        return {'full_sentences': bundle.score(rows, values),
                'all_tokens_unseen_sentences': bundle.score([rows[i] for i in unseen], [values[i] for i in unseen])}
    return {'before': section(predict(rows, **main_config())), 'after': section(predict(rows))}


def check_train(train):
    sel = train['selection']
    if not all(frozen(sel).values()):
        raise SystemExit(f'compiled tables differ from the selection survivors: {frozen(sel)}; freeze the code')
    failing = [k for k, row in train['rule_ablation'].items() if not row['passes']]
    if failing:
        raise SystemExit(f'rules fail the preregistered criterion: {failing}; remove them from the code')


def improves(section):
    return all(section['after'][p]['any'] > section['before'][p]['any'] for p in ABLATED)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--stage', choices=('train', 'dev', 'all'), required=True)
    parser.add_argument('--output', type=Path, default=RECORD)
    args = parser.parse_args()
    if args.stage == 'train':
        train = train_section()
        print(json.dumps({**train, 'frozen': frozen(train['selection'])}, ensure_ascii=False, indent=2))
        return
    if args.stage == 'dev':
        dev = heldout_section('dev')
        print(json.dumps({**dev, 'improves_both': improves(dev)}, indent=2))
        return
    train = train_section()
    check_train(train)
    dev = heldout_section('dev')
    if not improves(dev):
        raise SystemExit('dev confirmation failed; stop and report (nothing is iterated on dev)')
    result = {
        'preregistration': PREREG,
        'criterion': CRITERION,
        'Dakshina': {'train': train, 'dev': dev, 'test': heldout_section('test')},
        'BanglaTLit': banglatlit_section(),
        'before_definition': ('main before this change: loan-coda table of the eight 2026-10-14 entries, '
                              'no suffix entries (equals the pre-change runtime)'),
        'inputs_sha256': {
            **{f'benchmark/data/bengali/{s}.csv.gz': DAK_SHA[s] for s in ('train', 'dev', 'test')},
            'benchmark/data/banglatlit/test.csv.gz': bundle.BANGLATLIT_SHA,
            **{f'lang/bengali/{name}': sha((ROOT/'lang/bengali'/name).read_bytes())
               for name in ('rules.go', 'vowel_tree.json', 'lexicon.tsv')},
            'internal/bengalitables/tables.go': sha((ROOT/'internal/bengalitables/tables.go').read_bytes())},
    }
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True)+'\n')


if __name__ == '__main__':
    main()
