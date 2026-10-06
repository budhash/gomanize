import json
from pathlib import Path
import sys
import unittest
sys.path.insert(0, str(Path(__file__).parent/'bengali'))
from build_selector_parity import build
from bengali_training_data import ROOT, sha
from evaluate_runtime import score, dakshina_section, banglatlit_rows, banglatlit_section, predict
from bengali_training_data import verify


class BengaliRuntimeTest(unittest.TestCase):
    def test_frozen_artifact_and_synthetic_reconstruction(self):
        frozen = (ROOT/'docs/reviews/2026-10-04-bengali-b2-crossfit-selector.json').read_bytes()
        self.assertEqual((ROOT/'lang/bengali/selector.json').read_bytes(), frozen)
        old = json.loads((ROOT/'docs/reviews/2026-10-04-bengali-b2-crossfit.json').read_text())
        self.assertEqual(sha(frozen), old['artifact_sha256'])
        self.assertEqual(old['selected']['threshold'], .6)
        self.assertEqual(build(), (ROOT/'lang/bengali/testdata/selector_parity.json.gz').read_bytes())

    def test_committed_record_reproduces(self):
        # Recompute the sections that need no external archive and require the
        # committed runtime record to match (Aksharantar needs the pinned zip
        # and remains an offline reproduction step).
        record = json.loads((ROOT/'docs/reviews/2026-10-05-bengali-b2-runtime.json').read_text())
        self.assertEqual(json.loads(json.dumps(dakshina_section())), record['Dakshina'])
        groups = verify()
        _, bangla = banglatlit_rows()
        section = banglatlit_section(bangla, predict(bangla), groups['google-train'] | groups['dakshina-train'])
        self.assertEqual(json.loads(json.dumps(section)), record['BanglaTLit'])

    def test_external_pair_denominators_and_losses(self):
        rows = [('ক', ['a']), ('খ', ['ab'])]
        values = [['a', 'a', 'a', 'a', 'b', 'b'], ['x', 'x', 'x', 'x', 'ab', 'ab']]
        result = score(rows, values)
        self.assertEqual(result['rerank_vs_model'], {'wins': 1, 'losses': 1, 'changed': 2})
        self.assertEqual(result['rerank']['macro_minCER'], .5)
        self.assertIsNone(score([], [])['model']['accuracy'])
        with self.assertRaises(ValueError):
            score(rows, values[:1])


if __name__ == '__main__':
    unittest.main()
