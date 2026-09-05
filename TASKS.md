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

- [ ] (F-0006) [P2] [todo] WASM build + web demo (data flywheel) @shadow
  - [x] (T-0028) [P2] [done] Add GOOS=js GOARCH=wasm build target; verify go:embed assets load @done=2026-09-05
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

  - [ ] (T-0029) [P2] [todo] Static web page: paste Devanagari, romanize, toggle flags, inline-edit output
  - [ ] (T-0030) [P3] [todo] Consent-based user-correction capture as review-gated lexicon candidates
- [ ] (F-0007) [P2] [todo] Aksharantar-convention scheme (selectable vowel-doubling)

  - [ ] (T-0031) [P2] [todo] Per-scheme symbol maps (interface change; unblocks IAST + convention schemes)
  - [ ] (T-0032) [P2] [todo] aksharantar scheme: vowel-doubling conventions; measure on AK test set
- [ ] (F-0008) [P3] [todo] Additional Brahmic languages (Marathi/Nepali)
  - [ ] (T-0033) [P3] [todo] Parameterize renderer inherent vowel (hardcoded 'a'); audit Devanagari rule literals
  - [ ] (T-0034) [P3] [todo] Implement lang/marathi (symbol map + config + rules composing brahmic.SchwaRules)

- [ ] (T-0035) [P3] [todo] Grow lyrics gold set toward ~500 lines (PD in-repo; copyrighted via fetch script)
## Skipped

- [ ] (F-0003) [P1] [todo] H2: Push the rule ceiling (~86%→~90% pure, honestly) @shadow
  - [ ] (T-0010) [P1] [skipped] Medial ee/oo rule (ी→ee, ू→oo medial; i/u word-final) @deps=T-0005
  - [ ] (T-0011) [P2] [skipped] Narrow contextual ा→aa extension (never blanket) @deps=T-0005
---

# Notes

Use `## <ID>` headers (e.g., `## F-0001`) for structured notes per task.
Use `./tools/tasks.py show <id> --full` to display notes with task details.

- 2026-09-04: Initialized TASKS.md
