# yaml-parsing Specification

## Purpose

Provide YAML frontmatter parsing to extract structured data and markdown body from document files.

## ADDED Requirements

### Requirement: YAML Frontmatter Extraction

The system SHALL extract YAML frontmatter from markdown files using standard delimiters.

#### Scenario: Extract YAML between delimiters
**Given** a markdown file with frontmatter: `---\nkey: value\n---\ncontent`
**When** ParseDocument is called
**Then** it SHALL extract "key: value" as YAML
**And** it SHALL extract "content" as markdown body

#### Scenario: Handle missing delimiters
**Given** a markdown file without `---` delimiters
**When** ParseDocument is called
**Then** it SHALL treat entire file as markdown body
**And** it SHALL return an empty YAML map

#### Scenario: Handle empty frontmatter
**Given** a markdown file with empty frontmatter: `---\n---\ncontent`
**When** ParseDocument is called
**Then** it SHALL return an empty YAML map
**And** it SHALL extract "content" as markdown body

---

### Requirement: YAML Parsing by Document Type

The ParseDocument function SHALL parse YAML into appropriate struct based on document type.

#### Scenario: Parse into Agent struct
**Given** frontmatter with Agent fields (id, model, specialization, etc.)
**When** ParseDocument is called with docType "agent"
**Then** it SHALL unmarshal YAML into Agent struct
**And** it SHALL return an Agent pointer

#### Scenario: Parse into Command struct
**Given** frontmatter with Command fields (name, tools, files, etc.)
**When** ParseDocument is called with docType "command"
**Then** it SHALL unmarshal YAML into Command struct
**And** it SHALL return a Command pointer

#### Scenario: Parse into Memory struct
**Given** frontmatter with Memory fields (title, applies_to, etc.)
**When** ParseDocument is called with docType "memory"
**Then** it SHALL unmarshal YAML into Memory struct
**And** it SHALL return a Memory pointer

#### Scenario: Parse into Skill struct
**Given** frontmatter with Skill fields (name, description, etc.)
**When** ParseDocument is called with docType "skill"
**Then** it SHALL unmarshal YAML into Skill struct
**And** it SHALL return a Skill pointer

#### Scenario: Unknown document type returns error
**Given** ParseDocument is called with docType "unknown"
**Then** it SHALL return an error
**And** it SHALL indicate unsupported document type
