**Location**: `internal/output/`
**Parent**: See `/internal/AGENTS.md` for package overview
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/04-output.md`, `@.opencode/skills/golang-cli-architecture/references/05-errors.md`, `@.opencode/skills/golang-error-handling/SKILL.md` (`errors.As` dispatch table for `FormatError`)

---

# Output Package

Shared error formatting, output-format flags, and exporters (JSON, table) for read-only commands. Consumed by `cmd/` via direct import of `output`.

## Files

| File | Purpose |
|------|---------|
| `errors.go` | `FormatError(io, err)` — dispatches via `errors.As` on `*core.{Parse,Validation,Transform,File,Config,NotFound,PartialSuccess,Operation,Initialize,Usage}Error` and `*config.WriteError` (12-arm switch per `openspec/specs/cli-error-formatting/spec.md`) |
| `exporter.go` | `Exporter` interface; `JSONExporter` (2-space indent, trailing newline); `TableExporter` (`tab:"HEADER"` struct tag) |
| `output_flags.go` | `AddOutputFlags(cmd, *string)` — wires `--output` with completion for `json`/`table`/`plain` |
| `library.go` | `FormatResourcesList(lib)` — human-readable rendering of `library resources` |
| `output_test.go` | FormatError dispatch + exporter round-trip tests |

## Key Surface

- `FormatError(io *iostreams.IOStreams, err error)` — no-op when `err == nil`
- `ValidOutputFormats` — `["json", "table", "plain"]`
- `DefaultOutputFormat` — `"plain"`
- `TableExporter` reads `tab:"..."` tag on struct fields; `"-"` hides a field
- `FormatResourcesList(*library.Library) string` — stable byte-identical plain rendering of the library resources list
