## Context

Germinator currently uses Claude Code's document format as the source standard in `internal/models/`. The models (Agent, Command, Skill, Memory) contain platform-specific fields like `permissionMode` (Claude Code) and `context` (Claude Code fork mode). This tight coupling makes it difficult to support new platforms like OpenCode without significant refactoring.

The current serialization uses Go templates in `config/templates/claude-code/` to render output. However, there are no templates for OpenCode or other platforms. The models are tightly coupled to Claude Code format with hardcoded validation rules.

**Platform differences discovered:**
- **Claude Code**: `permissionMode` enum (default, acceptEdits, dontAsk, bypassPermissions, plan), flat tool arrays (tools, disallowedTools), short model names (sonnet, opus, haiku)
- **OpenCode**: `permission` object with allow/ask/deny values, structured tool config ({"bash": true}), agent mode field (primary/subagent), full model IDs (anthropic/claude-sonnet-4-20250514)
- **Permission systems are fundamentally different** - require complex transformation logic
- **Tool configuration formats don't map 1:1** - need hybrid approach

**Constraints:**
- No new dependencies required
- Must preserve existing test fixtures and golden files
- CLI must work with any input/output paths
- Platform differences handled via Go templates, not YAML adapters
- --platform flag always required (no default behavior)

## Goals / Non-Goals

**Goals:**
- Establish Germinator format as canonical source with ALL platform fields parseable
- Create platform-specific Go templates for serialization (Claude Code and OpenCode)
- Implement permission mode transformation logic from Claude Code to OpenCode
- Establish validation rules that are platform-agnostic (with platform-specific extensions)
- Support tool configurations: flat arrays (Claude Code) and structured objects (OpenCode)
- Add OpenCode-specific configuration fields (Temperature as *float64 pointer, Steps, Hidden, Prompt, Disable, etc.)
- Support full model IDs across platforms (e.g., `anthropic/claude-sonnet-4-20250514`)
- Support both file-based memory (paths) and narrative memory (content)
- Create comprehensive test fixtures and golden files
- Update documentation with OpenCode usage examples
- Extract platform validation logic to reduce code duplication

**Non-Goals:**
- Complete abstraction of all possible AI assistant platforms (focus on Claude Code and OpenCode)
- Runtime platform detection (user explicitly specifies target platform)
- Full validation of platform-specific features at model level (defer to templates and platform-specific validation)
- Automatic migration of user's custom schemas or parsers (document breaking changes)
- YAML-based adapter configuration (use Go templates instead)
- Bidirectional transformation (OpenCode to platform-agnostic model is out of scope)
- Complete feature parity between Claude Code and OpenCode (skip unsupported features silently)
- Automated migration scripts for existing Claude Code users

## Architectural Decisions

### 1. Canonical Source Format: Germinator YAML

**Decision**: Germinator YAML format is the canonical source containing ALL platform fields (Claude Code + OpenCode) in single structs. All fields have YAML and JSON tags for full parseability.

**Rationale**:
- Go doesn't support inheritance, single structs with all fields is idiomatic
- Germinator format serves as single source of truth with complete field set
- All fields parseable from YAML eliminates need for JSON/CLI overrides
- Clear separation of concerns: source (Germinator YAML) → transformation (templates) → output (target platform)
- Templates handle field filtering based on platform, not struct tags

**Trade-offs**:
- Larger structs with unused fields for each platform (acceptable for two platforms)
- Users must understand canonical format includes both platform fields

### 2. Template-Based Serialization

**Decision**: Use Go templates in `config/templates/<platform>/<docType>.tmpl` for serialization. Each platform has its own set of templates that transform platform-agnostic models to platform-specific output format.

**Rationale**:
- Existing codebase already uses Go templates (Claude Code)
- Go templates provide full programming power (conditionals, loops, custom functions)
- No external dependency on YAML configuration loading
- Type-safe at compile time
- Easy to extend for new platforms by adding new template directories

**Trade-offs**:
- More complex than declarative YAML configuration
- Requires template debugging skills for troubleshooting

### 3. Permission Mode Transformation: Claude Code → OpenCode

**Decision**: Implement custom Go template functions to transform Claude Code's `permissionMode` enum to OpenCode's permission objects.

**Rationale**:
- Permission systems are fundamentally incompatible (enum vs. nested object with command keys)
- `dontAsk` means "don't prompt user for approval" → allow both edit and bash
- `bypassPermissions` means "override permission restrictions" → allow both edit and bash
- `plan` is a restricted mode for analysis only → deny both edit and bash
- OpenCode's permission system supports command-level granularity that Claude Code's enum cannot represent
- Basic mapping provides useful approximation for common cases

**Limitations**:
- Only edit and bash tools explicitly mapped; 14 other tools remain undefined
- Complex permission configurations require manual configuration

### 4. Tool Configuration: Flat Array Representation

**Decision**: Use Claude Code's flat array format as the single internal representation. Templates handle transformation to OpenCode's structured object format with required case conversion.

**Rationale**:
- Claude Code format (flat arrays) is simpler and works as base representation
- Templates handle all transformations - cleaner model structure
- Single source of truth in the model
- No ambiguity or synchronization issues
- Simplifies validation logic

**Trade-offs**:
- OpenCode format requires templates to convert arrays to objects
- Case conversion (PascalCase → lowercase) applied in templates using Sprig's `lower` function

### 5. Validation: Single Function with Platform Parameter

**Decision**: Use a single `Validate(platform string)` function that applies both platform-agnostic and platform-specific validation rules. Platform parameter is always required.

**Rationale**:
- Single function is simpler and easier to understand
- No duplication between base and platform validation
- Platform is always required - no ambiguity
- Clear control flow with switch statement
- Easier to add new platforms in future

**Trade-offs**:
- Larger methods with switch statements (manageable for 2 platforms)
- Breaking change but ensures explicit platform selection

### 6. Agent Mode and OpenCode-Specific Fields

**Decision**: Add OpenCode-specific fields to Agent model (Mode, Temperature (*float64), Steps, Hidden, Prompt, Disable) to support OpenCode's agent configuration options. These fields are platform-specific and omitted when serializing to other platforms.

**Rationale**:
- OpenCode uses these fields for fine-grained control over agent behavior
- Preserving these fields enables accurate transformation
- Omitting from Claude Code templates maintains compatibility
- Temperature as `*float64` pointer distinguishes nil (not set) from 0.0 (valid deterministic value)
- Extensible for future platforms with similar configuration options

**Trade-offs**:
- Agent model grows larger with platform-specific fields
- Temperature pointer requires careful nil handling in templates and validation

### 7. Platform Constants

**Decision**: Define platform string constants in `internal/models/constants.go` to eliminate magic string literals throughout codebase.

**Rationale**:
- Platform names are repeated in validation logic across multiple models
- Constants provide compile-time checking for platform name correctness
- Reduces fragility for future platform additions
- Centralized definition makes refactoring easier

**Trade-offs**:
- Additional constants file to maintain
- Requires importing constants package in multiple places

## Architecture

```
Germinator Source Format (YAML)
         │
         ├── Parse → Platform-Agnostic Models (internal/models/)
         │                    │
         │                    ├── Validate(platform) → Errors
         │                    │    ├─ Common: required fields, types
         │                    │    └─ Platform-specific: services/transformer.go
         │                    │
          │                    └─ RenderDocument(platform)
          │                         │
          │                         ├── Select Template (config/templates/<platform>/)
          │                         │    ├─ claude-code/agent.tmpl
          │                         │    └─ opencode/agent.tmpl
          │                         │
          │                         └─ Apply Template Functions (Sprig + custom)
          │                              ├─ transformPermissionMode() (custom)
          │                              └─ Sprig functions: lower, upper, trim, etc.
         │
         └── Output (Platform-Specific)
              ├─ Claude Code format
              └─ OpenCode format
```

## Risks / Trade-offs

### Over-Abstraction Loss of Clarity
Complex platform-agnostic models may become difficult to understand and maintain.

**Mitigation**: Keep models simple, use clear field names, document platform-specific fields with comments. Avoid over-engineering.

### Information Loss During Transformation
Platform-specific features in source format may not map cleanly to target platform, resulting in dropped fields.

**Mitigation**: Document all field mappings clearly in spec files. Skip unsupported features silently (no warnings). Users can manually add missing features after transformation.

### Permission Mode Mapping Complexity
Transforming between Claude Code's `permissionMode` enum and OpenCode's permission objects is complex with edge cases.

**Mitigation**: Document all mapping rules explicitly. Write comprehensive tests for all permission mode transformations.

### Incomplete Permission Transformation
Permission transformation handles 8 tools (edit, bash, read, grep, glob, list, webfetch, websearch), leaving 7+ other OpenCode permissionable tools at undefined state.

**Mitigation**: Document this limitation prominently in field mapping tables. Clearly indicate which tools are mapped vs undefined in implementation.

### Template Complexity vs. Flexibility
Go templates provide full programming power but are harder to debug than declarative YAML.

**Mitigation**: Write comprehensive unit tests for template rendering. Break complex logic into custom template functions.

### Breaking Changes for Existing Users
Users with custom parsers or schemas will need to update their code.

**Mitigation**: Document breaking changes clearly in migration guide. Provide example migration path. Version JSON schemas clearly.

## Migration Plan

1. Refactor `internal/models/` to be platform-agnostic (keep existing fields)
2. Add OpenCode-specific fields to models
3. Implement template functions (permission transformation) using Sprig for string functions
4. Update `core.RenderDocument()` to register template functions (Sprig + custom) and select templates based on platform
5. Update validation to use single `Validate(platform string)` function
6. Create `config/templates/opencode/` directory with templates for all document types
7. Add comprehensive tests for permission mode transformations
8. Create platform-specific validation functions in `internal/services/transformer.go`
9. Create test fixtures for OpenCode (all document types, edge cases)
10. Create golden files for OpenCode (all document types)
11. Update all test fixtures to use lowercase tool names
12. Update tests to use platform parameter in validation
13. Update CLI to require --platform flag
14. Fix skill name regex to match OpenCode specs
15. Update documentation (README, AGENTS.md)
16. Run full validation (mise run check, lint, test)

**Rollback strategy**: Keep `config/templates/claude-code/` as backup. Revert by restoring old model structures, removing platform parameter, deleting OpenCode templates.
