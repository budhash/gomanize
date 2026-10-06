# B3: source-pinned lyrics pilot (T-0067)

The lyrics measurement path now runs end to end against 82 lines from the first
four songs in Tagore's 1913 *Gitanjali*, with committed source captures and six
runtime profiles. **The Roman references are assistant drafts, not reviewed
gold.** T-0068 (independent Bengali review) and T-0054 (B3 gold completion)
remain open. This PR changes evaluation tooling and data only.

[Dataset, attribution, frozen-reference protocol and review instructions](../../benchmark/data/bengali_lyrics/README.md).
[Machine-readable results, per-line predictions and source/artifact hashes](2026-10-05-bengali-b3-lyrics.json).

## Provisional reference agreement

All 82 occurrences count here; macro line CER includes punctuation and spaces.
These measurements describe agreement with one draft spelling per line, not
validated language accuracy or a new release gate.

| Profile | Exact lines / 82 | Macro line CER | Micro CER |
|---|---:|---:|---:|
| B1 pure rules | 43 | 0.021638 | 0.022478 |
| Vowel model, no lexicon | 52 | 0.017664 | 0.018482 |
| Model + native reranker, no lexicon | 52 | 0.017013 | 0.017982 |
| Rules + lexicon | 37 | 0.029390 | 0.029970 |
| Model + lexicon | 47 | 0.025304 | 0.025974 |
| Model + reranker + lexicon | 46 | 0.025234 | 0.025974 |

The pure reranker changes three lines: two improve CER and one worsens; exact
agreement has one win and one loss. With the lexicon, two lines change, one
improves CER and one worsens, with zero exact wins and one exact loss. Lexicon
profiles agree less with this draft than their corresponding pure profiles.
No runtime setting, reference, threshold or spelling exception was changed in
response to these results.

Examples for review: the lexicon changes `shokol` to `sakal` in three repeated
refrains, and `akash` to `aakash` in one line. These illustrate spelling-style
differences against the draft, not proven linguistic errors. The pure reranker
changes `korechho` to `korechh` (an exact-agreement loss), `jonmer` to `jonomer`
(lower CER), and `laghobo` to `laghob` (an exact-agreement win).

Refrains create 11 extra occurrences. Deduplicating to 71 native lines gives
42 exact lines for both model profiles; macro CER is 0.019730 for the model and
0.018978 with reranking (B1: 37 exact, CER 0.021711). Equal-weight song macro
CER is 0.017341 and 0.016727 respectively. The JSON includes all six profiles
for each slice and each song.

## Exposure and limits

There are 213 normalized native token types and 348 token occurrences. Of these,
142 types and 237 occurrences appear in the frozen Google/Dakshina training
union. Only seven complete lines have every token absent from that union:
model/reranker both match six, with macro CER 0.006803. That tiny slice is
insufficient evidence of generalization. The audit also discloses overlap with
existing dev/test vocabularies; it does not equate source novelty to unseen
words or guarantee independence from assistant pretraining.

Four adjacent songs by one poet are a reproducible starting sample, not broad
lyrics coverage. The single draft references omit valid variants and require
fluent review of nasals, clusters and archaic pronunciation. The strong draft
agreement must not be compared directly with Hindi's human-reviewed lyrics or
Bengali's Dakshina word metrics. Bengali has reached the lyrics **measurement**
stage; reviewed lyrics quality remains unestablished.

## Verification and follow-through

- Python integrity and scoring tests cover source extraction, hashes, status,
  repetition, unequal reference lengths, unequal song lengths and empty slices.
  Negative controls deliberately corrupt a snapshot/reference and rehash
  tampered native/status rows; all are rejected.
- The Go public API suite recomputes and logs the six macro CER results and
  exact counts, and verifies multiline rendering preserves per-line output.
  (As of the 2026-10-05 review, the Python suite asserts the record; see below.)
- Full local `make ci` passed, including the new pilot verification target and
  the frozen Hindi replay (689.700 seconds under race/coverage). Hindi remains
  1,147/1,330 strict and 1,236/1,330 match-any on curated Dakshina. No runtime
  code or learned artifact changed.
- T-0068 requires a Bengali reviewer to correct/attest the source-first worksheet
  before reference promotion. T-0054 remains open for gold and coverage review.
  Do not tune on the pilot while that work is pending.

## Review notes (2026-10-05)

Independent stack review; reference quality is out of scope (T-0068).

- **Record now asserted against the engine.** The Go test only logged its scores
  and the Python test re-scored the saved predictions, so a changed lexicon entry
  (সকল → `sokkol`) moved the logged CER with every test green.
  `test_record_matches_current_engine` re-runs the engine and requires the
  committed predictions, every slice, song weighting and overlap audit to match
  (provenance hashes excluded); the mutation now fails.
- **Regenerated.** After the B1/B2 review fixes every prediction and score is
  unchanged; only `sources_sha256` was stale and has been refreshed.
- Tracked: extraction strips every Unicode Cf character, including ZWJ/ZWNJ, which
  can be meaningful in Bengali (only a BOM occurs in the current captures); about
  16% of scored reference characters are spaces/punctuation, so a letters-only CER
  would help T-0068; full-page captures (~17 KB each) will add up if the pilot grows.
