# Bengali B0 baseline

B0 adds an experimental Go API language (`New("bengali")`) above the Part A
profile refactor. It is a mechanical baseline for subsequent measured work,
not a completed pronunciation engine. The CLI, npm wrapper, and web demo remain
Hindi-only. No learned Bengali components are included.

## Implemented scope

- Bengali consonants, independent vowels, matras, modifiers, digits, and both
  encodings of nukta letters. Ordinary conjuncts remain component-by-component;
  atomic/positional ক্ষ, জ্ঞ, হ্ম, হ্ন and phala behavior belong to B1.
- Decomposed ো/ৌ are registered for both segmentation and symbol lookup.
- Configured khanda ta (`ৎ`) has no inherent vowel even with rules disabled or
  before an independent vowel. It stays a consonant in its run; the following
  consonant has an after-halant-equivalent boundary. This is a generic parser
  hook with no change for existing Hindi configurations.
- Baseline rules keep medial inherent vowels and delete word-final vowels.
  Independent অ follows the same optional `InherentVowelA` style as default
  consonant vowels. No Hindi medial deletion, nasal rewrite, or learned model
  is imported. Existing Hindi-only options have no additional Bengali effect.
- Unknown characters, currency, historical fractions, and Assamese-only letters
  pass through. Explicit terminal virama remains the documented shared-parser
  limitation. Chandrabindu `n` and visarga `h` are provisional symbol spellings;
  positional rules and their goldens are B1 work.

Character identities/decompositions and dead-consonant semantics were checked
against the [Unicode Bengali chart](https://www.unicode.org/charts/PDF/U0980.pdf).
Latin spellings are this project's initial colloquial convention, not Unicode
pronunciation prescriptions.

## Dataset and metrics

Source: [Google Dakshina](https://github.com/google-research-datasets/dakshina),
CC BY-SA 4.0. `tools/build_bengali.py` reproducibly produces compressed train/dev/
test CSVs from the local v1.0 archive. See `benchmark/data/bengali/README.md` and
its manifest for hashes and normalization. All references/attestations are kept;
normalized native word sets are disjoint. No fixture is a learned artifact.

Strict top-1 uses the highest-vote reference (lexical tie break). Match-any and
macro minCER use all references, with one equally weighted observation per word.
The harness reports ref-count and best-attestation strata separately.

| Split | Words | o strict | o match-any | o macro minCER | a strict | a match-any | a macro minCER |
|---|---:|---:|---:|---:|---:|---:|---:|
| train | 25,000 | 7,044 (28.18%) | 12,080 (48.32%) | 0.10402 | 6,170 (24.68%) | 10,849 (43.40%) | 0.11414 |
| dev | 2,500 | 698 (27.92%) | 1,210 (48.40%) | 0.10898 | 638 (25.52%) | 1,107 (44.28%) | 0.11809 |
| test | 2,500 | 722 (28.88%) | 1,252 (50.08%) | 0.10548 | 647 (25.88%) | 1,132 (45.28%) | 0.11551 |

Test figures are aggregate reporting only; no test misses were mined or used to
adjust the baseline. Implementation and synthetic goldens preceded evaluation.

## Train/dev evidence for T-0055

The alternate engine changes only unmarked-vowel style. For words where the two
outputs differ, reference support is:

| Split | o only | a only | both | neither | identical candidates |
|---|---:|---:|---:|---:|---:|
| train | 1,883 | 652 | 3,340 | 8,955 | 10,170 |
| dev | 168 | 65 | 333 | 828 | 1,106 |

This supports keeping `o` as the **working B0 default**. It does not estimate the
full human inherent-vowel o/a distribution: other B0 errors exclude many words
from candidate matches, and 1,106 dev inputs do not distinguish the styles.
Per-position alignment and B1 remeasurement are needed before claiming the
register decision is settled. No per-word exceptions were added from this data.

| Best reference's attestation count | Train words | Dev words |
|---|---:|---:|
| 1 | 4,424 | 412 |
| 2 | 10,699 | 1,096 |
| 3 | 8,663 | 889 |
| 4+ | 1,214 | 103 |

A threshold of 4 would leave just 103 dev words. A threshold of 3 would leave
992; these are candidate curation sizes, not a selected benchmark definition.
T-0055 remains open for a final threshold/default decision with adequate
alignment evidence. The B1 success bar must be preregistered before rule tuning;
these B0 results alone are not a B1 acceptance floor.

## Verification

`make test-bengali` runs the fixture-builder unit tests and Bengali/benchmark
checks. Cases cover Unicode aliases, dependent-vowel segmentation, compositional
clusters, khanda-ta vowel suppression, public sentence boundaries, rule toggles,
and deterministic metric definitions. Hindi's frozen snapshot is reused without
regeneration; full CI verifies its 3,432,832 outputs alongside the Hindi suites.

Validation completed on 2026-10-04: `make ci` passed formatting, lint, build,
race/coverage tests, the unchanged 3,432,832-output Hindi snapshot, and accuracy
benchmarks. Hindi curated results remain 1,147/1,330 strict and 1,236/1,330
match-any; held-out constructs remain 126/129. `make test-bengali` and
`make npm-test` passed. Rebuilding the fixtures and manifest produced identical
bytes. A temporary Go overlay removing khanda-ta's structural flag made its
regression test fail as expected (`ৎই` rendered `toi` instead of `ti`).
