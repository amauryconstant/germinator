---
name: run-lint
description: Run linting and formatting checks on code
tools:
  - bash
execution:
  context: fork
  subtask: false
arguments:
  hint: "[options] <files...>"
model: anthropic/claude-sonnet-4-20250514
targets:
  claude-code:
    disable-model-invocation: false
---
This command runs comprehensive code quality checks including:
- Go fmt verification
- Go vet analysis
- golangci-lint with standard rules

Usage: run-lint [options] <files...>
