---
name: format-code
description: Formats code using project standards
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
Format the specified file $ARGUMENTS:

```bash
gofmt -w $ARGUMENTS
goimports -w $ARGUMENTS
```

Example usage:
```
/format-code main.go
/format-code internal/*.go
```
