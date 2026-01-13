# Design: Core Infrastructure Implementation

## Overview

This design establishes ultra-minimal foundational components for germinator CLI tool. It defines document models, YAML parsing, file loading, and struct validation. This MVP-focused approach eliminates all interfaces, abstractions, and patterns - just straightforward functions that do exactly what's needed.

## Architectural Decisions

### 1. Document Model Design

**Decision**: Define 4 concrete structs (Agent, Command, Memory, Skill) in a single models.go file without interfaces or base struct embedding.

**Rationale**:

**No Document Interface**:
- We only have 4 concrete types
- No polymorphic use cases yet
- Interface is premature abstraction (YAGNI)

**No BaseDocument**:
- Embedding adds complexity for minimal duplication (12 fields vs 1 BaseDocument + 4 embeddings)
- Clearer ownership when each document type owns its fields
- Simpler to understand

**Single File**:
- Only ~250 lines for all 4 document types
- Easier to understand in one place
- File splitting is premature optimization

**Struct Organization**:
```go
// pkg/models/models.go
type Agent struct {
    FilePath string        `yaml:"-"`
    Content  string        `yaml:"-"`
    ID          string    `yaml:"id"`
    LastChanged string    `yaml:"last_changed"`
    Model       string    `yaml:"model"`
    // ... more fields
}

func (a *Agent) Validate() []error {
    // Check required fields
    var errs []error
    if a.ID == "" {
        errs = append(errs, errors.New("id is required"))
    }
    // ... more checks
    return errs
}
```

### 2. YAML Parsing Design

**Decision**: Simple ParseDocument function that extracts YAML frontmatter and parses into appropriate struct based on docType parameter.

**Rationale**:

**No DocumentParser Interface**:
- Only one implementation (YAML parser)
- No alternate parsing strategies needed for MVP
- Interface adds unnecessary indirection

**Function-Based Approach**:
- Simpler than struct + interface pattern
- Direct function call: `ParseDocument(filepath, "agent")`
- Type parameter explicit, no magic

**Delimiter Detection**:
- Standard markdown frontmatter: `---\n...\n---`
- Handle missing delimiters gracefully (entire file is markdown body)
- Extract YAML content between delimiters
- Extract markdown body after second delimiter

**Type-Switch Parsing**:
```go
func ParseDocument(filepath string, docType string) (interface{}, error) {
    // Extract YAML and content
    yamlContent, content, err := extractFrontmatter(filepath)
    if err != nil {
        return nil, err
    }

    // Parse YAML into map
    var data map[string]interface{}
    if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
        return nil, err
    }

    // Switch on type to unmarshal into struct
    switch docType {
    case "agent":
        var doc Agent
        // Unmarshal with type tags
        return &doc, nil
    case "command":
        var doc Command
        return &doc, nil
    // ... memory, skill
    default:
        return nil, fmt.Errorf("unsupported document type: %s", docType)
    }
}
```

### 3. File Loading Design

**Decision**: Simple LoadDocument function that detects type from filename, parses file, and validates struct.

**Rationale**:

**No DocumentFactory**:
- Factory pattern is overkill for linear workflow
- Simple function equivalent
- No dependency injection needed for MVP

**Inline Type Detection**:
- Only 4 regex patterns
- Doesn't warrant separate file
- Clear and obvious in one place

**Linear Workflow**:
- Detect type → parse → validate → return
- No complex state or configuration
- Easy to understand and debug

**Implementation**:
```go
func LoadDocument(filepath string) (interface{}, error) {
    // 1. Detect type from filename
    docType := detectType(filepath) // inline regex patterns
    if docType == "" {
        return nil, fmt.Errorf("unrecognizable filename: %s (expected: agent-*.md, *-agent.md, etc.)", filepath)
    }

    // 2. Parse document
    doc, err := ParseDocument(filepath, docType)
    if err != nil {
        return nil, err
    }

    // 3. Validate
    var errs []error
    switch d := doc.(type) {
    case *Agent:
        errs = d.Validate()
    case *Command:
        errs = d.Validate()
    // ... memory, skill
    }

    if len(errs) > 0 {
        return doc, fmt.Errorf("validation failed: %v", errs)
    }

    return doc, nil
}
```

**Filename Patterns**:
- Agent: `agent-*.md` or `*-agent.md`
- Command: `command-*.md` or `*-command.md`
- Memory: `memory-*.md` or `*-memory.md`
- Skill: `skill-*.md` or `*-skill.md`

### 4. Validation Design

**Decision**: Each document type implements Validate() method that checks struct fields and returns slice of errors.

**Rationale**:

**No DocumentValidator Interface**:
- Each document already has Validate() method
- Interface would be wrapper with zero value
- Completely redundant abstraction

**Method-Based Validation**:
- Direct call: `doc.Validate()`
- Document owns its validation logic
- Clear and explicit

**Error Handling**:
- Return slice of standard Go errors
- Clear, actionable error messages
- Enum validation (e.g., Memory.AppliesTo)

**Validation Examples**:
```go
// Agent validation
func (a *Agent) Validate() []error {
    var errs []error
    if a.ID == "" {
        errs = append(errs, errors.New("id is required"))
    }
    if a.LastChanged == "" {
        errs = append(errs, errors.New("last_changed is required"))
    }
    // ... more field checks
    return errs
}

// Memory enum validation
func (m *Memory) Validate() []error {
    var errs []error
    validAppliesTo := map[string]bool{
        "generic":     true,
        "claude-code": true,
        "opencode":   true,
    }
    if !validAppliesTo[m.AppliesTo] {
        errs = append(errs, fmt.Errorf("applies_to must be one of: generic, claude-code, opencode (got: %s)", m.AppliesTo))
    }
    // ... more field checks
    return errs
}
```

### 5. Why No Abstractions?

This is a radical departure from "best practice", but appropriate for MVP:

**No Interfaces**:
- Single implementation doesn't need abstraction
- Interface before implementation is over-engineering
- Add interface when you have 2+ implementations

**No Factory Pattern**:
- Simple function is equivalent
- Pattern is overkill for linear workflow
- Add pattern when creation logic becomes complex

**No BaseDocument**:
- Embedding adds complexity for minimal benefit
- Duplication is minimal and explicit
- Add base struct when 10+ document types

**No Config Loader Interface**:
- Only read from filesystem
- No multiple loading strategies
- Add interface when you need remote loading or caching

## Integration Points

### With Existing Code

**pkg/models/**:
- Remove placeholder `doc.go`
- Create `models.go` with all 4 document types

**internal/core/**:
- Remove placeholder `doc.go`
- Create `parser.go` with ParseDocument function
- Create `loader.go` with LoadDocument function

**test/fixtures/**:
- Add valid and invalid test fixtures

### With Future Milestones

**Document Type Milestones (2-5)**:
- Will use defined document models
- May add document-specific validation logic

**CLI Integration Milestone**:
- Will use LoadDocument function to load documents
- Will use Validate() methods for validation

**Future Enhancements** (deferred):
- Add Document interface when polymorphic handling needed
- Add DocumentParser interface when second parser needed
- Add TemplateEngine when transforming for other platforms
- Add JSON Schema validation when complex rules needed

## Trade-offs

### Simplicity vs Purity

**Chosen Path**: Ultra-minimal, straightforward implementation.

**Pros**:
- Faster development (0.75-1.0 days vs 2-3 days)
- Easier to understand and maintain
- No cognitive overhead from abstractions
- Clear, linear code flow

**Cons**:
- Less "architecturally pure"
- May need refactoring later when adding features
- Duplicate field declarations (12 fields vs 1 BaseDocument)

### Current vs Future

**Chosen Path**: Build what's needed now, add patterns later.

**Pros**:
- Faster time to working MVP
- Learn from actual usage before adding features
- Avoid over-engineering

**Cons**:
- Will need to add features later (interfaces, patterns)
- May need refactoring when adding features

**Mitigation**:
- Code is structured to easily extract interfaces
- Functions are small and focused
- Comprehensive tests enable safe refactoring

## Risk Mitigation

### Risk: No Interfaces Makes Refactoring Harder

**Mitigation**:
- Keep functions small and focused
- Use type switches that are easy to refactor to polymorphic calls
- Document why we chose this approach
- Add interfaces when we actually need polymorphism

### Risk: Duplication Without BaseDocument

**Mitigation**:
- Document that 12 field declarations is acceptable
- Clearer ownership vs embedding
- Can extract BaseDocument when we have 10+ types

### Risk: Future Refactoring Required

**Mitigation**:
- Write tests that make refactoring safe
- Keep code structured for easy extraction
- Document decisions for future reference

## Success Metrics

1. **Completeness**: All 4 document types defined and working
2. **Parsing**: ParseDocument extracts YAML and parses into correct structs
3. **Loading**: LoadDocument detects type, parses, and validates correctly
4. **Validation**: Validate() methods catch all test cases with clear errors
5. **Test Coverage**: >90% coverage on models and core
6. **Performance**: LoadDocument processes typical file in <100ms
7. **Implementation Speed**: Complete in 0.75-1.0 days

## Implementation Sequencing

**Phase 1: Document Models**
1. Create models.go with all 4 document types
2. Add YAML tags to all fields
3. Implement Validate() methods for all types
4. Write unit tests

**Phase 2: YAML Parsing**
1. Create parser.go with ParseDocument function
2. Extract YAML frontmatter and markdown body
3. Parse YAML into appropriate struct
4. Write unit tests

**Phase 3: Document Loading**
1. Create loader.go with LoadDocument function
2. Implement inline type detection
3. Call ParseDocument and Validate()
4. Write unit tests

**Phase 4: Integration**
1. Create test fixtures
2. Write integration tests
3. Run all tests with coverage
4. Create validation script
5. Final quality checks

## Open Questions

None - scope is ultra-minimal and clear.

## Decisions Made

1. **No Interfaces**: Removed Document, DocumentParser, DocumentValidator (premature)
2. **No Factory**: Use simple LoadDocument function
3. **No BaseDocument**: Each document owns its fields directly
4. **Single Models File**: All 4 types in one models.go
5. **Inline Type Detection**: Patterns in LoadDocument function
6. **Struct Validation Only**: No JSON Schema for MVP
7. **Function-Based**: Simple functions instead of struct+interface patterns
8. **No Template Engine**: Deferred to future milestone
9. **No Platform Adapter**: Deferred to future milestone
