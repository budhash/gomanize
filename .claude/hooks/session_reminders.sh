#!/bin/bash
# Session start reminders for gomanize development
# This hook runs at the start of each Claude Code session

cat << 'EOF'

=== GOMANIZE DEVELOPMENT REMINDERS ===

1. FIRST TIME SETUP
   Run this once after cloning:

   make init           # Download deps + install pre-commit hooks

2. PR WORKFLOW
   Always use feature branches and PRs:

   git checkout -b feature/my-feature
   # make changes
   git add . && git commit -m "Description"
   git push -u origin feature/my-feature
   gh pr create --title "Title" --body "Description"

   NEVER commit directly to main branch!
   Pre-commit hooks will block this automatically.

3. USE MAKE TARGETS
   Always use Makefile targets:

   make init           # First-time setup (deps + hooks)
   make build          # Build the gomanize binary
   make test           # Run all tests
   make test-unit      # Run unit tests only (fast)
   make test-cover     # Run tests with coverage
   make lint           # Run golangci-lint
   make ci             # Full CI: fmt-check, lint, build, test
   make demo           # Demo with sample Hindi words

4. RUN make ci BEFORE PRs
   Before creating a PR, ALWAYS run:

   make ci

5. TASK TRACKING
   ./tools/tasks tree    # backlog (never hand-edit TASKS.md)
   ./tools/tasks next    # next actionable task

6. CURRENT STATUS (2026-10-09; v1.2.1 + experimental Bengali, keel-reviewed)
   Hindi curated Dakshina: 86.2% pure / 92.9% match-any / 94.8% with --rerank
   Hindi held-out: 70.7% (rerank) | Lyrics gold CER: 0.039 (--lexicon; human floor ~0.054)
   Bengali (--language=bengali, experimental): held-out match-any 56.5% rules /
   62.8% vowel model / 63.0% + selector; lyrics references unreviewed (T-0068)
   Keel follow-ups: F-0013. Run Python suites under 3.12 (hosted CI does).
   See CLAUDE.md for status and docs/reviews/ for all decision records.

==========================================

EOF
