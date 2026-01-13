# Proposal: Initialize Project Structure

## Summary

Establish the Go project structure, tooling, and foundational configuration for the germinator CLI tool. This creates the standard Go project layout with cmd/, internal/, pkg/, config/, test/, and scripts/ directories, initializes Go modules, sets up the Cobra CLI framework, and creates placeholder files for all packages.

## Motivation

The germinator CLI tool requires a proper Go project structure as the foundation for all subsequent implementation work. Following Go's standard project layout ensures consistency, makes the codebase approachable for Go developers, and provides the necessary scaffolding for:

- CLI command organization (cmd/)
- Internal application logic (internal/)
- Reusable library packages (pkg/)
- Configuration files (config/ schemas, templates, adapters)
- Test fixtures and golden files (test/)
- Task runner configuration (.mise/ for validation and tool management)

## Proposed Change

Create a complete Go project structure following the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) conventions:

```
germinator/
├── cmd/
│   └── root.go          # CLI entry point using Cobra
├── internal/
│   ├── core/            # Core interfaces and implementations
│   └── services/        # Business logic services
├── pkg/
│   └── models/          # Domain models (Document, Agent, Command, etc.)
├── config/
│   ├── schemas/         # JSON Schema files for validation
│   ├── templates/       # Template files for rendering output
│   └── adapters/        # Platform adapter configurations
├── test/
│   ├── fixtures/        # Test input documents
│   └── golden/          # Expected output files for comparison
└── .mise/               # mise task runner configuration
    ├── config.toml       # Task definitions and tools
    └── tasks/           # File-based task scripts
```

Initialize Go modules with:
- `go.mod` with appropriate module path
- `go.sum` populated after dependencies added

Set up Cobra CLI framework:
- Add `github.com/spf13/cobra` dependency
- Create basic root command structure

Create minimal placeholder files:
- doc.go files in each package with package documentation
- .gitkeep files in empty directories as needed

## Alternatives Considered

1. **Custom Directory Layout**: Could use a non-standard layout, but this would:
   - Reduce code discoverability for Go developers
   - Violate community conventions
   - Increase onboarding friction

2. **Flat Structure**: Could put all code in root directory, but this would:
   - Not scale as the codebase grows
   - Mix concerns (CLI, business logic, models)
   - Make testing and organization difficult

3. **Monolithic Packages**: Could combine all code into single packages, but this would:
   - Violate separation of concerns
   - Make parallel development difficult
   - Reduce code reusability

## Impact

**Positive Impacts**:
- Establishes clear project organization from day one
- Provides scaffolding for all subsequent development
- Follows Go community conventions
- Enables parallel development after core infrastructure

**Neutral Impacts**:
- No functional behavior changes (infrastructure only)
- No user-facing functionality yet

**No Negative Impacts**

## Dependencies

None. This is the foundational milestone with no dependencies on other work.

## Success Criteria

1. All standard Go directories exist with correct structure
2. Go modules initialized successfully (`go mod init`)
3. Cobra dependency added and root command scaffolding created
4. All packages have placeholder files with package declarations
5. Project builds successfully: `go build ./...`
6. Project structure matches Standard Go Project Layout conventions
7. Documentation exists explaining directory structure

## Validation Plan

- Verify all directories exist: `ls -la cmd/ internal/ pkg/ config/ test/ .mise/`
- Verify Go module: `test -f go.mod && cat go.mod`
- Verify Cobra dependency: `grep cobra go.mod`
- Verify build: `go build ./...`
- Verify package structure: `find . -name "*.go" -path "*/pkg/*" -o -path "*/internal/*"`
- Verify mise configuration: `test -f .mise/config.toml`

## Open Questions

1. **Module Name**: What should the Go module name be? (e.g., `github.com/username/germinator`, `gitlab.com/group/germinator`)

## Related Changes

This is the first feature in the Project Setup Milestone (docs/phase4/IMPLEMENTATION_PLAN.md:52-89).

## Timeline Estimate

1-2 hours for full setup and verification.
