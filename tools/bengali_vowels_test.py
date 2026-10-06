import hashlib
import json
from pathlib import Path
import sys
import unittest

sys.path.insert(0, str(Path(__file__).parent / 'bengali'))
from vowels import align, dataset, features, units
from train_vowels import predict, tree
from bengali_training_data import DEFAULT, GOOGLE_SHA, ROOT


class BengaliVowelsTest(unittest.TestCase):
    def test_exact_alignment_and_all_classes(self):
        for phones, expected in [('k', 0), ('k O', 1), ('k o', 2)]:
            self.assertEqual(align('ক', phones), ('aligned', ((0, expected),)))
        self.assertEqual(align('কমল', 'k O . m o l'), ('aligned', ((0, 1), (1, 2), (2, 0))))
        self.assertEqual(align('কাম', 'k a m'), ('aligned', ((2, 0),)))
        self.assertEqual(align('কা', 'k a'), ('no_slots', ()))
        self.assertEqual(align('কঅ', 'k O'), ('unsupported', None))

    def test_no_guessing_or_partial_phone_consumption(self):
        for phones in ('k a', 'k O x', 'x k O', 'k k O', ''):
            self.assertEqual(align('ক', phones), ('unmatched', None))
        for word in ('ক্ষ', 'ক্', 'চাঁদ', 'বাংলা', 'হওয়া', 'ঐ', 'ৎ', 'কাা', 'াক', 'aক', ''):
            self.assertIsNone(units(word))

    def test_variants_are_not_overweighted_or_arbitrarily_resolved(self):
        rows = [('ক', 1, 'ক', 'k O'), ('ক', 2, 'ক', 'k O', 'noun')]
        data, stats = dataset(rows)
        self.assertEqual(len(data), 1)
        self.assertEqual(stats['accepted_types'], 1)
        rows.append(('ক', 3, 'ক', 'k o'))
        data, stats = dataset(rows)
        self.assertEqual(data, [])
        self.assertEqual(stats['conflicting_variant_types'], 1)
        rows.append(('ক', 4, 'ক', 'k x'))
        data, stats = dataset(rows)
        self.assertEqual(data, [])
        self.assertEqual(stats['rejected_types'], 1)

    def test_normalized_feature_offsets(self):
        self.assertEqual(units('ড়ক'), units('ড়ক'))
        self.assertEqual(units('কো'), units('কো'))
        self.assertEqual(units('ক\u200dম'), units('কম'))
        self.assertEqual(features('ড়ক', 2)['prev'], '়')
        self.assertEqual(features('কাম', 2)['last'], '1')

    def test_artifact_provenance_and_fixture_predictions(self):
        model = json.loads((ROOT/'lang/bengali/vowel_tree.json').read_text())
        provenance = model['provenance']
        self.assertEqual(provenance['google_sha256'], GOOGLE_SHA)
        self.assertEqual(provenance['split_manifest_sha256'], hashlib.sha256((DEFAULT/'manifest.json').read_bytes()).hexdigest())
        for name, expected in provenance['training_scripts_sha256'].items():
            self.assertEqual(hashlib.sha256((ROOT/'tools'/name).read_bytes()).hexdigest(), expected)
        external = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-vowels-external.json').read_text())
        self.assertEqual(external['model_sha256'], hashlib.sha256((ROOT/'lang/bengali/vowel_tree.json').read_bytes()).hexdigest())
        fixtures = json.loads((ROOT/'lang/bengali/testdata/vowel_features.json').read_text())
        for row in fixtures:
            self.assertEqual(features(row['word'], row['index']), row['features'])
            self.assertEqual(predict(model['tree'], row['features']), row['label'])

    def test_three_class_tree_learns_and_ties_are_deterministic(self):
        data = []
        for word, phone in [('ক', 'k'), ('খ', 'kh O'), ('গ', 'g o')]:
            examples, _ = dataset([(word, 1, word, phone)])
            data += examples
        model = tree(data, 3, 1)
        self.assertEqual(model, tree(list(reversed(data)), 3, 1))
        self.assertEqual([predict(model, r[0]) for r in data], [0, 1, 2])


if __name__ == '__main__':
    unittest.main()
