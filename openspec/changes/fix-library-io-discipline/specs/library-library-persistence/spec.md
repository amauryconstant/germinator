# library-library-persistence Specification (delta)

## MODIFIED Requirements

### Requirement: Atomic library.yaml writes

All mutations to `library.yaml` SHALL be performed atomically via the write-temp-then-rename pattern:

1. Write new content to `library.yaml.tmp`
2. `fsync` the temp file
3. `os.Rename(tmp, library.yaml)`

If `os.Rename` fails with `syscall.EXDEV` (cross-device link, e.g., temp file on `/tmp` and target on `/home`), the writer SHALL fall back to a copy-then-remove sequence: `io.Copy(tmp, library.yaml)` to write the new content into `library.yaml`, then `os.Remove(tmp)`. The fallback is atomic-or-fail at the user-observable level: the new `library.yaml` is fully written before the old temp is removed.

**Change**: ADDED the `EXDEV` fallback clause. The pre-change implementation used `os.Rename` directly; cross-filesystem scenarios (e.g., `TMPDIR=/tmp`, library on `/home`) failed with a silent error. The fallback is the standard cross-filesystem rename pattern.

#### Scenario: Successful atomic save on same filesystem

- **WHEN** `Library.Save()` writes new metadata and the temp file is on the same filesystem as `library.yaml`
- **THEN** the new content SHALL be written to `library.yaml.tmp` first
- **AND** on successful close, `os.Rename` SHALL atomically replace `library.yaml`
- **AND** no partial data SHALL ever be visible at `library.yaml` (no torn writes)

#### Scenario: Cross-filesystem save uses copy-then-remove fallback

- **GIVEN** `TMPDIR` resolves to a different filesystem from the library directory (e.g., `TMPDIR=/tmp` and library on `/home/user/lib`)
- **WHEN** `Library.Save()` writes new metadata
- **AND** the initial `os.Rename` returns `syscall.EXDEV`
- **THEN** the writer SHALL fall back to `io.Copy(tmp, library.yaml)` to write the new content
- **AND** on successful copy, `os.Remove(tmp)` SHALL remove the temp file
- **AND** the resulting `library.yaml` SHALL contain the new content

#### Scenario: Crash mid-write preserves prior content

- **GIVEN** the temp-file write is interrupted (process killed before `os.Rename` or `io.Copy`)
- **WHEN** `Library.Load()` is called on the library directory
- **THEN** the prior `library.yaml` SHALL be returned intact
- **AND** `library.yaml.tmp` MAY remain as a partial file (cleaned up on next successful save)

## ADDED Requirements

### Requirement: Unix-only permission bits

The library package uses Unix permission bits (`0o750` for directories, `0o640` for `library.yaml`, `0o644` for resource files). On Windows, these bits are silently ignored by `os.Chmod`; the resulting file permissions follow the Windows default. Windows support is out of scope; the limitation SHALL be documented at the call sites in `internal/library/adder.go:105`, `internal/library/creator.go:33`, and `internal/library/saver.go`.

**Change**: NEW requirement documenting a pre-existing limitation. The pre-change code did not document the platform behavior; a future contributor might add a Windows fix without considering whether to claim Windows support.

#### Scenario: Permission bits on Unix/macOS

- **WHEN** `germinator library init` creates a library on Linux or macOS
- **THEN** directories SHALL be created with mode `0750`
- **AND** `library.yaml` SHALL be created with mode `0640`
- **AND** resource files SHALL be created with mode `0644`

#### Scenario: Permission bits on Windows

- **WHEN** `germinator library init` runs on Windows
- **THEN** the `os.Chmod` calls SHALL be no-ops (Windows ignores the bits)
- **AND** file permissions SHALL follow the Windows default
- **AND** the documentation comment at the call site SHALL state that Windows is out of scope
