## ADDED Requirements

### Requirement: validateOpenCodeAgent: mode values, temperature range, maxSteps constraint
### Requirement: validateOpenCodeCommand: template field required
### Requirement: validateOpenCodeSkill: name regex format, content required
### Requirement: validateOpenCodeMemory: paths or content required
### Requirement: All validation functions return []error
### Requirement: Validation errors are descriptive and helpful

#### Scenario: Agent mode validation - valid primary
- **GIVEN** Agent with mode="primary"
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent mode validation - valid subagent
- **GIVEN** Agent with mode="subagent"
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent mode validation - valid all
- **GIVEN** Agent with mode="all"
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent mode validation - invalid value
- **GIVEN** Agent with mode="invalid-mode"
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns error "mode must be one of: primary, subagent, all"

#### Scenario: Agent mode validation - empty value allowed
- **GIVEN** Agent with mode=""
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (mode is optional)
- **AND** Template will default to "all"

#### Scenario: Agent temperature validation - valid range
- **GIVEN** Agent with temperature=0.5
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent temperature validation - boundary 0.0
- **GIVEN** Agent with temperature=0.0
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent temperature validation - boundary 1.0
- **GIVEN** Agent with temperature=1.0
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent temperature validation - too high
- **GIVEN** Agent with temperature=1.5
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns error "temperature must be between 0.0 and 1.0"

#### Scenario: Agent temperature validation - negative
- **GIVEN** Agent with temperature=-0.5
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns error "temperature must be between 0.0 and 1.0"

#### Scenario: Agent temperature validation - not set
- **GIVEN** Agent with temperature=0.0 (Go zero value)
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (temperature is optional)

#### Scenario: Agent maxSteps validation - valid
- **GIVEN** Agent with maxSteps=50
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent maxSteps validation - minimum 1
- **GIVEN** Agent with maxSteps=1
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (no error)

#### Scenario: Agent maxSteps validation - too low
- **GIVEN** Agent with maxSteps=0
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns error "maxSteps must be >= 1"

#### Scenario: Agent maxSteps validation - negative
- **GIVEN** Agent with maxSteps=-5
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns error "maxSteps must be >= 1"

#### Scenario: Agent maxSteps validation - not set
- **GIVEN** Agent with maxSteps=0 (Go zero value)
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation passes (maxSteps is optional)

#### Scenario: Agent multiple validation errors
- **GIVEN** Agent with mode="invalid", temperature=2.0, maxSteps=0
- **WHEN** validateOpenCodeAgent is called
- **THEN** Validation returns []error with all three errors
- **AND** Each error is descriptive

#### Scenario: Command template validation - content present
- **GIVEN** Command with content="npm test"
- **WHEN** validateOpenCodeCommand is called
- **THEN** Validation passes (no error)

#### Scenario: Command template validation - empty
- **GIVEN** Command with content=""
- **WHEN** validateOpenCodeCommand is called
- **THEN** Validation returns error "template is required"

#### Scenario: Skill name validation - valid simple
- **GIVEN** Skill with name="git-workflow"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation passes (no error)
- **AND** Regex ^[a-z0-9]+(-[a-z0-9]+)*$ matches

#### Scenario: Skill name validation - valid multiple hyphens
- **GIVEN** Skill with name="code-review-tool-enhanced"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation passes (no error)
- **AND** Multiple hyphens are allowed (non-consecutive)

#### Scenario: Skill name validation - valid with numbers
- **GIVEN** Skill with name="git2-operations"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation passes (no error)
- **AND** Numbers are allowed

#### Scenario: Skill name validation - invalid consecutive hyphens
- **GIVEN** Skill with name="git--workflow"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error about invalid name format
- **AND** Consecutive hyphens not allowed

#### Scenario: Skill name validation - invalid starting hyphen
- **GIVEN** Skill with name="-git-workflow"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error about invalid name format
- **AND** Cannot start with hyphen

#### Scenario: Skill name validation - invalid ending hyphen
- **GIVEN** Skill with name="git-workflow-"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error about invalid name format
- **AND** Cannot end with hyphen

#### Scenario: Skill name validation - invalid uppercase
- **GIVEN** Skill with name="Git-Workflow"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error about invalid name format
- **AND** Uppercase not allowed

#### Scenario: Skill name validation - invalid special chars
- **GIVEN** Skill with name="git_workflow"
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error about invalid name format
- **AND** Underscores not allowed

#### Scenario: Skill content validation - present
- **GIVEN** Skill with content="Provides git operations..."
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation passes (no error)

#### Scenario: Skill content validation - empty
- **GIVEN** Skill with content=""
- **WHEN** validateOpenCodeSkill is called
- **THEN** Validation returns error "content is required"

#### Scenario: Memory validation - paths only
- **GIVEN** Memory with paths=["README.md"], content=""
- **WHEN** validateOpenCodeMemory is called
- **THEN** Validation passes (no error)
- **AND** Paths present is sufficient

#### Scenario: Memory validation - content only
- **GIVEN** Memory with paths=[], content="Project context..."
- **WHEN** validateOpenCodeMemory is called
- **THEN** Validation passes (no error)
- **AND** Content present is sufficient

#### Scenario: Memory validation - both paths and content
- **GIVEN** Memory with paths=["README.md"], content="Project context..."
- **WHEN** validateOpenCodeMemory is called
- **THEN** Validation passes (no error)
- **AND** Both fields present is valid

#### Scenario: Memory validation - both empty
- **GIVEN** Memory with paths=[], content=""
- **WHEN** validateOpenCodeMemory is called
- **THEN** Validation returns error "paths or content is required"
