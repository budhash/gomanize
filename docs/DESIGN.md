# Design

How gomanize works: the engine architecture, the rule system, the colloquial
romanization scheme, the embedded learned components, and where the design can
go next. The research grounding for every design choice is in
[`RESEARCH.md`](RESEARCH.md); dated primary sources in [`reviews/`](reviews/).

---

## 1. Architecture

Four layers with strict knowledge boundaries: `core` knows nothing about
Devanagari, `script` nothing about Hindi, `lang` nothing about output style:

```
core/     Engine mechanics: pipeline, rule engine, types, optional interfaces
script/   Script family (brahmic): parsing, rendering, schwa state, shared rules
lang/     Languages (hindi, experimental bengali): symbol map, language rules, learned artifacts
scheme/   Output style (colloquial): selects rules from the language catalog
```

Every word flows through a four-stage pipeline (`core/engine.go`):

```
Parse → Prepare → Rules → Render
```

1. **Parse** (`script/brahmic/parser.go`) — walks runes against the language's
   symbol map; combines nukta; **consumes halant without emitting a unit**,
   flagging the following consonant as after-halant (conjunct). Output: a
   `Word` of doubly-linked `Unit`s, each seeded with a base romanization and
   script-specific state in `Unit.ScriptData`. The parser resolves `Config.Profile`
   and stores it in `WordBrahmicData`; run preparation preserves this profile.
2. **Prepare** (`script/brahmic/runs.go`) — groups consecutive consonants
   between vowels into `ConsonantRun`s so schwa decisions can be coordinated.
3. **Rules** (`core/rule.go`) — four phases in fixed order: **Schwa →
   Consonant → Vowel → Render**, mutating `Unit.BaseRom` and schwa state.
4. **Render** (`script/brahmic/renderer.go`) — concatenates base romanizations,
   spelling retained inherent vowels from the script profile and per-unit
   `SchwaQuality`. Matras and configured bare independent vowels suppress the
   preceding inherent vowel; other independent vowels retain it regardless of
   the schwa decision. Halant-linked clusters suppress the preceding vowel,
   and otherwise `Delete` suppresses it while `Pending` behaves like `Keep`.

`BrahmicData.Gemination` supports `RepeatPrevious` (a halant-linked phala
substitutes the preceding simple onset's geminate pair) and `GeminateSelf`
(the unit repeats its own onset). Aspirated pairs use `kh → k + kh`.
Rendering does not mutate neighboring units or force a vowel decision. Atomic,
larger-cluster, and intrinsically dead onsets are rejected. These modes are
used by the Bengali B1 catalog and independently checked by prototype tests.

`brahmic.ScriptProfile` holds inherent/raised spellings, bare-vowel identities,
the aa-matra, sonorants, and the independent-vowel range. Hindi supplies
`DevanagariProfile()` explicitly. Legacy `Config` values and words without a
profile retain Devanagari defaults. Parsers copy supplied profile slices; parsed
profiles are read-only. A non-nil profile must set every field (`NewParser`
panics, naming the missing ones); explicit empty membership slices disable that
check; a nil `Profile` keeps Devanagari defaults. `Options.InherentVowelA` selects `a` for Default
quality; Raised uses the profile's raised spelling and Open uses `a`. This is a
renderer option, not a complete academic scheme or a rewrite of lexicon hits.
It does not change Hindi output and is not exposed as a CLI/WASM flag yet.

Two optional capabilities hook in *before* the pipeline, via interfaces a
language may implement (`core/engine.go`):

- `LexiconProvider` / `OptionsLexiconProvider` — `Options.Lexicon`: known words
  short-circuit to their attested spelling; OOV falls through unchanged. The
  option-aware interface takes precedence, including on a miss. Bengali uses it
  to bypass its default-style lexicon when alternate-style flags are requested;
  Hindi retains the original interface and behavior.
- `Reranker` — `Options.Rerank`: the engine runs the pipeline under candidate
  configurations and the language picks the most natural output.
- `NativeReranker` takes precedence over `Reranker`, including on a decline.
  It receives source text/options and a synchronous `CandidateRenderer` callback
  preserving the caller's scheme and rule controls. Candidate rendering disables
  recursive reranking, lexicon lookup and debug traces. Declines retain the original
  pipeline; handled results return no single-path debug trace. Bengali uses this
  for its frozen selector, requiring `SchwaModel` plus `Rerank` and default style.
  See the [runtime contract](reviews/2026-10-05-bengali-b2-runtime.md).

The sentence-level API (`gomanize.Translit`) segments on **all whitespace and
punctuation, preserved verbatim** — so multi-line lyrics romanize correctly and
an attached danda (जीत।) cannot defeat word-final rules. Devanagari combining
marks are Unicode marks, never split points.

## 2. The rule system

Rules are declarative structs (`core/rule.go`) with:

| Field | Meaning |
|---|---|
| `Phase` | Schwa / Consonant / Vowel / Render — fixed execution order |
| `Scope` | Universal (0) / Script (100) / Language (200) / Scheme (300) priority base |
| `Priority` | 0–99 within scope; effective priority = scope + priority, higher first |
| `Mode` | **Exclusive** (first match wins per unit) / **Always** / **Fallback** (only for untouched units) |
| `Conditional` | Option gate (e.g. `"SchwaModel"`, `"!KeepMedialSchwa"`) |
| `Condition` / `Action` | Predicates and mutations over `(Unit, Word)` |

Both `NewRuleEngine` and `AddRule` reject duplicate effective priorities within
a phase, including disabled rules; equal priorities in different phases are valid.

Rules identify characters by **source runes** where practical; a few conditions
still test intermediate `BaseRom` strings (the व→w converter's state guard, the
ज्ञ/ीय-suffix rules, and neighbor checks in the conjunct rule). Those remaining
sites are tracked for conversion — matching on output strings couples rules to
the symbol table.

The whole catalog is runtime-inspectable and toggleable: `--list-rules`,
`--disable-rule`, `--enable-rule`, and `--debug` traces changes to base romanization or script metadata
per unit, plus every applied Schwa rule. Normal and fallback passes both capture
metadata before/after actions; debug-off execution never calls the extractor.

### Schwa handling — the core mechanism
Every consonant starts `SchwaPending`; schwa-phase rules move it to `Keep` or
`Delete`; the renderer obeys. A `ConsonantRun` allows **at most one deletion per
run**, preventing cascade deletions (जनता→janta, never *jnt*).

Shared Brahmic schwa rules (`script/brahmic/schwa_rules.go` — Hindi uses all of
them; Bengali composes the word-final delete and default keep):

| Rule | Effect |
|---|---|
| `schwa.keep.sonorous-final` | Word-final conjunct in र/य/व keeps schwa (मंत्र→mantra) |
| `schwa.delete.ccv` | Medial C+C+V deletion (जनता→janta) — the classic rule |
| `schwa.delete.cccc-final` | 4-consonant words ending in consonants (मकसद→maksad) |
| `schwa.delete.before-cc` | Compound-boundary deletion (देशभर→deshbhar) |
| `schwa.delete.word-final` | Final schwa deleted (भारत→bharat) |
| `schwa.keep.default` | Fallback: keep |

Hindi adds language-specific schwa rules (ज्ञ-final, ीय-suffix, and the
learned-model rule below), plus consonant rules (व→v/w contexts, फ→f with the
फू exception), vowel rules (closed-final ा→aa, िए glide), and render rules
(nasal endings, ांव→aon, etc.).

## 3. The colloquial scheme

Design principles: **no diacritics** (aa not ā), **schwa deletion matching
spoken Hindi**, **phonetic ASCII spelling** (kh not ḵẖ), **vowel length marked
only where conventional** (काम→kaam but गाना→gana — the broader rule was
measured net-negative, see RESEARCH §5).

### Character mappings (verified against the engine)

**Vowels**

| Devanagari | Matra | Output | Notes |
|---|---|---|---|
| अ | (inherent) | a | subject to schwa deletion |
| आ | ा | a / aa | aa in closed final syllable (काम→kaam) |
| इ / ई | ि / ी | i | no length distinction (deliberate; see RESEARCH §5) |
| उ / ऊ | ु / ू | u | no length distinction |
| ऋ | ृ | ri | |
| ए / ऐ | े / ै | e / ai | |
| ओ / औ | ो / ौ | o / au | |

**Consonants**

| | | | | | |
|---|---|---|---|---|---|
| क k | ख kh | ग g | घ gh | ङ n | च ch |
| छ chh | ज j | झ jh | ञ ny | ट t | ठ th |
| ड d | ढ dh | ण n | त t | थ th | द d |
| ध dh | न n | प p | फ ph/f | ब b | भ bh |
| म m | य y | र r | ल l | व v/w | श sh |
| ष sh | स s | ह h | | | |

**Nukta (Perso-Arabic)**: क़ q · ख़ kh · ग़ gh · ज़ z · फ़ f · ड़ d · ढ़ dh

**Conjuncts**: क्ष ksh · त्र tr · ज्ञ gy (ज्ञान→gyaan) · श्र shr (श्री→shri)

### Comparison with standards

| Character | ISO 15919 | Hunterian | Gomanize |
|---|---|---|---|
| आ | ā | ā, a | aa, a |
| ई | ī | ī, i | i |
| च | ca | cha | ch |
| व | va | wa, v- | v (w in specific contexts) |
| ज्ञ | jña | gy | gy |

### Deliberate divergences
Measured and intentional (each backed by a decision record):
जनता→*janta* (schwa deleted; also attested), मंत्र→*mantra* (readability),
गाना→*gana* / संगीत→*sangit* (vowel-length rules measured net-negative). The
`--keep-medial-schwa` flag restores janata-style output.

## 4. Learned components (all embedded, zero runtime dependencies, opt-in)

| Component | Artifact | Integration | Measured value |
|---|---|---|---|
| Schwa classifier (`--schwa-model`) | CART tree, 34 KB JSON (`lang/hindi/schwa_tree.json`) | `schwa.model.predict` rule (Language:90, Exclusive) takes over inherent-schwa decisions | 90.67% per-schwa held-out; ties/beats the 8 hand rules word-level |
| Lexicon (`--lexicon`) | 8,367-entry TSV, ~204 KB (`lang/hindi/lexicon.tsv`) | `core.LexiconProvider` pre-pipeline short-circuit; lossless OOV fallthrough | 71.7% token coverage; COMI-LINGUA +6.7 pts, AK-NEI +7.8 pts, lyrics CER 0.0492→0.0394 |
| Re-ranker (`--rerank`) | Char 4-gram LM, 31K grams, 224 KB (`lang/hindi/roman_ngrams.tsv`) | `core.Reranker`: scores {default rules, schwa-model} outputs, stupid backoff, per-char normalized; ties keep the default | Held-out 69.3→70.7%, curated match-any 92.9→94.8%, lyrics CER 0.0492→0.0476 |
| Bengali vowel model (`SchwaModel`) | Three-class CART tree (`lang/bengali/vowel_tree.json`), Google pronunciation lexicon train partition | `schwa.bengali.model` rule on supported simple words; unsupported words keep B1 rules | Held-out match-any 56.52→62.76% |
| Bengali lexicon (`Lexicon`) | 8,976-entry TSV (`lang/bengali/lexicon.tsv`), Dakshina train | `OptionsLexiconProvider`; default style only | Zero held-out coverage by construction; helps known words |
| Bengali native selector (`SchwaModel` + `Rerank`) | Cross-fitted preference tree (`lang/bengali/selector.json`) | `core.NativeReranker`: up to eight single-slot vowel flips, threshold 0.6 | Held-out match-any 62.76→63.04% |

Design constraints that shaped them:
- **Train on train partitions only** (Hindi: Dakshina TRAIN; Bengali: the frozen
  schema-2 partitions, which also exclude spelling variants of held-out words).
- **Canonical input** — the engine canonicalizes each word once (Unicode NFC
  subset, format characters in script text) before any component sees it.
- **Candidate quality gates re-ranking** — only individually-strong candidates
  enter the pool; ablation showed weak candidates drag it below baseline.
- **Everything distills to data files + ~50 lines of Go inference** — the
  Python training pipelines (`tools/schwa/`, `tools/train_ngram.py`,
  `tools/schwa/build_lexicon.py`) are pure stdlib and reproducible.
- **Nothing auto-promotes into gold.** The candidate miner
  (`tools/mine_overrides.py`) measured 43% unreviewed precision — output is
  human-review-only by design.

## 5. What is implemented

- Languages: **hindi** and experimental **bengali**; one scheme (**colloquial**).
- Options: `InherentVowelA` (Go API only), `LongVowels`, `SimpleNasals`,
  `KeepMedialSchwa`, `SchwaModel`, `Lexicon`, `Rerank`, `Debug`
- CLI: `--language`, all options except `InherentVowelA` as flags, plus
  `--input=FILE` / `--test=FILE` / `--diff` / `--list-rules` / `--disable-rule` /
  `--enable-rule` / `--debug` / `--version`
- Public API: `New(lang)`, `NewWithOptions(lang, opts, engineOpts...)`,
  `Translit(text)`, `TranslitDebug(word)` (single word; no sentence splitting),
  rule management via `ListRules/DisableRule/EnableRule`
- Frontends: WASM/npm (`{ language, ...flags }`) and the web demo
- Evaluation: accuracy, parser-QA and Bengali gate/record suites run by `make ci`

## 6. Future directions

**Tractable next (unblocked by current design):**
- **Bengali (Bangla)** — implemented as experimental (F-0011; see the
  [design](reviews/2026-09-30-bengali-support-design.md) and its as-built notes,
  and the [keel review](reviews/2026-10-07-bengali-keel.md)). B1 uses
  source-identity predicates and per-unit gemination metadata for scoped phalas
  and visarga; its dev gate runs in `make ci`. Next steps are tracked in F-0013
  (language and option registries, one reranker protocol, performance) and
  T-0068 (lyrics reference attestation).
- **Marathi / Nepali** — implement `core.Language` (symbol map + config + rule
  catalog composing `brahmic.SchwaRules()`); parser/renderer/runs are reused.
  Supply a `Config.Profile` and audit which shared schwa rules fit the language;
  configurable literals do not make Hindi deletion heuristics universal.
- **Lexicon growth** past the 78.2% train-gold coverage ceiling — human review
  of mined candidates, or new attested sources (Xlit-Crowd is CC-BY-NC-SA).
- **Lyrics gold expansion** — more public-domain verse in-repo; Giitaayan
  (ITRANS→Devanagari) + LyricsTranslate curation out-of-repo.

**Requires design work:**
- **IAST / scholarly scheme** — schemes currently select rules from the
  language catalog, but base romanizations live in the symbol table (colloquial
  choices baked in). A clean IAST needs per-scheme symbol maps: an interface
  change.
- **Sentence-level context** — the engine is word-independent; disambiguating
  by context (Kirov et al. 2024's noisy-channel approach) would need a
  sentence-level LM pass.

**Out of scope by construction:**
- **Roman→Devanagari** — the pipeline is lossy (schwa deletion, ई/इ collapse,
  श/ष merge) and one-directional; the reverse task is a sequence-disambiguation
  problem best served by a separate model, not this engine.
