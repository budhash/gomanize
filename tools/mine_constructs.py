#!/usr/bin/env python3
"""Mine construct-bearing Devanagari tokens for a held-out regression gold set.

Part of F-0010 (structured parser QA), task T-0044. Aggregate accuracy hides
systematic parser bugs in *rare* orthographic constructs (this is how गई→"gi"
went unnoticed — the गई family is ~0.2% of the curated set). This tool builds a
stratified, frequency-ranked candidate list so each target construct class is
represented, then a human validates the romanizations.

Sources are passed in (nothing is hardcoded): any number of Devanagari
frequency lists ("word<TAB>count" or "word,count") and/or plain-text corpora.
Mined tokens are DEDUPED against every benchmark/data/*.csv native column, so
nothing already in the Dakshina train split (the sole training source for the
learned components) or any benchmark can leak into a "held-out" set.

Output is a candidate TSV (to stdout or --out) for human validation:
    class  freq  native  engine_default  engine_best  gold  context  source
The `gold` column is left BLANK for the human to fill (use `|` for accepted
variants, match-any style). `engine_*` are filled by shelling out to a built
./gomanize so the validator sees — but is not anchored to — the current output.

Only VALIDATED rows (native + human romanization = our own work) graduate into
benchmark/data, with a source citation and license check. Raw corpus text is
never committed.

Usage:
    make build
    tools/mine_constructs.py \
        --freq ~/Downloads/clean_lemma_freq.txt \
        --corpus ~/Downloads/RAW.tsv --corpus-col 2 \
        --out candidates.tsv
"""

from __future__ import annotations

import argparse
import csv
import glob
import os
import subprocess
import sys
from collections import defaultdict

# --- Devanagari codepoint sets ------------------------------------------------
DEV_LO, DEV_HI = 0x0900, 0x097F
INDEP_VOWEL = range(0x0904, 0x0915)          # अ .. औ (independent vowels)
MATRAS = set(range(0x093E, 0x094D)) | {0x0962, 0x0963}
HALANT = 0x094D
ANUSVARA = 0x0902
CHANDRABINDU = 0x0901
VISARGA = 0x0903
NUKTA = 0x093C
NUKTA_PRECOMPOSED = set(range(0x0958, 0x0960))  # क़ ख़ ग़ ज़ ड़ ढ़ फ़ य़
DEV_DIGITS = set(range(0x0966, 0x0970))

# Class priority: rarest / most bug-prone first. A token is assigned to the
# first class it matches, so common constructs (conjuncts) don't crowd out the
# rare ones (independent vowels) we actually need coverage for.
CLASS_ORDER = [
    "indep_vowel_medial",  # the गई family: independent vowel after position 0
    "chandrabindu",        # चाँद / कहाँ  (T-0045: nasal dropped mid-word)
    "nukta",               # क़ ज़ ड़  (ड़का→daka today)
    "visarga",             # दुःख
    "anusvara",            # संग / नहीं
    "conjunct",            # क्ष / त्र  (common; low quota)
]
DEFAULT_QUOTAS = {
    "indep_vowel_medial": 30,
    "chandrabindu": 20,
    "nukta": 15,
    "visarga": 5,
    "anusvara": 15,
    "conjunct": 15,
}


def classify(token: str) -> str | None:
    """Return the primary construct class of a token, or None if it targets no
    tracked construct."""
    cps = [ord(c) for c in token]
    has_indep_medial = any(cp in INDEP_VOWEL for cp in cps[1:])
    has_cbindu = CHANDRABINDU in cps
    has_nukta = (NUKTA in cps) or any(cp in NUKTA_PRECOMPOSED for cp in cps)
    has_visarga = VISARGA in cps
    has_anusvara = ANUSVARA in cps
    has_halant = HALANT in cps
    flags = {
        "indep_vowel_medial": has_indep_medial,
        "chandrabindu": has_cbindu,
        "nukta": has_nukta,
        "visarga": has_visarga,
        "anusvara": has_anusvara,
        "conjunct": has_halant,
    }
    for cls in CLASS_ORDER:
        if flags[cls]:
            return cls
    return None


def is_clean_token(token: str) -> bool:
    """Pure Devanagari, length>=2, no digits, no danda/punctuation, well-formed.

    Also rejects a common corpus encoding artifact: an independent vowel
    immediately after a halant (e.g. कोर्इ, गर्इ — where ई was split into
    र्+इ). A halant must attach to a following consonant, so halant→independent
    vowel is malformed Devanagari, not a real construct."""
    if len(token) < 2:
        return False
    cps = [ord(c) for c in token]
    for i, cp in enumerate(cps):
        if cp < DEV_LO or cp > DEV_HI:
            return False
        if cp in DEV_DIGITS or cp in (0x0964, 0x0965):  # digits, danda, double danda
            return False
        if cp in INDEP_VOWEL and i > 0 and cps[i - 1] == HALANT:
            return False
    return True


def load_benchmark_natives(repo_root: str) -> set[str]:
    """Every native (col 0) across benchmark/data/*.csv — the dedup/contamination
    guard."""
    natives: set[str] = set()
    pattern = os.path.join(repo_root, "benchmark", "data", "*.csv")
    for path in glob.glob(pattern):
        with open(path, encoding="utf-8", newline="") as fh:
            reader = csv.reader(fh)
            next(reader, None)  # skip header
            for row in reader:
                if row:
                    natives.add(row[0].strip())
    return natives


def iter_freq(path: str):
    """Yield (token, count) from a 'word<TAB>count' or 'word,count' file."""
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            line = line.rstrip("\n")
            if not line:
                continue
            if "\t" in line:
                parts = line.split("\t")
            else:
                parts = line.split(",")
            token = parts[0].strip()
            count = 0
            if len(parts) > 1 and parts[1].strip().isdigit():
                count = int(parts[1].strip())
            if token:
                yield token, count


def iter_corpus_tokens(path: str, col: int | None):
    """Yield (token, line) from a corpus. If col is set, the file is TSV and
    tokens come from that 1-based column; otherwise the whole line is tokenized.
    Splits on any non-Devanagari run."""
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            line = line.rstrip("\n")
            text = line
            if col is not None:
                cells = line.split("\t")
                if len(cells) < col:
                    continue
                text = cells[col - 1]
            # split into maximal Devanagari runs
            cur = []
            for c in text:
                if DEV_LO <= ord(c) <= DEV_HI:
                    cur.append(c)
                else:
                    if cur:
                        yield "".join(cur), text
                        cur = []
            if cur:
                yield "".join(cur), text


def run_engine(binary: str, tokens: list[str], flags: list[str]) -> list[str]:
    """Romanize tokens (one per line) through the built CLI, preserving order."""
    if not tokens:
        return []
    proc = subprocess.run(
        [binary, *flags],
        input="\n".join(tokens) + "\n",
        capture_output=True,
        text=True,
        check=True,
    )
    out = proc.stdout.split("\n")
    # drop a single trailing empty line from the final newline
    if out and out[-1] == "":
        out = out[:-1]
    if len(out) != len(tokens):
        raise SystemExit(
            f"engine returned {len(out)} lines for {len(tokens)} tokens "
            f"(flags={flags}); refusing to misalign"
        )
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--freq", action="append", default=[],
                    help="frequency list: 'word<TAB>count' or 'word,count' (repeatable)")
    ap.add_argument("--corpus", action="append", default=[],
                    help="plain-text or TSV corpus to tokenize (repeatable)")
    ap.add_argument("--corpus-col", type=int, default=None,
                    help="1-based TSV column to read from each --corpus (default: whole line)")
    ap.add_argument("--binary", default="./gomanize", help="built gomanize CLI")
    ap.add_argument("--best-flags", default="--lexicon --rerank",
                    help="flags for the 'engine_best' column")
    ap.add_argument("--per-class", type=int, default=None,
                    help="override every class quota with this number")
    ap.add_argument("--out", default="-", help="output TSV path (default: stdout)")
    args = ap.parse_args()

    if not args.freq and not args.corpus:
        ap.error("supply at least one --freq or --corpus source")

    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    dedup = load_benchmark_natives(repo_root)
    print(f"[mine] dedup guard: {len(dedup)} benchmark natives", file=sys.stderr)

    quotas = dict(DEFAULT_QUOTAS)
    if args.per_class is not None:
        quotas = {k: args.per_class for k in quotas}

    # Collect candidates: token -> {class, freq, context, source}. Frequency
    # sources rank by real-world count; corpus tokens contribute context lines
    # and on-domain (verse) coverage.
    cand: dict[str, dict] = {}
    seen = set(dedup)

    def consider(token, count, context, source):
        if token in seen and token not in cand:
            return
        if not is_clean_token(token):
            return
        cls = classify(token)
        if cls is None:
            return
        if token not in cand:
            cand[token] = {"class": cls, "freq": count, "context": context, "source": source}
            seen.add(token)
        else:
            # keep the higher freq and first non-empty context
            if count > cand[token]["freq"]:
                cand[token]["freq"] = count
            if context and not cand[token]["context"]:
                cand[token]["context"] = context

    for fp in args.freq:
        n = 0
        for token, count in iter_freq(fp):
            consider(token, count, "", os.path.basename(fp))
            n += 1
        print(f"[mine] {os.path.basename(fp)}: scanned {n} freq entries", file=sys.stderr)

    for cp in args.corpus:
        n = 0
        for token, line in iter_corpus_tokens(cp, args.corpus_col):
            consider(token, 0, line.strip()[:120], os.path.basename(cp))
            n += 1
        print(f"[mine] {os.path.basename(cp)}: scanned {n} corpus tokens", file=sys.stderr)

    # Rank within each class by frequency (desc), then apply quota.
    by_class: dict[str, list] = defaultdict(list)
    for token, meta in cand.items():
        by_class[meta["class"]].append((token, meta))
    selected = []
    for cls in CLASS_ORDER:
        rows = sorted(by_class.get(cls, []), key=lambda kv: kv[1]["freq"], reverse=True)
        selected.extend(rows[: quotas.get(cls, 0)])
        print(f"[mine] {cls}: {len(rows)} candidates -> kept {min(len(rows), quotas.get(cls,0))}",
              file=sys.stderr)

    natives = [t for t, _ in selected]
    eng_def = run_engine(args.binary, natives, [])
    eng_best = run_engine(args.binary, natives, args.best_flags.split())

    out = sys.stdout if args.out == "-" else open(args.out, "w", encoding="utf-8")
    w = csv.writer(out, delimiter="\t")
    w.writerow(["class", "freq", "native", "engine_default", "engine_best",
                "gold", "context", "source"])
    for (token, meta), ed, eb in zip(selected, eng_def, eng_best):
        w.writerow([meta["class"], meta["freq"], token, ed, eb, "",
                    meta["context"], meta["source"]])
    if out is not sys.stdout:
        out.close()
        print(f"[mine] wrote {len(selected)} candidates to {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
