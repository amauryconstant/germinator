## Purpose

Provide standalone validator functions for Agent, Command, Skill, and Memory documents that can be composed into pipelines. Generic validators apply to all platforms; platform-specific validators live in subpackages.

## Requirements

### Requirement: Agent validators in validation package

The system SHALL provide standalone validator functions for Agent in `internal/validation/validators.go`.

#### Scenario: ValidateAgentName checks name is required

- **WHEN** `ValidateAgentName(agent)` is called with agent.Name == ""
- **THEN** it SHALL return an error Result with ValidationError

#### Scenario: ValidateAgentName checks name format

- **WHEN** `ValidateAgentName(agent)` is called with agent.Name == "Invalid Name"
- **THEN** it SHALL return an error Result with message about pattern

#### Scenario: ValidateAgentName passes valid name

- **WHEN** `ValidateAgentName(agent)` is called with agent.Name == "valid-name"
- **THEN** it SHALL return `NewResult(true)`

---

### Requirement: Composed ValidateAgent function

The system SHALL provide a `ValidateAgent(agent *canonical.Agent) Result[bool]` function that composes all agent validators.

#### Scenario: ValidateAgent runs all validators

- **WHEN** `ValidateAgent(agent)` is called
- **THEN** it SHALL run ValidateAgentName, ValidateAgentDescription, ValidateAgentPermissionPolicy
- **AND** it SHALL return early on first error

#### Scenario: ValidateAgent passes valid agent

- **WHEN** `ValidateAgent(agent)` is called with a fully valid agent
- **THEN** it SHALL return `NewResult(true)`

---

### Requirement: Command validators

The system SHALL provide standalone validator functions for Command.

#### Scenario: ValidateCommandName checks name is required

- **WHEN** `ValidateCommandName(command)` is called with command.Name == ""
- **THEN** it SHALL return an error Result

#### Scenario: ValidateCommand composes validators

- **WHEN** `ValidateCommand(command)` is called
- **THEN** it SHALL run ValidateCommandName, ValidateCommandDescription, ValidateCommandExecution

---

### Requirement: Skill validators

The system SHALL provide standalone validator functions for Skill.

#### Scenario: ValidateSkillName checks name format

- **WHEN** `ValidateSkillName(skill)` is called with invalid name format
- **THEN** it SHALL return an error Result

#### Scenario: ValidateSkill composes validators

- **WHEN** `ValidateSkill(skill)` is called
- **THEN** it SHALL run ValidateSkillName, ValidateSkillDescription, ValidateSkillExecution

---

### Requirement: Memory validators

The system SHALL provide standalone validator functions for Memory.

#### Scenario: ValidateMemoryRequiresPathsOrContent

- **WHEN** `ValidateMemory(memory)` is called with memory.Paths == [] and memory.Content == ""
- **THEN** it SHALL return an error Result

---

### Requirement: OpenCode-specific validators in subpackage

The system SHALL provide platform-specific validators in `internal/validation/opencode/validators.go`.

#### Scenario: ValidateAgentMode checks valid mode

- **WHEN** `ValidateAgentMode(agent)` is called with agent.Behavior.Mode == "invalid"
- **THEN** it SHALL return an error Result

#### Scenario: ValidateAgentMode passes valid modes

- **WHEN** `ValidateAgentMode(agent)` is called with mode in ["primary", "subagent", "all", ""]
- **THEN** it SHALL return `NewResult(true)`

#### Scenario: ValidateAgentTemperature checks range

- **WHEN** `ValidateAgentTemperature(agent)` is called with temperature < 0.0 or > 1.0
- **THEN** it SHALL return an error Result

---

### Requirement: Composed OpenCode validators

The system SHALL provide composed validators for OpenCode platform.

#### Scenario: ValidateAgentOpenCode composes OpenCode validators

- **WHEN** `ValidateAgentOpenCode(agent)` is called
- **THEN** it SHALL run ValidateAgentMode, ValidateAgentTemperature

#### Scenario: ValidateCommandOpenCode composes OpenCode validators

- **WHEN** `ValidateCommandOpenCode(command)` is called
- **THEN** it SHALL run OpenCode-specific command validators

#### Scenario: ValidateSkillOpenCode composes OpenCode validators

- **WHEN** `ValidateSkillOpenCode(skill)` is called
- **THEN** it SHALL run OpenCode-specific skill validators
