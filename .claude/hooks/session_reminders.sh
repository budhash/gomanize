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

5. CURRENT STATUS
   Native Hindi accuracy: ~49% (Target: 80%+)
   Key issues: MISSING_SCHWA, V_VS_W, MISSING_FINAL_A
   See Claude.md and testbed/ISSUES.md for details.

==========================================

EOF
