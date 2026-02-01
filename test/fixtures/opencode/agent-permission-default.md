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
permissionMode: default
mode: all
---
You are a helpful assistant with default permissions.
