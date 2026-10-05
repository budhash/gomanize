#!/usr/bin/env python3
"""Verify the source-pinned, unreviewed B3 pilot (never a training input)."""
import csv
import gzip
import hashlib
from html.parser import HTMLParser
import io
import json
from pathlib import Path
import unicodedata

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / 'benchmark/data/bengali_lyrics'
STATUS = 'assistant_draft_unreviewed'
FIELDS = ['id', 'song', 'line', 'native', 'roman', 'reference_status']


def sha(blob):
    return hashlib.sha256(blob).hexdigest()


class Poems(HTMLParser):
    """Read only poem divs; br/p/div boundaries terminate verse lines."""
    def __init__(self):
        super().__init__()
        self.depth = 0
        self.text = []

    def handle_starttag(self, tag, attrs):
        if tag == 'div':
            if self.depth:
                self.depth += 1
            elif 'poem' in dict(attrs).get('class', '').split():
                self.depth = 1
        if self.depth and tag in ('br', 'p'):
            self.text.append('\n')

    def handle_endtag(self, tag):
        if self.depth and tag in ('p', 'div'):
            self.text.append('\n')
        if self.depth and tag == 'div':
            self.depth -= 1

    def handle_data(self, data):
        if self.depth:
            self.text.append(data)


def extract(blob):
    parser = Poems()
    parser.feed(blob.decode('utf-8'))
    text = unicodedata.normalize('NFC', ''.join(parser.text))
    text = ''.join(c for c in text if unicodedata.category(c) != 'Cf')
    return [line for raw in text.splitlines() if (line := ' '.join(raw.split()))]


def verify(data=DATA):
    manifest = json.loads((data/'manifest.json').read_text())
    if manifest['reference_status'] != STATUS:
        raise ValueError('pilot references must remain explicitly unreviewed')
    blob = (data/'pilot.csv').read_bytes()
    if sha(blob) != manifest['fixture_sha256']:
        raise ValueError('fixture checksum mismatch')
    reader = csv.DictReader(io.StringIO(blob.decode()))
    if reader.fieldnames != FIELDS:
        raise ValueError('fixture schema mismatch')
    rows = list(reader)
    expected = []
    for song in manifest['songs']:
        compressed = (data/song['snapshot']).read_bytes()
        if sha(compressed) != song['snapshot_sha256']:
            raise ValueError('snapshot checksum mismatch')
        raw = gzip.decompress(compressed)
        if sha(raw) != song['html_sha256']:
            raise ValueError('HTML checksum mismatch')
        lines = extract(raw)
        if len(lines) != song['lines']:
            raise ValueError('source line count mismatch')
        for number, native in enumerate(lines, 1):
            expected.append((f'g1913-{song["song"]:02d}-{number:02d}', str(song['song']), str(number), native))
    if [(r['id'], r['song'], r['line'], r['native']) for r in rows] != expected:
        raise ValueError('native rows differ from captured poems')
    refs = {}
    for row in rows:
        roman = row['roman']
        if row['reference_status'] != STATUS or not roman or roman != roman.strip():
            raise ValueError('invalid reference or status')
        if any('\u0980' <= c <= '\u09ff' for c in roman):
            raise ValueError('reference contains Bengali script')
        if row['native'] in refs and refs[row['native']] != roman:
            raise ValueError('inconsistent repeated-line reference')
        refs[row['native']] = roman
    if len(rows) != manifest['lines'] or len(refs) != manifest['unique_native_lines']:
        raise ValueError('fixture counts mismatch')
    return rows, manifest


if __name__ == '__main__':
    rows, manifest = verify()
    print(f"Verified {len(rows)} lines, {manifest['unique_native_lines']} unique; {STATUS}")
