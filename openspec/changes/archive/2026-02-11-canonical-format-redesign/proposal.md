## Why

The Germinator source format is fundamentally coupled to Claude Code's format, using Claude Code-specific enums (`permissionMode`), field names, and model aliases. This creates tight coupling throughout the codebase: hardcoded Claude Code validation rules in models, a 103-line `transformPermissionMode()` function, and templates that bolt on OpenCode support. Adding new platforms requires invasive changes to core parsing and validation logic.

## What Changes

**BREAKING**: Replace Claude Code-based source format with domain-driven canonical format expressing intent over platform mechanics.

- Add `permissionPolicy` enum (restrictive, balanced, permissive, analysis, unrestricted) to express permission intent, not Claude Code enums
- Add `targets` section for platform-specific overrides (e.g., `targets.claude-code.skills`) instead of inline platform fields
- Use split tool lists (`tools` + `disallowedTools`) supporting both platforms cleanly
- Add `behavior` object grouping agent settings (`mode`, `temperature`, `maxSteps`, `prompt`, `hidden`, `disabled`)
- Use simple `model` string (user-provided full ID) instead of platform-specific aliases
- Create canonical models in `internal/models/canonical/` independent of platform specifics
- Add platform adapters (`internal/adapters/claude-code`, `internal/adapters/opencode`) converting canonical to/from platform formats
- Replace `transformPermissionMode()` function with simple permission policy mapping table
- Update templates to render from canonical models
- Remove Claude Code-specific validation logic from core models
- Update all test fixtures to canonical format
- Remove `skills` field from canonical agents (move to `targets.claude-code`)

## Capabilities

### New Capabilities

- `canonical-source-format`: Domain-driven YAML format expressing AI coding assistant configuration intent independent of platform specifics, using permission policies, behavior objects, and target-specific overrides.

- `platform-adapters`: Bidirectional conversion between canonical format and platform-specific formats (Claude Code, OpenCode), handling field name conversions (lowercase to PascalCase/lowercase), permission policy mapping, and platform-specific field filtering.

### Modified Capabilities

- `germinator-source-format`: Source format replaced from Claude Code-based to domain-driven canonical format with permissionPolicy enum, targets section, and behavior objects.

- `domain-models`: Models updated from Claude Code-mirroring to domain-driven structures (Agent with PermissionPolicy, Behavior, Targets fields; all document types using canonical concepts).

- `permission-transformation`: Transform function replaced with simple permission policy mapping table (restrictive→default, balanced→acceptEdits, etc.) instead of complex enum-to-object conversion logic.

### Removed Capabilities

- `opencode-agent-transformation`: Transformation logic now subsumed by `platform-adapters` capability. Spec described template-specific rendering from Claude Code-based Germinator format, which is replaced by adapter pattern converting from canonical format.

- `opencode-command-transformation`: Transformation logic now subsumed by `platform-adapters` capability. Spec described template-specific field mappings from Claude Code-based format, replaced by adapter pattern.

- `opencode-skill-transformation`: Transformation logic now subsumed by `platform-adapters` capability. Spec described template-specific field mappings and rendering rules, replaced by adapter pattern.

## Impact

**Code affected:**
- `internal/models/` - New `canonical/` subpackage, remove Claude Code-based models
- `internal/core/` - Parser/serializer updated for canonical format
- `internal/adapters/` - New package with platform adapters
- `config/templates/` - All templates updated to render from canonical models
- `test/fixtures/` - All fixtures converted to canonical format
- `internal/services/` - Remove `transformPermissionMode()`, update validation logic

**Tests affected:**
- All unit tests updated for canonical format
- Golden file tests regenerated with canonical fixtures
- `transformPermissionMode()` tests removed

**Dependencies:**
- No new external dependencies

**Breaking changes:**
- All existing YAML files must be converted to canonical format
- No backward compatibility support (clean break)
