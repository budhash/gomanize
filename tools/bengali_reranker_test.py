import json
import math
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from train_reranker import fit, score, choose, select_trial, serialize, metrics, dev_results
from bengali_training_data import ROOT, training_rows, sha


class BengaliRerankerTest(unittest.TestCase):
    def test_weighted_counts_boundaries_case_and_pruning(self):
        counts, stats = fit([{'roman': 'Ab', 'attestations': '3'},
                             {'roman': 'ab', 'attestations': '2'},
                             {'roman': 'z', 'attestations': '1'},
                             {'roman': 'a-b', 'attestations': '9'}])
        self.assertEqual(counts['^ab$'], 5)
        self.assertEqual(counts['^'], 6)
        self.assertEqual(counts['b$'], 5)
        self.assertNotIn('z', counts)
        self.assertEqual(stats['attestation_mass'], 6)
        self.assertEqual(stats['skipped_non_ascii_letter_rows'], 1)
        with self.assertRaises(ValueError):
            fit([{'roman': 'ab', 'attestations': '0'}])

    def test_score_and_candidate_prior(self):
        counts, _ = fit([{'roman': 'ab', 'attestations': '3'}])
        self.assertEqual(score(counts, 'AB'), 0.0)
        # Each missing conditional order costs 0.4 before unigram fallback.
        self.assertAlmostEqual(score({'^': 3, 'a': 3, '$': 3}, 'a'),
                               (math.log(.4/3)+math.log(.4**2/3))/2)
        self.assertEqual(score(counts, 'a-b'), -math.inf)
        self.assertTrue(math.isfinite(score(counts, 'z')))
        self.assertEqual(choose('same', 'same', {}, 0), 'same')
        self.assertEqual(choose('b1', 'model', {'b1': 1, 'model': 1}, 0), 'model')
        self.assertEqual(choose('b1', 'model', {'b1': 2, 'model': 1}, 1), 'model')
        self.assertEqual(choose('b1', 'model', {'b1': 2, 'model': 1}, .5), 'b1')
        self.assertEqual(choose('b1', 'model', {'b1': -math.inf, 'model': -math.inf}, 0), 'model')

    def test_metric_denominators_and_reference_ties(self):
        rows = [('ক', {'a': 2, 'ab': 2}), ('খ', {'zz': 1})]
        self.assertEqual(metrics(rows, ['ab', 'zx']),
                         {'words': 2, 'strict': 0, 'any': 1, 'CER': .25})
        for refs, outputs in (([], []), (rows, ['ab'])):
            with self.assertRaises(ValueError):
                metrics(refs, outputs)

    def test_dev_guard_rejects_partial_gains_and_noop(self):
        baseline = {'any': 10, 'strict': 5, 'CER': .2}
        for trial in ({'any': 11, 'strict': 4, 'CER': .1},
                      {'any': 11, 'strict': 5, 'CER': .21},
                      {'any': 11, 'strict': 5, 'CER': .2}, baseline):
            self.assertIsNone(select_trial(baseline, [dict(trial, margin=0)]))
        good = {'any': 11, 'strict': 5, 'CER': .1, 'margin': .2}
        self.assertEqual(select_trial(baseline, [good, dict(good, margin=.4)]), good)

    def test_frozen_negative_result_and_train_only_reconstruction(self):
        report = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-rerank.json').read_text())
        counts, stats = fit(training_rows('dakshina'))
        self.assertEqual(sha(serialize(counts)), report['ngrams_sha256'])
        self.assertEqual(stats, report['training'])
        self.assertIsNone(select_trial(report['dev_model'], report['dev_trials']))
        self.assertIsNone(report['selected'])
        # Rerun the dev grid against the current engine: the committed outcome
        # must still reproduce (source hashes are provenance, not compared).
        baseline, trials, selected = dev_results(counts)
        self.assertEqual(baseline, report['dev_model'])
        self.assertEqual(trials, report['dev_trials'])
        self.assertIsNone(selected)


if __name__ == '__main__':
    unittest.main()
