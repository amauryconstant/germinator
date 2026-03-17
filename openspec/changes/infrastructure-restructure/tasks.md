## 1. Verify Infrastructure Directories

- [ ] 1.1 Verify `internal/infrastructure/` directory exists (already created)
- [ ] 1.2 Verify subdirectories exist: `adapters/`, `parsing/`, `serialization/`, `config/`, `library/`

## 2. Move Infrastructure Files

- [ ] 2.1 Move `internal/core/loader.go` to `internal/infrastructure/parsing/loader.go`
- [ ] 2.2 Move `internal/core/parser.go` to `internal/infrastructure/parsing/parser.go`
- [ ] 2.3 Move `internal/core/platform_parser.go` to `internal/infrastructure/parsing/platform_parser.go`
- [ ] 2.4 Move `internal/core/serializer.go` to `internal/infrastructure/serialization/serializer.go`
- [ ] 2.5 Move `internal/core/template_funcs.go` to `internal/infrastructure/serialization/template_funcs.go`
- [ ] 2.6 Move `internal/core/doc.go` to `internal/infrastructure/parsing/doc.go`
- [ ] 2.7 Move `internal/core/integration_test.go` to `internal/infrastructure/parsing/integration_test.go`
- [ ] 2.8 Move adapter root files: `internal/adapters/adapter.go`, `internal/adapters/helpers.go` to `internal/infrastructure/adapters/`
- [ ] 2.9 Move `internal/adapters/claude-code/` subdirectory to `internal/infrastructure/adapters/claude-code/`
- [ ] 2.10 Move `internal/adapters/opencode/` subdirectory to `internal/infrastructure/adapters/opencode/`
- [ ] 2.11 Move `internal/config/` directory to `internal/infrastructure/config/`
- [ ] 2.12 Move `internal/library/` directory to `internal/infrastructure/library/`

## 3. Rename Services Package

- [ ] 3.1 Rename `internal/services/` to `internal/service/`
- [ ] 3.2 Update package declarations in all service files

## 4. Update Infrastructure Import Paths

- [ ] 4.1 Update imports from `internal/core` → `internal/infrastructure/parsing` (loader, parser) or `internal/infrastructure/serialization` (serializer, template_funcs)
- [ ] 4.2 Update all imports from `internal/adapters` to `internal/infrastructure/adapters`
- [ ] 4.3 Update all imports from `internal/config` to `internal/infrastructure/config`
- [ ] 4.4 Update all imports from `internal/library` to `internal/infrastructure/library`
- [ ] 4.5 Update all imports from `internal/services` to `internal/service`
- [ ] 4.6 Verify compilation with `go build ./...`

## 5. Move Infrastructure Tests

- [ ] 5.1 Move `internal/core/loader_test.go`, `parser_test.go`, `platform_parser_test.go` to `internal/infrastructure/parsing/`
- [ ] 5.2 Move `internal/core/serializer_test.go` to `internal/infrastructure/serialization/`
- [ ] 5.3 Move `internal/adapters/helpers_test.go` to `internal/infrastructure/adapters/`
- [ ] 5.4 Move `internal/adapters/**/*_test.go` to `internal/infrastructure/adapters/` (preserve subdirectory structure)
- [ ] 5.5 Move `internal/config/*_test.go` to `internal/infrastructure/config/`
- [ ] 5.6 Move `internal/library/*_test.go` to `internal/infrastructure/library/`
- [ ] 5.7 Move `internal/services/*_test.go` to `internal/service/`
- [ ] 5.8 Update all test file import paths
- [ ] 5.9 Verify tests pass with `go test ./...`

## 6. Cleanup Old Directories

- [ ] 6.1 Remove empty `internal/core/` directory
- [ ] 6.2 Remove empty `internal/adapters/` directory
- [ ] 6.3 Remove empty `internal/config/` directory (if moved)
- [ ] 6.4 Remove empty `internal/library/` directory (if moved)
- [ ] 6.5 Remove empty `internal/services/` directory (after rename)
