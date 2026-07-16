---
description: Agent with balanced permission policy
mode: primary
model: anthropic/claude-sonnet-4-20250514
tools:
  bash: true
  grep: true
  read: true
  edit: true
  write: true
permission:
  edit:
    "*": allow
  bash:
    "*": ask
  read:
    "*": allow
  grep:
    "*": allow
  glob:
    "*": allow
  list:
    "*": allow
  webfetch:
    "*": allow
  websearch:
    "*": allow
temperature: 0.7
maxSteps: 20
---
This agent uses balanced permission policy.
File edits are automatically approved, bash commands ask for confirmation.
Good balance between automation and safety for development work.

