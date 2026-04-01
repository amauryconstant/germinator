# Capability: Library JSON Output

## Purpose

Add a `--json` flag to the parent `germinator library` command for consistent JSON output across all library subcommands. The flag inherits to all subcommands via Cobra's persistent flag mechanism.

## ADDED Requirements

### Requirement: Library parent command accepts --json flag

The `germinator library` parent command SHALL accept a `--json` flag that is inherited by all subcommands.

#### Scenario: JSON flag available on library command
- **GIVEN** the `germinator library` command help output
- **WHEN** user runs `germinator library --help`
- **THEN** the `--json` flag is listed in the flags section

#### Scenario: JSON flag inherited by resources subcommand
- **GIVEN** the `germinator library resources` command
- **WHEN** user runs `germinator library resources --help`
- **THEN** the `--json` flag is available

#### Scenario: JSON flag inherited by presets subcommand
- **GIVEN** the `germinator library presets` command
- **WHEN** user runs `germinator library presets --help`
- **THEN** the `--json` flag is available

#### Scenario: JSON flag inherited by show subcommand
- **GIVEN** the `germinator library show` command
- **WHEN** user runs `germinator library show --help`
- **THEN** the `--json` flag is available

#### Scenario: JSON flag inherited by init subcommand
- **GIVEN** the `germinator library init` command
- **WHEN** user runs `germinator library init --help`
- **THEN** the `--json` flag is available

### Requirement: Library resources outputs JSON when --json is set

The `germinator library resources --json` command SHALL output JSON format when the `--json` flag is set.

#### Scenario: Resources JSON output structure
- **GIVEN** a library with resources
- **WHEN** user runs `germinator library resources --json`
- **THEN** output is valid JSON with structure: `{"resources": [{"type": "...", "name": "...", "description": "...", "platform": "..."}, ...]}`

#### Scenario: Resources JSON output with empty library
- **GIVEN** a library with no resources
- **WHEN** user runs `germinator library resources --json`
- **THEN** output is valid JSON: `{"resources": []}`

### Requirement: Library presets outputs JSON when --json is set

The `germinator library presets --json` command SHALL output JSON format when the `--json` flag is set.

#### Scenario: Presets JSON output structure
- **GIVEN** a library with presets
- **WHEN** user runs `germinator library presets --json`
- **THEN** output is valid JSON with structure: `{"presets": [{"name": "...", "description": "...", "resources": ["...", "..."]}, ...]}`

#### Scenario: Presets JSON output with empty library
- **GIVEN** a library with no presets
- **WHEN** user runs `germinator library presets --json`
- **THEN** output is valid JSON: `{"presets": []}`

### Requirement: Library show outputs JSON when --json is set

The `germinator library show <ref> --json` command SHALL output JSON format when the `--json` flag is set.

#### Scenario: Show resource JSON output structure
- **GIVEN** a library with resource `skill/commit`
- **WHEN** user runs `germinator library show skill/commit --json`
- **THEN** output is valid JSON with structure: `{"resource": {"type": "...", "name": "...", "description": "...", "path": "..."}}`

#### Scenario: Show preset JSON output structure
- **GIVEN** a library with preset `git-workflow`
- **WHEN** user runs `germinator library show preset/git-workflow --json`
- **THEN** output is valid JSON with structure: `{"preset": {"name": "...", "description": "...", "resources": ["...", "..."]}}`

#### Scenario: Show nonexistent resource returns error in JSON format
- **GIVEN** a library without resource `skill/nonexistent`
- **WHEN** user runs `germinator library show skill/nonexistent --json`
- **THEN** command returns error with exit code 6 and JSON error body

### Requirement: Library init outputs JSON when --json is set

The `germinator library init --json` command SHALL output JSON format when the `--json` flag is set.

#### Scenario: Init success JSON output
- **WHEN** user runs `germinator library init --path /tmp/test-lib --json`
- **THEN** output is valid JSON with structure: `{"success": true, "path": "...", "message": "..."}`

#### Scenario: Init failure JSON output
- **WHEN** user runs `germinator library init --path /existing/library --force --json`
- **THEN** if initialization fails, output is valid JSON with structure: `{"success": false, "error": "...", "path": "..."}`

### Requirement: JSON encoding uses proper formatting

All JSON output SHALL use `json.NewEncoder(c.OutOrStdout()).SetIndent("", "  ")` for formatted output with 2-space indentation.

#### Scenario: JSON output is pretty-printed
- **GIVEN** a library with resources
- **WHEN** user runs `germinator library resources --json`
- **THEN** the JSON output has 2-space indentation and is human-readable
