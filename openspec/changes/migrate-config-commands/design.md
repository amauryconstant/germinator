# Design — Migrate config init and config validate

## Context

After change-7 (`migrate-library-rest`) deletes the legacy shell, the only remaining commands using the old pattern are the two config commands and the two shell commands (`completion`, `version`). Change-8 migrates the config commands; change-9 finishes the migration with `completion`, `version`, and documentation.

## Goals / Non-Goals

**Goals:**

- `cmd/config.go` is slimmed to the parent only; sub-commands move to flat `cmd/config_init.go` and `cmd/config_validate.go` (`package cmd`).
- Both `config init` and `config validate` follow the `NewCmdXxx(f, runF) + runXxx(opts)` pattern.
- The legacy `--output` flag is renamed to `--output-path` to disambiguate from the new `--output` format flag.
- The default config path resolution (`$XDG_CONFIG_HOME/germinator/config.toml` or `~/.config/germinator/config.toml`) is preserved.
- Config init produces byte-identical output to the pre-change build.

**Non-Goals:**

- Migrating `completion`, `version` — change-9.
- Adding `--output json` to config commands — they produce text output (TOML files or validation errors), not structured data.
- Changing the config file format or template — preserved as-is.

## Decisions

### 1. `--output` renamed to `--output-path`

**Choice**: Both `config init` and `config validate` accept `--output-path <file>` instead of `--output <file>`.

**Rationale**: the `--output` flag now belongs to the `cli-output-formats` capability (`--output json|table|plain`); reusing it for a file path would be confusing. Renaming to `--output-path` makes the file path explicit.

**Alternatives considered**:
- *Keep `--output` and add `--output-format` for the format flag* — leaks the format-vs-path overlap into help text and confuses users reading scripts.
- *Use `--config-file` instead* — accurate for `validate`, but misleading for `init` (which writes the file rather than selecting one). `--output-path` works for both verbs.

### 2. Config template is moved into the init command file

**Choice**: The config template (the `scaffoldedConfig` multi-line TOML string, currently in `cmd/config.go:22-53`) moves into `cmd/config_init.go` as a private constant.

**Rationale**: matches the "implementations live next to the caller" principle; the template is small and rarely changes. (Note: the template is NOT in `internal/config/config.go` — that file holds only the `Config`/`CompletionConfig` structs, `Validate`, and `ExpandPaths`.)

**Alternatives considered**:
- *Keep template in `internal/config/`* — would couple `cmd/config_init.go` to a private constant in another package; either leaks the template via an exported function or forces an awkward getter.
- *Extract to a new `internal/templates/` package* — over-engineering for a single multi-line string; not worth the package boundary yet.

### 3. Config validation delegates to `config.Validate()` and returns the error

**Choice**: `runConfigValidate` reads the file, parses with `koanf`, and delegates to the existing `config.Validate()` (platform-only today). It **returns** the error; rendering is `main.go`'s job (`main.go:41` calls `output.FormatError` once), per the single-handling rule (errors are either rendered OR returned, never both).

**Rationale**: matches the rest of the codebase's error boundary. Calling `output.FormatError` inside the command AND returning the error would double-print (main.go always re-renders the returned error, and `root.go` sets `SilenceErrors: true`). Scope stays platform-only for now; if per-field collect-all is added later, two things must change: (a) `runConfigValidate` composes the result via `errors.Join`, and (b) `output.FormatError` must be enhanced to iterate `errors.Join` unwrap chains — it currently renders only the first typed error (`internal/output/errors.go:31-50`).

**Alternatives considered**:
- *Render each error inside the command via `output.FormatError` and also return it* — violates the single-handling rule and double-prints; the existing `cmd/validate.go:120-124` has this latent smell and this change should not replicate it.
- *Reuse `internal/parser` pipeline* — that pipeline validates `Resource` documents, not TOML config files; would force awkward adaptations.
- *Use a third-party TOML schema validator (e.g., `go-playground/validator`)* — adds a dependency for one command; not warranted while scope is platform-only.

### 4. Commands stay flat in `cmd/` (`package cmd`)

**Choice**: `cmd/config.go` is slimmed to the parent command only; sub-commands live in `cmd/config_init.go` and `cmd/config_validate.go`, all `package cmd`. This matches the flat layout every other migrated group uses (`library.go` + `library_init.go` + `library_add.go` + …, `init.go`, `adapt.go`, `validate.go`).

**Rationale**: consistency with the rest of `cmd/`. The `golang-project-layout` rule "packages MUST match their directory name" also rules out a `cmd/config/` sub-directory: `package config` would collide with the existing `internal/config` import (forcing an alias at every call site), while `package cmd` inside `cmd/config/` contradicts the match-directory convention.

**Alternatives considered**:
- *`cmd/config/` sub-directory (`config/root.go` + `init.go` + `validate.go`)* — introduces a layout no migrated group uses; creates the package-name collision described above.
- *Single-file `cmd/config.go` containing all three* — works for very small commands but the parent + two sub-commands structure already warrants one file per command, matching `library_*`.

### 5. "File already exists" uses `*core.FileError`, not `*core.ConfigError`

**Choice**: When `config init` targets an existing file and `--force` is not set, return `core.NewFileError(opts.OutputPath, "create", "config file already exists (use --force to overwrite)", nil)`. (Use the constructor — `FileError` fields are unexported, so composite literals like `*core.FileError{Path: ...}` do not compile.)

**Rationale**: matches the established precedent in `internal/library/creator.go` where `library init` reports an existing-target condition via `core.NewFileError(opts.Path, "create", ...)`. `core.FileError` already carries the path, operation, and a wrapped cause; `output.FormatError` (`internal/output/errors.go`) dispatches it correctly. Using `*core.ConfigError` would be misleading — it signals "a field in this configuration is invalid", not "this filesystem path is occupied".

## Risks / Trade-offs

- **BREAKING rename of `--output`** — script breakage. **Mitigation:** CHANGELOG entry only. The deprecation canary does NOT cover this (unknown flags → Cobra usage error → exit 2; the canary fires only on exit 1).
- **Config file format unchanged** — but tests must still byte-compare. **Mitigation:** golden file test in task 8.2.5.
- **No new dependencies** — uses existing `koanf` and `pflag`.
