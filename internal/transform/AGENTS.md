**Location**: `internal/transform/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Transform Package

Document-transformation shell package. Lifts the canonical parse → render → write pipeline out of `cmd/transformer.go`. The cmd-side `Transformer` interface is satisfied by `transform.NewService(...)` via structural typing.

## Files

| File | Purpose |
|------|---------|
| `transform.go` | `Request` struct, `Service` interface, `transformService` adapter, `NewService` constructor, `Transform` body |
| `transform_test.go` | Table-driven tests covering happy-path round-trip per platform, missing-input error, parse error, render error, file-write error using `t.TempDir()` fixtures |

## Key Surface

- `Request{InputPath, OutputPath, Platform string}` — the per-call input (was `cmd.TransformRequest`)
- `Service` interface — single method `Transform(ctx, *Request) (*core.TransformResult, error)`
- `NewService(p *parser.Parser, s *renderer.Serializer) Service` — constructor with injected dependencies (parser + serializer are constructed once, not on every call)
- `transformService.Transform` — orchestrator (calls `parser.LoadDocument` → `renderer.RenderDocument` → `os.WriteFile`)
- `transformService` holds the injected `parser` and `serializer` so callers can reuse the same Service across many calls

## Why this package exists

The `cmd/` layer is the parse + execute + respond shell (Cobra commands); document-transformation requires filesystem reads (the input doc), filesystem writes (the rendered output), and dispatches across `parser` and `renderer` shell units. Per the imperative-shell discipline in `golang-cli-architecture/SKILL.md`, the orchestrator belongs at the package edge in `internal/<x>/` so the cmd layer remains a thin handler. The cmd-side `Transformer` interface (declared in `cmd/adapt.go`) imports this package's `Request` type directly so the cmd layer never re-implements the adapter body.

## Production Wiring

```go
t := transform.NewService(parser.NewParser(), renderer.NewSerializer())
result, err := t.Transform(ctx, &transform.Request{
    InputPath:  args[0],
    OutputPath: args[1],
    Platform:   platform, // pre-validated by core.ValidatePlatform upstream
})
```

On error, returns either a wrapped load error, a `*core.TransformError` (render failure), or a `*core.FileError` (write failure) so the cmd layer wraps it once with `fmt.Errorf("transforming document: %w", err)` and `main.go`'s `output.FormatError` renders the correct prefix.

## Pre-flight Validation Contract

| Input       | Caller responsibility                                       | Validation site                |
|-------------|-------------------------------------------------------------|--------------------------------|
| Platform    | `core.ValidatePlatform(platform)` in `cmd/adapt.go:runAdapt` | `cmd/adapt.go`                 |
| InputPath   | filesystem-stat at parse time                                | `parser.LoadDocument`          |
| OutputPath  | filesystem write                                             | `os.WriteFile` here            |

The Service does NOT re-run validation for Platform; that's the cmd layer's job (single source of truth for the pre-flight contract).

## Dependencies

- `internal/core` (typed errors + result types)
- `internal/parser` (filesystem-backed parse)
- `internal/renderer` (template-driven render)
- `context` (cancellation), `os` (`os.WriteFile`)
- No other I/O packages

## Coverage

`transform_test.go` covers happy-path round-trip for both supported platforms (claude-code / opencode) via `t.TempDir()` fixtures, the missing-input-file failure path, the render-failure path (malformed source), and the file-write-error path (read-only output directory).