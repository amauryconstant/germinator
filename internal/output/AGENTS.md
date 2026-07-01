**Location**: `internal/output/`
**Parent**: See `/internal/AGENTS.md` for package overview

---

# Output Package

Shared error formatting, output format flags, and exporters (JSON, table) for read-only commands. Re-exported through `cmdutil.AddOutputFlags` in slice 2+; not yet consumed by `cmd/` in slice 1.

## Files

| File | Purpose |
|------|---------|
| `errors.go` | `FormatError(io, err)` — dispatches via `errors.As` on `*core.{Parse,Validation,Transform,File,Config,PartialSuccess}Error` |
| `exporter.go` | `Exporter` interface; `JSONExporter` (2-space indent, trailing newline); `TableExporter` (`tab:"HEADER"` struct tag) |
| `output_flags.go` | `AddOutputFlags(cmd, *string)` — wires `--output` with completion for `json`/`table`/`plain` |
| `library.go` | `FormatResourcesList(lib)` — human-readable rendering of `library resources` (moved from cmd/library_formatters.go in slice-7 7.4.7) |
| `output_test.go` | FormatError dispatch + exporter round-trip tests |

## Key Surface

- `FormatError(io *iostreams.IOStreams, err error)` — no-op when `err == nil`
- `ValidOutputFormats` — `["json", "table", "plain"]`
- `DefaultOutputFormat` — `"plain"`
- `TableExporter` reads `tab:"..."` tag on struct fields; `"-"` hides a field
- `FormatResourcesList(*library.Library) string` — stable byte-identical plain rendering of the library resources list (slice-7 7.4.7)
