---
name: agent-plan-permission
description: Agent with plan permission mode
tools:
  - grep
  - read
disallowedTools:
  - bash
  - edit
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: analysis
behavior:
  mode: all
  temperature: 0.7
  steps: 20
---
You are a planning assistant that analyzes without editing or executing.
