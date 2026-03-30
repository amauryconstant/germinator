## 1. Create cmd/config.go

- [x] 1.1 Create NewConfigCommand constructor returning cobra.Command
- [x] 1.2 Add config command to root command in cmd/root.go

## 2. Implement config init command

- [x] 2.1 Create NewConfigInitCommand constructor
- [x] 2.2 Add --output flag (default: config.GetConfigPath())
- [x] 2.3 Add --force flag
- [x] 2.4 Implement RunE: check file exists, write scaffolded config
- [x] 2.5 Create parent directories if needed (os.MkdirAll)
- [x] 2.6 Print success message with output path

## 3. Implement config validate command

- [x] 3.1 Create NewConfigValidateCommand constructor
- [x] 3.2 Add --output flag (default: config.GetConfigPath())
- [x] 3.3 Implement RunE: use config.NewConfigManager().Load()
- [x] 3.4 Return error if file not found, parse error, or validation error
- [x] 3.5 Print success message if valid

## 4. Add scaffolded config template

- [x] 4.1 Define scaffolded config content as constant string in cmd/config.go

## 5. Tests

- [x] 5.1 Add unit tests for config init command
- [x] 5.2 Add unit tests for config validate command
- [x] 5.3 Run mise run check (lint, format, test, build)
