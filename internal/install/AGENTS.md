**Location**: `internal/install/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Install Package

Resource-installation shell package. Lifts the per-ref parse → render → write loop out of `cmd/initializer.go`. The cmd-side `Initializer` interface is satisfied by `install.NewService(...)` via structural typing.

The package name `install` was chosen over `init` to avoid collision with Go's reserved `init` identifier (per design Decision 1).

## Files

| File | Purpose |
|------|---------|
| `install.go` | `Request` struct, `Service` interface, `installService` adapter, `NewService` constructor, `Initialize` body |
| `install_test.go` | Table-driven tests for per-ref scenarios: load → render → write, dry-run short-circuit, existing-file check (force vs no-force), ctx cancellation, partial-success aggregation |

## Key Surface

- `Request{Library, Platform, OutputDir, Refs, DryRun, Force}` — the per-call input (was `cmd.InitializeRequest`)
- `Service` interface — single method `Initialize(ctx, *Request) ([]core.InitializeResult, error)`
- `NewService(p *parser.Parser, s *renderer.Serializer) Service` — constructor with injected dependencies (parser + serializer are constructed once, not on every call)
- `installService.Initialize` — orchestrator (loops over `req.Refs`, resolves each via `library.ResolveResource` + `library.ParseRef` + `library.GetOutputPath`, then runs the canonical load → render → write pipeline)
- `installService` holds the injected `parser` and `serializer` so callers can reuse the same Service across many calls

## Why this package exists

The `cmd/` layer is the parse + execute + respond shell (Cobra commands); resource installation requires filesystem reads (per-ref input docs), filesystem writes (the rendered per-ref outputs), and dispatches across `parser`, `renderer`, and `library` shell units. Per the imperative-shell discipline in `golang-cli-architecture/SKILL.md`, the orchestrator belongs at the package edge in `internal/<x>/` so the cmd layer remains a thin handler. The cmd-side `Initializer` interface (declared in `cmd/init.go`) imports this package's `Request` type directly so the cmd layer never re-implements the adapter body.

## Production Wiring

```go
svc := install.NewService(parser.NewParser(), renderer.NewSerializer())
results, err := svc.Initialize(ctx, &install.Request{
    Library:   lib,
    Platform:  platform, // pre-validated by core.ValidatePlatform upstream
    OutputDir: outputDir,
    Refs:      refs,
    DryRun:    dryRun,
    Force:     force,
})
```

The error return is reserved for transport-level failures; per-resource outcomes always live in `result.Error`, allowing callers to synthesize `*core.PartialSuccessError`.

## Pre-flight Validation Contract

| Input       | Caller responsibility                                  | Validation site             |
|-------------|---------------------------------------------------------|-----------------------------|
| Platform    | `core.ValidatePlatform(platform)` in `cmd/init.go:runInit` | `cmd/init.go`               |
| Library     | `opts.Library()` (lazy loader built by NewCmdInit)       | lazy; cmd-side              |
| Refs XOR Preset | `runInit` mutex check                                  | `cmd/init.go`               |
| Per-ref     | `library.ResolveResource` / `library.ParseRef` / `library.GetOutputPath` | `library` shell unit |

The Service does NOT re-run validation for Platform or the Refs/Preset mutex; that's the cmd layer's job (single source of truth for the pre-flight contract).

## Dependencies

- `internal/core` (typed errors + result types)
- `internal/library` (resolve + parse ref + output path)
- `internal/parser` (filesystem-backed parse)
- `internal/renderer` (template-driven render)
- `context` (cancellation), `os` (`os.Stat`, `os.MkdirAll`, `os.WriteFile`), `path/filepath`
- No other I/O packages

## Coverage

`install_test.go` covers per-ref scenarios matching the spec's cli-init-command scenarios: load → render → write happy path, dry-run short-circuit (no file writes, success result returned), existing-file check (force overwrites, no-force returns *core.FileError), missing-ref failure, ctx cancellation, partial-success aggregation across multiple refs.