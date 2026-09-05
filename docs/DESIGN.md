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
lang/     Language (hindi): symbol map, language rules, learned artifacts
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
   script-specific state in `Unit.ScriptData`.
2. **Prepare** (`script/brahmic/runs.go`) — groups consecutive consonants
   between vowels into `ConsonantRun`s so schwa decisions can be coordinated.
3. **Rules** (`core/rule.go`) — four phases in fixed order: **Schwa →
   Consonant → Vowel → Render**, mutating `Unit.BaseRom` and schwa state.
4. **Render** (`script/brahmic/renderer.go`) — concatenates base romanizations,
   emitting the inherent `a` after a consonant unless a vowel follows, the unit
   is part of a conjunct, or a schwa rule decided `Delete`.

Two optional capabilities hook in *before* the pipeline, via interfaces a
language may implement (`core/engine.go`):

- `LexiconProvider` — `Options.Lexicon`: known words short-circuit to their
  attested spelling; OOV falls through unchanged (lossless).
- `Reranker` — `Options.Rerank`: the engine runs the pipeline under candidate
  configurations and the language picks the most natural output.

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

Rules identify characters by **source runes** where practical; a few conditions
still test intermediate `BaseRom` strings (the व→w converter's state guard, the
ज्ञ/ीय-suffix rules, and neighbor checks in the conjunct rule). Those remaining
sites are tracked for conversion — matching on output strings couples rules to
the symbol table.

The whole catalog is runtime-inspectable and toggleable: `--list-rules`,
`--disable-rule`, `--enable-rule`, and `--debug` traces every rule application
per unit.

### Schwa handling — the core mechanism
Every consonant starts `SchwaPending`; schwa-phase rules move it to `Keep` or
`Delete`; the renderer obeys. A `ConsonantRun` allows **at most one deletion per
run**, preventing cascade deletions (जनता→janta, never *jnt*).

Shared Brahmic schwa rules (`script/brahmic/schwa_rules.go` — reused by any
future Brahmic language):

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
| Lexicon (`--lexicon`) | 8,367-entry TSV, ~204 KB (`lang/hindi/lexicon.tsv`) | `core.LexiconProvider` pre-pipeline short-circuit; lossless OOV fallthrough | 71.1% token coverage; +5.9 to +7.9 pts on three independent evals |
| Re-ranker (`--rerank`) | Char 4-gram LM, 31K grams, 224 KB (`lang/hindi/roman_ngrams.tsv`) | `core.Reranker`: scores {default rules, schwa-model} outputs, stupid backoff, per-char normalized; ties keep the default | Improved held-out (69.0→70.4%), curated match-any (92.8→94.7%), and lyrics CER (0.0492→0.0465) |

Design constraints that shaped them:
- **Train on Dakshina TRAIN only** (splits are type-disjoint, so held-out results are uncontaminated).
- **Candidate quality gates re-ranking** — only individually-strong candidates
  enter the pool; ablation showed weak candidates drag it below baseline.
- **Everything distills to data files + ~50 lines of Go inference** — the
  Python training pipelines (`tools/schwa/`, `tools/train_ngram.py`,
  `tools/schwa/build_lexicon.py`) are pure stdlib and reproducible.
- **Nothing auto-promotes into gold.** The candidate miner
  (`tools/mine_overrides.py`) measured 43% unreviewed precision — output is
  human-review-only by design.

## 5. What is implemented (v1.0 surface)

- One language (**hindi**) and one scheme (**colloquial**); the legacy comparison engine was removed pre-1.0.
- Options: `LongVowels`, `SimpleNasals`, `KeepMedialSchwa`, `SchwaModel`,
  `Lexicon`, `Rerank`, `Debug`
- CLI: all options as flags, plus `--list-rules` / `--disable-rule` /
  `--enable-rule` / `--debug` / `--test=FILE` / `--diff` / `--version`
- Public API: `New(lang)`, `NewWithOptions(lang, opts, engineOpts...)`,
  `Translit(text)`, `TranslitDebug(word)`, rule management via
  `ListRules/DisableRule/EnableRule`
- Evaluation: five benchmark suites (see RESEARCH §3–4) run by `make ci`

## 6. Future directions

**Tractable next (unblocked by current design):**
- **Marathi / Nepali** — implement `core.Language` (symbol map + config + rule
  catalog composing `brahmic.SchwaRules()`); parser/renderer/runs are reused.
  Caveats: renderer's inherent-vowel is hardcoded `"a"`; a few Hindi rules
  carry Devanagari literals worth auditing per language.
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
