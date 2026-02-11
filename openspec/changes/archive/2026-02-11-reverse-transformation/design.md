## Context

Germinator currently implements forward transformation (canonical → platform) via `services.TransformDocument()` which uses adapter.FromCanonical() and template-based rendering. The adapters (ClaudeCodeAdapter, OpenCodeAdapter) already have ToCanonical() methods implemented but are only used for testing, not production workflows.

**Current state**:

- Forward transformation: canonical.yaml → platform.md (fully implemented)
- Reverse transformation: platform.md → canonical.yaml (not implemented)
- Adapter interface supports bidirectional conversion with ToCanonical() and FromCanonical() methods
- Platform parsing logic exists in adapters but is only used in unit tests

**Constraints**:

- Must preserve platform-agnostic simplicity of canonical format
- Cannot introduce complex dependencies or significant data model changes
- Must follow existing architecture patterns (templates, adapter interface, validation pipeline)
- Platform-specific logic must remain in adapters only
- No comments in code unless explicitly requested
- First iteration: keep scope minimal (no auto-detection, no lossless round-trip)

**Assumptions**:

- Users know the document type of files they want to convert
- Platform files are valid YAML with expected structure
- OpenCode files may have fine-grained permissions that will be collapsed to 5 policies
- Markdown content after YAML frontmatter should be preserved as-is

---

## Goals / Non-Goals

**Goals:**

- Provide CLI command for converting Claude Code or OpenCode documents to canonical YAML format
- Implement reverse transformation pipeline (Parse → Validate → Serialize → Write) mirroring forward transformation
- Reuse existing adapter.ToCanonical() methods for platform parsing
- Create canonical templates for YAML serialization (consistent with platform templates)
- Validate canonical models before writing output
- Support all document types (agent, command, skill, memory) and platforms (claude-code, opencode)

**Non-Goals:**

- Auto-detection of document type from YAML content (requires type flag)
- Lossless round-trip preservation of fine-grained platform permissions
- Batch conversion or directory processing (users can script with bash loop if needed)
- Preservation of all platform-specific fields not in canonical format
- Migration tools for existing configurations
- Dry-run mode or diff output
- Interactive mode

---

## Decisions

### Decision 1: Template-based serialization vs yaml.Marshal

**Choice**: Use canonical templates in `config/templates/canonical/` for serializing canonical models to YAML, consistent with existing platform templates.

**Rationale**:

- **Consistency**: Matches existing architecture pattern for platform templates
- **Formatting control**: Full control over field ordering, indentation, and empty field omission
- **Extensibility**: Easy to add comments, sections, or platform-specific escapes later
- **Future-proof**: Templates can evolve independently of struct tag definitions
- **Familiar patterns**: Existing codebase already has template rendering infrastructure (getTemplatePath, templateContext, Sprig functions)

**Alternatives considered**:

- **yaml.Marshal**: Simpler (one-line implementation), automatic, faster execution. Drawback: No formatting control, different from platform templates, less maintainable if canonical format evolves.
- **Mixed approach**: Use templates for complex types, yaml.Marshal for simple types. Drawback: Inconsistent, harder to maintain, violates single responsibility principle.

---

### Decision 2: User-provided type flag vs auto-detection

**Choice**: Require `--type <agent|command|skill|memory>` flag in CLI command, no auto-detection in first iteration.

**Rationale**:

- **Unambiguous**: User knows what they're converting, explicit flag prevents guessing errors
- **Simple implementation**: No need for complex heuristics or content analysis
- **Fast feedback**: Clear error if user provides wrong type
- **Preserves simplicity**: Keeps canonical format focused on structure, not detection logic
- **Lower maintenance**: No risk of detection edge cases or ambiguous content patterns
- **Consistent with existing pattern**: The forward `adapt` command also requires users to know their document type, making this symmetrical

**Alternatives considered**:

- **Content-based detection**: Inspect YAML fields to infer type (tools, extensions, execution, paths). Drawback: Ambiguous cases (name + tools could be agent, command, or skill), requires complex heuristics, higher maintenance burden.
- **Filename patterns**: Use regex patterns like `agent-*` or `*-agent.md`. Drawback: Inconsistent naming across projects, files may not follow conventions, brittle.

---

### Decision 3: Output file extension (.yaml vs .md)

**Choice**: Use `.yaml` extension for canonical output files.

**Rationale**:

- **Semantic clarity**: Canonical format is configuration data, not documentation
- **Distinguishes formats**: Makes clear which format a file contains (canonical.yaml vs agent.md)
- **Tool compatibility**: YAML tools/editors recognize and handle properly
- **Future validation**: Easy to create schemas or linters for .yaml files
- **Convention**: Most configuration files use .yaml extension

**Alternatives considered**:

- **.md extension**: Consistent with platform documents (all use .md). Drawback: Ambiguous about content, markdown editors may try to render YAML as markdown, less semantically accurate.
- **Mixed extension (.canonical.yaml)**: Very explicit about format. Drawback: Non-standard, longer, potentially confusing.

---

### Decision 4: Content preservation strategy

**Choice**: Preserve markdown content after YAML frontmatter as-is, including leading/trailing whitespace, by extracting everything after the second `---` delimiter.

**Rationale**:

- **Simple**: Reuse existing `extractFrontmatter()` logic from core/parser.go
- **Lossless**: No modification of user's markdown content
- **Consistent**: Works for all document types with content sections
- **Testable**: Existing tests already cover frontmatter extraction

**Alternatives considered**:

- **Parse and re-serialize markdown**: Use a markdown parser to normalize content. Drawback: Complex, may alter formatting, unnecessary for this use case.
- **Strip leading/trailing whitespace**: Normalize content during extraction. Drawback: May remove intentional spacing, harder to preserve exact formatting.

---

### Decision 5: Input validation scope

**Choice**: Validate only canonical models after conversion, not validate input platform files.

**Rationale**:

- **Single validation point**: Canonical models already have comprehensive Validate() methods
- **Leverages existing logic**: No duplication of validation rules
- **Platform-agnostic**: Adapters handle platform-specific issues during parsing
- **Error clarity**: Validation errors are in canonical format, easier to understand

**Alternatives considered**:

- **Validate both platform and canonical**: Two-stage validation before and after conversion. Drawback: More complex, duplicate work, adapter already handles platform structure.
- **No validation until after parsing**: Rely on adapter parsing to catch errors. Drawback: Less specific error messages, may pass invalid canonical structures.

---

### Decision 6: Error handling strategy

**Choice**: Fail fast with descriptive error messages at each pipeline stage, following existing pattern in TransformDocument().

**Rationale**:

- **Consistent**: Matches existing error handling in services/transformer.go
- **Predictable**: User knows exactly where failure occurred (parse, validate, marshal, write)
- **Debuggable**: Clear error messages guide troubleshooting
- **Testable**: Each stage can be tested in isolation

**Alternatives considered**:

- **Collect all errors**: Parse, validate, and marshal all stages even if earlier fails, then report all errors. Drawback: More complex, may waste effort on invalid files, existing pattern is fail-fast.
- **Graceful degradation**: Try to produce partial output even if some stages fail. Drawback: Invalid output is worse than no output, harder to reason about.

---

### Decision 7: Template function map for canonical templates

**Choice**: Create new `canonicalTemplateContext` struct with only Doc field (no Adapter), use minimal Sprig functions (toYaml, fromYaml for edge cases).

**Rationale**:

- **Canonical templates don't need adapters**: No platform-specific conversions required
- **Simpler**: Less context passed to templates, easier to reason about
- **Faster**: No adapter instantiation or platform logic needed
- **Clear**: Canonical format is final state, no transformation needed
- **Avoids confusion**: Using existing `templateContext` with Adapter field suggests adapters are needed, which could lead to accidental adapter method calls

**Alternatives considered**:

- **Reuse existing templateContext**: Include Adapter field even if unused. Drawback: Confusing, suggests adapters are needed, potential for accidental adapter method calls.
- **Use map[string]any{} directly**: Pass simple map to templates. Drawback: Less type-safe, no clear structure for future maintenance.

---

### Decision 8: Content preservation validation in templates

**Choice**: Do not validate markdown content preservation in templates; trust parser's extraction logic.

**Rationale**:

- **Already tested**: `extractFrontmatter()` in parser.go is thoroughly tested in parser_test.go
- **Validation already exists**: Canonical model validation catches missing content with `"paths or content is required"` error
- **No duplication risk**: Adding template validation duplicates logic without adding value
- **Clear separation of concerns**: Templates should render, not validate their input
- **Parser responsibility**: If extraction is broken, it's a parser bug caught by parser tests, not a serialization concern

**Alternatives considered**:

- **Verify Content field in templates**: Add template checks to confirm content was extracted. Drawback: Duplicates parser logic, unnecessary complexity, templates should be simple.
- **Add validation in MarshalCanonical()**: Check Content field before rendering. Drawback: Already handled by canonical model validation, redundant.

---

### Decision 9: CLI --dry-run flag

**Choice**: Do not implement `--dry-run` flag in first iteration; keep scope minimal.

**Rationale**:

- **Consistent with existing pattern**: The `adapt` and `validate` commands don't have dry-run flags
- **Minimal scope**: First iteration focuses on core functionality, not convenience features
- **User workaround**: Users can test with temporary output path: `germinator canonicalize input.md /tmp/test.yaml --platform opencode --type agent`
- **Add on demand**: Common pattern in CLI tools (git, make) to add features based on user feedback
- **No breaking changes**: Can add `--dry-run` in future iteration without affecting existing interface

**Alternatives considered**:

- **Implement --dry-run now**: Preview output without writing to disk. Drawback: Adds implementation complexity, not essential for MVP, existing workaround suffices.
- **Implement --output flag instead of positional argument**: Allow `-o output.yaml` pattern with optional `-` for stdout. Drawback: Different from existing `adapt` command pattern, changes interface.

---

### Decision 10: Validation error message examples

**Choice**: Include examples of correct format in validation error messages for complex fields only (enum values, regex patterns, ranges); omit examples for simple required fields.

**Rationale**:

- **Matches existing pattern**: models.go already includes examples for complex fields:
  - `"permissionPolicy must be one of: restrictive, balanced, permissive, analysis, unrestricted (got: %s)"` (line 110)
  - `"name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$ (got: %s)"` (line 101)
- **Simple fields clear enough**: `"name is required"` and `"description is required"` are self-explanatory
- **Improves UX for complex rules**: Users learn valid values from error messages without consulting documentation
- **Reduces support burden**: Common error scenarios are self-correcting
- **No information overload**: Examples only where helpful, avoiding clutter

**Fields requiring examples**:
- `permissionPolicy`: Enum values (restrictive, balanced, permissive, analysis, unrestricted)
- `name`: Regex pattern `^[a-z0-9]+(-[a-z0-9]+)*$`
- `behavior.mode`: Valid modes (primary, subagent, all)
- `behavior.temperature`: Range (0.0-1.0)

**Alternatives considered**:

- **Include examples for all errors**: Add examples for simple required fields. Drawback: Unnecessary verbosity, "name is required" is self-evident.
- **No examples at all**: Keep errors minimal, rely on documentation. Drawback: Poor UX for new users, higher support burden, doesn't match existing pattern.
- **Link to documentation instead**: Include doc URLs or references. Drawback: Interrupts workflow, requires internet or local docs lookup.

---

## Risks / Trade-offs

### Risk 1: Permission granularity loss on round-trip

[Risk] OpenCode supports command-level permission rules (e.g., `{"bash": {"git push": "deny"}}`) that canonical format's 5 policies cannot represent. Converting platform → canonical → platform will lose fine-grained permissions.

**Mitigation**:

- Document limitation clearly in CLI help text and proposal
- Recommend users edit canonical output directly if they need fine-grained control
- Future iteration: Add `targets.{platform}.raw-permission` escape hatch for preservation
- Most common use cases covered by 5 policies (restrictive, balanced, permissive, analysis, unrestricted)

### Risk 2: User provides wrong document type

[Risk] User specifies `--type agent` but file is actually a skill, leading to parsing errors or unexpected behavior.

**Mitigation**:

- Clear error messages from adapters when document structure doesn't match expected type
- Validation errors will catch missing required fields (e.g., skill needs name, agent needs name + description)
- Document expected structure for each type in CLI help
- Future iteration: Add type auto-detection as suggestion, not requirement

### Risk 3: Template formatting inconsistencies

[Risk] Canonical templates may produce different YAML structure (field ordering, indentation) than users expect, making diffing against manually-written canonical files noisy.

**Mitigation**:

- Use consistent indentation (2 spaces, matches existing templates)
- Maintain logical field ordering (identity first, then tools, then behavior, then targets)
- Run golden file tests to ensure stable output
- Document canonical format structure clearly

### Risk 4: Markdown content edge cases

[Risk] Files without markdown content, or with malformed frontmatter (single `---`, no `---`), may have unexpected content handling.

**Mitigation**:

- Reuse existing `extractFrontmatter()` which handles edge cases
- Return empty string for content if no markdown section found
- Validate canonical models (some require content, e.g., Memory needs paths or content)
- Test with edge cases in unit tests

### Risk 5: Platform-specific field mapping complexity

[Risk] Adapters may not correctly map all platform fields to canonical models, especially edge cases or deprecated fields.

**Mitigation**:

- Leverage existing adapter.ToCanonical() methods (already tested in adapter tests)
- Add comprehensive unit tests for ParsePlatformDocument() covering both platforms
- Use golden file tests to verify full pipeline
- Document which fields are mapped and which are omitted

### Trade-off: Manual type flag vs auto-detection

[Trade-off] Choosing `--type` flag over auto-detection increases user friction but decreases implementation complexity and maintenance burden.

**Rationale**:

- First iteration prioritizes simplicity over convenience
- Users know their document types (they created the files)
- Can add auto-detection in future iteration without breaking changes
- Explicit flag provides clear contract (no ambiguity)

### Trade-off: No lossless round-trip in first iteration

[Trade-off] Not implementing escape hatch for fine-grained permissions means platform → canonical → platform is not lossless, but keeps implementation simple.

**Rationale**:

- Most users don't need command-level permissions
- 5 policies cover 90% of use cases
- Escape hatch adds complexity to adapters and templates
- Can add in future iteration when real need is demonstrated
- Users can edit canonical output directly for advanced use cases

---

## Migration Plan

### Deployment Steps

1. Implement ParsePlatformDocument() in internal/core/platform_parser.go
2. Create canonical templates in config/templates/canonical/
3. Implement MarshalCanonical() in internal/core/serializer.go
4. Implement CanonicalizeDocument() in internal/services/canonicalizer.go
5. Implement CLI command in cmd/canonicalize.go
6. Write unit tests for ParsePlatformDocument() and MarshalCanonical()
7. Write integration tests for CanonicalizeDocument()
8. Write CLI tests for canonicalize command
9. Run mise run check to validate all code quality
10. Run mise run test to verify all tests pass

### Rollback Strategy

- No breaking changes (new command, doesn't modify existing functionality)
- Can rollback by removing cmd/canonicalize.go and associated files
- Git revert provides clean rollback
- Existing adapt and validate commands unchanged

---

## Future Enhancements

Potential improvements for future iterations based on user feedback:

- Add `--dry-run` flag to preview output without writing files
- Implement type auto-detection from YAML content to eliminate `--type` flag requirement
- Support batch conversion for multiple files or directories
- Add `--output` flag with `-` for stdout to enable piping
- Preserve fine-grained platform permissions via escape hatch in `targets.{platform}.raw-permission`
