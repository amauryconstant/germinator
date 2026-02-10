## 1. Core Implementation

- [ ] 1.1 Create internal/core/platform_parser.go file with ParsePlatformDocument(path, platform, docType) function that reads platform YAML file, unmarshals to map[string]interface{}, instantiates adapter (claudecode.New() for "claude-code", opencode.New() for "opencode"), calls adapter.ToCanonical(), wraps result in Canonical* struct (adding FilePath and Content), and returns canonical model or error

- [ ] 1.2 Add MarshalCanonical() function to internal/core/serializer.go that defines canonicalTemplateContext struct with Doc field only, determines docType from canonical model, loads canonical template from config/templates/canonical/{docType}.tmpl, creates canonicalTemplateContext instance, executes template with minimal Sprig functions, returns YAML string or error

- [ ] 1.3 Create internal/services/canonicalizer.go file with CanonicalizeDocument(inputPath, outputPath, platform, docType) function that calls ParsePlatformDocument(), validates canonical model using existing Validate() methods, calls MarshalCanonical(), writes output to file with os.WriteFile(outputPath, yamlBytes, 0644), returns nil on success or error on any stage failure

## 2. Template Creation

- [ ] 2.1 Create config/templates/canonical/ directory structure

- [ ] 2.2 Create config/templates/canonical/agent.tmpl template with YAML frontmatter structure rendering name, description, tools array (if non-empty), disallowedTools array (if non-empty), permissionPolicy (if set), behavior object (if any behavior fields set), model (if set), extensions.hooks (if non-empty), targets section (if set), and Content field after second ---

- [ ] 2.3 Create config/templates/canonical/command.tmpl template with YAML frontmatter structure rendering name, description, tools array (if non-empty), execution object (if any execution fields set), arguments.hint (if set), model (if set), and Content field after second ---

- [ ] 2.4 Create config/templates/canonical/skill.tmpl template with YAML frontmatter structure rendering name, description, tools array (if non-empty), extensions object (if any extension fields set), execution object (if any execution fields set), model (if set), and Content field after second ---

- [ ] 2.5 Create config/templates/canonical/memory.tmpl template with YAML frontmatter structure rendering paths array (if non-empty), content field with pipe syntax (if non-empty and paths empty), and Content field after second ---

## 3. CLI Command

- [ ] 3.1 Create cmd/canonicalize.go file with canonicalizeCmd Cobra command that defines Usage as "canonicalize <input> <output>", Short description as "Convert a platform document to canonical format", Long description with supported platforms and document types, Requires ExactArgs(2) for input and output paths

- [ ] 3.2 Add platform flag to canonicalizeCmd with StringVar pointing to platform variable, default empty string, description "Source platform (required: claude-code, opencode)", and MarkFlagRequired("platform")

- [ ] 3.3 Add type flag to canonicalizeCmd with StringVar pointing to docType variable, default empty string, description "Document type (required: agent, command, skill, memory)", and MarkFlagRequired("type")

- [ ] 3.4 Implement Run function that validates both platform and docType flags are set (exit with error to stderr if missing), calls services.CanonicalizeDocument(inputPath, outputPath, platform, docType), prints error message to stderr and os.Exit(1) if error returned, prints success message to stdout on completion

- [ ] 3.5 Register canonicalizeCmd with rootCmd.AddCommand() in init() function

## 4. Testing

- [ ] 4.1 Write unit tests for internal/core/platform_parser.go covering ParsePlatformDocument() with valid Claude Code agent, valid OpenCode agent, invalid YAML syntax, file not found, content preservation with frontmatter, all four document types for both platforms

- [ ] 4.2 Write unit tests for internal/core/serializer.go MarshalCanonical() function covering canonicalTemplateContext struct usage, canonical agent with all fields, canonical agent with minimal fields, canonical command, canonical skill, canonical memory with paths only, canonical memory with content only, canonical model with all empty optional fields, verify no Adapter field is accessed in templates

- [ ] 4.3 Write integration tests for internal/services/canonicalizer.go covering CanonicalizeDocument() successful pipeline for all document types, fail-fast on parse error, fail-fast on validation error, file write error, round-trip both platforms

- [ ] 4.4 Write CLI tests in cmd/cmd_test.go for canonicalizeCmd covering valid invocation with all flags, missing --platform flag, missing --type flag, invalid platform value, invalid type value, non-existent input file, successful conversion with output confirmation

- [ ] 4.5 Create test fixtures in test/fixtures/claude-code/ directory with agent.md, command.md, skill.md, memory.md files in Claude Code format

- [ ] 4.6 Create test fixtures in test/fixtures/opencode/ directory with agent.md, command.md, skill.md, memory.md files in OpenCode format

- [ ] 4.7 Create golden files in test/golden/canonical/ directory with agent.yaml.golden, command.yaml.golden, skill.yaml.golden, memory.yaml.golden showing expected canonical YAML output

## 5. Validation and Quality

- [ ] 5.1 Run mise run check to verify all code quality checks pass (gofmt, govet, golangci-lint, YAML/TOML/JSON validation, file hygiene)

- [ ] 5.2 Run mise run test to verify all tests pass including new unit tests, integration tests, CLI tests, and golden file tests

- [ ] 5.3 Run mise run test:coverage to verify test coverage is sufficient for new code paths

- [ ] 5.4 Run mise run build to verify binary compiles successfully with new canonicalize command included
