#!/usr/bin/env python3
"""Build deterministic, split-preserving Bengali Dakshina benchmark fixtures.

Source: Google Dakshina v1.0, CC BY-SA 4.0; see benchmark/data/bengali/README.md.
No attestation filtering, spelling selection, or training is performed here.
"""
import argparse
import csv
import gzip
import hashlib
import io
import json
from pathlib import Path
import tarfile
import unicodedata


def native_key(text):
    """Canonical comparison key; matches the parser's removal of format marks."""
    return unicodedata.normalize("NFC", "".join(c for c in text if unicodedata.category(c) != "Cf"))


def parse_split(raw):
    counts = {}
    source_rows = 0
    changed_rows = 0
    for row in csv.reader(io.StringIO(raw.decode("utf-8")), delimiter="\t"):
        if len(row) != 3 or not row[0] or not row[1]:
            raise ValueError(f"invalid lexicon record at row {source_rows + 1}")
        count = int(row[2])
        if count <= 0:
            raise ValueError("attestations must be positive")
        native = native_key(row[0])
        if not native:
            raise ValueError("normalization produced empty native text")
        source_rows += 1
        changed_rows += native != row[0]
        key = native, row[1]
        counts[key] = counts.get(key, 0) + count
    return counts, source_rows, changed_rows


def build(archive, output):
    datasets = {}
    manifest = {
        "source": "https://storage.googleapis.com/gresearch/dakshina/dakshina_dataset_v1.0.tar",
        "license": "CC-BY-SA-4.0",
        "normalization": "remove Unicode Cf; NFC native only; aggregate duplicate native/roman pairs; preserve roman case",
        "unicode_version": unicodedata.unidata_version,
        "splits": {},
    }
    seen = set()
    with tarfile.open(archive, "r:") as tar:
        for split in ("train", "dev", "test"):
            member = f"dakshina_dataset_v1.0/bn/lexicons/bn.translit.sampled.{split}.tsv"
            source = tar.extractfile(member)
            if source is None:
                raise ValueError(f"missing {member}")
            raw = source.read()
            counts, source_rows, changed = parse_split(raw)
            natives = {n for n, _ in counts}
            if seen & natives:
                raise ValueError("normalized word types overlap across splits")
            seen.update(natives)
            buf = io.StringIO(newline="")
            writer = csv.writer(buf, lineterminator="\n")
            writer.writerow(["native", "roman", "attestations"])
            for (native, roman), count in sorted(counts.items()):
                writer.writerow([native, roman, count])
            packed = io.BytesIO()
            with gzip.GzipFile(filename="", mode="wb", fileobj=packed, mtime=0) as gz:
                gz.write(buf.getvalue().encode("utf-8"))
            datasets[split] = packed.getvalue()
            manifest["splits"][split] = {
                "member": member,
                "source_sha256": hashlib.sha256(raw).hexdigest(),
                "fixture_sha256": hashlib.sha256(datasets[split]).hexdigest(),
                "source_rows": source_rows,
                "normalized_rows": len(counts),
                "normalized_words": len(natives),
                "changed_native_rows": changed,
                "native_keys_sha256": hashlib.sha256("\n".join(sorted(natives)).encode("utf-8")).hexdigest(),
            }
    output.mkdir(parents=True, exist_ok=True)
    for split, data in datasets.items():
        (output / f"{split}.csv.gz").write_bytes(data)
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2, ensure_ascii=False) + "\n")
    for split, info in manifest["splits"].items():
        print(f"{split}: {info['normalized_words']} words, {info['normalized_rows']} references, {len(datasets[split])} compressed bytes")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--archive", type=Path, default=Path("datasets/dakshina/dakshina.tar"))
    parser.add_argument("--output", type=Path, default=Path("benchmark/data/bengali"))
    args = parser.parse_args()
    build(args.archive, args.output)
