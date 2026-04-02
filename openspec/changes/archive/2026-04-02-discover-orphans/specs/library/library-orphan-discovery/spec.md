# Delta Spec: Library Orphan Discovery - Discover Orphans Enhancement

## MODIFIED Requirements

### Requirement: Discover orphaned resource files

The system SHALL scan library directories recursively for files not registered in library.yaml.

#### Scenario: Discover orphans recursively in skills directory
- **GIVEN** a library with no `skill/nested/deep` registered
- **AND** a file `skills/nested/deep.md` exists with valid frontmatter
- **WHEN** AddResource is called with `--discover`
- **THEN** the file is reported as an orphan with type `skill`, name `nested/deep`

#### Scenario: Discover orphans recursively in all resource directories
- **GIVEN** a library with orphan files in nested directories: `agents/team/reviewer.md`, `commands/git/commit.md`, `memory/notes/todo.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** all orphans from all directories (including nested) are reported

#### Scenario: Discover orphans in deeply nested structure
- **GIVEN** a library with orphan files at multiple nesting levels: `skills/sub1/skill1.md`, `skills/sub1/sub2/skill2.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** both `skill1` and `skill2` are reported as orphans

### Requirement: Enhanced discover result structure

The system SHALL return a DiscoverResult with comprehensive information for batch integration.

#### Scenario: Discover result contains orphans with path information
- **GIVEN** a library with orphan file `skills/commit.md`
- **WHEN** AddResource is called with `--discover`
- **THEN** the result includes orphan with Path, Type, Name, and optional Issue fields

#### Scenario: Discover result contains summary statistics
- **GIVEN** a library with 5 `.md` files scanned and 3 orphans found
- **WHEN** AddResource is called with `--discover`
- **THEN** the result Summary includes TotalScanned=5 and TotalOrphans=3

#### Scenario: Discover result tracks added resources
- **GIVEN** a library with orphan `skills/new-skill.md`
- **WHEN** AddResource is called with `--discover --force`
- **THEN** the result Added contains the successfully registered orphan

#### Scenario: Discover result tracks conflicts
- **GIVEN** a library with existing resource `skill/commit`
- **AND** an orphan file `skills/commit.md` not in library.yaml
- **WHEN** AddResource is called with `--discover`
- **THEN** the result Conflicts contains the conflict with Issue="name_conflict"
