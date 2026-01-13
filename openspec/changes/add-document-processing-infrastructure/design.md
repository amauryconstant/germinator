# Design: Document Processing Infrastructure

## Context

The current codebase provides document parsing and validation through `ParseDocument()` and document-level `Validate()` methods. To enable core workflows (validate, adapt), we need minimal infrastructure for serialization and transformation orchestration.

### Constraints
- Follow minimal approach from existing codebase
- No premature abstractions or interfaces
- Simple functions over patterns until needed
- Single-file implementations where possible
- Integrate with existing Validate() methods (no JSON Schema yet)

## Goals / Non-Goals

**Goals:**
- Enable validate and adapt CLI commands
- Provide template-based serialization to YAML + markdown
- Orchestrate transformation with simple linear workflow

**Non-Goals:**
- Document interface for polymorphic handling (not needed for MVP)
- Multiple serializer implementations (single Go template approach)
- External validation via JSON Schema (existing Validate() methods sufficient)
- Batch/directory processing (single document transformation for MVP)
- Platform adapter system (deferred until actual cross-platform need)
- Schema inspection command (deferred until clear user need)

## Decisions

### Decision 1: Simple Functions Over Interfaces

**What**: Use simple functions like `RenderDocument(doc, platform)` instead of interfaces like `type Serializer interface { Serialize(doc) }`.

**Why**:
- Single implementation needed (Go templates)
- No polymorphic use cases yet
- Matches existing codebase pattern (`ParseDocument`, `LoadDocument`)
- Simpler to understand and maintain
- Faster to implement

**Alternatives considered**:
- DocumentSerializer interface: Would add indirection without value
- TemplateEngine interface: Would abstract standard library (text/template)

**Trade-offs**:
- **Pro**: Minimal code, linear flow, easier to understand
- **Con**: Need to refactor if multiple serializer types emerge (unlikely soon)
- **Mitigation**: Defer interfaces until concrete use case exists

---

### Decision 2: Template-Based Serialization

**What**: Use Go's standard library `text/template` for rendering documents to YAML + markdown.

**Why**:
- Built into Go, no external dependencies
- Familiar to Go developers
- Simple and direct (no abstraction layer)
- Can handle YAML frontmatter + markdown body

**Alternatives considered**:
- Custom YAML marshaling: Would lose markdown body preservation
- External template engine: Adds dependency for no benefit
- Struct tags only: Insufficient for complex formatting

**Trade-offs**:
- **Pro**: No dependencies, simple, direct
- **Con**: Need to learn template syntax (standard)
- **Mitigation**: Templates are simple for this use case (YAML field rendering)

**Template structure**:
```
config/templates/<platform>/<docType>.tmpl
```
Example: `config/templates/claude-code/agent.tmpl`
```yaml
---
name: {{.Name}}
description: {{.Description}}
tools:
{{- range .Tools}}
  - {{.}}
{{- end}}
---
{{.Content}}
```

---

### Decision 3: Linear Transformation Pipeline

**What**: Implement `TransformDocument(input, output, platform)` as a simple function that orchestrates linear workflow: LoadDocument → Validate → Serialize → Write.

**Why**:
- Matches existing linear pattern (`LoadDocument` calls `ParseDocument` calls `Validate`)
- No branching or conditional logic needed for MVP
- Clear, easy to understand

**Alternatives considered**:
- TransformationPipeline interface: Overkill for linear workflow
- Pipeline builder pattern: Adds complexity for no benefit
- Async processing: Not needed for single document

**Trade-offs**:
- **Pro**: Simple, linear, easy to understand
- **Con**: Need to refactor if parallelization needed (unlikely soon)
- **Mitigation**: Defer complexity until needed

---

### Decision 4: Direct Validation from CLI

**What**: Call document's `Validate()` method directly from CLI commands instead of creating a wrapper service.

**Why**:
- Existing `Validate()` methods already return clear, descriptive errors
- Wrapper service would add unnecessary layer of indirection
- No need for centralization in a simple CLI tool

**Alternatives considered**:
- ValidateDocument wrapper function: Would add no value over calling Validate() directly
- ValidationService interface: Would add indirection without value

**Trade-offs**:
- **Pro**: Minimal code, direct calls, no wrapper layer
- **Con**: Error handling may be duplicated across commands (minimal)
- **Mitigation**: Each command is simple, error handling is straightforward

---

### Decision 5: CLI Command Structure

**What**: Create separate command files (`cmd/validate.go`, `cmd/adapt.go`) using Cobra framework.

**Why**:
- Matches existing CLI framework from `initialize-project-structure`
- Cobra provides help, flags, and subcommand structure
- Separate files for modularity

**Alternatives considered**:
- Single file with all commands: Would be large and hard to maintain
- Different CLI framework: Inconsistent with existing setup

**Trade-offs**:
- **Pro**: Standard Cobra usage, modular
- **Con**: Multiple files (standard for CLI apps)
- **Mitigation**: Each file is simple and focused

---

### Decision 6: Template File Organization

**What**: Store templates in `config/templates/<platform>/<docType>.tmpl`.

**Why**:
- Clear directory structure
- Easy to add new platforms and document types
- User-editable (no recompilation needed)

**Alternatives considered**:
- Embedded templates (Go code): Harder to modify, not user-editable
- Single directory with naming conventions: More cluttered
- Database storage: Overkill for file-based tool

**Trade-offs**:
- **Pro**: User-editable, clear structure, extensible
- **Con**: More files to manage (standard for configuration)
- **Mitigation**: Only one platform (Claude Code) initially

**Directory structure**:
```
config/
  templates/
    claude-code/
      agent.tmpl
      command.tmpl
      skill.tmpl
      memory.tmpl
```

---

### Decision 7: Error Handling Strategy

**What**: Return errors from functions and fail fast with clear error messages. Use existing error messages from Validate() methods.

**Why**:
- Matches Go idioms (return errors, don't panic)
- Clear error messages for CLI users
- Fail fast prevents cascading issues
- Existing Validate() methods already provide good error messages

**Alternatives considered**:
- Custom error types: Unnecessary for simple CLI error display
- Warnings for non-critical issues: Adds complexity, not needed
- Panic on errors: Unidiomatic, hard to handle

**Trade-offs**:
- **Pro**: Idiomatic, clear, fail-fast
- **Con**: None (standard Go approach)
- **Mitigation**: None needed

---

### Decision 8: No Test Coverage Targets

**What**: Focus on integration tests and manual verification instead of code coverage percentages.

**Why**:
- Coverage targets incentivize testing implementation details
- Infrastructure quality better measured by integration tests
- Faster to implement without chasing arbitrary coverage metrics

**Alternatives considered**:
- >80% coverage targets: Would require testing implementation details
- 100% coverage: Unrealistic and time-consuming

**Trade-offs**:
- **Pro**: Faster implementation, focus on behavior
- **Con**: Less assurance of code path coverage
- **Mitigation**: Integration tests cover actual user workflows

---

## Risks / Trade-offs

### Risk 1: Template Syntax Complexity

**Risk**: Developers may need to learn Go template syntax for creating new platform templates.

**Mitigation**: Templates are simple for this use case (YAML field rendering). Provide examples and documentation.

**Acceptance**: Low risk, high benefit (flexibility, no dependencies).

---

## Migration Plan

### Phase 1: Template Engine (document-serialization)
1. Create `internal/core/serializer.go`
2. Implement `RenderDocument(doc, platform)` function
3. Create Claude Code templates for all document types
4. Write unit tests for rendering

### Phase 2: Transformation Pipeline (document-transformation)
1. Create `internal/services/transformer.go`
2. Implement `TransformDocument(input, output, platform)` function
3. Orchestrate LoadDocument → Validate → Serialize → Write
4. Write unit tests and integration tests for complete workflow

### Phase 3: CLI Commands (cli-framework modification)
1. Create `cmd/validate.go` with Validate command
2. Create `cmd/adapt.go` with Adapt command
3. Register subcommands in `cmd/root.go`
4. Write integration tests for CLI commands

### Rollback Plan
If any phase introduces issues:
- Revert to previous working state via git
- Phase isolation allows targeted rollback
- Minimal dependencies between phases

---

## Open Questions

None - design is minimal and clear based on constraints and requirements.
