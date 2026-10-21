import hashlib
import json
from pathlib import Path
import sys
import unittest

sys.path.insert(0, str(Path(__file__).parent / 'bengali'))
from vowels import INVENTORY, LEGACY_SCOPE, SCOPE, align, dataset, eligible, features, units
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
        for word in ('ক্', '্ক', 'ক্্ক', 'ক্ষ্ম্য', 'জ্ঞান', 'ক্য়', 'ক্অ', 'চাঁদ', 'বাংলা', 'হওয়া', 'ঐ', 'ৎ', 'কাা', 'াক', 'aক', ''):
            self.assertIsNone(units(word))
        # The legacy simple-word scope still rejects every cluster.
        for word in ('ক্ষ', 'কর্ম', 'মধ্য'):
            self.assertIsNone(units(word, 1))
            self.assertFalse(eligible(word, 1))
            self.assertTrue(eligible(word))

    def test_cluster_units(self):
        # Members before a halant have no slot; the last member has one unless a matra follows.
        self.assertEqual(units('কর্ম'), [(0, 'ক', True), (1, 'র', False), (2, '্', False), (3, 'ম', True)])
        self.assertEqual(units('স্ত্রী'), [(0, 'স', False), (1, '্', False), (2, 'ত', False), (3, '্', False),
                                           (4, 'র', False), (5, 'ী', False)])
        self.assertFalse(eligible('স্ত্রী'))  # supported, but no inherent slot
        self.assertEqual(units('ড়্গ'), [(0, 'ড়', False), (2, '্', False), (3, 'গ', True)])
        self.assertIsNone(units('ক্কঅ'))  # slot before an independent vowel stays B1
        self.assertEqual(features('মধ্য', 3)['prev'], '্')
        self.assertEqual(features('কর্ম', 0)['next2'], '্')
        self.assertEqual(features('কর্ম', 3)['last'], '1')

    def test_cluster_alignment_alternatives(self):
        # Own phone, gemination of the previous member, or silence, for non-first members.
        self.assertEqual(align('কর্ম', 'k O r . m o'), ('aligned', ((0, 1), (3, 2))))
        self.assertEqual(align('পত্র', 'p O t . t o'), ('aligned', ((0, 1), (3, 2))))
        self.assertEqual(align('কন্যা', 'k O n . n a'), ('aligned', ((0, 1),)))
        self.assertEqual(align('দ্বার', 'd a r'), ('aligned', ((4, 0),)))
        # ক্ষ: ষ may read kh after ক; ক may be silent (train-failure alternative).
        self.assertEqual(align('পক্ষ', 'p O k . kh o'), ('aligned', ((0, 1), (3, 2))))
        self.assertEqual(align('ক্ষমা', 'kh O . m a'), ('aligned', ((2, 1),)))
        # First members never go silent or geminate, and phones are never invented.
        self.assertEqual(align('কর্ম', 'k O m o'), ('unmatched', None))
        self.assertEqual(align('পত্র', 'p O t . t r o'), ('unmatched', None))
        self.assertEqual(align('মধ্য', 'm O d . dh o'), ('unmatched', None))
        # Vowel handling is unchanged: a ya-phala read as E does not align.
        self.assertEqual(align('ব্যাগ', 'b E g'), ('unmatched', None))
        self.assertEqual(align('জ্ঞান', 'g a n'), ('unsupported', None))

    def test_shared_inventory(self):
        self.assertEqual((SCOPE, LEGACY_SCOPE, INVENTORY['max_cluster']), ('halant-clusters-v2', 'simple-whole-words-v1', 3))
        fixtures = json.loads((ROOT/'lang/bengali/testdata/vowel_eligibility.json').read_text())
        self.assertGreater(len(fixtures), 100)
        for row in fixtures:
            self.assertEqual((eligible(row['word']), eligible(row['word'], 1)), (row['eligible'], row['legacy_eligible']), row['word'])

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
        self.assertEqual(model['scope'], SCOPE)
        self.assertEqual(provenance['inventory_sha256'], hashlib.sha256((ROOT/'lang/bengali/vowel_inventory.json').read_bytes()).hexdigest())
        record = json.loads((ROOT/'docs/reviews/2026-10-21-bengali-vowel-conjuncts.json').read_text())
        self.assertEqual(record['inputs_sha256']['lang/bengali/vowel_tree.json'], hashlib.sha256((ROOT/'lang/bengali/vowel_tree.json').read_bytes()).hexdigest())
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
