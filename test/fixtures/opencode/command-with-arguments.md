---
name: format-code
description: Formats code using project standards
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
