# Tasks: Refactor Document Models to Match Claude Code Specs

## Task List

### 1. Define All Document Models

**Objective**: Create all 4 document types matching actual Claude Code specifications.

**Steps**:
- [x] Completely rewrite `pkg/models/models.go` with Claude Code-matching structs
- [x] Define Agent struct with name, description, tools, disallowedTools, model, permissionMode, skills
- [x] Define Command struct with allowed-tools, argument-hint, context, agent, description, model, disable-model-invocation (name from filename)
- [x] Define Memory struct with paths (optional), markdown content only
- [x] Define Skill struct with name, description, allowed-tools, model, context, agent, user-invocable
- [x] Implement Agent.Validate() method checking name, description, model, permissionMode values
- [x] Implement Command.Validate() method (no required fields, validation optional only if provided)
- [x] Implement Memory validation (paths parsing, optional only, markdown content)
- [x] Implement Skill.Validate() method checking name format and description max length
- [x] Add YAML struct tags to all frontmatter fields
- [x] Write table-driven unit tests for all 4 document types
- [x] Test YAML unmarshaling for each document type

**Verification**:
- `go test ./pkg/models/... -run TestDocumentModels -v` passes
- `grep "type.*struct" pkg/models/models.go` shows all 4 document types
- All Validate() methods return appropriate errors for Claude Code specs
- All tests achieve >90% coverage

**Dependencies**: None

### 2. Refactor All Document Models

**Objective**: Rewrite all 4 document models to match verified Claude Code specifications.

**Steps**:
- [x] Completely rewrite `pkg/models/models.go` with Claude Code-matching structs
- [x] Update `pkg/models/models_test.go` with new validation tests
- [x] Remove all old test fixtures that don't match Claude Code specs
- [x] Ensure no validation required for incorrect model fields

**Dependencies**: Task 1 (Define All Document Models)

### 3. Implement YAML Parsing and Memory Handling

**Objective**: Create ParseDocument function to extract YAML frontmatter and parse into structs, with special handling for Memory (markdown-only).

**Steps**:
- [x] Completely rewrite `pkg/models/models.go` with Claude Code-matching structs
- [x] Create `internal/core/parser.go` with ParseDocument function
- [x] Detect markdown frontmatter delimiter (---)
- [x] Extract YAML content between delimiters
- [x] Extract markdown body content
- [x] Parse YAML into map[string]interface{}
- [x] Switch on docType to unmarshal into appropriate struct:
  - "agent" → Agent struct
  - "command" → Command struct (name from filename)
  - "memory" → Load as markdown content only (no frontmatter struct)
  - "skill" → Skill struct
- [x] Return parsed document or error
- [x] Add godoc comments
- [x] Write unit tests for parsing valid and invalid files
- [x] Test delimiter detection and edge cases (missing delimiters, empty frontmatter)
- [x] Test Memory markdown-only handling (no frontmatter, no YAML parsing)

**Verification**:
- `go test ./internal/core/... -run TestParser -v` passes
- `grep "func ParseDocument" internal/core/parser.go` succeeds
- Parses YAML frontmatter correctly for agent, command, skill
- Handles Memory as markdown content (no YAML struct)
- Extracts markdown body correctly
- Handles missing delimiters gracefully
- Unmarshals to correct struct type based on docType

**Dependencies**: Task 1 (document models), Task 2 (yaml.v3 dependency)

### 4. Implement Document Loading with Filename and Memory Handling

**Objective**: Create LoadDocument function with type detection, parsing, and validation, handling Commands (name from filename) and Memory (markdown content).

**Steps**:
- [x] Create `internal/core/loader.go`
- [x] Implement LoadDocument(filepath string) function
- [x] Detect type from filename using regex patterns:
  - agent-*.md or *-agent.md → "agent"
  - command-*.md or *-command.md → "command"
  - memory-*.md or *-memory.md → "memory"
  - skill-*.md or *-skill.md → "skill"
- [x] Return error if no pattern matches
- [x] Call ParseDocument(filepath, detectedType) to parse file
- [x] Call document.Validate() on parsed document
- [x] Return validated document or error
- [x] Add godoc comments
- [x] Write unit tests for each document type
- [x] Test error cases (unrecognizable filename, validation failures)
- [x] Test Command name extraction from filename
- [x] Test Memory markdown-only handling (no frontmatter)

**Verification**:
- `go test ./internal/core/... -run TestLoader -v` passes
- `grep "func LoadDocument" internal/core/loader.go` succeeds
- Loads and validates Agent documents
- Loads and validates Command documents (extracts name from filename)
- Loads and validates Memory documents (markdown only)
- Loads and validates Skill documents
- Returns error for unrecognizable filenames
- Returns validation errors for invalid documents

**Dependencies**: Task 3 (parser), Task 1 (document models)

### 5. Create Test Fixtures

**Objective**: Create test fixtures matching Claude Code specifications for integration testing.

**Steps**:
- [x] Create test fixtures directory if not exists: `test/fixtures/`
- [x] Create valid Agent test fixture (agent-test.md) with name and description
- [x] Create valid Command test fixture (command-test.md) with optional fields and name from filename
- [x] Create valid Memory test fixture (memory-test.md) as markdown content
- [x] Create valid Skill test fixture (skill-test.md) with all optional fields
- [x] Create invalid Agent test fixture (missing description field)
- [x] Create invalid Command test fixture (invalid YAML)
- [x] Create unrecognizable filename fixture (my-document.md)

**Verification**:
- `test -f test/fixtures/agent-test.md` succeeds
- All valid fixtures exist
- All invalid fixtures exist
- Fixtures use Claude Code specification format

**Dependencies**: None

### 6. Integration Test: End-to-End Workflow

**Objective**: Test complete workflow: load document → parse → validate → verify.

**Steps**:
- [x] Write integration tests using LoadDocument function:
  - [x] Load valid Agent document and verify type
  - [x] Load valid Command document and verify type
  - [x] Load valid Memory document and verify type
  - [x] Load valid Skill document and verify type
  - [x] Load invalid documents and verify validation errors
  - [x] Load unrecognizable filename and verify error
  - [x] Verify all fields are parsed correctly
  - [x] Verify validation messages are clear
  - [x] Test error cases (missing file, invalid YAML)

**Verification**:
- `go test ./internal/core/... -run TestIntegration -v` passes
- End-to-end workflow works for all 4 document types
- Error handling works correctly
- All valid fixtures load successfully
- All invalid fixtures return appropriate errors

**Dependencies**: Tasks 4, 5 (loader, test fixtures)

### 7. Run Tests with Coverage

**Objective**: Verify all tests pass with >90% coverage.

**Steps**:
- [x] Run `go test ./pkg/models/... -v -cover`
- [x] Run `go test ./internal/core/... -v -cover`
- [x] Verify coverage >90% for pkg/models/
- [x] Verify coverage >90% for internal/core/
- [x] Add tests if coverage is below target

**Verification**:
- `go test ./pkg/models/... -coverprofile=coverage.out` succeeds
- `go tool cover -func=coverage.out | grep total` shows >90% coverage
- `go test ./internal/core/... -coverprofile=coverage.out` succeeds
- `go tool cover -func=coverage.out | grep total` shows >90% coverage

**Dependencies**: Tasks 1-6 (all implementation complete)

### 8. Final Validation and Code Quality

**Objective**: Run all validation checks to ensure quality.

**Steps**:
- [x] Run `go build ./pkg/models/...` to verify models compile
- [x] Run `go build ./internal/core/...` to verify core compiles
- [x] Run `go mod tidy` to clean dependencies
- [x] Run `go vet ./...` for static analysis
- [x] Run `golangci-lint run` for linting
- [x] Run `mise run validate` for all checks
- [x] Fix any issues found

**Verification**:
- All builds succeed
- `go vet ./...` exits with code 0
- `golangci-lint run` exits with code 0
- `mise run validate` exits with code 0

**Dependencies**: Task 7 (tests passing with >90% coverage)

## Parallelizable Work

The following tasks can be executed in parallel after dependencies are met:

**Task 1 and Task 2** can be parallel
**Task 5** can be parallel with Tasks 1-4 (test fixtures)
**Task 6** can be parallel with Tasks 1-5 (integration tests)
**Task 7** can be parallel with Tasks 1-6 (coverage)
**Task 8** can be parallel after all other tasks complete

## Dependencies Graph

```
Task 1 (Document Models) ──┐
                           ├── Task 2 (YAML Parser) ──┐
Task 1 (Document Models) ─────┤                    ├── Task 3 (Loader) ────┴── Task 6 (Integration)
                            │                          ├── Task 5 (Test Fixtures) ────┤
                            │                          ├── Task 7 (Coverage) ────┤
                            └── Task 8 (Final Validation) ──────────┘
```

## Open Questions

- Should we update the design.md file to reflect this ultra-minimal approach without interfaces?

## Decisions Made

1. **No Interfaces**: Removed Document, DocumentParser, DocumentValidator interfaces - premature and unnecessary for 4 concrete types with linear workflow

2. **No Factory**: Use simple LoadDocument function instead of Factory pattern - clearer and simpler

3. **No BaseDocument**: Each document type owns its fields directly - no embedding complexity for 4 types

4. **Single Models File**: All 4 document types in one models.go file - reduces duplication

5. **Inline Type Detection**: Patterns in LoadDocument function, not separate file - less complex

6. **Struct Validation Only**: No JSON Schema validation for MVP - simple field checks sufficient

7. **No Template Engine**: Deferred to future milestone - not needed for MVP

8. **No Platform Adapter**: Deferred to future milestone - not needed for MVP

## Timeline Estimate

1-2 days for implementation and testing (ultra-minimal, straightforward refactoring).

## Related Changes

This is **Core Infrastructure Milestone** (docs/phase4/IMPLEMENTATION_PLAN.md:93-138) - ultra-simplified to absolute minimum based on verified Claude Code specs. Required before all document type milestones (Agent, Command, Memory, Skill).
