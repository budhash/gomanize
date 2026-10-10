# TASKS.md

Single source of truth for features + tasks in this repo.

Use `./tools/tasks.py help` for CLI commands.

---

# Meta

## Info

This file is managed by `./tools/tasks.py` CLI tool.

- Run `./tools/tasks.py help` for full documentation
- Manual edits may break parsing - use CLI commands instead
- Task IDs (F-####, T-####) are auto-generated and must be unique
- Checkbox state must match status: `[x]` for done, `[ ]` otherwise

## Schema

Format: `- [ ] (ID) [PRIO] [STATUS] Title @tags...`

Example:
```
- [ ] (F-0001) [P0] [todo] Feature title @issue=42 @tags=security,mvp
  - [ ] (T-0001) [P0] [todo] Subtask @deps=T-0002 @effort=4h
  - [x] (T-0002) [P0] [done] Another subtask @done=2025-01-01
```

Tags: `@deps=` `@rel=` `@branch=` `@pr=` `@issue=` `@tags=` `@effort=` `@system=` `@done=`

---

# Tasks

## Now

- [x] (F-0001) [P0] [done] H0: Absorb skein tooling & process @branch=feature/review-and-tooling-absorption @done=2026-09-04
  - [x] (T-0001) [P0] [done] Vendor tasks.py + ./tools/tasks wrapper @done=2026-09-04
  - [x] (T-0002) [P0] [done] Fix dead make targets (test-dakshina/test-unit run legacy only) @done=2026-09-04
  - [x] (T-0003) [P1] [done] Add docs/PROCESS.md + PR template + task-aware CLAUDE.md section @done=2026-09-04
  - [x] (T-0004) [P2] [done] Pre-commit parity: no-commit-to-branch + large-file/whitespace guards @done=2026-09-04

- [ ] (F-0011) [P2] [todo] Bengali (Bangla) support @shadow
  - [x] (T-0050) [P2] [done] Part A: generalize Brahmic layer (ScriptProfile config + SchwaRom); zero Hindi regression + golden snapshot @done=2026-10-04 @tags=as-built-SchwaQuality
  - [ ] (T-0058) [P2] [todo] A.2: measure second-unit cccc-final index correction against attested Hindi data; review per-input snapshot deltas separately @deps=T-0050
  - [x] (T-0051) [P2] [done] B0: lang/bengali symbol table + config + compositional conjuncts; wire New(); bn benchmark harness @deps=T-0050 @branch=feature/bengali-b0 @done=2026-10-04
  - [x] (T-0056) [P2] [done] B1 prototype: phala/self gemination modes, hao/howa contexts, khanda-ta and metadata-only rule traces; gate catalog expansion @deps=T-0051 @branch=feature/bengali-b1-prototype @done=2026-10-04
  - [x] (T-0055) [P2] [done] Measure o/a split + attestation histogram on Dakshina bn (gates Q1 default + curation threshold) @deps=T-0051 @branch=feature/bengali-evaluation-gate @done=2026-10-04
  - [x] (T-0052) [P2] [done] B1: Bengali rules-only schwa (final-cluster keep, ɔ→o raising, phalas, anusvara/visarga) @deps=T-0055,T-0056 @branch=feature/bengali-b1-rules @done=2026-10-04 @tags=scope-narrowed-see-T-0108
  - [x] (T-0060) [P2] [done] Audit user-supplied Bengali repositories and import pinned BanglaTLit test-only evaluation with provenance @done=2026-10-04 @branch=feature/bengali-b1-rules
  - [x] (T-0057) [P2] [done] Freeze Bengali normalized split manifests and cross-source exclusions before B2 training; assert provenance and report overlap @deps=T-0051 @branch=feature/bengali-b2-data @done=2026-10-04 @pr=117

  - [x] (T-0061) [P2] [done] B2 vowel alignment and three-class model: train on isolated Google data and gate on Bengali dev @deps=T-0052,T-0057 @branch=feature/bengali-b2-vowels @done=2026-10-04 @pr=118
  - [x] (T-0062) [P2] [done] B2 isolated Bengali spelling lexicon with style-aware lookup and measured coverage @deps=T-0057,T-0061 @branch=feature/bengali-b2-lexicon @done=2026-10-04 @pr=119
  - [x] (T-0063) [P2] [done] B2 Bengali candidate reranking: isolated character model, dev selection and held-out validation @deps=T-0057,T-0061,T-0062 @branch=feature/bengali-b2-rerank @done=2026-10-04 @tags=dev-rejected,heldout-not-run @pr=120
  - [x] (T-0064) [P3] [done] B2 native-conditioned candidate selection: train-only features and fixed dev gate after character-only rejection @deps=T-0063 @branch=feature/bengali-b2-native-selector @tags=dev-rejected,heldout-not-run @done=2026-10-04 @pr=121
  - [x] (T-0065) [P3] [done] B2 candidate expansion and out-of-fold selector training: preregister before further dev selection @deps=T-0064 @branch=feature/bengali-b2-crossfit @done=2026-10-04 @pr=122
  - [x] (T-0066) [P2] [done] B2 promote cross-fitted selector: native-aware runtime contract, parity, style fallbacks and external validation @deps=T-0065 @branch=feature/bengali-b2-runtime-rerank @rel=T-0054 @done=2026-10-05 @pr=123
  - [x] (T-0067) [P2] [done] B3 source-pinned Bengali lyrics pilot, provisional references and line-level evaluation @deps=T-0066 @branch=feature/bengali-b3-lyrics-seed @done=2026-10-05 @pr=124
  - [ ] (T-0068) [P2] [todo] B3 independent Bengali reference review: correct and attest the lyrics pilot before gold promotion @deps=T-0067
  - [x] (T-0069) [P2] [done] Expose Bengali in CLI with explicit language selection and Hindi-default compatibility @deps=T-0067 @branch=feature/bengali-cli @done=2026-10-05 @pr=125
  - [x] (T-0070) [P2] [done] Expose Bengali through npm/WASM and browser language selection with Hindi defaults @deps=T-0069 @branch=feature/bengali-web-npm @done=2026-10-05 @pr=126
  - [x] (T-0109) [P2] [done] Retrain after excluding variant spellings of held-out words (partitions schema 2) @tags=data @done=2026-10-09 @pr=127
  - [ ] (T-0117) [P2] [todo] Lyrics review worksheet: record each reviewer's স (s/sh), ছ (ch/chh) and o/a register preference, with several accepted spellings per word @tags=quality-review,eval
  - [ ] (T-0118) [P2] [todo] Bengali convention change: স -> s everywhere (retire s-cluster) and ছ -> ch; train-only ablation, deliberate pin update (dev +187 words measured) @tags=quality-review,linguistics
  - [ ] (T-0119) [P2] [todo] Bengali rule bundle: keep -িত/-ৃত/-ূত and final ত after গ র হ ন শ ষ; loan-coda delete table; C+ও; ya-phala after sibilants -> y; ba-phala after শ/স -> w, দ্ব -> db @tags=quality-review,linguistics
  - [ ] (T-0120) [P2] [todo] Bengali suffix-boundary deletion: drop stem-final inherent vowel before a train-derived suffix table (কে টি গুলো তে দের ...) @tags=quality-review,linguistics
  - [ ] (T-0121) [P2] [todo] Extend the Bengali vowel model to conjunct-bearing words (align halant clusters; one shared eligibility inventory for model, trainer and selector) @tags=quality-review,model
  - [ ] (T-0122) [P2] [todo] Retire the Bengali native selector (keep records); recommend the vowel model as the Bengali profile @tags=quality-review,cleanup
  - [ ] (T-0123) [P3] [todo] Shared schwa.delete.ccv for the Bengali rules-only default (dev +142, train +924 with 1,352 losses); decide after the rule bundle @tags=quality-review,linguistics
  - [ ] (T-0124) [P3] [todo] Optional Bengali lexicon filter: skip entries that differ from rule output only by o/a (470), measured on attested lyrics @tags=quality-review,lexicon
- [x] (F-0009) [P2] [done] npm distribution: @budhash/gomanize (WASM engine + JS/TS wrapper) @shadow @done=2026-10-09
  - [x] (T-0036) [P2] [done] Package the WASM engine as @budhash/gomanize with loader, types, make target, and Node smoke test @done=2026-09-08
- [ ] (F-0006) [P2] [todo] WASM build + web demo (data flywheel) @shadow
  - [x] (T-0028) [P2] [done] Add GOOS=js GOARCH=wasm build target; verify go:embed assets load @done=2026-09-05
  - [x] (T-0029) [P2] [done] Static web page: paste Devanagari, romanize, toggle flags, inline-edit output @done=2026-09-05
- [x] (F-0005) [P2] [done] Real-world evaluation (frequency-weighted + lyrics) @branch=feature/c-real-world-eval @done=2026-09-05
  - [x] (T-0019) [P2] [done] Frequency-weighted eval on Shabd (CC0) + lexicon token coverage @done=2026-09-04
  - [x] (T-0020) [P2] [done] Frequency-rank/expand lexicon toward 5-10k high-conf entries @done=2026-09-05
  - [x] (T-0021) [P3] [done] Build curated ~500-line lyrics gold set (Giitaayan+LyricsTranslate, fetch-script) @done=2026-09-05
  - [x] (T-0022) [P3] [done] Add COMI-LINGUA (CC-BY) redistributable sentence benchmark @done=2026-09-05
## Backlog

- [x] (F-0002) [P1] [done] H1: Fix the measurement (honest, multi-reference eval) @branch=feature/h1-honest-measurement @done=2026-09-05

  - [x] (T-0005) [P1] [done] Multi-reference eval: match-any-attested-variant (minCER) in benchmark @done=2026-09-04
  - [x] (T-0006) [P1] [done] Add CER metric alongside top-1 accuracy @done=2026-09-04
  - [x] (T-0007) [P1] [done] Integrate Aksharantar Hindi test set (AK-Freq/Uni/NE slices) @done=2026-09-05
  - [x] (T-0008) [P1] [done] Tighten CI gate to track PURE accuracy; report pure as headline @done=2026-09-04
  - [x] (T-0009) [P2] [done] Reconcile/prune override_hi.csv (fix कर्मकांड self-contradiction) @done=2026-09-04
- [x] (F-0003) [P1] [done] H2: Push the rule ceiling (~86%→~90% pure, honestly) @done=2026-09-04 @branch=feature/a-rule-hardening

  - [x] (T-0012) [P2] [done] Refactor: match rules on SOURCE chars, not BaseRom output strings @done=2026-09-04
  - [x] (T-0013) [P2] [done] Extract universal/script-scoped rules out of lang/hindi @done=2026-09-04
  - [x] (T-0014) [P1] [done] Add unit tests for lang/hindi, script/brahmic, cmd (currently 0) @done=2026-09-04
- [x] (F-0004) [P2] [done] H3: Break the ceiling with a learned component @branch=feature/h3-schwa-classifier @done=2026-09-05
  - [x] (T-0015) [P2] [done] Distill Arora et al. schwa classifier → dependency-free Go decision trees @deps=T-0007 @done=2026-09-04
  - [x] (T-0016) [P2] [done] Lexicon layer: attestation-weighted top-~50k, rules as OOV fallback @deps=T-0007 @done=2026-09-04
  - [x] (T-0017) [P3] [done] Offline model-mining tool for override candidates (IndicXlit/LLM) @done=2026-09-05
  - [x] (T-0018) [P3] [done] Candidate generation + tiny char n-gram re-ranker @done=2026-09-05

- [x] (T-0023) [P2] [done] Fix CLI whitespace handling: split words on all whitespace, not just spaces @done=2026-09-05

- [ ] (T-0024) [P3] [todo] Finish source-rune conversion: remaining BaseRom checks (gya/iya rules, conjunct neighbor tests)

- [ ] (T-0025) [P3] [todo] NFC-normalize engine input or dataset keys (7-373 non-NFC rows across datasets)

- [ ] (T-0026) [P3] [todo] Call-scoped debug traces: make TransliterateDebug/rule-toggling race-free

- [x] (T-0027) [P3] [done] Add unit tests for cmd/gomanize and scheme/colloquial (currently none) @done=2026-09-05

- [ ] (F-0006) [P2] [todo] WASM build + web demo (data flywheel)

  - [ ] (T-0030) [P3] [todo] Consent-based user-correction capture as review-gated lexicon candidates
- [ ] (F-0007) [P2] [todo] Aksharantar-convention scheme (selectable vowel-doubling)

  - [ ] (T-0031) [P2] [todo] Per-scheme symbol maps (interface change; unblocks IAST + convention schemes)
  - [ ] (T-0032) [P2] [todo] aksharantar scheme: vowel-doubling conventions; measure on AK test set
- [ ] (F-0008) [P3] [todo] Additional Brahmic languages (Marathi/Nepali)
  - [ ] (T-0033) [P3] [todo] Parameterize renderer inherent vowel (hardcoded 'a'); audit Devanagari rule literals @done=2026-10-09 @tags=inherent-vowel-delivered-by-T-0050,literal-audit-open
  - [ ] (T-0034) [P3] [todo] Implement lang/marathi (symbol map + config + rules composing brahmic.SchwaRules)

- [ ] (T-0035) [P3] [todo] Grow lyrics gold set toward ~500 lines (PD in-repo; copyrighted via fetch script)

- [ ] (F-0009) [P2] [todo] npm distribution: @budhash/gomanize (WASM engine + JS/TS wrapper)
  - [x] (T-0037) [P3] [done] CI: gated npm publish workflow on version tag (needs NPM_TOKEN) @done=2026-09-08

- [x] (F-0010) [P2] [done] Structured parser QA: construct-coverage + invariants (find systematic parsing bugs) @done=2026-09-11
  - [x] (T-0038) [P2] [done] Tier 1: per-construct accuracy analyzer (make target + benchmark; rank by prevalence x error) @done=2026-09-08
  - [x] (T-0039) [P2] [done] Tier 2: parser invariant/property tests (schwa-kept=>vowel; independent-vowel=>own unit; matra != independent; nuclei count) @done=2026-09-08
  - [x] (T-0043) [P2] [done] Corpus-diff transition matrix (rules-only; exact + minCER + variant-drift; stratified by dataset/split/construct/mode; lexical/acronym tagged) @done=2026-09-08
  - [x] (T-0044) [P3] [done] Held-out running-text regression set (50-150 curated lyrics/sentence tokens covering the target constructs) @done=2026-09-09
  - [x] (T-0040) [P2] [done] Tier 5 anchor + Phase B fix: consonant + independent-vowel parsing (गई class), construct-anchored golden, benchmark-gated @done=2026-09-08
  - [x] (T-0045) [P2] [done] Fix render.chandrabindu.final-silent to require IsWordFinal (drops nasal mid-word: चाँद→chaad); un-skip TestGoldenChandrabinduNasal @done=2026-09-08
  - [x] (T-0046) [P3] [done] Chandrabindu-before-labial: data-driven review of ँ→m vs n (साँप saanp/saamp); align with anusvara labial assimilation if warranted @done=2026-09-10
  - [x] (T-0047) [P3] [done] Add --schwa-model coverage to the chandrabindu golden set (optional; confirm nasal output holds under the learned schwa model) @done=2026-09-10
  - [x] (T-0048) [P3] [done] Held-out misses: schwa-retention gaps (अंततः→antatah, दरअस्ल→darasal, बाईं→bai) surfaced by TestBenchmarkHeldoutConstructs; measure candidate fix on Dakshina pure gate before changing rules @done=2026-09-11
  - [x] (T-0049) [P3] [done] Held-out misses: vowel-length aa boundary (फाँसी→faansi, तांगा→taanga) from TestBenchmarkHeldoutConstructs; confirm against DESIGN §3 no-length convention (likely won't-fix — document the decision) @done=2026-09-11
  - [x] (T-0041) [P3] [done] Tier 3: combinatorial construct enumeration + parse-well-formedness coverage test (CI) @done=2026-09-11
  - [x] (T-0042) [P3] [done] Tier 4: differential parse vs Unicode akshara (UAX #29) segmentation reference @done=2026-09-11

- [ ] (F-0011) [P2] [todo] Bengali (Bangla) support
  - [x] (T-0053) [P3] [done] B2: Bengali learned components (3-way schwa classifier, lexicon, reranker) @deps=T-0052,T-0057,T-0061,T-0062,T-0063,T-0064,T-0065,T-0066 @done=2026-10-05 @pr=123
  - [ ] (T-0054) [P3] [todo] B3: Bengali PD lyrics gold set (Tagore Gitabitan via Wikisource) + line-level suite @rel=T-0066 @deps=T-0067,T-0068 @tags=pilot-is-1913-gitanjali-T-0067
- [ ] (F-0012) [P2] [todo] Roman-to-Hindi/Bengali transliteration: ambiguity, candidate ranking, and independent evaluation
  - [ ] (T-0059) [P2] [todo] Define reverse API and candidate-ranking contract; audit bntranslit and Romabangla references before implementation

- [ ] (F-0013) [P2] [todo] Bengali keel follow-ups (docs/reviews/2026-10-07-bengali-keel.md) @tags=keel
  - [ ] (T-0071) [P2] [todo] K9: language registry (gomanize.Languages metadata) consumed by CLI, WASM, npm, d.ts and web @tags=keel,arch
  - [ ] (T-0072) [P2] [todo] K10: single option/flag registry; engine evaluates Rule.Conditional; fix --list-rules mentioning nonexistent --inherent-vowel-a @tags=keel,arch
  - [ ] (T-0073) [P2] [todo] K11: port Hindi reranker to NativeReranker and remove the legacy Reranker protocol from core @tags=keel,arch
  - [ ] (T-0074) [P2] [todo] K12: Bengali rerank and vowel-model per-word cost (reuse candidate engines, compute features once, cache normalized profile) @tags=keel,perf
  - [ ] (T-0075) [P2] [todo] K13: CI cost: run Hindi snapshot outside -race; stop re-running Go tests in Bengali make targets @tags=keel,ci
  - [ ] (T-0076) [P2] [todo] K14: move shared helpers out of rejected-experiment tools; reproduce-* targets for #120/#121; rename test-bengali-rerank @tags=keel,tooling
  - [ ] (T-0077) [P2] [todo] K15: one Python normalization module + Go/Python parity fixture (bengali_lyrics NFC-before-Cf) @tags=keel,tooling
  - [ ] (T-0078) [P2] [todo] K16: cross-fit folds keyed by collision_key (57 variant pairs across folds) @tags=keel,data
  - [ ] (T-0079) [P2] [todo] K17: decide runtime khanda-ta folding (ৎ vs ত্ vs ত্+ZWJ) for Bengali lexicon and selector features @tags=keel,unicode
  - [ ] (T-0080) [P2] [todo] K18: parser sets Nukta flag on the MultiChar path (Hindi decomposed nukta) @tags=keel,unicode
  - [ ] (T-0081) [P2] [todo] K19: assert remaining committed records (BanglaTLit external sections, external-overlap.json, manifest counts, reranker digest, Hindi artifact hashes, Hindi lyrics rerank CER) @tags=keel,tests
  - [ ] (T-0082) [P2] [todo] K20: Hindi schwa model and lexicon loaders fail closed (validate like Bengali) @tags=keel,robustness
  - [ ] (T-0083) [P2] [todo] K21: CLI --input/--test handle lines over 64 KB @tags=keel,cli
  - [ ] (T-0084) [P2] [todo] K22: construct analyzer asserts or is relabelled; per-construct heldout floors; Bengali construct/akshara/invariant suites @tags=keel,tests
  - [ ] (T-0085) [P2] [todo] K23: expose InherentVowelA in CLI/npm/WASM or document API-only; reject unknown npm option keys @tags=keel,api
  - [ ] (T-0086) [P3] [todo] K24: per-language registration so binaries need not embed every language's data @tags=keel,size
  - [ ] (T-0087) [P3] [todo] K25: naming: SchwaModel gates Bengali vowel model; renderer hard-codes roman onsets @tags=keel,naming
  - [ ] (T-0088) [P3] [todo] K26: dedupe test helpers (sha pins, assert_close, sha), drop single-implementation interfaces and 'new architecture' strings @tags=keel,cleanup
  - [ ] (T-0089) [P3] [todo] K27: bare U+09D7 leak; akshara corpus test skips on any load error; bare nuktas vanish; nukta on non-nukta consonant dropped @tags=keel,unicode
  - [ ] (T-0090) [P3] [todo] K28: npm load() caches the promise (concurrent load creates two runtimes) @tags=keel,npm
  - [ ] (T-0091) [P3] [todo] K29: CLI polish (WASM ready log to stdout, --input FILE form, --debug with --input, whitespace exit codes) @tags=keel,cli
  - [ ] (T-0092) [P3] [todo] Partial ScriptProfile should fail at engine construction, not first Translit @tags=review,api
  - [ ] (T-0093) [P3] [todo] Retire bengaliLexiconKey / vowelWordView now that core canonicalizes @tags=review,cleanup
  - [ ] (T-0094) [P3] [todo] B1 gate scope for later stages: bound paired losses, count strict losses, gate train and alternate style @tags=review,eval
  - [ ] (T-0095) [P3] [todo] Commit protocol before results for future dev-selection experiments; record attempt counts after a missed gate @tags=review,process
  - [ ] (T-0096) [P3] [todo] Commit train-only ablation evidence for B1 post-miss rules (ফ→f, loan-coda, h-nasal) @tags=review,eval
  - [ ] (T-0097) [P3] [todo] Bengali independent অ: train favors a (1,257 vs 888) but renders o; decide convention @tags=review,linguistics
  - [ ] (T-0098) [P3] [todo] Bengali gemination after anusvara coda (সংখ্যা -> shongkkha) @tags=review,linguistics
  - [ ] (T-0099) [P3] [todo] Malformed অা renders literally (অাবার -> oabar); consider folding to আ @tags=review,unicode
  - [ ] (T-0100) [P3] [todo] BanglaTLit scoring is case-sensitive (1,344 references uppercase); casefold or report both @tags=review,eval
  - [ ] (T-0101) [P3] [todo] Go-side hash pin of vowel_tree.json; selector parity fixture pins threshold and tie cases @tags=review,tests
  - [ ] (T-0102) [P3] [todo] Out-of-alphabet reranker candidates score -inf silently; count and report @tags=review,tooling
  - [ ] (T-0103) [P3] [todo] Lyrics pilot: ZWJ/ZWNJ extraction policy, letters-only CER for T-0068, trimmed captures before growth @tags=review,lyrics
  - [ ] (T-0104) [P3] [todo] npm: warn or error for Bengali rerank without schwaModel (CLI warns) @tags=review,npm
  - [ ] (T-0105) [P3] [todo] renderOnset type assertions before Gemination check (hot-path nit) @tags=review,perf
  - [ ] (T-0106) [P3] [todo] isSonorousRune reads only Runes[0] (inert today; fix before Bengali uses sonorous-final) @tags=review,unicode
  - [ ] (T-0107) [P3] [todo] Replace deprecated goreleaser brews config @tags=review,release
  - [ ] (T-0108) [P3] [todo] Bengali ɔ→o raising and broader anusvara/visarga coverage (narrowed from T-0052) @tags=review,linguistics
  - [x] (T-0110) [P2] [done] Keel review of the Bengali milestone (audit doc) @tags=keel @done=2026-10-09 @pr=128
  - [x] (T-0111) [P1] [done] Keel should-fix K2-K8 (format chars, long tokens, npm guard, style bypass, profiles, coverage assert, debug race) @tags=keel @done=2026-10-09 @pr=129
  - [x] (T-0112) [P1] [done] Keel K1: one canonical spelling per word across all components @tags=keel,unicode @done=2026-10-09 @pr=130
  - [ ] (T-0113) [P2] [todo] Doc sweep after the Bengali milestone (keel lens 3 checklist) @tags=keel,docs @branch=docs/bengali-milestone-sweep
  - [ ] (T-0114) [P2] [todo] Confirm redistribution terms for docs/reference/Hindi-Marathi-Nepali-Transliteration.pdf or replace it with a link @tags=licensing
  - [ ] (T-0115) [P3] [todo] Runtime/lyrics JSON records: rename 'accuracy' (it is the match-any rate) and fill strict_profiles @tags=records
  - [ ] (T-0116) [P2] [todo] Pin Hindi headline numbers in tests (curated pure/match-any/rerank, held-out, COMI, frequency-weighted, lyrics CER) per the PROCESS asserted-numbers rule @tags=tests
  - [ ] (T-0125) [P3] [todo] Table-driven Bengali rule helper for growing data tables (loan codas, suffixes, -িত endings) @tags=quality-review,arch
## Skipped

- [ ] (F-0003) [P1] [todo] H2: Push the rule ceiling (~86%→~90% pure, honestly) @shadow
  - [ ] (T-0010) [P1] [skipped] Medial ee/oo rule (ी→ee, ू→oo medial; i/u word-final) @deps=T-0005
  - [ ] (T-0011) [P2] [skipped] Narrow contextual ा→aa extension (never blanket) @deps=T-0005
---

# Notes

Use `## <ID>` headers (e.g., `## F-0001`) for structured notes per task.
Use `./tools/tasks.py show <id> --full` to display notes with task details.

- 2026-09-04: Initialized TASKS.md
