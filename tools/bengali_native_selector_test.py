import json
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from train_native_selector import (features, training_references, target, fit,
                                   probability, predictions, select)
from bengali_training_data import ROOT, training_rows, verify


class BengaliNativeSelectorTest(unittest.TestCase):
    def test_excludes_candidate_model_training_and_rejects_unauthorized_rows(self):
        groups = {'dakshina-train': {'ক', 'খ', 'ড়'}, 'google-train': {'খ'}}
        def row(word, roman, votes=1):
            return {'native': word, 'roman': roman, 'attestations': str(votes)}
        rows = [row('ক', 'ko'), row('খ', 'kho'), row('ড়', 'r', 2), row('ড়', 'r', 3)]
        self.assertEqual(training_references(rows, groups), [('ক', {'ko': 1}), ('ড়', {'r': 5})])
        with self.assertRaisesRegex(ValueError, 'unauthorized'):
            training_references(rows + [row('গ', 'go')], groups)
        with self.assertRaisesRegex(ValueError, 'invalid'):
            training_references([row('ক', 'ko', 0)], groups)

    def test_all_frozen_training_exclusions(self):
        groups = verify()
        rows = training_references(training_rows('dakshina'), groups)
        keys = {word for word, _ in rows}
        expected = groups['dakshina-train'] - groups['google-train']
        self.assertEqual(sorted(keys - expected)[:5], [], 'unexpected training types (first five)')
        self.assertEqual(sorted(expected - keys)[:5], [], 'missing training types (first five)')
        for name, words in groups.items():
            if name != 'dakshina-train':
                self.assertFalse(keys & words, name)

    def test_native_and_edit_features(self):
        x = features('কুমার', 'kumar', 'kumor')
        self.assertEqual((x['has_u'], x['last'], x['last2']), ('True', 'র', 'ার'))
        self.assertEqual((x['edit_b1'], x['edit_model'], x['before'], x['after']), ('a', 'o', 'm', 'r'))
        self.assertEqual(features('ড়', 'r', 'ro'), features('ড়', 'r', 'ro'))
        self.assertEqual(features('ক', 'k', 'ko')['edit_b1'], '')
        self.assertEqual(features('ক', 'k', 'ko')['edit_at_end'], 'True')

    def test_reference_preference_and_ties(self):
        self.assertEqual(target({'ko': 2, 'ka': 1}, 'ka', 'xx'), 1)
        self.assertEqual(target({'ko': 2, 'ka': 1}, 'ka', 'ko'), 0)
        self.assertEqual(target({'ko': 2, 'ka': 1}, 'ko', 'ka'), 1)
        self.assertEqual(target({'ko': 1}, 'xo', 'xx'), 1)
        self.assertEqual(target({'ko': 1}, 'xo', 'kx'), 0)

    def test_tree_learns_and_respects_minimum_leaf_and_threshold(self):
        rows = [({'last': 'ক'}, 1)]*4 + [({'last': 'খ'}, 0)]*4
        tree = fit(rows, depth=1, min_leaf=4)
        self.assertEqual(tree, fit(list(reversed(rows)), depth=1, min_leaf=4))
        self.assertEqual(probability(tree, {'last': 'ক'}), 1)
        self.assertEqual(probability(tree, {'last': 'খ'}), 0)
        self.assertEqual(fit(rows, depth=5, min_leaf=5), {'counts': [4, 4]})
        self.assertEqual(fit(rows, depth=0, min_leaf=1), {'counts': [4, 4]})
        sample = [('ক', {}), ('খ', {})]
        self.assertEqual(predictions(tree, sample, [('k', 'ko'), ('kh', 'kho')], .5), ['k', 'kho'])
        self.assertEqual(predictions(tree, sample, [('k', 'ko'), ('kh', 'kho')], 1), ['ko', 'kho'])
        self.assertEqual(predictions(tree, [('ক', {})], [('same', 'same')], .5), ['same'])
        with self.assertRaises(ValueError):
            fit([])
        with self.assertRaises(ValueError):
            predictions(tree, sample, [], .5)

    def test_selection_rejects_noop_and_partial_gains(self):
        baseline = {'any': 10, 'strict': 5, 'CER': .2}
        for change in ({}, {'any': 11, 'strict': 4, 'CER': .1},
                       {'any': 11, 'CER': .2}, {'any': 11, 'CER': .21}):
            self.assertIsNone(select(baseline, [dict(baseline, threshold=.5, **change)]))
        good = {'any': 11, 'strict': 5, 'CER': .1, 'threshold': .7}
        self.assertEqual(select(baseline, [dict(good, threshold=.5), good]), good)
        report = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-native-selector.json').read_text())
        self.assertIsNone(select(report['dev_model'], report['dev_trials']))
        self.assertIsNone(report['selected'])


if __name__ == '__main__':
    unittest.main()
