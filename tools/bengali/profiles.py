#!/usr/bin/env python3
"""Render the shipped Bengali profiles through the Go runtime; never fit or select."""
import json
import subprocess
from bengali_training_data import ROOT

PROFILES = ('B1', 'model', 'lexicon', 'model_lexicon')


def predict(rows):
    proc = subprocess.run(['go', 'run', './tools/bengali/render_profiles'], cwd=ROOT,
                          input=''.join(json.dumps(w, ensure_ascii=False)+'\n' for w, _ in rows),
                          text=True, capture_output=True, check=True)
    values = [json.loads(line) for line in proc.stdout.splitlines()]
    if len(values) != len(rows) or any(len(p) != len(PROFILES) for p in values):
        raise ValueError('prediction shape mismatch')
    return values
