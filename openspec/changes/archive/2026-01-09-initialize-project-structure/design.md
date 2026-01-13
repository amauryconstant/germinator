# Design: Project Structure Initialization

## Overview

This design establishes the foundational project structure for germinator, a Go-based CLI tool that adapts Claude Code documents between platforms. The structure follows Go's Standard Project Layout conventions and provides scaffolding for the complete implementation pipeline.

## Architectural Decisions

### 1. Standard Go Project Layout

**Decision**: Adopt the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) as the organizational foundation.

**Rationale**:
- **Community Convention**: Widely recognized by Go developers, reducing onboarding friction
- **Discoverability**: Familiar structure makes code navigation intuitive
- **Tooling Support**: Many Go tools (gopls, golangci-lint) expect this layout
- **Documentation**: Extensive community documentation and examples available

**Alternatives Considered**:
- **Custom Layout**: Would require explaining conventions to all contributors
- **Minimal Layout** (all code in root/): Would not scale as project grows
- **Monorepo Structure**: Overkill for single-purpose CLI tool

### 2. Internal vs Public Packages

**Decision**: Use `internal/` for application code and `pkg/` for reusable libraries.

**Rationale**:

**internal/** directory:
- Go compiler enforces import restrictions (cannot be imported from external packages)
- Clear separation of private application logic
- Encourages clean boundaries between components
- Prevents external consumers from depending on implementation details

**pkg/** directory:
- Contains stable, reusable components
- Can be imported by external projects if needed
- Contains domain models (Document, Agent, Command, Memory, Skill)
- Provides API stability guarantees

**Package Organization**:
```
internal/
  core/           # Core interfaces: DocumentParser, SchemaValidator, TemplateEngine, ConfigLoader
  services/       # Business logic: ValidationService, TransformationService

pkg/
  models/         # Domain models: BaseDocument, Agent, Command, Memory, Skill
```

### 3. CLI Entry Point Structure

**Decision**: Use Cobra framework with cmd/ for CLI entry points.

**Rationale**:
- **Cobra Benefits**: Built-in flag parsing, subcommand routing, help generation
- **Industry Standard**: Used by major Go CLIs (kubectl, hugo, docker CLI)
- **Composable**: Easy to add subcommands (validate, adapt, schema)
- **Best Practices**: Provides patterns for CLI design (verbs, flags, persistent flags)

**cmd/** Structure:
```
cmd/
  root.go         # Main entry point, root command
  validate/       # Validate subcommand (future)
  adapt/          # Adapt subcommand (future)
  schema/         # Schema subcommand (future)
```

**Why Cobra over Alternatives**:
- **urfave/cli**: Simpler but less feature-rich, Cobra is more mature
- **flag package**: Too low-level, manual subcommand routing required
- **Kingpin**: Good alternative but Cobra has larger ecosystem

### 4. Configuration Organization

**Decision**: Separate configuration into schemas/, templates/, and adapters/ directories.

**Rationale**:

**config/schemas/**:
- JSON Schema files for document validation
- Version-controlled schema definitions
- Can be inspected with `dotai schema` command
- Enables schema evolution tracking

**config/templates/**:
- Go template files for output rendering
- Version-controlled output formats
- Template changes trackable in git

**config/adapters/**:
- Platform-specific transformation rules
- Adapter configuration files
- Enables new platform support via configuration

**Benefits of Separate Config**:
- Clear separation of concerns
- Configuration can be updated without code changes
- Easy to inspect and modify configuration
- Supports plugin-like extensibility

### 5. Test Organization

**Decision**: Separate test fixtures and golden files into test/ directory.

**Rationale**:

**test/fixtures/**:
- Test input documents (valid and invalid examples)
- Real-world document samples
- Edge case documents
- Organized by document type (agents/, commands/, memories/, skills/)

**test/golden/**:
- Expected output files for snapshot testing
- Known-good transformation outputs
- Reference outputs for regression testing
- Enables output comparison with `diff`

**Benefits**:
- Test data separate from test code
- Golden master pattern for regression testing
- Easy to add new test cases
- Clear test data organization

### 6. mise Task Runner Organization

**Decision**: mise task runner configuration and scripts directory at project root.

**Rationale**:
- .mise/ provides unified task system for validation, formatting, etc.
- mise handles automatic tool installation
- Easy to run from project root with `mise run <task>`
- Version-controlled task definitions and scripts
- Automation-friendly (CI/CD integration)
- File-based tasks get proper editor support

### 7. Go Module Naming Convention

**Decision**: Use standard Go module path based on VCS location.

**Pattern**: `go mod init <vcs-host>/<username>/<project-name>`

**Examples**:
- GitHub: `github.com/username/germinator`
- GitLab: `gitlab.com/group/germinator`
- Personal domain: `example.com/germinator`

**Open Question**: Module name not yet determined (awaiting user input).

### 8. Package Granularity

**Decision**: Coarse-grained packages with clear responsibilities.

**Package Sizes**:
- `pkg/models/`: All domain models together (shared Document interface)
- `internal/core/`: All core interfaces together (shared dependencies)
- `internal/services/`: All business logic services together

**Rationale**:
- **Avoid Over-Engineering**: Too many packages adds complexity
- **Clear Boundaries**: Each package has single responsibility
- **Maintainability**: Easier to understand relationships
- **Avoid Circular Dependencies**: Clear hierarchy prevents cycles

**When to Split**:
- Package grows beyond 500 lines
- Distinct subdomains emerge
- Testing becomes difficult

## Integration Points

### With Future Milestones

**Core Infrastructure Milestone**:
- `internal/core/` interfaces will be implemented
- `pkg/models/` will get concrete struct definitions
- Schemas will be added to `config/schemas/`

**Document Type Milestones (2-5)**:
- Templates added to `config/templates/`
- Tests use `test/fixtures/` and `test/golden/`
- Services extended in `internal/services/`

**CLI Integration Milestone**:
- Subcommands added to `cmd/`
- Root command features extended
- Help and documentation enhanced

## Trade-offs

### Simplicity vs Scalability

**Chosen Path**: Balanced approach with moderate structure upfront.

**Pros**:
- Room to grow without restructuring
- Not over-engineered for current scope
- Clear organization from start

**Cons**:
- More directories than minimal project
- Requires understanding of layout conventions

### Convention vs Customization

**Chosen Path**: Follow Go community conventions.

**Pros**:
- Immediate familiarity for Go developers
- Tooling support out-of-the-box
- Best practices baked in

**Cons**:
- Less flexibility for unconventional needs
- May have empty directories initially

## Risk Mitigation

### Risk: Structure Changes Required Later

**Mitigation**:
- Use standard layout to minimize future refactoring
- Keep packages coarse-grained initially
- Document rationale in design docs

### Risk: Over-Engineering

**Mitigation**:
- Start with placeholder files (minimal code)
- Defer detailed implementation to future milestones
- Focus on structure, not implementation

### Risk: Module Name Incompatibility

**Mitigation**:
- Document open question about module name
- Easy to change with `go mod edit -module=...`
- No code imports until Core Infrastructure milestone

## Success Metrics

1. **Build Success**: `go build ./...` succeeds
2. **Tooling Support**: golangci-lint, go vet work correctly
3. **Documentation**: Clear README explaining structure
4. **Developer Onboarding**: New developers can navigate codebase in <5 minutes
5. **Extensibility**: New features can be added without structural changes

## References

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Go Package Names](https://go.dev/blog/package-names)
- [Effective Go: Package Names](https://go.dev/doc/effective_go#package-names)
- [Go Workspace Layout](https://github.com/golang-standards/project-layout)
