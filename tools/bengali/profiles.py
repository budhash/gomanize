#!/usr/bin/env python3
"""Render and score the shipped Bengali profiles through the Go runtime; never fit or select."""
from collections import defaultdict
import atexit
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
from bengali_training_data import ROOT, dak_rows
from evaluate_external import distance

PROFILES = ('B1', 'model', 'lexicon', 'model_lexicon')


_binary = []


def render_binary():
    """Build render_profiles once per process (go run would rebuild per call)."""
    if not _binary:
        directory = tempfile.mkdtemp(prefix='render_profiles')
        atexit.register(shutil.rmtree, directory, True)
        path = Path(directory)/'render_profiles'
        subprocess.run(['go', 'build', '-o', str(path), './tools/bengali/render_profiles'], cwd=ROOT, check=True)
        _binary.append(str(path))
    return _binary[0]


def tables():
    """The rule tables compiled into the engine: {'coda': [...], 'suf': [...]}."""
    proc = subprocess.run([render_binary(), '-list-tables'], cwd=ROOT, text=True, capture_output=True, check=True)
    return {name: set(entries) for name, entries in json.loads(proc.stdout).items()}


def predict(rows, disable=(), drop=(), add=()):
    """Outputs per row for PROFILES; disable names rules to ablate in every
    profile; drop and add name 'table:entry' items to edit the rule tables."""
    command = [render_binary()]
    if disable:
        command += ['-disable', ','.join(disable)]
    if drop:
        command += ['-drop-entry', ','.join(drop)]
    if add:
        command += ['-add-entry', ','.join(add)]
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
