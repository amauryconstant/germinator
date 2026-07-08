# cli-output-formats Specification (delta)

## MODIFIED Requirements

### Requirement: AddOutputFlags helper

The `output.AddOutputFlags` function SHALL add a `--output` string flag to a Cobra command.

**Change**: rehome the function from `internal/cmdutil` to `internal/output`. The previous `cmdutil.AddOutputFlags` re-export (at `internal/cmdutil/output_flags.go`) was deleted in change `remove-cmdutil-output-reexport` because the re-export covered only 1 of ~6 `output` symbols consumed by cmd files; every cmd file already imports `internal/output` for the other symbols, so the re-export provided no convenience.

#### Scenario: Flag registration

- **WHEN** `output.AddOutputFlags(cmd, &opts.Output)` is called
- **THEN** `cmd` SHALL have a `--output` string flag with default value `"plain"`
- **AND** valid values are `json`, `table`, `plain`
- **AND** shell completion is wired for those three values via `cobra.RegisterFlagCompletionFunc`

#### Scenario: Default value

- **WHEN** the `--output` flag is not provided on the command line
- **THEN** `opts.Output` SHALL be `"plain"` after flag parsing

#### Scenario: Function location is internal/output

- **WHEN** the codebase is searched for `AddOutputFlags`
- **THEN** it SHALL be defined in `internal/output/output_flags.go`
- **AND** it SHALL NOT be re-exported by `internal/cmdutil` (the previous re-export was deleted)

### Requirement: PersistentFlags variant for parent commands

The `output.AddOutputFlags` function SHALL accept a `*cobra.Command` parameter; for parent commands that need inherited flags, callers SHALL use `cmd.PersistentFlags()` directly instead of `cmd.Flags()`.

**Change**: NEW requirement documenting the limitation of `output.AddOutputFlags` (it binds to `cmd.Flags()`, which is local-only). Library subcommands that need inherited flags (e.g., `library remove`) MUST wire the `--output` flag manually via `cmd.PersistentFlags()`.

#### Scenario: PersistentFlags wiring for library remove

- **WHEN** a parent command like `library remove` needs the `--output` flag to be inherited by its subcommands (`resource`, `preset`)
- **THEN** the parent SHALL wire the flag manually using `cmd.PersistentFlags().StringVar(&outputFormat, "output", "plain", "Output format")` and `cmd.RegisterFlagCompletionFunc`
- **AND** it SHALL NOT call `output.AddOutputFlags(cmd, ...)` (which binds to local `cmd.Flags()`)
