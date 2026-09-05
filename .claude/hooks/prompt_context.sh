#!/bin/bash
# UserPromptSubmit hook - injects context that survives compaction
# This runs before each user message is processed

cat << 'EOF'
<user-prompt-submit-hook>
GOMANIZE RULES (always follow):
- PRs: Use feature branches, never commit directly to main
- Commands: Use make targets (make build, make test, make ci, make lint)
- Testing: Run make test-unit for fast tests, make ci before PRs
- Tasks: track via ./tools/tasks (never hand-edit TASKS.md)
- Accuracy: report PURE + match-any; CI gates pure >= 85% (curated 86.1% pure, 92.8% match-any, 94.7% w/--rerank)
</user-prompt-submit-hook>
EOF
