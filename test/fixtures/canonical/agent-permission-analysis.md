---
name: agent-analysis
description: Agent with analysis permission policy
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
  mode: primary
  temperature: 0.3
  steps: 20
---
This agent uses analysis permission policy.
Read-only analysis mode for exploration and planning.
No file modifications or command execution allowed.
Perfect for code review, documentation, and research.
