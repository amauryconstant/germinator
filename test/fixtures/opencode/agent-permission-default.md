---
name: agent-default-permission
description: Agent with default permission mode
tools:
  - bash
  - grep
  - read
  - edit
disallowedTools:
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: restrictive
behavior:
  mode: all
  temperature: 0.7
  steps: 20
---
You are a helpful assistant with default permissions.
