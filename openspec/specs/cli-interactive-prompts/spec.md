# cli-interactive-prompts Specification

## Purpose

Define the "flags-first, prompt-as-fallback" policy for interactive input: every command accepts its inputs as flags; if a required flag is missing AND stdin is a TTY, the CLI prompts interactively. In non-interactive contexts (CI, pipes, redirects) the prompt is skipped and the missing flag is treated as a usage error.

> **Library:** `github.com/charmbracelet/huh` for forms and prompts.

## Requirements

### Requirement: Flags-first input model

Every command SHALL accept its inputs via flags. Required flags SHALL be marked with `MarkFlagRequired` (so Cobra emits a usage error if missing), with one exception: commands that support interactive prompting (e.g., `germinator init` — but only when `IOStreams.IsInteractive()` is true) MAY leave a flag optional and prompt for it.

#### Scenario: Required flag without prompt

- **GIVEN** command `adapt` requires `--platform`
- **WHEN** `germinator adapt input.md output.md` is invoked (no `--platform`)
- **THEN** Cobra SHALL emit a "required flag(s) "platform" not set" usage error
- **AND** exit code SHALL be 2

#### Scenario: Prompt only on TTY

- **GIVEN** command `init` optionally prompts for `--resources`
- **AND** `IOStreams.IsInteractive() == true`
- **WHEN** `germinator init --platform opencode` is invoked without `--resources` or `--preset`
- **THEN** the CLI SHALL print an interactive prompt for resource refs
- **AND** on user input, the command SHALL proceed with the entered values

### Requirement: huh-based prompts

Interactive prompts SHALL be implemented via `github.com/charmbracelet/huh`. Prompts are rendered to `ErrOut` (not `Out`) so stdout stays clean for piped output.

#### Scenario: huh Form on stderr

- **GIVEN** a `huh.Form` is constructed for resource input
- **WHEN** `form.WithOutput(opts.IO.ErrOut).Run()` is called
- **THEN** the form SHALL render to `ErrOut`
- **AND** `Out` SHALL NOT receive any prompt output

### Requirement: Non-interactive never prompts

When `IOStreams.IsInteractive()` returns false (stdin OR stdout is not a TTY), the CLI SHALL NOT attempt to prompt. Missing required input SHALL produce a clear usage error mentioning the relevant `--flag`.

#### Scenario: Non-interactive missing flag

- **GIVEN** stdin is not a TTY (piped from another command)
- **WHEN** `germinator init --platform opencode` is invoked without `--resources` or `--preset`
- **THEN** the CLI SHALL return `*core.NewValidationError("init", "resources/preset", "", "either --resources or --preset is required (or run interactively in a terminal)")`
- **AND** exit code SHALL be 2 (`ExitCodeUsage`)

### Requirement: --non-interactive / --yes bypass

Destructive or interactive commands SHALL accept `--force` or equivalent to bypass prompts. The exact flag name (`--force`, `--yes`, `--non-interactive`) is per-command; the policy is that *some* bypass flag exists for every interactive command.

#### Scenario: --yes skips prompts

- **GIVEN** command accepts `--yes` to bypass prompts
- **WHEN** `germinator init --platform opencode --resources skill/commit --yes` is invoked
- **THEN** no prompts SHALL be printed
- **AND** the command SHALL proceed with the flag values only

### Requirement: Prompt helpers

Shared prompt helpers SHALL live in `internal/output/prompt.go` (or co-located with the prompt user). Each helper SHALL gate on `io.IsInteractive()` and SHALL return a typed error if the prompt cannot be displayed.

The canonical helpers include:

- `promptConfirm(io, message) (bool, error)` — yes/no confirmation
- `promptString(io, title, placeholder) (string, error)` — single-line string input
- `promptMultiSelect(io, title, options) ([]string, error)` — multi-select from a list

#### Scenario: promptConfirm gates on TTY

- **GIVEN** an `IOStreams` from `iostreams.Test()` (non-interactive)
- **WHEN** `promptConfirm(io, "delete foo?")` is called
- **THEN** it SHALL return a typed error indicating the prompt cannot be displayed
- **AND** no input SHALL be read from stdin
