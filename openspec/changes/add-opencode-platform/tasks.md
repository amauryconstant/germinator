## 1. Model Structure Updates

- [x] 1.1 Add OpenCode fields to Agent model (Mode, Temperature (*float64 pointer), Steps, Hidden, Prompt, Disable)
- [x] 1.2 Add Subtask field to Command model
- [x] 1.3 Add OpenCode fields to Skill model (License, Compatibility, Metadata, Hooks)
- [x] 1.4 Update model struct tags to remove yaml:"-" from OpenCode fields and add proper yaml:"field,omitempty" tags
- [x] 1.6 Update Memory model to make Content field parseable from YAML
- [x] 1.7 Remove model validation for short model names (sonnet, opus, haiku) - allow full IDs
- [x] 1.8 Add struct tags for JSON compatibility (json:"fieldname,omitempty")
- [x] 1.9 Add Agent name regex validation (^[a-z0-9-]+(-[a-z0-9]+)*$)

## 2. Template Functions Implementation

- [x] 2.1 Implement all template functions in internal/core/template_funcs.go
       - transformPermissionMode (Claude Code → OpenCode, single-direction)
       - Add Go doc comments to all functions
       - Handle all 5 Claude Code modes: default, acceptEdits, dontAsk, bypassPermissions, plan
       - Return nil for unknown modes

- [x] 2.2 Add unit tests for all template functions
       - transformPermissionMode: test all 5 modes + unknown
       - Verify correct permission objects returned
       - Verify nil returned for unknown modes

- [x] 2.3 Verify all template function tests pass

## 3. Template Function Registration

- [x] 3.1 Implement createTemplateFuncMap() in internal/core/serializer.go
       - Register transformPermissionMode function
       - Add documentation for FuncMap structure
       - Return map[string]interface{} for template usage

- [x] 3.2 Update RenderDocument to use custom func map
       - Load templates with registered functions
       - Pass funcMap to template parsing
       - Add test verifying transformPermissionMode is available in templates

## 4. Validation Updates

- [x] 4.1 Update Agent.Validate() to Validate(platform string)
       - Add platform parameter requirement check (return error if empty)
       - Add Agent name format validation (^[a-z0-9]+(-[a-z0-9]+)*$)
       - Add unknown platform error handling
       - Return []error (all errors, not just first)

- [x] 4.2 Update Command.Validate() to Validate(platform string)
       - Add platform parameter requirement check
       - Add unknown platform error handling
       - Fix Name field YAML tag to yaml:"name"
       - Return []error

- [x] 4.3 Update Skill.Validate() to Validate(platform string)
       - Add platform parameter requirement check
       - Fix skill name regex to ^[a-z0-9]+(-[a-z0-9]+)*$
       - Add unknown platform error handling
       - Return []error

- [x] 4.4 Update Memory.Validate() to Validate(platform string)
       - Add platform parameter requirement check
       - Add paths or content required validation
       - Add unknown platform error handling
       - Return []error

- [x] 4.5 Add tests for multiple validation errors (all models)
       - Verify []error contains all validation issues, not just first
       - Test Agent with missing name AND description
       - Test Command with missing name AND description

- [x] 4.6 Add tests for platform parameter requirement
       - Verify error returned when platform is empty string
       - Test all four models (Agent, Command, Skill, Memory)

- [x] 4.7 Add tests for unknown platform error
       - Verify error message when passing "invalid-platform"
       - Verify error lists available platforms: claude-code, opencode
       - Test all four models

## 5. OpenCode Agent Template

- [x] 5.1 Create config/templates/opencode/agent.tmpl with complete structure
       - YAML frontmatter with all Agent fields
       - Name field (required by OpenCode)
       - Mode field (default to "all")
       - Tools array → {tool: true} map conversion using range
       - DisallowedTools array → {tool: false} map conversion using range
       - Permission object using transformPermissionMode() function
       - Optional fields: temperature, steps, hidden, prompt, disable
       - Model field preservation (full provider-prefixed IDs)
       - Content after frontmatter
       - Omit Claude Code-specific fields (skills list)

- [x] 5.2 Add TestRenderOpenCodeAgent tests
       - Table-driven test with minimal agent scenario
       - Table-driven test with full agent scenario
       - Table-driven test with mixed tools scenario
       - Verify template renders correctly

## 6. OpenCode Command Template

- [x] 6.1 Create config/templates/opencode/command.tmpl
       - YAML frontmatter with all Command fields
       - Template field rendering
       - $ARGUMENTS placeholder preservation in content
       - Optional fields: agent, model, subtask
       - Omit Claude Code-specific fields (allowedTools, argumentHint, context, disableModelInvocation)
       - Preserve content indentation and special characters

- [x] 6.2 Add TestRenderOpenCodeCommand tests
       - Table-driven test for minimal command
       - Table-driven test for command with $ARGUMENTS
       - Table-driven test for full command
       - Verify template rendering preserves content

## 7. OpenCode Skill Template

- [x] 7.1 Create config/templates/opencode/skill.tmpl
       - YAML frontmatter with all Skill fields
       - Name and description fields
       - License field (optional)
       - Compatibility field rendered as YAML list
       - Metadata field rendered as YAML key-value map
       - Hooks field rendered as YAML map (optional)
       - Omit Claude Code-specific fields (allowedTools, userInvocable)

- [x] 7.2 Add TestRenderOpenCodeSkill tests
       - Table-driven test for minimal skill
       - Table-driven test for skill with all OpenCode fields
       - Verify license, compatibility, metadata, hooks rendering

## 8. OpenCode Memory Template

- [x] 8.1 Create config/templates/opencode/memory.tmpl
       - AGENTS.md format (no YAML frontmatter)
       - Paths array → @ file references conversion (one per line)
       - Content rendered as project context narrative
       - Teaching instructions for Read tool usage at top
       - Support paths-only, content-only, and both modes
       - Preserve markdown formatting in content

- [x] 8.2 Add TestRenderOpenCodeMemory tests
       - Table-driven test for paths-only scenario
       - Table-driven test for content-only scenario
       - Table-driven test for both paths and content
       - Verify @ file references conversion
       - Verify teaching instructions presence

## 9. Platform-Specific Validation Functions

- [x] 9.1 Implement validateOpenCodeAgent in internal/services/transformer.go
       - Mode validation (primary/subagent/all)
       - Temperature range validation (0.0-1.0, inclusive)
       - Steps validation (>= 1)
       - Return []error with all violations

- [x] 9.2 Add tests for validateOpenCodeAgent
       - Valid modes: primary, subagent, all
       - Invalid mode: test error message
       - Temperature boundaries: 0.0 (pass), 0.5 (pass), 1.0 (pass), -0.5 (error), 1.5 (error)
       - Steps boundaries: 1 (pass), 50 (pass), 0 (error), -5 (error)
       - Multiple validation errors test

- [x] 9.3 Implement validateOpenCodeCommand in internal/services/transformer.go
       - Template field required validation
       - Return []error

- [x] 9.4 Add tests for validateOpenCodeCommand
       - Template present: verify no error
       - Template empty: verify error message
       - Test with empty string vs nil

- [x] 9.5 Implement validateOpenCodeSkill in internal/services/transformer.go
       - Name regex validation (^[a-z0-9]+(-[a-z0-9]+)*$)
       - Content required validation
       - Return []error

- [x] 9.6 Add tests for validateOpenCodeSkill
       - Valid names: git-workflow, code-review-tool-enhanced, git2-operations
       - Invalid names: git--workflow (consecutive hyphens), -git-workflow (leading), git-workflow- (trailing)
       - Invalid names: Git-Workflow (uppercase), git_workflow (underscores)
       - Content present: verify no error
       - Content empty: verify error message

- [x] 9.7 Implement validateOpenCodeMemory in internal/services/transformer.go
       - Paths or content required validation
       - Return []error

- [x] 9.8 Add tests for validateOpenCodeMemory
       - Paths only: verify no error
       - Content only: verify no error
       - Both paths and content: verify no error
       - Both empty: verify error message

## 10. Test Fixtures

- [x] 10.1 Create test/fixtures/opencode directory

- [x] 10.2 Create Agent fixtures in Germinator format
       - Create test/fixtures/opencode/code-reviewer-agent.md (minimal, Claude Code fields only)
       - Create test/fixtures/opencode/agent-full.md (all fields: Claude Code + OpenCode)
       - Create test/fixtures/opencode/agent-mixed-tools.md

- [x] 10.3 Create Command fixtures in Germinator format
       - Create test/fixtures/opencode/run-tests-command.md (minimal)
       - Create test/fixtures/opencode/command-full.md (all fields)
       - Create test/fixtures/opencode/command-with-arguments.md

- [x] 10.4 Create Skill fixtures in Germinator format
       - Create test/fixtures/opencode/git-workflow-skill subdirectory
       - Create test/fixtures/opencode/git-workflow-skill/git-workflow-skill.md (minimal)
       - Create test/fixtures/opencode/skill-full.md (all fields)

- [x] 10.5 Create Memory fixtures in Germinator format
       - Create test/fixtures/opencode/memory-paths-only.md
       - Create test/fixtures/opencode/memory-content-only.md
       - Create test/fixtures/opencode/memory-both.md (paths and content)

## 11. Golden Files

- [x] 11.1 Create test/golden/opencode directory

- [x] 11.2 Create Agent golden files (from agent.tmpl)
       - Create test/golden/opencode/code-reviewer-agent.md.golden (minimal)
       - Create test/golden/opencode/agent-full.md.golden (all fields)
       - Create test/golden/opencode/agent-mixed-tools.md.golden (mixed tools)
       - Create golden files for all permission modes (default, acceptEdits, dontAsk, bypassPermissions, plan)

- [x] 11.3 Create Command golden files (from command.tmpl)
       - Create test/golden/opencode/run-tests-command.md.golden (minimal)
       - Create test/golden/opencode/command-full.md.golden (all fields)
       - Create test/golden/opencode/command-with-arguments.md.golden ($ARGUMENTS placeholder)

- [x] 11.4 Create Skill golden files (from skill.tmpl)
       - Create test/golden/opencode/git-workflow-skill subdirectory
       - Create test/golden/opencode/git-workflow-skill/git-workflow-skill.md.golden (minimal)
       - Create test/golden/opencode/skill-full.md.golden (all fields)

- [x] 11.5 Create Memory golden files (from memory.tmpl)
       - Create test/golden/opencode/memory-paths-only.md.golden
       - Create test/golden/opencode/memory-content-only.md.golden
       - Create test/golden/opencode/memory-both.md.golden (paths and content)

## 12. Transformation Tests

- [x] 12.1 Add comprehensive Agent transformation tests
       - Table-driven test: minimal agent (name, description, content only)
       - Table-driven test: full agent (all fields)
       - Table-driven test: mixed tools (allowed and disallowed)
       - Table-driven test: all 5 permission modes (default, acceptEdits, dontAsk, bypassPermissions, plan)
       - Table-driven test: agent mode default (empty → "all")
       - Table-driven test: agent mode explicit (primary, subagent)
       - Table-driven test: OpenCode-specific fields (temperature, steps, hidden, prompt, disable)

- [x] 12.2 Add comprehensive Command transformation tests
       - Table-driven test: minimal command (name, description, content only)
       - Table-driven test: command with $ARGUMENTS placeholder
       - Table-driven test: full command (all optional fields)
       - Table-driven test: subtask field (true, false)
       - Table-driven test: agent and model fields
       - Table-driven test: content preservation and indentation
       - Table-driven test: special characters in content ($, *, #)

- [x] 12.3 Add comprehensive Skill transformation tests
       - Table-driven test: minimal skill (name, description, content only)
       - Table-driven test: full skill (all OpenCode fields)
       - Table-driven test: license field
       - Table-driven test: compatibility field (YAML list)
       - Table-driven test: metadata field (YAML map)
       - Table-driven test: hooks field (YAML map)
       - Table-driven test: multi-line content preservation
       - Table-driven test: markdown preservation (#, **, -)

- [x] 12.4 Add comprehensive Memory transformation tests
       - Table-driven test: paths-only scenario
       - Table-driven test: content-only scenario
       - Table-driven test: both paths and content
       - Table-driven test: multiple paths
       - Table-driven test: nested directory paths
       - Table-driven test: @ file reference conversion
       - Table-driven test: teaching instructions presence
       - Table-driven test: markdown preservation in content
       - Table-driven test: relative paths (./)
       - Table-driven test: absolute paths (/)
       - Table-driven test: long content (>1024 chars)
       - Table-driven test: special characters in paths (spaces)

- [x] 12.5 Add permission transformation tests
       - Table-driven test: default → {"edit": {"*": "ask"}, "bash": {"*": "ask"}}
       - Table-driven test: acceptEdits → {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
       - Table-driven test: dontAsk → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
       - Table-driven test: bypassPermissions → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
       - Table-driven test: plan → {"edit": {"*": "deny"}, "bash": {"*": "deny"}}
       - Table-driven test: unknown mode → nil

## 13. Update Existing Tests

- [x] 13.1 Update test fixtures to use lowercase tool names
       - Find all existing Claude Code fixtures
       - Update tool names from PascalCase to lowercase (if any)
       - Run tests to verify fixtures still parse correctly

- [x] 13.2 Update all existing test calls to use Validate(platform="claude-code")
       - Find all Validate() calls in tests
       - Update to Validate("claude-code") for Claude Code tests
       - Verify tests compile

- [x] 13.3 Update golden files to match new output format (if any changes)
       - Compare old and new golden files
       - Update if serialization format changed

- [x] 13.4 Add test for platform parameter requirement (empty platform returns error)
       - Test all four models
       - Verify error message is clear

- [x] 13.5 Add test for unknown platform error
       - Test with "invalid-platform"
       - Verify error lists available platforms

## 14. CLI Updates

- [x] 14.1 Add --platform flag to adapt command (required, no default)
       - Add flag definition to Cobra command
       - Document flag in help text
       - Make flag required (return error if not provided)

- [x] 14.2 Update CLI help text to document --platform flag requirement
       - Add description of --platform flag
       - List available platforms (claude-code, opencode)
       - Document Germinator source format as canonical input
       - Document breaking change (existing Claude Code YAML incompatible)

- [x] 14.3 Add validation to ensure --platform is provided (error if empty)
       - Check flag value before processing
       - Return clear error message

- [x] 14.4 Update CLI help text to list available platforms (claude-code, opencode)
       - Update --help output
       - Ensure platform names are clear

- [x] 14.5 Add CLI tests for --platform flag requirement
       - Test running adapt without --platform flag (should error)
       - Verify error message is helpful

- [x] 14.6 Add CLI tests for invalid platform error
       - Test with --platform invalid-platform
       - Verify error message lists available platforms

- [x] 14.7 Add CLI tests for valid platform (claude-code and opencode)
       - Test with --platform claude-code (should work)
       - Test with --platform opencode (should work)
       - Verify both platforms produce correct output

## 15. Documentation

- [x] 15.1 Update README.md with Germinator source format and OpenCode platform
       - Add "Germinator Source Format" section documenting canonical YAML format
       - Add "OpenCode Platform" section
       - Overview of OpenCode support
       - Supported document types (Agent, Command, Skill, Memory)
       - Document unidirectional transformation flow

- [x] 15.2 Add usage examples for all document types
       - Agent: `germinator adapt agent.yaml opencode-agent.yaml --platform opencode`
       - Command: `germinator adapt command.yaml opencode-command.yaml --platform opencode`
       - Skill: `germinator adapt skill.yaml .opencode/skills/git-workflow/SKILL.md --platform opencode`
       - Memory: `germinator adapt memory.yaml AGENTS.md --platform opencode`

- [x] 15.3 Add field mapping tables for all models
       - Agent mapping table: Germinator (all fields) → Claude Code + OpenCode
       - Command mapping table: Germinator (all fields) → Claude Code + OpenCode
       - Skill mapping table: Germinator (all fields) → Claude Code + OpenCode
       - Memory mapping table: Germinator (all fields) → Claude Code + OpenCode

- [x] 15.4 Document known limitations
       - Permission mode basic approximation (only top-level edit and bash)
       - Skipped fields (skills list, allowedTools, userInvocable, disableModelInvocation)
       - Command-level permission rules not supported
       - DisallowedTools not supported in OpenCode (forward compatibility only)
       - No bidirectional conversion (Germinator → target only)

- [x] 15.5 Document breaking changes and migration
       - Germinator source format as canonical input (breaking change for existing Claude Code YAML)
       - Migration guide for converting existing Claude Code YAML to Germinator format
       - --platform flag requirement (no default)
       - Validate() signature change (platform parameter added)
       - Field name changes (if any)

- [x] 15.6 Update AGENTS.md
       - Add Germinator source format documentation
       - Add OpenCode platform support notes
       - Document CLI changes (--platform flag)
       - Add field mapping reference
       - Update usage examples

## 16. Verification

- [x] 16.1 Run full validation (mise run check)

- [x] 16.2 Run linting (mise run lint)

- [x] 16.3 Run all tests (mise run test)

- [x] 16.4 Verify test coverage (mise run test:coverage)

- [x] 16.5 Test end-to-end transformations
       - Agent: germinator adapt input.yaml output.yaml --platform opencode
       - Command: germinator adapt input.yaml output.yaml --platform opencode
       - Skill: germinator adapt input.yaml .opencode/skills/name/SKILL.md --platform opencode
       - Memory: germinator adapt input.yaml AGENTS.md --platform opencode

- [x] 16.6 Verify golden file tests pass
       - Run go test with golden file comparison
       - Verify all golden files match output

- [x] 16.7 Verify error messages are descriptive
       - Test invalid agent mode (should mention valid values)
       - Test invalid temperature (should mention range)
       - Test invalid steps (should mention minimum)
       - Test unknown platform (should list available platforms)

- [x] 16.8 Verify platform-specific validation works
       - OpenCode constraints enforced (mode, temperature, steps)
       - Claude Code still validates correctly
       - Platform parameter required

- [x] 16.9 Verify all permission modes transform correctly
       - Test all 5 Claude Code modes
       - Verify correct OpenCode permission objects generated
       - Verify unknown mode handled gracefully

- [x] 16.10 Final integration test
       - Full workflow: parse source → validate → transform → serialize
       - Verify output matches golden files
       - Verify no data loss
       - Verify all document types work

## 17. Golden File Testing

- [x] 17.1 Research golden file test patterns
       - Review Go testing patterns for golden file comparison
       - Research best practices for golden file testing in Go projects
       - Document patterns and approaches found

- [x] 17.2 Evaluate current golden files against transformer output
       - List all files in `test/golden/opencode/`
       - Run TransformDocument on corresponding fixtures
       - Compare actual output with existing golden files
       - Document any discrepancies found

- [x] 17.3 Create transformer_golden_test.go with table-driven tests
       - Create `internal/services/transformer_golden_test.go` for golden file testing
       - Use table-driven pattern with test cases for each golden file
       - Load input fixture, run TransformDocument, read expected golden
       - Compare actual output vs expected golden byte-by-byte
       - Use `filepath.Join()` for cross-platform path construction
       - Implement comparison using `cmp.Diff` or byte-by-byte equality

- [x] 17.4 Add update-golden mechanism for regenerating files
       - Add environment variable support for updating golden files
       - Update test documentation with update instructions
       - Add helper function to handle file writing

- [x] 17.5 Update CI to verify golden files match
       - Add test job to CI pipeline to run golden file tests
       - Ensure golden file tests pass before merging
       - Add note about -update-golden flag for developers

- [x] 17.6 Document golden file testing convention in test/README.md
       - Add "Golden File Testing" section to test/README.md
       - Document how to add new golden file tests
       - Explain update-golden workflow
       - Provide examples for new developers

## 18. Testing System Improvements

- [x] 18.1 Remove custom contains() helper function from cmd/cmd_test.go
       - Replace custom `contains()` and `containsMiddle()` functions with `strings.Contains()`
       - Update TestAdaptCommand to use strings.Contains()
       - Run tests to verify no regression

- [x] 18.2 Fix hardcoded platform in integration tests
       - Update TestLoadDocumentIntegration to accept platform parameter
       - Add table-driven test cases for both "claude-code" and "opencode"
       - Verify both platforms load correctly

- [x] 18.3 Add cmd package tests to increase coverage
       - Add tests for adapt.go command functionality
       - Add tests for validate.go command functionality
       - Add tests for root.go setup
       - Add tests for version.go command

- [x] 18.4 Add version package tests
       - Create internal/version/version_test.go
       - Test version variable
       - Test commit variable
       - Test date variable
       - Add edge case tests

- [x] 18.5 Unify test data setup patterns
       - Decide on standard approach: t.TempDir() (dynamic) vs fixtures (static)
       - Update tests to use consistent pattern
       - Document pattern in test/README.md

- [x] 18.6 Fix fragile path resolution in integration tests
       - Replace relative path navigation with robust path resolution
       - Add getProjectRoot() and getFixturesDir() helper functions
       - Ensure tests work from any working directory

- [x] 18.7 Add loader unit tests
       - Create internal/core/loader_test.go
       - Test DetectType() edge cases (empty paths, invalid extensions, etc.)
       - Test LoadDocument() validation error propagation
       - Test DetectType() with all valid document types

- [x] 18.8 Expand test/README.md documentation
       - Document when to use fixtures vs golden files
       - Add section on test naming conventions
       - Document platform testing expectations
       - Add examples for adding new tests
       - Explain table-driven test pattern

- [x] 18.9 Standardize error counting patterns across tests
       - Choose pattern: explicit errorCount field or len(errs) > 0
       - Update inconsistent tests to use chosen pattern
       - Document pattern in test/README.md
       - Decision: Use `errorCount` field for precise validation tests, `len(errs) > 0` for binary pass/fail

- [x] 18.10 Verify and reduce duplicate test coverage
       - Identify overlapping test cases
       - Consolidate overlapping tests
       - Remove redundant assertions
       - Maintain coverage while reducing duplication

- [x] 18.11 Run coverage analysis after improvements
        - Run mise run test:coverage
        - Verify cmd package coverage >70%
        - Verify version package coverage >80%
        - Document coverage metrics in tasks

## 19. Critical Fixes from Review

- [x] 19.1 Add Sprig dependency and replace toLowerCase with lower function
        - Add github.com/Masterminds/sprig/v3 to go.mod
        - Update internal/core/serializer.go to use sprig.FuncMap()
        - Replace all tool name conversion in templates from toLowerCase() to | lower
        - Update documentation to reference Sprig's lower function
        - Run `go mod tidy` after adding dependency
        - Verify templates render lowercase tool names correctly

- [x] 19.2 Change Temperature field from float64 to *float64 pointer
        - Update internal/models/models.go: Temperature *float64
        - Update ValidateOpenCodeAgent to handle nil checks for Temperature
        - Update config/templates/opencode/agent.tmpl condition from ne .Temperature 0.0 to .Temperature
        - Add tests for Temperature nil (omit from output) vs 0.0 (render)
        - Verify nil check logic in validation

- [x] 19.3 Fix field mapping documentation (AGENTS.md and README.md)
        - Update Agent/Command/Skill field mapping tables to reflect actual template output
        - Correct permission transformation format (nested objects, not booleans)
        - Remove "web" permission from docs (not implemented)
        - Distinguish between "parseable from source" vs "output to target"
        - Clarify that OpenCode uses filename as identifier (name field not rendered)
        - Verify documentation matches template rendering behavior

- [x] 19.4 Add CLI integration tests
        - Test adapt command with actual CLI flag parsing and file I/O
        - Test validate command with error messages and exit codes
        - Verify platform flag validation and error handling
        - Test root command help text completeness
        - Target: Increase cmd package coverage from 26.5% to >70%

- [x] 19.5 Resolve name field documentation mismatch
        - Update AGENTS.md to reflect that name field is omitted for OpenCode
        - Explain that OpenCode uses filename as identifier
        - Update opencode-agent-transformation/spec.md to reflect name field omission
        - Verify documentation-to-implementation consistency

## 20. High Priority Fixes

- [x] 20.1 Expand permission transformation to 8+ tools
         - Add mapping for read, grep, glob, list, webfetch, websearch tools
         - Update internal/core/template_funcs.go transformPermissionMode function
         - Document remaining 7+ undefined tools clearly in field mapping tables
         - Add unit tests for newly mapped tools
         - Verify permission objects generated correctly for all 8 tools

- [x] 20.2 Extract platform constants to internal/models/constants.go
         - Create internal/models/constants.go with PlatformClaudeCode and PlatformOpenCode constants
         - Replace magic string literals "claude-code" and "opencode" throughout codebase
         - Update validation functions to use constants instead of strings
         - Update CLI commands to use constants
         - Eliminates 4x code duplication in platform validation

- [x] 20.3 Add tool configuration to command.tmpl
         - Update config/templates/opencode/command.tmpl to include allowedTools/disallowedTools
         - Convert tool arrays to lowercase tool maps using Sprig's | lower
         - Test with commands containing tool restrictions
         - Verify tool configuration not silently dropped

- [x] 20.4 Fix regexp.MatchString error handling
         - Update internal/models/models.go:47 and :189 to handle error from regexp.MatchString
         - Properly check error return value instead of using _
         - Follow Go error handling best practices
         - Add unit test for regex validation error path

- [x] 20.5 Make agent mode field truly optional
         - Update config/templates/opencode/agent.tmpl to omit mode when empty (not default to "all")
         - Update spec to reflect mode as optional field
         - Verify template omits mode when .Mode is empty string
         - Test with agents that have empty mode field

- [x] 20.6 Fix hidden/disable boolean output logic
         - Update config/templates/opencode/agent.tmpl to only output hidden/disable when true
         - Change from {{- if .Hidden}} to {{- if .Hidden}}{{- if eq .Hidden true}}}
         - Verify false values are not rendered in output
         - Test with hidden=false and disable=false scenarios

- [x] 20.7 Add CLI end-to-end tests
         - Test adapt command through CLI with real files
         - Test validate command through CLI with validation errors
         - Verify platform flag validation, error messages, exit codes
         - Add tests for help text completeness
         - Increase cmd package coverage significantly



