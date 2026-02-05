---
name: code-analyzer
description: Advanced code analysis with multiple backends
tools:
  - bash
  - grep
  - read
extensions:
  license: MIT
  compatibility:
    - claude-code
    - opencode
  metadata:
    version: "1.0.0"
    author: "Germinator Team"
execution:
  context: fork
  agent: code-reviewer
  userInvocable: true
model: anthropic/claude-haiku-4-20250514
targets:
  claude-code:
    disable-model-invocation: false
---
This skill provides advanced code analysis capabilities:
- Pattern matching and detection
- Code complexity metrics
- Dependency graph analysis
- Security vulnerability scanning

The skill can be invoked by users directly and automatically
by code-reviewer agent for related tasks.
