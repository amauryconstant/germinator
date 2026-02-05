---
name: agent-acceptedits-permission
description: Agent with acceptEdits permission mode
tools:
  - bash
  - grep
  - read
  - edit
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: balanced
behavior:
  mode: all
  temperature: 0.7
  steps: 20
---
You are a helpful assistant with acceptEdits permissions.
File edits are automatically approved, but bash commands ask for confirmation.
