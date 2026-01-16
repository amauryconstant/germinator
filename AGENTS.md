<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

---

# Project Overview

**What this tool is**: A configuration adapter that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses Claude Code's document standard as the source format and adapts it for other platforms.

**Use case**: Users who test different AI coding assistants regularly. The tool enables them to:
1. Maintain **one source of truth** for their coding assistant setup
2. Quickly **switch platforms** without rewriting their configuration
3. **Adapt** their setup to new projects easily

---

# Development Workflow

## Prerequisites
- Go 1.25.5 or later
- mise task runner (for unified development commands)

## Build Commands
```bash
# Build the CLI
go build -o germinator ./cmd

# Build all packages
go build ./...

# Run tests
go test ./...

# Verify with go vet
go vet ./...
```

## mise Task Runner
This project uses mise for unified development workflow:

```bash
# Build CLI to bin/germinator
mise run build

# Run all validation checks
mise run check

# Run linting
mise run lint

# Auto-fix linting issues
mise run lint:fix

# Format Go code
mise run format

# Run tests
mise run test

# Run tests with coverage
mise run test:coverage

# Clean build artifacts
mise run clean

# Version management
mise run version:patch   # Bump patch version
mise run version:minor   # Bump minor version
mise run version:major   # Bump major version

# Discover all tasks
mise tasks
```

**Important**: Run `mise run check` before committing. Run `go mod tidy` after any dependency changes.

For tools installed via mise, use `mise exec -- <command>` (e.g., `mise exec -- golangci-lint run`).

---

# Directory Structure

```
germinator/
 ├── cmd/                    # CLI entry point (Cobra framework)
 │   ├── root.go          # Main command registration
 │   ├── validate.go       # Validate document subcommand
 │   └── adapt.go          # Transform document subcommand
 ├── internal/              # Private application code
 │   ├── core/             # Core interfaces and implementations: DocumentParser, DocumentSerializer
 │   │   ├── parser.go       # Parse documents from files
 │   │   ├── loader.go       # Load and validate documents
 │   │   └── serializer.go  # Serialize documents to templates
 │   └── services/         # Business logic: ValidationService, TransformationService
 │       └── transformer.go   # Orchestrate document transformation pipeline
 ├── pkg/                   # Public library code
 │   └── models/           # Domain models: Document, Agent, Command, Memory, Skill
 ├── config/               # Configuration files
 │   ├── schemas/          # JSON Schema files for document validation
 │   ├── templates/        # Go template files for output rendering
 │   │   └── claude-code/ # Claude Code platform templates
 │   │       ├── agent.tmpl
 │   │       ├── command.tmpl
 │   │       ├── skill.tmpl
 │   │       └── memory.tmpl
 │   └── adapters/         # Platform adapter configurations
 ├── test/                 # Test artifacts
 │   ├── fixtures/         # Test input documents (valid/invalid examples)
 │   └── golden/           # Expected outputs for snapshot testing
 └── .mise/                # Task runner configuration
     ├── config.toml       # Task definitions and tool configurations
     └── tasks/            # File-based bash scripts for tasks
```

---

# Key Constraints & Architectural Principles

When making decisions, keep these constraints in mind:

1. **No predefined directory structure** - The CLI works with any input/output paths. Do not hardcode directory layouts.

2. **Platform differences handled automatically** - Tool names, permissions, and conventions are mapped appropriately via adapters. Do not force platform-specific logic in core parsing.

3. **Source content preserved** - Documents are only adapted/enriched for the target platform. Do not alter or discard original content unless it's truly incompatible.

4. **No forced compatibility** - If a target platform doesn't support a feature from the source, it's simply not supported. Do not try to make incompatible features work.

5. **Follow Go standard layout** - The project uses `internal/` for private code and `pkg/` for public libraries. Respect these boundaries.

6. **Use mise for all task operations** - Use `mise run <task>` instead of running scripts directly. Use `mise exec --` for tools installed via mise.

---

# CLI Usage Pattern

Users interact with the tool via a simple CLI interface:

```bash
cli action input_file target_platform [options]
```

For implementation milestones, follow the IMPLEMENTATION_PLAN.md for specific command structures (validate, adapt, schema).

---

# Release Management

## Creating a Release

**Important**: Version bumping is manual. Git tags are created automatically by CI when `internal/version/version.go` changes.

1. Bump version using mise tasks:
   ```bash
   mise run version:patch   # Bump patch version (0.2.0 → 0.2.1)
   mise run version:minor   # Bump minor version (0.2.0 → 0.3.0)
   mise run version:major   # Bump major version (0.2.0 → 1.0.0)
   ```

2. Validate release prerequisites:
   ```bash
   mise run release:validate
   ```
   This checks:
   - Git working directory is clean (no uncommitted changes)
   - Current branch is main
   - GoReleaser configuration is valid

3. Commit the version bump:
   ```bash
   git add .
   git commit -m "chore: bump version to 0.3.0"
   git push origin main
   ```

4. **GitLab CI will automatically create a Git tag and release**:
   - Tag stage detects the version change
   - Creates Git tag with format `v<VERSION>` (e.g., `v0.3.0`)
   - Tag triggers release stage automatically
   - Release builds cross-platform binaries:
     - Linux/macOS/Windows (amd64/arm64)
     - Archives (.tar.gz for Unix, .zip for Windows)
     - SHA256 checksums
     - SBOMs (SPDX format)
     - Auto-generated release notes

**Note**: The release job automatically runs `mise run release:validate` in its before_script, so validation failures will cause the release to fail early with clear error messages. Tag creation is idempotent - if a tag already exists, the tag stage will skip creation and continue the pipeline.

## Release Artifacts

Releases include:
- **germinator_0.3.0_linux_amd64.tar.gz**
- **germinator_0.3.0_linux_arm64.tar.gz**
- **germinator_0.3.0_darwin_amd64.tar.gz**
- **germinator_0.3.0_darwin_arm64.tar.gz**
- **germinator_0.3.0_windows_amd64.zip**
- **checksums.txt** (SHA256)
- **germinator_0.3.0_sbom.spdx.json** (Software Bill of Materials)

## Tool Management

Check for tool updates:
```bash
mise run tools:check
```

Update tool versions:
```bash
mise run tools:update
```

After updating tools:
1. Review changes: `git diff .mise/config.toml`
2. Install updated tools: `mise install --yes`
3. Commit and push

## CI Image Maintenance

Custom CI Docker image (`registry.gitlab.com/amoconst/germinator/ci:latest`) contains:
- Go 1.25.5
- mise (latest stable)
- golangci-lint
- GoReleaser
- Docker CLI

---

# Technology Stack

- **Language**: Go 1.25.5
- **CLI Framework**: Cobra
- **Validation**: JSON Schema (xeipuuv/gojsonschema)
- **Task Runner**: mise
- **Release Management**: GoReleaser
- **Linting**: golangci-lint (gofmt, govet, errcheck, typecheck, misspell, revive)
- **CI/CD**: GitLab CI with custom Docker image

---

# Pre-commit Hooks

The project uses pre-commit hooks to ensure code quality before commits. Install them after cloning:

```bash
pre-commit install
```

Hooks run automatically on commit and include:
- Go formatting (gofmt)
- Go vet
- golangci-lint checks
- YAML/TOML/JSON validation
- File hygiene checks (trailing whitespace, end-of-file, merge conflicts)

**Note**: All hooks are blocking - commits will fail if any hook fails.

---

# Troubleshooting

## CI/CD Issues

**Cache not invalidating after tool update:**
- Verify .mise/config.toml is included in cache key
- Check that setup job ran after changing tool versions
- Force cache rebuild by modifying any cache key file

**Mirror job not running:**
- Check that GITHUB_ACCESS_TOKEN and GITHUB_REPO_URL are set in GitLab CI variables
- Verify current branch is main
- Confirm job is skipped (not failed) when variables are missing - this is expected behavior

**Expensive jobs (lint, test, release) skipping unexpectedly:**
- Verify code changes are outside openspec/ directory
- Check pipeline logs for "Skipped due to rules" message
- This is expected behavior for documentation-only changes

## Release Issues

**release:validate failing:**
- Check for uncommitted changes: `git status`
- Verify you're on main branch: `git branch --show-current`
- Validate GoReleaser config: `goreleaser check`

**Release job failing in CI:**
- Check CI job logs for `mise run release:validate` output
- Look for specific validation failure messages
- Fix the issue locally, push the fix, and try again

**Tag not being created:**
- Verify `internal/version/version.go` was modified
- Check that changes are pushed to main branch
- Review tag stage logs for errors
- Ensure $GITLAB_USER_EMAIL and $GITLAB_USER_NAME variables are set
- Ensure $GITLAB_ACCESS_TOKEN is set with `write_repository` scope (required for tag stage)
- Ensure $GITLAB_RELEASE_TOKEN is set with `api` scope (required for release stage)
- If seeing 403 errors, verify tokens have correct permissions and are not expired

**Tag already exists, but release failed:**
- If tag exists but release failed, delete and re-push:
  ```bash
  git tag -d v0.3.0
  git push origin :refs/tags/v0.3.0
  ```
  Then re-run pipeline or push a new version bump
- Or manually create new tag with incremented version

**Need to recreate tag after fixing issues:**
- Delete existing tag locally: `git tag -d v0.3.0`
- Delete remote tag: `git push origin :refs/tags/v0.3.0`
- Re-run pipeline (tag stage will recreate tag)

**Version tag mismatch errors:**
- Ensure tag format is vX.Y.Z (with 'v' prefix)
- Check that version in internal/version/version.go is correct
- Tag stage automatically creates correct format, this should not happen

---

# CI/CD Pipeline

The project uses GitLab CI/CD for automated testing and deployment. The pipeline runs on:

- Merge requests
- Pushes to the main branch

**Pipeline stages:**
1. **build-ci** - Build and push CI Docker image (when Dockerfile.ci or .mise/config.toml changes)
2. **setup** - Download Go module dependencies
3. **lint** - Run golangci-lint
4. **test** - Run all tests
5. **tag** - Create Git tag when internal/version/version.go changes (automatic, idempotent)
6. **release** - Create GitLab releases on tag push (triggered by tag stage)
7. **mirror** - Push to GitHub mirror (main branch only, requires GITHUB_ACCESS_TOKEN and GITHUB_REPO_URL)

**Cache Configuration:**
- Cache key includes: .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum
- Cache invalidates when any of these files change
- Setup job writes cache (pull-push), other jobs read only (pull)
- Resource group `cache_updates` serializes writes to prevent corruption
- Artifacts expire after 24 hours

**Tag Stage Behavior:**
- Runs automatically when internal/version/version.go changes on main branch
- Extracts version from internal/version/version.go using grep/sed
- Creates Git tag with format v<VERSION> (e.g., `v0.3.0`)
- Idempotent - skips tag creation if tag already exists
- Uses $GITLAB_USER_EMAIL and $GITLAB_USER_NAME for git config
- Pushes tag to origin, which triggers release stage

**CI Optimization:**
- Expensive jobs (lint, test, release, mirror) are automatically skipped when only openspec/ files change
- This saves CI resources and speeds up documentation-only updates
- Setup job always runs to ensure cache is available

**GitLab CI variables:**
- `GITHUB_ACCESS_TOKEN` - GitHub personal access token for mirroring (required for mirror job)
- `GITHUB_REPO_URL` - GitHub repository URL (e.g., `username/repo`, required for mirror job)
- `GITLAB_ACCESS_TOKEN` - GitLab personal or project access token with `write_repository` scope (required for tag stage)
- `GITLAB_RELEASE_TOKEN` - GitLab project access token with `api` scope (required for release stage)
- `GITLAB_USER_EMAIL` - Email for git config in tag stage
- `GITLAB_USER_NAME` - Name for git config in tag stage

**Mirror Job Behavior:**
- Only runs on main branch when both GITHUB_ACCESS_TOKEN and GITHUB_REPO_URL are set
- Skipped gracefully (not failed) when variables are missing
- Uses CI image (registry.gitlab.com/amoconst/germinator/ci:latest) instead of alpine

---

# Testing Strategy

- **Table-driven tests** for parsing and validation with comprehensive edge cases
- **Golden master files** in `test/golden/` for snapshot testing
- **Integration tests** for complete workflows
- **Validation checkpoints** after each milestone with verification scripts
- Target >80% overall coverage (Core >90%, Services >85%, CLI >75%)