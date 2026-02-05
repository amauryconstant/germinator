---
name: agent-balanced
description: Agent with balanced permission policy
tools:
  - bash
  - grep
  - read
  - edit
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: balanced
behavior:
  mode: primary
  temperature: 0.7
  steps: 20
---
This agent uses balanced permission policy.
File edits are automatically approved, bash commands ask for confirmation.
Good balance between automation and safety for development work.
