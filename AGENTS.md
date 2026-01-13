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
# Install golangci-lint
mise install golangci-lint

# Run all validation checks
mise run validate

# Quick build check
mise run smoke-test

# Format Go code
mise run format

# Discover all tasks
mise run --help
```

**Important**: Run `mise run validate` before committing. Run `go mod tidy` after any dependency changes.

For tools installed via mise, use `mise exec -- <command>` (e.g., `mise exec -- golangci-lint run`).

---

# Directory Structure

```
germinator/
├── cmd/                    # CLI entry point (Cobra framework)
│   └── root.go
├── internal/              # Private application code
│   ├── core/             # Core interfaces: DocumentParser, SchemaValidator, TemplateEngine
│   └── services/         # Business logic: ValidationService, TransformationService
├── pkg/                   # Public library code
│   └── models/           # Domain models: Document, Agent, Command, Memory, Skill
├── config/               # Configuration files
│   ├── schemas/          # JSON Schema files for document validation
│   ├── templates/        # Go template files for output rendering
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

# Technology Stack

- **Language**: Go 1.25.5
- **CLI Framework**: Cobra
- **Validation**: JSON Schema (xeipuuv/gojsonschema)
- **Task Runner**: mise
- **Linting**: golangci-lint (gofmt, govet, errcheck)

---

# Testing Strategy

- **Table-driven tests** for parsing and validation with comprehensive edge cases
- **Golden master files** in `test/golden/` for snapshot testing
- **Integration tests** for complete workflows
- **Validation checkpoints** after each milestone with verification scripts
- Target >80% overall coverage (Core >90%, Services >85%, CLI >75%)