#!/bin/bash
# PreToolUse hook - prevents dangerous git operations
# This runs before Bash tool executions

# Read the tool input from stdin (JSON format from Claude Code)
INPUT=$(cat)

# Extract the command from the JSON input
# The input format is: {"tool_name": "Bash", "tool_input": {"command": "..."}}
COMMAND=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"command"[[:space:]]*:[[:space:]]*"//' | sed 's/"$//' || echo "")

# Check for direct commits to main/master
if echo "$COMMAND" | grep -qE 'git\s+commit.*' && ! echo "$COMMAND" | grep -qE 'git\s+checkout|git\s+switch'; then
    # Get current branch
    CURRENT_BRANCH=$(git branch --show-current 2>/dev/null)
    if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" ]]; then
        echo "BLOCKED: Cannot commit directly to $CURRENT_BRANCH branch."
        echo "Create a feature branch first: git checkout -b <branch-name>"
        exit 1
    fi
fi

# Check for force push
if echo "$COMMAND" | grep -qE 'git\s+push.*(-f|--force)'; then
    echo "BLOCKED: Force push detected. This is dangerous and requires explicit user approval."
    echo "If you really need to force push, ask the user first."
    exit 1
fi

# Check for force push to main/master
if echo "$COMMAND" | grep -qE 'git\s+push.*(main|master).*(-f|--force)' || \
   echo "$COMMAND" | grep -qE 'git\s+push.*(-f|--force).*(main|master)'; then
    echo "BLOCKED: Force push to main/master is extremely dangerous and not allowed."
    exit 1
fi

# Check for hard reset
if echo "$COMMAND" | grep -qE 'git\s+reset\s+--hard'; then
    echo "WARNING: Hard reset detected. This will discard uncommitted changes."
    echo "Proceeding, but be careful."
fi

# All checks passed
exit 0
