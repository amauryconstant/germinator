---
name: agent-analysis
description: Agent with analysis permission policy
tools:
  - Grep
  - Read
disallowedTools:
  - Bash
  - Edit
  - Write
model: claude-sonnet-4-5-20250929
permissionMode: plan
---
This agent uses analysis permission policy.
Read-only analysis mode for exploration and planning.
No file modifications or command execution allowed.
Perfect for code review, documentation, and research.
