# Resource Installation

## ADDED Requirements

### Requirement: Install resources to target project

The system SHALL install resources from the library to a target project directory.

#### Scenario: Install single resource
- **GIVEN** a library with resource `skill/commit`
- **WHEN** InitializeResources is called with refs `["skill/commit"]` and platform `opencode`
- **THEN** the resource is loaded, transformed, and written to `.opencode/skills/commit/SKILL.md`

#### Scenario: Install multiple resources
- **GIVEN** a library with resources `skill/commit` and `skill/merge-request`
- **WHEN** InitializeResources is called with both refs
- **THEN** both resources are installed to their respective output paths

#### Scenario: Install resources from preset
- **GIVEN** a library with preset `git-workflow` containing `["skill/commit", "skill/merge-request"]`
- **WHEN** InitializeResources is called with preset name
- **THEN** all resources in the preset are installed

### Requirement: Derive platform-specific output paths

The system SHALL derive output paths from resource type, name, and platform.

#### Scenario: Derive output path for OpenCode skill
- **GIVEN** resource type `skill` and name `commit`
- **WHEN** GetOutputPath is called with platform `opencode`
- **THEN** path `.opencode/skills/commit/SKILL.md` is returned

#### Scenario: Derive output path for Claude Code skill
- **GIVEN** resource type `skill` and name `commit`
- **WHEN** GetOutputPath is called with platform `claude-code`
- **THEN** path `.claude/skills/commit/SKILL.md` is returned

#### Scenario: Derive output path for agent
- **GIVEN** resource type `agent` and name `reviewer`
- **WHEN** GetOutputPath is called with platform `opencode`
- **THEN** path `.opencode/agents/reviewer.md` is returned

#### Scenario: Derive output path for command
- **GIVEN** resource type `command` and name `test`
- **WHEN** GetOutputPath is called with platform `opencode`
- **THEN** path `.opencode/commands/test.md` is returned

#### Scenario: Derive output path for memory
- **GIVEN** resource type `memory` and name `context`
- **WHEN** GetOutputPath is called with platform `opencode`
- **THEN** path `.opencode/memory/context.md` is returned

### Requirement: Handle existing files

The system SHALL handle existing output files according to force flag.

#### Scenario: Handle existing file without force
- **GIVEN** an output path that already exists
- **WHEN** InitializeResources is called without `--force`
- **THEN** an error is returned indicating the file exists

#### Scenario: Overwrite existing file with force
- **GIVEN** an output path that already exists
- **WHEN** InitializeResources is called with `--force`
- **THEN** the existing file is overwritten

### Requirement: Support dry-run mode

The system SHALL preview changes without writing files in dry-run mode.

#### Scenario: Dry-run mode
- **GIVEN** dry-run mode enabled
- **WHEN** InitializeResources is called
- **THEN** output paths are printed but no files are written

### Requirement: Fail-fast on errors

The system SHALL stop processing on first error (fail-fast strategy).

#### Scenario: Fail-fast on error
- **GIVEN** resources `["skill/commit", "skill/invalid", "skill/merge-request"]`
- **WHEN** InitializeResources is called and `skill/invalid` fails
- **THEN** processing stops at the error, `skill/merge-request` is not processed

### Requirement: Create output directories

The system SHALL create parent directories as needed.

#### Scenario: Create output directories
- **GIVEN** output path `.opencode/skills/commit/SKILL.md` where `.opencode/skills/` does not exist
- **WHEN** InitializeResources is called
- **THEN** parent directories are created before writing

### Requirement: Reuse existing transformation pipeline

The system SHALL use existing LoadDocument and RenderDocument for transformation.

#### Scenario: Reuse existing transformation pipeline
- **GIVEN** a canonical resource file
- **WHEN** InitializeResources processes it
- **THEN** the existing LoadDocument and RenderDocument functions are used for transformation
