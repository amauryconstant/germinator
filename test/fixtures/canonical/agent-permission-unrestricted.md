---
name: agent-unrestricted
description: Agent with unrestricted permission policy
tools:
  - bash
  - grep
  - read
  - edit
  - write
  - webfetch
  - websearch
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: unrestricted
behavior:
  mode: primary
  temperature: 0.7
  steps: 50
---
This agent uses unrestricted permission policy.
All operations are automatically approved without any confirmation.
Maximum automation for fully trusted environments.
Use with caution in production contexts.
