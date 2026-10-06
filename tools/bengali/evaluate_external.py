#!/usr/bin/env python3
"""Score the frozen B2 vowel model on external corpora; never select a model here."""
import argparse
import csv
import gzip
import io
import json
from pathlib import Path
import subprocess
import sys
import zipfile
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, pinned, sha, verify
from audit_bengali_training_overlap import bengali_tokens


def distance(a, b):
    previous = list(range(len(b)+1))
    for i, x in enumerate(a, 1):
        current = [i]
        for j, y in enumerate(b, 1):
            current.append(min(current[-1]+1, previous[j]+1, previous[j-1]+(x != y)))
        previous = current
    return previous[-1]


def score(rows, predictions, with_lexicon=False):
    result = {'items': len(rows)}
    for column, name in enumerate(('B1', 'model', 'lexicon', 'model_lexicon') if with_lexicon else ('B1', 'model')):
        matches = 0
        cer = 0.0
        for (native, references), values in zip(rows, predictions):
            out = values[column]
            matches += out in references
            cer += min(distance(out, ref)/max(len(ref), 1) for ref in references)
        result[name] = {'match_any': matches, 'accuracy': matches/len(rows) if rows else None,
                        'macro_minCER': cer/len(rows) if rows else None}
    if with_lexicon:
        wins = losses = changed = 0
        for (_, references), values in zip(rows, predictions):
            before, after = values[1] in references, values[3] in references
            wins += after and not before
            losses += before and not after
            changed += values[1] != values[3]
        result['lexicon_added_to_model'] = {'wins': wins, 'losses': losses, 'changed': changed}
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('aksharantar_zip', type=Path)
    parser.add_argument('--lexicon', action='store_true', help='include isolated lexicon and combined scores')
    args = parser.parse_args()
    raw = pinned(args.aksharantar_zip, '4ab6edcc6ab556040d8f43ee75eec73cc25af0cf72d09e4650dab26b196d5fe7')
    with zipfile.ZipFile(io.BytesIO(raw)) as archive:
        member = archive.read('ben_test.json')
    from bengali_training_data import native_key
    words = {}
    for line in member.decode().splitlines():
        row = json.loads(line)
        words.setdefault(native_key(row['native word']), set()).add(row['english word'])
    ak = [(word, sorted(refs)) for word, refs in sorted(words.items())]
    blob = pinned(ROOT/'benchmark/data/banglatlit/test.csv.gz', '386721b4221e2d93b16ab73f3761c7c4e4cd295adf1160fe9c95c06cb34c5386')
    bangla = [(r['native'], [r['roman']]) for r in csv.DictReader(io.StringIO(gzip.decompress(blob).decode()))]
    rows = ak + bangla
    encoded = ''.join(json.dumps(word, ensure_ascii=False)+'\n' for word, _ in rows)
    proc = subprocess.run(['go', 'run', './tools/bengali/evaluate'], cwd=ROOT, input=encoded,
                          text=True, capture_output=True, check=True)
    predictions = [json.loads(line) for line in proc.stdout.splitlines()]
    if len(predictions) != len(rows):
        raise ValueError('prediction count mismatch')
    groups = verify()
    training = groups['google-train'] | groups['dakshina-train']
    selected = [(r, p) for r, p in zip(ak, predictions[:len(ak)]) if r[0] not in training]
    unseen = score([r for r, _ in selected], [p for _, p in selected], args.lexicon)
    # Sentence-only references cannot establish token-level correctness.
    selected = [(r, p) for r, p in zip(bangla, predictions[len(ak):])
                if (tokens := set(bengali_tokens(r[0]))) and not tokens & training]
    unseen_sentences = score([r for r, _ in selected], [p for _, p in selected], args.lexicon)
    result = {'model_sha256': sha((ROOT/'lang/bengali/vowel_tree.json').read_bytes()),
              'aksharantar_archive_sha256': sha(raw), 'aksharantar_test_sha256': sha(member),
              'banglatlit_fixture_sha256': sha(blob),
              'unseen_definition': 'absent from union of frozen Google and Dakshina training vocabularies (conservative; not just successfully aligned Google types)',
              'Aksharantar': {'full_types': score(ak, predictions[:len(ak)], args.lexicon), 'unseen_types': unseen},
              'BanglaTLit': {'full_sentences': score(bangla, predictions[len(ak):], args.lexicon), 'all_tokens_unseen_sentences': unseen_sentences,
                            'limitation': 'sentence-level Roman references do not support isolated unseen-token accuracy'},
              'lyrics_gold': 'not yet built (T-0054)'}
    if args.lexicon:
        lexicon_data = (ROOT/'lang/bengali/lexicon.tsv').read_bytes()
        result['lexicon_sha256'] = sha(lexicon_data)
        entries = {line.split('\t')[0] for line in lexicon_data.decode().splitlines()}
        result['Aksharantar']['lexicon_hit_types'] = sum(word in entries for word, _ in ak)
        result['lexicon_policy'] = 'fixed >=3 attestations and unique winner; no external tuning'
        if any(p[0] != p[2] or p[1] != p[3] for r, p in zip(ak, predictions[:len(ak)]) if r[0] not in training):
            raise ValueError('lexicon changed an unseen evaluation word')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
