# Changelog

All notable changes to gomanize are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The npm package [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize)
is the same engine compiled to WebAssembly and shares this version line. Each
tagged release also has auto-generated notes on the
[GitHub releases page](https://github.com/budhash/gomanize/releases).

## [Unreleased]

### Fixed (keel review, 2026-10-07)

- Text without any character of the selected script (emoji ZWJ sequences, flag
  tags, RTL marks, soft hyphens) now passes through unchanged instead of losing
  its format characters.
- Learned per-word components (schwa/vowel models, Bengali selector) skip tokens
  longer than 64 runes; very long unspaced input no longer costs quadratic time
  (a 2,000-character Bengali word took ~30 s in the web demo).
- Debug tracing is collected per call: `Translit`/`TranslitDebug` with
  `Options.Debug` are now safe for concurrent use on a shared instance.
- npm wrappers snapshot options into validated primitives before calling WASM, so
  a getter or Proxy cannot pass a BigInt through and crash the runtime.
- A non-nil `brahmic.ScriptProfile` must set every field; partial profiles now
  panic instead of silently inheriting Devanagari values.
- New `core.Options.DefaultStyle()` centralizes the default-style check used by
  Bengali's lexicon and selector, covering any style option added later.
- The Hindi lexicon coverage test now asserts zero held-out coverage.

### Changed

- npm/WASM accepts per-call `language: "bengali"`; omitted language remains Hindi.
  The browser adds explicit language selection and remembers options per language.
  Unsupported JavaScript language values throw without terminating the WASM engine.

- CLI accepts `--language=bengali` or `--language bengali` for every input mode;
  omitted language remains Hindi. Bengali learned components stay opt-in.
  Unsupported names fail with the supported list; Bengali `--rerank` without
  `--schwa-model` prints a warning instead of being silently ignored.
- Bengali training partitions (schema 2) exclude khanda-ta (ৎ / ত্ / word-final
  ত) and malformed-আ spelling variants of held-out words; vowel model, lexicon
  (8,976 entries) and native selector retrained. Held-out results are unchanged.
- npm wrappers reject non-object options and non-boolean flags (a BigInt flag
  previously crashed the WASM runtime). The npm license field and release
  archives now carry the embedded-data licenses (`NOTICE.md`).

- Debug rule traces now retain metadata-only changes in all phases and fallback
  passes; metadata extraction remains disabled outside debug execution.

- Generalized the Brahmic renderer and shared rule constants through an optional
  script profile, with separate retained-vowel quality. Legacy callers retain
  Devanagari defaults; Hindi output is protected by a frozen corpus snapshot.
- Rule-engine construction now rejects same-phase effective-priority conflicts,
  matching `AddRule`, including rules disabled by default.

### Added

- Experimental Bengali B1 rules: scoped phalas/conjuncts, final-cluster vowels,
  f spelling, and হও/হওয়া handling; fixed dev acceptance gate now enforced in CI.
- Pinned BanglaTLit sentence test fixture and importer, plus reviewed Bengali
  and reverse-transliteration references with provenance/overlap findings.

- Bengali train/dev vowel evidence, frozen B0 dev outputs, and a preregistered
  B1 acceptance command (`make test-bengali-b1-gate`; B0 intentionally fails).

- Brahmic render-time gemination modes with independent vowel ownership,
  validated by test-only Bengali B1 prototypes. Public Bengali rules remain B0.

- Experimental Bengali B0 through the Go API: symbols, Unicode aliases,
  compositional conjuncts, khanda-ta support, and split-aware Dakshina baseline.
  Bengali pronunciation rules and CLI/npm/web selection are not yet complete.
- `Options.InherentVowelA` for profile-aware rendering; no Hindi output change.
  This API option is not yet a CLI flag or a complete academic scheme.

## [1.2.1] - 2026-09-13

### Added

- Homebrew install: `brew install budhash/tools/gomanize` (formula auto-published
  to the shared [`budhash/homebrew-tools`](https://github.com/budhash/homebrew-tools)
  by GoReleaser on each release). No engine changes since 1.2.0.

## [1.2.0] - 2026-09-11

Behaviour-changing romanization fixes for two common construct classes, plus a
structured parser-QA program that found them and guards against regressions.

### Fixed

- **Consonant + independent vowel** now keeps the consonant's inherent vowel
  instead of collapsing onto the vowel: गई → `gai` (was `gi`), नई → `nai`,
  कई → `kai`, गए → `gae`, and similar. An independent vowel starts its own
  syllable; only a matra binds to the preceding consonant.
- **Chandrabindu nasalization** is now kept mid-word and in longer word-final
  forms: चाँद → `chaand` (was `chaad`), साँस → `saans`, पाँच → `paanch`,
  कहाँ → `kahaan`, गाँव → `gaanv`. The nasal is dropped only for the lexical
  exception माँ → `maa`.

### Added

- Structured parser-QA test program: per-construct accuracy analyzer, gold-free
  structural invariants, combinatorial construct enumeration, an akshara
  segmentation differential, a held-out construct regression set (129 curated
  words, match-any), and a corpus-diff regression harness. New make targets:
  `test-constructs`, `test-heldout`, `test-combinatorial`, `test-akshara`,
  `regression-baseline` / `regression-diff`.
- `docs/reference/candidate-datasets.md`: a ledger of external Hindi datasets
  evaluated for future use, and `tools/mine_constructs.py` for building
  construct-stratified candidate sets.

### Changed

- Browser demo sample text is now a public-domain Ghalib couplet (was a
  copyrighted film lyric), and the README was simplified.
- Accuracy on curated Dakshina edged up with the fixes: 86.2% pure /
  92.9% match-any / 94.8% with `--rerank` (from 86.1 / 92.8 / 94.7). No public
  API change; romanization output changes only for the construct classes above.

## [1.1.0] - 2026-09-08

### Added

- Browser demo — the full engine compiled to WebAssembly, running client-side
  at [budhash.com/gomanize](https://budhash.com/gomanize) (`make wasm` /
  `make wasm-serve`, auto-deployed by `pages.yml`).
- npm package [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize)
  — the WASM engine plus a JS/TS wrapper (not a reimplementation, so output is
  identical to the CLI); published via OIDC trusted publishing on version tags.

## [1.0.0] - 2026

Initial public release: Go library and CLI romanizing Devanagari (Hindi) to
Latin, with a rule-based engine and optional embedded learned components
(schwa classifier, attested lexicon, character-LM re-ranker). MIT licensed.

[1.2.1]: https://github.com/budhash/gomanize/releases/tag/v1.2.1
[1.2.0]: https://github.com/budhash/gomanize/releases/tag/v1.2.0
[1.1.0]: https://github.com/budhash/gomanize/releases/tag/v1.1.0
[1.0.0]: https://github.com/budhash/gomanize/releases/tag/v1.0.0
