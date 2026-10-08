# Keel review: Bengali milestone (2026-10-07)

**Scope:** whole system at `main` 2273008, after the Bengali stack (#111–#126) and
the variant-spelling retrain (#127). Read-only, multi-lens audit following the
keel process (milestone gate: what per-PR review is structurally blind to).
**Result:** no blockers. 8 should-fix items, 15 within-next items, 6 later items,
and a documentation sweep. Every finding below was reproduced or verified in code.

## Lenses

| Lens | Focus | Hunt classes |
|---|---|---|
| 1 Architecture | design symmetry, centralizers, third-language readiness, cost | H1 bypassed centralizer · H2 symmetry · H3 next-layer readiness · H4 dead code/naming · H5 cost |
| 2 Correctness | assertion strength under the shipped CI config, Unicode, fallbacks, concurrency, fail-closed guards | C1 computed-not-asserted · C2 canonical equivalence · C3 silent fallbacks · C4 concurrency · C5 fail-closed |
| 3 Docs and data | docs↔code, number integrity, tasks, isolation as documented vs enforced, licensing, paths | D1–D6 |
| 4 Reproduction | adversarial reproduction through real frontends, edge inputs, mutation testing, packaging | A–E |

**Deviation:** the keel process calls for a Codex adversarial pass. Codex could not
run (its configured model is not available for the account in use), so lens 4 was a
Claude subagent under the same reproduction mandate. Re-run with Codex when available.

### Class results

| Class | Result | Evidence (short) |
|---|---|---|
| H1 bypassed centralizer | found | No core canonicalization; 4 Go and 4 Python normalizers; style bypass duplicated |
| H2 symmetry | found | Hindi lacks Bengali-era guards (encoding, style bypass, artifact validation) |
| H3 next-layer readiness | found | Language list hard-coded in ~8 places; partial profiles default to Devanagari |
| H4 dead code / naming | found | Rejected-experiment code imported by shipped tools and run in CI |
| H5 cost | found | Learned paths quadratic in word length; race-instrumented snapshot 139 s |
| C1 computed-not-asserted | found (minor) | All pins bite under `make ci`; some committed records still unasserted |
| C2 canonical equivalence | found | Hindi diverges on precomposed nukta, Cf, mark order; Bengali only on mark order |
| C3 silent fallbacks | found (mostly documented) | Runtime artifact decode failures degrade silently but are caught by tests |
| C4 concurrency | found | `Options.Debug` and concurrent `SetOptions` race |
| C5 fail-closed | found (1 gap) | Training guard denied 22/22 bad states; cross-fit fold guard uses exact keys |
| D1 docs↔code | found | ~45 stale statements (CLAUDE.md, README, DESIGN, RESEARCH, ROADMAP, CHANGELOG, web) |
| D2 numbers | found (small) | Bengali numbers reproduce; Hindi lyrics rerank CER 0.0465 → 0.0476; in-sample framing |
| D3 tasks | found | Three task titles overclaim; ~10 deferrals lack task IDs |
| D4 isolation | found | Hindi lexicon coverage test documented as asserting, but only logs |
| D5 licensing | found | README License section, Homebrew formula, a third-party PDF |
| D6 paths/links | found (minor) | Review cache path in 6 docs; two dead path references |
| A isolation via frontends | clean | 3,000 interleaved WASM calls × 64 flag sets; CLI order-independent; web logic |
| B edge inputs | found | Quadratic time; Cf stripped from non-Indic text; CLI 64 KB line limit |
| C mutation testing | mostly clean | One analyzer described as a regression net cannot fail |
| D packaging | clean | npm contents, NOTICE in archives, offline WASM load |

## Punch-list

**Disposition (maintainer, 2026-10-07):** fix all 🟠 items now (K2–K8 in one PR;
K1 separately, with a deliberate, reported Hindi snapshot update); track every 🟡/🟢
item as a TASKS.md entry during the doc sweep; nothing is dropped.

### 🟠 Should fix (this milestone)

| ID | Finding | Evidence | Fix |
|---|---|---|---|
| K1 | Hindi output depends on Unicode encoding | Precomposed nukta (U+0958–095F): schwa model differs on 1,624/12,929 words, rerank 340, lexicon 63; Cf (ZWJ/ZWNJ/BOM) misses the lexicon (3,165/263,922); virama/nukta order changes default output on 427/839 Dakshina words (Bengali 5/5 synthetic) | One mandatory canonicalization step in the core engine (Cf policy, nukta form, combining-class order, split matras) from a per-script table, applied before lexicon, rerank and parse; retire per-language copies; encoding-variant goldens for both languages |
| K2 | Cf characters stripped from all text, including non-Indic | `👨‍👩‍👧` loses its ZWJs; flag tag sequences, soft hyphen and RLM removed, in every frontend | Remove Cf only inside Indic script runs (part of K1) |
| K3 | Learned paths quadratic in word length | Bengali vowel model + rerank: 1.7 s at 960 runes; WASM 31 s for a 2k-character word (web demo renders per keystroke); Hindi model 18.6 s at 40k runes | Decode once per word; bound learned-model work per word (fall back to rules beyond a length cap) |
| K4 | npm type guard bypassable by a getter or Proxy; the Go panic later kills the Node process | Options getter returns `true` then `7n` | Snapshot options into a validated plain object; check value types in Go instead of `Truthy()` |
| K5 | Style-option bypass is a duplicated allowlist | `lang/bengali/lexicon.go:102`, `reranker.go:148`; a new style field would fall through | Single core helper plus a reflect test that fails when `Options` gains a field |
| K6 | Partial `ScriptProfile` silently inherits Devanagari values | `script/brahmic/profile.go` `normalize()` | Require a complete profile |
| K7 | Hindi lexicon coverage test documented as asserting, but only logs | `benchmark/benchmark_test.go` (lexicon coverage); RESEARCH cites it as asserted | Fail on any held-out coverage |
| K8 | Concurrency contract is overstated | `Options.Debug` with concurrent `Translit` races (`EnableDebug` mutates the shared rule engine); concurrent `SetOptions` races | `Translit` ignores `Debug` (only `TranslitDebug` honours it); document or guard `SetOptions` |

### 🟡 Within next milestone (tracked as tasks)

| ID | Finding |
|---|---|
| K9 | Language registry hard-coded in ~8 places (Go switch, CLI, WASM, npm, `d.ts`, web) → `gomanize.Languages()` metadata consumed by every frontend |
| K10 | No single option registry; `Rule.Conditional` is display-only; `--list-rules` mentions a nonexistent flag |
| K11 | Core legacy `Reranker` hard-codes Hindi's candidate strategy; two reranker protocols → port Hindi to `NativeReranker` |
| K12 | Bengali rerank ~68× default per word (fresh engine and catalog per candidate, O(R²) overrides) |
| K13 | CI cost: Hindi snapshot 139 s under `-race` (single goroutine); Bengali targets re-run Go tests already run |
| K14 | Rejected-experiment code imported by shipped tools and run in CI; `test-bengali-rerank` tests the rejected model, not the shipped one |
| K15 | Python normalization copies diverge (`bengali_lyrics.py` applies NFC before Cf removal) → one module plus Go/Python parity fixture |
| K16 | Cross-fit fold guard uses exact keys (57 variant pairs across folds) → fold by `collision_key` |
| K17 | Bengali runtime lexicon key does not fold khanda-ta encodings (`সৎ` vs `সত্‍`) → decide and document |
| K18 | Parser does not set the nukta flag on the MultiChar path (Hindi decomposed nukta) |
| K19 | Still unasserted: BanglaTLit sections of external records, `external-overlap.json`, manifest descriptive counts, reranker output digest, Hindi artifact content hashes, Hindi lyrics rerank CER |
| K20 | Hindi model and lexicon loaders fail open (schwa model returns "delete" on load failure) |
| K21 | CLI `--input`/`--test` abort on lines over 64 KB (stdin is fine) |
| K22 | Construct analyzer described as a regression net cannot fail; heldout "per-construct floors" exist only for chandrabindu; construct/akshara/invariant suites are Hindi-only |
| K23 | `InherentVowelA` unreachable from CLI/npm/WASM; unknown npm option keys silently ignored |

### 🟢 Later

| ID | Finding |
|---|---|
| K24 | All languages' data linked into every binary (~790 KB) → per-language registration at N languages |
| K25 | Naming: `SchwaModel` gates Bengali's vowel model; renderer hard-codes roman onset strings |
| K26 | Duplication: ~12 hand-rolled SHA pin sites, duplicated `assert_close`/`sha` helpers, single-implementation interfaces, "new architecture" strings, parser rebuilt per word |
| K27 | Bare U+09D7 leaks into output; akshara corpus test skips on any load error; bare nuktas vanish |
| K28 | Concurrent npm `load()` creates two runtimes (cache the promise) |
| K29 | CLI polish: WASM logs "ready" to stdout; `--input FILE` space form unsupported; `--debug` ignored with `--input`; whitespace-only argument vs stdin exit codes differ |

### Documentation sweep

Lens 3's file-by-file checklist (≈45 items) drives the doc sweep: CLAUDE.md
(tagline, status, architecture tree, commands), README (tagline, flag table,
License section and NOTICE link, strict/minCER alongside match-any), DESIGN §4–6,
RESEARCH (Bengali dataset rows, §4/§5 placement, stale numbers, in-sample framing,
dev reuse across selection rounds), ROADMAP, PROCESS, CHANGELOG `[Unreleased]`, web
About panel and README, RERANKER.md, training and benchmark data READMEs, the design
doc's status/B.6/chandrabindu/lyrics-plan as-built notes, Makefile `help`, CLI help
text, session hook status text, dead path references, leaked cache paths, Homebrew
licence, a root NOTICE, the third-party reference PDF's licence record, and TASKS.md
(retitle T-0050/T-0052/T-0054, close T-0033, epic statuses, task IDs for every
tracked deferral and K9–K29).

## Prevent recurrence

| Class | Why per-PR review missed it | Gap closure |
|---|---|---|
| H1/C2 encoding | Each PR added a correct local normalizer; duplication and Hindi gaps only show together | Canonicalization lives in core (K1); CLAUDE.md rule; encoding-variant goldens for every language |
| H2 symmetry | Bengali PRs were scoped "Hindi unchanged" and the frozen snapshot has no encoding variants | Keel checklist item: back-apply new guards to existing languages |
| C1/D2 assertions | Each PR pinned its own artifact; prose numbers and cross-record fields belong to no PR | Every headline number in README/RESEARCH must come from an asserting test |
| D1 docs | No PR owned cross-cutting docs | PR checklist names CLAUDE.md/README/DESIGN/CHANGELOG; doc sweep at each milestone |
| D3 tasks | "Tracked:" notes were accepted without task IDs | Deferrals must carry a TASKS ID |
| C4 concurrency | Races need `Debug` plus shared instances; no test combined them | Race test with mixed options on a shared engine |
| B/K3 scale | Per-PR benchmarks used dictionary-sized words | Long-input timing test with a ceiling |
| Codex gap | Tooling config, not process | Re-run the adversarial pass with Codex at the next keel |
