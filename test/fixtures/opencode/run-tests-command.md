---
name: run-tests
description: Runs all project tests
tools:
  - bash
execution:
  context: fork
  subtask: false
model: anthropic/claude-sonnet-4-20250514
targets:
  claude-code:
    disable-model-invocation: false
---
Execute test suite with coverage reporting:
```bash
go test ./... -v -cover
```
