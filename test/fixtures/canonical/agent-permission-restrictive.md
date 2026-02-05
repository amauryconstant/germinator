---
name: agent-restrictive
description: Agent with restrictive permission policy
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
  mode: primary
  temperature: 0.7
  steps: 20
---
This agent uses restrictive permission policy.
All tool operations require confirmation before execution.
Use when you want to review every action before it's taken.
