---
name: agent-bypasspermissions-permission
description: Agent with bypassPermissions permission mode
tools:
  - bash
  - grep
  - read
  - edit
  - write
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: unrestricted
behavior:
  mode: all
  temperature: 0.7
  steps: 20
---
You are a helpful assistant with bypassPermissions permissions.
All operations are automatically approved without any confirmation.
