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
## Backlog

- [ ] (F-0002) [P1] [todo] H1: Fix the measurement (honest, multi-reference eval) @branch=feature/h1-honest-measurement

  - [x] (T-0005) [P1] [done] Multi-reference eval: match-any-attested-variant (minCER) in benchmark @done=2026-09-04
  - [x] (T-0006) [P1] [done] Add CER metric alongside top-1 accuracy @done=2026-09-04
  - [ ] (T-0007) [P1] [deferred] Integrate Aksharantar Hindi test set (AK-Freq/Uni/NE slices)
  - [x] (T-0008) [P1] [done] Tighten CI gate to track PURE accuracy; report pure as headline @done=2026-09-04
  - [x] (T-0009) [P2] [done] Reconcile/prune override_hi.csv (fix कर्मकांड self-contradiction) @done=2026-09-04
- [x] (F-0003) [P1] [done] H2: Push the rule ceiling (~86%→~90% pure, honestly) @done=2026-09-04 @branch=feature/a-rule-hardening

  - [x] (T-0012) [P2] [done] Refactor: match rules on SOURCE chars, not BaseRom output strings @done=2026-09-04
  - [x] (T-0013) [P2] [done] Extract universal/script-scoped rules out of lang/hindi @done=2026-09-04
  - [x] (T-0014) [P1] [done] Add unit tests for lang/hindi, script/brahmic, cmd (currently 0) @done=2026-09-04
- [ ] (F-0004) [P2] [todo] H3: Break the ceiling with a learned component @branch=feature/h3-schwa-classifier
  - [x] (T-0015) [P2] [done] Distill Arora et al. schwa classifier → dependency-free Go decision trees @deps=T-0007 @done=2026-09-04
  - [x] (T-0016) [P2] [done] Lexicon layer: attestation-weighted top-~50k, rules as OOV fallback @deps=T-0007 @done=2026-09-04
  - [ ] (T-0017) [P3] [todo] Offline model-mining tool for override candidates (IndicXlit/LLM)
  - [ ] (T-0018) [P3] [todo] Candidate generation + tiny char n-gram re-ranker
## Skipped

- [ ] (F-0003) [P1] [todo] H2: Push the rule ceiling (~86%→~90% pure, honestly) @shadow
  - [ ] (T-0010) [P1] [skipped] Medial ee/oo rule (ी→ee, ू→oo medial; i/u word-final) @deps=T-0005
  - [ ] (T-0011) [P2] [skipped] Narrow contextual ा→aa extension (never blanket) @deps=T-0005
---

# Notes

Use `## <ID>` headers (e.g., `## F-0001`) for structured notes per task.
Use `./tools/tasks.py show <id> --full` to display notes with task details.

- 2026-09-04: Initialized TASKS.md
