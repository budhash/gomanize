# Bengali Support — Design

**Date:** 2026-09-30
**Status:** Design proposal (not yet implemented). Tracked as F-0011.
**Scope:** Add Bengali (Bangla) as the second language, in two dependency-ordered
pieces: (A) make the shared Brahmic layer genuinely script-general with zero
change to Hindi output, then (B) build the `lang/bengali` package on top of it.

This doc is grounded in two research passes (2026-09-30): a linguistics review of
Bengali vs. Hindi for the engine, and a dataset-availability survey. Primary
sources are cited at the end; the standing project discipline (train-split-only
training, overrides-are-not-accuracy, measure-before-shipping, negative results
recorded) carries over unchanged.

---

## 1. Summary and the one hard finding

The data story is strong: Bengali has permissively-licensed analogues for every
Hindi data source (Dakshina `bn`, Aksharantar `ben`, BanglaTLit, IndicCorp v2),
plus a *better* lyrics opportunity than Hindi ever had (Tagore's Gitabitan, 2,232
public-domain songs). The engine is partly ready: the parser, category system,
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
first (a config value + a per-unit spelling slot); Part B supplies the second (a
Bengali rule set, and eventually a Bengali-trained classifier).

A second consequence is for measurement: Dakshina `bn` has **3.8 attested
romanizations per word vs. Hindi's 1.8**, and the colloquial target is bimodal
(`o` vs `a` for the same vowel depending on register). Strict top-1 will be
lower than Hindi's regardless of engine quality. **Match-any / minCER is the
Bengali headline metric**, and the pure-accuracy gate must be set empirically
from the `bn` data, not copied from Hindi's 85%.

---

## Part A — Generalize the Brahmic layer (zero Hindi regression)

### A.1 The Devanagari-specific sites in the "shared" layer

Verified against current `main` (v1.2.0 prep):

| Location | Hardcoded (Devanagari) | Role |
|---|---|---|
| `script/brahmic/renderer.go` (kept-schwa writes) | literal `"a"` | spelling of a kept inherent vowel |
| `script/brahmic/renderer.go` `isInherentAVowel` | `অ`→ no; `अ` U+0905 / ऄ U+0904 | the bare-a vowel that coalesces with a preceding schwa |
| `script/brahmic/schwa_rules.go` `isSonorousRune` | र/य/व | sonorant-final schwa retention |
| `script/brahmic/schwa_rules.go` `before-cc` | aa-matra `ा` U+093E | Sanskrit-heuristic deletion guard |
| `script/brahmic/schwa_rules.go` `isWordInitialConjunct` | indep-vowel range U+0905–U+0914 | word-initial conjunct detection |

`Config` already carries `Halant`, `Nukta`, `MultiChar` per language — the parser
is clean. These five are the remaining leaks.

### A.2 `ScriptProfile`: lift the literals into config

Extend `brahmic.Config` (or add a `Profile` field) with:

```go
type Config struct {
    Halant    string
    Nukta     string
    MultiChar []string
    // New (script-general parameters; Hindi values reproduce today's literals):
    InherentVowel        string   // Hindi "a"; Bengali "o"
    BareVowelRunes       []rune   // Hindi {0x0905, 0x0904}; Bengali {0x0985}
    AaMatra              rune     // Hindi 0x093E; Bengali 0x09BE
    SonorousRunes        []rune   // Hindi {र,य,व}; Bengali {র,য,ব,...}
    IndependentVowelRange [2]rune // Hindi {0x0905,0x0914}; Bengali {0x0985,0x0994}
}
```

The five sites read from config instead of literals. **Hindi's config reproduces
the current values exactly**, so this step is behavior-preserving by construction.

### A.3 Separate "keep vs delete" from "what a kept schwa spells"

This is the structural change the inherent vowel forces. Keep `SchwaState` binary
(`Keep` / `Delete`) — that decision is shared and correct for both languages.
Add a per-unit spelling slot so the renderer never hardcodes a vowel:

```go
// in BrahmicData
SchwaRom string // optional; romanization of a KEPT inherent vowel.
                // Empty => renderer uses Config.InherentVowel.
```

Renderer: where it writes `"a"` for a kept schwa, write
`unitSchwaRom(unit)` which returns `BrahmicData.SchwaRom` if set, else
`Config.InherentVowel`.

- **Hindi**: no rule ever sets `SchwaRom`; `InherentVowel="a"` → every call site
  emits `"a"` exactly as today. Byte-identical.
- **Bengali**: `InherentVowel="o"` gives the common case for free; a Bengali
  raising rule (or the classifier) sets `SchwaRom="o"`/`"a"`/`"ô"` per unit for
  the ɔ-vs-o distinction without the shared layer knowing anything Bengali.

This cleanly maps the three-way Bengali outcome onto existing machinery:
`Delete` → delete; `ɔ`/`o` → `Keep` + `SchwaRom`. The classifier's three classes
collapse to (binary state) + (spelling string).

### A.4 Zero-regression guarantee — three independent nets

1. **Byte-identical benchmark gate.** The full Hindi suite
   (`make test-dakshina` + the five benchmark suites) must not move a single
   count: 86.2% pure, 92.9% match-any, 94.8% rerank, held-out, lyrics, COMI.
2. **Frozen Hindi golden snapshot.** Before the refactor, commit a snapshot of
   current Hindi output across the entire curated set (and a sample of each other
   suite) as `testdata/hindi_golden.tsv`; a test diffs live output to it and
   must be exactly zero. This catches any per-word drift the aggregate rates
   could mask.
3. **Ship Part A alone, first.** Part A merges as its own PR, green, *before any
   Bengali code exists*. If a count moves, there is no Bengali to blame — it is a
   pure refactor regression and gets fixed in that PR.

### A.5 Scope boundary — what stays Hindi-only

Everything Hindi-specific already lives in `lang/hindi` and does **not** move:
the Hindi symbol table, the फ→f rule, the व→w rules, the Hindi anusvara
(homorganic n/m) render rules, and `schwa_tree.json` / `lexicon.tsv` /
`roman_ngrams.tsv` (Hindi-trained data). Part A touches only `script/brahmic` and
the Hindi `Config` construction in `lang/hindi/symbols.go`.

---

## Part B — The `lang/bengali` package

Built entirely on the generalized layer from Part A. Mirrors `lang/hindi`'s file
layout: `symbols.go`, `rules.go`, `*_test.go`, and (in later phases) the three
learned-component data files.

### B.1 Symbol table (Bengali block U+0980–U+09FF)

Full inventory is in the linguistics research (consonant, vowel/matra, conjunct,
modifier tables). Notable mappings that differ from Hindi:

- **Consonants:** ঙ→`ng`, ঞ→`n`, ফ→`ph` (not Hindi's `f`), ব→`b` (covers b and
  Sanskrit-v), য→`j`, য়→`y`, ড়→`r`, ঢ়→`rh`, স→`sh` by default / `s` in
  clusters, ৎ (khanda ta)→`t` (vowel-less), ণ merged to `n`, ষ→`sh`.
- **Vowels/matras:** linear logical storage confirmed (no reordering needed for a
  left-to-right rune walk, even for pre-base ি/ে and split ো/ৌ).
- **Aliases (reuse Hindi's `init()` pattern):** ড়/ঢ়/য় are composition-excluded,
  so NFC keeps them decomposed while keyboards emit precomposed — register both.
- **MultiChar:** `ক্ষ`, `জ্ঞ`, `হ্ম`, `হ্ন`, and the **decomposed two-part
  matras** `ে`+`া` (→ো) and `ে`+`ৗ` (→ৌ), plus lone `ৗ`. Alternatively
  NFC-normalize at `Translit` entry; registering MultiChar is more surgical.

### B.2 Bengali `Config`

```go
InherentVowel:         "o",
BareVowelRunes:        []rune{0x0985},            // অ
AaMatra:               0x09BE,                     // া
SonorousRunes:         []rune{'র','য','ব', ...},   // tune against bn data
IndependentVowelRange: [2]rune{0x0985, 0x0994},
Halant:                "্",
Nukta:                 "়",
```

### B.3 Bengali rules (the real linguistic work)

Bengali composes *some* of `brahmic.SchwaRules()` but overrides the word-final
behavior, which is where it diverges most. New/changed rules, all expressible in
the current rule engine:

- **Final-cluster keep (Choudhury R4):** Bengali disallows complex codas, so a
  schwa after a final consonant cluster is kept, spelled `o` (অন্ত *onto*, শব্দ
  *shobdo*, কর্ম *kormo*) — opposite of Hindi (*ant, shabd, karm*).
- **Final-হ keep:** গ্রহ *groho*.
- **Weaker word-final deletion generally** (Bengali lacks Hindi's strong final
  prohibition); plus a documented homograph error class (বল, হল, কোন, মত).
- **ɔ→o raising** via `ConsonantRun` lookahead: high-vowel harmony (বলি *boli*,
  মধু *modhu*), ra-phala-initial (প্রথম *prothom*), any retained final → `o`.
  Implemented by setting `SchwaRom="o"` (A.3).
- **Sanskritic medial retention:** Bengali keeps many medial schwas Hindi deletes
  (রচনা *rochona* not *rochna*); the `ccv` deletion needs a Bengali-tuned guard.
- **Positional phala rules** (no Hindi precedent; mutate neighbors):
  ya-phala ্য (initial → æ vowel; medial → geminate + silent য), ba-phala ্ব
  (initial silent; medial geminate), ma-phala ্ম (similar), with short exception
  lists.
- **Positional conjuncts:** ক্ষ (initial `kh` / medial `kkh`), জ্ঞ (initial
  `g`/`gæ` / medial `gg`).
- **Modifiers:** anusvara ং → `ng` always (retire Hindi's homorganic rules for
  Bengali); visarga ঃ medial → geminate next consonant (দুঃখ *dukkho*);
  chandrabindu ঁ → default drop (colloquial), `n` as a scheme option.

### B.4 The inherent-vowel classifier (Phase B2)

Rules alone cap near ~85% on Bengali deletion and cannot do ɔ-vs-o reliably. The
learned path mirrors Hindi's schwa model but with **three output classes
(∅/ɔ/o)**, which the tree-inference loader must support (minor generalization).
Training data: the Google `bn` pronunciation lexicon (60k+ entries with phonemic
transcription → ∅/ɔ/o labels directly; train-split-only under our contamination
rules), with Dakshina `bn` supplying the colloquial *spelling* of a kept vowel.
Dakshina `bn` remains the benchmark, never a training source.

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
  default, `a` on demand, with the lexicon/classifier handling the per-word
  exceptions in between.
- **Chandrabindu** — default drop (colloquial); flag to emit `n`.
- The existing **generic** `--enable-rule` / `--disable-rule` pattern mechanism
  applies to Bengali rules unchanged, so any individual rule is switchable for
  testing without a new flag.

Each flag is covered by the same before/after measurement discipline as Hindi's.

---

## 2. Data plan

| Role | Hindi used | Bengali equivalent | License |
|---|---|---|---|
| Primary benchmark + training | Dakshina hi | **Dakshina bn** — 25k/2.5k/2.5k word types, 3.8 variants/type | CC BY-SA 4.0 |
| Independent human benchmark | Aksharantar hi test | **Aksharantar ben test** — 5,009 Karya-annotated (AK-Freq 1,071 / AK-Uni 1,198 / AK-NEF 1,059 / AK-NEI 1,681) | CC-BY |
| Naturally-typed (COMI role) | COMI-LINGUA | **BanglaTLit** — 42,705 Banglish↔Bengali sentence pairs | MIT |
| Frequency ranking (Shabd role) | Shabd (CC0) | **derive from IndicCorp v2 Bengali** (~29.6M rows, CC0); Leipzig `ben_wikipedia_2021` as quick first pass | CC0 / CC-BY |
| Lyrics gold | 43 PD lines | **Tagore Gitabitan** (2,232 PD songs, bn.wikisource) + Lalon, D.L. Roy, Atulprasad; Nazrul excluded until 2037 | PD |

Pipelines mirror the Hindi `tools/build_*.py`. The one genuine gap vs. Hindi:
no ready CC0 psycholinguistic frequency DB, so the Bengali frequency list is
derived (documented pipeline) rather than adopted.

A scratchpad Bengali frequency list (`bn_50k.txt`, "word count" format) is already
present from the research pass and can seed the frequency suite for a first pass;
provenance/license to be confirmed before anything is committed.

---

## 3. Milestones, effort, and gates

Dependency-ordered. Each ships as its own green PR.

| Phase | Deliverable | Rough effort | Gate |
|---|---|---|---|
| **A** | Generalized Brahmic layer + Hindi golden snapshot | 2–3 days | Hindi byte-identical (3 nets, A.4) |
| **B0** | `lang/bengali` symbol table, config, compositional conjuncts, `bengali` wired into `New()`; naive output; `bn` benchmark harness that **emits the o/a-split and attestation histograms** (gates the Q1 default and the curation threshold) | 3–5 days | builds; Hindi untouched; baseline + histograms measured |
| **B1 (light)** | Non-schwa rules in full; only the cheap/obvious schwa wins (final-cluster keep, ɔ→o raising) + the phala/anusvara/visarga rules; behavioral flags (B.6). Deliberately **not** an exhaustive schwa-rule tuning pass | 1 week | match-any on Dakshina `bn` beats the B0 baseline; bn-specific gate set empirically |
| **B2** | Bengali learned components (3-way schwa classifier, lexicon, reranker) | 1–2 weeks | each improves held-out `bn` match-any; contamination asserted |
| **B3** | Bengali PD lyrics gold set + line-level suite | 2–4 days | line-CER reported |

A usable, honestly-measured rules-only Bengali exists after **A + B0 + B1**
(~3 weeks). Full Hindi-parity (learned components + lyrics) is ~6–7 weeks.
These are rough; the rule-tuning in B1 is the least predictable because of the
inherent-vowel irregularity.

**Metrics discipline for Bengali:** match-any / minCER is the headline (2×
variance, bimodal target). Report AK-Freq / AK-NEF / AK-NEI slices as for Hindi.
Set the pure-accuracy CI gate from the observed `bn` distribution; do not copy
85%. The Hindi gate and suites are untouched.

---

## 4. Risks and honest limitations

- **The inherent vowel caps Bengali below Hindi.** Deletion rules ~85%; ɔ/o
  ~92.5% even for a trained model; the colloquial o/a target is bimodal. Bengali
  top-1 will trail Hindi's and that is inherent to the task, not a defect.
- **Word-final homographs** (বল, হল, কোন, মত) are unresolvable at word level.
  Accept and document the error class, as RESEARCH §5 does for Hindi.
- **æ (ya-phala-initial) has no settled Banglish spelling** → those words are
  inherently multi-reference.
- **Rule-engine expressiveness:** phala rules mutate *neighboring* units
  (gemination, following-vowel æ). The engine passes `*core.Unit` + `*core.Word`
  so this is expressible, but it is new; prototype one phala rule early in B1 to
  confirm before committing to the full set.
- **Frequency data** must be derived, not adopted (no Shabd equivalent).

## 5. Decisions and remaining open questions

**Decided (this review):**
- **Scheme:** reuse `colloquial` (B.5). Not language-coupled.
- **Kept-vowel default:** `o` as the rules default, exposed as a switchable flag
  (B.6); the `a`-exceptions (Sanskritic words, proper names) are handled by the
  lexicon and ɔ/o classifier, not hardcoded as rules. **Precondition:** measure
  the actual o/a split on Dakshina `bn` before locking the default value — the
  B0 benchmark harness emits this histogram first (T-0055). Front-load the Bengali
  **lexicon for proper nouns**, since names are high-salience (a mangled name is
  noticed where a mangled common word is forgiven), exactly as Hindi's lexicon
  earned its keep on loanwords.
- **Milestone shape:** "B1 light." Build B0 + the non-schwa rules fully (the bulk
  of correctness, needed regardless), implement only the cheap/obvious schwa wins
  (final-cluster keep, ɔ→o raising — which double as the classifier's OOV
  fallback), measure the honest baseline, then invest the saved rule-tuning time
  in the B2 classifier. We do **not** repeat Hindi's exhaustive rule-ceiling arc:
  Johny & Jansche already established it (~85% rules, classifier needed).
- **Success bar for v1:** honestly measured and genuinely useful, **trailing
  Hindi by design** — not matching Hindi's 92.9%. Match-any / minCER is the
  headline; the pure gate is set empirically from `bn`.

**Still open (resolve against `bn` data):**
1. æ romanization for ya-phala-initial (`a` / `e` / `ya`).
2. Curated-set attestation threshold for `bn` (Hindi's ≥4 likely yields too few
   given 3.8 variants/type — inspect the histogram alongside the o/a split).
3. The exact `o` vs `a` default, pending the T-0055 measurement (the mechanism is
   decided; the default *value* is data-gated).

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
Full citations and per-source detail live in the two 2026-09-30 research reports
(session artifacts).
