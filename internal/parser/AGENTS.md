**Location**: `internal/parser/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Parser Package

Loads, type-detects, and parses Germinator source documents (agent / command / skill / memory) and platform documents (claude-code, opencode) into canonical structs. Pure file I/O + YAML; no validation rules (those live in `internal/core` and the validating services).

## Files

| File | Purpose |
|------|---------|
| `doc.go` | Package doc (`package parser`) |
| `loader.go` | `LoadDocument`, `DetectType`, `validatePlatform`, `Parser` struct |
| `parser.go` | `ParseDocument` + `CanonicalAgent`/`CanonicalCommand`/`CanonicalSkill`/`CanonicalMemory` types |
| `platform_parser.go` | `ParsePlatformDocument` — bridges to `internal/claude-code` and `internal/opencode` adapters |
| `parser_test.go` | Unit tests |

## Public Surface

- `LoadDocument(ctx context.Context, filepath, platform string) (interface{}, error)` — canonical-source entry point: validates platform, detects doc type from filename, parses. `ctx` is checked at entry and forwarded to `DetectType` and `ParseDocument`. Returns `*core.ParseError` / `*core.FileError` on failure.
- `ParseDocument(ctx context.Context, filePath, docType string) (interface{}, error)` — checks `ctx.Err()` then reads the file and dispatches to `parseMemory` (handles optional frontmatter) or `parseDocumentWithFrontmatter` (agent/command/skill). `ctx` is forwarded to both helpers.
- `DetectType(ctx context.Context, filepath string) string` — filename regex → `"agent" | "command" | "skill" | "memory" | ""`. Recognizes both `agent-*.md` / `*-agent.md` (and `.yaml`) shapes. `ctx` is checked between regex iterations; detection is regex-only (no I/O), so ctx is accept-and-may-ignore per the cli-framework spec.
- `ParsePlatformDocument(ctx context.Context, path, platform, docType string) (interface{}, error)` — checks `ctx.Err()` then reads a platform YAML file, selects the adapter via `claudecode.New()` / `opencode.New()`, calls `ToCanonical`, and wraps the result in the corresponding `Canonical*` struct.
- `Parser` / `NewParser()` — thin method-style wrapper around `LoadDocument` for callers that prefer instance syntax.

### Canonical struct types

Each extends the embedded `core.*` type with `FilePath string` and `Content string` (the markdown body after the YAML frontmatter):

- `CanonicalAgent` — embeds `core.Agent`
- `CanonicalCommand` — embeds `core.Command`
- `CanonicalSkill` — embeds `core.Skill`
- `CanonicalMemory` — embeds `core.Memory`

## Design Notes

**Why this package exists.** The parser owns file→struct extraction only. It deliberately does **not** run validation (see the comment in `loader.go:48-49`); the service layer calls `core.Validate*` separately. Keeping parsing and validation decoupled lets the same parser feed both the `validate` and `adapt` commands.

**Dependencies.** Imports `internal/core` (types + errors) and, via `platform_parser.go`, the platform adapters `internal/claude-code` and `internal/opencode`. YAML decoding uses `gopkg.in/yaml.v3`.

**Filename convention.** `DetectType` is the only source of truth for "what type is this file?" — both `LoadDocument` and the library discovery walk rely on it. New doc-type patterns must be added here (and only here).

**Frontmatter handling.** `parseMemory` is dual-mode: it accepts either a body-only file or a `---`-delimited frontmatter + body file. `parseDocumentWithFrontmatter` always requires frontmatter for agent/command/skill. The shared `extractFrontmatter` helper returns `(yaml, body, err)` and never errors by design (annotated `//nolint:unparam`).
