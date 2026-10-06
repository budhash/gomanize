import csv
import gzip
import hashlib
import io
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import build_banglatlit as builder


class ImportTests(unittest.TestCase):
    def test_pinned_source_and_deterministic_preservation(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'data').mkdir()
            (root / 'LICENSE').write_text('test license')
            text = io.StringIO(newline='')
            writer = csv.writer(text)
            writer.writerow(['id', 'text_transliterated', 'text_bengali'])
            for i in range(2500):
                writer.writerow([str(i), 'Keep Case!', 'ক\u200dো'])
            raw = text.getvalue().encode()
            (root / 'data/BanglaTLiT_test.csv').write_bytes(raw)
            with self.assertRaises(ValueError):
                builder.build(root, root / 'rejected')
            with patch.object(builder, 'SOURCE_SHA', hashlib.sha256(raw).hexdigest()):
                builder.build(root, root / 'one')
                builder.build(root, root / 'two')
            self.assertEqual((root / 'one/test.csv.gz').read_bytes(), (root / 'two/test.csv.gz').read_bytes())
            with gzip.open(root / 'one/test.csv.gz', 'rt') as stream:
                rows = list(csv.reader(stream))
            self.assertEqual(rows[1], ['0', 'কো', 'Keep Case!'])
            self.assertEqual(rows[-1][0], '2499')
            manifest = json.loads((root / 'one/manifest.json').read_text())
            self.assertEqual(manifest['changed_native_rows'], 2500)
            self.assertEqual(manifest['unique_normalized_native'], 1)


if __name__ == '__main__':
    unittest.main()
