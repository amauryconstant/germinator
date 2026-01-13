# Proposal: Add Core Infrastructure

## Summary

Build foundational components for germinator CLI tool: document models, YAML parsing, struct validation, and file loading. This ultra-minimal MVP provides exactly what's needed to parse and validate documents without premature abstractions.

## Motivation

Core infrastructure provides the absolute minimum for document processing based on **verified Claude Code specifications**:

- **Document Models**: Simple structs matching actual Claude Code specs with YAML tags and Validate() methods
- **YAML Parsing**: Extract frontmatter and markdown body from files, handling Memory as markdown-only
- **Struct Validation**: Check required fields with appropriate validation rules (format, enum values)
- **File Loading**: Detect type, parse, and validate in one function with correct architecture

This MVP-focused approach defers all interfaces, abstractions, and patterns until they're actually needed. No Document interface, no Factory pattern, no separate validator - just straightforward functions that do the work.

## Proposed Change

**Feature1: Document Model Definitions**
- Define `Agent`, `Command`, `Memory`, `Skill` structs in single `models.go` file
- Each struct has all frontmatter fields with YAML tags
- Each struct implements `Validate() []error` method
- No BaseDocument embedding - each type owns its fields directly

**Feature2: YAML Parsing**
- Implement `ParseDocument(filepath string, docType string)` function
- Extract YAML frontmatter (between `---` delimiters)
- Parse YAML into appropriate struct based on docType
- Extract markdown body content

**Feature3: File Loading**
- Implement `LoadDocument(filepath string)` function
- Detect type from filename patterns (agent-*.md, command-*.md, etc.)
- Call `ParseDocument` to parse file
- Call `document.Validate()` to validate struct
- Return validated document or error

## Alternatives Considered

1. **Document interface**: Could provide polymorphism, but this would:
   - Be overkill for 4 concrete types with no polymorphic use cases
   - Add unnecessary abstraction layer
   - Defer until we need to treat documents polymorphically

2. **DocumentParser interface**: Could enable multiple parsers, but this would:
   - Add indirection for single YAML implementation
   - Be premature before we need alternate parsing strategies
   - Defer until we have multiple parser implementations

3. **DocumentValidator interface**: Could centralize validation, but this would:
   - Be completely redundant - each document already has Validate() method
   - Add wrapper around calling existing methods
   - Provide zero value for MVP

4. **DocumentFactory pattern**: Could follow design pattern, but this would:
   - Be overkill for simple detect→parse→validate workflow
   - Add unnecessary abstraction for linear process
   - Replace with simple `LoadDocument` function

5. **BaseDocument struct**: Could reduce duplication, but this would:
   - Add embedding complexity for minimal benefit (12 fields vs 1 struct + 4 embeddings)
   - Be clearer to have each document own its fields
   - Keep code simple and explicit

6. **JSON Schema validation**: Could be powerful, but this would:
   - Add external dependency and complexity
   - Require document → JSON conversion
   - Be overkill for simple struct field validation

## Impact

**Affected Specs**:
- Add new capability: document-models
- Remove obsolete capability: core-interfaces (interfaces removed)

**Affected Code**:
- `pkg/models/models.go` - Complete rewrite of all structs and validation
- `pkg/models/models_test.go` - Complete rewrite of all tests
- `test/fixtures/*.md` - Recreate all test fixtures
- `internal/core/parser.go` - Updates for Memory markdown-only handling
- `internal/core/loader.go` - Updates for Memory markdown and Command name-from-filename handling

**Affected Code**:
- Create `pkg/models/models.go` (single file with all 4 document types)
- Create `internal/core/parser.go` (ParseDocument function)
- Create `internal/core/loader.go` (LoadDocument function with inline type detection)

**Positive Impacts**:
- Ultra-minimal implementation (0 interfaces, 0 patterns)
- Faster development (0.75-1.0 days vs 1-1.5 days)
- Easier to understand and maintain
- Straightforward, linear code flow

**Neutral Impacts**:
- No polymorphic document handling (not needed for MVP)
- Slightly more code (no BaseDocument embedding)

**No Negative Impacts**

## Dependencies

Depends on `initialize-project-structure`, `setup-configuration-structure`, and `setup-development-tooling` (Project Setup Milestone complete).

## Success Criteria

1. All document models match actual Claude Code specifications exactly
2. ParseDocument correctly extracts YAML frontmatter and markdown body
3. LoadDocument correctly detects types, parses, and validates all document types
4. Unit tests achieve >90% coverage on models and core
5. Integration test passes: Load valid and invalid documents with correct Claude Code format
6. All code passes `mise run validate`

## Validation Plan

- Run `go build ./pkg/models/...` and `go build ./internal/core/...`
- Run `go test ./pkg/models/... -v -cover` (target >90%)
- Run `go test ./internal/core/... -v -cover` (target >90%)
- Integration test: LoadDocument from fixtures → verify type, parsing, and validation
- Run `mise run validate` for final quality check

## Related Changes

This is the Core Infrastructure Milestone (docs/phase4/IMPLEMENTATION_PLAN.md:93-138) - ultra-simplified to absolute minimum. Required before all document type milestones (Agent, Command, Memory, Skill).

## Timeline Estimate

0.75-1.0 days for implementation and testing.

## Open Questions

None - scope is clear and ultra-minimal.
