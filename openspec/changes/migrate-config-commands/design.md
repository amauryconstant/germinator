# Design — Migrate config init and config validate

## Context

After change-7 (`migrate-library-rest`) deletes the legacy shell, the only remaining commands using the old pattern are the two config commands and the two shell commands (`completion`, `version`). Change-8 migrates the config commands; change-9 finishes the migration with `completion`, `version`, and documentation.

## Goals / Non-Goals

**Goals:**

- `cmd/config.go` is split into a sub-directory: `cmd/config/{root,init,validate}.go`.
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

**Rationale**: the `--output` flag now belongs to the `output-formats` capability (`--output json|table|plain`); reusing it for a file path would be confusing. Renaming to `--output-path` makes the file path explicit.

### 2. Config template is moved into the command file

**Choice**: The config template (a multi-line TOML string) moves from `internal/infrastructure/config/config.go` (already renamed to `internal/config/config.go` in change-1) into `cmd/config/init.go` as a private constant.

**Rationale**: matches the "implementations live next to the caller" principle; the template is small and rarely changes.

### 3. Config validation uses `koanf` directly

**Choice**: `runConfigValidate` reads the file, parses with `koanf`, and checks each known field's type. Validation errors are rendered via `output.FormatError`.

**Rationale**: `koanf` is already a dependency; explicit field validation gives clearer error messages than koanf's default behavior.

### 4. Config sub-directory uses `cmd/config/` convention

**Choice**: `cmd/config.go` becomes `cmd/config/root.go` (parent only); sub-commands live in `cmd/config/init.go` and `cmd/config/validate.go`. This matches the convention for groups with 2+ sub-commands (the proposal notes that even 2 commands may justify a sub-directory "because of the sub-grouping convention").

## Risks / Trade-offs

- **BREAKING rename of `--output`** — script breakage. **Mitigation:** CHANGELOG entry; one-version deprecation canary.
- **Config file format unchanged** — but tests must still byte-compare. **Mitigation:** golden file test in task 8.2.3.
- **No new dependencies** — uses existing `koanf` and `pflag`.
