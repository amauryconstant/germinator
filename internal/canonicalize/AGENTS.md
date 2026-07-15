**Location**: `internal/canonicalize/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Canonicalize Package

Document-canonicalization shell package. Lifts the per-platform canonicalization pipeline (parse-platform-doc → core validators → marshal-canonical → write-output) out of `cmd/canonicalize.go`. The cmd-side `Canonicalizer` interface is satisfied by `canonicalize.NewService()` via structural typing.

## Files

| File | Purpose |
|------|---------|
| `canonicalize.go` | `Request` struct, `Service` interface, `canonicalizeService` adapter, `NewService` constructor, `Canonicalize` body, `validateCanonicalDoc` and `unwrapCanonicalErrors` helpers |
| `canonicalize_test.go` | Table-driven tests for happy-path, parse-error, validation-error, marshal-error, file-write-error using `t.TempDir()` fixtures |
| `canonicalize_golden_test.go` | (build tag `golden`) Byte-identical golden-file comparison against `test/golden/canonical/*.yaml.golden` for all platform/doc-type variants |

## Key Surface

- `Request{InputPath, OutputPath, Platform, DocType string}` — the per-call input (was `cmd.CanonicalizeRequest`)
- `Service` interface — single method `Canonicalize(ctx, *Request) (*core.CanonicalizeResult, error)`
- `NewService() Service` — returns the production wiring (zero-size service; orchestration is delegated to the imported shell units)
- `canonicalizeService.Canonicalize` — orchestrator (calls `parser.ParsePlatformDocument` → `validateCanonicalDoc` → `renderer.MarshalCanonical` → `os.WriteFile`)
- `validateCanonicalDoc` — per-doctype validator dispatch (returns `[]error`)
- `unwrapCanonicalErrors` — flatten helper for the `error` interface with `Unwrap() []error`

## Why this package exists

The `cmd/` layer is the parse + execute + respond shell (Cobra commands); canonicalization requires filesystem reads (the input doc), filesystem writes (the output YAML), and dispatches across `parser`, `core`, and `renderer` shell units. Per the imperative-shell discipline in `golang-cli-architecture/SKILL.md`, the orchestrator belongs at the package edge in `internal/<x>/` so the cmd layer remains a thin handler. The cmd-side `Canonicalizer` interface (declared in `cmd/canonicalize.go`) imports this package's `Request` type directly so the cmd layer never re-implements the adapter body.

## Production Wiring

```go
c := canonicalize.NewService()
result, err := c.Canonicalize(ctx, &canonicalize.Request{
    InputPath:  args[0],
    OutputPath: args[1],
    Platform:   platform, // pre-validated by core.ValidatePlatform upstream
    DocType:    docType,  // pre-validated by core.ValidateDocumentType upstream
})
```

On error, returns a typed `*core.ParseError` / `*core.ValidationError` / `*core.TransformError` / `*core.FileError` directly so the cmd layer wraps it once with `fmt.Errorf("canonicalizing document: %w", err)` and `main.go`'s `output.FormatError` renders the correct prefix.

## Pre-flight Validation Contract

| Input       | Caller responsibility                                                | Validation site                          |
|-------------|----------------------------------------------------------------------|------------------------------------------|
| Platform    | `core.ValidatePlatform(platform)` in `cmd/canonicalize.go:runCanonicalize` | `cmd/canonicalize.go`                    |
| DocType     | `core.ValidateDocumentType(docType)` in same                          | `cmd/canonicalize.go`                    |
| InputPath   | filesystem-stat at parse time                                         | `parser.ParsePlatformDocument`           |
| OutputPath  | filesystem write                                                      | `os.WriteFile` here                      |

The Service does NOT re-run validation for Platform/DocType; that's the cmd layer's job (single source of truth for the pre-flight contract).

## Dependencies

- `internal/core` (typed errors + functional-core validators)
- `internal/parser` (filesystem-backed parse)
- `internal/renderer` (canonical YAML marshal)
- `context` (cancellation), `os` (`os.WriteFile`)
- No other I/O packages

## Coverage

`canonicalize_test.go` covers happy-path round-trip (input doc → canonical YAML bytes → on-disk file), the parse-error path, the validation-error path, the marshal-error path, and the file-write-error path. `canonicalize_golden_test.go` (build tag `golden`) byte-compares against the full `test/golden/canonical/` corpus.
