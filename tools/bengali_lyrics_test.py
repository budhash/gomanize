import csv
import io
import json
import math
from pathlib import Path
import shutil
import sys
import tempfile
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from bengali_lyrics import ROOT, DATA, extract, verify, sha
from evaluate_lyrics import score, evaluate


def assert_close(case, got, want, path=''):
    """Exact match except floats, which may differ in the last ulp across
    Python versions (3.12 sum() uses compensated summation)."""
    if isinstance(want, dict):
        case.assertEqual(set(got), set(want), path)
        for key in want:
            assert_close(case, got[key], want[key], path+'.'+key)
    elif isinstance(want, float):
        case.assertTrue(math.isclose(got, want, rel_tol=1e-12, abs_tol=1e-15), f'{path}: {got} != {want}')
    else:
        case.assertEqual(got, want, path)


class BengaliLyricsTest(unittest.TestCase):
    def test_source_reconstruction(self):
        rows, manifest = verify()
        self.assertEqual([s['lines'] for s in manifest['songs']], [20, 20, 20, 22])
        self.assertEqual(len(rows), 82)
        self.assertEqual(manifest['unique_native_lines'], 71)

    def test_saved_report_recomputes_from_frozen_references(self):
        rows, manifest = verify()
        report = json.loads((ROOT/'docs/reviews/2026-10-05-bengali-b3-lyrics.json').read_text())
        self.assertEqual(report['fixture_sha256'], manifest['fixture_sha256'])
        predictions = report['predictions']
        self.assertEqual([(r['id'], r['native'], r['roman']) for r in rows],
                         [(p['id'], p['native'], p['reference']) for p in predictions])
        from evaluate_runtime import PROFILES
        values = [[p['profiles'][name] for name in PROFILES] for p in predictions]
        assert_close(self, score(rows, values), report['scores']['all_lines'])

    def test_record_matches_current_engine(self):
        # Re-run the engine on the frozen lines and require the whole committed
        # record (predictions, every slice, overlap) to match; only provenance
        # hashes are excluded. Logged-only Go scores would not catch a drift.
        from evaluate_lyrics import evaluate, predict, training_groups
        rows, _ = verify()
        record = json.loads((ROOT/'docs/reviews/2026-10-05-bengali-b3-lyrics.json').read_text())
        rebuilt = evaluate(rows, predict([(r['native'], [r['roman']]) for r in rows]), training_groups())
        for key in ('predictions', 'scores', 'equal_weight_song_macro_CER', 'overlap'):
            if key == 'predictions':
                self.assertEqual(json.loads(json.dumps(rebuilt[key], ensure_ascii=False)), record[key])
            else:
                assert_close(self, json.loads(json.dumps(rebuilt[key])), record[key], key)

    def test_extraction_boundaries(self):
        html = '<p>ignore</p><div class="poem"><p>ক&nbsp; খ<br/><span>গ\u200b</span></p></div>ignore<div class="poem"><p>ঘ</p></div>'
        self.assertEqual(extract(html.encode()), ['ক খ', 'গ', 'ঘ'])

    def test_tampering_negative_controls(self):
        for mutation in ('reference', 'native_rehashed', 'status_rehashed', 'snapshot'):
            with self.subTest(mutation=mutation), tempfile.TemporaryDirectory() as tmp:
                data = Path(tmp)/'data'
                shutil.copytree(DATA, data)
                manifest = json.loads((data/'manifest.json').read_text())
                if mutation == 'snapshot':
                    path = data/manifest['songs'][0]['snapshot']
                    path.write_bytes(path.read_bytes()+b'bad')
                else:
                    rows = list(csv.DictReader(io.StringIO((data/'pilot.csv').read_text())))
                    field = {'reference':'roman', 'native_rehashed':'native', 'status_rehashed':'reference_status'}[mutation]
                    rows[0][field] = 'bad'
                    with (data/'pilot.csv').open('w', newline='') as f:
                        w = csv.DictWriter(f, fieldnames=list(rows[0]), lineterminator='\n')
                        w.writeheader(); w.writerows(rows)
                    if mutation.endswith('_rehashed'):
                        manifest['fixture_sha256'] = sha((data/'pilot.csv').read_bytes())
                        (data/'manifest.json').write_text(json.dumps(manifest))
                with self.assertRaises(ValueError):
                    verify(data)

    def test_denominators_and_paired_changes(self):
        rows = [{'roman':'a'}, {'roman':'abc'}]
        preds = [['x','a','x','a','x','x'], ['ab','ab','ab','ab','abc','abc']]
        result = score(rows, preds)
        self.assertAlmostEqual(result['B1']['macro_line_CER'], 2/3)
        self.assertEqual(result['B1']['micro_CER'], .5)
        self.assertEqual(result['rerank_vs_model'], {'changed':2, 'CER_better':1, 'CER_worse':1, 'CER_equal':0, 'exact_wins':1, 'exact_losses':1})
        self.assertIsNone(score([], [])['B1']['macro_line_CER'])
        for bad in (preds[:1], [preds[0][:5],preds[1]], [[None]*6, preds[1]]):
            with self.assertRaises(ValueError):
                score(rows, bad)

    def test_duplicates_song_weight_and_unseen_filter(self):
        rows = [{'id':str(i), 'native':w, 'roman':r, 'song':s} for i,(w,r,s) in enumerate([
            ('ক','a','1'), ('ক','a','1'), ('খ','a','2')])]
        groups = {'google-train': {'ক'}, 'dakshina-train':set()}
        result = evaluate(rows, [['b']*6, ['b']*6, ['a']*6], groups)
        self.assertAlmostEqual(result['scores']['all_lines']['B1']['macro_line_CER'], 2/3)
        self.assertEqual(result['scores']['unique_native_lines']['B1']['macro_line_CER'], .5)
        self.assertEqual(result['equal_weight_song_macro_CER']['B1'], .5)
        self.assertEqual(result['scores']['all_tokens_unseen_lines']['lines'], 1)
        self.assertEqual(result['overlap']['training_token_occurrences'], 2)


if __name__ == '__main__':
    unittest.main()
