import unittest

from build_bengali import native_key, parse_split


class BengaliFixtureTest(unittest.TestCase):
    def test_canonical_forms_and_format_marks(self):
        self.assertEqual(native_key("\u09dc"), native_key("\u09a1\u09bc"))
        self.assertEqual(native_key("\u09af\u09bc"), native_key("\u09df"))
        self.assertEqual(native_key("ক\u09c7\u09be"), native_key("ক\u09cb"))
        self.assertEqual(native_key("ক\u09c7\u09d7"), native_key("ক\u09cc"))
        self.assertEqual(native_key("ক্\u200dষ"), "ক্ষ")

    def test_duplicate_votes_aggregate_without_spelling_selection(self):
        rows = "ড়\tr\t2\nড়\tr\t3\nড়\trh\t1\n"
        counts, total, changed = parse_split(rows.encode())
        self.assertEqual(total, 3)
        self.assertEqual(changed, 1)
        self.assertEqual(counts, {("ড়", "r"): 5, ("ড়", "rh"): 1})

    def test_invalid_rows_fail(self):
        for row in ("ক\tk\t0\n", "ক\t\t1\n", "ক\tk\n"):
            with self.subTest(row=row), self.assertRaises(ValueError):
                parse_split(row.encode())


if __name__ == "__main__":
    unittest.main()
