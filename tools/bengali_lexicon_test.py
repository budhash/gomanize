import gzip
import json
from pathlib import Path
import sys
import unittest
import unicodedata
from unittest.mock import patch
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
import build_lexicon as builder
from build_lexicon import build, select
from build_lexicon_keys import cases
from bengali_training_data import ROOT, assert_lexicon_isolation, verify, sha


class BengaliLexiconTest(unittest.TestCase):
    def test_selection_requires_unique_attested_winner(self):
        def row(word, roman, votes):
            return {'native': word, 'roman': roman, 'attestations': str(votes)}
        entries, stats = select([row('ক', 'K', 3), row('ক', 'ko', 2),
                                 row('খ', 'kho', 3), row('খ', 'kha', 3),
                                 row('গ', 'go', 2), row('২', 'dui', 4)])
        self.assertEqual(entries, {'ক': 'K'})
        self.assertEqual((stats['tied_top'], stats['below_threshold'], stats['nonword']), (1, 1, 1))
        entries, _ = select([row('ড়', 'r', 2), row('ড়', 'r', 1)])
        self.assertEqual(entries, {'ড়': 'r'})
        for bad in (row('ক', 'k', 0), row('ক', '', 3), row('ক', 'a\tb', 3)):
            with self.assertRaises(ValueError):
                select([bad])

    def test_reconstruction_provenance_and_all_exclusions(self):
        payload, manifest = build()
        self.assertEqual(payload, (ROOT/'lang/bengali/lexicon.tsv').read_bytes())
        self.assertEqual(manifest, (ROOT/'lang/bengali/lexicon_manifest.json').read_bytes())
        external = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-lexicon-external.json').read_text())
        self.assertEqual(external['lexicon_sha256'], sha(payload))
        # Shared columns must agree with the vowel-model report, so a stale
        # rerun of either report cannot go unnoticed.
        vowels = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-vowels-external.json').read_text())
        self.assertEqual(external['model_sha256'], vowels['model_sha256'])
        for corpus in ('Aksharantar', 'BanglaTLit'):
            for subset, columns in vowels[corpus].items():
                if not isinstance(columns, dict):
                    continue
                for column in ('B1', 'model'):
                    if column in columns:
                        self.assertEqual(external[corpus][subset][column], columns[column], (corpus, subset, column))
        keys = {line.split('\t')[0] for line in payload.decode().splitlines()}
        groups = verify()
        for name, heldout in groups.items():
            if name.endswith('-train'):
                continue
            self.assertFalse(keys & heldout, name)
            with self.subTest(name=name), self.assertRaises(ValueError):
                assert_lexicon_isolation(keys | {next(iter(heldout))})

    def test_builder_rejects_an_injected_heldout_row(self):
        word = next(w for w in sorted(verify()['dakshina-test'])
                    if all('\u0980' <= c <= '\u09ff' and unicodedata.category(c)[0] in 'LM' for c in w))
        injected = [{'native': word, 'roman': 'test', 'attestations': '3'}]
        with patch.object(builder, 'training_rows', return_value=injected):
            with self.assertRaisesRegex(ValueError, 'held-out or unauthorized'):
                build()

    def test_unicode_parity_fixture_is_current(self):
        fixture = gzip.decompress((ROOT/'lang/bengali/testdata/lexicon_keys.json.gz').read_bytes())
        self.assertEqual(json.loads(fixture), cases())


if __name__ == '__main__':
    unittest.main()
