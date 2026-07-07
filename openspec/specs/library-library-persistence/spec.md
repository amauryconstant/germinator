# library-library-persistence Specification

## Purpose

Define the persistence guarantees for the library directory: atomic writes for `library.yaml`, file permissions on the directory tree, and the single-writer contract (with optional cross-process locking for advanced use cases).

## Requirements

### Requirement: Atomic library.yaml writes

All mutations to `library.yaml` SHALL be performed atomically via the write-temp-then-rename pattern:

1. Write new content to `library.yaml.tmp`
2. `fsync` the temp file
3. `os.Rename(tmp, library.yaml)`

A crash between steps 1 and 3 leaves the prior `library.yaml` intact.

#### Scenario: Successful atomic save

- **WHEN** `Library.Save()` writes new metadata
- **THEN** the new content SHALL be written to `library.yaml.tmp` first
- **AND** on successful close, `os.Rename` SHALL atomically replace `library.yaml`
- **AND** no partial data SHALL ever be visible at `library.yaml` (no torn writes)

#### Scenario: Crash mid-write preserves prior content

- **GIVEN** the temp-file write is interrupted (process killed before `os.Rename`)
- **WHEN** `Library.Load()` is called on the library directory
- **THEN** the prior `library.yaml` SHALL be returned intact
- **AND** `library.yaml.tmp` MAY remain as a partial file (cleaned up on next successful save)

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
