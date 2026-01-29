## Why

Germinator currently uses Claude Code's document format as the source standard, which tightly couples the implementation to one platform's conventions. This limits extensibility and makes it difficult to support new platforms like OpenCode without significant refactoring.

To add OpenCode as a target platform, we need to:
1. Refactor models to be platform-agnostic (separate common from platform-specific fields)
2. Establish template-based serialization infrastructure
3. Create OpenCode templates and validation
4. Add comprehensive tests and documentation

This change provides a complete solution: a platform-agnostic foundation that serves as a single source of truth, with concrete OpenCode transformation logic implemented through Go templates.

## What Changes

### Model Refactoring
- Refactor `internal/models/` to support multiple platforms with shared common fields and platform-specific extensions for Agent, Command, Skill, and Memory models
- Add OpenCode-specific fields to Agent model: Mode (primary/subagent/all, default: all), Temperature (0.0-1.0), MaxSteps (> 0), Hidden (bool), Prompt, Disable
- Add OpenCode-specific field to Command model: Subtask
- Add OpenCode-specific fields to Skill model: License, Compatibility (list), Metadata (map), Hooks
- Support both file-based and narrative memory types (paths and content fields)
- Support full model IDs (e.g., `anthropic/claude-sonnet-4-20250514`) across platforms
- OpenCode-specific fields (Mode, Temperature, MaxSteps, Hidden, Prompt, Disable, Subtask, License, Compatibility, Metadata, Hooks) can ONLY be provided via JSON files or CLI flags when targeting OpenCode platform
- These fields have `yaml:"-"` tags to prevent YAML serialization to non-target platforms

### Template Infrastructure
- Create platform-specific Go templates in `config/templates/<platform>/` for serialization (Claude Code and OpenCode)
- Implement template functions: permission transformation (basic approximation)
- Establish template function registration in serializer

### OpenCode Templates
- Create 4 Go template files in `config/templates/opencode/`:
  - `agent.tmpl` - Transform Agent model to OpenCode agent format with mode, tools map, permissions map
  - `command.tmpl` - Transform Command model with template field and $ARGUMENTS placeholder
  - `skill.tmpl` - Transform Skill model to `.opencode/skills/<name>/SKILL.md` structure with frontmatter
  - `memory.tmpl` - Transform Memory model to AGENTS.md format with @ file references
- Implement transformation logic within templates:
  - Conditional rendering for optional fields
  - Convert tools allowed/disallowed lists to `{tool: true|false}` map
  - Permission mapping using transformPermissionModeToOpenCode() function

### Validation
- Establish platform-agnostic validation rules with single `Validate(platform string)` function (platform always required)
- Define tool configuration schema using flat arrays as single internal representation (no ToolConfig field)
- Fix skill name regex to match OpenCode specs: `^[a-z0-9]+(-[a-z0-9]+)*$`
- Add platform-specific validation functions in `internal/services/transformer.go` for OpenCode constraints

### Testing
- Create comprehensive test fixtures and golden files for all 4 document types in `test/fixtures/` and `test/golden/`
- Add table-driven tests for template transformations, validation, and edge cases

### Documentation
- Update `README.md` and `AGENTS.md` with OpenCode usage examples and field mapping notes
- Document CLI changes (--platform flag requirement)
- Document known limitations (permission mode approximation, skipped Claude Code-specific fields)

## Capabilities

### New Capabilities

**Models:**
- `platform-agnostic-models`: Core domain models (Agent, Command, Skill, Memory) with shared common fields and platform-specific extensions. OpenCode-specific fields: Mode, Temperature, MaxSteps, Hidden, Prompt, Disable for Agent; Subtask for Command; License, Compatibility, Metadata, Hooks for Skill. Platform-specific fields have yaml:"-" tags to prevent cross-platform contamination. OpenCode-specific fields can ONLY be provided via JSON files or CLI flags, not YAML.

**Transformation:**
- `permission-transformation`: Custom template function to transform Claude Code's `permissionMode` enum to OpenCode's permission object format. Preserves semantic intent: dontAsk→allow/allow, bypassPermissions→allow/allow, plan→deny/deny, default→ask/ask, acceptEdits→allow/ask.
- `opencode-agent-transformation`: Template-based transformation from platform-agnostic Agent model to OpenCode agent YAML format with mode field (default "all"), tools map conversion (arrays to `{tool: true|false}`), and permissions map using transformPermissionModeToOpenCode().
- `opencode-command-transformation`: Template-based transformation from platform-agnostic Command model to OpenCode command format with template field and $ARGUMENTS placeholder for argument substitution.
- `opencode-skill-transformation`: Template-based transformation from platform-agnostic Skill model to `.opencode/skills/<name>/SKILL.md` directory structure with YAML frontmatter, handling OpenCode-specific fields: License, Compatibility, Metadata, Hooks.
- `opencode-memory-transformation`: Template-based transformation from platform-agnostic Memory model to AGENTS.md format with @ file references (e.g., @README.md) and project context narrative.

**Validation:**
- `platform-agnostic-validation`: All models implement `Validate(platform string)` method that applies both common validation (required fields, data types, format constraints) and platform-specific validation rules. Platform parameter is always required.
- `opencode-platform-validation`: Platform-specific validation functions to enforce OpenCode constraints: Agent mode values (primary/subagent/all), temperature range (0.0-1.0), MaxSteps constraint (> 0), skill name regex (`^[a-z0-9]+(-[a-z0-9]+)*$`).

**Infrastructure:**
- `platform-field-mappings`: Documentation of all field mappings between Germinator models, Claude Code format, and OpenCode format. Indicates field type (common, platform-specific), transformation logic, and skipped fields.

### Modified Capabilities

**Documents:**
- `template-based-serialization`: MODIFIED - Go templates in `config/templates/<platform>/` that transform platform-agnostic models to platform-specific outputs. Now includes custom template function registration for permission transformations.

## Impact

- **New files**:
  - `internal/models/models.go` (refactored with platform-agnostic structure)
  - `internal/core/template_funcs.go` (custom template functions)
  - `config/templates/opencode/agent.tmpl`
  - `config/templates/opencode/command.tmpl`
  - `config/templates/opencode/skill.tmpl`
  - `config/templates/opencode/memory.tmpl`
  - `test/fixtures/opencode/` (fixture files for all 4 document types)
  - `test/golden/opencode/` (golden files for all 4 document types)

- **Modified files**:
  - `internal/core/serializer.go` (template function registration, platform-aware template loading)
  - `internal/services/transformer.go` (add OpenCode validation functions)
  - `internal/models/models.go` (restructured domain models)
  - `README.md` (add OpenCode usage examples)
  - `AGENTS.md` (add OpenCode field mapping notes, CLI changes)

- **Dependencies**: No new dependencies required

- **Breaking changes**:
  - Existing JSON schemas and model definitions will be replaced
  - `Validate()` method signature changes from `Validate()` to `Validate(platform string)` (platform always required)
  - `--platform` flag becomes required for all CLI operations (no default to Claude Code)
  - Users with custom schemas or parsers will need to migrate
  - Users upgrading from v0.x will need to explicitly specify --platform flag for all operations

- **Claude Code-specific fields not supported in OpenCode**:
   - Agent.skills list (skipped without warnings)
   - Skill.userInvocable (skipped without warnings)
   - Command.argumentHint (skipped without warnings) **NOTE: Preserved in implementation for backward compatibility with existing tests and codebase**
   - Command.disableModelInvocation (skipped without warnings)
   - Agent.permissionMode (transformed to OpenCode permission object via transformPermissionModeToOpenCode())
- **Permission mode mapping preserves distinction**: dontAsk → allow both, bypassPermissions → allow both, plan → deny both
