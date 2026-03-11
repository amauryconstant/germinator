## MODIFIED Requirements

### Requirement: Germinator format serves as canonical source

The system SHALL use a domain-driven canonical YAML format expressing configuration intent, NOT Claude Code format with platform-specific fields.

#### Scenario: Canonical format uses domain-driven fields

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Common fields SHALL be present (Name, Description, Model, Content)
- **AND** PermissionPolicy enum SHALL be present (restrictive, balanced, permissive, analysis, unrestricted)
- **AND** Behavior object SHALL group settings (mode, temperature, maxSteps, prompt, hidden, disabled)
- **AND** Targets section SHALL contain platform-specific overrides
- **AND** Claude Code-specific fields (permissionMode, skills) SHALL NOT be at top level
- **AND** OpenCode-specific fields (mode, temperature, steps, hidden, prompt, disable) SHALL NOT be at top level (except in behavior)

#### Scenario: Tools use split lists with lowercase names

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Tools array SHALL contain lowercase tool names
- **AND** DisallowedTools array SHALL contain lowercase tool names

#### Scenario: Model is simple string with full ID

- **GIVEN** A canonical source YAML file
- **WHEN** The file is parsed
- **THEN** Model SHALL be a string with provider/id format (e.g., "anthropic/claude-sonnet-4-20250514")
- **AND** No alias resolution or normalization SHALL occur
- **AND** Model ID SHALL be passed to output as-is

#### Scenario: Targets section for platform overrides

- **GIVEN** A canonical source YAML with targets section
- **WHEN** The file is parsed
- **THEN** targets.claude-code SHALL contain Claude Code-specific fields (skills array, disableModelInvocation bool)
- **AND** targets.opencode SHALL contain OpenCode-specific overrides (can override behavior object fields)
- **AND** Other platform keys in targets SHALL be supported for future extensibility
- **AND** Empty targets section SHALL NOT cause parsing errors

## REMOVED Requirements

### Requirement: All platform fields MUST be parseable from YAML

**Reason**: Canonical format uses domain-driven field names, not platform-specific enums. Platform adapters handle conversion to/from platform-specific formats. This requirement describes the old approach of having all platform fields in one format.

**Migration**:
- Use canonical models with domain-driven field names
- Create platform adapters in `internal/adapters/` for platform-specific conversions
- Update templates to render from canonical models
- Remove YAML parsing logic that handled platform-specific fields directly

### Requirement: Claude Code-specific fields parseable

**Reason**: Canonical format does NOT include Claude Code-specific fields (permissionMode, skills) at top level. These are now in targets.claude-code section for explicit platform overrides.

**Migration**:
- Parse targets.claude-code section for platform-specific Claude Code fields
- Pass canonical PermissionPolicy enum to adapters for conversion
- Remove direct parsing of permissionMode enum from top-level YAML

### Requirement: OpenCode-specific fields parseable

**Reason**: Canonical format does NOT include OpenCode-specific fields (mode, temperature, maxSteps, hidden, prompt, disable) at top level. These are in behavior object or targets.opencode section.

**Migration**:
- Parse behavior object for OpenCode agent settings
- Flatten behavior object in OpenCode adapter (move fields to top level for output)
- Parse targets.opencode section for platform-specific overrides

### Requirement: Transformation is unidirectional

**Reason**: Canonical format supports bidirectional conversion via adapters (platform→canonical and canonical→platform). This requirement described old unidirectional approach from Germinator format.

**Migration**:
- Implement ToCanonical() method in adapters (parse platform format)
- Implement FromCanonical() method in adapters (render canonical to platform format)
- Support bidirectional transformation for migration and round-trip testing

### Requirement: Templates filter platform-specific fields

**Reason**: Templates now render from canonical models. Platform-specific field filtering is handled by adapters converting canonical format before template rendering.

**Migration**:
- Update templates to use canonical model fields (permissionPolicy, behavior, tools arrays)
- Remove conditional rendering based on platform in templates
- Simplify templates to render canonical fields directly

### Requirement: Source YAML files can contain optional platform fields

**Reason**: Canonical format uses targets section for platform-specific overrides, not inline optional platform fields. This separates concerns clearly.

**Migration**:
- Add targets section schema for platform-specific overrides
- Parse targets.platform keys for platform configuration
- Move inline platform-specific validation to adapter logic

### Requirement: All model fields MUST have YAML and JSON tags

**Reason**: This requirement remains valid but applies to canonical models with new field names. Tags shall use new field names (permissionPolicy, behavior, targets).

**Migration**:
- Update YAML/JSON tags to use canonical field names
- Ensure all new model fields have proper tags
- Remove tags for old platform-specific fields

### Requirement: Temperature nil vs 0.0 distinction

**Reason**: This requirement applies to behavior.temperature pointer field in canonical format. Same logic applies (nil vs 0.0 distinction).

**Migration**:
- Keep behavior.temperature as *float64 pointer
- Template rendering checks nil presence (not zero value)
- Preserve distinction between unset and explicitly set to 0.0
