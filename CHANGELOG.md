# Changelog

All notable changes to gomanize are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The npm package [`@budhash/gomanize`](https://www.npmjs.com/package/@budhash/gomanize)
is the same engine compiled to WebAssembly and shares this version line. Each
tagged release also has auto-generated notes on the
[GitHub releases page](https://github.com/budhash/gomanize/releases).

## [Unreleased]

### Added

- **Experimental Bengali** (`gomanize.New("bengali")`, CLI `--language=bengali`,
  npm `{ language: "bengali" }`, web language selector). Omitted language remains
  Hindi everywhere; every Bengali learned component is opt-in.
  - B0: symbols, Unicode aliases, compositional conjuncts, khanda-ta support and a
    split-aware Dakshina baseline.
  - B1 rules: scoped phalas/conjuncts, gemination, final-cluster vowels, ফ→f,
    হও/হওয়া handling; a preregistered dev gate enforced in CI
    (`make test-bengali-b1-gate`).
  - B2 learned components: a three-class inherent-vowel model (`SchwaModel`), an
    8,976-entry spelling lexicon (`Lexicon`), and a cross-fitted native selector
    (`SchwaModel` + `Rerank`). Held-out Dakshina match-any: rules 56.52%, model
    62.76%, model + selector 63.04%.
  - Frozen, normalized training partitions with held-out exclusions (schema 2
    also excludes khanda-ta and malformed-আ variants of held-out words).
  - Evaluation fixtures: pinned BanglaTLit test split, Aksharantar overlap audit,
    and a source-pinned lyrics pilot (first four songs of the 1913 Gitanjali)
    with **unreviewed** references pending attestation.
- `Options.InherentVowelA` for profile-aware rendering (Go API only); no Hindi
  output change.
- `core.Options.DefaultStyle()`, `core.Canonicalizer`, `brahmic.Canonicalize`
  and `brahmic.ScriptProfile` for script-general rendering and canonicalization.

### Changed

- The Brahmic renderer and shared rule constants take a script profile; legacy
  callers keep Devanagari defaults and Hindi output is guarded by a frozen corpus
  snapshot. A non-nil `ScriptProfile` must set every field (partial profiles
  now panic instead of inheriting Devanagari values).
- Rule-engine construction rejects same-phase effective-priority conflicts,
  matching `AddRule`, including rules disabled by default.
- Script text is canonicalized once per word before lexicon lookup, ranking and
  parsing. Hindi output no longer depends on input encoding; 110 non-canonically
  encoded snapshot inputs changed deliberately and headline accuracy is unchanged
  (docs/reviews/2026-10-09-keel-k1-canonicalization.md).
- CLI: unsupported language names fail with the supported list; Bengali
  `--rerank` without `--schwa-model` prints a warning.
- npm/WASM: options are snapshotted into validated primitives; non-object
  options and non-boolean flags throw `TypeError`. The npm license field is
  `MIT AND CC-BY-SA-4.0 AND CC-BY-4.0`; release archives include `NOTICE.md`.
- Debug rule traces retain metadata-only changes in all phases and fallback
  passes; metadata extraction remains disabled outside debug execution.

### Fixed

- Text without any character of the selected script (emoji ZWJ sequences, flag
  tags, RTL marks, soft hyphens) passes through unchanged.
- Learned per-word components skip tokens longer than 64 runes, so very long
  unspaced input no longer costs quadratic time (a 2,000-character Bengali word
  took ~30 s in the web demo); canonicalization is linear on adversarial input.
- Debug tracing is collected per call: `Translit`/`TranslitDebug` with
  `Options.Debug` are safe for concurrent use on a shared instance.
- A BigInt, getter or Proxy option can no longer crash the npm WASM runtime.
- The Hindi lexicon coverage test asserts zero held-out coverage, as documented.

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
