
---

**Project**: Configuration adapter for AI coding assistant documents. Transforms from Germinator canonical YAML format to Claude Code or OpenCode.

**Use**: Maintain single source of truth, switch platforms, adapt to new projects.

---

# OpenSpec (OPSX)

Spec-driven development with artifact-based workflow.

| Command | Purpose |
|---------|---------|
| `/opsx:explore` | Think through ideas |
| `/opsx:new <name>` | Start change |
| `/opsx:continue` | Next artifact |
| `/opsx:ff` | Fast-forward all |
| `/opsx:apply` | Implement |
| `/opsx:verify` | Validate |
| `/opsx:sync` | Sync specs |
| `/opsx:archive` | Archive done |

**Config**: `openspec/config.yaml` injects conventions, defines artifact rules, sets workflow schema.
**Schemas**: `openspec schemas --json`

---

# Development

**Prerequisites**: Go 1.25.5+, mise task runner

## mise Commands

```bash
mise run build            # Build CLI
mise run check            # All validation
mise run lint             # Lint code
mise run lint:fix         # Auto-fix
mise run format           # Format Go
mise run test             # Run tests
mise run test:coverage    # With coverage
mise run clean            # Clean artifacts
mise run version:patch    # Bump patch
mise run version:minor    # Bump minor
mise run version:major    # Bump major
mise tasks                # List all
```

**Rules**:
- `mise run check` before committing
- `go mod tidy` after dependency changes
- Use `mise exec -- <command>` for mise-installed tools (e.g., `mise exec -- golangci-lint run`)

## Directory

```
germinator/
 ├── cmd/              # CLI entry (Cobra)
 ├── internal/         # Private code
 │   ├── core/        # Parser, loader, serializer
 │   └── services/    # Validation, transformation
 ├── config/          # Schemas, templates, adapters
 ├── test/            # Fixtures, golden
 └── .mise/           # Task runner config
```

---

# Key Constraints

1. **No predefined paths** - CLI works with any input/output
2. **Platform mapping via adapters** - Defer to adapters, don't hardcode in core parsing
3. **Preserve source content** - Only adapt/enrich, don't discard
4. **No forced compatibility** - Skip unsupported features
5. **Go standard layout** - `internal/` private, `pkg/` public
6. **mise for all tasks** - Use `mise run <task>`, not direct scripts

---

# CLI Pattern

```bash
cli action input_file output_file --platform <platform> [options]
```

**Platforms**: `claude-code`, `opencode` (required, no default)

**Examples**:
```bash
# Validate
./germinator validate agent.yaml --platform claude-code

# Adapt to Claude Code
./germinator adapt agent.yaml .claude/agents/my-agent.yaml --platform claude-code

# Adapt to OpenCode
./germinator adapt skill.yaml .opencode/skills/my-skill/SKILL.md --platform opencode
```

See `IMPLEMENTATION_PLAN.md` for command structures.

---

# Germinator Source Format

Germinator uses a canonical YAML format containing ALL fields for ALL platforms. This enables unidirectional transformation to either Claude Code or OpenCode.

**Key Principles**:
- Source YAML includes both Claude Code and OpenCode fields
- Platform-specific fields are used or skipped based on target platform
- Model IDs are user-provided platform-specific values (no normalization)
- Permission modes are transformed from Claude Code enum to OpenCode objects

**Document Types**:
- **Agent**: Tools, permission modes, model configuration, prompts
- **Command**: Tool permissions, execution context, agent references
- **Skill**: Metadata, hooks, compatibility, tool restrictions
- **Memory**: File paths and narrative content for project context

---

# Field Mapping Reference

## Agent

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ⚠ omitted (uses filename as identifier) |
| description | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| tools | ✓ | ✓ (converted to lowercase) |
| disallowedTools | ✓ | ✓ (converted to lowercase, set false) |
| permissionMode | ✓ | → Permission object (nested with ask/allow/deny) |
| skills | ✓ | ⚠ skipped (not supported) |
| mode | - | ✓ (primary/subagent/all, defaults to all) |
| temperature | - | ✓ (*float64 pointer, omits when nil) |
| maxSteps | - | ✓ |
| hidden | - | ✓ (omits when false) |
| prompt | - | ✓ |
| disable | - | ✓ (omits when false) |

## Command

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ⚠ omitted (uses filename as identifier) |
| description | ✓ | ✓ |
| allowed-tools | ✓ | ✓ (converted to lowercase) |
| disallowed-tools | ✓ | ✓ (converted to lowercase, set false) |
| subtask | ✓ | ✓ |
| argument-hint | ✓ | ⚠ skipped |
| context | ✓ (fork) | ✓ (fork) |
| agent | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| disable-model-invocation | ✓ | ⚠ skipped (not supported) |

## Skill

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| name | ✓ | ✓ |
| description | ✓ | ✓ |
| allowed-tools | ✓ | ✓ (converted to lowercase) |
| disallowed-tools | ✓ | ✓ (converted to lowercase) |
| license | ✓ | ✓ |
| compatibility | ✓ | ✓ |
| metadata | ✓ | ✓ |
| hooks | ✓ | ✓ |
| model | ✓ | ✓ (full provider-prefixed ID) |
| context | ✓ (fork) | ✓ (fork) |
| agent | ✓ | ✓ |
| user-invocable | ✓ | ⚠ skipped (not supported) |

## Memory

| Germinator Field | Claude Code | OpenCode |
|------------------|-------------|----------|
| paths | ✓ | → @ file references (one per line) |
| content | ✓ | → Narrative context (rendered as-is) |

**Legend**: ✓ = Supported, → = Transformed, ⚠ = Skipped

---

# OpenCode Platform Support

OpenCode support added in v0.4.0 with the following characteristics:

**Transformation Flow**:
1. Parse Germinator YAML (all fields)
2. Validate platform-specific rules
3. Render using OpenCode templates
4. Transform permission modes to permission objects
5. Generate platform-specific output format

**OpenCode-Specific Features**:
- Agent `name` field is omitted in output (uses filename as identifier)
- Permission objects with nested structures: `{"edit": {"*": "ask"}, "bash": {"*": "allow"}}`
- Agent modes: `primary`, `subagent`, `all` (default when empty)
- Temperature as *float64 pointer (nil omits field, 0.0 renders explicitly)
- MaxSteps configuration
- Hidden and disable boolean fields (omit when false)
- Custom prompts

**Known Limitations**:
- Permission mode transformation is basic approximation
- Agent `skills` list not supported (skipped)
- Command `disable-model-invocation` not supported (skipped)
- Skill `user-invocable` not supported (skipped)
- No bidirectional conversion

**Breaking Changes in v0.4.0**:
- `--platform` flag required (no default)
- Germinator source format replaces Claude Code YAML
- `Validate(platform string)` signature change


---

# Release

Git tags are source of truth.

```bash
# Prerequisites
mise run release:validate    # Clean, main, valid config

# Create tag (auto-bumps internal/version/version.go)
mise run release:tag patch   # v0.3.20 → v0.3.21
mise run release:tag minor   # v0.3.20 → v0.4.0
mise run release:tag major   # v0.3.20 → v1.0.0

# Test before tagging
mise run release:dry-run
```

**CI**: Tag push triggers GitLab CI → lint, test, GoReleaser, release artifacts.

**Artifacts**: Cross-platform binaries (linux/darwin/windows, amd64/arm64), checksums, SBOM.

## Tool Management

```bash
mise run tools:check     # Check updates
mise run tools:update    # Update versions
mise install --yes       # Install after update
git diff .mise/config.toml  # Review changes
```

---

# Pre-commit Hooks

```bash
pre-commit install  # One-time setup
```

Hooks: gofmt, govet, golangci-lint, YAML/TOML/JSON validation, file hygiene.

**CI image**: `registry.gitlab.com/amoconst/germinator/ci:latest` (Go 1.25.5, mise, golangci-lint, GoReleaser, Docker CLI).

---

# Troubleshooting

**CI cache not invalidating**: Verify `.mise/config.toml` in cache key, check setup job ran.

**Mirror job skipped**: Expected when `GITHUB_ACCESS_TOKEN` or `GITHUB_REPO_URL` unset.

**release:validate fails**: Check `git status`, verify main branch, `goreleaser check`.

**Tag not created**: Verify `internal/version/version.go` changed, pushed to main, review logs. Required vars: `GITLAB_USER_EMAIL`, `GITLAB_USER_NAME`, `GITLAB_ACCESS_TOKEN` (write_repository), `GITLAB_RELEASE_TOKEN` (api).

**Recreate tag**:
```bash
git tag -d v0.3.0 && git push origin :refs/tags/v0.3.0
```

---

# CI/CD Pipeline

**Stages**: build-ci, setup, lint, test, tag, release, mirror.
**Triggers**: MRs, main branch pushes.

**Tag stage**: Auto-runs when `internal/version/version.go` changes on main. Idempotent, creates format `vX.Y.Z`.

**Optimization**: Expensive jobs skip when only `openspec/` files change.

**Required vars**: `GITHUB_ACCESS_TOKEN`, `GITHUB_REPO_URL` (mirror), `GITLAB_ACCESS_TOKEN` (tag), `GITLAB_RELEASE_TOKEN` (release), `GITLAB_USER_EMAIL`, `GITLAB_USER_NAME`.

---

# Testing

- Table-driven tests for parsing/validation (edge cases)
- Integration tests for workflows
- Validation checkpoints after milestones
