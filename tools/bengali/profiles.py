#!/usr/bin/env python3
"""Render and score the shipped Bengali profiles through the Go runtime; never fit or select."""
from collections import defaultdict
import json
import subprocess
from bengali_training_data import ROOT, dak_rows
from evaluate_external import distance

PROFILES = ('B1', 'model', 'lexicon', 'model_lexicon')
# Build tag that swaps in the pre-T-0121 simple-word vowel model
# (lang/bengali/vowel_model_v1.go); evaluation only.
SIMPLE_V1 = ('bengali_vowel_v1',)


def predict(rows, disable=(), tags=()):
    """Outputs per row for PROFILES; disable names rules to ablate in every profile,
    tags selects Go build tags (SIMPLE_V1 = the previous vowel model)."""
    command = ['go', 'run'] + (['-tags', ','.join(tags)] if tags else []) + ['./tools/bengali/render_profiles']
    if disable:
        command += ['-disable', ','.join(disable)]
    proc = subprocess.run(command, cwd=ROOT,
                          input=''.join(json.dumps(w, ensure_ascii=False)+'\n' for w, _ in rows),
                          text=True, capture_output=True, check=True)
    values = [json.loads(line) for line in proc.stdout.splitlines()]
    if len(values) != len(rows) or any(len(p) != len(PROFILES) for p in values):
        raise ValueError('prediction shape mismatch')
    return values


def references(split):
    """Dakshina (native, {roman: attestations}) rows for one split, sorted by native."""
    grouped = defaultdict(dict)
    for row in dak_rows(split):
        grouped[row['native']][row['roman']] = int(row['attestations'])
    return sorted(grouped.items())


def score(rows, values):
    """Per-profile match-any, accuracy and macro minimum CER (record schema)."""
    result = {'items': len(rows)}
    for column, name in enumerate(PROFILES):
        matches = sum(p[column] in refs for (_, refs), p in zip(rows, values))
        cer = sum(min(distance(p[column], ref)/max(len(ref), 1) for ref in refs)
                  for (_, refs), p in zip(rows, values))
        result[name] = {'match_any': matches, 'accuracy': matches/len(rows) if rows else None,
                        'macro_minCER': cer/len(rows) if rows else None}
    return result


def strict(rows, values, column):
    """Strict (top-attested reference), match-any and macro minimum CER."""
    top = any_match = 0
    cer = 0.0
    for (_, refs), p in zip(rows, values):
        top += p[column] == min(refs, key=lambda r: (-refs[r], r))
        any_match += p[column] in refs
        cer += min(distance(p[column], ref)/max(len(ref), 1) for ref in refs)
    return {'words': len(rows), 'strict': top, 'any': any_match, 'CER': cer/len(rows)}
