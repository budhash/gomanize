<!-- The Acceptance block below is a FLOOR — the MINIMUM, never the review. A green gate ≠ "reviewed":
     the substance is the adversarial read (independent recompute, try-to-break-it, flag what the boxes
     don't name). Ticking every box is necessary, not sufficient. -->

## Why
<!-- The problem / motivation — written BEFORE reviewers dispatch, so they don't reconstruct it from the diff. -->

## Approach
<!-- How this solves it; the key decisions. -->

## Scope
<!-- What is deliberately NOT included (guards silent scope-creep). -->

## Prereqs
<!-- Task IDs (T-####) that MUST be [done] before this merges — or "None". -->
None

## Verification
<!-- How you verified. For accuracy-affecting changes, show before/after on `make test-dakshina`. -->

## Acceptance
<!-- Ticked checkboxes act as a forcing function — an unchecked box blocks merge, so silent scope-trim is impossible. -->
- [ ] `make ci` green (fmt-check, lint, build, test-cover, benchmark)
- [ ] Accuracy not regressed — pure (no-override) number reported before/after for any rule/engine change
- [ ] New behavior covered by tests (+ a sanity-revert for any regression guard — prove it fails without the fix)
- [ ] `TASKS.md` updated via `./tools/tasks` (relevant task moved to `done`/linked to this PR)
- [ ] Every review finding handled — Fixed (commit ref) / Rejected (reason) / Deferred (task ID) — no silent discards

## Checklist
- [ ] No secrets, credentials, or API keys committed
- [ ] Docs updated if behavior, architecture, or a gotcha changed
