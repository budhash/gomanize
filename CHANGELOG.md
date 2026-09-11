# Changelog

All notable changes to gomanize are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The npm package [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize)
is the same engine compiled to WebAssembly and shares this version line. Each
tagged release also has auto-generated notes on the
[GitHub releases page](https://github.com/budhash/gomanize/releases).

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

[1.2.0]: https://github.com/budhash/gomanize/releases/tag/v1.2.0
[1.1.0]: https://github.com/budhash/gomanize/releases/tag/v1.1.0
[1.0.0]: https://github.com/budhash/gomanize/releases/tag/v1.0.0
