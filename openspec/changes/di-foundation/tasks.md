## 1. Foundation

- [x] 1.1 Create `cmd/container.go` with `ServiceContainer` struct
- [x] 1.2 Update `CommandConfig` to embed `*ServiceContainer` (requires cli-infrastructure)

## 2. Command Constructors

- [x] 2.1 Create `NewValidateCommand(config *CommandConfig)` in `cmd/validate.go`
- [x] 2.2 Create `NewAdaptCommand(config *CommandConfig)` in `cmd/adapt.go`
- [x] 2.3 Create `NewCanonicalizeCommand(config *CommandConfig)` in `cmd/canonicalize.go`
- [x] 2.4 Create `NewVersionCommand(config *CommandConfig)` in `cmd/version.go`
- [x] 2.5 Remove `init()` functions and global command variables from all command files

## 3. Root Command

- [x] 3.1 Create `NewRootCommand(config *CommandConfig)` in `cmd/root.go`
- [x] 3.2 Remove global `rootCmd` variable
- [x] 3.3 Remove `main()` function from `cmd/root.go`

## 4. Composition Root

- [x] 4.1 Create `main.go` at project root
- [x] 4.2 Wire ServiceContainer in main()
- [x] 4.3 Wire CommandConfig with Services
- [x] 4.4 Call NewRootCommand and execute

## 5. Verification

- [x] 5.1 Run `mise run check` (lint, format, test, build)
- [x] 5.2 Verify CLI behavior unchanged (adapt, validate, canonicalize, version commands work)
