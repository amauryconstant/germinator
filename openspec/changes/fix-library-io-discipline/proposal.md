## Why

The 2026-07-08 code review identified 4 I/O discipline findings in `internal/library/` that violate the Unix contract and the cross-filesystem safety contract:

1. **B-003** — `internal/library/adder.go:92` writes dry-run output directly to `os.Stdout`, polluting piped consumers' stdout.
2. **BD-005 / B-004** — `internal/library/creator.go:38` writes dry-run output directly to `os.Stdout`, same issue.
3. **C-009** — `internal/library/adder.go:333` uses `os.Rename` without an `EXDEV` fallback; cross-filesystem renames (e.g., `/tmp` → `/home`) fail.
4. **C-018** — `internal/library/adder.go:785` `DiscoverOrphans` early-breaks on `ctx.Err()` at the top of the directory loop only; per-directory scan cannot be cancelled in parallel.

The first two are Top-10 fixes (#7 and #8 in the review). The library package must not write to `os.Stdout` — the cmd layer is responsible for rendering user-facing output. Cross-filesystem rename safety and per-directory cancellation are correctness fixes for the persistence contract documented in `library-library-persistence/spec.md`.

This change enforces the I/O discipline contract for the library package. It is a **production-code refactor** with spec deltas.

## What Changes

- **MODIFY** `internal/library/adder.go` — add `stdout io.Writer` parameter to `AddResource` / `BatchAddResources`; thread `opts.IO.Out` from `cmd/library_add.go`. Remove the `fmt.Fprintln(os.Stdout, ...)` dry-run block; render via the writer.
- **MODIFY** `internal/library/creator.go` — add `stdout io.Writer` parameter to `CreateLibrary`; thread `opts.IO.Out` from `cmd/library_init.go`. Remove the `fmt.Fprintln(os.Stdout, ...)` dry-run block.
- **MODIFY** `internal/library/adder.go:333` — add `EXDEV` fallback in `os.Rename`:
  ```go
  if err := os.Rename(tmpPath, yamlPath); err != nil {
      if errors.Is(err, syscall.EXDEV) {
          // Cross-filesystem: copy + remove
          if err := copyFile(tmpPath, yamlPath); err != nil { ... }
          _ = os.Remove(tmpPath)
      } else {
          return core.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
      }
  }
  ```
- **MODIFY** `internal/library/adder.go:785` — wrap the `DiscoverOrphans` directory scan in `errgroup.Group` for per-directory cancellation. Each goroutine checks `ctx.Err()` before processing its directory.
- **DOCUMENT** Windows permission-bit limitation (`0o750`, `0o644` ignored on Windows) at `internal/library/adder.go:105`. Two options: (a) document as known limitation, or (b) add `runtime.GOOS` switch. Decision 1 in `design.md` evaluates both.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- **`library-library-persistence`** — add EXDEV fallback requirement for atomic writes; document cross-filesystem safety.
- **`library-library-scaffolding`** — `CreateLibrary` SHALL accept an `io.Writer` for dry-run output; `internal/library` SHALL NOT write to `os.Stdout` directly.
- **`library-library-resource-import`** — `AddResource` SHALL accept an `io.Writer` for dry-run output.
- **`library-library-batch-add`** — `BatchAddResources` SHALL accept an `io.Writer` for dry-run output.
- **`library-library-orphan-discovery`** — `DiscoverOrphans` SHALL use `errgroup` for per-directory cancellation; cancellation MUST be checked at the directory level, not just at the top of the loop.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `internal/library/adder.go:92` | `os.Stdout` → `io.Writer` param | -2 / +3 |
| `internal/library/adder.go:333` | EXDEV fallback | +8 |
| `internal/library/adder.go:785` | `errgroup` wrapping | +15 / -3 |
| `internal/library/adder.go:105` | Windows doc comment | +3 |
| `internal/library/creator.go:38` | `os.Stdout` → `io.Writer` param | -2 / +3 |
| `internal/library/adder_test.go` | Test signature updates | +5 / -5 |
| `internal/library/creator_test.go` | Test signature updates | +5 / -5 |
| `cmd/library_add.go:347` | Pass `opts.IO.Out` to `AddResource` | +1 |
| `cmd/library_init.go:144` | Pass `opts.IO.Out` to `CreateLibrary` | +1 |

### Affected systems

- **Public API:** all 3 function signatures (`AddResource`, `BatchAddResources`, `CreateLibrary`) gain a new `io.Writer` parameter. The cmd layer (sole caller) is updated. **BREAKING** for any external consumer, but the package is `internal/`; zero external consumers.
- **Cross-filesystem safety:** atomic writes now succeed across filesystems (e.g., when `tmpPath` is on `/tmp` and `yamlPath` is on `/home`). This was previously a silent failure path.
- **Per-directory cancellation:** `DiscoverOrphans` now respects ctx cancellation at every directory, not just the top of the loop. Tests with `t.Cleanup` that cancel mid-scan no longer leak goroutines.
- **Windows perm bits:** documented as known limitation; no behavior change on Unix/macOS. Windows support remains out of scope (consistent with the existing `isTerminalFile` caveat in `internal/iostreams`).

## Risks

- **EXDEV fallback introduces a small window where the new file exists at `yamlPath` but the temp file is not yet removed.** A crash in that window leaves both files; the next read sees the new file. **Mitigation**: document the window in the `library-library-persistence/spec.md` delta; the next `LoadLibrary` call will read the new file (correct semantics) and ignore the temp file.
- **`errgroup` adds a goroutine per directory.** For libraries with thousands of directories, this is a non-trivial memory cost. **Mitigation**: cap concurrency at 8 workers via `errgroup.WithContext` + a semaphore channel; design Decision 3 evaluates the cap value.
- **`io.Writer` parameter is a signature break** for `AddResource` / `BatchAddResources` / `CreateLibrary`. **Mitigation**: the cmd layer is the sole caller (per the project convention); the migration is mechanical. Test files are updated in the same commit.
- **Windows documentation is a non-fix.** The review's C-016 finding is documented but not fixed. **Mitigation**: design Decision 1 evaluates the `runtime.GOOS` switch as an alternative; if the user prefers the active fix, the scope expands by ~15 LOC.
