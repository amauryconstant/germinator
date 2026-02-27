## 1. Foundation

- [ ] 1.1 Create `cmd/container.go` with `ServiceContainer` struct
- [ ] 1.2 Update `CommandConfig` to embed `*ServiceContainer` (requires cli-infrastructure)

## 2. Command Constructors

- [ ] 2.1 Create `NewValidateCommand(config *CommandConfig)` in `cmd/validate.go`
- [ ] 2.2 Create `NewAdaptCommand(config *CommandConfig)` in `cmd/adapt.go`
- [ ] 2.3 Create `NewCanonicalizeCommand(config *CommandConfig)` in `cmd/canonicalize.go`
- [ ] 2.4 Create `NewVersionCommand(config *CommandConfig)` in `cmd/version.go`
- [ ] 2.5 Remove `init()` functions and global command variables from all command files

## 3. Root Command

- [ ] 3.1 Create `NewRootCommand(config *CommandConfig)` in `cmd/root.go`
- [ ] 3.2 Remove global `rootCmd` variable
- [ ] 3.3 Remove `main()` function from `cmd/root.go`

## 4. Composition Root

- [ ] 4.1 Create `main.go` at project root
- [ ] 4.2 Wire ServiceContainer in main()
- [ ] 4.3 Wire CommandConfig with Services
- [ ] 4.4 Call NewRootCommand and execute

## 5. Verification

- [ ] 5.1 Run `mise run check` (lint, format, test, build)
- [ ] 5.2 Verify CLI behavior unchanged (adapt, validate, canonicalize, version commands work)
