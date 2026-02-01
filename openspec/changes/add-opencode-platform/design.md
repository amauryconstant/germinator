## Context

Germinator currently uses Claude Code's document format as the source standard in `internal/models/`. The models (Agent, Command, Skill, Memory) contain platform-specific fields like `permissionMode` (Claude Code) and `context` (Claude Code fork mode). This tight coupling makes it difficult to support new platforms like OpenCode without significant refactoring.

The current serialization uses Go templates in `config/templates/claude-code/` to render output. However, there are no templates for OpenCode or other platforms. The models are tightly coupled to Claude Code format with hardcoded validation rules.

**Platform differences discovered**:
- **Claude Code**: `permissionMode` enum (default, acceptEdits, dontAsk, bypassPermissions, plan), flat tool arrays (tools, disallowedTools), short model names (sonnet, opus, haiku)
- **OpenCode**: `permission` object with allow/ask/deny values, structured tool config ({"bash": true}), agent mode field (primary/subagent), full model IDs (anthropic/claude-sonnet-4-20250514)
- **Permission systems are fundamentally different** - require complex transformation logic
- **Tool configuration formats don't map 1:1** - need hybrid approach

**Current state**:
- Single set of models tightly coupled to Claude Code format
- Platform-specific validation embedded in model `Validate()` methods
- Template-based serialization exists only for Claude Code
- Memory only supports file paths, not narrative content
- No permission mode transformation between platforms
- No template functions for transformations

**Constraints**:
- No new dependencies required
- Must preserve existing test fixtures and golden files
- CLI must work with any input/output paths
- Platform differences handled via Go templates, not YAML adapters
- Use kebab-case for YAML (matches fixtures), PascalCase for Go structs
- --platform flag always required (no default behavior)

## Goals / Non-Goals

**Goals:**
- Refactor models to be platform-agnostic while preserving all information from source formats
- Create platform-specific Go templates for serialization (extending existing Claude Code templates)
- Implement permission mode transformation logic from Claude Code to OpenCode
- Establish validation rules that are platform-agnostic (with platform-specific extensions)
- Support tool configurations: flat arrays (Claude Code) and structured objects (OpenCode)
- Add Agent Mode field to support OpenCode's primary/subagent distinction
- Add OpenCode-specific configuration fields (Temperature, MaxSteps, Hidden, Prompt, Disable, Subtask, License, Compatibility, Metadata, Hooks)
- Support full model IDs (e.g., `anthropic/claude-sonnet-4-20250514`) across platforms
- Support both file-based memory (paths) and narrative memory (content)
- Create OpenCode templates for all 4 document types
- Add platform-specific validation for OpenCode constraints
- Create comprehensive test fixtures and golden files
- Update documentation with OpenCode usage examples
- Make --platform flag always required (breaking change)

**Non-Goals:**
- Complete abstraction of all possible AI assistant platforms (focus on Claude Code and OpenCode)
- Runtime platform detection (user explicitly specifies target platform)
- Full validation of platform-specific features at model level (defer to templates and platform-specific validation functions)
- Automatic migration of user's custom schemas or parsers (document breaking changes clearly)
- YAML-based adapter configuration (use Go templates instead)
- Bidirectional transformation (OpenCode to platform-agnostic model is out of scope)
- Complete feature parity between Claude Code and OpenCode (skip unsupported features silently)
- Live validation or runtime adaptation beyond template rendering and platform-specific validation
- Automated migration scripts for existing Claude Code users
- Warnings for skipped Claude Code-specific fields (skip silently)

## Decisions

### 1. Model Structure: Germinator Format as Canonical Source

**Decision**: Germinator YAML format is the canonical source containing ALL platform fields (Claude Code + OpenCode) in single structs. All fields have YAML and JSON tags for full parseability.

**Rationale**:
- Go doesn't support inheritance, single structs with all fields is idiomatic
- Germinator format serves as single source of truth with complete field set
- All fields parseable from YAML eliminates need for JSON/CLI overrides
- Clear separation of concerns: source (Germinator YAML) → transformation (templates) → output (target platform)
- Templates handle field filtering based on platform, not struct tags

**Alternatives considered**:
- Separate structs per platform (rejected: would duplicate code, lose single-source-of-truth)
- Claude Code YAML + JSON overrides (rejected: complex workflow, multiple files)
- Tag-based conditional fields (rejected: complex to maintain, poor compile-time safety)

### 2. Field Mapping: Go Template-Based Serialization

**Decision**: Use Go templates in `config/templates/<platform>/<docType>.tmpl` for serialization. Each platform has its own set of templates that transform platform-agnostic models to platform-specific output format.

**Rationale**:
- Existing codebase already uses Go templates (Claude Code)
- Go templates provide full programming power (conditionals, loops, custom functions)
- No external dependency on YAML configuration loading
- Type-safe at compile time
- Easy to extend for new platforms by adding new template directories

**Structure**:
```
config/templates/
 ├── claude-code/
 │   ├── agent.tmpl
 │   ├── command.tmpl
 │   ├── skill.tmpl
 │   └── memory.tmpl
 └── opencode/
     ├── agent.tmpl
     ├── command.tmpl
     ├── skill.tmpl
     └── memory.tmpl
```

**Template example** (opencode/agent.tmpl):
```go
---
name: {{.Name}}
description: {{.Description}}
{{if .Tools}}
tools:
{{range .Tools}}
  {{.}}: true
{{end}}
{{end}}
{{if .DisallowedTools}}
disallowedTools:
{{range .DisallowedTools}}
  {{.}}: false
{{end}}
{{end}}
{{if .PermissionMode}}
permission:
{{transformPermissionMode .PermissionMode}}
{{end}}
{{if .Mode}}
mode: {{.Mode}}
{{end}}
{{if not .Mode}}
mode: all
{{end}}
---
{{.Content}}
```

### 3. Permission Mode Transformation: Claude Code → OpenCode

**Decision**: Implement custom Go template functions to transform Claude Code's `permissionMode` enum to OpenCode's permission objects. This is a complex mapping because the two systems are fundamentally different. We implement basic mapping only (edit and bash tools) and document limitations.

**Important limitation**: OpenCode's permission system requires nested objects with command keys (e.g., `{"bash": {"*": "ask", "git *": "allow"}}`). Claude Code's enum cannot represent this granularity. Our transformation provides a basic approximation that maps to top-level edit and bash permissions only.

**Mapping logic** (Claude Code → OpenCode - Preserves semantic intent):
```go
default → {"edit": {"*": "ask"}, "bash": {"*": "ask"}}
acceptEdits → {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
dontAsk → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
bypassPermissions → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
plan → {"edit": {"*": "deny"}, "bash": {"*": "deny"}}
```

**Rationale**:
- `dontAsk` means "don't prompt user for approval" → allow both edit and bash
- `bypassPermissions` means "override permission restrictions" → allow both edit and bash
- `plan` is a restricted mode for analysis only → deny both edit and bash
- This preserves the semantic distinction between Claude Code's permission modes

**Rationale**:
- Permission systems are fundamentally incompatible (enum vs. nested object with command keys)
- OpenCode's permission system supports command-level granularity (e.g., `"git push": "deny"`) that Claude Code's enum cannot represent
- Global wildcards like `"*": "allow"` are not supported by OpenCode - must use tool-specific wildcards
- Basic mapping provides useful approximation for common cases
- Complex permission configurations require manual configuration
- Custom template functions provide clean separation of concerns

**Implementation**:
```go
func transformPermissionMode(mode string) map[string]interface{} {
    switch mode {
    case "default":
        return map[string]interface{}{
            "edit": map[string]string{"*": "ask"},
            "bash": map[string]string{"*": "ask"},
        }
    case "acceptEdits":
        return map[string]interface{}{
            "edit": map[string]string{"*": "allow"},
            "bash": map[string]string{"*": "ask"},
        }
    case "dontAsk":
        return map[string]interface{}{
            "edit": map[string]string{"*": "allow"},
            "bash": map[string]string{"*": "allow"},
        }
    case "bypassPermissions":
        return map[string]interface{}{
            "edit": map[string]string{"*": "allow"},
            "bash": map[string]string{"*": "allow"},
        }
    case "plan":
        return map[string]interface{}{
            "edit": map[string]string{"*": "deny"},
            "bash": map[string]string{"*": "deny"},
        }
    default:
        return nil // Unknown mode, handle gracefully
    }
}
```

**Delta Spec**: See `specs/transformation/permission-transformation/` for detailed requirements including scenarios for all permission modes, limitations, and testing requirements.

### 4. Tool Configuration: Flat Array Representation Only

**Decision**: Use Claude Code's flat array format as the single internal representation. Templates handle transformation to OpenCode's structured object format.

**Platform-agnostic model structure**:
```go
type Agent struct {
    Tools           []string `yaml:"tools,omitempty"`
    DisallowedTools []string `yaml:"disallowedTools,omitempty"`
    // ... other fields
}

type Command struct {
    AllowedTools    []string `yaml:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty"`
    // ... other fields
}

type Skill struct {
    AllowedTools    []string `yaml:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty"`
    // ... other fields
}
```

**Transformation logic**:
- Claude Code templates use flat arrays as-is: `tools: [bash, edit]`, `disallowedTools: [write]`
- OpenCode templates transform arrays to objects: `tools: {bash: true, edit: true}`
- DisallowedTools preserved in Agent and Skill models for forward compatibility (not used by OpenCode currently)
- Single source of truth in the model
- No ambiguity or synchronization issues
- No ToolConfig field (removed from design)

**Rationale**:
- Claude Code format (flat arrays) is simpler and works as base representation
- Templates handle all transformations - cleaner model structure
- Single source of truth eliminates conflicts
- Simplifies validation logic
- Transformation in templates is flexible and platform-specific
- ToolConfig field is unnecessary and would add confusion

### 5. Validation: Single Function with Platform-Aware Logic

**Decision**: Use a single `Validate(platform string)` function that applies both platform-agnostic and platform-specific validation rules. Platform parameter is always required.

**Implementation**:
```go
func (a *Agent) Validate(platform string) []error {
    var errs []error

    // Platform-agnostic validation (all platforms)
    if a.Name == "" {
        errs = append(errs, errors.New("name is required"))
    }
    if a.Description == "" {
        errs = append(errs, errors.New("description is required"))
    }

    // Platform-specific validation
    switch platform {
    case "claude-code":
        // Claude Code-specific rules (e.g., validModelNames)
    case "opencode":
        // OpenCode-specific rules (e.g., mode values, temperature range)
    default:
        errs = append(errs, fmt.Errorf("unknown platform: %s (available: claude-code, opencode)", platform))
    }

    return errs
}
```

**Rationale**:
- Single function is simpler and easier to understand
- No duplication between base and platform validation
- Platform is always required - no ambiguity
- Clear control flow with switch statement
- Easier to add new platforms in future
- Breaking change but ensures explicit platform selection

**Alternative Considered**: Separate validation methods per platform

```go
func (a *Agent) ValidateOpenCode() []error
func (a *Agent) ValidateClaudeCode() []error
```

**Rationale for chosen approach**:
- Single entry point simplifies caller code (CLI, tests, services)
- Platform parameter is explicit, preventing accidental validation with wrong platform
- Trade-off: Larger methods with switch statements, but manageable for 2 platforms
- Adding platforms requires updating switch statements but doesn't change caller code

**Note**: If platforms grow beyond 3, reconsider this decision for platform-specific validation methods to avoid unwieldy switch statements.

### 6. Agent Mode: Primary vs Subagent Support

**Decision**: Add `Mode` field to Agent model to support OpenCode's primary/subagent distinction. This field is platform-specific but useful for generically distinguishing agent types.

**Platform differences**:
- **Claude Code**: No explicit mode field (inferred by invocation)
- **OpenCode**: `mode: "primary" | "subagent" | "all"`

**Platform-agnostic model structure (Germinator format - all fields parseable)**:
```go
type Agent struct {
    Mode    string   `yaml:"mode,omitempty" json:"mode,omitempty"`
    // ... other fields
}
```

**Template handling**:
- Claude Code templates: omit `Mode` field (not used by Claude Code)
- OpenCode templates: include `mode: {{.Mode}}` if specified
- Default to "all" if not specified when serializing to OpenCode

**Rationale**:
- Provides clear distinction between agent types across platforms
- OpenCode uses this for UI organization and permissions
- Doesn't impact Claude Code functionality (field simply omitted)
- Extensible for future platforms with similar concepts

### Decision 6.1: Agent Name Validation

**Decision**: Agent names follow the same format as Skill names: `^[a-z0-9]+(-[a-z0-9]+)*$`

**Rationale**: Consistency across document types prevents confusion and ensures compatibility with OpenCode's strict naming requirements.

**Platform differences**:
- **Claude Code**: No strict validation, but conventionally kebab-case
- **OpenCode**: Enforces kebab-case with alphanumeric characters and hyphens (no consecutive/leading/trailing hyphens)

### 7. OpenCode-Specific Agent Configuration

**Decision**: Add `Mode`, `Temperature`, `MaxSteps`, `Hidden`, `Prompt`, and `Disable` fields to Agent model to support OpenCode's agent configuration options. These fields are OpenCode-specific and should be omitted when serializing to other platforms.

**Platform differences**:
- **Claude Code**: No equivalent fields (skipped)
- **OpenCode**:
  - `mode` (string): Agent type - primary, subagent, or all
  - `temperature` (float): Controls randomness of LLM responses, range 0.0-1.0
  - `maxSteps` (int): Maximum number of agentic iterations before forced response, must be > 0
  - `hidden` (bool): Hide subagent from @ autocomplete menu
  - `prompt` (string): Custom system prompt override
  - `disable` (bool): Disable agent from @ autocomplete

**Platform-agnostic model structure**:
```go
type Agent struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION (single representation - flat arrays)
    Tools           []string `yaml:"tools,omitempty" json:"tools,omitempty"`
    DisallowedTools []string `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`

     // OPENCODE-SPECIFIC (parseable from Germinator YAML)
    Mode        string  `yaml:"mode,omitempty" json:"mode,omitempty"`
    Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    MaxSteps    int     `yaml:"maxSteps,omitempty" json:"maxSteps,omitempty"`
    Hidden      bool    `yaml:"hidden,omitempty" json:"hidden,omitempty"`
    Prompt      string  `yaml:"prompt,omitempty" json:"prompt,omitempty"`
    Disable     bool    `yaml:"disable,omitempty" json:"disable,omitempty"`

    // CLAUDE CODE-SPECIFIC
    PermissionMode string   `yaml:"permissionMode,omitempty" json:"permissionMode,omitempty"`
    Skills        []string `yaml:"skills,omitempty" json:"skills,omitempty"`

    // COMMON (but different per platform)
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
}
```

**Template handling**:
- Claude Code templates: omit OpenCode-specific fields (Mode, Temperature, MaxSteps, Hidden, Prompt, Disable)
- OpenCode templates: include all OpenCode fields when specified
- Validation: Temperature must be 0.0-1.0, MaxSteps must be >= 1, Mode must be primary/subagent/all

**Rationale**:
- OpenCode uses these fields for fine-grained control over agent behavior
- Preserving these fields in the model allows accurate transformation
- Omitting from Claude Code templates maintains compatibility
- Extensible for future platforms with similar configuration options

### Decision 7.1: Germinator Format as Canonical Source

**Decision**: Germinator YAML format is the canonical source format containing ALL platform fields. Source files include both Claude Code and OpenCode fields, with all fields parseable from YAML. Unidirectional transformation: Germinator format → Claude Code OR OpenCode.

**Rationale**:
- Single canonical source eliminates ambiguity about what format to author
- All fields available in one file simplifies configuration
- Templates filter fields based on target platform, not struct tags
- Unidirectional flow matches current use case (convert Germinator source to platform-specific output)
- No need for CLI flags or JSON overrides - all fields in source YAML

**Implementation**:
- Model structs have all fields with proper YAML tags: `yaml:"field,omitempty" json:"field,omitempty"`
- Source files written in Germinator format with complete field set
- Templates use conditionals to render platform-specific output
- `RenderDocument(doc, platform)` filters based on platform parameter

### 8. OpenCode-Specific Command Fields

**Decision**: Add `Subtask` field to Command model to support OpenCode's subtask configuration.

**Platform differences**:
- **Claude Code**: No subtask field (skipped)
- **OpenCode**: `subtask` (bool): Whether this command can be invoked as a subtask by agents

**Platform-agnostic model structure**:
```go
type Command struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION
    AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

     // OPENCODE-SPECIFIC (parseable from Germinator YAML)
     Subtask bool `yaml:"subtask,omitempty" json:"subtask,omitempty"`

    // CLAUDE CODE-SPECIFIC
    Context                string `yaml:"context,omitempty" json:"context,omitempty"`
    Agent                  string `yaml:"agent,omitempty" json:"agent,omitempty"`
    DisableModelInvocation bool   `yaml:"disable-model-invocation,omitempty" json:"disable-model-invocation,omitempty"`

    // COMMON (but different per platform)
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
}
```

**Rationale**:
- OpenCode uses subtask flag to control command invocation
- Preserving this field enables accurate transformation
- Omitted in Claude Code templates (not rendered)
- DisallowedTools is preserved in model for forward compatibility even if OpenCode doesn't currently support it

### 9. OpenCode-Specific Skill Fields

**Decision**: Add `License`, `Compatibility`, `Metadata`, and `Hooks` fields to Skill model to support OpenCode's skill configuration.

**Platform differences**:
- **Claude Code**: No equivalent fields
- **OpenCode**:
  - `license` (string): License identifier (e.g., MIT, Apache-2.0)
  - `compatibility` (list): List of compatible AI platforms
  - `metadata` (map): Key-value metadata
  - `hooks` (map): Pre/post execution hooks

**Platform-agnostic model structure**:
```go
type Skill struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION
    AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

     // OPENCODE-SPECIFIC (parseable from Germinator YAML)
    License       string            `yaml:"license,omitempty" json:"license,omitempty"`
    Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
    Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
    Hooks        map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`

    // CLAUDE CODE-SPECIFIC
    Model          string `yaml:"model,omitempty" json:"model,omitempty"`
    Context        string `yaml:"context,omitempty" json:"context,omitempty"`
    Agent          string `yaml:"agent,omitempty" json:"agent,omitempty"`
    UserInvocable bool   `yaml:"user-invocable,omitempty" json:"user-invocable,omitempty"`
}
```

**Rationale**:
- OpenCode uses these fields for skill metadata and configuration
- Preserving these fields enables accurate transformation
- Skipped in Claude Code templates (fields omitted)

### 10. Memory: Dual Support for Files and Narrative

**Decision**: Memory model supports both `paths` (file paths to load into context) and `content` (narrative text directly included). All fields parseable from YAML.

**Rationale**:
- Some platforms use file-based memory (Claude Code)
- Some platforms support narrative memory (OpenCode skills)
- Adapters normalize both forms to platform-specific output
- Germinator format allows specifying both or either

**Model structure**:
```go
type Memory struct {
    Paths   []string `yaml:"paths,omitempty" json:"paths,omitempty"`    // file paths
    Content string   `yaml:"content,omitempty" json:"content,omitempty"` // narrative content
    FilePath string  `yaml:"-" json:"-"`
}
```

**Template handling**:
- Claude Code templates: render paths as list
- OpenCode templates: transform paths to @ file references, render content as project context, and include explicit teaching instructions for Read tool usage at the top of AGENTS.md

### 11. Model IDs: Full Platform-Specific Format

**Decision**: Store full provider-prefixed model IDs as user-provided (e.g., `anthropic/claude-sonnet-4-20250514`, `openai/gpt-4-1`). Do not normalize or transform between short names and full IDs - user provides exact format needed for their target platform.

**Platform differences**:
- **Claude Code**: Short names (sonnet, opus, haiku) but full IDs also supported
- **OpenCode**: Full provider-prefixed with version (anthropic/claude-sonnet-4-20250514)

**Platform-agnostic model structure**:
```go
type Agent struct {
    Model    string `yaml:"model"` // User-provided, platform-specific ID
    // ... other fields
}
```

**Template handling**:
- Templates use `.Model` directly without transformation
- No normalization or mapping between formats
- User responsible for providing correct model ID for target platform

**Rationale**:
- Model IDs are platform-specific and can change frequently
- Provider prefixes and version numbers differ across platforms
- Users know their target platform's model IDs
- Avoids maintaining a mapping table that quickly becomes outdated
- Templates preserve user-provided value exactly

### 12. Platform Requirement: Always Required

**Decision**: --platform flag is always required for all CLI operations. No default to Claude Code.

**Rationale**:
- Forces explicit platform selection, avoiding confusion
- Prevents accidental transformations to wrong platform
- Breaking change but ensures correctness
- Clear error message when platform not provided

**Implementation**:
```go
// In CLI command
if platform == "" {
    return errors.New("--platform flag is required (available: claude-code, opencode)")
}
```

## Risks / Trade-offs

### Risk: Over-Abstraction Loss of Clarity

**Risk**: Complex platform-agnostic models may become difficult to understand and maintain.

**Mitigation**: Keep models simple, use clear field names, document platform-specific fields with comments. Avoid over-engineering - only abstract what's truly different between platforms.

### Risk: Information Loss During Transformation

**Risk**: Platform-specific features in source format may not map cleanly to target platform, resulting in dropped fields.

**Mitigation**: Document all field mappings clearly in spec files. Skip unsupported features silently (no warnings). Users can manually add missing features after transformation.

### Risk: Permission Mode Mapping Complexity

**Risk**: Transforming between Claude Code's `permissionMode` enum and OpenCode's permission objects is complex and may have edge cases.

**Mitigation**: Document all mapping rules explicitly. Write comprehensive tests for all permission mode transformations. Provide clear examples in documentation.

### Trade-off: Template Complexity vs. Flexibility

**Trade-off**: Go templates provide full programming power but are harder to debug than declarative YAML.

**Decision**: Use Go templates for flexibility (conditionals, loops, custom functions). Write comprehensive unit tests for template rendering.

### Risk: Breaking Changes for Existing Users

**Risk**: Users with custom parsers or schemas will need to update their code.

**Mitigation**: Document breaking changes clearly in migration guide. Provide example migration path. Version JSON schemas clearly.

### Trade-off: Validation Complexity

**Trade-off**: Platform-agnostic validation vs. platform-specific validation.

**Decision**: Layered approach (base + platform-specific) balances consistency and flexibility. Base validation catches obvious errors; platform validation catches platform-specific issues.

## Testing Strategy

### Current Testing Issues Identified

Based on comprehensive testing system review, the following issues require resolution:

**1. Platform Coverage Gap**
- Integration tests hardcode "claude-code" platform, never testing OpenCode
- `internal/core/integration_test.go:73` uses fixed platform string
- Risk: OpenCode-specific code paths untested in integration scenarios

**2. Golden Files Not Tested**
- Golden files exist in `test/golden/opencode/` but are never referenced in test code
- No automated verification of golden file correctness
- Risk: Golden files become stale, manual verification only

**3. Custom Utility Functions**
- `cmd/cmd_test.go` implements custom `contains()` and `containsMiddle()` functions
- Standard library `strings.Contains()` provides equivalent functionality
- Risk: Unnecessary complexity, potential bugs

**4. Inconsistent Test Data Setup**
- `transformer_test.go`: Uses `t.TempDir()` (dynamic test files)
- `integration_test.go`: Uses static fixtures from `test/fixtures/`
- Mixed patterns create navigation and understanding issues

**5. Coverage Gaps**
- `cmd` package: 20.6% coverage (minimal tests)
- `version` package: 0% coverage (no tests)
- Core functionality tested, but CLI and version components neglected

**6. Fragile Path Resolution**
- `integration_test.go:17` uses relative path navigation
- Assumes specific working directory structure
- Risk: Tests fail from different working directories

**7. Missing Loader Unit Tests**
- No dedicated tests for `loader.go` functions
- Only indirect testing through integration tests
- Edge cases in `DetectType()` and `LoadDocument()` untested

**8. Test Documentation Gaps**
- `test/README.md` is minimal (17 lines)
- No guidance on test patterns or conventions
- Unclear when to use fixtures vs golden files

### Testing Patterns and Conventions

**Table-Driven Tests**
- Primary pattern for test organization
- Structure: `tests := []struct { name, input, expect }`
- Run with `for _, tt := range tests { t.Run(tt.name, ...) }`
- Used in: `models_test.go`, `serializer_test.go`, `transformer_test.go`

**Test Data Setup**
- **Preferred**: `t.TempDir()` for dynamic test file creation
- **Alternative**: Static fixtures in `test/fixtures/` for shared test data
- **Decision**: Use dynamic for test-specific data, static for reusable fixtures

**Error Assertions**
- **Preferred**: Explicit error count with `errorCount` field
- **Alternative**: Check `len(errs) > 0`
- **Decision**: Use explicit count for expected error scenarios

**Platform Testing**
- All validation tests parameterize platform field
- Test both "claude-code" and "opencode" where applicable
- Use table-driven tests for platform scenarios

### Testing Goals

**Coverage Targets**
- `cmd` package: >70% (from current 20.6%)
- `version` package: >80% (from current 0%)
- `models` package: Maintain >90% (currently 91.6%)
- `core` package: Maintain >80% (currently 83.6%)
- `services` package: Maintain >70% (currently 71.4%)

**Quality Standards**
- No custom utility functions when standard library available
- Golden files either automated or removed (no manual-only)
- Platform testing across all supported platforms
- Path resolution works from any working directory
- All public APIs have unit tests

**Test Organization Principles**
- **Unit tests**: Single function/method, isolated dependencies
- **Integration tests**: End-to-end workflows, multiple components
- **Transformation tests**: Input → Output verification (with golden files if automated)
- **Validation tests**: Edge cases, error scenarios, platform-specific constraints

### Risk: Template Complexity May Become Unmaintainable

**Risk**: Complex transformation logic in templates may become difficult to maintain.

**Mitigation**: Break complex logic into custom template functions. Keep templates focused on structure, not algorithmic complexity. Document template functions with examples.

## Migration Plan

1. **Phase 1**: Refactor `internal/models/` to be platform-agnostic (keep existing fields)
2. **Phase 2**: Add OpenCode-specific fields to models (Mode, Temperature, MaxSteps, Hidden, Prompt, Disable for Agent; Subtask for Command; License, Compatibility, Metadata, Hooks for Skill)
3. **Phase 3**: Implement template functions (permission transformation)
4. **Phase 4**: Update `core.RenderDocument()` to register template functions and select templates based on platform parameter
5. **Phase 5**: Update validation to use single `Validate(platform string)` function (platform always required)
6. **Phase 6**: Create `config/templates/opencode/` directory with templates for all document types
7. **Phase 7**: Add comprehensive tests for permission mode transformations (forward)
8. **Phase 8**: Create platform-specific validation functions in `internal/services/transformer.go`
9. **Phase 9**: Create test fixtures for OpenCode (all document types, edge cases)
10. **Phase 10**: Create golden files for OpenCode (all document types)
11. **Phase 11**: Update all test fixtures to use lowercase tool names
12. **Phase 12**: Update tests to use platform parameter in validation
13. **Phase 13**: Update CLI to require --platform flag
14. **Phase 14**: Fix skill name regex to match OpenCode specs
15. **Phase 15**: Update documentation (README, AGENTS.md)
16. **Phase 16**: Run full validation (mise run check, lint, test)

**Rollback strategy**: Keep `config/templates/claude-code/` as backup. Revert by restoring old model structures, removing platform parameter, deleting OpenCode templates.

## Open Questions

None - all design decisions have been resolved based on user requirements and codebase constraints.

## Spec Organization

This change follows the domain-based spec organization defined in `openspec/config.yaml`. Delta specs are created in the change directory and synced to main specs via `openspec archive`.

**Domain Structure:**
- `specs/models/` - Core data structures
- `specs/documents/` - Document lifecycle
- `specs/transformation/` - Platform conversion
- `specs/quality/` - Code quality enforcement
- `specs/infrastructure/` - Infrastructure support and meta-specifications

**Delta Specs for This Change:**
- `specs/documents/document-serialization/` - MODIFIED: Add template function registration
- `specs/transformation/permission-transformation/` - ADDED: New permission transformation logic
- Other specs in domain folders - ADDED: New capabilities (platform-agnostic models, OpenCode transformations, validation)

**References:**
- Template function registration: See `specs/documents/document-serialization/` delta spec
- Permission transformation: See `specs/transformation/permission-transformation/` delta spec

## Concrete Struct Examples

### Agent Model

**BEFORE (Claude Code-centric):**
```go
type Agent struct {
    Name            string   `yaml:"name"`
    Description     string   `yaml:"description"`
    Tools           []string `yaml:"tools"`
    DisallowedTools []string `yaml:"disallowedTools"`
    Model           string   `yaml:"model"`
    PermissionMode  string   `yaml:"permissionMode"`
    Skills          []string `yaml:"skills"`
    FilePath        string   `yaml:"-"`
    Content         string   `yaml:"-"`
}
```

**AFTER (Germinator Format - All Fields Parseable):**
```go
type Agent struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION (platform-agnostic)
    Tools           []string `yaml:"tools,omitempty" json:"tools,omitempty"`
    DisallowedTools []string `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`

    // OPENCODE-SPECIFIC (all parseable from YAML)
    Mode        string  `yaml:"mode,omitempty" json:"mode,omitempty"`
    Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    MaxSteps    int     `yaml:"maxSteps,omitempty" json:"maxSteps,omitempty"`
    Hidden      bool    `yaml:"hidden,omitempty" json:"hidden,omitempty"`
    Prompt      string  `yaml:"prompt,omitempty" json:"prompt,omitempty"`
    Disable     bool    `yaml:"disable,omitempty" json:"disable,omitempty"`

    // CLAUDE CODE-SPECIFIC
    PermissionMode string   `yaml:"permissionMode,omitempty" json:"permissionMode,omitempty"`
    Skills        []string `yaml:"skills,omitempty" json:"skills,omitempty"`

    // COMMON (but different per platform)
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
}
```

### Command Model

**BEFORE:**
```go
type Command struct {
    Name                   string   `yaml:"name"`
    AllowedTools           []string `yaml:"allowed-tools"`
    ArgumentHint           string   `yaml:"argument-hint"`
    Context                string   `yaml:"context,omitempty"`
    Agent                  string   `yaml:"agent,omitempty"`
    Description            string   `yaml:"description"`
    Model                  string   `yaml:"model,omitempty"`
    DisableModelInvocation bool     `yaml:"disable-model-invocation,omitempty"`
    FilePath               string   `yaml:"-"`
    Content                string   `yaml:"-"`
}
```

**AFTER (Germinator Format - All Fields Parseable):**
```go
type Command struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION
    AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

    // OPENCODE-SPECIFIC (parseable from YAML)
    Subtask bool `yaml:"subtask,omitempty" json:"subtask,omitempty"`

    // CLAUDE CODE-SPECIFIC
    ArgumentHint           string `yaml:"argument-hint,omitempty" json:"argument-hint,omitempty"`
    Context                string `yaml:"context,omitempty" json:"context,omitempty"`
    Agent                  string `yaml:"agent,omitempty" json:"agent,omitempty"`
    DisableModelInvocation bool   `yaml:"disable-model-invocation,omitempty" json:"disable-model-invocation,omitempty"`

    // COMMON (but different per platform)
    Model string `yaml:"model,omitempty" json:"model,omitempty"`
}
```

### Skill Model

**BEFORE:**
```go
type Skill struct {
    Name           string   `yaml:"name"`
    Description    string   `yaml:"description"`
    AllowedTools   []string `yaml:"allowed-tools"`
    Model          string   `yaml:"model,omitempty"`
    Context        string   `yaml:"context,omitempty"`
    Agent          string   `yaml:"agent,omitempty"`
    UserInvocable  bool     `yaml:"user-invocable,omitempty"`
    FilePath       string   `yaml:"-"`
    Content        string   `yaml:"-"`
}
```

**AFTER (Germinator Format - All Fields Parseable):**
```go
type Skill struct {
    // REQUIRED (all platforms)
    Name        string `yaml:"name" json:"name"`
    Description string `yaml:"description" json:"description"`
    Content     string `yaml:"-" json:"-"`
    FilePath    string `yaml:"-" json:"-"`

    // TOOL CONFIGURATION
    AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
    DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

    // OPENCODE-SPECIFIC (parseable from YAML)
    License       string            `yaml:"license,omitempty" json:"license,omitempty"`
    Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
    Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
    Hooks        map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`

    // CLAUDE CODE-SPECIFIC
    Model          string `yaml:"model,omitempty" json:"model,omitempty"`
    Context        string `yaml:"context,omitempty" json:"context,omitempty"`
    Agent          string `yaml:"agent,omitempty" json:"agent,omitempty"`
    UserInvocable bool   `yaml:"user-invocable,omitempty" json:"user-invocable,omitempty"`
}
```

### Memory Model

**BEFORE:**
```go
type Memory struct {
    Paths    []string `yaml:"paths"`
    FilePath string   `yaml:"-"`
    Content  string   `yaml:"-"`
}
```

**AFTER (Platform-agnostic):**
```go
type Memory struct {
    // DUAL STORAGE MODES
    Paths   []string `yaml:"paths,omitempty" json:"paths,omitempty"`
    Content string   `yaml:"content,omitempty" json:"content,omitempty"`
    FilePath string  `yaml:"-" json:"-"`
}
```

### Platform-Agnostic Model YAML Example

**Example: Agent with all field types**
```yaml
---
name: code-reviewer
description: Reviews code for security and best practices
model: anthropic/claude-sonnet-4-20250514
tools:
  - bash
  - grep
  - read
disallowedTools:
  - write
  - edit
permissionMode: acceptEdits
skills:
  - skill-creator
mode: primary
temperature: 0.1
maxSteps: 50
hidden: false
prompt: "You are an expert code reviewer..."
disable: false
---
You are a code reviewer...
```

When parsed into the platform-agnostic Agent model:
- Common fields: name, description, model
- Tool configuration: tools (flat array), disallowedTools (flat array)
- Claude Code-specific: permissionMode, skills
- OpenCode-specific: mode, temperature, maxSteps, hidden, prompt, disable
- Metadata: Content (markdown body), FilePath (source location)
