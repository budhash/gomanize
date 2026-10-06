# Bengali B2 runtime selector and external validation

T-0066 integrates the frozen cross-fitted selector as an experimental opt-in.
Enable **both `SchwaModel` and `Rerank`** through the Bengali Go API. No default
changes, training, threshold tuning, or CLI/npm/web language-selection expansion
are included. The selector artifact is byte-identical to the T-0065 artifact;
threshold 0.6 and eight possible slot flips remain fixed.

## Runtime contract

Core adds optional `NativeReranker` and its synchronous `CandidateRenderer`
callback. The native provider takes precedence over legacy `Reranker`, including
on a decline. A decline resumes the original pipeline with the original options.
An accepted result returns nil debug information rather than pretending one
candidate trace explains the whole selection. Lexicon hits still win first.
Hindi retains its legacy character-LM path unchanged.

Candidate rendering constructs a fresh engine with the caller's scheme and
current rule enable/disable state. It suppresses recursive reranking, repeat
lexicon lookup and candidate debug traces. Bengali's candidate language replaces
only the existing model rule's action; its condition, priority and pending-slot
ownership remain intact. Each closure carries one flip index and no shared
mutable word/model state.

Bengali generates model, B1, then source-ordered single-slot flips, deduplicates
strings, and selects only when preference strictly exceeds 0.6 and the current
best. It uses native/edit features and stable candidate order for ties. Canonical
aliases and format-character removal use the existing Bengali normalization.
Unsupported words or any `InherentVowelA`, `LongVowels`, `SimpleNasals`, or
`KeepMedialSchwa` setting bypass selection. `Rerank` without `SchwaModel` also
retains the existing output. Malformed embedded selector/model data declines.

## Frozen parity and measured results

The Go selector matches the offline Python selector **word for word on all
5,000 dev/test words**, not just in aggregate. Lexicon lookup has zero effect on
these excluded words. All results below use no overrides.

| Dakshina split | Model strict → rerank strict | Model match-any → rerank match-any | Mean minCER before → after | Wins / losses / changed |
|---|---:|---:|---:|---:|
| Dev, 2,500 | 894 → 897 | 1,544 → 1,552 (61.76% → 62.08%) | 0.07851746 → 0.07820237 | 10 / 2 / 21 |
| Test, 2,500 | 914 → 916 | 1,569 → 1,576 (62.76% → 63.04%) | 0.07642207 → 0.07572118 | 8 / 1 / 21 |

Wins/losses refer to match-any, and changed includes cases that remain wrong or
remain accepted variants. The seven-word net held-out gain is a small point
estimate; statistical significance is not claimed.

| External set / profile | Match-any before → after | Mean minCER before → after | Wins / losses / changed |
|---|---:|---:|---:|
| Aksharantar, all 7,134 types, pure | 2,200 → 2,205 (30.84% → 30.91%) | 0.21640005 → 0.21603412 | 8 / 3 / 41 |
| Aksharantar, 6,036 unseen types, pure | 2,052 → 2,058 (34.00% → 34.10%) | 0.20464678 → 0.20429022 | 8 / 2 / 33 |
| Aksharantar, all types, with lexicon | 2,296 → 2,302 (32.18% → 32.27%) | 0.21117049 → 0.21077652 | 8 / 2 / 40 |
| BanglaTLit, 2,500 sentences, pure | 12 → 12 exact | 0.31286250 → 0.31286312 | 0 / 0 / 23 |
| BanglaTLit, with lexicon | 17 → 17 exact | 0.28142890 → 0.28144658 | 0 / 0 / 21 |

**BanglaTLit character error worsens slightly**, despite the word-level gains.
No threshold or feature was adjusted to hide that result. The selector remains
opt-in; this is not a blanket recommendation for lyrics. Independent public-domain
lyrics validation is the next stage (T-0054), including these sentence tradeoffs.
The 38 BanglaTLit sentences whose every Bengali token is unseen remain unchanged
(2 exact; mean minCER 0.52095747), an insufficient sample for broad conclusions.

Aksharantar contains all Dakshina test words, so it is not an independent held-out
replication. “Unseen” excludes the union of frozen Google and Dakshina training
vocabularies, not merely successfully aligned Google types. Sentence references
do not establish token-level accuracy. Lexicon gains include known spellings and
must not be described as unseen-word generalization.

The [machine-readable report](2026-10-05-bengali-b2-runtime.json) records every
profile, strict Dakshina results, paired outcomes, source hashes and data hashes.
The [artifact record](../../lang/bengali/RERANKER.md) retains CC BY-SA attribution
to Dakshina and CC BY attribution to Google's Bengali pronunciation source.

## Verification and reproduction

- 583 synthetic cases verify native/edit features, candidate order, outputs,
  Unicode aliases and fallback styles; 15 select an alternative. Cases are built
  from synthetic words and frozen-tree paths, not held-out reference mining.
- Core tests verify native-provider precedence, option forwarding, scheme/rule
  controls, lexicon priority, recursion prevention and decline behavior.
- Bengali tests cover malformed trees, disabled model rules, debug semantics,
  shared-engine concurrent calls, and punctuation/whitespace preservation.
- Benchmark tests enforce frozen dev/test metrics and held-out lexicon isolation.
- Python tests verify artifact identity, synthetic fixture reconstruction and
  external scoring denominators/paired losses.
- Negative controls remove inherited rule controls and set the selector threshold
  to 1.0 using temporary Go overlays. Both corresponding tests fail; repository
  source files are not changed for the negative controls.
- Full local `make ci` runs the race/coverage suites, Hindi frozen snapshot,
  accuracy benchmarks and all Bengali gates. WASM/npm smoke validation checks
  that the existing Hindi surface still builds and runs.

```sh
GOCACHE=/tmp/gomanize-review-go-cache make test-bengali-runtime
GOCACHE=/tmp/gomanize-review-go-cache python3 tools/bengali/evaluate_runtime.py \
  /path/to/pinned/ben.zip --output /tmp/bengali-runtime.json
cmp /tmp/bengali-runtime.json docs/reviews/2026-10-05-bengali-b2-runtime.json
make ci
make npm-test
```

The runtime evaluator does no fitting or selection. It checks artifact identity,
compares every Dakshina output against the independent offline renderer/selector,
and separately reports full/unseen external slices. The historical frozen
experiment's source-hash guard intentionally identifies its original source
revision; this runtime validation has its own source provenance.

## Review notes (2026-10-05)

Independent stack review. Tables above record the runtime as submitted.

- **Regenerated after the B1 review fixes** (canonical nukta letters; explicit
  final hasant). The selector is unchanged and word-for-word runtime/offline parity
  still holds on all 5,000 dev and test words. Dakshina dev: model 895 / 1,545 /
  0.07839302 → rerank 898 / 1,553 / 0.07807792 (B1 775 / 1,371); **test unchanged**
  (914 → 916 strict, 1,569 → 1,576 match-any, 0.07572118). Aksharantar full
  2,201 → 2,206 (minCER 0.21617089 → 0.21580496), unseen 2,053 → 2,059, with lexicon
  2,297 → 2,303. BanglaTLit pure minCER 0.31285796 → 0.31285858 (still a slight
  regression, +6.2e-7), with lexicon 0.28142436 → 0.28144204. Paired win/loss and
  exact-match counts are unchanged.
- **Records are now asserted in CI.** `test_committed_record_reproduces` recomputes
  the Dakshina and BanglaTLit sections (no external archive needed) and requires the
  committed record to match; the stale record failed it. Aksharantar needs the pinned
  archive and remains an offline step. The Go test now also pins the embedded
  selector's SHA-256.
- **Licensing.** `selector.json` (CC BY-SA 4.0) and `vowel_tree.json` (CC BY 4.0)
  are embedded in every build; the README License section now names both.
- Tracked: threshold and tie behaviour are protected by the aggregate dev/test pins
  rather than unit fixtures; enabling Rerank costs about 6× time per word (a fresh
  engine per candidate).
