## 1. Create cmd/config.go

- [ ] 1.1 Create NewConfigCommand constructor returning cobra.Command
- [ ] 1.2 Add config command to root command in cmd/root.go

## 2. Implement config init command

- [ ] 2.1 Create NewConfigInitCommand constructor
- [ ] 2.2 Add --output flag (default: config.GetConfigPath())
- [ ] 2.3 Add --force flag
- [ ] 2.4 Implement RunE: check file exists, write scaffolded config
- [ ] 2.5 Create parent directories if needed (os.MkdirAll)
- [ ] 2.6 Print success message with output path

## 3. Implement config validate command

- [ ] 3.1 Create NewConfigValidateCommand constructor
- [ ] 3.2 Add --output flag (default: config.GetConfigPath())
- [ ] 3.3 Implement RunE: use config.NewConfigManager().Load()
- [ ] 3.4 Return error if file not found, parse error, or validation error
- [ ] 3.5 Print success message if valid

## 4. Add scaffolded config template

- [ ] 4.1 Define scaffolded config content as constant string in cmd/config.go

## 5. Tests

- [ ] 5.1 Add unit tests for config init command
- [ ] 5.2 Add unit tests for config validate command
- [ ] 5.3 Run mise run check (lint, format, test, build)
