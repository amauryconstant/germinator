# Proposal: Setup Development Tooling

## Summary

Configure minimal development tooling for germinator CLI tool, including golangci-lint configuration for code quality and mise task runner for validation and tool installation.

## Motivation

Minimal development tooling provides immediate value while keeping the project lean:

- **Code Quality**: Automated linting catches common issues before review
- **Consistency**: All developers use the same tools and configurations
- **Unified Task System**: mise provides one command to run all tasks
- **Auto-Installation**: mise automatically installs tools like golangci-lint
- **Parallel Execution**: Tasks run in parallel by default
- **CI/CD Ready**: Tooling configurations can be reused in CI pipelines

## Proposed Change

**golangci-lint Configuration**:
- Create `.golangci.yml` with core Go linting rules
- Enable 3 core linters: gofmt, govet, errcheck
- Configure output formats and exclusions
- Exclude generated code from linting

**mise Task Runner Configuration**:
- Create `mise.toml` with task definitions
- Configure golangci-lint as tool (mise auto-installs)
- Define validation task that runs: go build, go mod tidy, go vet, golangci-lint
- Define smoke-test task for quick build check
- Define format task for code formatting

**File-Based Tasks**:
- Create `.mise/tasks/validate.sh` (executable bash script)
- Create `.mise/tasks/smoke-test.sh` (executable bash script)
- Store scripts with proper shebangs for editor support

**Workflow Documentation**:
- Document `mise run validate` for comprehensive checks
- Document `mise run smoke-test` for quick verification
- Document `mise run --help` for task discovery
- Document when to run `go mod tidy` (after dependency changes)

## Alternatives Considered

1. **Bash Scripts**: Could use bash scripts, but mise provides:
   - Unified task system with one command
   - Automatic tool installation
   - Parallel execution
   - Cross-platform consistency

2. **Use Different Linter**: Could use individual linters, but golangci-lint is:
   - Industry standard for Go projects
   - Supports multiple linters with single configuration
   - Fast and well-maintained

3. **Multiple Task Runners**: Could use make/npm scripts, but mise is:
   - More modern and powerful
   - Better cross-platform support
   - Has built-in tool installation

## Impact

**Positive Impacts**:
- Enforces code quality standards from the start
- Reduces code review time by catching issues early
- Provides consistent developer experience via unified task system
- Automatic tool installation removes setup burden
- Parallel task execution improves performance
- File-based tasks get proper editor support

**Neutral Impacts**:
- Adds minimal configuration files to repository
- Developers learn mise command syntax (well-documented)

**No Negative Impacts**

## Dependencies

Depends on `initialize-project-structure` change (project structure and Go module must exist).

## Success Criteria

1. golangci-lint runs successfully with no configuration errors
2. `mise run validate` executes all validation checks successfully
3. `mise run smoke-test` executes quick build check successfully
4. `mise run --help` lists all available tasks
5. golangci-lint is automatically installed by mise when needed
6. Documentation explains tooling usage and workflow

## Validation Plan

- Run `golangci-lint run` to verify configuration works
- Run `mise run validate` to verify validation task works
- Run `mise run smoke-test` to verify smoke test works
- Run `mise run --help` to verify tasks are discoverable
- Verify mise auto-installs golangci-lint: `mise use golangci-lint@latest`

## Decisions Made

1. **Linter Strictness**: Start with 3 core linters (gofmt, govet, errcheck), add more incrementally as codebase matures
2. **Pre-commit Hooks**: Skip entirely, use manual validation with `mise run validate` before commits
3. **Format Tool**: Use gofmt (standard Go formatter)
4. **Task Runner**: Use mise instead of bash scripts for unified task system
5. **File-Based Tasks**: Store scripts in `.mise/tasks/` directory for proper editor support

## Related Changes

This is Feature 2 in the Project Setup Milestone (docs/phase4/IMPLEMENTATION_PLAN.md:65-69), depends on Feature 1 (initialize-project-structure).

## Timeline Estimate

1-2 hours for configuration and testing.
