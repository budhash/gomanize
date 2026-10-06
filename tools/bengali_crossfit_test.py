from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from train_crossfit_selector import (fold, fold_examples, grouped_training, variants,
                                     decisions, pair_examples, choose, FOLDS, MAX_EDITS)
from bengali_training_data import verify, training_rows, sha
from train_crossfit_selector import dumps, evaluate, ROOT
from evaluate_crossfit import frozen_selection, verify_sources
import json


class BengaliCrossfitTest(unittest.TestCase):
    def test_word_grouped_folds_exclude_candidate_training(self):
        self.assertEqual(fold('ড়'), fold('ড়'))
        self.assertEqual(fold('ক\u200dম'), fold('কম'))
        groups = verify()
        examples = [({}, 0, word, 0) for word in sorted(groups['google-train'])]
        predicted = set()
        for i in range(FOLDS):
            held = {word for word in groups['dakshina-train'] if fold(word) == i}
            fitted = {r[2] for r in fold_examples(examples, i)}
            self.assertFalse(fitted & held, 'prediction words leaked into candidate fit')
            self.assertTrue(all(fold(word) != i for word in fitted))
            self.assertFalse(predicted & held)
            predicted |= held
        self.assertEqual(len(predicted), len(groups['dakshina-train']))

    def test_reference_training_isolation(self):
        groups = verify()
        rows = training_rows('dakshina')
        refs = grouped_training(rows, groups['dakshina-train'])
        self.assertEqual({w for w, _ in refs}, groups['dakshina-train'])
        for split in ('google-dev', 'google-test', 'dakshina-dev', 'dakshina-test'):
            bad = {'native': min(groups[split]), 'roman': 'x', 'attestations': '1'}
            with self.assertRaisesRegex(ValueError, 'unauthorized'):
                grouped_training(rows + [bad], groups['dakshina-train'])

    def test_candidate_expansion_is_bounded_and_one_slot_only(self):
        labels = {i: i % 3 for i in range(12)}
        result = variants(labels)
        self.assertEqual(len(result), MAX_EDITS+1)
        self.assertEqual(result[0], labels)
        for i, candidate in enumerate(result[1:]):
            self.assertEqual([j for j in labels if labels[j] != candidate[j]], [i])
            self.assertEqual(candidate[i], 1 if labels[i] == 0 else 0)
        self.assertEqual(variants({}), [{}])
        result[0][0] = 2
        self.assertEqual(labels[0], 0)
        self.assertEqual(decisions('ক্ত', {'leaf':1}), {})
        self.assertEqual(decisions('কাম', {'leaf':1}), {2:1})

    def test_heldout_requires_unchanged_dev_selected_artifact(self):
        artifact = {'tree': {'counts': [1, 3]}}
        baseline = {'words': 100, 'strict': 40, 'any': 60, 'CER': .2}
        trial = dict(baseline, strict=41, any=61, CER=.19, threshold=.6)
        report = {'artifact_sha256': sha(dumps(artifact)), 'selected': trial,
                  'dev': {'baseline': baseline, 'trials': [trial]}}
        self.assertEqual(frozen_selection(artifact, report), .6)
        with self.assertRaisesRegex(ValueError, 'source changed'):
            verify_sources({'provenance': {'sources_sha256': {
                'tools/bengali/train_crossfit_selector.py': 'wrong'}}})
        with self.assertRaisesRegex(ValueError, 'model or split changed'):
            verify_sources({'provenance': {'sources_sha256': {},
                            'vowel_model_sha256': 'wrong', 'split_manifest_sha256': 'wrong'}})
        with self.assertRaisesRegex(ValueError, 'artifact changed'):
            frozen_selection({'tree': {'counts': [2, 3]}}, report)
        with self.assertRaisesRegex(ValueError, 'did not select'):
            frozen_selection(artifact, dict(report, selected=None))
        with self.assertRaisesRegex(ValueError, 'differs'):
            frozen_selection(artifact, dict(report, selected=dict(trial, threshold=.7)))
        rejected = dict(baseline, threshold=.6)
        with self.assertRaisesRegex(ValueError, 'differs'):
            frozen_selection(artifact, dict(report, selected=rejected,
                             dev={'baseline': baseline, 'trials': [rejected]}))

    def test_pair_labels_and_stable_selection(self):
        rows = [('কম', {'kom':3, 'kmo':1})]
        pairs = pair_examples(rows, [['kom', 'kmo', 'km']])
        self.assertEqual([y for _, y in pairs], [0, 0])
        self.assertEqual([y for _, y in pair_examples(rows, [['km', 'kom']])], [1])
        tree = {'counts':[1, 3]}
        self.assertEqual(choose(tree, 'কম', ['km', 'kom', 'kmo'], .7), 'kom')
        self.assertEqual(choose(tree, 'কম', ['km', 'kom'], .75), 'km')
        self.assertEqual(choose(tree, 'কম', ['km'], .5), 'km')
        with self.assertRaises(ValueError):
            pair_examples(rows, [])
        with self.assertRaises(ValueError):
            choose(tree, 'কম', [], .5)

    def test_committed_records_reproduce(self):
        # Re-evaluate the committed selector against the current engine; the
        # committed dev and held-out records must match. This only asserts the
        # frozen outcome; it never re-selects. Provenance hashes are excluded
        # (they drift with any engine edit); a full retrain needs the raw
        # Google lexicon and stays an offline reproduction step.
        reviews = ROOT/'docs/reviews'
        artifact = json.loads((reviews/'2026-10-04-bengali-b2-crossfit-selector.json').read_text())
        report = json.loads((reviews/'2026-10-04-bengali-b2-crossfit.json').read_text())
        heldout = json.loads((reviews/'2026-10-04-bengali-b2-crossfit-heldout.json').read_text())
        threshold = frozen_selection(artifact, report)
        self.assertEqual(json.loads(json.dumps(evaluate(artifact, 'dev'))), report['dev'])
        self.assertEqual(heldout['frozen_threshold'], threshold)
        self.assertEqual(json.loads(json.dumps(evaluate(artifact, 'test', threshold))), heldout['test'])


if __name__ == '__main__':
    unittest.main()
