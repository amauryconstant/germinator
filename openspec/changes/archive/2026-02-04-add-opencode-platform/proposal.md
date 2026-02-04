# Proposal: Add OpenCode Platform Support

## Why

Germinator currently uses Claude Code's document format as the source standard, tightly coupling implementation to one platform's conventions. This limits extensibility and makes it difficult to support new platforms like OpenCode without significant refactoring.

To add OpenCode as a target platform, we need:
- Canonical source format containing all platform fields
- Platform-agnostic models with transformation infrastructure
- OpenCode templates and validation
- Unidirectional transformation (Germinator → target platform)

This change provides a complete solution: a Germinator source format that serves as the single source of truth, with concrete OpenCode transformation logic implemented through Go templates.

## What Changes

### Canonical Source Format
- Germinator YAML format as single source of truth containing all platform fields
- All fields parseable from YAML (no `yaml:"-"` tags)
- Platform-specific fields included in same structs

### Transformation Infrastructure
- Go templates for both Claude Code and OpenCode serialization
- Permission mode transformation: Claude Code enum → OpenCode permission object
- Tool name case conversion: PascalCase → lowercase for OpenCode via Sprig library
- Template functions: `transformPermissionMode()`, Sprig's `lower` function

### New OpenCode-Specific Fields
- **Agent**: Mode, Temperature (*float64 pointer for nil/0.0 distinction), Steps, Hidden, Prompt, Disable
- **Command**: Subtask
- **Skill**: License, Compatibility, Metadata, Hooks
- **Memory**: Support both file paths and narrative content

## Impact

### New Files
- `internal/core/template_funcs.go` - Custom template functions
- `config/templates/opencode/*.tmpl` - 4 template files (agent, command, skill, memory)
- `test/fixtures/opencode/` - Fixture files for all document types
- `test/golden/opencode/` - Golden files for all document types

### Modified Files
- `internal/models/models.go` - Platform-agnostic model structures
- `internal/core/serializer.go` - Template function registration
- `internal/services/transformer.go` - OpenCode validation functions
- `README.md` - OpenCode usage examples
- `AGENTS.md` - Field mapping notes and CLI changes

### Breaking Changes
- **Germinator YAML format is now canonical source** - Existing Claude Code-only YAML files incompatible
- **Validate() signature change**: `Validate()` → `Validate(platform string)` - platform always required
- **--platform flag always required** - No default to Claude Code
- Users must migrate existing Claude Code YAML to Germinator format

### Limitations
- Permission transformation: Approximation with 8 tools mapped (edit, bash, read, grep, glob, list, webfetch, websearch), 7+ tools remain undefined
- No bidirectional transformation (Germinator → target only)
- Some Claude Code fields skipped (skills, userInvocable, disableModelInvocation, allowedTools)
- Command-level permission rules not supported in transformation

## Risks

### Tool Name Case Conversion Errors
Claude Code uses PascalCase tool names (Bash, Read, Edit) but OpenCode requires lowercase (bash, read, edit). If templates don't perform case conversion consistently, OpenCode will not recognize tools.

**Mitigation**: Use Sprig library's `lower` function and apply it to all tool name outputs in OpenCode templates. Add comprehensive tests verifying tool names are correctly lowercased.

### Incomplete Permission Transformation
Permission transformation handles 8 tools (edit, bash, read, grep, glob, list, webfetch, websearch), leaving 7+ other OpenCode permissionable tools at undefined state. Users may have incorrect expectations about permission behavior.

**Mitigation**: Document this limitation prominently in field mapping tables. Clearly indicate which tools are mapped vs undefined in implementation.

### Documentation Accuracy Gaps
Field mapping documentation may claim fields as "supported" that are not actually output to target format, leading to user confusion about data loss.

**Mitigation**: Maintain strict synchronization between templates and documentation. Distinguish between "parseable from source" vs "output to target". Add validation tests verifying actual template output matches documented behavior.

### Breaking Changes for Existing Users
Users with custom schemas or parsers will need to update their code.

**Mitigation**: Document breaking changes clearly in migration guide. Provide example migration path. Version JSON schemas clearly.
