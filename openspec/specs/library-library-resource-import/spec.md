# Capability: Library Resource Import

## Purpose

The Library Resource Import capability enables importing existing canonical or platform documents into the library. It handles type detection, canonicalization, validation, file copying, and library.yaml updates.

## Requirements

### Requirement: Import resource to library

The system SHALL import a resource from a source file to the library.

#### Scenario: Import canonical skill to library
- **GIVEN** a library at `/tmp/library` and source file `skill-commit.md` in canonical format
- **WHEN** AddResource is called with source `skill-commit.md`
- **THEN** the file is copied to `library/skills/skill-commit.md` and library.yaml is updated with entry `skill/commit: {path: skills/skill-commit.md, description: ...}`

#### Scenario: Import platform document and canonicalize
- **GIVEN** a library and source file `code-reviewer.md` in OpenCode format with type agent
- **WHEN** AddResource is called with source `code-reviewer.md` and platform `opencode`
- **THEN** the document is canonicalized before being copied to `library/agents/code-reviewer.md`

#### Scenario: Import with explicit type override
- **GIVEN** a library and source file with ambiguous name but frontmatter indicating skill type
- **WHEN** AddResource is called with `--type skill`
- **THEN** the resource is treated as skill regardless of filename or frontmatter

### Requirement: Auto-detect resource type

The system SHALL detect resource type from flag, frontmatter, or filename.

#### Scenario: Detect type from --type flag
- **GIVEN** a source file with no type indication
- **WHEN** AddResource is called with `--type agent`
- **THEN** type is determined to be `agent`

#### Scenario: Detect type from frontmatter
- **GIVEN** a source file with frontmatter containing `type: skill`
- **WHEN** AddResource is called without `--type` flag
- **THEN** type is detected as `skill`

#### Scenario: Detect type from filename pattern
- **GIVEN** a source file named `agent-reviewer.md`
- **WHEN** AddResource is called without `--type` flag and no frontmatter type
- **THEN** type is detected as `agent` from filename pattern

#### Scenario: Unknown type detection fails
- **GIVEN** a source file with no type indication and non-matching filename
- **WHEN** AddResource is called
- **THEN** an error is returned indicating type could not be detected

### Requirement: Auto-detect resource name

The system SHALL detect resource name from flag or frontmatter.

#### Scenario: Detect name from --name flag
- **GIVEN** a source file with frontmatter name `old-name`
- **WHEN** AddResource is called with `--name new-name`
- **THEN** the resource name is `new-name`

#### Scenario: Detect name from frontmatter
- **GIVEN** a source file with frontmatter `name: commit-skill`
- **WHEN** AddResource is called without `--name` flag
- **THEN** the resource name is `commit-skill`

#### Scenario: Missing name fails
- **GIVEN** a source file with no name in frontmatter or flag
- **WHEN** AddResource is called
- **THEN** an error is returned indicating name could not be detected

### Requirement: Pre-flight ref validation via core.CanInstallResource

When `library add` is invoked in Mode 1 (explicit files), the command SHALL validate each input's resulting ref via `core.CanInstallResource(name)` before any I/O is performed. If validation fails, the command SHALL return `*core.ValidationError` (mapped to exit 1 by `cmdutil.ExitCodeFor` via the default-error case at `internal/cmdutil/exit.go:77`).

#### Scenario: Valid ref passes pre-flight validation

- **GIVEN** a `--name commit` flag and `--type skill`
- **WHEN** `germinator library add <file> --name commit --type skill` is invoked
- **THEN** `core.CanInstallResource("skill/commit")` SHALL return `nil`
- **AND** the import proceeds (file is copied, library.yaml updated)

#### Scenario: Invalid type fails pre-flight validation

- **GIVEN** a `--name commit` flag and `--type skills` (plural, not a valid type)
- **WHEN** `germinator library add <file> --name commit --type skills` is invoked
- **THEN** `core.CanInstallResource("skills/commit")` SHALL return a non-nil error
- **AND** the error SHALL be `*core.ValidationError`
- **AND** `output.FormatError` SHALL render `Error: ref type must be one of skill, agent, command, memory\n` to stderr
- **AND** no I/O is performed (no file copy, no library.yaml update)

#### Scenario: Empty name fails pre-flight validation

- **GIVEN** `--name ""` (or no `--name` flag and no frontmatter name) and `--type skill`
- **WHEN** `germinator library add <file> --name "" --type skill` is invoked
- **THEN** `core.CanInstallResource("skill/")` SHALL return a non-nil `*core.ValidationError`
- **AND** the import is aborted before I/O

#### Scenario: Malformed ref (no slash) fails pre-flight validation

- **GIVEN** `--name "commit"` and `--type ""` (no type detected)
- **WHEN** the resolved ref would be `"commit"` (no slash)
- **THEN** `core.CanInstallResource("commit")` SHALL return a non-nil `*core.ValidationError`
- **AND** the import is aborted before I/O

### Requirement: Auto-detect description

The system SHALL detect description from flag or frontmatter.

#### Scenario: Detect description from --description flag
- **GIVEN** a source file with frontmatter description `Original description`
- **WHEN** AddResource is called with `--description "New description"`
- **THEN** the resource description is `New description`

#### Scenario: Detect description from frontmatter
- **GIVEN** a source file with frontmatter `description: Git commit best practices`
- **WHEN** AddResource is called without `--description` flag
- **THEN** the resource description is `Git commit best practices`

### Requirement: Handle existing resources

The system SHALL handle resource conflicts according to force flag.

#### Scenario: Error on existing resource without force
- **GIVEN** a library with existing resource `skill/commit`
- **WHEN** AddResource is called for `skill/commit` without `--force`
- **THEN** an error is returned indicating resource already exists

#### Scenario: Replace existing resource with force
- **GIVEN** a library with existing resource `skill/commit`
- **WHEN** AddResource is called with `--force` for `skill/commit`
- **THEN** the file is replaced and library.yaml entry is updated

### Requirement: Support dry-run mode

The system SHALL preview changes without modifying library in dry-run mode. The dry-run preview SHALL be written to the `Stdout io.Writer` field on `AddRequest` (populated by the cmd layer from `opts.IO.Out`). The library package SHALL NOT write to `os.Stdout` directly; the cmd layer is responsible for selecting the writer.

#### Scenario: Dry-run shows what would happen
- **GIVEN** a library and valid source file
- **WHEN** AddResource is called with `--dry-run`
- **THEN** no files are modified
- **AND** the expected action is described via the `Stdout io.Writer` field on `AddRequest` (gated on `opts.Stdout != nil`, not directly to `os.Stdout`)

#### Scenario: Library package does not write to os.Stdout
- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** the dry-run output SHALL be written via the `Stdout io.Writer` field on `AddRequest`

### Requirement: Validate canonical document

The system SHALL validate the canonical document before adding to library.

#### Scenario: Valid document is added
- **GIVEN** a valid canonical skill document
- **WHEN** AddResource is called
- **THEN** the document passes validation and is added to library

#### Scenario: Invalid document is rejected
- **GIVEN** a canonical skill document with missing required fields
- **WHEN** AddResource is called
- **THEN** an error is returned indicating validation failure

### Requirement: Validate library after update

The system SHALL validate library by loading it after update.

#### Scenario: Valid library after update
- **GIVEN** a library and successful resource add
- **WHEN** LoadLibrary is called on the updated library
- **THEN** the library loads successfully with the new resource

#### Scenario: Library validation catches corruption
- **GIVEN** a library update that corrupted library.yaml
- **WHEN** LoadLibrary is called
- **THEN** an error is returned

### Requirement: Discover library path

The system SHALL discover library path via flag, environment, or default.

#### Scenario: Discover library from --library flag
- **GIVEN** `--library=/custom/path` flag
- **WHEN** AddResource is called
- **THEN** the library at `/custom/path` is used

#### Scenario: Discover library from environment
- **GIVEN** `GERMINATOR_LIBRARY=/env/path` env var and no flag
- **WHEN** AddResource is called
- **THEN** the library at `/env/path` is used

#### Scenario: Discover library from default
- **GIVEN** no `--library` flag and no `GERMINATOR_LIBRARY` env
- **WHEN** AddResource is called
- **THEN** `~/.config/germinator/library/` is used

### Requirement: Canonicalize platform documents

The system SHALL convert platform documents to canonical format before storing.

#### Scenario: Canonicalize OpenCode agent
- **GIVEN** an OpenCode format agent document
- **WHEN** AddResource is called with `--platform opencode`
- **THEN** the document is parsed, converted to canonical Agent, and marshaled

#### Scenario: Canonicalize Claude Code skill
- **GIVEN** a Claude Code format skill document
- **WHEN** AddResource is called with `--platform claude-code`
- **THEN** the document is parsed, converted to canonical Skill, and marshaled

#### Scenario: Detect platform from frontmatter
- **GIVEN** a source file with `platform: opencode` in frontmatter
- **WHEN** AddResource is called without `--platform`
- **THEN** platform is detected as `opencode`

#### Scenario: Skip canonicalization for canonical format
- **GIVEN** a source file already in canonical format (has `name:`, `description:`, `tools:` fields)
- **WHEN** AddResource is called
- **THEN** the document is validated but not re-canonicalized

### Requirement: Support orphan discovery mode

The system SHALL support discovering orphaned resource files via --discover flag.

#### Scenario: Discover flag shows orphans without modifying
- **GIVEN** a library with orphan files not in library.yaml
- **WHEN** AddResource is called with `--discover`
- **THEN** orphaned files are reported
- **AND** library.yaml is not modified

#### Scenario: Discover with force registers orphans
- **GIVEN** a library with orphan file `skills/new-skill.md`
- **WHEN** AddResource is called with `--discover --force`
- **THEN** the orphan is registered in library.yaml

#### Scenario: Discover requires explicit flag
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called without `--discover`
- **THEN** no discovery occurs
- **AND** orphans are not reported

#### Scenario: Discover dry-run previews registration
- **GIVEN** a library with orphan files
- **WHEN** AddResource is called with `--discover --dry-run`
- **THEN** no files are modified
- **AND** expected additions are described
