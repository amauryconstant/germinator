## 1. Model Structure Updates

- [ ] 1.1 Add OpenCode fields to Agent model (Mode, Temperature, MaxSteps, Hidden, Prompt, Disable)
- [ ] 1.2 Add Subtask field to Command model
- [ ] 1.3 Add OpenCode fields to Skill model (License, Compatibility, Metadata, Hooks)
- [ ] 1.4 Update model struct tags to separate common from platform-specific fields (use yaml:"-" for platform-specific)
- [ ] 1.5 Update Agent model to support both paths and content for memory
- [ ] 1.6 Update Memory model to support both paths and content fields
- [ ] 1.7 Remove model validation for short model names (sonnet, opus, haiku) - allow full IDs
- [ ] 1.8 Add struct tags for JSON compatibility (json:"fieldname,omitempty")
- [ ] 1.9 Add Agent name regex validation (^[a-z0-9]+(-[a-z0-9]+)*$)

## 2. Template Functions Implementation

**Section 2 depends on: Section 1 complete (model updates)**

- [ ] 2.1 Implement all template functions in internal/core/template_funcs.go
      - transformPermissionMode (Claude Code → OpenCode, single-direction)
      - Add Go doc comments to all functions
      - Handle all 5 Claude Code modes: default, acceptEdits, dontAsk, bypassPermissions, plan
      - Return nil for unknown modes

- [ ] 2.2 Add unit tests for all template functions
      - transformPermissionMode: test all 5 modes + unknown
      - Verify correct permission objects returned
      - Verify nil returned for unknown modes

- [ ] 2.3 Verify all template function tests pass

## 3. Template Function Registration

**Section 3 depends on: Section 2 complete (functions implemented)**

- [ ] 3.1 Implement createTemplateFuncMap() in internal/core/serializer.go
      - Register transformPermissionMode function
      - Add documentation for FuncMap structure
      - Return map[string]interface{} for template usage

- [ ] 3.2 Update RenderDocument to use custom func map
      - Load templates with registered functions
      - Pass funcMap to template parsing
      - Add test verifying transformPermissionMode is available in templates

## 4. Validation Updates

**Section 4 depends on: Section 1 complete (model updates)**

- [ ] 4.1 Update Agent.Validate() to Validate(platform string)
      - Add platform parameter requirement check (return error if empty)
      - Add Agent name format validation (^[a-z0-9]+(-[a-z0-9]+)*$)
      - Add unknown platform error handling
      - Return []error (all errors, not just first)

- [ ] 4.2 Update Command.Validate() to Validate(platform string)
      - Add platform parameter requirement check
      - Add unknown platform error handling
      - Fix Name field YAML tag to yaml:"name"
      - Return []error

- [ ] 4.3 Update Skill.Validate() to Validate(platform string)
      - Add platform parameter requirement check
      - Fix skill name regex to ^[a-z0-9]+(-[a-z0-9]+)*$
      - Add unknown platform error handling
      - Return []error

- [ ] 4.4 Update Memory.Validate() to Validate(platform string)
      - Add platform parameter requirement check
      - Add paths or content required validation
      - Add unknown platform error handling
      - Return []error

- [ ] 4.5 Add tests for multiple validation errors (all models)
      - Verify []error contains all validation issues, not just first
      - Test Agent with missing name AND description
      - Test Command with missing name AND description

- [ ] 4.6 Add tests for platform parameter requirement
      - Verify error returned when platform is empty string
      - Test all four models (Agent, Command, Skill, Memory)

- [ ] 4.7 Add tests for unknown platform error
      - Verify error message when passing "invalid-platform"
      - Verify error lists available platforms: claude-code, opencode
      - Test all four models

## 5. OpenCode Agent Template

**Section 5 depends on: Sections 2-3 complete (functions implemented and registered)**

- [ ] 5.1 Create config/templates/opencode/agent.tmpl with complete structure
      - YAML frontmatter with all Agent fields
      - Mode field (default to "primary")
      - Tools array → {tool: true} map conversion using range
      - DisallowedTools array → {tool: false} map conversion using range
      - Permission object using transformPermissionMode() function
      - Optional fields: temperature, maxSteps, hidden, prompt, disable
      - Model field preservation (full provider-prefixed IDs)
      - Content after frontmatter
      - Omit Claude Code-specific fields (skills list)

- [ ] 5.2 Create golden file tests for agent.tmpl
      - Create test/golden/opencode/code-reviewer-agent.md.golden (minimal)
      - Create test/golden/opencode/agent-full.md.golden (all fields)
      - Create test/golden/opencode/agent-mixed-tools.md.golden (mixed tools)
      - Create golden files for all permission modes

- [ ] 5.3 Add TestRenderOpenCodeAgent tests
      - Table-driven test with minimal agent scenario
      - Table-driven test with full agent scenario
      - Table-driven test with mixed tools scenario
      - Verify template renders correctly
      - Verify golden file output matches

## 6. OpenCode Command Template

**Section 6 depends on: Sections 2-3 complete (functions implemented and registered)**

- [ ] 6.1 Create config/templates/opencode/command.tmpl
      - YAML frontmatter with all Command fields
      - Template field rendering
      - $ARGUMENTS placeholder preservation in content
      - Optional fields: agent, model, subtask
      - Omit Claude Code-specific fields (allowedTools, argumentHint, context, disableModelInvocation)
      - Preserve content indentation and special characters

- [ ] 6.2 Create golden file tests for command.tmpl
      - Create test/golden/opencode/run-tests-command.md.golden (minimal)
      - Create test/golden/opencode/command-full.md.golden (all fields)
      - Create test/golden/opencode/command-with-arguments.md.golden ($ARGUMENTS placeholder)

- [ ] 6.3 Add TestRenderOpenCodeCommand tests
      - Table-driven test for minimal command
      - Table-driven test for command with $ARGUMENTS
      - Table-driven test for full command
      - Verify template rendering preserves content
      - Verify golden file output matches

## 7. OpenCode Skill Template

**Section 7 depends on: Sections 2-3 complete (functions implemented and registered)**

- [ ] 7.1 Create config/templates/opencode/skill.tmpl
       - YAML frontmatter with all Skill fields
       - Name and description fields
       - License field (optional)
       - Compatibility field rendered as YAML list
       - Metadata field rendered as YAML key-value map
       - Hooks field rendered as YAML map (optional)
       - Omit Claude Code-specific fields (allowedTools, userInvocable)

- [ ] 7.2 Create golden file tests for skill.tmpl
      - Create test/golden/opencode/git-workflow-skill/SKILL.md.golden (minimal)
      - Create test/golden/opencode/skill-full.md.golden (all fields)

- [ ] 7.3 Add TestRenderOpenCodeSkill tests
       - Table-driven test for minimal skill
       - Table-driven test for skill with all OpenCode fields
       - Verify license, compatibility, metadata, hooks rendering
       - Verify golden file output matches

## 8. OpenCode Memory Template

**Section 8 depends on: Sections 2-3 complete (functions implemented and registered)**

- [ ] 8.1 Create config/templates/opencode/memory.tmpl
      - AGENTS.md format (no YAML frontmatter)
      - Paths array → @ file references conversion (one per line)
      - Content rendered as project context narrative
      - Teaching instructions for Read tool usage at top
      - Support paths-only, content-only, and both modes
      - Preserve markdown formatting in content

- [ ] 8.2 Create golden file tests for memory.tmpl
      - Create test/golden/opencode/memory-paths-only.md.golden
      - Create test/golden/opencode/memory-content-only.md.golden
      - Create test/golden/opencode/memory-both.md.golden (paths and content)

- [ ] 8.3 Add TestRenderOpenCodeMemory tests
      - Table-driven test for paths-only scenario
      - Table-driven test for content-only scenario
      - Table-driven test for both paths and content
      - Verify @ file references conversion
      - Verify teaching instructions presence
      - Verify golden file output matches

## 9. Platform-Specific Validation Functions

**Section 9 depends on: Section 4 complete (validation signatures updated)**

- [ ] 9.1 Implement validateOpenCodeAgent in internal/services/transformer.go
      - Mode validation (primary/subagent/all)
      - Temperature range validation (0.0-1.0, inclusive)
      - MaxSteps validation (>= 1)
      - Return []error with all violations

- [ ] 9.2 Add tests for validateOpenCodeAgent
      - Valid modes: primary, subagent, all
      - Invalid mode: test error message
      - Temperature boundaries: 0.0 (pass), 0.5 (pass), 1.0 (pass), -0.5 (error), 1.5 (error)
      - MaxSteps boundaries: 1 (pass), 50 (pass), 0 (error), -5 (error)
      - Multiple validation errors test

- [ ] 9.3 Implement validateOpenCodeCommand in internal/services/transformer.go
      - Template field required validation
      - Return []error

- [ ] 9.4 Add tests for validateOpenCodeCommand
      - Template present: verify no error
      - Template empty: verify error message
      - Test with empty string vs nil

- [ ] 9.5 Implement validateOpenCodeSkill in internal/services/transformer.go
      - Name regex validation (^[a-z0-9]+(-[a-z0-9]+)*$)
      - Content required validation
      - Return []error

- [ ] 9.6 Add tests for validateOpenCodeSkill
      - Valid names: git-workflow, code-review-tool-enhanced, git2-operations
      - Invalid names: git--workflow (consecutive hyphens), -git-workflow (leading), git-workflow- (trailing)
      - Invalid names: Git-Workflow (uppercase), git_workflow (underscores)
      - Content present: verify no error
      - Content empty: verify error message

- [ ] 9.7 Implement validateOpenCodeMemory in internal/services/transformer.go
      - Paths or content required validation
      - Return []error

- [ ] 9.8 Add tests for validateOpenCodeMemory
      - Paths only: verify no error
      - Content only: verify no error
      - Both paths and content: verify no error
      - Both empty: verify error message

## 10. Test Fixtures

**Section 10 depends on: Sections 5-8 complete (templates created)**

- [ ] 10.1 Create test/fixtures/opencode directory

- [ ] 10.2 Create Agent fixtures
      - Create test/fixtures/opencode/code-reviewer-agent.md (minimal)
      - Create test/fixtures/opencode/agent-full.md (all fields)
      - Create test/fixtures/opencode/agent-mixed-tools.md

- [ ] 10.3 Create Command fixtures
      - Create test/fixtures/opencode/run-tests-command.md (minimal)
      - Create test/fixtures/opencode/command-full.md (all fields)
      - Create test/fixtures/opencode/command-with-arguments.md

- [ ] 10.4 Create Skill fixtures
      - Create test/fixtures/opencode/git-workflow-skill subdirectory
      - Create test/fixtures/opencode/git-workflow-skill/SKILL.md (minimal)
      - Create test/fixtures/opencode/skill-full.md (all fields)

- [ ] 10.5 Create Memory fixtures
      - Create test/fixtures/opencode/memory-paths-only.md
      - Create test/fixtures/opencode/memory-content-only.md
      - Create test/fixtures/opencode/memory-both.md (paths and content)

## 11. Golden Files

**Section 11 depends on: Sections 5-8 complete (templates created)**

- [ ] 11.1 Create test/golden/opencode directory

- [ ] 11.2 Create Agent golden files
      - Create test/golden/opencode/code-reviewer-agent.md.golden (from agent.tmpl)
      - Create test/golden/opencode/agent-full.md.golden (from agent.tmpl)
      - Create test/golden/opencode/agent-mixed-tools.md.golden (from agent.tmpl)

- [ ] 11.3 Create Command golden files
      - Create test/golden/opencode/run-tests-command.md.golden (from command.tmpl)
      - Create test/golden/opencode/command-full.md.golden (from command.tmpl)
      - Create test/golden/opencode/command-with-arguments.md.golden (from command.tmpl)

- [ ] 11.4 Create Skill golden files
      - Create test/golden/opencode/git-workflow-skill subdirectory
      - Create test/golden/opencode/git-workflow-skill/SKILL.md.golden (from skill.tmpl)
      - Create test/golden/opencode/skill-full.md.golden (from skill.tmpl)

- [ ] 11.5 Create Memory golden files
      - Create test/golden/opencode/memory-paths-only.md.golden (from memory.tmpl)
      - Create test/golden/opencode/memory-content-only.md.golden (from memory.tmpl)
      - Create test/golden/opencode/memory-both.md.golden (from memory.tmpl)

## 12. Transformation Tests

**Section 12 depends on: Sections 5-8, 10-11 complete (templates, fixtures, golden)**

- [ ] 12.1 Add comprehensive Agent transformation tests
      - Table-driven test: minimal agent (name, description, content only)
      - Table-driven test: full agent (all fields)
      - Table-driven test: mixed tools (allowed and disallowed)
       - Table-driven test: all 5 permission modes (default, acceptEdits, dontAsk, bypassPermissions, plan)
       - Table-driven test: agent mode default (empty → "all")
       - Table-driven test: agent mode explicit (primary, subagent)
      - Table-driven test: OpenCode-specific fields (temperature, maxSteps, hidden, prompt, disable)

- [ ] 12.2 Add comprehensive Command transformation tests
      - Table-driven test: minimal command (name, description, content only)
      - Table-driven test: command with $ARGUMENTS placeholder
      - Table-driven test: full command (all optional fields)
      - Table-driven test: subtask field (true, false)
      - Table-driven test: agent and model fields
      - Table-driven test: content preservation and indentation
      - Table-driven test: special characters in content ($, *, #)

- [ ] 12.3 Add comprehensive Skill transformation tests
       - Table-driven test: minimal skill (name, description, content only)
       - Table-driven test: full skill (all OpenCode fields)
       - Table-driven test: license field
       - Table-driven test: compatibility field (YAML list)
       - Table-driven test: metadata field (YAML map)
       - Table-driven test: hooks field (YAML map)
       - Table-driven test: multi-line content preservation
       - Table-driven test: markdown preservation (#, **, -)

- [ ] 12.4 Add comprehensive Memory transformation tests
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

- [ ] 12.5 Add permission transformation tests
       - Table-driven test: default → {"edit": {"*": "ask"}, "bash": {"*": "ask"}}
       - Table-driven test: acceptEdits → {"edit": {"*": "allow"}, "bash": {"*": "ask"}}
       - Table-driven test: dontAsk → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
       - Table-driven test: bypassPermissions → {"edit": {"*": "allow"}, "bash": {"*": "allow"}}
       - Table-driven test: plan → {"edit": {"*": "deny"}, "bash": {"*": "deny"}}
       - Table-driven test: unknown mode → nil

## 13. Update Existing Tests

**Section 13 depends on: Section 4 complete (validation updated)**

- [ ] 13.1 Update test fixtures to use lowercase tool names
      - Find all existing Claude Code fixtures
      - Update tool names from PascalCase to lowercase (if any)
      - Run tests to verify fixtures still parse correctly

- [ ] 13.2 Update all existing test calls to use Validate(platform="claude-code")
      - Find all Validate() calls in tests
      - Update to Validate("claude-code") for Claude Code tests
      - Verify tests compile

- [ ] 13.3 Update golden files to match new output format (if any changes)
      - Compare old and new golden files
      - Update if serialization format changed

- [ ] 13.4 Add test for platform parameter requirement (empty platform returns error)
      - Test all four models
      - Verify error message is clear

- [ ] 13.5 Add test for unknown platform error
      - Test with "invalid-platform"
      - Verify error lists available platforms

## 14. CLI Updates

**Section 14 depends on: Sections 4, 9 complete (validation updated and functions implemented)**

- [ ] 14.1 Add --platform flag to adapt command (required, no default)
      - Add flag definition to Cobra command
      - Document flag in help text
      - Make flag required (return error if not provided)

- [ ] 14.2 Update CLI help text to document --platform flag requirement
      - Add description of --platform flag
      - List available platforms (claude-code, opencode)
      - Document breaking change

- [ ] 14.3 Add validation to ensure --platform is provided (error if empty)
      - Check flag value before processing
      - Return clear error message

- [ ] 14.4 Update CLI help text to list available platforms (claude-code, opencode)
      - Update --help output
      - Ensure platform names are clear

- [ ] 14.5 Add CLI tests for --platform flag requirement
      - Test running adapt without --platform flag (should error)
      - Verify error message is helpful

- [ ] 14.6 Add CLI tests for invalid platform error
      - Test with --platform invalid-platform
      - Verify error message lists available platforms

- [ ] 14.7 Add CLI tests for valid platform (claude-code and opencode)
      - Test with --platform claude-code (should work)
      - Test with --platform opencode (should work)
      - Verify both platforms produce correct output

## 15. Documentation

**Section 15 depends on: All implementation complete**

- [ ] 15.1 Update README.md with OpenCode platform section
      - Add "OpenCode Platform" section
      - Overview of OpenCode support
      - Supported document types (Agent, Command, Skill, Memory)

- [ ] 15.2 Add usage examples for all document types
      - Agent: `germinator adapt agent.yaml opencode-agent.yaml --platform opencode`
      - Command: `germinator adapt command.yaml opencode-command.yaml --platform opencode`
      - Skill: `germinator adapt skill.yaml .opencode/skills/git-workflow/SKILL.md --platform opencode`
      - Memory: `germinator adapt memory.yaml AGENTS.md --platform opencode`

- [ ] 15.3 Add field mapping tables for all models
      - Agent mapping table: Germinator → OpenCode
      - Command mapping table: Germinator → OpenCode
      - Skill mapping table: Germinator → OpenCode
      - Memory mapping table: Germinator → OpenCode

- [ ] 15.4 Document known limitations
      - Permission mode basic approximation (only top-level edit and bash)
       - Skipped fields (skills list, allowedTools, userInvocable, disableModelInvocation)
      - Command-level permission rules not supported
      - DisallowedTools not supported in OpenCode (forward compatibility only)

- [ ] 15.5 Document breaking changes and migration
      - --platform flag requirement (no default)
      - Validate() signature change (platform parameter added)
      - Field name changes (if any)
      - Migration guide for existing users

- [ ] 15.6 Update AGENTS.md
      - Add OpenCode platform support notes
      - Document CLI changes (--platform flag)
      - Add field mapping reference
      - Update usage examples

## 16. Verification

**Section 16 depends on: All previous sections complete**

- [ ] 16.1 Run full validation (mise run check)

- [ ] 16.2 Run linting (mise run lint)

- [ ] 16.3 Run all tests (mise run test)

- [ ] 16.4 Verify test coverage (mise run test:coverage)

- [ ] 16.5 Test end-to-end transformations
      - Agent: germinator adapt input.yaml output.yaml --platform opencode
      - Command: germinator adapt input.yaml output.yaml --platform opencode
      - Skill: germinator adapt input.yaml .opencode/skills/name/SKILL.md --platform opencode
      - Memory: germinator adapt input.yaml AGENTS.md --platform opencode

- [ ] 16.6 Verify golden file tests pass
      - Run go test with golden file comparison
      - Verify all golden files match output

- [ ] 16.7 Verify error messages are descriptive
      - Test invalid agent mode (should mention valid values)
      - Test invalid temperature (should mention range)
      - Test invalid maxSteps (should mention minimum)
      - Test unknown platform (should list available platforms)

- [ ] 16.8 Verify platform-specific validation works
      - OpenCode constraints enforced (mode, temperature, maxSteps)
      - Claude Code still validates correctly
      - Platform parameter required

- [ ] 16.9 Verify all permission modes transform correctly
      - Test all 5 Claude Code modes
      - Verify correct OpenCode permission objects generated
      - Verify unknown mode handled gracefully

- [ ] 16.10 Final integration test
      - Full workflow: parse source → validate → transform → serialize
      - Verify output matches golden files
      - Verify no data loss
      - Verify all document types work
