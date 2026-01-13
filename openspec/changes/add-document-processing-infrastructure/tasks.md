## 1. Template Engine Implementation

- [ ] 1.1 Create `internal/core/serializer.go` file
- [ ] 1.2 Implement `RenderDocument(doc interface{}, platform string) (string, error)` function
- [ ] 1.3 Implement template loading from `config/templates/<platform>/<docType>.tmpl`
- [ ] 1.4 Implement YAML frontmatter rendering with struct fields
- [ ] 1.5 Implement markdown body preservation
- [ ] 1.6 Create `config/templates/claude-code/agent.tmpl` template file
- [ ] 1.7 Create `config/templates/claude-code/command.tmpl` template file
- [ ] 1.8 Create `config/templates/claude-code/skill.tmpl` template file
- [ ] 1.9 Create `config/templates/claude-code/memory.tmpl` template file
- [ ] 1.10 Write unit tests for RenderDocument with all document types
- [ ] 1.11 Write unit tests for markdown body preservation

## 2. Transformation Pipeline Implementation

- [ ] 2.1 Create `internal/services/transformer.go` file
- [ ] 2.2 Implement `TransformDocument(inputPath, outputPath, platform string) error` function
- [ ] 2.3 Implement orchestration: LoadDocument → doc.Validate() → RenderDocument → Write
- [ ] 2.4 Implement output file writing
- [ ] 2.5 Implement fail-fast error handling (parse, validate, write errors)
- [ ] 2.6 Write unit tests for successful transformation (valid document)
- [ ] 2.7 Write unit tests for validation failure (invalid document)
- [ ] 2.8 Write unit tests for parse failure (invalid YAML)
- [ ] 2.9 Write unit tests for write error (read-only directory)
- [ ] 2.10 Write integration tests for complete workflow with all document types

## 3. CLI Commands Implementation

- [ ] 3.1 Create `cmd/validate.go` file
- [ ] 3.2 Implement validate command with Cobra
- [ ] 3.3 Implement error handling and exit codes
- [ ] 3.4 Implement error display (print each error on separate line)
- [ ] 3.5 Write unit tests for validate command with valid document
- [ ] 3.6 Write unit tests for validate command with invalid document
- [ ] 3.7 Write unit tests for validate command with missing file
- [ ] 3.8 Create `cmd/adapt.go` file
- [ ] 3.9 Implement adapt command with Cobra
- [ ] 3.10 Add --platform flag (required)
- [ ] 3.11 Implement error handling for validation and write failures
- [ ] 3.12 Write unit tests for adapt command with valid document
- [ ] 3.13 Write unit tests for adapt command with invalid document
- [ ] 3.14 Write unit tests for adapt command with read/write errors

## 4. CLI Integration and Registration

- [ ] 4.1 Modify `cmd/root.go` to register validate subcommand
- [ ] 4.2 Modify `cmd/root.go` to register adapt subcommand
- [ ] 4.3 Verify commands appear in `germinator --help` output
- [ ] 4.4 Test validate command help: `germinator validate --help`
- [ ] 4.5 Test adapt command help: `germinator adapt --help`

## 5. End-to-End Testing

- [ ] 5.1 Create test fixtures for all document types (valid and invalid)
- [ ] 5.2 Test validate command with valid agent document
- [ ] 5.3 Test validate command with invalid command document
- [ ] 5.4 Test adapt command with valid skill document to Claude Code
- [ ] 5.5 Test adapt command with invalid agent document
- [ ] 5.6 Test complete workflow: parse, validate, transform, serialize for all document types

## 6. Code Quality and Validation

- [ ] 6.1 Run `go build ./internal/core/...` - verify compilation
- [ ] 6.2 Run `go build ./internal/services/...` - verify compilation
- [ ] 6.3 Run `go build ./cmd/...` - verify compilation
- [ ] 6.4 Run `go test ./internal/core/... -v -cover`
- [ ] 6.5 Run `go test ./internal/services/... -v -cover`
- [ ] 6.6 Run `go test ./cmd/... -v -cover`
- [ ] 6.7 Run `go vet ./...` - verify code quality
- [ ] 6.8 Run `mise run validate` - run all validation checks
- [ ] 6.9 Run `mise run format` - ensure code is formatted
- [ ] 6.10 Fix any linting or formatting issues

## 7. Documentation

- [ ] 7.1 Update README.md with new CLI commands usage
- [ ] 7.2 Document template file format and examples
- [ ] 7.3 Add examples for validate and adapt commands
- [ ] 7.4 Update AGENTS.md with new infrastructure details
