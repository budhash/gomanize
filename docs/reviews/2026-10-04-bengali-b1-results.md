# Bengali B1 rules and independent evaluation

Implements T-0052 above the evaluation contract in PR #115. Rules were developed
from the approved design and Dakshina **train** examples; two aggregate dev
checks selected the final candidate. The thresholds were not changed. No test
misses or external dataset examples were used to tune pronunciation.

## Behavior

Bengali keeps the final vowel after native consonant clusters and final হ, with
closed-coda exceptions for র্ট/র্ড/স্ট loans. It adds scoped initial/medial
ya/ba/ma phalas, ক্ষ/জ্ঞ, হ্ম/হ্ন, স before selected cluster members, `ফ → f`,
and the exact হও/হওয়া transformations. The দুঃখ family suppresses visarga and
geminates the following খ; unrelated visarga contexts retain the baseline.
Rules inspect source identities, mutate only their own unit, and use the
prototype gemination machinery with separate vowel decisions. Renderer support
for a rewritten `f` onset was added so a phala never resurrects the old `ph`.

Representative outputs: কর্ম → kormo; মধ্য → moddho; দুঃখ → dukkho;
ক্ষতি → khoti; কক্ষ → kokkho; জ্ঞান → gyan; চিহ্ন → chinno;
হও → hao; হওয়া → howa. Initial ya-phala uses the explicit `ya` convention,
including Open vowel metadata, rather than claiming one universal æ spelling.

Final-cluster retention is motivated by rule R4 in
[Choudhury, Basu & Sarkar (2004)](https://aclanthology.org/W04-0103.pdf).
The spelling/loan exceptions are project choices checked against train evidence,
not a claim that this phonotactic generalization covers every modern loan.
[Johny & Jansche (2018)](https://www.isca-archive.org/sltu_2018/johny18_sltu.pdf)
distinguish schwa realization/deletion and vowel quality; B1 does not add a
learned classifier or claim to resolve that full ambiguity.

Train evidence for f spelling: among 942 words containing ফ, switching the
initial candidate from ph to f gained 186 word matches and lost 27; 194 matched
both and 535 neither. English-origin spellings such as `ph` in names remain a
known tradeoff. The named `consonant.bengali.f-spelling` rule is disableable.
Chandrabindu keeps its existing n: a train-only drop comparison on the frozen
candidate's 312 relevant words gained 34 matches and lost 52 (128 both, 98
neither). No unsupported drop-default or new option was added. Anusvara remains ng.

## Preregistered gate results

| Set / metric | Frozen B0 | B1 | Required |
|---|---:|---:|---:|
| Full dev match-any | 1,210/2,500 (48.40%) | 1,370 (54.80%) | ≥1,339 |
| Full dev strict | 698 (27.92%) | 774 (30.96%) | ≥698 |
| Full dev macro minCER | 0.1089795752 | 0.0914997633 | ≤0.1035305965 |
| Curated dev match-any | 545/992 (54.94%) | 595 (59.98%) | ≥590 |
| Curated dev strict | 424 (42.74%) | 459 (46.27%) | ≥424 |
| Curated dev macro minCER | 0.1140507370 | 0.1004382211 | ≤0.1083482002 |

Paired full-dev wins/losses: **180/20**. Curated: **58/8**. The first candidate
missed the gate (full-dev 1,333 matches; curated 571). The train-supported f,
h-nasal and loan-coda changes produced the final candidate above. Rule tuning
stopped before held-out/external evaluation. These are pure rules-only results;
no lexicon, overrides, classifier, or reranking was used.

| Split | B0 strict | B1 strict | B0 match-any | B1 match-any | B0 minCER | B1 minCER |
|---|---:|---:|---:|---:|---:|---:|
| Train (25,000) | 28.18% | 32.06% | 48.32% | 55.61% | 0.10402 | 0.08472 |
| Dev (2,500) | 27.92% | 30.96% | 48.40% | 54.80% | 0.10898 | 0.09150 |
| Test (2,500) | 28.88% | 32.00% | 50.08% | 56.52% | 0.10548 | 0.08772 |

Alternate a-style test match-any is 51.04%, minCER 0.09921, reported without
using test to select the style. `make ci` now enforces `test-bengali-b1-gate`.
The original B0 dev output/reference hashes remain unchanged. Historical B0
mechanical tests and slot measurements select only their original three rules.
The snapshot capture helper now refuses changed output rather than allowing a
new candidate to be mislabeled B0.

## User-supplied repositories and external test

All four supplied repositories are retained as viable references, especially
bengali-romanizer's pronunciation/rendering separation and uncertainty plumbing.
See [the pinned review](../reference/bengali-repositories.md) for source revisions,
licenses, measured overlap, and the reverse-direction roadmap (F-0012).

The official BanglaTLit test split was imported **after** the B1 rule candidate
was fixed. It retains all 2,500 rows, 2,498 native sentence types, and its MIT
license; IDs and Roman references are unchanged. A deterministic importer pins
the source hash and records NFC/Cf normalization of native strings. The large
upstream `train` pool, which contains every official dev/test pair, was not
imported or used for training. No pretrained model was run.

External forward-transfer result: B0 exact sentence matches 5/2,500, macro CER
0.34312474; B1 8/2,500, macro CER 0.33149376. These scores include spelling,
spacing, punctuation, and informal sentence variation; they are not the
Dakshina word gate. An initial harness attempt used an unsupported leading
wildcard and selected identical catalogs; the corrected harness disables the
four explicit phase prefixes and asserts a known B0/B1 output difference before
scoring. No production rule changed in response to external evaluation.

## Limits and validation

This is an experimental rules-only baseline. Word-final homographs and retained
simple final vowels remain unresolved: e.g. চিহ্নিত currently yields chinnit,
not the desired chinnito. Larger clusters and excluded phala predecessors keep
compositional behavior; visarga expansion beyond the scoped family is not claimed.
These remaining pronunciation ambiguities belong to T-0053's measured learned
work, gated by T-0057's split/exclusion pipeline. CLI/npm/web language selection
is still Hindi-only. No reverse API is introduced; F-0012 tracks that next goal.

Focused tests cover production outputs, style behavior, negative contexts,
metadata-only traces, unique catalog priorities, and rule ownership. Disabling
each new rule breaks its corresponding construct as expected. Disabling all B1
phase prefixes through a temporary overlay makes the numerical gate fail at the
unchanged B0 counts. No baseline fixture was regenerated. BanglaTLit builder
hash rejection, transformation, row preservation, and deterministic output are
tested; a real fixture rebuild was byte-identical. Bengali benchmark tests passed
under race detection, and the WASM/npm smoke tests passed.

Full `make ci` passed on 2026-10-04, including the now-enforced B1 gate and all
3,432,832 frozen Hindi outputs. Hindi pure strict stays 1,147/1,330 (86.2%),
match-any 1,236/1,330 (92.9%), and held-out 126/129 (97.7%). After the external
fixture/test addition, format/lint, focused Bengali tests, fixture-builder tests,
and Bengali benchmark race tests also passed. T-0052 and T-0060 are complete;
B2 and the reverse feature remain open. Nothing in the stack has been merged.

## Review fixes and notes (2026-10-05)

Independent stack review. The gate table above records the candidate as
submitted; two correctness fixes followed, neither tuned on dev or test:

- **Canonical equivalence for nukta letters.** Phala and gemination checks
  required a one-rune unit, so a decomposed nukta letter (base + U+09BC) behaved
  as a cluster while its precomposed form did not (রয়্যালটি: `royjaloti` vs
  `royyaloti`; five Dakshina types diverged, all new in B1). The parser now marks
  base + nukta letters and `brahmic.IsSingleLetter` treats them as one letter.
- **Explicit final hasant.** The parser consumed a trailing U+09CD without a
  trace, so B1 keep rules re-added a vowel (আল্লাহ্ → `allaho`). The parser now
  records `TrailingHalant` and the Bengali keep rules respect it.

Effect (default-o, rules only): dev strict 774 → 775, match-any 1,370 → 1,371,
macro minCER 0.09150 → 0.09138; train strict 8,015 → 8,016, match-any 13,902 →
13,903; curated dev 595 → 596. Test is unchanged (strict 800, match-any 1,413).
BanglaTLit B1 macro CER 0.33149376 → 0.33148921 (exact unchanged at 8). Hindi
frozen outputs are unchanged. The preregistered gate still passes.

Further review notes (tracked, not changed here):

- **Train evidence for the post-miss rules.** Only the ফ→`f` evidence is recorded
  above. Reviewer ablations on the final catalog: ফ→`f` train match-any +186/−26
  (dev +30); loan-coda train +54/−2 (dev +6/−0); h-nasal train +16/−1 (dev 0).
  These account for the 1,333 → 1,370 step and every rule is net positive on
  train, so the disclosure holds; a committed train-only ablation would make it
  reproducible.
- **Paired losses.** Full-dev losses vs B0: 7 strict, 20 match-any (final-h on
  Arabic loans, s-cluster changes, medial ba-phala gemination, loan codas outside
  র্ট/র্ড/স্ট). Train: strict +1,105/−134, match-any +2,082/−260.
- **Gemination after a coda.** সংখ্যা renders `shongkkha` (B0 `shongkhja` was also
  wrong, so the gate does not register it).
- **BanglaTLit scoring is case-sensitive.** 1,344/2,500 references contain
  uppercase while the engine emits lowercase; exact and CER figures include that
  penalty.
- **T-0052 scope.** Anusvara is unchanged (`ng`) and visarga covers the দুঃখ
  family only; the task title overstates the delivered scope.
