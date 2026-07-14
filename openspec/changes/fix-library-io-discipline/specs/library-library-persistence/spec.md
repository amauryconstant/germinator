# library-library-persistence Specification (delta)

## MODIFIED Requirements

### Requirement: Atomic library.yaml writes

All mutations to `library.yaml` SHALL be performed atomically via the write-temp-then-rename pattern:

1. Write new content to `library.yaml.tmp`
2. `Sync()` the temp file
3. `os.Rename(tmp, library.yaml)`

If `os.Rename` fails with `syscall.EXDEV` (cross-device link, e.g., temp file on `/tmp` and target on `/home`), the writer SHALL fall back to a copy-then-remove sequence: open `tmp` for read, open `library.yaml` with `O_WRONLY|O_CREATE|O_TRUNC`, `io.Copy`, `Sync()` the target, then `os.Remove(tmp)`. The fallback is atomic-or-fail at the user-observable level: the new `library.yaml` is fully written before the old temp is removed.

All library.yaml writers SHALL delegate to a single `library.atomicWriteFile(path, data, perm)` helper (`internal/library/saver.go`, next to `SaveLibrary`). The four call sites are:

- `internal/library/adder.go:330-335` — `AddResource` save block
- `internal/library/remover.go:190-195` — `RemoveResource` save block
- `internal/library/remover.go:223-228` — `RemovePreset` save block
- `internal/library/saver.go:30-35` — `SaveLibrary` save block

`SaveLibrary` (`internal/library/saver.go:15`) is also a consumer; it calls `atomicWriteFile` for its `library.yaml` write — there are no remaining direct `os.Rename` library.yaml write sites in `internal/library/`.

**Change**: ADDED the `EXDEV` fallback clause AND the `atomicWriteFile` helper factoring. Pre-change state: three sites (`adder.go:333`, `remover.go:193`, `remover.go:226`) used the temp+rename pattern **without** EXDEV fallback, and one site (`saver.go:33`, `SaveLibrary`) used non-atomic direct `os.WriteFile` with no rename at all. Post-change: all four sites delegate to `atomicWriteFile`, gaining (a) EXDEV fallback for the three rename sites, and (b) atomicity for the `SaveLibrary` site. Cross-filesystem scenarios (e.g., `TMPDIR=/tmp`, library on `/home`) for the three rename sites now succeed via the copy+remove fallback. The `SaveLibrary` change is a torn-write fix, not an EXDEV fix.

#### Scenario: Successful atomic save on same filesystem

- **WHEN** any library.yaml mutation is invoked (via `AddResource`, `RemoveResource`, `RemovePreset`, or `SaveLibrary`) and the temp file is on the same filesystem as `library.yaml`
- **THEN** the new content SHALL be written to `library.yaml.tmp` first via `atomicWriteFile`
- **AND** on successful close, `os.Rename` SHALL atomically replace `library.yaml`
- **AND** no partial data SHALL ever be visible at `library.yaml` (no torn writes)

#### Scenario: Cross-filesystem save uses copy-then-remove fallback

- **GIVEN** `TMPDIR` resolves to a different filesystem from the library directory (e.g., `TMPDIR=/tmp` and library on `/home/user/lib`)
- **WHEN** any library.yaml mutation is invoked
- **AND** the initial `os.Rename` returns `syscall.EXDEV`
- **THEN** `atomicWriteFile` SHALL fall back to copy-then-remove (`io.Copy` + `os.Remove(tmp)`) to write the new content
- **AND** the resulting `library.yaml` SHALL contain the new content

#### Scenario: Crash mid-write preserves prior content

- **GIVEN** the temp-file write is interrupted (process killed before `os.Rename` or `io.Copy`)
- **WHEN** `LoadLibrary()` is called on the library directory
- **THEN** the prior `library.yaml` SHALL be returned intact
- **AND** `library.yaml.tmp` MAY remain as a partial file (cleaned up on next successful save)

#### Scenario: All library.yaml writers use atomicWriteFile

- **WHEN** the codebase is searched for `os.Rename` in `internal/library/`
- **THEN** exactly one match SHALL exist: the `os.Rename` call inside `atomicWriteFile` itself (pre-change count: 3, in `adder.go:333`, `remover.go:193`, `remover.go:226`)
- **AND** the three former EXDEV-vulnerable direct-call sites (`adder.go:333`, `remover.go:193`, `remover.go:226`) SHALL be gone, replaced by calls to `atomicWriteFile`
- **AND** the former non-atomic direct-`os.WriteFile` site (`saver.go:33`, `SaveLibrary`) SHALL also be gone, replaced by a call to `atomicWriteFile`
- **AND** `rg "atomicWriteFile" internal/library/` SHALL return five matches: one definition plus four callers

## ADDED Requirements

### Requirement: Unix-only permission bits

The library package uses Unix permission bits. On Windows, these bits are silently ignored by `os.Chmod`; the resulting file permissions follow the Windows default. Windows support is out of scope; the limitation SHALL be documented at each call site.

The pre-existing permission values are preserved by this change (no behavior change on Unix/macOS):

- Directories: `0o750` (set via `os.MkdirAll` at `internal/library/adder.go:105`, `internal/library/creator.go:57`, `internal/library/saver.go:21`).
- `library.yaml` initial creation via `CreateLibrary`: `0o644` (`internal/library/creator.go:65`).
- `library.yaml` incremental mutations (via `atomicWriteFile` from `AddResource`, `RemoveResource`, `RemovePreset`, `SaveLibrary`): `0o600`.
- Resource files: `0o644` (`internal/library/adder.go:124`).

The split between `CreateLibrary` (`0o644`) and the mutation path via `atomicWriteFile` (`0o600`) for `library.yaml` is pre-existing behavior; this requirement documents both without prescribing unification. Unifying them is out of scope for this change.

**Change**: NEW requirement documenting a pre-existing limitation. The pre-change code did not document the platform behavior; a future contributor might add a Windows fix without considering whether to claim Windows support.

#### Scenario: Permission bits on Unix/macOS

- **WHEN** `germinator library init` creates a library on Linux or macOS
- **THEN** directories SHALL be created with mode `0750`
- **AND** the initial `library.yaml` SHALL be created with mode `0644`
- **AND** subsequent library.yaml mutations (via `atomicWriteFile`) SHALL write `library.yaml` with mode `0600`
- **AND** resource files SHALL be created with mode `0644`

#### Scenario: Permission bits on Windows

- **WHEN** `germinator library init` runs on Windows
- **THEN** the `os.Chmod` calls SHALL be no-ops (Windows ignores the bits)
- **AND** file permissions SHALL follow the Windows default
- **AND** the documentation comment at each call site (`adder.go:105`, `creator.go:57`, `saver.go:21`) SHALL state that Windows is out of scope
