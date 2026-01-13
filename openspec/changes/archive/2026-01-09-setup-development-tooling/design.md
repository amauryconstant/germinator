# Design: Development Tooling Setup

## Overview

This design establishes minimal development tooling configuration for germinator, including golangci-lint for code quality and mise task runner for unified task execution and tool installation.

## Architectural Decisions

### 1. golangci-lint as Primary Linter

**Decision**: Use golangci-lint as primary linting tool.

**Rationale**:
- **Performance**: Runs multiple linters in parallel, faster than running each individually
- **Comprehensive**: Supports 50+ linters with single configuration
- **Industry Standard**: Widely adopted in Go projects, familiar to developers
- **Easy Integration**: Simple to integrate into CI/CD and mise tasks
- **Active Maintenance**: Well-maintained project with regular updates
- **Configuration as Code**: Single `.golangci.yml` file versioned in repository

**Linter Selection (Initial Set)**:
- **gofmt**: Standard Go formatter (non-negotiable)
- **govet**: Go static analysis (official tool, catches real bugs)
- **errcheck**: Check for unchecked errors (production reliability)

**Deferred Linters** (add incrementally):
- **staticcheck**: Advanced static analysis (good but can be opinionated, higher false positives)
- **ineffassign**: Detect ineffectual assignments (useful but not critical for MVP)
- **gocyclo**: Cyclomatic complexity (can add later if needed)
- **dupl**: Code duplication (can add later if needed)

**Incremental Strategy**:
- Start with 3 core linters that catch 80% of real issues with 20% of friction
- Add linters progressively as codebase matures and team gains experience
- Use warning mode before error mode for new linters

**Alternatives Considered**:
- **Individual Linters**: Would require separate configuration, slower execution
- **revive**: Good alternative, but golangci-lint has broader community adoption
- **No Linter**: Would lead to inconsistent code quality, more code review burden

### 2. mise Task Runner

**Decision**: Use mise as task runner and tool installer.

**Rationale**:
- **Unified Task System**: All tasks defined in `.mise/config.toml`, one command: `mise run <task>`
- **Auto-Installation**: mise automatically installs tools like golangci-lint when needed
- **Parallel Execution**: Tasks run in parallel by default with dependency management
- **Incremental Builds**: Can skip tasks if source files haven't changed (via `sources`/outputs`)
- **Watch Mode**: Auto-rebuild on file changes
- **Cross-Platform**: Consistent behavior across macOS, Linux, Windows
- **File-Based Tasks**: Store scripts in `.mise/tasks/` directory for proper editor support
- **Extensive Tool Registry**: 1000+ tools including golangci-lint

**Task Configuration**:
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

**Tool Configuration**:
```toml
[tools]
golangci-lint = "latest"
```

**Why Not Bash Scripts**:
- mise provides unified task system (one command for all tasks)
- Automatic tool installation removes setup burden
- Parallel execution improves performance
- Incremental builds save time on subsequent runs
- File-based tasks get proper editor support with syntax highlighting
- Cross-platform consistency (bash scripts can have issues on Windows)

**Why Not Make/npm Scripts**:
- mise is more modern and powerful
- Better cross-platform support
- Built-in tool installation and management
- Watch mode and incremental builds out of the box
- Cleaner TOML configuration vs Makefile syntax

### 3. Manual Validation Workflow

**Decision**: Use manual validation before commits, no pre-commit hooks.

**Rationale**:
- **Lightweight**: No setup scripts, no maintenance burden
- **Transparent**: Developers see what's being checked
- **Fast**: No hook overhead during commits
- **CI Safety Net**: CI validates everything regardless of local workflow

**Workflow**:
- Developer makes changes
- Developer runs `mise run validate`
- Developer commits if validation passes
- CI runs validation on pull request

**Why Not Pre-commit Hooks**:
- Git hooks require setup scripts (not in git), high maintenance burden
- pre-commit framework is Python-based, overkill for Go project
- Go community prefers lightweight approach (Kubernetes, etcd, Prometheus use CI + manual)
- Adds complexity without significant benefit for MVP

### 4. gofmt as Standard Formatter

**Decision**: Use gofmt for code formatting.

**Rationale**:
- **Universal Compatibility**: Ships with Go, works everywhere
- **Industry Standard**: 99% of Go projects use it (Kubernetes, Docker, Prometheus)
- **Low Risk**: Never breaks code, no surprises
- **Zero Learning Curve**: Developers already know it
- **Avoids Style Debates**: gofumpt can create contentious formatting opinions

**Why Not gofumpt**:
- Stricter formatter can create team friction
- Adds opinions without clear consensus
- Easy to switch to gofumpt later if requested
- Conservative approach for MVP: use standard tool all Go developers expect

**Deferred Decision**:
- Reconsider gofumpt if team specifically requests stricter formatting
- Requires team consensus on style preferences
- Can evaluate with actual code examples later

### 5. File-Based Tasks in .mise/tasks/

**Decision**: Store task scripts in `.mise/tasks/` directory.

**Rationale**:
- **Editor Support**: Bash scripts get proper syntax highlighting and auto-completion
- **Version Control**: Scripts are versioned in repository
- **Flexibility**: Can use any language (bash, python, etc.)
- **Clarity**: Separate files are easier to maintain than inline scripts

**Task Structure**:
```
.mise/
├── config.toml
└── tasks/
    ├── validate.sh
    ├── smoke-test.sh
    └── format.sh
```

**Why Not Inline Tasks**:
- File-based tasks have better editor support
- Easier to debug and test individual scripts
- Can use shell linters (shellcheck) on files
- Better separation of concerns

### 6. Minimal Documentation

**Decision**: Keep documentation concise and focused.

**Rationale**:
- Go developers already know `go get` and `go mod tidy`
- Don't over-explain standard Go workflows
- Focus on tooling-specific information

**What to Document**:
- mise task runner commands (`mise run validate`, `mise run --help`)
- Automatic tool installation via mise
- Workflow: "Run `mise run validate` before committing"
- When to run `go mod tidy`: "After any dependency changes"

**What Not to Document**:
- Detailed workflows for adding/removing dependencies (standard Go commands)
- Pre-commit hooks (not using them)
- IDE setup (not providing it)
- Bash script details (using mise instead)

## Integration Points

### With Development Workflow

**Local Development**:
- Developer makes changes
- Runs `mise run validate`
- Commits if validation passes

**Pull Request**:
- CI runs `mise run validate`
- All checks must pass before merge

**Continuous Integration**:
- Same validation runs on main branch
- Ensures quality gates for production

### With initialize-project-structure

**Directory Structure**:
- `.mise/tasks/` directory created
- go.mod already initialized
- No structural changes needed

**Go Module**:
- Tooling operates on existing Go module
- No module changes required

**.mise/config.toml**:
- Already exists with golangci-lint configured
- Tasks section added to existing config

## Trade-offs

### Simplicity vs Features

**Chosen Path**: Minimal tooling, add features incrementally.

**Pros**:
- Low maintenance burden
- Quick setup and onboarding
- Clear and straightforward
- Easy to understand

**Cons**:
- Some features not available immediately
- Must add features later if requested

### Strictness vs Adoption

**Chosen Path**: Start moderate, tighten over time.

**Pros**:
- Lower initial friction
- Easier adoption
- Focus on high-value rules first

**Cons**:
- Some issues may slip through
- Requires discipline to tighten later

### Bash Scripts vs mise

**Chosen Path**: Use mise task runner.

**Pros**:
- Unified task system with one command
- Automatic tool installation
- Parallel execution
- Cross-platform consistency
- File-based tasks get editor support
- Better for team workflows

**Cons**:
- Developers need to learn mise commands (well-documented)
- Requires mise installation

## Risk Mitigation

### Risk: Tooling Too Minimal

**Mitigation**:
- 3 core linters catch 80% of real issues
- Easy to add more linters incrementally
- CI provides comprehensive validation
- Can expand tooling when pain points emerge

### Risk: Validation Too Slow

**Mitigation**:
- mise runs tasks in parallel by default
- golangci-lint runs linters in parallel
- Use incremental builds (sources/outputs) to skip unchanged tasks
- Modern CI and local machines run checks quickly
- Monitor performance, optimize if needed

### Risk: Tooling Incompatibility

**Mitigation**:
- Use stable, well-maintained tools
- mise handles cross-platform compatibility
- Pin tool versions in mise.toml if needed
- Document installation requirements
- Test tooling across platforms (Linux, macOS, Windows)

### Risk: mise Adoption

**Mitigation**:
- mise has excellent documentation
- Simple TOML configuration
- Task discovery via `mise run --help`
- Common use case (validation, formatting) is straightforward

## Success Metrics

1. **Linter Configuration**: `golangci-lint run` executes without errors
2. **mise Tasks**: `mise run validate` runs successfully
3. **Tool Installation**: mise automatically installs golangci-lint when needed
4. **Task Discovery**: `mise run --help` lists all available tasks
5. **File-Based Tasks**: Scripts in `.mise/tasks/` execute correctly
6. **Adoption**: Team uses `mise run validate` regularly
7. **Code Quality**: Linter catches real issues before code review

## References

- [golangci-lint Documentation](https://golangci-lint.run/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [mise Documentation](https://mise.jdx.dev/)
- [mise Tasks](https://mise.jdx.dev/tasks/)
- [mise Tools](https://mise.jdx.dev/dev-tools/)
- [Go Module Reference](https://go.dev/ref/mod)
- [Effective Go: Formatting](https://go.dev/doc/effective_go#formatting)
