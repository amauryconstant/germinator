# Init Command

## ADDED Requirements

### Requirement: Install with explicit resources

The CLI SHALL install explicitly specified resources.

#### Scenario: Install with explicit resources
- **GIVEN** a library with resources `skill/commit` and `skill/merge-request`
- **WHEN** `germinator init --platform opencode --resources skill/commit,skill/merge-request` is run
- **THEN** both resources are installed to the current directory

### Requirement: Install with preset

The CLI SHALL install all resources from a preset.

#### Scenario: Install with preset
- **GIVEN** a library with preset `git-workflow`
- **WHEN** `germinator init --platform opencode --preset git-workflow` is run
- **THEN** all resources in the preset are installed

### Requirement: Validate required flags

The CLI SHALL require platform and either resources or preset flags.

#### Scenario: Require platform flag
- **GIVEN** no `--platform` flag
- **WHEN** `germinator init --resources skill/commit` is run
- **THEN** an error indicates `--platform` is required

#### Scenario: Require resources or preset
- **GIVEN** no `--resources` or `--preset` flag
- **WHEN** `germinator init --platform opencode` is run
- **THEN** an error indicates either `--resources` or `--preset` is required

#### Scenario: Reject both resources and preset
- **GIVEN** both `--resources` and `--preset` flags
- **WHEN** `germinator init --platform opencode --resources skill/commit --preset git-workflow` is run
- **THEN** an error indicates flags are mutually exclusive

### Requirement: Support custom library path

The CLI SHALL accept a custom library path via flag.

#### Scenario: Custom library path
- **GIVEN** a library at `/custom/library`
- **WHEN** `germinator init --platform opencode --resources skill/commit --library /custom/library` is run
- **THEN** resources are loaded from the custom library

### Requirement: Support custom output directory

The CLI SHALL accept a custom output directory via flag.

#### Scenario: Custom output directory
- **GIVEN** output directory `/target/project`
- **WHEN** `germinator init --platform opencode --resources skill/commit --output /target/project` is run
- **THEN** resources are installed to `/target/project/.opencode/skills/commit/SKILL.md`

### Requirement: Support dry-run preview

The CLI SHALL support dry-run mode to preview changes.

#### Scenario: Dry-run preview
- **GIVEN** dry-run mode
- **WHEN** `germinator init --platform opencode --resources skill/commit --dry-run` is run
- **THEN** output shows what would be written without creating files

### Requirement: Support force overwrite

The CLI SHALL support force flag to overwrite existing files.

#### Scenario: Force overwrite
- **GIVEN** existing output files
- **WHEN** `germinator init --platform opencode --resources skill/commit --force` is run
- **THEN** existing files are overwritten

### Requirement: Format success output

The CLI SHALL display a summary of installed resources on success.

#### Scenario: Success output
- **GIVEN** successful installation of resources
- **WHEN** init completes
- **THEN** a summary lists installed resources and their paths

### Requirement: Format error output

The CLI SHALL display errors with resource reference and file path.

#### Scenario: Error output with file paths
- **GIVEN** a resource that fails to load
- **WHEN** init encounters an error
- **THEN** the error message includes the resource reference and file path

### Requirement: Validate platform value

The CLI SHALL validate the platform flag value.

#### Scenario: Validate platform value
- **GIVEN** an invalid platform `invalid-platform`
- **WHEN** `germinator init --platform invalid-platform --resources skill/commit` is run
- **THEN** an error indicates valid platforms are `opencode` and `claude-code`
