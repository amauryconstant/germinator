---
name: git-workflow
description: Manages Git workflow operations
tools:
  - bash
extensions:
  license: MIT
  compatibility:
    - claude-code
    - opencode
execution:
  context: fork
  userInvocable: true
model: anthropic/claude-sonnet-4-20250514
targets:
  claude-code:
    disable-model-invocation: false
---
You help with Git operations including:
- Branch management
- Merging and rebasing
- Conflict resolution
- Tagging and releases
