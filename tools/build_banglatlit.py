#!/usr/bin/env python3
"""Import only the pinned BanglaTLit official test split as evaluation data."""
import argparse
import csv
import gzip
import hashlib
import io
import json
from pathlib import Path
import unicodedata

SOURCE_SHA = "2e6d3d3e3a1eab578011971be14807fdb20ae13384bc749c1d54eb209234c872"
SOURCE_COMMIT = "f780fcdd5176ddfe6364ffe7948acb6a1c2ea58e"


def native_key(text):
    return unicodedata.normalize("NFC", "".join(c for c in text if unicodedata.category(c) != "Cf"))


def build(checkout, output):
    source = checkout / "data/BanglaTLiT_test.csv"
    raw = source.read_bytes()
    if hashlib.sha256(raw).hexdigest() != SOURCE_SHA:
        raise ValueError("source test CSV differs from the pinned revision")
    reader = csv.DictReader(io.StringIO(raw.decode("utf-8")))
    if reader.fieldnames != ["id", "text_transliterated", "text_bengali"]:
        raise ValueError("unexpected source columns")
    rows = list(reader)
    if len(rows) != 2500 or len({r["id"] for r in rows}) != 2500:
        raise ValueError("unexpected source counts or repeated IDs")
    text = io.StringIO(newline="")
    writer = csv.writer(text, lineterminator="\n")
    writer.writerow(["id", "native", "roman"])
    for row in rows:
        if not row["text_bengali"] or not row["text_transliterated"]:
            raise ValueError("empty test pair")
        writer.writerow([row["id"], native_key(row["text_bengali"]), row["text_transliterated"]])
    packed = io.BytesIO()
    with gzip.GzipFile(fileobj=packed, mode="wb", filename="", mtime=0) as stream:
        stream.write(text.getvalue().encode("utf-8"))
    blob = packed.getvalue()
    manifest = {
        "source_repository": "https://github.com/farhanishmam/BanglaTLit",
        "source_commit": SOURCE_COMMIT,
        "source_path": "data/BanglaTLiT_test.csv",
        "source_sha256": SOURCE_SHA,
        "fixture_sha256": hashlib.sha256(blob).hexdigest(),
        "license": "MIT; Copyright (c) 2024 Aplycaebous",
        "role": "external sentence evaluation only; neither training nor rule tuning",
        "normalization": "native: remove Unicode Cf then NFC; roman, IDs and row order unchanged",
        "unicode_version": unicodedata.unidata_version,
        "rows": len(rows),
        "unique_normalized_native": len({native_key(r["text_bengali"]) for r in rows}),
        "changed_native_rows": sum(native_key(r["text_bengali"]) != r["text_bengali"] for r in rows),
    }
    output.mkdir(parents=True, exist_ok=True)
    (output / "test.csv.gz").write_bytes(blob)
    (output / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (output / "LICENSE").write_bytes((checkout / "LICENSE").read_bytes())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--checkout", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    build(args.checkout, args.output)
