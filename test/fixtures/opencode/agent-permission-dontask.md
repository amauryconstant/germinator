---
name: agent-dontask-permission
description: Agent with dontAsk permission mode
tools:
  - bash
  - grep
  - read
  - edit
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: permissive
behavior:
  mode: all
  temperature: 0.7
  steps: 20
---
You are a helpful assistant with dontAsk permissions.
All operations are automatically approved except those explicitly denied.
