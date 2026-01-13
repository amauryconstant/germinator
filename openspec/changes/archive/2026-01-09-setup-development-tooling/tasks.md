# Tasks: Setup Development Tooling

## Task List

### - [x] 1. Configure golangci-lint

**Objective**: Set up golangci-lint configuration file with core Go linting rules.

**Steps**:
1. Create `.golangci.yml` in project root
2. Configure linters section with core linters:
   - Enable: gofmt, govet, errcheck
   - Disable other linters initially
3. Configure run section:
   - Set timeout (e.g., 5m)
4. Configure output section:
   - Set format to colored-line-number
5. Configure exclusions:
   - Exclude generated files (e.g., `*_gen.go`)
   - Exclude vendor directory

**Verification**:
- `test -f .golangci.yml` succeeds
- `golangci-lint run` succeeds without configuration errors

**Dependencies**: initialize-project-structure must be applied (go.mod must exist)

---

### - [x] 2. Create mise.toml Configuration

**Objective**: Set up mise task runner configuration.

**Steps**:
1. Create `mise.toml` in project root
2. Configure tools section:
   ```toml
   [tools]
   golangci-lint = "latest"
   ```
3. Configure tasks section:
   ```toml
   [tasks.validate]
   description = "Run all validation checks"
   run = [
     "go build ./...",
     "go mod tidy",
     "go vet ./...",
     "golangci-lint run",
   ]

   [tasks.smoke-test]
   description = "Quick build check"
   run = ["go build ./cmd"]
   sources = ["cmd/**/*.go"]
   outputs = ["germinator"]

   [tasks.format]
   description = "Format Go code"
   run = ["gofmt -w ./..."]
   sources = ["**/*.go"]
   ```
4. Configure task dependencies (if needed)

**Verification**:
- `test -f mise.toml` succeeds
- `mise run --help` lists all tasks
- `mise run validate` executes all checks

**Dependencies**: Task 1 (golangci-lint must be configured)

---

### - [x] 3. Create File-Based Tasks

**Objective**: Create executable scripts in `.mise/tasks/` directory.

**Steps**:
1. Create `.mise/tasks/` directory
2. Create `.mise/tasks/validate.sh`:
   - Add shebang: `#!/usr/bin/env bash`
   - Add `set -e` for error handling
   - Run checks: go build, go mod tidy, go vet, golangci-lint run
   - Make executable: `chmod +x .mise/tasks/validate.sh`
3. Create `.mise/tasks/smoke-test.sh`:
   - Add shebang: `#!/usr/bin/env bash`
   - Run: `go build ./cmd`
   - Make executable: `chmod +x .mise/tasks/smoke-test.sh`
4. Create `.mise/tasks/format.sh`:
   - Add shebang: `#!/usr/bin/env bash`
   - Run: `gofmt -w ./...`
   - Make executable: `chmod +x .mise/tasks/format.sh`

**Verification**:
- `test -d .mise/tasks/` succeeds
- `test -x .mise/tasks/validate.sh` succeeds
- `test -x .mise/tasks/smoke-test.sh` succeeds
- `test -x .mise/tasks/format.sh` succeeds
- File-based tasks execute correctly

**Dependencies**: Task 2 (mise.toml must exist)

---

### - [x] 4. Document mise Workflow

**Objective**: Add documentation for mise task runner usage.

**Steps**:
1. Add section to README.md about development tooling
2. Document mise task commands:
   - `mise run validate` - run all validation checks
   - `mise run smoke-test` - quick build check
   - `mise run format` - format Go code
   - `mise run --help` - discover all tasks
3. Document automatic tool installation:
   - mise automatically installs golangci-lint
   - Run `mise use golangci-lint@latest` to install
4. Document workflow: "Run `mise run validate` before committing"

**Verification**:
- `grep "mise run" README.md` succeeds
- `grep "golangci-lint" README.md` succeeds
- Documentation is clear and actionable

**Dependencies**: Tasks 1-3 (all tooling must be set up)

---

### - [x] 5. Final Validation

**Objective**: Verify all tooling works correctly together.

**Steps**:
1. Run `golangci-lint run` to verify linting works
2. Run `mise run validate` to verify validation task works
3. Run `mise run smoke-test` to verify smoke test works
4. Run `mise run --help` to verify task discovery works
5. Verify golangci-lint auto-installs: `mise use golangci-lint@latest`
6. Verify file-based tasks execute correctly

**Verification**:
- All commands run without errors
- mise lists all tasks
- Tools are installed automatically

**Dependencies**: All previous tasks

---

## Parallelizable Work

The following tasks can be executed in parallel after dependencies are met:
- Task 4 (Document mise workflow) - can be done anytime after Tasks 1-3

## Dependencies Graph

```
initialize-project-structure
├── Task 1 (golangci-lint) ─── Task 2 (mise.toml) ─── Task 3 (File-based tasks)
└── Task 4 (mise workflow docs) ─── Task 5 (Validation)
```

## Decisions Made

1. **Linter Strictness**: Core linters only (gofmt, govet, errcheck), add more incrementally
2. **Pre-commit Hooks**: Skip, use manual validation with `mise run validate` before commits
3. **Format Tool**: gofmt (standard Go formatter)
4. **Task Runner**: Use mise instead of bash scripts for unified task system
5. **File-Based Tasks**: Store scripts in `.mise/tasks/` directory for proper editor support
