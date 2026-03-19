## Phase A: Adapters Package

**Goal**: Move `internal/adapters/` → `internal/infrastructure/adapters/`

- [x] A.1 Verify `internal/infrastructure/adapters/` directory exists
- [x] A.2 Move adapter files
  - `git mv internal/adapters/adapter.go internal/infrastructure/adapters/`
  - `git mv internal/adapters/helpers.go internal/infrastructure/adapters/`
- [x] A.3 Move adapter subdirectories
  - `git mv internal/adapters/claude-code internal/infrastructure/adapters/`
  - `git mv internal/adapters/opencode internal/infrastructure/adapters/`
- [x] A.4 Update package declarations in all adapter files to `package adapters`
- [x] A.5 Update imports in consuming files: `internal/adapters` → `internal/infrastructure/adapters`
  - Files: `internal/core/platform_parser.go`, `internal/core/loader.go`
- [x] A.6 Move test files: `helpers_test.go`, `claude-code/*_test.go`, `opencode/*_test.go`
- [x] A.7 Update test imports
- [x] A.8 Verify: `go build ./internal/infrastructure/adapters/...`
- [x] A.9 Verify: `go test ./internal/infrastructure/adapters/...`
- [x] A.10 Commit this phase using `osx-commit`

---

## Phase B: Config Package

**Goal**: Move `internal/config/` → `internal/infrastructure/config/`

- [x] B.1 Verify `internal/infrastructure/config/` directory exists
- [x] B.2 Move config files: `git mv internal/config/*.go internal/infrastructure/config/`
- [x] B.3 Update package declarations to `package config`
- [x] B.4 Update imports in consuming files: `internal/config` → `internal/infrastructure/config`
- [x] B.5 Move test files
- [x] B.6 Update test imports
- [x] B.7 Verify: `go build ./internal/infrastructure/config/...`
- [x] B.8 Verify: `go test ./internal/infrastructure/config/...`
- [x] B.9 Commit this phase using `osx-commit`

---

## Phase C: Library Package

**Goal**: Move `internal/library/` → `internal/infrastructure/library/`

- [x] C.1 Verify `internal/infrastructure/library/` directory exists
- [x] C.2 Move library files: `git mv internal/library/*.go internal/infrastructure/library/`
- [x] C.3 Update package declarations to `package library`
- [x] C.4 Update imports in consuming files: `internal/library` → `internal/infrastructure/library`
  - Files: `internal/services/*.go`
- [x] C.5 Move test files
- [x] C.6 Update test imports
- [x] C.7 Verify: `go build ./internal/infrastructure/library/...`
- [x] C.8 Verify: `go test ./internal/infrastructure/library/...`
- [x] C.9 Commit this phase using `osx-commit`

---

## Phase D: Parsing Package

**Goal**: Move parsing files from `internal/core/` → `internal/infrastructure/parsing/`

- [x] D.1 Verify `internal/infrastructure/parsing/` directory exists
- [x] D.2 Move parsing files
  - `git mv internal/core/loader.go internal/infrastructure/parsing/`
  - `git mv internal/core/parser.go internal/infrastructure/parsing/`
  - `git mv internal/core/platform_parser.go internal/infrastructure/parsing/`
  - `git mv internal/core/doc.go internal/infrastructure/parsing/`
  - `git mv internal/core/integration_test.go internal/infrastructure/parsing/`
- [x] D.3 Update package declarations to `package parsing`
- [x] D.4 Update imports in parsing files (now reference infrastructure/adapters)
- [x] D.5 Update imports in consuming files: `internal/core` → `internal/infrastructure/parsing`
  - Files: `internal/services/*.go`, `cmd/**/*.go`
- [x] D.6 Move test files: `loader_test.go`, `parser_test.go`, `platform_parser_test.go`
- [x] D.7 Update test imports
- [x] D.8 Verify: `go build ./internal/infrastructure/parsing/...`
- [x] D.9 Verify: `go test ./internal/infrastructure/parsing/...`
- [x] D.10 Commit this phase using `osx-commit`

---

## Phase E: Serialization Package

**Goal**: Move serialization files from `internal/core/` → `internal/infrastructure/serialization/`

- [x] E.1 Verify `internal/infrastructure/serialization/` directory exists
- [x] E.2 Move serialization files
  - `git mv internal/core/serializer.go internal/infrastructure/serialization/`
  - `git mv internal/core/template_funcs.go internal/infrastructure/serialization/`
- [x] E.3 Update package declarations to `package serialization`
- [x] E.4 Update imports in consuming files: `internal/core` → `internal/infrastructure/serialization`
- [x] E.5 Move test files: `serializer_test.go`, `template_funcs_test.go`
- [x] E.6 Update test imports
- [x] E.7 Verify: `go build ./internal/infrastructure/serialization/...`
- [x] E.8 Verify: `go test ./internal/infrastructure/serialization/...`
- [x] E.9 Commit this phase using `osx-commit`

---

## Phase F: Service Rename

**Goal**: Rename `internal/services/` → `internal/service/`

- [x] F.1 Rename directory: `git mv internal/services internal/service`
- [x] F.2 Update package declarations in all files to `package service`
- [x] F.3 Update imports in consuming files: `internal/services` → `internal/service`
  - Files: `cmd/**/*.go`, `internal/application/*.go`
- [x] F.4 Verify: `go build ./internal/service/...`
- [x] F.5 Verify: `go test ./internal/service/...`
- [x] F.6 Commit this phase using `osx-commit`

---

## Phase G: Final Verification & Cleanup

- [x] G.1 Remove empty `internal/core/` directory
- [x] G.2 Remove empty `internal/adapters/` directory
- [x] G.3 Remove empty `internal/config/` directory
- [x] G.4 Remove empty `internal/library/` directory
- [x] G.5 Verify: `go build ./...`
- [x] G.6 Verify: `go test ./...`
- [x] G.7 Verify: `mise run check`
- [x] G.8 Commit this phase using `osx-commit`

(End of file - total 120 lines)
