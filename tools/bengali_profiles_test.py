"""Held-out and external guards for the shipped Bengali profiles.

Recomputes the vowel-conjunct record (docs/reviews/2026-10-21-bengali-vowel-conjuncts.json):
the Dakshina train selection that shipped the halant-cluster vowel model
(paired against the previous simple-word model, build tag bengali_vowel_v1),
Dakshina dev/test before and after, and BanglaTLit for the four shipped
profiles, and requires them to match. The 2026-10-14 rule-bundle and
2026-10-05 runtime records are history. Aksharantar needs the pinned external
archive and remains an offline step.
"""
import json
import math
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from bengali_training_data import ROOT, DAK_SHA, sha
import evaluate_vowel_conjuncts as conjuncts

RECORD = ROOT/'docs/reviews/2026-10-21-bengali-vowel-conjuncts.json'


class BengaliProfilesTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.record = json.loads(RECORD.read_text())

    def assert_close(self, got, want, path):
        if isinstance(want, dict):
            self.assertEqual(set(got), set(want), path)
            for key in want:
                self.assert_close(got[key], want[key], path+'.'+key)
        elif isinstance(want, float):
            self.assertTrue(math.isclose(got, want, rel_tol=1e-12, abs_tol=1e-15), f'{path}: {got} != {want}')
        else:
            self.assertEqual(got, want, path)

    def test_record_inputs_and_criterion(self):
        self.assertEqual(self.record['criterion'], conjuncts.CRITERION)
        self.assertEqual(self.record['confirmation'], conjuncts.CONFIRMATION)
        for split in ('train', 'dev', 'test'):
            self.assertEqual(self.record['inputs_sha256'][f'benchmark/data/bengali/{split}.csv.gz'], DAK_SHA[split])
        # The record describes the shipped model, inventory and rules.
        self.assertEqual(self.record['inputs_sha256'], conjuncts.inputs())
        self.assertTrue(self.record['Dakshina']['train']['ships'])
        self.assertTrue(self.record['Dakshina']['dev']['confirmed'])
        self.assertEqual(self.record['inputs_sha256']['lang/bengali/testdata/vowel_tree_simple_v1.json'],
                         sha((ROOT/'lang/bengali/testdata/vowel_tree_simple_v1.json').read_bytes()))

    def test_train_selection(self):
        # The retrained model must still win on train, paired, per eligibility stratum.
        self.assert_close(json.loads(json.dumps(conjuncts.train_section())), self.record['Dakshina']['train'], 'train')

    def test_dakshina_heldout_profiles(self):
        dev = conjuncts.heldout_section('dev')
        dev['confirmed'] = conjuncts.confirmed(dev)
        self.assert_close(dev, self.record['Dakshina']['dev'], 'dev')
        self.assert_close(conjuncts.heldout_section('test'), self.record['Dakshina']['test'], 'test')

    def test_banglatlit_profiles(self):
        self.assert_close(conjuncts.banglatlit_section(), self.record['BanglaTLit'], 'BanglaTLit')


if __name__ == '__main__':
    unittest.main()
