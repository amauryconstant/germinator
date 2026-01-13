# Proposal: Add Document Processing Infrastructure

## Summary

Build minimal infrastructure to enable core workflows: validate and adapt AI coding assistant documents. This provides template rendering and transformation pipeline with zero premature abstractions.

## Motivation

Current implementation handles parsing and validation, but lacks infrastructure for serialization and CLI workflows. We need:

- **Template Rendering**: Convert document structs back to YAML + markdown format
- **Transformation Pipeline**: Parse → Validate → Render → Write workflow
- **CLI Commands**: validate and adapt commands for users

This ultra-minimal approach follows the existing codebase pattern: simple functions, no interfaces until needed, single-file implementations.

## Proposed Change

**Feature 1: Template Engine (document-serialization)**
- Implement `RenderDocument(doc interface{}, platform string) (string, error)` function
- Load Go templates from `config/templates/<platform>/<docType>.tmpl`
- Render YAML frontmatter with all struct fields
- Preserve markdown body content exactly
- Single file implementation: `internal/core/serializer.go`

**Feature 2: Transformation Pipeline (document-transformation)**
- Implement `TransformDocument(inputPath, outputPath, platform string) error` function
- Orchestrates: LoadDocument → doc.Validate() → RenderDocument → Write
- Fail fast on validation errors
- Write output to specified file
- Single file implementation: `internal/services/transformer.go`

**Feature 3: CLI Commands**
- Add `validate` command: `germinator validate <file>`
- Add `adapt` command: `germinator adapt <input> <output> --platform <platform>`
- Commands defined in `cmd/validate.go`, `cmd/adapt.go`

## Alternatives Considered

1. **DocumentSerializer interface**: Could enable multiple serializers, but this would:
   - Be overkill for single Go template implementation
   - Add indirection without value
   - Defer until we need alternate serialization strategies (JSON, XML, etc.)

2. **ValidationService wrapper**: Could centralize validation, but this would:
   - Be redundant - existing Validate() methods already work
   - Add wrapper around direct function calls
   - Provide zero value for calling doc.Validate() directly from CLI

3. **Adapter system**: Could enable platform transformations, but this would:
   - Be overkill for Claude Code pass-through (MVP)
   - Add unnecessary infrastructure (Adapter structs, config loading)
   - Defer until we actually need cross-platform transformations (Cursor, Windsurf)

4. **Schema inspection**: Could display field requirements, but this would:
   - Require complex reflection to parse Validate() methods
   - Add infrastructure without clear MVP value
   - Users can inspect struct definitions directly or read Claude Code docs

## Impact

**Affected Specs**:
- Add new capability: document-transformation
- Add new capability: document-serialization
- Modify existing capability: cli-framework (add validate, adapt commands)

**Affected Code**:
- Create `internal/core/serializer.go` (template rendering)
- Create `internal/services/transformer.go` (transformation pipeline)
- Create `cmd/validate.go`, `cmd/adapt.go` (CLI commands)
- Create `config/templates/claude-code/{agent,command,skill,memory}.tmpl` (templates)
- Modify `cmd/root.go` (register new subcommands)

**Positive Impacts**:
- Enables document validation and adaptation workflows
- Minimal approach (3 new files, 0 interfaces)
- Clear separation of concerns
- Easy to understand and maintain

**Neutral Impacts**:
- No platform adapter system (deferred until actual cross-platform need)
- No schema inspection command (deferred until clear user need)
- No JSON output format (deferred until needed for tooling)

**No Negative Impacts**

## Dependencies

Depends on `add-core-infrastructure` (document models, parsing, loading complete).

## Success Criteria

1. Template engine renders documents correctly with YAML + markdown
2. Transformation pipeline handles end-to-end workflow
3. Validate command displays errors clearly for invalid documents
4. Adapt command transforms documents successfully
5. Unit tests and integration tests pass for all document types
6. All code passes `mise run validate`

## Validation Plan

- Run `go build ./internal/core/...` and `go build ./internal/services/...`
- Run `go test ./internal/core/... -v -cover`
- Run `go test ./internal/services/... -v -cover`
- Test template rendering with all document types
- Test transformation pipeline with valid and invalid documents
- Run `mise run validate` for final quality check
- Manual testing of CLI commands with sample documents

## Related Changes

This is the Document Processing Infrastructure Milestone - required before Platform Integration Milestone (actual transformations for Cursor, Windsurf, etc.).

## Timeline Estimate

0.75-1.0 days for implementation and testing (reduced from 1.5-2.0 days due to scope reduction).

## Open Questions

None - scope is clear and minimal based on end goals.
