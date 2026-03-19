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
- [ ] A.10 Commit this phase using `osx-commit`

---

## Phase B: Config Package

**Goal**: Move `internal/config/` → `internal/infrastructure/config/`

- [ ] B.1 Verify `internal/infrastructure/config/` directory exists
- [ ] B.2 Move config files: `git mv internal/config/*.go internal/infrastructure/config/`
- [ ] B.3 Update package declarations to `package config`
- [ ] B.4 Update imports in consuming files: `internal/config` → `internal/infrastructure/config`
- [ ] B.5 Move test files
- [ ] B.6 Update test imports
- [ ] B.7 Verify: `go build ./internal/infrastructure/config/...`
- [ ] B.8 Verify: `go test ./internal/infrastructure/config/...`
- [ ] B.9 Commit this phase using `osx-commit`

---

## Phase C: Library Package

**Goal**: Move `internal/library/` → `internal/infrastructure/library/`

- [ ] C.1 Verify `internal/infrastructure/library/` directory exists
- [ ] C.2 Move library files: `git mv internal/library/*.go internal/infrastructure/library/`
- [ ] C.3 Update package declarations to `package library`
- [ ] C.4 Update imports in consuming files: `internal/library` → `internal/infrastructure/library`
  - Files: `internal/services/*.go`
- [ ] C.5 Move test files
- [ ] C.6 Update test imports
- [ ] C.7 Verify: `go build ./internal/infrastructure/library/...`
- [ ] C.8 Verify: `go test ./internal/infrastructure/library/...`
- [ ] C.9 Commit this phase using `osx-commit`

---

## Phase D: Parsing Package

**Goal**: Move parsing files from `internal/core/` → `internal/infrastructure/parsing/`

- [ ] D.1 Verify `internal/infrastructure/parsing/` directory exists
- [ ] D.2 Move parsing files
  - `git mv internal/core/loader.go internal/infrastructure/parsing/`
  - `git mv internal/core/parser.go internal/infrastructure/parsing/`
  - `git mv internal/core/platform_parser.go internal/infrastructure/parsing/`
  - `git mv internal/core/doc.go internal/infrastructure/parsing/`
  - `git mv internal/core/integration_test.go internal/infrastructure/parsing/`
- [ ] D.3 Update package declarations to `package parsing`
- [ ] D.4 Update imports in parsing files (now reference infrastructure/adapters)
- [ ] D.5 Update imports in consuming files: `internal/core` → `internal/infrastructure/parsing`
  - Files: `internal/services/*.go`, `cmd/**/*.go`
- [ ] D.6 Move test files: `loader_test.go`, `parser_test.go`, `platform_parser_test.go`
- [ ] D.7 Update test imports
- [ ] D.8 Verify: `go build ./internal/infrastructure/parsing/...`
- [ ] D.9 Verify: `go test ./internal/infrastructure/parsing/...`
- [ ] D.10 Commit this phase using `osx-commit`

---

## Phase E: Serialization Package

**Goal**: Move serialization files from `internal/core/` → `internal/infrastructure/serialization/`

- [ ] E.1 Verify `internal/infrastructure/serialization/` directory exists
- [ ] E.2 Move serialization files
  - `git mv internal/core/serializer.go internal/infrastructure/serialization/`
  - `git mv internal/core/template_funcs.go internal/infrastructure/serialization/`
- [ ] E.3 Update package declarations to `package serialization`
- [ ] E.4 Update imports in consuming files: `internal/core` → `internal/infrastructure/serialization`
- [ ] E.5 Move test files: `serializer_test.go`, `template_funcs_test.go`
- [ ] E.6 Update test imports
- [ ] E.7 Verify: `go build ./internal/infrastructure/serialization/...`
- [ ] E.8 Verify: `go test ./internal/infrastructure/serialization/...`
- [ ] E.9 Commit this phase using `osx-commit`

---

## Phase F: Service Rename

**Goal**: Rename `internal/services/` → `internal/service/`

- [ ] F.1 Rename directory: `git mv internal/services internal/service`
- [ ] F.2 Update package declarations in all files to `package service`
- [ ] F.3 Update imports in consuming files: `internal/services` → `internal/service`
  - Files: `cmd/**/*.go`, `internal/application/*.go`
- [ ] F.4 Verify: `go build ./internal/service/...`
- [ ] F.5 Verify: `go test ./internal/service/...`
- [ ] F.6 Commit this phase using `osx-commit`

---

## Phase G: Final Verification & Cleanup

- [ ] G.1 Remove empty `internal/core/` directory
- [ ] G.2 Remove empty `internal/adapters/` directory
- [ ] G.3 Remove empty `internal/config/` directory
- [ ] G.4 Remove empty `internal/library/` directory
- [ ] G.5 Verify: `go build ./...`
- [ ] G.6 Verify: `go test ./...`
- [ ] G.7 Verify: `mise run check`
- [ ] G.8 Commit this phase using `osx-commit`
