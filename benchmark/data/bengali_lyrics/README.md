# Bengali B3 lyrics pilot — unreviewed references

This is an evaluation-only pilot, **not human-attested gold**. T-0067 builds the
pilot; T-0068 covers independent Bengali reference review. T-0054 stays open
until that review establishes a defensible gold set. Do not train on this data
or tune rules, thresholds, or references to improve its scores.

## Source and attribution

Rabindranath Tagore, *Gitanjali*, third edition, Indian Press, Allahabad, 1913.
The first four complete numbered songs were selected in edition order before
running predictions. They contain 82 line occurrences and 71 unique native
lines; repeated refrains remain in their original positions.

Text is transcribed by **Bengali Wikisource contributors**. The
[source edition](https://bn.wikisource.org/wiki/গীতাঞ্জলি_(১৯১৩)) identifies the
original work as public domain and dates the edition to 1913. This import
preserves Wikisource's [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)
terms for the transcriptions and adaptations, including `pilot.csv`. Dataset
licensing is separate from the repository's MIT-licensed code. Changes made
here: HTML verse extraction, whitespace/Unicode normalization, identifiers,
and newly drafted Roman references. No upstream Roman transcription was used.

| Song | Source revision | Contributor history | Lines |
|---|---|---|---:|
| 1 | [1936474](https://bn.wikisource.org/w/index.php?oldid=1936474) | [History](https://bn.wikisource.org/w/index.php?title=গীতাঞ্জলি_(১৯১৩)/১&action=history) | 20 |
| 2 | [1936544](https://bn.wikisource.org/w/index.php?oldid=1936544) | [History](https://bn.wikisource.org/w/index.php?title=গীতাঞ্জলি_(১৯১৩)/২&action=history) | 20 |
| 3 | [1936555](https://bn.wikisource.org/w/index.php?oldid=1936555) | [History](https://bn.wikisource.org/w/index.php?title=গীতাঞ্জলি_(১৯১৩)/৩&action=history) | 20 |
| 4 | [1936566](https://bn.wikisource.org/w/index.php?oldid=1936566) | [History](https://bn.wikisource.org/w/index.php?title=গীতাঞ্জলি_(১৯১৩)/৪&action=history) | 22 |

`manifest.json` records capture date, source revisions, compressed snapshot and
uncompressed HTML checksums, and the fixture checksum. Old revisions can
transclude changing pages/templates; **the committed HTML captures, not merely
the oldid URLs, pin the text**. The captures retain transcluded page names and
source links for attribution and scan checking. They are data, not executable
web assets. A fresh download may differ and must not silently replace them.

## Text and reference protocol

`tools/bengali_lyrics.py` extracts only `div.poem` blocks. Verse breaks come from
`br`, paragraph/div boundaries and source newlines; empty lines are dropped.
NFC, removal of Unicode format characters, and collapsed/trimmed whitespace
remove layout artifacts. Indentation is not part of the fixture. Punctuation,
hyphens, apostrophes and old spellings remain, including শকতি and অহঙ্কার.

Every Roman line was drafted by the assistant directly from the captured
Bengali text **before any pilot engine predictions**, then frozen under SHA-256
`44b33280eb0cdd342d12a58e083f008953b83a1379a9e5c724d39a7fedf424ae`.
All rows say `assistant_draft_unreviewed`. This is neither independent human
attestation nor a claim that the assistant has never encountered these famous
songs in pretraining.

The draft uses lowercase colloquial Latin letters, generally without vowel
length doubling, `sh` for the common sibilant and pronunciation-oriented
clusters such as `rokkha`/`poddo`. Native punctuation is copied, including danda
and curly apostrophe; the reference is not fully ASCII. Nasalization and
poetic/archaic forms need particular review (`thain`, `danrao`, `shokoti`).
These are proposed conventions, not a validated Bengali standard. Single
references do not accept alternate valid spellings.

## Reproduce

From the repository root:

```sh
python3 tools/bengali_lyrics.py
make test-bengali-lyrics
python3 tools/bengali/evaluate_lyrics.py --output /tmp/bengali-lyrics.json
```

Scoring uses Unicode character Levenshtein distance, including punctuation and
spaces. Macro line CER averages `edits/reference characters` over lines; micro
CER divides total edits by total reference characters. Exact-line agreement,
per-song scores, equal-weight song macro CER, deduplicated native-line scores,
and an all-tokens-unseen slice are separate. Empty slices report null rates.
No quality threshold is enforced while references are unreviewed.

Training exposure uses normalized Bengali letter/mark tokens and the union of
frozen Google/Dakshina training vocabularies. Source independence alone does
not imply word-level novelty. The evaluator records every prediction for
recomputation, source/artifact hashes, and paired reranker wins **and losses**.
The Go suite independently computes macro CER and exercises multiline public
API rendering for all six profiles. Python tests reject corrupt snapshots,
changed references, source/native mismatches even after rehashing, and false
reference status; they check scoring denominators and duplicate weighting.

## Review handoff (T-0068)

Start with `pilot.csv` and the source links, without consulting engine
predictions. A fluent Bengali reviewer should check all 82 lines for faithful
source extraction, pronunciation, poetic forms, nasalization, word boundaries
and punctuation; decide acceptable spelling variants and a consistent style;
then record reviewer identity/attestation, date and per-row corrections in a
versioned review artifact. Check the linked scans when transcription is
uncertain. Repeated lines must receive consistent decisions.

Reference corrections must have linguistic/source reasons, never improved
engine agreement. Preserve this version and its report, bump dataset version,
and publish a new hash and evaluation after review. A review does not by itself
make four Tagore songs representative of modern lyrics, other poets or regional
pronunciation. Broader coverage remains part of the B3 completion review.
