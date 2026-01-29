---
name: code-reviewer-mixed
description: Code reviewer with both allowed and disallowed tools
tools:
  - read
  - write
  - bash
disallowedTools:
  - dangerous-command
  - system-config
model: anthropic/claude-sonnet-4-20250514
permissionMode: dontAsk
---

You are a code reviewer with restricted tool access.
