# Tasks: Initialize Project Structure

## Task List

### 1. Create Standard Go Directory Structure

**Objective**: Create all required directories following Standard Go Project Layout.

**Steps**:
1. Create cmd/ directory for CLI entry point
2. Create internal/ directory for private application code
3. Create internal/core/ for core interfaces and implementations
4. Create internal/services/ for business logic services
5. Create pkg/ directory for public library code
6. Create pkg/models/ for domain models
7. Create config/ directory for configuration files
8. Create config/schemas/ for JSON Schema files
9. Create config/templates/ for template files
10. Create config/adapters/ for platform adapter configurations
11. Create test/ directory for test artifacts
12. Create test/fixtures/ for test input documents
13. Create test/golden/ for expected output files
14. Create scripts/ directory for utility scripts

**Verification**:
- Run `tree -L 2 -d` (or `ls -R` if tree not available) to verify all directories exist
- All directories should be present with correct hierarchy

**Dependencies**: None

---

### 2. Initialize Go Module

**Objective**: Set up Go module with appropriate module name.

**Steps**:
1. Determine module name (awaiting user input on open question)
2. Run `go mod init <module-name>`
3. Verify go.mod was created with correct module path
4. Create .gitignore for Go projects (bin/, .vscode/, etc.)

**Verification**:
- `test -f go.mod` succeeds
- `cat go.mod` shows correct module name and Go version

**Dependencies**: Task 1 (directory structure must exist)

---

### 3. Add Cobra CLI Framework Dependency

**Objective**: Install Cobra framework for CLI command structure.

**Steps**:
1. Run `go get github.com/spf13/cobra@latest`
2. Run `go mod tidy` to update go.sum
3. Verify cobra is listed in go.mod

**Verification**:
- `grep cobra go.mod` succeeds
- `go list -m github.com/spf13/cobra` shows version
- `go build ./...` succeeds without errors

**Dependencies**: Task 2 (Go module must be initialized)

---

### 4. Create CLI Entry Point Structure

**Objective**: Set up basic Cobra root command in cmd/ directory.

**Steps**:
1. Create cmd/root.go with main() function
2. Initialize Cobra root command with name "germinator"
3. Set root command description from PURPOSE.md
4. Add basic --help flag (default from Cobra)

**Verification**:
- `test -f cmd/root.go` succeeds
- `go build ./cmd` succeeds
- `./germinator --help` displays help text

**Dependencies**: Task 3 (Cobra must be installed)

---

### 5. Create Package Placeholder Files

**Objective**: Create minimal placeholder files with package declarations for all packages.

**Steps**:
1. Create internal/core/doc.go with package documentation
2. Create internal/services/doc.go with package documentation
3. Create pkg/models/doc.go with package documentation

**Verification**:
- `go build ./internal/...` succeeds
- `go build ./pkg/...` succeeds
- All files have proper package declarations
- `grep -r "^package" internal/ pkg/` shows all packages

**Dependencies**: Task 1 (directories must exist)

---

### 6. Create Configuration Directory Placeholders

**Objective**: Create minimal placeholder files in config/ directories.

**Steps**:
1. Create config/.gitkeep to preserve empty config/ directory if needed

**Verification**:
- Directories exist from Task 1
- If .gitkeep created, it exists

**Dependencies**: Task 1 (directories must exist)

---

### 7. Create Test Directory Placeholders

**Objective**: Set up minimal test structure.

**Steps**:
1. Create test/.gitkeep to preserve empty test/ directory if needed

**Verification**:
- Directories exist from Task 1
- If .gitkeep created, it exists

**Dependencies**: Task 1 (directories must exist)

---

### 8. Create Scripts Directory Placeholder

**Objective**: Set up minimal scripts structure.

**Steps**:
1. Create scripts/.gitkeep to preserve empty scripts/ directory if needed

**Verification**:
- Directory exists from Task 1
- If .gitkeep created, it exists

**Dependencies**: Task 1 (directories must exist)

---

### 9. Create Project Documentation

**Objective**: Add minimal documentation for project structure.

**Steps**:
1. Create or update README.md with project overview
2. Document directory structure in README.md
3. Document build instructions in README.md

**Verification**:
- `test -f README.md` succeeds
- `grep "Directory Structure" README.md` succeeds
- `grep "Build" README.md` succeeds

**Dependencies**: Task 1 (structure must be in place)

---

### 10. Final Validation and Build Check

**Objective**: Verify entire project structure builds correctly.

**Steps**:
1. Run `go build ./...` to verify all packages compile
2. Run `go mod tidy` to ensure dependencies are clean
3. Run `go vet ./...` for static analysis
4. Verify all directories exist with expected structure

**Verification**:
- `go build ./...` exits with code 0
- `go mod tidy` succeeds
- `go vet ./...` exits with code 0
- `tree -L 2` (or equivalent) shows correct structure

**Dependencies**: All previous tasks

---

## Parallelizable Work

The following tasks can be executed in parallel after Task 1:
- Task 5 (Create Package Placeholder Files)
- Task 6 (Config Directory Placeholders)
- Task 7 (Test Directory Placeholders)
- Task 8 (Scripts Directory Placeholder)
- Task 9 (Project Documentation)

## Dependencies Graph

```
Task 1 (Directory Structure)
├── Task 2 (Go Module) ─── Task 3 (Cobra) ─── Task 4 (CLI Entry)
├── Task 5 (Package Placeholders)
├── Task 6 (Config Placeholders)
├── Task 7 (Test Placeholders)
├── Task 8 (Script Placeholders)
└── Task 9 (Project Documentation)

All tasks → Task 10 (Final Validation)
```

## Open Questions to Resolve

1. **Module Name**: What is the Go module name? (github.com/username/germinator, gitlab.com/group/germinator, etc.)
