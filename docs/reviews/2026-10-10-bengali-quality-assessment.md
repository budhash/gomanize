# Bengali quality assessment and improvement plan (2026-10-10)

**Status:** assessment plus an independent review. Written after the Bengali
milestone closed (stack #111–#126, retrain #127, [keel review](2026-10-07-bengali-keel.md)
#128, fixes #129–#130, doc sweep #131). Part 1 takes stock of the arc; part 2 is an
independent quantitative review (Fable, dev and train only) that measured the
shortcomings and corrected several of part 1's claims. No production code or
pinned number changes here.

**Bottom line.** Bengali's remaining errors are mostly convention- and
rule-governed, not lexical. One symbol choice (স as `sh` where Dakshina annotators
write `s`) is the largest single error class; it and seven further train-positive
rules lift the vowel-model profile on dev from **61.80% to 72.44% match-any**
(strict 35.80% → 43.60%). The structural gap after that is the vowel model's
exclusion of every word with a conjunct.

## Part 1: the arc

Design → B0 baseline → preregistered dev gate → B1 rules → B2 learned components
(vowel tree, lexicon, rejected character reranker, shipped cross-fitted native
selector) → runtime contract → CLI/npm/web frontends → B3 lyrics pilot → stack
review → variant-spelling retrain → keel review → K1–K8 fixes → doc sweep.

**What went well**

- **Method.** Gate thresholds were frozen before rules existed, training partitions
  are frozen, isolation guards fail closed (22/22 bad states denied), and negative
  results are recorded (character reranker rejected, chandrabindu drop 34 gained /
  52 lost, BanglaTLit CER regression under the selector). Test data never tuned
  anything.
- **Hindi was protected.** The frozen snapshot (264,064 inputs × 13 profiles) let a
  second language land without silently moving Hindi output; its one change (K1)
  was deliberate and every changed output is a correction.
- **Layered review.** Per-PR review caught local bugs (nukta letters in phalas,
  explicit final hasant, Python 3.12 portability); only the keel caught the
  cross-cutting ones (encoding-dependent Hindi output, quadratic long words, the
  `Debug` race, ~45 stale docs).

**What went less well**

- **Rigor without closure.** Many results were computed and recorded but not
  asserted; correct local normalizers were added four times in Go and four in
  Python instead of once; docs drifted; the variant-spelling leak passed the
  original isolation design. Much of the review work converted claims into tests.
- **Rigor aimed at the wrong layer.** The project measured learned components to
  three decimal places but never measured the symbol table against the data: the
  `sh` default for স was linguistically motivated and convention-wrong, and it cost
  more than every learned component gained (part 2).
- **Complexity ahead of payoff.** The native selector brings a second reranker
  protocol, a fresh engine per candidate, cross-fitting, records and pins, for
  **+7 words of 2,500** on test (62.76% → 63.04% match-any) and a slightly worse
  BanglaTLit CER.
- **Bespoke rather than general.** `"bengali"` is hard-coded in 8 files, there are
  two reranker protocols and no option registry (K9–K11). A third language would
  copy Bengali's pattern rather than plug into a framework.

## Current quality

Dakshina held-out test, 2,500 words, no overrides (recorded figures):

| Profile | Strict | Match-any | Mean minCER |
|---|---:|---:|---:|
| B1 rules | 32.00% | 56.52% | 0.0877 |
| Vowel model | 36.56% | 62.76% | 0.0764 |
| Vowel model + selector | 36.64% | 63.04% | 0.0757 |

- **About one word in three is wrong under every accepted reference.** In a lyric
  line of 5–8 words, most lines contain a visible error. Experimental is the right
  label today.
- **Low strict is partly the problem itself.** 979 of 2,500 dev words have
  references that differ only by o/a; 258 have a single reference.
- **External sets are much weaker:** Aksharantar 30.9–34.1% match-any; BanglaTLit
  macro CER ≈0.31 (inflated by case-sensitive scoring, T-0100). Part 1 blamed
  "Dakshina's conventions"; part 2 shows the biggest item is the reverse: `sh` for
  স is the engine's idiosyncrasy, and natural typists write `s` too.
- **Lyrics quality is unmeasured.** 52/82 exact lines are against unreviewed
  assistant-drafted references (T-0068), written in the engine's own `sh`/o style.
- **The selector is at its ceiling.** The best candidate in its set reaches only
  65.44% (dev) / 66.48% (test); the generator, not the chooser, is the limit.

## Part 2: independent review (Fable, 2026-10-10)

**Data discipline.** Dakshina `bn` train and dev only; test, Aksharantar,
BanglaTLit and the lyrics pilot were not run or inspected. Candidate rules were
chosen on train evidence and dev confirms them, but this is a fifth use of dev for
selection; dev figures are optimistic by a small unknown amount (every rule's
train and dev deltas agree in sign and relative size). Each candidate ran as an
extra Bengali rule through the real engine, so interactions with phalas, visarga,
hasant and the vowel model are real; all B1 construct examples (kormo, moddho,
dukkho, khoti, kokkho, gyan, chinno, hao, howa) are unchanged; Hindi is untouched.
The baseline reproduces the pinned dev figures exactly (B1 775 / 1,371 / 0.09138;
model 895 / 1,545 / 0.07839).

**Independent verification.** The headline lever was reproduced separately by
editing the স symbol in a scratch worktree and scoring with a separate scorer:
dev model 1,545 → 1,706 match-any, strict 895 → 1,031; train 15,262 → 16,678; dev
rules 1,371 → 1,503; train rules +1,312. Selector cost was re-timed (below).

### Error classes (dev, vowel-model profile, 955 misses)

Each miss is classified against its nearest reference ("pure" = every edit in one
class; the classifier is heuristic, the rule measurements below are exact).

| Class | Pure misses | In misses | Part 1 listed it? |
|---|---:|---:|---|
| স as `sh`, reference `s` (plus a few ছ) | **194 (20.3%)** | 262 | **no** |
| Medial inherent vowel kept/deleted | 176 (18.4%) | 276 | not as such |
| Mixed (two or more classes) | 210 (22.0%) | — | — |
| Final inherent vowel | 85 (8.9%) | 127 | yes |
| o/a vowel quality | 79 (8.3%) | 126 | yes |
| English-spelling loans (k→c, i→e, ksh→x…) | ~40 | ~60 | no |
| Gemination | 36 (3.8%) | 53 | yes, over-weighted |
| j/y/z for জ/য | 8 | 10 | no |
| ya-phala æ | 5 | 19 | no |
| Nasal ng/n/m | **2** | 2 | yes, over-weighted |
| ফ `f`/`ph` | **1** | 1 | yes, over-weighted |

Findings behind the classes:

- **স.** Attestation votes where the `s` and `sh` forms differ: train 3,830 vs
  1,064, dev 512 vs 121. The existing `s-cluster` rule already writes `s` before
  some cluster members; everywhere else B1 writes `sh`.
- **Model coverage.** 1,310/2,500 dev words (52.4%) are eligible for the vowel
  model; 946 of the 1,190 ineligible words contain a conjunct. Eligible words hit
  71.1%, ineligible 51.6%, and 576 of 955 misses are in ineligible words. There
  B1 has **no medial deletion rule at all**, so a stem-final vowel before a suffix
  is kept: অঞ্চলকে → `oncholoke` (references `oncholke`). The aligner discards
  every halant-bearing training type: 20,691/50,681 Google types (41%) and
  10,811/23,099 Dakshina train types (47%).
- **Final vowel** errors are the participle ending, not final vowels in general:
  of 430 train words where the engine picks the uniquely wrong side, 170 end in
  -িত (references keep 219 vs delete 21), plus ৃত 30/0, গত 41/0, ূত 15/5, ংশ 13/0.
  Final র ন ক ম ট দ are handled correctly thousands of times.
- **Verb morphology** is not a measured class. `korechho → korechh` in the lyrics
  pilot was the selector's error; dev references for করেছ are `korecho`/`karecho`
  (ছ as `ch`, final vowel kept). Final bare ছ occurs in 6 train words.
- **Loan codas** (train keep/delete; engine keeps): ন্ট 0/31, ন্ড 9/15, ঞ্জ 3/9,
  ন্স 0/9, ক্স 1/3, র্স 0/5, র্ক 2/9, ল্ড 0/9. The `loanCoda` list covers only
  র্ট/র্ড/স্ট.
- **Phalas.** Medial ya-phala after শ/ষ/স wants `y` (দৃশ্যের `drishsher` →
  `drishyer`); ba-phala after শ/স wants `w` (বিশ্বাস `bishwas`); medial দ্ব wants
  `db` (উদ্বোধন). ক্ষ as `kkh` is right (`ksh` loses 147 on train); য as `j` is
  right (`y` loses 305).
- **Initial অ** (left open by the gate review): writing `a` is net −17 train, −4
  dev on both profiles. Keep `o`.

### Measured fixes

Paired net match-any words against the same profile without the rule.

| Rule | Mechanism | Train B1 | Train model (wins/losses) | Dev model |
|---|---|---:|---:|---:|
| `s` | স → `s` everywhere; retire `s-cluster` | +1,312 | +1,416 (1,600/184) | **+161** |
| `chh` | ছ → `ch` | +139 | +187 (209/22) | +26 |
| `ito` | keep final vowel after -িত/-ীত/-ৃত/-ূত | +249 | +196 (207/11) | +8 |
| `tacons` | keep final ত after bare গ র হ ন শ ষ | +45 | +25 | +1 |
| `coda` | delete after final ন্ট ন্ড ঞ্জ ন্স ক্স র্স র্ক ল্ড | +75 | +75 (90/15) | +8 |
| `osuf` | word-final ও after a bare consonant: no extra `o` | +70 | +70 (78/8) | +5 |
| `yph` | medial ya-phala after শ/ষ/স → `y`, no gemination | +28 | +28 (35/7) | +10 |
| `bph` | ba-phala after শ/স → `w`; medial দ্ব → `db` | +8 | +8 (37/29) | +1 (marginal) |
| `suf` | delete stem-final vowel before a suffix table (কে টি টা গুলো গুলি তে দের রা খানা…) | +634 | +198 (235/37) | +21 |
| `ccv` | shared `schwa.delete.ccv` | +924 (2,276/1,352) | +43 (604/561) | +6 |
| `bcc` | shared `schwa.delete.before-cc` | −6 | −6 | −1 |
| initial অ → `a` | | −15 | −17 | −4 |

Combinations (dev strict / match-any / minCER; train match-any; `pos` = s, ito,
tacons, coda, osuf, yph, bph):

| Profile | Dev strict | Dev match-any | Dev minCER | Train match-any |
|---|---:|---:|---:|---:|
| Shipped vowel model | 895 (35.80%) | 1,545 (61.80%) | 0.07839 | 61.05% |
| + `s` | 1,031 (41.24%) | 1,706 (68.24%) | 0.06236 | 66.71% |
| + `pos` | 1,055 (42.20%) | 1,760 (70.40%) | 0.05781 | 68.81% |
| + `pos` + `chh` | 1,080 (43.20%) | 1,787 (71.48%) | 0.05611 | 69.58% |
| + `pos` + `chh` + `suf` | **1,090 (43.60%)** | **1,811 (72.44%)** | **0.05470** | 70.45% |
| B1 rules + `pos` + `chh` + `suf` | 1,008 (40.32%) | 1,682 (67.28%) | 0.06281 | 66.56% |
| B1 rules + `pos` + `ccv` | 1,043 (41.72%) | 1,714 (68.56%) | 0.05965 | 67.03% |

The 24 dev losses of `pos`+`chh` against the shipped model are almost all English
loans whose references keep `sh` (sector, second, thesis, bypass), one -িত word
(উপস্থিত) and ছাতকে; no native-word regression was found. Most of the vowel
model's gain over B1 is medial deletion that the shared CCV rule already provides
(`ccv` is a coin flip under the model, 604/561, because there it only touches
conjunct words, where deletion is not rule-governed).

### What remains

After `pos`+`chh`, 713 dev misses remain (28.5%, from 38.2%): medial inherent vowel
212 pure, mixed 151, o/a 78, final vowel 73, loans wanting `sh` 46, gemination 18,
loan spellings/æ/y-glides ~60. **461 of the 713 are in words the vowel model
cannot see**; hit rates are 80.8% on eligible words and 61.3% on ineligible ones.
Dev oracles on this profile: a perfect o/a chooser 75.48%; o/a + final vowel
79.76%; + gemination and vowel length 81.16%.

### Lexicon style (part 1 overstated it)

On train (in-sample, 8,976 entries) 4,957 entries (55.2%) equal the model's output.
Of the 4,019 that differ, 532 differ only by `sh`→`s` (834 contain it), 361 by a
medial vowel, 279 only by o→a (470 contain any o→a, **5.2%** of the lexicon), 185 by
a final vowel and 165 by gemination. For 281 of the 470 a-style entries the
o-style form is not attested on train at all, and for none does it reach 3 votes;
a mechanical restyle would fabricate unattested spellings. The lexicon is
Dakshina-style, not "often a-style", and its biggest clash with the rules is the
sibilant, which is the rules' error. The lyrics drop (52 → 47 lines) was measured
against draft references written in the engine's own `sh`/o style.

### Selector (cost corrected)

Warm-engine cost over the 2,500 dev words (Fable, re-timed independently):
rules 4.1–4.2 µs/word, model 10.7–11.0, model + lexicon 12.1, **model + selector
46.7–49.2** = 4.2–4.6× the model profile and ~11× B1, not the ~68× in keel K12
(that figure is likely a different measurement; T-0074 should state its method).
At ~20,000 words/s the cost argument is weak; the complexity argument stands.

On the improved base the candidate-set oracles are: model alone 71.48%, model +
B1 73.44%, model + whole-word a-style (`InherentVowelA`) 74.76%, all three 76.72%;
the a-style candidate alone wins 82 words. Single-slot flips are the wrong axis;
whole-word register has the headroom. Every selector artifact, pin and record is
invalidated by the rule changes (its candidates are model/B1 outputs).

## Corrections to part 1 (as first written)

- The error-class list omitted the largest class (স as `sh`, 20% of misses) and
  listed three that are negligible on dev: `f`/`ph` (1 miss), anusvara `ng` (2),
  visarga (≤3 dev words contain one). Coda gemination is 36 misses, not top-tier.
- "Verb morphology: korechho vs korechh" was a selector artefact.
- "Remaining failures are lexical, not rule-governed" is true for Hindi and false
  for Bengali: +266 dev words are recoverable with eight rules.
- "The lexicon carries Dakshina's spellings, often a-style": 5.2% are a-style.
- "~68× cost" for the selector: measured 4.2–4.6× the model profile.
- Item ordering: convention fixes and the conjunct-coverage gap come before more
  training data or a character sequence model.

## Improvement plan (revised, ranked by dev gain per effort)

Tracked as T-0117 (item 1), T-0118 (2–3), T-0119 (4), T-0120 (5), T-0121 (6),
T-0123 (7), T-0122 (8), T-0124 (9) and T-0125 (table-driven helper, item 11).

Validation for every rule item: a named, disableable rule; a committed train-only
ablation table (as for ফ→`f`); freeze; report dev, then test once; regenerate
`bnPinnedRules`/`bnPinnedModel` and the external and lyrics records deliberately;
the Hindi snapshot must not move.

1. **Decide house conventions with the lyrics review (T-0068).** The data says স →
   `s` and ছ → `ch`; lyric usage has both (Amar Shonar / Sonar Bangla). The
   worksheet must record each reviewer's preference for স (`s`/`sh`), ছ
   (`ch`/`chh`) and register (o/a), with several accepted spellings per word, or
   the lyrics number measures convention, not correctness.
2. **স → `s`** (+161 dev / +1,416 train; strict +5.4 pts; minCER −20%). Hours.
   Risk: lyric readers expecting `sh`; ~20 `sh`-spelled English loans. Also removes
   the largest lexicon/rules inconsistency.
3. **ছ → `ch`** (+26 dev / +187 train). Hours; decide with item 2.
4. **Rule bundle:** -িত/-ৃত/-ূত keep; final ত after গ র হ ন শ ষ; loan-coda
   delete table; C+ও; ya-phala after sibilants → `y`; ba-phala after শ/স → `w` and
   দ্ব → `db` (keep or drop `bph` on its own ablation). +54 dev / +525 train on top
   of item 2. One day.
5. **Suffix-boundary deletion** (the measurable morphology layer: nominal suffixes
   on conjunct-bearing stems, not verb endings). +24 dev on top of 2–4; +88 for
   rules-only. Half a day; the suffix table grows from train only.
6. **Extend the vowel model to conjunct-bearing words.** Align halant clusters in
   `tools/bengali/vowels.py`, mirror `vowelUnits`, regenerate the parity fixture,
   with one shared eligibility inventory. *Estimate*, not measurement: +120–160 dev
   words if coverage reaches ~90%. 2–4 days; needs the pinned Google lexicon.
7. **Shared CCV for the rules-only default** (not under the model): +142 dev /
   +924 train, with 1,352 train losses (a net win, not a clean one). The default
   profile is what CLI/npm/web users get.
8. **Retire the native selector** (keep the records); recommend the vowel model as
   the Bengali profile. Revisit reranking only with an attested lyrics target, a
   whole-word a-style candidate and a preregistered gate.
9. **Lexicon: no restyle.** Optionally a documented filter that skips entries
   differing from rule output only by o/a (470), measured on the attested lyrics.
10. **Then** more training data, and a design note for a compact character-level
    sequence model, only if items 2–6 leave a gap worth the constraint cost.
11. **Architecture before a third language:** K9–K11 (T-0071–T-0073), a
    table-driven rule helper (the loan-coda, suffix and -িত tables will grow), one
    shared model-eligibility inventory (today three copies: `vowel_model.go`,
    `vowels.py`, `reranker.go`), and style flags only after the option registry
    (`Options.DefaultStyle()` disables lexicon and selector for any new bool). Pin
    Hindi's headline numbers in tests (T-0116).

**Not worth pursuing (measured):** initial অ → `a` (−17 train), `before-cc` (−6),
ক্ষ → `ksh` (−147), য → `y` (−305), C+ই suffix (+4 train, 0 dev), more single-slot
selector flips.

**Recommendation.** Settle the স/ছ/register conventions through T-0068 first,
then land items 2–5 as one rules PR (about +10 points on dev), retire the selector,
and only then extend the vowel model to conjunct words. Bengali stays experimental
until the lyrics references are attested.

## Uncertainty and provenance

- Dev was consulted during this review; dev figures are mildly optimistic. Test and
  external sets were not touched and must be reported once, after freezing.
- Item 6's gain is coverage arithmetic, not a measurement.
- Whether lyric readers prefer `s`/`ch` is unmeasured; only Dakshina evidence exists.
- The review's scripts and per-word outputs were scratch analysis and are not
  committed. Each rule PR must recreate its evidence as a committed train-only
  ablation; the headline `s` result and the cost figures were reproduced
  independently as described above.
