"""Held-out and external guards for the shipped Bengali profiles.

Recomputes the Dakshina dev/test and BanglaTLit sections of the B2 runtime
record (docs/reviews/2026-10-05-bengali-b2-runtime.json) for the four shipped
profiles and requires them to match. The selector rows of that record are
history (retired 2026-10-13) and are not checked. Aksharantar needs the pinned
external archive and remains an offline reproduction step.
"""
from collections import defaultdict
import csv
import gzip
import io
import json
import math
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from profiles import PROFILES, predict
from bengali_training_data import ROOT, dak_rows, pinned, verify
from evaluate_external import distance
from audit_bengali_training_overlap import bengali_tokens

RECORD = ROOT/'docs/reviews/2026-10-05-bengali-b2-runtime.json'
BANGLATLIT_SHA = '386721b4221e2d93b16ab73f3761c7c4e4cd295adf1160fe9c95c06cb34c5386'


def references(split):
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


class BengaliProfilesTest(unittest.TestCase):
    def assert_close(self, got, want, path):
        if isinstance(want, dict):
            for key in want:
                self.assert_close(got[key], want[key], path+'.'+key)
        elif isinstance(want, float):
            self.assertTrue(math.isclose(got, want, rel_tol=1e-12, abs_tol=1e-15), f'{path}: {got} != {want}')
        else:
            self.assertEqual(got, want, path)

    def shipped(self, section):
        return {k: v for k, v in section.items() if k in PROFILES or k == 'items'}

    def test_dakshina_heldout_profiles(self):
        record = json.loads(RECORD.read_text())['Dakshina']
        for split in ('dev', 'test'):
            rows = references(split)
            values = predict(rows)
            # Held-out words are excluded from the lexicon, so it must not move them.
            self.assertTrue(all(p[0] == p[2] and p[1] == p[3] for p in values), split)
            self.assert_close(score(rows, values), self.shipped(record[split]), split)
            for column, name in enumerate(PROFILES):
                self.assert_close(strict(rows, values, column), record[split]['strict_profiles'][name],
                                  f'{split}.strict.{name}')

    def test_banglatlit_profiles(self):
        record = json.loads(RECORD.read_text())['BanglaTLit']
        blob = pinned(ROOT/'benchmark/data/banglatlit/test.csv.gz', BANGLATLIT_SHA)
        rows = [(r['native'], [r['roman']]) for r in csv.DictReader(io.StringIO(gzip.decompress(blob).decode()))]
        values = predict(rows)
        groups = verify()
        training = groups['google-train'] | groups['dakshina-train']
        unseen = [(r, p) for r, p in zip(rows, values)
                  if (tokens := set(bengali_tokens(r[0]))) and not tokens & training]
        self.assert_close(score(rows, values), self.shipped(record['full_sentences']), 'full')
        self.assert_close(score([r for r, _ in unseen], [p for _, p in unseen]),
                          self.shipped(record['all_tokens_unseen_sentences']), 'unseen')


if __name__ == '__main__':
    unittest.main()
