# Bengali Support — Design

**Date:** 2026-09-30
**Status:** Design v4; Part A config/quality refactor implemented on the feature
branch, 2026-10-04. Experimental Bengali B0 is implemented in its dependent
branch. The [B1 structural prototype](2026-10-04-bengali-b1-prototype.md) is
implemented separately. The [B1 rules catalog](2026-10-04-bengali-b1-results.md)
now passes the fixed dev gates; B2/B3 and A.2' remain unimplemented. B0 measurements and remaining
T-0055 decisions are in [the B0 report](2026-10-04-bengali-b0-baseline.md) and
the [evaluation contract](2026-10-04-bengali-evaluation-gate.md). T-0055 is now
resolved: operational o default, max-reference-votes ≥3 secondary curation,
and explicit full/curated dev gates before production B1 tuning.
Sections 1.5 and 1.6 record the earlier Fable/Codex reviews; §1.7 records the
latest Codex review and accepted resolutions. Part A is independent of Bengali
implementation; Bengali behavior must pass the B1 prototype gate. Tracked as F-0011.
**Scope:** Add Bengali (Bangla) as the second language, in two dependency-ordered
pieces: (A) make the shared Brahmic layer script-general with zero
change to Hindi output, then (B) build the `lang/bengali` package on top of it.

This doc is grounded in two research passes (2026-09-30): a linguistics review of
Bengali vs. Hindi for the engine, and a dataset-availability survey. Primary
sources are cited at the end; the standing project discipline (train-split-only
training, overrides-are-not-accuracy, measure-before-shipping, negative results
recorded) carries over unchanged.

---

## 1. Summary and the one hard finding

Bengali has permissively-licensed analogues for every Hindi data source
(Dakshina `bn`, Aksharantar `ben`, BanglaTLit, IndicCorp v2), plus a large
public-domain lyrics source with no Hindi equivalent (Tagore's Gitabitan, 2,232
songs). The engine is partly ready: the parser, category system,
run identification, rule-engine mechanics, scheme layer, and the three
learned-component loaders are all reusable.

The hard part is one thing, and it is not mechanical: **the inherent vowel.**

- Hindi's inherent vowel is `a`, kept or deleted — a binary decision the engine
  models with `SchwaState{Keep,Delete}` and a renderer that writes the literal
  `"a"`.
- Bengali's inherent vowel অ is /ɔ/, **raised to [o] in systematic contexts and
  deleted in others** — a three-way outcome (∅ / ɔ / o) on a single grapheme.
  A kept Bengali inherent vowel is usually spelled `o`, not `a`.
- Bengali inherent-vowel deletion is **less regular than Hindi's**. The same
  published rule family scores ~96% on Hindi but ~85% on Bengali (Choudhury et
  al. 2004); the best neural classifier caps at ~97.5% for deletion and ~92.5%
  for the ɔ-vs-o choice (Johny & Jansche 2018). The irregularity is concentrated
  exactly where Hindi is most regular — word-final — and includes a layer of
  word-level homographs rules cannot resolve (বল *bôlo* 'speak!' vs *bol* 'ball').

Consequence for the architecture: the shared renderer's assumption that a kept
inherent vowel has one fixed spelling is false for Bengali, and the shared
word-final deletion rule encodes Hindi regularity Bengali lacks. Part A fixes the
first (a config value + a per-unit vowel-quality slot); Part B supplies the second (a
Bengali rule set, and eventually a Bengali-trained classifier).

A second consequence is for measurement: Dakshina `bn` has **3.8 attested
romanizations per word vs. Hindi's 1.8**, and the colloquial target is bimodal
(`o` vs `a` for the same vowel depending on register). Strict top-1 will be
lower than Hindi's regardless of engine quality. **Match-any / minCER is the
Bengali headline metric**, and the pure-accuracy gate must be set empirically
from the `bn` data, not copied from Hindi's 85%.

---

## 1.5 Review outcome and resolutions (v2)

Two independent adversarial reviews (Fable subagent + Codex, both read-only)
converged on the same verdict: **the core representation is viable, but the
plumbing was underspecified and several Bengali edge cases were missing — not
implementation-ready as v1 was written.** Both confirmed the sound parts: the
five identified literals are the real remaining ones; binary keep/delete plus a
separate vowel-quality field can model the three-way outcome; reusing
`colloquial.SelectRules` is right; shipping Part A alone is the right discipline.
These are historical review outcomes; v4 refinements in §1.7 and the operative
Part A/B contracts below supersede earlier wording where noted.

**Convergent must-fixes (both reviewers):**

1. **Config reaches only the parser today.** The renderer, `PrepareWord`, and
   `brahmic.SchwaRules()` receive no `Config`. **Resolution:** the parser stamps
   the resolved profile onto the word (`WordBrahmicData`); renderer and rules read
   it there. Fix `IdentifyRuns`, which currently replaces `WordBrahmicData`
   wholesale (`runs.go:53`) — it must merge, not clobber. No `core.Script`
   signature change. (A.2)
2. **Zero-value `Config` would silently break Hindi for public callers** (exported
   struct, documented as `Config{Halant, Nukta}`; new fields default empty →
   renderer writes nothing). **Resolution:** `Config.normalize()` fills Devanagari
   defaults (`InherentVowel="a"`, bare-vowel runes, aa-matra, indep range,
   sonorants); add a "zero-value Config still romanizes Hindi" test. (A.2)
3. **Store vowel *quality*, not a literal spelling.** A `SchwaRom string` conflates
   phonological class with output style. **Resolution:** a semantic
   `SchwaQuality` (`Default`/`Raised`) on the unit, mapped to a spelling at render
   time by the scheme/flag. (A.3)
4. **Neighbor mutation is real but hazardous** — the `acted` map never marks the
   mutated neighbor, `BaseRom`-matching rules mis-fire after a string rewrite, and
   it is invisible to `--debug`. ("No Hindi precedent" was wrong: `aanv-to-aon`
   already mutates `u.Next`.) **Resolution (J2, approved):** gemination is a
   render-time field (`BrahmicData.Geminate` + a per-consonant geminate onset in
   the symbol table), applied by the renderer — never a `BaseRom` string rewrite.
   V4 replaces the boolean with explicit modes and also changes trace recording;
   the field alone does not resolve debug visibility. See B.3.1 and B.3.3.
5. **Rule-order ties unchecked + unstable sort** (`NewRuleEngine` lacks the
   tie-check `AddRule` has; `sort.Slice` is unstable). Hindi is safe only because
   all its effective priorities are distinct. **Resolution:** Bengali's catalog
   adds a priority-uniqueness test; Part A must not reorder Hindi's priorities. (A.2)
6. **The frozen snapshot is too weak** (samples other suites; aggregate counts
   hide different-wrong-words; default options only; lexicon short-circuits the
   pipeline). **Resolution:** per-input frozen output for every suite **and** every
   meaningful option profile (SchwaModel, Lexicon, Rerank, KeepMedialSchwa,
   LongVowels, SimpleNasals, disable-rule configs), plus the invariants constructs
   (गई, कऋ, दरअसल, ऄ). (A.4)
7. **Khanda ta ৎ is not expressible** via `SymbolInfo{Category, BaseRom}`.
   **Resolution:** add a `BrahmicData.NoInherentVowel` flag (parser-set for ৎ),
   honored by the renderer and run logic. (B.1)
8. **ো/ৌ need to be `SymbolMap` keys**, not only `MultiChar` (segmentation vs
   lookup are separate). Register the decomposed two-part forms as `CatMatra`
   keys *and* MultiChar. (B.1)

**Sharpest single-reviewer catches (adopted):**

- **The ɔ→o raising rule cannot live in `PhaseSchwa`** — the acted-map would
  swallow the keep/delete decision. It must be a *post-Schwa* (Vowel/Render) rule
  conditioned on `SchwaKeep`; the harmony lookahead is only valid after schwa
  decisions. **And** under the default `o` style, raising is a no-op (ɔ→o, o→o),
  so colloquial B1 does not need it — it defers to the academic/phonemic scheme
  and the B2 classifier. (B.3, milestones)
- **`InherentVowel="o"` + the unconditional pre-independent-vowel write → "oo"**:
  হও→*hoo*, হওয়া→*hooya*. Bengali needs a `C+ও` rule on day one of B1. (B.3)
- **`"Delete"→delete` is not authoritative** before an independent vowel — the
  renderer there ignores `SchwaState`. The Bengali classifier/rules must exclude
  those contexts. (A.3)
- **Rune-index fragility (J1, approved).** Only `cccc-final` (`u.Start.Rune==1`)
  is rune-fragile (`isWordInitialConjunct` uses a unit index). Fix is one
  predicate → logical second-consonant. **This is behavior-*changing*** for
  nukta-initial Hindi words (fixes a latent bug), so it is split out as **A.2'
  (measured), separate from the byte-identical config lift (A.1).**
- **Independent অ (অতি→oti) needs the ɔ/o/style treatment too** — `SchwaQuality`
  only covers consonant-borne vowels; a vowel-phase `BaseRom` rule (and the
  classifier) must cover independent অ. (B.3)
- **Metrics:** pair match-any with macro minCER, strict top-1, attestation strata,
  and the Aksharantar number (full and Dakshina-test-unseen slices: Aksharantar
  contains the Dakshina test vocabulary, so it is not independent); match-any is *inflated* by `bn`'s 3.8
  refs/word vs Hindi's 1.8, so "trailing Hindi" could mislead — report the
  reference-count-controlled number and say so. (§3)
- **B1 gate "beats B0" is near-vacuous** — pre-register a bar from the B0
  histograms. (milestones)

**Deferred / open-flagged (recorded, decided in B-phases against `bn` data):**
lexicon short-circuit ignores the style flag (style-key the lexicon or disable it
under non-default style); rerank candidates are Hindi-shaped (base + SchwaModel) —
Bengali's axis is o-vs-a; the schwa-model loader is Hindi-package-private with
Devanagari tables (lifting it to shared is extra regression surface — likely copy
into `lang/bengali` for B2); options must be bools (`Rule.Conditional` is a
bool-option name); terminal explicit virama is a pre-existing Brahmic limitation
to note.

## 1.6 Second review pass (v3) — Part A cleared, B1 pins recorded

A focused second pass (Fable + Codex, read-only) checked whether the v2
resolutions held. Both confirmed **A.1 (the config lift) is ready**, and found a
short list of refinements. The two that gated Part A are **applied in this v3**:

- **A.2' predicate was mis-worded.** `u.Start.Rune == 1` means "second *unit*",
  which in a vowel-initial word is the *first* consonant (अजगर→*ajgar*); "second
  *consonant*" would wrongly flip every vowel-initial word (→*ajagar*). The
  faithful predicate is **`w.Units[1] == u`**, and the affected class is **any
  multi-rune first unit** (nukta- *and* ज्ञ/क्ष-initial), not just nukta. Also:
  A.2' is **not assumed neutral** — `cccc-final` already mis-deletes some words
  (जबरदस्त→*jabradast*), so the fix can propagate that to nukta-initial forms; it
  must be *measured* against data, which is exactly why it is gated separately.
- **The `SchwaQuality`→spelling map cannot read the `Scheme`.** The renderer
  receives only the word; a `Scheme` is a rule selector (`colloquial.go:23`). The
  style switch is therefore a **`core.Options` bool** (bool because
  `Rule.Conditional` only names bool options), and the readers take a **nil-profile
  Devanagari fallback** (a future/mocked `Word` not built by the brahmic parser).

The v3 pins covered gemination ownership, an `Open` vowel quality for হও,
khanda-ta parser support, and priority validation. **Superseded in v4:** one
"emit Prev's onset" operation cannot handle both phalas and post-visarga
consonants; the debug extractor alone cannot record metadata-only rule traces;
and the হও rule cannot also generate হওয়া. The operative contracts are B.3.1–3.
Khanda ta still requires `Config.VowellessConsonants []rune` (or a dead-consonant
category), with the following consonant treated as after-halant-equivalent for
run lookaheads. Priority-tie validation belongs in `NewRuleEngine`, with a Hindi
catalog tie test in Part A; existing Hindi priorities must remain unchanged.

Also fixed in v3: two internal contradictions — §5 no longer lists ɔ→o raising
among B1 wins (it is deferred; no-op under `o`), and B.6's chandrabindu is
data-gated (J3), not "default drop".

**Verdict (both, second pass): Part A is implementation-ready; the B1 pins above
are resolved before B1 starts, not before A.**

## 1.7 Third review and accepted resolutions (v4, 2026-10-04)

The latest review checked remote commit `1524d43` against the current engine.
Five findings are resolved as design contracts below, not claimed implemented:

1. **Gemination ownership:** separate phala substitution from self-gemination
   after visarga; specify aspirated onset pairs and vowel ownership (B.3.1).
2. **Independent ও:** separate the exact হও and হওয়া transformations, including
   the glide and suppressed য় onset; broader contexts remain data-gated (B.3.2).
3. **Pronunciation versus spelling:** the three-class classifier cannot select
   default-style `a` exceptions when both kept classes render as `o`. The lexicon
   owns those exceptions initially; a learned spelling selector is deferred (B.4).
4. **Debug traces:** record metadata-only rule changes as well as exposing final
   unit metadata; a focused regression test must prove the rule is visible (B.3.3).
5. **Training isolation:** freeze normalized word-type exclusions before external
   training, assert disjointness, and report other corpus overlap (B.4.1).

**Delivery order:** revise this design and task dependencies first; implement
Part A with frozen Hindi outputs independently; then prove phala, visarga,
khanda-ta, and independent-vowel examples in a small B0/B1 prototype before
expanding the rule catalog. Keep A.2' separately measurable. No implementation
or benchmark improvement is asserted by this documentation revision.

**Task mapping:** T-0056 gates B1 catalog expansion with the prototype and trace
checks; T-0057 gates B2 training with split manifests and exclusions. Existing
T-0050 title shorthand `SchwaRom` means the semantic `SchwaQuality` contract in
A.3; T-0052's older raising label is deferred to B2 as specified in B.3.

**Acceptance:** every proposed field and rule must produce its documented
examples without an unspecified extra transformation. A Bengali example is an
acceptance target until the prototype passes, not evidence that a rule works.

---

## Part A — Generalize the Brahmic layer (zero Hindi regression)

### A.1 The Devanagari-specific sites in the "shared" layer

Verified against current `main` (v1.2.0 prep):

| Location | Hardcoded (Devanagari) | Role |
|---|---|---|
| `script/brahmic/renderer.go` (kept-schwa writes) | literal `"a"` | spelling of a kept inherent vowel |
| `script/brahmic/renderer.go` `isInherentAVowel` | only `अ` U+0905 / `ऄ` U+0904 (Bengali `অ` U+0985 not recognized) | the bare-a vowel that coalesces with a preceding schwa |
| `script/brahmic/schwa_rules.go` `isSonorousRune` | र/य/व | sonorant-final schwa retention |
| `script/brahmic/schwa_rules.go` `before-cc` | aa-matra `ा` U+093E | Sanskrit-heuristic deletion guard |
| `script/brahmic/schwa_rules.go` `isWordInitialConjunct` | indep-vowel range U+0905–U+0914 | word-initial conjunct detection |

`Config` already carries `Halant`, `Nukta`, `MultiChar` per language — the parser
is clean. These five are the remaining leaks.

### A.2 `ScriptProfile`: lift the literals into config

Implemented as `brahmic.Config.Profile *ScriptProfile`; the existing Halant,
Nukta, and MultiChar fields stay on Config. The resolved profile contains:

```go
type ScriptProfile struct {
    InherentVowel         string   // Hindi "a"; future Bengali "o"
    RaisedVowel           string   // "o" by default
    BareVowelRunes        []rune   // Hindi {0x0905, 0x0904}
    AaMatra               rune     // Hindi 0x093E
    SonorousRunes         []rune   // Hindi {र,य,व}
    IndependentVowelRange [2]rune  // Hindi {0x0905,0x0914}
}
```

**Plumbing (the review's #1 gap).** Only the parser receives `Config` today. The
parser resolves the profile and stamps it onto the word via `WordBrahmicData`;
the renderer and `brahmic.SchwaRules()` read it from the word. `IdentifyRuns`
merges runs into existing `WordBrahmicData` rather than replacing the profile. No `core.Script` / `NewRenderer` /
`PrepareWord` signature changes. **Defaulting:** `Config.normalize()` fills the
Devanagari values for zero scalar fields and nil slices; explicit empty rune
slices disable membership. Nil Profile and legacy `Config{Halant, Nukta}` callers
retain Hindi behavior. The parser copies caller-owned profile slices; parsed
profiles are read-only. Dedicated tests cover partial defaults and nil-profile
fallbacks, including manually built words.

The five sites then read from config instead of literals. **Hindi's config
reproduces the current values exactly**, so this lift is behavior-preserving by
construction.

**A.1 vs A.2' — two risk classes, kept separate.** The config lift above (A.1) is
byte-identical *by construction*. The `cccc-final` rune-index fix (J1) is a
*separate* step (A.2'): replacing `u.Start.Rune == 1` with the **second-unit**
predicate `w.Units[1] == u` (NOT "second consonant" — that would wrongly flip
vowel-initial words like अजगर). This is behavior-*changing* for any **multi-rune
first unit** (nukta- and ज्ञ/क्ष-initial), so it ships as its own commit with its
own measurement (tracked separately as T-0058). It is **not** assumed neutral: `cccc-final` already mis-deletes
some words (जबरदस्त→*jabradast*), so this can propagate that deletion to
nukta-initial forms — evaluate the moved words against attested data, update the
golden snapshot *deliberately*, and justify. It does **not** shelter under the
byte-identical guarantee (which covers A.1 only).

### A.3 Separate "keep vs delete" from "what a kept schwa spells"

This is the structural change the inherent vowel forces. Keep `SchwaState` binary
(`Keep` / `Delete`) — that decision is shared and correct for both languages.
Store a per-unit *semantic quality*, not a literal string (the review's #3 —
a literal `SchwaRom` conflates phonological class with output style, and a global
`"o"` rewrite would also rewrite genuinely raised /o/):

```go
// in BrahmicData
SchwaQuality SchwaQuality // SchwaDefault | SchwaRaised | SchwaOpen.
                          // Default => Profile.InherentVowel (or "a" with
                          // Options.InherentVowelA); Raised => Profile.RaisedVowel;
                          // Open => "a" (the scoped হও case).
```

Renderer: where it writes `"a"` for a kept schwa, map the quality using the
resolved profile and `Word.Options`; it does not read the `Scheme`.

- **Hindi**: no rule sets `SchwaQuality` (stays Default); `InherentVowel="a"` →
  every site emits `"a"` exactly as today. Byte-identical.
- **Bengali**: `InherentVowel="o"` gives the common case for free; a post-Schwa
  raising rule (or the classifier) sets `Raised` where harmony/ra-phala/final
  applies. The quality→spelling map is read from a **`core.Options` bool** (the
  renderer sees `Word.Options`, not the `Scheme`): a future Bengali profile
  maps Default and Raised both to `o`; `InherentVowelA` maps Default→`a` and
  leaves Raised→`o`. The `ô` convention and full academic scheme are deferred.
  Readers fall back to Devanagari defaults when the profile is nil.

This maps the three-way outcome onto existing machinery: `Delete` → delete;
ɔ/o → `Keep` + `SchwaQuality`. **Caveat (review):** before an *independent* vowel
the renderer ignores `SchwaState` entirely (it emits unconditionally), so the
classifier/rules must exclude those contexts; and independent অ itself
(অতি→oti) is handled by a vowel-phase `BaseRom` rule, not `SchwaQuality` (which
only covers consonant-borne inherent vowels).

### A.4 Zero-regression guarantee — three independent nets

1. **Byte-identical benchmark gate (A.1 only).** For the config lift, the full
   Hindi suite (`make test-dakshina` + the five benchmark suites) must not move a
   single count: 86.2% pure, 92.9% match-any, 94.8% rerank, held-out, lyrics,
   COMI. A.2' is exempt (it changes a measured set of nukta/multi-rune-initial
   words); its snapshot delta is reviewed and justified, not required to be zero.
2. **Frozen Hindi golden snapshot (strengthened per review #6).** Before the
   refactor, commit **per-input** output — not aggregate counts, which can hide
   different-wrong-words — across every suite **and** every meaningful option
   profile (default, SchwaModel, Lexicon, Rerank, KeepMedialSchwa, LongVowels,
   SimpleNasals, representative disable-rule configs), plus the invariants
   constructs (गई, कऋ, दरअसल, ऄ) and the `SchwaPending` render path (the new code
   must keep `!= SchwaDelete`, not switch to `== SchwaKeep`). A test diffs live
   output to it and must be exactly zero. The A.2' index fix updates this snapshot
   deliberately (see A.2), the config lift must not touch it at all.
3. **Ship Part A alone, first.** Part A merges as its own PR, green, *before any
   Bengali code exists*. If a count moves, there is no Bengali to blame — it is a
   pure refactor regression and gets fixed in that PR.

The initial frozen fixture is committed before the refactor in `6f59b22`:
264,064 distinct inputs from all ten Hindi CSVs plus invariant constructs,
13 option/rule profiles, and 3,432,832 exact outputs. See
`testdata/hindi_snapshot/README.md` for provenance and deliberate regeneration.
The fixture is a compatibility baseline, not an accuracy benchmark.

Part A validation (2026-10-04): all 3,432,832 outputs matched without regenerating
the fixture. A deliberate default-vowel mutation made the snapshot fail and was
restored. `make ci` passed formatting, lint, build, race/coverage (62.3% total),
and accuracy suites: curated pure 1147/1330 (86.2%), match-any 1236/1330 (92.9%),
and held-out constructs 126/129 default (97.7%). The full instrumented snapshot
replay took 515 seconds; the coverage target now permits 20 minutes per package
for slower workers. The rune-index correction is still pending as T-0058.

### A.5 Scope boundary — what stays Hindi-only

Everything Hindi-specific already lives in `lang/hindi` and does **not** move:
the Hindi symbol table, the फ→f rule, the व→w rules, the Hindi anusvara
(homorganic n/m) render rules, and `schwa_tree.json` / `lexicon.tsv` /
`roman_ngrams.tsv` (Hindi-trained data). Part A touches `script/brahmic` and
the Hindi `Config` construction in `lang/hindi/symbols.go`, plus the necessary
`core.Options` style field, constructor priority validation, and regression tests.
B1 owns gemination modes and the metadata-only trace change (B.3.3).

---

## Part B — The `lang/bengali` package

Built entirely on the generalized layer from Part A. Mirrors `lang/hindi`'s file
layout: `symbols.go`, `rules.go`, `*_test.go`, and (in later phases) the three
learned-component data files.

### B.1 Symbol table (Bengali block U+0980–U+09FF)

Full inventory is in the linguistics research (consonant, vowel/matra, conjunct,
modifier tables). Notable mappings that differ from Hindi:

*As built in B0 (see [B0 baseline](2026-10-04-bengali-b0-baseline.md)):* conjuncts
stay compositional (no `ক্ষ`/`জ্ঞ`/`হ্ম`/`হ্ন` MultiChar) and স always maps to `sh`;
atomic/positional conjunct and phala behavior is B1 rule work. Only the split
matras are MultiChar.

- **Consonants:** ঙ→`ng`, ঞ→`n`, ফ→`ph` (not Hindi's `f`), ব→`b` (covers b and
  Sanskrit-v), য→`j`, য়→`y`, ড়→`r`, ঢ়→`rh`, স→`sh` by default / `s` in
  clusters, ৎ (khanda ta)→`t` (vowel-less), ণ merged to `n`, ষ→`sh`.
- **Vowels/matras:** linear logical storage confirmed (no reordering needed for a
  left-to-right rune walk, even for pre-base ি/ে and split ো/ৌ).
- **Aliases (reuse Hindi's `init()` pattern):** ড়/ঢ়/য় are composition-excluded,
  so NFC keeps them decomposed while keyboards emit precomposed — register both.
- **MultiChar:** `ক্ষ`, `জ্ঞ`, `হ্ম`, `হ্ন`, and the **decomposed two-part
  matras** `ে`+`া` (→ো) and `ে`+`ৗ` (→ৌ). Note (review #8): MultiChar only controls
  *segmentation*; the parser then does an exact `SymbolMap` lookup of the matched
  string, so ো/ৌ (and precomposed forms) must **also** be `CatMatra` keys in the
  symbol map, or they become empty/unknown units.
- **Khanda ta ৎ (review #7):** not expressible as `SymbolInfo{Category, BaseRom}`
  — as a consonant it would acquire an inherent vowel (হঠাৎ→*hothato* once final
  deletion weakens); as a symbol it breaks run lookaheads. Add a parser-set
  `BrahmicData.NoInherentVowel` flag honored by the renderer and run logic.

### B.2 Bengali `Config`

```go
brahmic.Config{
    Halant: "্",
    Nukta:  "়",
    Profile: &brahmic.ScriptProfile{
        InherentVowel:         "o",
        RaisedVowel:           "o",
        BareVowelRunes:        []rune{0x0985}, // অ
        AaMatra:               0x09BE,        // া
        SonorousRunes:         []rune{'র', 'য', 'ব'}, // tune against bn data
        IndependentVowelRange: [2]rune{0x0985, 0x0994},
    },
}
```

### B.3 Bengali rules

Bengali composes *some* of `brahmic.SchwaRules()` but overrides the word-final
behavior, which is where it diverges most. New/changed rules, all expressible in
the current rule engine:

- **Final-cluster keep (Choudhury R4):** Bengali disallows complex codas, so a
  schwa after a final consonant cluster is kept, spelled `o` (অন্ত *onto*, শব্দ
  *shobdo*, কর্ম *kormo*) — opposite of Hindi (*ant, shabd, karm*).
- **Final-হ keep:** গ্রহ *groho*.
- **Weaker word-final deletion generally** (Bengali lacks Hindi's strong final
  prohibition); plus a documented homograph error class (বল, হল, কোন, মত).
- **ɔ→o raising** — a *post-Schwa* (Vowel/Render-phase) rule conditioned on
  `SchwaKeep` (not Schwa phase; see §1.5), setting `SchwaQuality=Raised` on
  harmony/ra-phala-initial/retained-final consonants. **Deferred out of B1:**
  under the colloquial `o` default this is a no-op (both map to `o`); it only
  matters for the academic/phonemic scheme and the B2 classifier. Independent
  অ (অতি→oti) is handled by a separate vowel-phase `BaseRom` rule.
- **Independent ও (B1, day one):** use the separate, explicitly scoped
  transformations in B.3.2. Do not compose a blanket `C+ও → Open` rule with a
  blanket medial `ও→w` rule: they do not yield the documented হওয়া output.
- **Sanskritic medial retention:** Bengali keeps many medial schwas Hindi deletes
  (রচনা *rochona* not *rochna*); the `ccv` deletion needs a Bengali-tuned guard.
- **Positional phala rules** (ya/ba/ma). The mechanism is a **render-time
  gemination mode**, not neighbor `BaseRom` mutation (§1.5 #4, J2): the phala
  carries `RepeatPrevious`, with onset-pair rendering defined in B.3.1
  (`kh→kkh`, never `khkh`). ya-phala ্য (initial → æ, via `SchwaQuality`/`BaseRom`
  on the phala-bearing unit — *not* a following vowel, which may not exist;
  medial → geminate + silent য), ba-phala ্ব and ma-phala ্ম (initial silent;
  medial geminate), each with an exception set. "Initial" needs a cluster-position
  predicate, not `IsWordInitial()` (the phala unit always has a preceding
  consonant).
- **Positional conjuncts:** ক্ষ (initial `kh` / medial `kkh`), জ্ঞ (initial
  `g`/`gæ` / medial `gg`).
- **Modifiers:** anusvara ং → `ng` always (retire Hindi's homorganic rules for
  Bengali); visarga ঃ medial → the following consonant carries `GeminateSelf`
  and the visarga suppresses its own output (দুঃখ *dukkho*, B.3.1);
  chandrabindu ঁ → **data-gated**
  (J3, see T-0055 — likely positional, `n` medially; the earlier "drop" default
  was an unmeasured guess and is retracted).

### B.3.1 Gemination modes and vowel ownership (B1 contract)

Use `GeminationMode{None, RepeatPrevious, GeminateSelf}` on `BrahmicData`.
`None` is zero and preserves existing rendering. Rules set only their own unit's
mode; source identity and `BaseRom` remain available to later rules.

- **RepeatPrevious:** owned by the phala. Replace its consonant onset with the
  right member of the preceding consonant's geminate pair; emit the left member
  at that preceding consonant. The renderer may inspect adjacent modes, but no
  rule mutates the preceding unit. The pair is derived from the preceding unit's
  final `BaseRom` after consonant rules: `t → (t,t)`, `kh → (k,kh)`. Thus an
  aspirated onset is emitted as `kkh`, not `khkh`. Validate adjacency/cluster
  eligibility; MultiChar and already-clustered exceptions must not be doubled
  automatically. The preceding consonant has no vowel before the halant-linked
  phala; the phala retains its own vowel slot or following matra.
- **GeminateSelf:** owned by the consonant immediately following visarga. Emit
  both members of its own final onset pair, then its own vowel slot or matra.
  Never derive this onset from `Prev`, which is the modifier. The visarga unit's
  separate rule suppresses only its own output.

Prototype acceptance examples: মধ্য parses as `ম | ধ | য(after-halant)` and
emits `mo | d | dh+o → moddho` when the phala selects the `(d,dh)` pair;
দুঃখ parses as `দ | ু | ঃ | খ` and emits `d | u | empty | kkh+o → dukkho`.
The phala/post-visarga vowel keep decisions are explicit Schwa-phase rules;
gemination must not silently force retention. Test aspirated/unaspirated pairs,
following matras, changed final onsets, and cluster exceptions independently.

### B.3.2 Independent ও: exact prototype transformations

Start with exact source-word contexts for these acceptance examples. Broader
rules require attested positive and negative cases; the examples alone do not
justify a general `C+ও` rewrite. Match source identity, including both য় aliases,
not a `BaseRom` changed by an earlier phase. Each rule acts on its own unit.

| Input / parsed units | Explicit actions | Rendered pieces (default style) |
|---|---|---|
| হও: `হ \| ও` | On হ, set `Open`; leave independent ও as `o` | `h+a \| o → hao` |
| হওয়া: `হ \| ও \| য় \| া` | Leave হ at Default; on ও set `BaseRom=w`; on য় suppress its own onset; leave া as `a` | `h+o \| w \| empty \| a → howa` |

The হও-specific `Open` rule must not match হওয়া. য় stays a consonant with its
following matra, so it acquires no extra inherent vowel when its onset is empty.
Changing ও's romanization must not change its independent-vowel unit type; the
preceding consonant still takes the pre-independent-vowel renderer path.
These are scoped linguistic rules, not lexicon hits; test with Lexicon/Rerank
both off. Require equivalent decomposed/precomposed য় outputs and negative
cases for standalone ও, unrelated medial ও, and য়+া outside হওয়া. Assert rule
traces for every affected unit. Test the style flag separately: Default changes
with style, while Open stays `a`; do not promise `howa` under every style.

### B.3.3 Metadata-only rule tracing

Extending `DebugMetaExtractor` exposes final fields but does not fix rule traces:
`core.RuleEngine.traceRule` currently discards non-Schwa rules unless `BaseRom`
changes. During debug execution, capture metadata before and after each applied
rule and retain a trace when either metadata or `BaseRom` changes (preserve the
existing Schwa trace behavior). Apply this to both normal and fallback passes.
Keep extraction behind the debug flag. Include gemination mode, vowel quality,
and `NoInherentVowel` in Brahmic metadata. A focused test must show a Render-phase
metadata-only rule in `Traces`, with its rule name and resulting metadata; also
cover unchanged rules and debug-off execution. Final unit metadata alone is not
a passing test.

### B.4 The inherent-vowel classifier (Phase B2)

Rules alone cap near ~85% on Bengali deletion and cannot do ɔ-vs-o reliably. The
learned path mirrors Hindi's schwa model but with **three output classes
(∅/ɔ/o)**, which the tree-inference loader must support (minor generalization).
Training data: the Google `bn` pronunciation lexicon (60k+ entries). Derive
∅/ɔ/o labels through grapheme–phoneme alignment, with rejected/ambiguous
alignments counted; a phoneme string alone does not identify each unit's label.
The classifier predicts phonological outcomes only. It never uses `Open` as a
proxy for an attested Latin `a` spelling; `Open` belongs to the scoped B1 rules.
Under default style both retained classes render as `o`, so classifier quality
alone cannot resolve Sanskritic/name `a` spellings. Initially those spellings
belong exclusively to the train-only lexicon. A learned spelling selector is
separate, deferred work requiring measured benefit. Do not promise lexical
exception coverage for unseen names.

Dakshina **train** supplies attested spellings for the lexicon and reranker;
**dev** supports tuning and B0 histograms; **test** is held out for final
reporting. This replaces the contradictory v3 statement that all of Dakshina is
benchmark-only. No held-out romanization is a training/mining input. The
non-default-style lexicon policy (style-key or bypass) must be chosen and tested
before B2 ships, so a lookup cannot silently override the requested style.

### B.4.1 Normalization, splits, and overlap assertions

Before training any B2 artifact, commit source versions/checksums, the split
seed, and normalized word-type manifests. Normalize comparison keys consistently
across all sources: Unicode NFC (including composition-excluded nukta aliases),
format-character removal matching the parser, and canonical equivalence of split
matras. Preserve source text for audit; do not merge distinct spellings using
romanized output. If normalization creates cross-split duplicates, exclude the
conflicting training entries rather than moving held-out types into training.

Exclude the **union** of normalized Google dev/test and Dakshina dev/test types
from **every** training pool (the Google pronunciation pool and the Dakshina-train
spelling pool alike) **before alignment, feature extraction, or lexicon
construction**. Split the remaining external
pool by normalized word type; all pronunciations and schwa instances of a word
stay together. Train on its train subset only and tune on its dev subset. Assert
zero overlap between external training and both external held-out sets and
Dakshina dev/test. Apply the Dakshina exclusion check to lexicon and reranker
training inputs too, and require zero held-out lexicon coverage.

Report overlap with Aksharantar, BanglaTLit, and lyrics gold separately, including
full-corpus and unseen-type metrics where applicable. They remain evaluation
sources, never sources of learned spelling entries. Record counts before/after
normalization and exclusion, ambiguous alignment counts, and manifest hashes
with every trained artifact. BanglaTLit's upstream `train` CSV contains every
official dev/test pair; only the pinned official test split may be imported, and
only for evaluation. A pipeline rerun must fail on violated exclusions;
a random split of the Google lexicon alone is not evidence of Dakshina isolation.
Update RESEARCH's Hindi-only training policy with this explicit Bengali extension
in the B2 implementation PR. No overlap counts have been measured in this design.

### B.5 Scheme (decided: reuse `colloquial`)

**Decision:** reuse the existing `colloquial` scheme as-is; do **not** add a
`bengali-colloquial` scheme. `colloquial.SelectRules` is `return
catalog.AllRules()` — already language-agnostic. All of Bengali's rule divergence
lives in the Bengali `RuleCatalog`, not in the scheme. The scheme axis is output
*style* (colloquial vs. academic/IAST), orthogonal to language; coupling style to
language would start an N×M scheme explosion. An academic/ISO-15919 scheme
(inherent vowel `a`, graphemic) is a natural later addition that would then serve
both languages for free. Out of scope for v1.

### B.6 Behavioral options (flags)

Following Hindi's opt-in philosophy (sensible colloquial default; named flags to
switch behavior), Bengali's headline behavioral choices are exposed as `Options`
fields / CLI flags rather than baked in:

- **Kept-vowel style** — the colloquial default spells a kept inherent vowel `o`
  (B.3); a flag selects the academic/graphemic `a` (and later `ô` for the
  ɔ-marking convention). This is the switch for the Q1 register tension: `o` by
  default, `a` on demand, with the lexicon handling attested per-word spelling
  exceptions. The classifier predicts pronunciation, not the spelling register.
- **Chandrabindu** — data-gated (J3, T-0055), likely positional (`n` medially);
  not a hardcoded "drop" default. Exposed as a flag once measured.
- The existing **generic** `--enable-rule` / `--disable-rule` pattern mechanism
  applies to Bengali rules unchanged, so any individual rule is switchable for
  testing without a new flag.

Each flag is covered by the same before/after measurement discipline as Hindi's.

---

## 2. Data plan

| Role | Hindi used | Bengali equivalent | License |
|---|---|---|---|
| Primary benchmark + training | Dakshina hi | **Dakshina bn** — train/dev/test: 25k/2.5k/2.5k word types, 3.8 variants/type; roles and exclusions in B.4.1 | CC BY-SA 4.0 |
| Phonological classifier training | Dakshina hi alignment | **Google bn pronunciation lexicon** — aligned labels; external splits and Dakshina exclusions (B.4.1) | CC BY 4.0 |
| Human benchmark (**not independent**: contains the Dakshina test vocabulary; report full and unseen slices) | Aksharantar hi test | **Aksharantar ben test** — published slices AK-Freq 1,071 / AK-Uni 1,198 / AK-NEF 1,059 / AK-NEI 1,681 (5,009); the released test file deduplicates to more word types (count recorded at evaluation time) | CC-BY |
| Naturally-typed (COMI role) | COMI-LINGUA | **BanglaTLit** — 42,705 Banglish↔Bengali sentence pairs; evaluation-only, official test split only (upstream `train` CSV contains dev/test pairs) | MIT |
| Frequency ranking (Shabd role) | Shabd (CC0) | **derive from IndicCorp v2 Bengali** (~29.6M rows, CC0); Leipzig `ben_wikipedia_2021` as quick first pass | CC0 / CC-BY |
| Lyrics gold | 43 PD lines | **Tagore Gitabitan** (2,232 PD songs, bn.wikisource) + Lalon, D.L. Roy, Atulprasad; Nazrul excluded until 2037 | PD |

Derived artifacts inherit their sources' terms: code stays MIT; artifacts built
from the Google lexicon carry CC BY 4.0 attribution, and Dakshina-derived
artifacts (spelling lexicon, any selector trained on Dakshina spellings) are
CC BY-SA 4.0. Packaged notices must carry these.

Pipelines mirror the Hindi `tools/build_*.py`. The one gap vs. Hindi:
no ready CC0 psycholinguistic frequency DB, so the Bengali frequency list is
derived (documented pipeline) rather than adopted.

---

## 3. Milestones, effort, and gates

Dependency-ordered. Each implementation phase ships as its own green PR. The
v4 design revision stays on `feature/bengali-design`. B0 freezes evaluation boundaries and uses
train/dev for exploratory histograms; test results must not guide rule tuning.

| Phase | Deliverable | Rough effort | Gate |
|---|---|---|---|
| **A** | Generalized Brahmic layer + Hindi golden snapshot | 2–3 days | Hindi byte-identical (3 nets, A.4) |
| **B0** | `lang/bengali` symbol table, config, compositional conjuncts, `bengali` wired into `New()`; naive output; `bn` benchmark harness that **emits the o/a-split and attestation histograms** (gates the Q1 default and the curation threshold) | 3–5 days | builds; Hindi untouched; baseline + histograms measured |
| **B1 (light)** | Non-schwa rules in full; cheap schwa wins (final-cluster keep) + phala (explicit gemination modes) + anusvara/visarga + the `C+ও` rule + behavioral flags (B.6); priority-tie test on the composed catalog. ɔ→o raising deferred (no-op under `o`). Deliberately **not** exhaustive schwa tuning | 1 week | B.3.1–3 prototype examples and traces pass before catalog expansion; match-any on Dakshina `bn` clears a bar **pre-registered from the B0 histograms** (not merely "beats B0"); bn-specific gate set empirically |
| **B2** | Bengali learned components (3-way schwa classifier, lexicon, reranker) | 1–2 weeks | each selected on Dakshina **dev** under a recorded protocol; held-out test reported once after selection, never used to accept/reject; lexicon judged on in-sample/external coverage with no held-out regression; normalized exclusions and artifact provenance asserted (B.4.1) — *amended 2026-10-05* |
| **B3** | Bengali PD lyrics gold set + line-level suite | 2–4 days | line-CER reported |

A usable, honestly-measured rules-only Bengali exists after **A + B0 + B1**
(~3 weeks). Full Hindi-parity (learned components + lyrics) is ~6–7 weeks.
These are rough; the rule-tuning in B1 is the least predictable because of the
inherent-vowel irregularity.

**Amendment (2026-10-05, stack review):** the v4 B2 gate read "each improves
held-out `bn` match-any". That contradicted B.4 (test is for final reporting only)
and B.4.1 (zero held-out lexicon coverage, so the lexicon cannot move a held-out
score). As applied, B2 selected on dev: the vowel model used a *relative* dev
acceptance guard (match-any gain, no strict regression, lower CER) rather than
pre-recorded numeric bars; later selector work fixed its dev protocol before
held-out validation. Test was reported once after selection in each case.

**Metrics discipline for Bengali:** match-any / minCER is the headline (2×
variance, bimodal target), but report it **alongside** strict top-1, macro minCER,
attestation-count strata, and the Aksharantar number (full and Dakshina-test-unseen
slices) — because `bn`'s
3.8 refs/word (vs Hindi's 1.8) *inflates* match-any, so a naive "trailing Hindi"
read could invert and mislead (review). Report AK-Freq / AK-NEF / AK-NEI slices as
for Hindi. Set the pure-accuracy CI gate from the observed `bn` distribution; do
not copy 85%. The Hindi gate and suites are untouched.

---

## 4. Risks and honest limitations

- **The inherent vowel caps Bengali below Hindi.** Deletion rules ~85%; ɔ/o
  ~92.5% even for a trained model; the colloquial o/a target is bimodal. Bengali
  top-1 will trail Hindi's and that is inherent to the task, not a defect.
- **Word-final homographs** (বল, হল, কোন, মত) are unresolvable at word level.
  Accept and document the error class, as RESEARCH §5 does for Hindi.
- **æ (ya-phala-initial) has no settled Banglish spelling** → those words are
  inherently multi-reference.
- **Gemination correctness.** Explicit render-time modes avoid neighbor mutation,
  but aspiration placement, vowel ownership, and cluster exceptions still require
  the B.3.1 prototype gate. Metadata-only trace visibility requires the core
  change in B.3.3, not just a new field or extractor.
- **Cluster/phala ordering.** The "স→s in clusters" rule must test the next unit's
  *identity*, not just after-halant, or স্বামী → *sami* instead of *shami* (review).
- **Frequency data** must be derived, not adopted (no Shabd equivalent).
- **Pre-existing:** a terminal explicit virama is dropped by the parser (the
  preceding consonant regains its inherent vowel); noted, not introduced here.

## 5. Decisions and remaining open questions

**Decided (this review):**
- **Scheme:** reuse `colloquial` (B.5). Not language-coupled.
- **Kept-vowel default:** `o` as the rules default, exposed as a switchable flag
  (B.6); the `a`-exceptions (Sanskritic words, proper names) are handled by the
  train-only lexicon; the ɔ/o classifier cannot select `a` under default style.
  **Precondition:** measure the actual o/a split on Dakshina `bn` train/dev before
  locking the default value — the
  B0 benchmark harness emits this histogram first (T-0055). Front-load the Bengali
  **lexicon for proper nouns**, since names are high-salience (a mangled name is
  noticed where a mangled common word is forgiven), exactly as Hindi's lexicon
  earned its keep on loanwords.
- **Milestone shape:** "B1 light." Build B0 + the non-schwa rules fully (the bulk
  of correctness, needed regardless), implement only the cheap/obvious schwa wins
  (final-cluster keep — the classifier's OOV fallback; ɔ→o raising is deferred, a
  no-op under the `o` default), measure the honest baseline, then invest the saved
  rule-tuning time
  in the B2 classifier. We do **not** repeat Hindi's exhaustive rule-ceiling arc:
  Johny & Jansche already established it (~85% rules, classifier needed).
- **Success bar for v1:** honestly measured and useful, **trailing
  Hindi by design** — not matching Hindi's 92.9%. Match-any / minCER is the
  headline; the pure gate is set empirically from `bn`.

**Original open questions (resolution status, 2026-10-04):**
1. æ romanization for ya-phala-initial (`a` / `e` / `ya`).
2. Resolved by T-0055: max-reference-votes ≥3 for the secondary subset; full
   dev remains primary. See the evaluation contract above.
3. Resolved by T-0055: keep o as the operational default, with explicit a style
   available. Conditional slot evidence and independent অ limitations are
   documented in the evaluation contract.

## 6. References

Linguistics: Johny & Jansche 2018 (SLTU),
<https://www.isca-archive.org/sltu_2018/johny18_sltu.pdf>; Choudhury, Basu &
Sarkar 2004 (SIGPHON), <https://aclanthology.org/W04-0103.pdf>; Google `bn`
lexicon, <https://github.com/google/language-resources/tree/master/bn/data>;
Unicode U+0980 chart and UCD. Data: Dakshina,
<https://github.com/google-research-datasets/dakshina>; Aksharantar,
<https://arxiv.org/abs/2205.03018>; BanglaTLit (EMNLP Findings 2024),
<https://aclanthology.org/2024.findings-emnlp.859.pdf>; IndicCorp v2 /
Sangraha (AI4Bharat); Tagore PD status,
<https://spicyip.com/2015/01/guest-post-iprs-indian-railways-rabindrasangeet.html>;
Gitabitan on Bengali Wikisource, <https://bn.wikisource.org/wiki/গীতবিতান>.
