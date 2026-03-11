---
name: agent-restrictive
description: Agent with restrictive permission policy
tools:
  - Bash
  - Grep
  - Read
  - Edit
disallowedTools:
  - Write
model: claude-sonnet-4-5-20250929
permissionMode: default
---
This agent uses restrictive permission policy.
All tool operations require confirmation before execution.
Use when you want to review every action before it's taken.
