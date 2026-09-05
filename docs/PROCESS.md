# Development Process

How work is planned, tracked, and shipped in gomanize. Absorbed from the
`skein` process substrate and adapted for this Go repo.

## Principles

- **Feature branches only.** Never commit directly to `main` — pre-commit hooks
  (`no-commit-to-branch`) enforce this. One PR per logical change.
- **`make ci` before every PR.** It runs fmt-check, lint, build, coverage tests,
  and the accuracy benchmark. A green gate is the *floor*, not the review.
- **Zero tech debt forward.** A shortcut is either fixed in the same PR or tracked
  as an explicit task via `./tools/tasks new` with a rationale. "I'll clean it up
  later" without a tracked task is a process violation.
- **Docs of record.** Behavior/architecture changes update the relevant doc in the
  same PR. Deep-dive findings live in `docs/reviews/YYYY-MM-DD-*.md`.

## Task tracking (`TASKS.md`)

`TASKS.md` is the single source of truth for features and tasks. It is managed by
a vendored, zero-dependency CLI — **edit it only through the CLI**, never by hand
(manual edits can break parsing; `@done` is auto-stamped).

```bash
./tools/tasks tree              # See everything, grouped by feature
./tools/tasks next              # Next actionable task (respects deps)
./tools/tasks current           # What's in progress
./tools/tasks show F-2 --full   # A feature + its notes
./tools/tasks new feature "Title" --prio P1 --section Backlog
./tools/tasks new task "Title" --prio P1 --under F-2 --deps T-5
./tools/tasks start T-10        # → doing (enforces single work-in-progress)
./tools/tasks done  T-10        # → done (auto-stamps @done)
./tools/tasks defer T-10        # → deferred
./tools/tasks validate          # Sanity-check the file (run in CI)
```

Also available via `make tasks ARGS="next"`. Canonical source of the tool:
<https://github.com/budhash/tasks>.

**Schema:** `- [ ] (ID) [PRIO] [STATUS] Title @tags...`
Features are `F-####` (parents), tasks are `T-####` (children). Priorities
`P0..P3`. Status `[todo] [doing] [done] [deferred] [skipped]`. Durable design
context for a feature goes in the `# Notes` block, not in chat.

The founding roadmap (F-0001…F-0005: tooling → measurement → rules → learned
components → real-world validation) is complete as of 2026-09-05; new work gets
new features via `./tools/tasks new`. See
[`docs/reviews/2026-09-04-state-of-project-and-path-to-next-level.md`](reviews/2026-09-04-state-of-project-and-path-to-next-level.md)
for the reasoning behind it.

## Branch & PR flow

```bash
git checkout -b feature/my-change
# ... work, keeping TASKS.md updated via ./tools/tasks ...
make ci
git add -A && git commit -m "Description"
git push -u origin feature/my-change
gh pr create   # fill in .github/PULL_REQUEST_TEMPLATE.md
```

The PR template's Acceptance block is a **floor** — ticking the boxes is necessary,
not sufficient. The substance is the adversarial read: recompute the numbers
independently, try to break the change, and flag what the boxes don't name.

## Accuracy discipline

- Report **pure** (no-override) accuracy as the headline; overrides are a
  secondary line. Overrides are an exception lexicon, not engine skill.
- Any accuracy-affecting change must show before/after on `make test-dakshina`
  (curated) in the PR's Verification section.
- A regression guard needs a sanity-revert: prove the test fails without the fix.
