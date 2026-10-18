"""Held-out and external guards for the shipped Bengali profiles.

Recomputes the rule-bundle record (docs/reviews/2026-10-14-bengali-rule-bundle.json):
the train leave-one-out ablation that selected each bundle rule, Dakshina
dev/test before and after the bundle, and BanglaTLit for the four shipped
profiles, and requires them to match. The 2026-10-05 runtime record is history.
Aksharantar needs the pinned external archive and remains an offline step.
"""
import json
import math
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from bengali_training_data import ROOT, DAK_SHA
import evaluate_rule_bundle as bundle

RECORD = ROOT/'docs/reviews/2026-10-14-bengali-rule-bundle.json'


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
        self.assertEqual(self.record['criterion'], bundle.CRITERION)
        for split in ('train', 'dev', 'test'):
            self.assertEqual(self.record['inputs_sha256'][f'benchmark/data/bengali/{split}.csv.gz'], DAK_SHA[split])
        self.assertTrue(all(row['shipped'] for row in self.record['Dakshina']['train']['ablation'].values()))
        self.assertFalse(set(bundle.REJECTED) & set(bundle.BUNDLE))

    def test_train_ablation(self):
        # Every bundle rule must still earn its place on train, alone and paired.
        self.assert_close(json.loads(json.dumps(bundle.train_section())), self.record['Dakshina']['train'], 'train')

    def test_dakshina_heldout_profiles(self):
        for split in ('dev', 'test'):
            self.assert_close(bundle.heldout_section(split), self.record['Dakshina'][split], split)

    def test_banglatlit_profiles(self):
        self.assert_close(bundle.banglatlit_section(), self.record['BanglaTLit'], 'BanglaTLit')


if __name__ == '__main__':
    unittest.main()
