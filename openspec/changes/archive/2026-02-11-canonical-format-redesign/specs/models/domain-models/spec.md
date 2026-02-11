## MODIFIED Requirements

### Requirement: Document Models

The project SHALL define canonical document models (Agent, Command, Memory, Skill) with domain-driven field names independent of platform specifics.

#### Scenario: Canonical Agent struct

- **GIVEN** Canonical Agent struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have DisallowedTools []string field with yaml tag "disallowedTools" (lowercase names)
- **AND** it SHALL have PermissionPolicy PermissionPolicy enum field with yaml tag "permissionPolicy" (not string)
- **AND** it SHALL have Behavior AgentBehavior struct field with yaml tag "behavior"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical AgentBehavior struct

- **GIVEN** Canonical AgentBehavior struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Mode string field with yaml tag "mode" (values: primary, subagent, all)
- **AND** it SHALL have Temperature *float64 field with yaml tag "temperature"
- **AND** it SHALL have MaxSteps int field with yaml tag "maxSteps"
- **AND** it SHALL have Prompt string field with yaml tag "prompt"
- **AND** it SHALL have Hidden bool field with yaml tag "hidden"
- **AND** it SHALL have Disabled bool field with yaml tag "disabled"

#### Scenario: Canonical Command struct

- **GIVEN** Canonical Command struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have Execution CommandExecution struct field with yaml tag "execution"
- **AND** it SHALL have Arguments CommandArguments struct field with yaml tag "arguments"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical CommandExecution struct

- **GIVEN** Canonical CommandExecution struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Context string field with yaml tag "context"
- **AND** it SHALL have Subtask bool field with yaml tag "subtask"
- **AND** it SHALL have Agent string field with yaml tag "agent"

#### Scenario: Canonical CommandArguments struct

- **GIVEN** Canonical CommandArguments struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Hint string field with yaml tag "hint"

#### Scenario: Canonical Memory struct

- **GIVEN** Canonical Memory struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Paths []string field with yaml tag "paths"
- **AND** it SHALL have Content string field with yaml tag "content"
- **AND** it SHALL have FilePath string field (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical Skill struct

- **GIVEN** Canonical Skill struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Name string field with yaml tag "name"
- **AND** it SHALL have Description string field with yaml tag "description"
- **AND** it SHALL have Tools []string field with yaml tag "tools" (lowercase names)
- **AND** it SHALL have Extensions SkillExtensions struct field with yaml tag "extensions"
- **AND** it SHALL have Execution SkillExecution struct field with yaml tag "execution"
- **AND** it SHALL have Model string field with yaml tag "model"
- **AND** it SHALL have Targets map[string]PlatformConfig field with yaml tag "targets"
- **AND** it SHALL have FilePath and Content string fields (ignored from YAML with `yaml:"-"`)

#### Scenario: Canonical SkillExtensions struct

- **GIVEN** Canonical SkillExtensions struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have License string field with yaml tag "license"
- **AND** it SHALL have Compatibility []string field with yaml tag "compatibility"
- **AND** it SHALL have Metadata map[string]string field with yaml tag "metadata"
- **AND** it SHALL have Hooks map[string]string field with yaml tag "hooks"

#### Scenario: Canonical SkillExecution struct

- **GIVEN** Canonical SkillExecution struct is defined
- **WHEN** A developer inspects the struct
- **THEN** it SHALL have Context string field with yaml tag "context"
- **AND** it SHALL have Agent string field with yaml tag "agent"
- **AND** it SHALL have UserInvocable bool field with yaml tag "userInvocable"

## REMOVED Requirements

### Requirement: Agent struct has Claude Code PermissionMode field

**Reason**: Canonical format uses PermissionPolicy enum (domain-driven), not Claude Code's permissionMode string enum.

**Migration**:
- Remove PermissionMode string field from Agent struct
- Add PermissionPolicy enum field with values: restrictive, balanced, permissive, analysis, unrestricted
- Create PermissionPolicy type with String() method for YAML serialization
- Update all code referencing PermissionMode to use PermissionPolicy

### Requirement: Agent struct has Claude Code Skills field

**Reason**: Canonical format uses targets.claude-code.skills array for platform-specific configuration. Skills are Claude Code-only concept.

**Migration**:
- Remove Skills []string field from Agent struct (top level)
- Add Targets map[string]PlatformConfig field
- PlatformConfig struct contains fields for each platform (e.g., ClaudeCodeConfig with Skills []string)
- Move skills validation to adapter logic

### Requirement: Command struct has Claude Code AllowedTools field

**Reason**: Canonical format uses tools array (unified for both platforms) instead of Claude Code-specific allowedTools.

**Migration**:
- Remove AllowedTools []string field from Command struct
- Use Tools []string field for all tool access control (both platforms)
- Adapter handles conversion to Claude Code allowedTools format if needed
- Update all code referencing AllowedTools to use Tools

### Requirement: Command struct has Claude Code ArgumentHint field

**Reason**: Canonical format uses arguments.hint field in arguments object for clearer structure.

**Migration**:
- Remove ArgumentHint string field from Command struct
- Add Arguments CommandArguments struct with Hint string field
- Update templates to use arguments.hint path
- Update all code referencing ArgumentHint to use arguments.hint

### Requirement: Command struct has Claude Code DisableModelInvocation field

**Reason**: Canonical format uses targets.claude-code.disableModelInvocation for platform-specific configuration. DisableModelInvocation is Claude Code-only concept.

**Migration**:
- Remove DisableModelInvocation bool field from Command struct
- Add Targets map[string]PlatformConfig field
- ClaudeCodeConfig struct contains DisableModelInvocation bool field
- Move disableModelInvocation validation to adapter logic

### Requirement: Command struct has Claude Code Context field

**Reason**: Canonical format uses execution.context field for execution settings (applies to both platforms).

**Migration**:
- Remove Context string field from Command struct
- Add Execution CommandExecution struct with Context string field
- Update templates to use execution.context path
- Update all code referencing Context to use execution.context

### Requirement: Skill struct has Claude Code AllowedTools field

**Reason**: Canonical format uses tools array (unified for both platforms) instead of Claude Code-specific allowedTools.

**Migration**:
- Remove AllowedTools []string field from Skill struct
- Use Tools []string field for all tool access control
- Adapter handles conversion to Claude Code allowedTools format if needed
- Update all code referencing AllowedTools to use Tools

### Requirement: Skill struct has Claude Code UserInvocable field

**Reason**: Canonical format uses execution.userInvocable field in execution object for execution settings.

**Migration**:
- Remove UserInvocable bool field from Skill struct
- Add Execution SkillExecution struct with UserInvocable bool field
- Update templates to use execution.userInvocable path
- Update all code referencing UserInvocable to use execution.userInvocable
