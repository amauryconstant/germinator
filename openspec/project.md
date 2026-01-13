# Project Context

## Purpose
A configuration adapter that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. Uses Claude Code's document standard as the source format and adapts it for other platforms. Enables users to maintain one source of truth for their coding assistant setup and quickly switch platforms without rewriting their configuration.

## Tech Stack
- **Language**: Go 1.25.5
- **CLI Framework**: Cobra
- **Validation**: JSON Schema (xeipuuv/gojsonschema)
- **Task Runner**: mise
- **Linting**: golangci-lint (gofmt, govet, errcheck, staticcheck, misspell, revive, gosec)
- **CI/CD**: GitLab CI/CD

## Project Conventions

### Code Style
- Follow Go standard layout with `internal/` for private code and `pkg/` for public libraries
- Respect boundaries between internal and public packages
- Use `mise run <task>` for all task operations
- Run `mise exec -- <command>` for tools installed via mise
- Run `mise run check` before committing
- Run `go mod tidy` after dependency changes
- NO comments in code unless explicitly requested

### Architecture Patterns
- **No predefined directory structure**: CLI works with any input/output paths
- **Platform differences handled automatically**: Tool names, permissions, and conventions mapped via adapters
- **Source content preserved**: Only adapt/enrich for target platform, don't alter or discard original content
- **No forced compatibility**: If target platform doesn't support a feature, it's simply not supported
- **Clear separation of concerns**:
  - `cmd/`: CLI entry point (Cobra framework)
  - `internal/`: Private application code (core interfaces, services)
  - `pkg/`: Public library code (domain models)
  - `config/`: Configuration files (schemas, templates, adapters)

### Testing Strategy
- Table-driven tests for parsing and validation with comprehensive edge cases
- Golden master files in `test/golden/` for snapshot testing
- Integration tests for complete workflows
- Target coverage: >80% overall (Core >90%, Services >85%, CLI >75%)
- Use `mise run test` for running tests
- Use `mise run test:coverage` for coverage reports

### Git Workflow
- Pre-commit hooks enforce code quality (gofmt, go vet, golangci-lint, YAML/TOML/JSON validation)
- GitLab CI/CD pipeline runs on merge requests and pushes to main
- Pipeline stages: setup → lint → test → tag → mirror
- Auto-create Git tags for versions on main branch
- Mirror to GitHub requires GITHUB_ACCESS_TOKEN and GITHUB_REPO_URL

## Domain Context
The project operates in the AI coding assistant ecosystem. It understands:
- Document types: agents, commands, memory, skills
- Platform-specific conventions and mappings
- Template-based output rendering using Go templates
- JSON Schema validation for document structure
- Platform adapter configurations that map features between different AI coding assistant platforms

## Important Constraints
1. **No hardcoded directory layouts**: CLI must work with any input/output paths provided by user
2. **Platform-specific logic only in adapters**: Core parsing should remain platform-agnostic
3. **Preserve original content**: Only transform for compatibility, don't lose or alter source meaning
4. **Graceful degradation**: Unsupported features are omitted rather than forced
5. **All dependencies must be verified**: Never assume libraries are available; check project first
6. **Pre-commit hooks are blocking**: Commits will fail if any hook fails

## External Dependencies
- **xeipuuv/gojsonschema**: JSON Schema validation for documents
- **Cobra**: CLI framework for command structure and help generation
- **GitLab CI/CD**: Automated testing and deployment pipeline
- **mise**: Task runner for unified development workflow
- **GitHub**: Mirror repository (optional, requires GITHUB_ACCESS_TOKEN)
