---
name: agent-permissive
description: Agent with permissive permission policy
tools:
  - bash
  - grep
  - read
  - edit
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: permissive
behavior:
  mode: primary
  temperature: 0.7
  steps: 20
---
This agent uses permissive permission policy.
All operations are automatically approved except those explicitly denied.
Maximum automation for trusted environments.
