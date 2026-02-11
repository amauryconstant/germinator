## Context

Germinator's current source format is Claude Code's format with OpenCode fields added (permissionMode enum, PascalCase tools, model aliases, Skills field inline). This creates tight coupling: validation logic hardcoded for Claude Code enums, 103-line transformPermissionMode() function converting enum to OpenCode permission object, and templates with platform-specific conditional rendering. Adding a new platform would require invasive changes to core models, validation, and templates.

Constraints:
- **Clean break**: No backward compatibility support (per proposal)
- **Decoupling goal**: Platform-agnostic canonical format expressing intent
- **Scope**: Only 2 platforms (Claude Code, OpenCode) currently, but design must support future additions easily
- **Permission granularity**: Delay fine-grained tool permissions to later change (use simple policies)

## Goals / Non-Goals

**Goals:**
- Define domain-driven canonical YAML format expressing AI coding assistant configuration intent independent of platform specifics
- Implement platform adapters converting canonical format to/from Claude Code and OpenCode
- Replace transformPermissionMode() function with simple permission policy mapping
- Eliminate Claude Code-specific enums and field names from canonical models
- Support easy addition of new platforms without core model changes
- Maintain template-based rendering from canonical models

**Non-Goals:**
- Backward compatibility with old Claude Code-based format (clean break)
- Fine-grained tool permissions (command-level rules like `{"bash": {"git push": "deny"}}`)
- Model alias normalization (user provides full ID as string)
- Bidirectional transformation from any platform to any other platform (only canonical → platform and platform → canonical)

## Decisions

### Decision 1: Permission Policy Enum vs Platform-Specific Enums

**Choice**: Use canonical `permissionPolicy` enum (restrictive, balanced, permissive, analysis, unrestricted) instead of Claude Code's `permissionMode` enum (default, acceptEdits, dontAsk, bypassPermissions, plan).

**Rationale:**
- **Intent over implementation**: `permissionPolicy` expresses security posture without naming specific platform features
- **Platform-agnostic**: Policy names don't reference Claude Code concepts ("acceptEdits" is Claude Code terminology)
- **Simple mapping**: Straightforward table mapping canonical policy to platform-specific values (restrictive→default, balanced→acceptEdits)
- **Future-proof**: New platform can map policies to its permission system without changing canonical format

**Alternatives considered:**
- Keep Claude Code permissionMode enum: Would preserve coupling, not solve core problem
- OpenCode-style permission objects: Too complex for canonical format, we're delaying fine-grained permissions
- No permission policy: Would lose important configuration capability, both platforms need permission concepts

### Decision 2: Targets Section vs Inline Platform Fields

**Choice**: Use `targets` section for platform-specific overrides (e.g., `targets.claude-code.skills`) instead of inline Claude Code/OpenCode fields.

**Rationale:**
- **Clear boundaries**: Explicit separation between platform-independent and platform-specific configuration
- **Avoids clutter**: Main sections don't mix platforms (no `permissionMode` + `mode` at same level)
- **Easy to skip**: Templates can check `targets.claude-code` existence and include or omit entire section
- **Extensible**: New platform adds `targets.new-platform` key without modifying canonical models

**Alternatives considered:**
- Inline platform fields (current): Creates confusion, which fields apply to which platform?
- Separate YAML files per platform: Defeats single source of truth purpose
- Platform-specific top-level keys (e.g., `claude-code: {...}`): Similar to targets, but `targets` is more explicit about "where these configurations go"

### Decision 3: Split Tool Lists vs Single List

**Choice**: Use split lists `tools` and `disallowedTools` instead of single list or boolean map in canonical format.

**Rationale:**
- **Both platforms support it**: Claude Code has tools + disallowedTools; OpenCode has tools map (true/false values)
- **Simple representation**: Two string arrays are easy to read and write
- **Adapter handles conversion**: OpenCode adapter converts split lists to map, Claude Code adapter uses as-is
- **Explicit disallowed**: Clearer than mixing positive and negative in one structure

**Alternatives considered:**
- Single tools list: Can't represent disallowed tools
- Boolean map (`{bash: true, write: false}`): More complex, unnecessary in canonical (adapter handles platform differences)

### Decision 4: Behavior Object vs Inline Agent Fields

**Choice**: Group agent settings into `behavior` object (mode, temperature, maxSteps, prompt, hidden, disabled) instead of inline fields.

**Rationale:**
- **Logical grouping**: These fields all control agent execution behavior, not identity or tools
- **Separates concerns**: `name`, `description`, `model`, `tools` define what agent IS; `behavior` defines how it ACTS
- **Clearer YAML**: Top-level has 4 sections: identity (name, description), tools, behavior, model, targets
- **Extensible**: New behavior-related fields can be added without cluttering top level

**Alternatives considered:**
- Inline fields (current): Flat structure, but mixes concerns and makes YAML harder to scan
- Multiple sections (behavior, execution, etc.): Over-engineering for 4-5 fields

### Decision 5: Simple Model String vs Aliases Object

**Choice**: Use simple `model` string (user-provided full ID like `anthropic/claude-sonnet-4-20250514`) instead of aliases object or platform-specific fields.

**Rationale:**
- **User control**: Models often customized by users, let them provide exact ID they want
- **No normalization complexity**: Avoid alias mapping tables (sonnet→full-id), user handles provider-specific IDs
- **Pass-through simplicity**: Adapter passes model string to output as-is (after case conversion if needed)
- **Future-proof**: New model formats (provider/id) work without changes

**Alternatives considered:**
- Aliases object (`{id: "...", claude-code: "sonnet"}`): Adds complexity, user rarely needs aliases
- Platform-specific fields: Brings back coupling we're trying to remove
- No model field (use defaults): Removes important user configuration option

### Decision 6: Adapter Pattern vs Enhanced Templates

**Choice**: Create platform adapters (`internal/adapters/claude-code`, `internal/adapters/opencode`) with bidirectional conversion methods instead of only template-based transformation.

**Rationale:**
- **Separation of concerns**: Templates handle rendering; adapters handle platform-specific logic
- **Bidirectional support**: Can parse existing Claude Code/OpenCode files into canonical format (future feature)
- **Testable**: Adapters are pure functions, easier to unit test than templates
- **Extensible**: New platform = new adapter + template, no core changes

**Alternatives considered:**
- Template-only transformation (current): Puts logic in templates, hard to test, limited reuse
- Config-driven mapping (YAML field mappings): More moving parts, debugging complexity, runtime overhead

### Decision 7: Permission Policy Mapping Table vs Transform Function

**Choice**: Replace 103-line `transformPermissionMode()` function with simple mapping table (struct/map lookup).

**Rationale:**
- **Simpler code**: Map lookup replaces switch statement with YAML generation
- **Easier to maintain**: Add new policy = add map entry, no function logic changes
- **Clear mappings**: Table format shows canonical→Claude Code, canonical→OpenCode mappings explicitly
- **Remove complexity**: No need for string building in Go code

**Implementation:**
```go
var permissionPolicyMappings = map[string]map[string]PermissionMapping{
    "restrictive": {
        claudeCode: "default",
        openCode: PermissionMap{Edit: Ask, Bash: Ask, ...},
    },
    "balanced": {
        claudeCode: "acceptEdits",
        openCode: PermissionMap{Edit: Allow, Bash: Ask, ...},
    },
    // ...
}
```

**Alternatives considered:**
- Keep transformPermissionMode(): 103 lines of switch statements and string building, complex to maintain
- Config-driven mapping: YAML file for mappings, adds runtime parsing, harder to debug

### Decision 8: Separated Validation vs Unified Validation

**Choice**: Use two-stage validation approach - canonical models validate domain structure (platform-agnostic), adapters validate platform-specific constraints.

**Rationale:**
- **Clear ownership**: Canonical models own structure validation (required fields, format rules, domain constraints); adapters own platform rules (platform-specific fields, constraints)
- **Platform independence**: Canonical models remain truly platform-agnostic, no knowledge of target platforms in model layer
- **Easier testing**: Each component can be tested in isolation without platform dependencies
- **Extensibility**: Adding new platform requires only new adapter validation, no changes to canonical models
- **Single responsibility**: Models validate structure, adapters validate platform compliance

**Implementation:**
- Canonical `Validate()` methods check: required fields, name regex, permissionPolicy enum values, temperature range, tool array structure
- Adapter `FromCanonical()` methods check: platform-specific field presence/absence, platform-specific constraints (e.g., Claude Code skills array format)
- Validation errors reported in sequence: structural errors from canonical models first, platform-specific errors during adapter conversion

**Alternatives considered:**
- Unified validation in adapters: Would require adapters to know canonical validation rules, creating duplication and coupling
- Platform-aware canonical models: Would violate platform-agnostic design goal, require platform parameter in Validate() methods
- Single validation pass: Would blur separation of concerns, make code harder to maintain

## Risks / Trade-offs

### Risk 1: User Migration Effort

[Risk] All existing YAML files must be converted to canonical format (clean break). Users with many configurations face manual or tool-assisted migration.

**Mitigation**:
- Provide clear migration guide documenting field mapping (old→new)
- Example side-by-side comparisons (old Claude Code format vs canonical format)
- Suggest providing migration tool (`germinator migrate`) in future change
- Document breaking change prominently in release notes

### Risk 2: Adapter Complexity

[Risk] Adapters must handle field name conversions (lowercase to PascalCase/lowercase), permission mapping, and platform-specific field filtering. Logic may spread across multiple adapter methods.

**Mitigation**:
- Use shared helper functions in adapter package (e.g., `convertToolNameCase`)
- Write comprehensive unit tests for adapter methods
- Document adapter responsibilities clearly (what each method does)
- Use maps for name conversions (no switch statements)

### Risk 3: Template Rendering Complexity

[Risk] Templates must render from canonical models but target different formats (name field presence, tool case, permission object vs enum). May need complex conditional logic.

**Mitigation**:
- Keep templates simple: use if/else for platform-specific fields, avoid template logic
- Leverage Sprig functions (lower, upper) for case conversions
- Adapter methods provide transformed values when needed (e.g., `adapter.PermissionObject(policy)`)
- Test templates with representative fixtures for each platform

### Risk 4: Permission Granularity Limitations

[Risk] Canonical permission policies (5 levels) can't represent command-level rules (`{"bash": {"git push": "deny"}}`) that OpenCode supports but Claude Code doesn't. Users needing this must edit output directly.

**Mitigation**:
- Document limitation clearly in proposal and design
- Plan future change for fine-grained permissions when needed
- Ensure `targets` section can include platform-specific permission objects (escape hatch)
- Most use cases covered by 5 policies (common security postures)

### Risk 5: New Platform Integration Effort

[Risk] While architecture is designed for easy platform addition, each new platform still requires: adapter implementation, template set, and validation rules.

**Mitigation**:
- Provide adapter interface/template as clear example (Claude Code adapter serves as reference)
- Document platform integration process step-by-step
- Share helper functions across adapters (field name conversions, etc.)
- Add integration tests verifying new platform works end-to-end
