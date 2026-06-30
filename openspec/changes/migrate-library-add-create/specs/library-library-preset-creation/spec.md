# library-library-preset-creation Specification (delta)

> **Behavior preserved, validation moved earlier.** The migration of `library create preset` adds `core.CanInstallResource` as a pre-flight ref validator. This is a behavioral addition: each ref in `opts.Resources` is now validated before any I/O, and malformed refs produce `*core.ValidationError` instead of a generic error.

## MODIFIED Requirements

### Requirement: Pre-flight ref validation via core.CanInstallResource

When `library create preset` is invoked, the command SHALL validate each ref in `opts.Resources` via `core.CanInstallResource(ref)` before calling `library.CreatePreset`. If any ref fails validation, the command SHALL return `*core.ValidationError` (mapped to exit 1 by `cmdutil.ExitCodeFor` via the default-error case at `internal/cmdutil/exit.go:77`).

> New requirement added in slice-6. The existing scenarios for "Create preset with valid resources", "Validate referenced resources exist", "Reject preset name with whitespace", "Reject duplicate preset without --force" all remain valid.

#### Scenario: Valid refs pass pre-flight validation

- **GIVEN** `--resources skill/commit,agent/reviewer`
- **WHEN** `germinator library create preset dev-setup --resources skill/commit,agent/reviewer` is invoked
- **THEN** `core.CanInstallResource("skill/commit")` SHALL return `nil`
- **AND** `core.CanInstallResource("agent/reviewer")` SHALL return `nil`
- **AND** the preset is created

#### Scenario: First malformed ref fails pre-flight validation

- **GIVEN** `--resources skills/commit,agent/reviewer` (invalid type `skills`)
- **WHEN** `germinator library create preset dev-setup --resources skills/commit,agent/reviewer` is invoked
- **THEN** `core.CanInstallResource("skills/commit")` SHALL return a non-nil `*core.ValidationError`
- **AND** `output.FormatError` SHALL render `Error: ref type must be one of skill, agent, command, memory\n` to stderr
- **AND** no preset is created (no library.yaml update)

#### Scenario: Second malformed ref (after valid first) still fails

- **GIVEN** `--resources skill/commit,agent/` (empty name in second ref)
- **WHEN** `germinator library create preset dev-setup --resources skill/commit,agent/` is invoked
- **THEN** `core.CanInstallResource("skill/commit")` SHALL return `nil`
- **THEN** `core.CanInstallResource("agent/")` SHALL return a non-nil `*core.ValidationError`
- **AND** the preset creation is aborted before `CreatePreset` is called

#### Scenario: Empty resources flag fails pre-flight validation

- **GIVEN** `--resources ""` (empty string)
- **WHEN** `germinator library create preset dev-setup --resources ""` is invoked
- **THEN** the command SHALL return a usage error (exit 2)
- **AND** `output.FormatError` SHALL render the usage error to stderr
- **AND** no preset is created

> **Why this is a delta, not a new capability**: same rationale as the `library-library-resource-import` delta above. `core.CanInstallResource` is a private helper; its call-site behavior in `library create preset` belongs to the `library-library-preset-creation` capability.
