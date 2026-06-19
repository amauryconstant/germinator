# Tasks — Migrate config init and config validate

**Slice 8 of 9.** Splits `cmd/config.go` into a `cmd/config/` sub-directory. Migrates `config init` and `config validate` to the new pattern. Renames legacy `--output` flag to `--output-path`.

Each task ends with `mise run check` passing.

## 8.1 Split `cmd/config.go` into sub-directory

- [ ] 8.1.1 Create `cmd/config/` directory
- [ ] 8.1.2 Move `cmd/config.go` to `cmd/config/root.go` (parent command only; no flags on the parent)
- [ ] 8.1.3 Update the package declaration if needed (still `package cmd`)

## 8.2 Migrate `cmd/config/init.go`

- [ ] 8.2.1 Create `cmd/config/init.go`
- [ ] 8.2.2 Define `configInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`, `Force bool`
- [ ] 8.2.3 Implement `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command`:
  - Add `--output-path` (string; renamed from legacy `--output`)
  - Add `--force` flag
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runConfigInit(opts)`
- [ ] 8.2.4 Implement `runConfigInit(opts *configInitOptions) error`:
  - Resolve `opts.OutputPath` (default `$XDG_CONFIG_HOME/germinator/config.toml` or `~/.config/germinator/config.toml`)
  - Check if file exists; if so and `--force` not set, return `*core.ConfigError`
  - Write the config template (moved from `internal/config/config.go`) to the file
  - Print success message to `opts.IO.Out`
- [ ] 8.2.5 Add golden file test in `cmd/config/init_test.go` confirming byte-identical output
- [ ] 8.2.6 Run `mise run check`

## 8.3 Migrate `cmd/config/validate.go`

- [ ] 8.3.1 Create `cmd/config/validate.go`
- [ ] 8.3.2 Define `configValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`
- [ ] 8.3.3 Implement `NewCmdConfigValidate(f, runF)` and `runConfigValidate(opts)`:
  - Add `--output-path` flag (no `--force`, no `--output` format flag)
  - Read the file at `opts.OutputPath`
  - Parse with `koanf`; validate known fields
  - For each validation error, call `output.FormatError(opts.IO, err)`
  - Return the first typed error (or nil if all valid)
- [ ] 8.3.4 Add tests in `cmd/config/validate_test.go`: valid config, missing required field, invalid type, missing file
- [ ] 8.3.5 Run `mise run check`

## 8.4 Update tests

- [ ] 8.4.1 Move `cmd/config_test.go` content into `cmd/config/init_test.go` and `cmd/config/validate_test.go`; convert to `iostreams.Test()` + `runF` injection
- [ ] 8.4.2 Add coverage for the `--output-path` rename: assert old `--output` returns a usage error
- [ ] 8.4.3 Run `mise run check`

## 8.5 Verification

- [ ] 8.5.1 Run `mise run lint` — confirm no new violations
- [ ] 8.5.2 Run `mise run test` — confirm all unit tests pass
- [ ] 8.5.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 8.5.4 Run `mise run test:coverage` — confirm coverage for `cmd/config/` ≥ 70%
- [ ] 8.5.5 Run `mise run test:e2e` — confirm E2E tests for config pass
- [ ] 8.5.6 Smoke-test:
  - `germinator config init`
  - `germinator config init --output-path /tmp/test-config.toml`
  - `germinator config init --output-path /tmp/test-config.toml --force` (overwrite)
  - `germinator config init --output /tmp/test-config.toml` (should fail with usage error)
  - `germinator config validate`
  - `germinator config validate --output-path /tmp/test-config.toml`
  - `germinator config validate --output-path /nonexistent.toml` (should fail with FileError)
- [ ] 8.5.7 Manually verify byte-identical output for `germinator config init` against a pre-change build
- [ ] 8.5.8 Update `cmd/config/AGENTS.md` (create it if needed) with the new pattern
- [ ] 8.5.9 Confirm all other commands still work (regression check)
