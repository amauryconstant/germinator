# Tasks — Migrate config init and config validate

**Slice 8 of 9.** Slims `cmd/config.go` to the parent only and migrates `config init` and `config validate` to flat per-command files (`cmd/config_init.go`, `cmd/config_validate.go`, `package cmd`). Renames legacy `--output` flag to `--output-path`.

Each task ends with `mise run check` passing.

## 8.1 Slim `cmd/config.go` to the parent only

- [x] 8.1.1 Remove the `init`/`validate` constructors and the `scaffoldedConfig` template from `cmd/config.go`, leaving only `NewConfigCommand`; keep `package cmd`
- [x] 8.1.2 Stop discarding the Factory in `NewConfigCommand` (currently `NewConfigCommand(_ *cmdutil.Factory)`) and forward it: `cmd.AddCommand(NewCmdConfigInit(f, nil))` / `cmd.AddCommand(NewCmdConfigValidate(f, nil))`

## 8.2 Migrate `cmd/config_init.go`

- [x] 8.2.1 Create `cmd/config_init.go` (`package cmd`)
- [x] 8.2.2 Define `configInitOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`, `Force bool`
- [x] 8.2.3 Implement `NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command`:
  - Add `--output-path` (string; renamed from legacy `--output`)
  - Add `--force` flag
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runConfigInit(opts)`
- [x] 8.2.4 Implement `runConfigInit(opts *configInitOptions) error`:
  - Resolve `opts.OutputPath` (default `$XDG_CONFIG_HOME/germinator/config.toml` or `~/.config/germinator/config.toml`)
  - Check if file exists; if so and `--force` not set, return `core.NewFileError(opts.OutputPath, "create", "config file already exists (use --force to overwrite)", nil)` (constructor — `FileError` fields are unexported)
  - Write the config template (the `scaffoldedConfig` constant moved out of `cmd/config.go`) to the file
  - Write a single success line to `opts.IO.Out`; nothing on `opts.IO.ErrOut`
- [x] 8.2.5 Add golden file test in `cmd/config_init_test.go` confirming byte-identical output
- [x] 8.2.6 Run `mise run check`

## 8.3 Migrate `cmd/config_validate.go`

- [x] 8.3.1 Create `cmd/config_validate.go` (`package cmd`)
- [x] 8.3.2 Define `configValidateOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`, `OutputPath string`
- [x] 8.3.3 Implement `NewCmdConfigValidate(f, runF)` and `runConfigValidate(opts)`:
  - Add `--output-path` flag (no `--force`, no `--output` format flag)
  - Read the file at `opts.OutputPath`; on missing file return `core.NewFileError(opts.OutputPath, "read", "config file not found", statErr)` (not-found is derived from the wrapped `os.Stat` cause via `FileError.IsNotFound()`)
  - Parse with `koanf`; delegate to `config.Validate()` (platform-only today)
  - **Return** the error — do NOT call `output.FormatError` inside the command (single-handling rule; `main.go:41` renders once)
  - On success, write a single line to `opts.IO.Out`; nothing on `opts.IO.ErrOut`
- [x] 8.3.4 Add tests in `cmd/config_validate_test.go`: valid config, invalid platform value, malformed TOML, missing file
- [x] 8.3.5 Run `mise run check`

## 8.4 Update tests

- [x] 8.4.1 Move `cmd/config_test.go` content into `cmd/config_init_test.go` and `cmd/config_validate_test.go`; convert to `iostreams.Test()` + `runF` injection
- [x] 8.4.2 Add coverage for the `--output-path` rename: assert old `--output` returns a usage error
- [x] 8.4.3 Run `mise run check`

## 8.5 Verification

- [x] 8.5.1 Run `mise run lint` — confirm no new violations
- [x] 8.5.2 Run `mise run test` — confirm all unit tests pass
- [x] 8.5.3 Run `mise run build` — confirm `bin/germinator` builds
- [x] 8.5.4 Run `mise run test:coverage` — confirm coverage for the config commands (`cmd/config.go`, `cmd/config_init.go`, `cmd/config_validate.go`) ≥ 70%
- [x] 8.5.5 Run `mise run test:e2e` — confirm new E2E tests for config pass (`test/e2e/config_init_test.go`, `test/e2e/config_validate_test.go`)
- [x] 8.5.6 Smoke-test:
  - `germinator config init`
  - `germinator config init --output-path /tmp/test-config.toml`
  - `germinator config init --output-path /tmp/test-config.toml --force` (overwrite)
  - `germinator config init --output /tmp/test-config.toml` (should fail with usage error)
  - `germinator config validate`
  - `germinator config validate --output-path /tmp/test-config.toml`
  - `germinator config validate --output-path /nonexistent.toml` (should fail with FileError)
- [x] 8.5.7 Manually verify byte-identical output for `germinator config init` against a pre-change build
- [x] 8.5.8 Regression check — run and confirm exit code 0 + expected subcommand tree for:
  - `germinator --help`
  - `germinator config --help` (shows `init`, `validate`)
  - `germinator adapt --help`
  - `germinator canonicalize --help`
  - `germinator validate --help`
  - `germinator library --help` (shows existing subcommands)
