**Location**: `internal/validate/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Validate Package

Document-validation shell package. Lifts the per-platform validation pipeline (parse → core validators + opencode platform validators → joined-error flatten) out of `cmd/validate.go`. The cmd-side `Validator` interface is satisfied by `validate.NewService()` via structural typing.

## Files

| File | Purpose |
|------|---------|
| `validate.go` | `Request` struct, `Service` interface, `validateService` adapter, `NewService` constructor, `Validate` body, `unwrapJoinedErrors` helper |
| `validate_test.go` | Table-driven tests covering happy-path and per-platform coverage via `t.TempDir()` fixtures |

## Key Surface

- `Request{InputPath, Platform string}` — the per-call input (was `cmd.ValidateRequest`)
- `Service` interface — single method `Validate(ctx, *Request) (*core.ValidateResult, error)`
- `NewService() Service` — returns the production wiring (zero-size service; orchestration lives in the type switch + per-doc-type validators)
- `validateService.Validate` — orchestrator (calls `parser.DetectType` → `parser.ParseDocument` → `core.ValidateX` (+ optional opencode validators) → flatten joined errors)
- `unwrapJoinedErrors` — flatten helper for the `error` interface with `Unwrap() []error`

## Why this package exists

The `cmd/` layer is the parse + execute + respond shell (Cobra commands); document-validation requires filesystem I/O (`parser.DetectType`, `parser.ParseDocument`) and depends on multiple functional-core validators. Per the imperative-shell discipline in `golang-cli-architecture/SKILL.md`, I/O packages live at the package edge in `internal/<x>/` so the cmd layer remains a thin orchestrator. The cmd-side `Validator` interface (declared in `cmd/validate.go`) imports this package's `Request` type directly so the cmd layer never re-implements the adapter body.

## Production Wiring

```go
v := validate.NewService()
result, err := v.Validate(ctx, &validate.Request{
    InputPath: args[0],
    Platform:  platform, // already validated via core.ValidatePlatform
})
```

`result.Valid()` returns `true` only when `result.Errors` is empty; otherwise the first error is the per-call return value so the cmd layer can wrap it with `fmt.Errorf("validating document: %w", err)`.

## Dependencies

- `internal/core`, `internal/core/opencode` (functional core validators + opencode platform validators)
- `internal/parser` (filesystem-backed parse)
- `context` (cancellation)
- No I/O packages outside stdlib are imported; the package performs I/O only through the imported shell units above.

## Coverage

`validate_test.go` covers happy-path for each doc type (agent / command / skill / memory) per platform, the unrecognizable-filename failure path, the unknown-doc-type failure path, and the `unwrapJoinedErrors` flatten behavior across `nil`, single-error, and joined-error inputs.
