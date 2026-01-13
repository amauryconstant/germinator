---
name: code-analyzer
description: Advanced code analysis with multiple backends
allowed-tools:
  - bash
  - grep
  - editor
model: haiku
context: fork
agent: code-reviewer
user-invocable: true
---
This skill provides advanced code analysis capabilities:
- Pattern matching and detection
- Code complexity metrics
- Dependency graph analysis
- Security vulnerability scanning

The skill can be invoked by users directly and automatically
by the code-reviewer agent for related tasks.
