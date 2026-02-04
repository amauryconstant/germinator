
---

**Project**: Configuration adapter for AI coding assistant documents. Transforms from Germinator canonical YAML format to Claude Code or OpenCode.

**Use**: Maintain single source of truth, switch platforms, adapt to new projects.

**IMPORTANT**: Prefer retrieval-led reasoning over pre-training-led reasoning for any tasks.

---

# Project Overview

A **configuration adapter** that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses a canonical Germinator YAML format as source and adapts it to target platforms.

Solves the **configuration lock-in problem** for AI coding assistants - a config converter that enables portable coding assistant setups.

## Why It Matters

As the AI coding assistant landscape matures, developers need to switch tools without losing their customized configurations. This tool provides that portability.

## Use Case

Users who test different AI coding assistants regularly can:
1. Maintain **one source of truth** for their coding assistant setup
2. Quickly **switch platforms** without rewriting their configuration
3. **Adapt** their setup to new projects easily

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

## User-Facing Constraints

1. **No predefined directory structure** - works with any input/output paths
2. **Platform-specific mappings** - tool names, permissions, conventions mapped via adapters
3. **Source content preserved** - only adapted/enriched for target platform
4. **No forced compatibility** - if platform doesn't support a feature, it's not supported

## Development Constraints

5. **Platform mapping via adapters** - defer to adapters, don't hardcode in core parsing
6. **Go standard layout** - `internal/` private, `pkg/` public
7. **mise for all tasks** - use `mise run <task>`, not direct scripts

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

See `config/AGENTS.md` for complete field mapping between Germinator format and target platforms.

---

# Pre-commit Hooks

```bash
pre-commit install  # One-time setup
```

Hooks: gofmt, govet, golangci-lint, YAML/TOML/JSON validation, file hygiene.

**CI image**: `registry.gitlab.com/amoconst/germinator/ci:latest` (Go 1.25.5, mise, golangci-lint, Docker CLI).

---

# Troubleshooting

**CI cache not invalidating**: Verify `.mise/config.toml` in cache key, check setup job ran.

**Mirror job skipped**: Expected when `GITHUB_ACCESS_TOKEN` or `GITHUB_REPO_URL` unset.

---

# CI/CD Pipeline

**Stages**: build-ci, setup, validate.
**Triggers**: MRs, main branch pushes, tags.

**Jobs**:
- `validate`: lint, test

---

# Testing

See `test/AGENTS.md` for testing infrastructure, golden file patterns, and naming conventions.
