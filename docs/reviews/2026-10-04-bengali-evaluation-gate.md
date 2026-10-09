# Bengali default, curation, and preregistered B1 gate

T-0055 decision, recorded before any production B1 rule changes. The parent is
`6a82b6ac5fddc597f1bab70b976e010f0973e0cd` (B0 plus test-only structural
prototypes). This closes the measurement/default/curation prerequisite for
T-0052. It does not claim B1 is implemented or passes its gate.

## Measurement protocol

`TestBengaliVowelEvidence` opens only the committed train and dev fixtures. It
marks the B0 renderer's retained consonant vowels separately from independent
অ. Matras and independent ও stay literal. Each reference must match every
non-marked character exactly; a marked slot may match `o`, `a`, or nothing.
The matcher considers **all** valid alignments. If a slot has multiple possible
values, it records an ambiguous mask rather than choosing a convenient path.
Synthetic tests cover ambiguous adjacent slots, omissions, matra ownership,
and mismatched consonant anchors.

These are conditional spelling counts, not phoneme labels or a census of all
Bengali vowels. B0-deleted final vowels are not slots, consonant errors exclude
references, and matras remain fixed anchors. Coverage is therefore part of the
result, not an accuracy metric. No test examples or misses are inspected.

| Coverage | Train | Dev |
|---|---:|---:|
| Total words | 25,000 | 2,500 |
| Words with a marked slot | 14,830 | 1,394 |
| Marked slots | 23,188 | 2,188 |
| Slot-bearing words with an aligned reference | 8,736 | 902 |
| All references | 94,546 | 9,279 |
| Aligned references, including words without slots | 22,027 | 2,303 |

The following counts weight each aligned slot/reference by that reference's
attestations. An omitted slot is separate from a spelling preference.

| Slot / spelling | Train | Dev |
|---|---:|---:|
| Retained consonant vowel: o | 18,113 | 1,837 |
| Retained consonant vowel: a | 11,593 | 1,227 |
| Retained consonant vowel: omitted | 7,718 | 887 |
| Independent অ: o | 888 | 110 |
| Independent অ: a | 1,257 | 119 |
| Independent অ: omitted | 13 | 0 |
| Ambiguous slot assignments | 0 | 0 |

Zero ambiguity here is an observed property of the matched subset; the matcher
and synthetic ambiguity tests do not assume it. Giving each slot-bearing word
one vote by its larger o/a attestation total yields train o/a/tie/no-informative
counts of 4,329/1,547/1,377/7,577 and dev 436/153/127/678. This sensitivity check
still conditions on exact anchors and is not an unbiased population estimate.

## Default decision

Keep **o** as the operational B1 default for retained consonant vowels. Both the
conditional slot counts and whole-word results favor it: full dev match-any is
1,210 with default o versus 1,107 with alternate a, and minCER is 0.10897958
versus 0.11809216. Preserve `InherentVowelA` as an explicit style option.

Independent অ does not show the same preference. The existing public default
continues to render it as o for the initial unified style; this is a product
convention, not a claim that the independent-vowel data favors o. Any later
independent-vowel exception needs separate evidence and before/after measurement.
No word exceptions or learned components are introduced by this decision.

## Curation decision

The primary gate remains **all 2,500 dev words**. The secondary curated subset
includes a word iff at least one reference has **three or more attestations**;
it retains **all** that word's references. Do not sum votes across variants for
eligibility, discard low-vote references, or select on whether B0 gets it right.

| Maximum reference votes at least | Train words | Dev words | Dev o match-any | Dev o minCER |
|---|---:|---:|---:|---:|
| 1 | 25,000 | 2,500 | 1,210 | 0.10897958 |
| 2 | 20,576 | 2,088 | 1,067 | 0.10662244 |
| 3 | 9,877 | 992 | 545 | 0.11405074 |
| 4 | 1,214 | 103 | 53 | 0.08997623 |

Three retains 39.68% of dev and nearly ten times the four-vote subset's size.
It supplies a useful secondary slice with repeated support for one spelling.
It is not a validated human gold standard or a replacement for full dev. This
choice is based on support/coverage, not maximizing B0 scores: the four-vote
slice actually favors alternate a on minCER (0.08316694 versus 0.08997623).

## B1 acceptance contract

The numerical bar is an **engineering target**, not a statistically inferred
natural rule ceiling. A 10% net reduction of the baseline's word misses requires
substantially more than a few corrected examples; the 5% minCER reduction checks
whether partially wrong outputs improve too. Both full dev and curated dev
must pass, without losing strict top-1 matches.

| Gate | Full dev (2,500) | Curated dev (992) |
|---|---:|---:|
| B0 match-any | 1,210 | 545 |
| Required B1 match-any | **≥1,339 (53.56%)** | **≥590 (59.48%)** |
| B0 strict top-1 | 698 | 424 |
| Required B1 strict top-1 | **≥698** | **≥424** |
| B0 macro minCER | 0.1089795752 | 0.1140507370 |
| Required B1 macro minCER | **≤0.1035305965** | **≤0.1083482002** |

Implementation uses unrounded sums: `baseline.any + ceil(0.10 * misses)` and
`candidate.CERsum <= 0.95 * baseline.CERsum` with a tiny floating-point tolerance.
Word matches are pure rules-only default style, with no lexicon, reranking,
learned model, or overrides. Report paired wins and losses, not just net gain.
The 10%/5% choices were fixed before rule tuning and must not be weakened to
make a later candidate pass. If the bar is missed, report it and revisit scope
explicitly; do not quietly change the subset or substitute test results.

Additional mandatory gates remain: T-0056 construct/trace tests, corresponding
rule-disable negative controls, unchanged Hindi frozen outputs and accuracy,
and full CI. After a candidate is frozen on train/dev, report held-out test
aggregates once; test does not select rules, thresholds, or examples. Existing
B0 test scores have already been reported, so this is not a never-seen dataset.

## Frozen artifacts and enforcement

`benchmark/data/bengali/b0-dev.csv.gz` stores the sorted 2,500 B0 default outputs
from the parent implementation. Its SHA-256 is
`20f8a790aebc41fbf7f2af6a58bce3ba3f6770f97e910833f703c57c2d571a3f`.
The gate also pins the existing dev reference fixture hash, verifies every native
key, and recomputes baseline scores from those immutable outputs and references.
No new native split or training artifact is created. T-0057 still owns B2
cross-source exclusions.

Commands:

```sh
# Train/dev slot evidence, curation sizes, and synthetic alignment tests
go test ./benchmark -run 'TestBengali(AlignmentDefinitions|VowelEvidence)' -count=1 -v
# Definition tests and current status; normal CI does not claim B1 readiness
go test ./benchmark -run 'TestBengali(GateDefinitions|DevGate)' -count=1 -v
# Actual B1 acceptance: deliberately fails on B0
make test-bengali-b1-gate
```

T-0052 must add the enforcing target to CI when landing production B1; leaving
it informational would not satisfy B1 acceptance. Existing B0/prototype CI validates the fixture and definitions without claiming
readiness; the verbose status command explicitly reports B1 NOT READY.

The one-time snapshot capture path is `BENGALI_CAPTURE_DEV=/tmp/new-file.csv.gz`
with `TestBengaliDevGate` at this evaluation commit (production code remains at
parent 6a82b6a). It refuses to overwrite an existing destination. Never use later
B1 outputs to regenerate this baseline. The compressed fixture is deterministic.

## Verification

Focused definition/evidence/gate tests pass. `make test-bengali-b1-gate` fails
on unchanged B0 with both expected match-any and CER shortfalls, serving as the
integration negative control. Synthetic tests independently reject each missed
criterion and wrong subset rounding. A fresh snapshot capture is byte-identical.
A separate Python CSV/Levenshtein calculation reproduced both subsets' baseline
counts, macro minCER, and required thresholds without using the Go metric code.

Full `make ci` passed on 2026-10-04: formatting, lint, build, race/coverage,
all 3,432,832 frozen Hindi outputs, and accuracy benchmarks. Hindi pure strict
remains 1,147/1,330 (86.2%), match-any 1,236/1,330 (92.9%), and held-out
126/129 (97.7%). `make test-bengali` also passed. T-0055 is complete;
production B1 (T-0052) remains open and does not yet satisfy this contract.

## Review notes (2026-10-05)

Independent stack review; the preregistered thresholds and subset above are
unchanged.

- **Curation rationale, qualified.** Three votes also has the highest B0 dev
  match-any rate of the four candidate thresholds (54.9%, versus 48.4% / 51.1% /
  51.5% for ≥1 / ≥2 / ≥4). This does not ease the gate, whose bar is relative to
  each subset's own misses, but "not maximizing B0 scores" should be read with it.
- **Independent অ is not mixed on train.** Train favors `a` (1,257 vs 888, 59%);
  only dev is close (119 vs 110). Rendering it `o` is a product-style convention
  that the evidence does not support on train; it remains an open follow-up.
- **Gate scope.** The gate compares net match-any, net strict, and minCER sum on
  dev default style. Paired losses are logged but not bounded, and train and the
  alternate style are not gated. Later stages should report strict losses and
  consider a loss cap or train non-regression.
- **Evidence is now pinned.** The evidence harness uses the fixed B0 rule subset
  (`schwa.delete.word-final`, `schwa.keep.default`), and `TestBengaliVowelEvidence`
  pins a digest of every logged line per split, so the tables above stay
  reproducible after B1 rules land. Any non-empty `BENGALI_REQUIRE_B1` enforces.
