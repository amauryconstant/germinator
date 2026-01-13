---
allowed-tools:
  - bash
  - editor
context: fork
description: Run linting and formatting checks on code
disable-model-invocation: false
---
This command runs comprehensive code quality checks including:
- Go fmt verification
- Go vet analysis
- golangci-lint with standard rules

Usage: run-lint [options] <files...>
