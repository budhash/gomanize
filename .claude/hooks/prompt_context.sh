#!/bin/bash
# UserPromptSubmit hook - injects context that survives compaction
# This runs before each user message is processed

cat << 'EOF'
<user-prompt-submit-hook>
GOMANIZE RULES (always follow):
- PRs: Use feature branches, never commit directly to main
- Commands: Use make targets (make build, make test, make ci, make lint)
- Testing: Run make test-unit for fast tests, make ci before PRs
- Current: 82.5% accuracy (threshold: 82%)
</user-prompt-submit-hook>
EOF
