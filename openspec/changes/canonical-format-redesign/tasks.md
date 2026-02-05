## 1. Setup

- [x] 1.1 Create internal/models/canonical/ package structure
- [x] 1.2 Define PermissionPolicy enum (restrictive, balanced, permissive, analysis, unrestricted)
- [x] 1.3 Define canonical Agent struct (PermissionPolicy, Behavior, Tools arrays, Targets map)
- [x] 1.4 Define canonical AgentBehavior struct (Mode, Temperature, MaxSteps, Prompt, Hidden, Disabled)
- [x] 1.5 Define canonical Command struct (Tools, Execution, Arguments, Targets map)
- [x] 1.6 Define canonical CommandExecution struct (Context, Subtask, Agent)
- [x] 1.7 Define canonical CommandArguments struct (Hint)
- [x] 1.8 Define canonical Memory struct (Paths, Content)
- [x] 1.9 Define canonical Skill struct (Tools, Extensions, Execution, Targets map)
- [x] 1.10 Define canonical SkillExtensions struct (License, Compatibility, Metadata, Hooks)
- [x] 1.11 Define canonical SkillExecution struct (Context, Agent, UserInvocable)
- [x] 1.12 Add Validate() methods to all canonical structs (platform-agnostic rules)

## 2. Platform Adapters

- [x] 2.1 Create internal/adapters/ package structure
- [x] 2.2 Define Adapter interface with ToCanonical() and FromCanonical() methods
- [x] 2.3 Create internal/adapters/claude-code/ adapter implementation
- [x] 2.4 Implement ClaudeCodeAdapter.ToCanonical() (parse Claude Code YAML to canonical models)
- [x] 2.5 Implement ClaudeCodeAdapter.FromCanonical() (render canonical models to Claude Code format)
- [x] 2.6 Implement ClaudeCodeAdapter.PermissionPolicyToPlatform() (map canonical policies to Claude Code enum)
- [x] 2.7 Implement ClaudeCodeAdapter.ConvertToolNameCase() (lowercase → PascalCase)
- [x] 2.8 Create internal/adapters/opencode/ adapter implementation
- [x] 2.9 Implement OpenCodeAdapter.ToCanonical() (parse OpenCode YAML to canonical models)
- [x] 2.10 Implement OpenCodeAdapter.FromCanonical() (render canonical models to OpenCode format)
- [x] 2.11 Implement OpenCodeAdapter.PermissionPolicyToPlatform() (map canonical policies to permission objects)
- [x] 2.12 Implement OpenCodeAdapter.ConvertToolNameCase() (lowercase → lowercase, identity)
- [x] 2.13 Implement OpenCode adapter tool list splitting (tools array → tools map, disallowedTools → false values)
- [x] 2.14 Implement OpenCode adapter behavior flattening (behavior.mode → mode, behavior.maxSteps → maxSteps, etc.)
- [x] 2.15 Add shared helper functions (ToPascalCase, ToLowerCase) in adapters package
- [x] 2.16 Define permission mapping data structures (PermissionMapping, PermissionMap, PermissionAction)
- [x] 2.17 Initialize permission mapping table for all 5 policies (restrictive, balanced, permissive, analysis, unrestricted)

## 3. Core Package Updates

- [x] 3.1 Update internal/core/parser.go to parse canonical YAML format
- [x] 3.2 Update internal/core/serializer.go to render canonical models to output
- [x] 3.3 Remove transformPermissionMode() function from internal/core/template_funcs.go
- [x] 3.4 Remove permission mode transformation template function calls from templates
- [x] 3.5 Update internal/core/loader.go to use canonical models instead of Claude Code models
- [x] 3.6 Update internal/models/constants.go to remove Claude Code platform constants (if any)
- [x] 3.7 Add PlatformConfig struct for Targets section (platform-specific configurations)
- [x] 3.8 Update internal/services/validator.go to validate canonical models
- [x] 3.9 Remove platform-specific validation from services (OpenCode-specific validation now in adapters)
- [x] 3.10 Update internal/services/transformer.go to use adapters for conversion
- [x] 3.11 Remove ValidateOpenCode*() methods from services (validation moved to adapters)
- [x] 3.12 Rename canonical AgentBehavior.MaxSteps field to Steps to match OpenCode platform field name
- [x] 3.13 Update OpenCode adapter to use Steps field directly (no conversion needed)
- [x] 3.14 Write unit tests for ClaudeCodeAdapter methods (ToCanonical, FromCanonical, PermissionPolicyToPlatform)
- [x] 3.15 Write unit tests for OpenCodeAdapter methods (ToCanonical, FromCanonical, PermissionPolicyToPlatform)
- [x] 3.16 Write unit tests for shared helper functions (case conversions)

## 4. Templates

- [ ] 4.1 Create config/templates/canonical/ directory for future use (optional)
- [ ] 4.2 Update config/templates/claude-code/agent.tmpl to render from canonical models
- [ ] 4.3 Update config/templates/claude-code/command.tmpl to render from canonical models
- [ ] 4.4 Update config/templates/claude-code/skill.tmpl to render from canonical models
- [ ] 4.5 Update config/templates/claude-code/memory.tmpl to render from canonical models
- [ ] 4.6 Update config/templates/opencode/agent.tmpl to render from canonical models
- [ ] 4.7 Update config/templates/opencode/command.tmpl to render from canonical models
- [ ] 4.8 Update config/templates/opencode/skill.tmpl to render from canonical models
- [ ] 4.9 Update config/templates/opencode/memory.tmpl to render from canonical models
- [ ] 4.10 Remove permission transformation template function calls from all templates
- [ ] 4.11 Add permission policy enum to template context via adapter methods
- [ ] 4.12 Test templates with canonical fixtures for both platforms

## 5. Test Fixtures Conversion

- [ ] 5.1 Convert test/fixtures/agent-valid.md to canonical format
- [ ] 5.2 Convert test/fixtures/agent-invalid.md to canonical format (show validation errors)
- [ ] 5.3 Convert test/fixtures/command-valid.md to canonical format
- [ ] 5.4 Convert test/fixtures/command-invalid.md to canonical format
- [ ] 5.5 Convert test/fixtures/skill-valid.md to canonical format
- [ ] 5.6 Convert test/fixtures/skill-invalid.md to canonical format
- [ ] 5.7 Convert test/fixtures/memory-valid.md to canonical format
- [ ] 5.8 Convert test/fixtures/memory-invalid.md to canonical format
- [ ] 5.9 Convert all test/fixtures/opencode/*.md fixtures to canonical format
- [ ] 5.10 Create new canonical agent fixtures with permission policies (all 5 policies)
- [ ] 5.11 Create new canonical fixtures demonstrating targets section (Claude Code skills, OpenCode overrides)

## 6. Test Updates

- [ ] 6.1 Update internal/models/canonical/*_test.go files for canonical model validation
- [ ] 6.2 Update internal/core/parser_test.go for canonical YAML parsing
- [ ] 6.3 Update internal/core/serializer_test.go for canonical model serialization
- [ ] 6.4 Update internal/services/transformer_test.go for adapter-based transformations
- [ ] 6.5 Add internal/adapters/claude-code/claude_code_adapter_test.go
- [ ] 6.6 Add internal/adapters/opencode/opencode_adapter_test.go
- [ ] 6.7 Remove internal/core/template_funcs_test.go tests for transformPermissionMode()
- [ ] 6.8 Regenerate golden files for both platforms using canonical fixtures
- [ ] 6.9 Run mise run test to verify all tests pass
- [ ] 6.10 Run mise run lint to verify code quality
- [ ] 6.11 Run mise run check to verify all validation passes
- [ ] 6.12 Remove Claude Code-based models from internal/models/ (Agent, Command, Memory, Skill)
- [ ] 6.13 Remove Claude Code-specific validation (permissionMode enum, model aliases) from old Validate()

## 7. Documentation and Cleanup

- [ ] 7.1 Update internal/models/canonical/doc.go with package documentation
- [ ] 7.2 Update internal/adapters/claude-code/doc.go with adapter documentation
- [ ] 7.3 Update internal/adapters/opencode/doc.go with adapter documentation
- [ ] 7.4 Update AGENTS.md with canonical format examples
- [ ] 7.5 Create migration guide document (old Claude Code format → canonical format)
- [ ] 7.6 Document breaking change in release notes
- [ ] 7.7 Update config/AGENTS.md with canonical template information
- [ ] 7.8 Remove obsolete specs (opencode-*-transformation specs from openspec/specs/transformation/) - Add note that these specs are subsumed by platform-adapters capability
- [ ] 7.9 Clarify OpenCode behavior flattening in design.md or create FAQ - Document why canonical behavior object is flattened to top level (OpenCode doesn't support nested behavior objects)
- [ ] 7.10 Run go mod tidy after removing packages

## 8. Verification

- [ ] 8.1 Run mise run check to verify all validation passes
- [ ] 8.2 Run mise run test:coverage to verify test coverage
- [ ] 8.3 Test end-to-end workflow with sample canonical YAML files
- [ ] 8.4 Verify both Claude Code and OpenCode adapters work correctly
- [ ] 8.5 Verify permission policy mapping matches design documentation
- [ ] 8.6 Confirm no Claude Code-specific enums remain in codebase
- [ ] 8.7 Confirm transformPermissionMode() function removed from codebase
- [ ] 8.8 Verify templates render canonical models correctly
