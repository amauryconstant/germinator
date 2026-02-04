## ADDED Requirements

### Requirement: validateOpenCodeAgent validates OpenCode-specific constraints
Validates Agent mode values, temperature range, and steps constraint.

#### Scenario: Agent mode validation
- **GIVEN** An Agent with mode set
- **WHEN** validateOpenCodeAgent is called
- **THEN** mode="primary", mode="subagent", mode="all" SHALL pass validation
- **AND** mode="" (empty) SHALL pass (mode is optional, defaults to "all" in template)
- **AND** mode="invalid" SHALL return error "mode must be one of: primary, subagent, all"

#### Scenario: Agent temperature validation
- **GIVEN** An Agent with Temperature (*float64 pointer)
- **WHEN** validateOpenCodeAgent is called
- **THEN** Temperature=nil SHALL pass validation (optional field)
- **AND** Temperature=0.0 SHALL pass (valid deterministic value)
- **AND** Temperature=0.5 SHALL pass
- **AND** Temperature=1.0 SHALL pass (max randomness)
- **AND** Temperature=-0.5 SHALL fail with "temperature must be between 0.0 and 1.0"
- **AND** Temperature=1.5 SHALL fail with "temperature must be between 0.0 and 1.0"

#### Scenario: Agent maxSteps validation
- **GIVEN** An Agent with MaxSteps field
- **WHEN** validateOpenCodeAgent is called
- **THEN** MaxSteps=1 SHALL pass validation (minimum valid value)
- **AND** MaxSteps=50 SHALL pass validation
- **AND** MaxSteps=0 (Go zero value) SHALL pass (field is optional)
- **AND** MaxSteps=-5 SHALL fail with "maxSteps must be >= 1"

#### Scenario: Agent multiple validation errors
- **GIVEN** An Agent with mode="invalid", temperature=2.0, maxSteps=0
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation SHALL return []error with all three errors
- **AND** Each error SHALL be descriptive

### Requirement: validateOpenCodeCommand validates template field
Validates that Command template content is present.

#### Scenario: Command template validation
- **GIVEN** A Command model
- **WHEN** validateOpenCodeCommand is called
- **THEN** content="npm test" SHALL pass validation
- **AND** content="" (empty string) SHALL return error "template is required"

### Requirement: validateOpenCodeSkill validates OpenCode-specific constraints
Validates Skill name format (kebab-case) and requires content.

#### Scenario: Skill name validation
- **GIVEN** A Skill with Name field
- **WHEN** validateOpenCodeSkill is called
- **THEN** name="git-workflow" SHALL pass (simple valid)
- **AND** name="code-review-tool-enhanced" SHALL pass (multiple hyphens, non-consecutive)
- **AND** name="git2-operations" SHALL pass (numbers allowed)
- **AND** name="git--workflow" SHALL fail with error about invalid format (consecutive hyphens)
- **AND** name="-git-workflow" SHALL fail (leading hyphen)
- **AND** name="git-workflow-" SHALL fail (trailing hyphen)
- **AND** name="Git-Workflow" SHALL fail (uppercase)
- **AND** name="git_workflow" SHALL fail (underscores)
- **AND** Regex SHALL be `^[a-z0-9]+(-[a-z0-9]+)*$`

#### Scenario: Skill content validation
- **GIVEN** A Skill model
- **WHEN** validateOpenCodeSkill is called
- **THEN** content="Provides git operations..." SHALL pass validation
- **AND** content="" SHALL return error "content is required"

### Requirement: validateOpenCodeMemory validates paths or content presence
Validates that Memory has at least one of paths or content populated.

#### Scenario: Memory validation
- **GIVEN** A Memory model
- **WHEN** validateOpenCodeMemory is called
- **THEN** paths=["README.md"], content="" SHALL pass (paths present)
- **AND** paths=[], content="Project context..." SHALL pass (content present)
- **AND** paths=["README.md"], content="Project context..." SHALL pass (both present)
- **AND** paths=[], content="" SHALL fail with "paths or content is required"

### Requirement: Agent and Skill name regex uses correct character class
The regex pattern for validating Agent and Skill names SHALL use `[a-z0-9-]` character class to properly allow hyphens as separators while preventing consecutive hyphens, leading hyphens, and trailing hyphens.

#### Scenario: Regex character class and pattern
- **GIVEN** Name validation regex
- **WHEN** Pattern is defined
- **THEN** Character class SHALL be `[a-z0-9-]` (hyphen included in character class for segments)
- **AND** Pattern SHALL be `^[a-z0-9]+(-[a-z0-9]+)*$` (hyphens outside character class as separators)
- **AND** Agent and Skill validation SHALL both use this regex
