#!/bin/bash
# UserPromptSubmit hook - injects context that survives compaction
# This runs before each user message is processed

cat << 'EOF'
<user-prompt-submit-hook>
GOMANIZE RULES (always follow):
- PRs: Use feature branches, never commit directly to main
- Commands: Use make targets (make build, make test, make ci, make lint)
- Testing: Run make test-unit for fast tests, make ci before PRs
- Goal: Improve Dakshina accuracy from ~49% to 80%+
</user-prompt-submit-hook>
EOF
