import copy
import gzip
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch

import bengali_training_data as data
from audit_bengali_training_overlap import bengali_tokens


class TrainingIsolationTest(unittest.TestCase):
    def test_source_variants_and_labels_survive(self):
        source = '# attribution\nড়\tr\tnoun \nড়\trh\n'
        with patch.object(data, 'pinned', return_value=gzip.compress(source.encode())):
            rows = data.google_rows('synthetic')
        self.assertEqual(rows, [('ড়', 2, 'ড়', 'r', 'noun '), ('ড়', 3, 'ড়', 'rh')])
        groups = data.split_keys({r[0] for r in rows}, {'train': set(), 'dev': set(), 'test': set()})
        self.assertEqual(sum(len(v) for k, v in groups.items() if k.startswith('google-')), 1)

    def test_external_overlap_token_boundaries(self):
        self.assertEqual(list(bengali_tokens('ক্\u200dষ, বাংলা। abc ১২৩')), ['ক্ষ', 'বাংলা'])

    def test_normalization(self):
        for left, right in [('ড়', 'ড়'), ('ক\u09c7\u09be', 'কো'), ('ক্\u200dষ', 'ক্ষ')]:
            self.assertEqual(data.native_key(left), data.native_key(right))

    def test_variant_spelling_collisions_are_pinned(self):
        # Khanda-ta and ta + hasant are not canonically equivalent but render alike.
        self.assertNotEqual(data.native_key('উত্সাহ'), data.native_key('উৎসাহ'))
        self.assertEqual(data.collision_key('উত্সাহ'), data.collision_key('উৎসাহ'))
        self.assertEqual(data.collision_key('অাবার'), data.collision_key('আবার'))
        groups = data.verify()
        self.assertEqual(data.variant_collisions(groups), data.KNOWN_VARIANT_COLLISIONS)
        self.assertEqual(len({c[3] for c in data.KNOWN_VARIANT_COLLISIONS}), 15)
        leaked = copy.deepcopy(groups)
        leaked['dakshina-train'].add('অর্থাত্')  # variant of dakshina-dev অর্থাৎ
        self.assertIn(('dakshina-train', 'অর্থাত্', 'dakshina-dev', 'অর্থাৎ'), data.variant_collisions(leaked))

    def test_exclusion_precedes_partition(self):
        google = {'ক', 'খ', 'গ', 'ঘ'}
        dak = {'train': {'ক', 'খ'}, 'dev': {'গ'}, 'test': {'ঘ'}}
        groups = data.split_keys(google, dak)
        self.assertFalse(set.union(*(groups['google-' + s] for s in data.DAK_SHA)) & {'গ', 'ঘ'})
        self.assertEqual(groups, data.split_keys(set(reversed(sorted(google))), dak))
        leaked = copy.deepcopy(groups)
        leaked['dakshina-train'].add('গ')
        with self.assertRaisesRegex(ValueError, 'leakage'):
            data.validate_groups(leaked)

    def test_real_training_and_lexicon_guard(self):
        groups = data.verify()
        rows = data.training_rows('dakshina')
        keys = {data.native_key(r['native']) for r in rows}
        self.assertEqual(keys, groups['dakshina-train'])
        data.assert_lexicon_isolation(keys)
        for name in ('google-dev', 'google-test', 'dakshina-dev', 'dakshina-test'):
            with self.subTest(name=name), self.assertRaises(ValueError):
                data.assert_lexicon_isolation(keys | {next(iter(groups[name]))})
        with self.assertRaises(ValueError):
            data.training_rows('test')

    def test_rehashed_leak_still_fails(self):
        # Negative control: leak a test word and update its hash/count too.
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp) / 'data'
            shutil.copytree(data.DEFAULT, directory)
            groups = data.verify(directory)
            groups['dakshina-train'].add(next(iter(groups['dakshina-test'])))
            payload = data.key_bytes(groups['dakshina-train'])
            (directory / 'dakshina-train.txt.gz').write_bytes(gzip.compress(payload, mtime=0))
            manifest = json.loads((directory / 'manifest.json').read_text())
            manifest['partitions']['dakshina-train'] = {
                'words': len(groups['dakshina-train']), 'sha256': data.sha(payload)}
            (directory / 'manifest.json').write_text(json.dumps(manifest))
            with self.assertRaisesRegex(ValueError, 'leakage'):
                data.verify(directory)

    def test_removed_google_holdout_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp) / 'data'
            shutil.copytree(data.DEFAULT, directory)
            groups = data.verify(directory)
            # Remove a Google-only protected word so Dakshina checks cannot mask this guard.
            dak_train = {data.native_key(r['native']) for r in data.dak_rows('train')}
            key = next(iter(groups['google-test'] - dak_train))
            groups['google-test'].remove(key)
            payload = data.key_bytes(groups['google-test'])
            (directory / 'google-test.txt.gz').write_bytes(gzip.compress(payload, mtime=0))
            manifest = json.loads((directory / 'manifest.json').read_text())
            manifest['partitions']['google-test'] = {
                'words': len(groups['google-test']), 'sha256': data.sha(payload)}
            (directory / 'manifest.json').write_text(json.dumps(manifest))
            with self.assertRaisesRegex(ValueError, 'frozen Google'):
                data.verify(directory)

    def test_bad_source_fails_before_reading_rows(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / 'source.gz'
            path.write_bytes(gzip.compress(b'bad source'))
            with self.assertRaisesRegex(ValueError, 'checksum'):
                data.google_rows(path)


if __name__ == '__main__':
    unittest.main()
