---
name: code-review-tool-enhanced
description: Advanced code review with automated checks
tools:
  - read
  - grep
  - bash
model: anthropic/claude-sonnet-4-20250514
extensions:
  license: MIT
  compatibility:
    - claude-code
    - opencode
  metadata:
    version: "1.0.0"
    author: "Germinator Team"
    category: "code-quality"
  hooks:
    pre-review: "run-linters"
    post-review: "update-metrics"
execution:
  context: fork
  agent: code-reviewer
  userInvocable: true
---
You are an enhanced code review tool with automated checks.

## Features
- Automated linting and static analysis
- Security vulnerability detection
- Performance bottleneck identification
- Best practices enforcement

## Usage
Call this skill to review code changes before merging.
