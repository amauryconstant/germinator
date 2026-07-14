# library-library-persistence Specification

## Purpose

Define the persistence guarantees for the library directory: atomic writes for `library.yaml`, file permissions on the directory tree, and the single-writer contract (with optional cross-process locking for advanced use cases).

## Requirements

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

### Requirement: Library directory permissions

When the system creates a new library directory, the permissions SHALL be `0750` (rwxr-x---). This restricts access to the owner plus a same-group read; other users have no access.

#### Scenario: library init creates 0750 directory

- **WHEN** `germinator library init --path /new/lib` creates a fresh library
- **THEN** `/new/lib` SHALL be created with mode `0750`

### Requirement: library.yaml permissions

The `library.yaml` file SHALL be created with mode `0640` (rw-r-----). This restricts write access to the owner, and read access to the owner plus a same-group read.

#### Scenario: library.yaml is 0640

- **WHEN** `Library.Init` writes the initial `library.yaml`
- **THEN** the file permissions SHALL be `0640`
- **AND** when `Library.Save` rewrites it via the atomic pattern, the resulting `library.yaml` SHALL ALSO be `0640`

### Requirement: Resource subdirectory permissions

The four resource subdirectories (`skills/`, `agents/`, `commands/`, `memory/`) SHALL each be created with mode `0750`.

#### Scenario: resource subdirectories are 0750

- **WHEN** `germinator library init` creates the library
- **THEN** `skills/`, `agents/`, `commands/`, `memory/` SHALL each exist with mode `0750`

### Requirement: Single-writer contract

The library SHALL be designed for single-writer usage (one user, one process at a time on a given library directory). Concurrent writes from multiple processes SHALL NOT be supported without external coordination.

#### Scenario: Two processes writing concurrently

- **GIVEN** two `germinator` processes modify the same library directory simultaneously
- **WHEN** both call `Library.Save()` in overlapping windows
- **THEN** the final `library.yaml` SHALL reflect whichever `Save()`'s `os.Rename` ran last (last-writer-wins, no merging)
- **AND** no corruption SHALL occur (each save is a complete snapshot, atomic rename prevents torn writes)

### Requirement: Optional cross-process locking (advanced)

For multi-writer scenarios (parallel CI, shared library on a network filesystem), the library MAY be wrapped in an advisory file lock via `github.com/gofrs/flock`. The lock file SHALL live at `<library-path>/.lock` (or a path the user provides). Locking is opt-in; the single-writer contract is the default.

#### Scenario: flock-protected library

- **GIVEN** the library is configured with `flock`-based locking
- **WHEN** process A acquires the lock and begins a save
- **THEN** process B's `TryLock` SHALL return `false`
- **AND** process B SHALL either wait, return an error, or proceed without locking (per configuration)

#### Scenario: stale lock cleanup

- **GIVEN** process A acquired the lock then crashed without releasing it
- **WHEN** the next `germinator` invocation acquires the lock
- **THEN** `gofrs/flock` SHALL detect the stale lock and grant the new acquisition (default behavior of `flock.New.TryLock`)
- **AND** no manual intervention SHALL be required

### Requirement: Lock-free default

The default library implementation SHALL NOT acquire any lock. Wrapping in `flock` is opt-in via configuration or explicit user action.

#### Scenario: default save is lock-free

- **WHEN** `Library.Save()` is called with default configuration (no lock option set)
- **THEN** no `flock.New(...)` call SHALL be made
- **AND** the save SHALL complete via the atomic rename pattern alone

### Requirement: Unix-only permission bits

The library package uses Unix permission bits. On Windows, these bits are silently ignored by `os.Chmod`; the resulting file permissions follow the Windows default. Windows support is out of scope; the limitation SHALL be documented at each call site.

The pre-existing permission values are preserved by this change (no behavior change on Unix/macOS):

- Directories: `0o750` (set via `os.MkdirAll` at `internal/library/adder.go:105`, `internal/library/creator.go:57`, `internal/library/saver.go:21`).
- `library.yaml` initial creation via `CreateLibrary`: `0o644` (`internal/library/creator.go:65`).
- `library.yaml` incremental mutations (via `atomicWriteFile` from `AddResource`, `RemoveResource`, `RemovePreset`, `SaveLibrary`): `0o600`.
- Resource files: `0o644` (`internal/library/adder.go:124`).

The split between `CreateLibrary` (`0o644`) and the mutation path via `atomicWriteFile` (`0o600`) for `library.yaml` is pre-existing behavior; this requirement documents both without prescribing unification. Unifying them is out of scope for this change.

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
