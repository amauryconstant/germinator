# Capability: Library Batch Add

## Purpose

The Library Batch Add capability enables importing multiple resources to the library in a single command. It processes files and directories, collects results by category (added/skipped/failed), and returns a structured result enabling rich CLI output and scripting.

## Requirements

### Requirement: Batch mode accepts multiple arguments

The system SHALL accept multiple positional arguments when `--batch` flag is set.

#### Scenario: Batch add with multiple files
- **GIVEN** `--batch` flag is set and source files `skill-a.md` and `skill-b.md` exist
- **WHEN** `library add --batch skill-a.md skill-b.md` is called
- **THEN** both files are processed

#### Scenario: Batch add with file and directory
- **GIVEN** `--batch` flag is set, file `agent.md` exists, and directory `./skills/` contains `skill-1.md`
- **WHEN** `library add --batch agent.md ./skills/` is called
- **THEN** `agent.md` and all `.md` files in `./skills/` are processed

#### Scenario: Batch add without --batch flag
- **GIVEN** `--batch` flag is NOT set
- **WHEN** `library add file.md` is called with a single file
- **THEN** the command works as before (single file add)

#### Scenario: Batch add with no arguments
- **GIVEN** `--batch` flag is set but no arguments provided
- **WHEN** `library add --batch` is called
- **THEN** an error is returned requiring at least one source

### Requirement: Directories are scanned recursively for .md files

The system SHALL recursively scan directories to find all `.md` files.

#### Scenario: Batch add with nested directories
- **GIVEN** directory `./skills/` contains `skills/subdir/nested-skill.md`
- **WHEN** `library add --batch ./skills/` is called
- **THEN** `skills/subdir/nested-skill.md` is discovered and processed

#### Scenario: Batch add skips non-markdown files
- **GIVEN** directory `./data/` contains `data/file.md` and `data/readme.txt`
- **WHEN** `library add --batch ./data/` is called
- **THEN** only `data/file.md` is processed

#### Scenario: Batch add with empty directory
- **GIVEN** directory `./empty/` exists but contains no `.md` files
- **WHEN** `library add --batch ./empty/` is called
- **THEN** the directory is skipped with no error

### Requirement: Processing continues on error

The system SHALL continue processing all inputs even if some fail.

#### Scenario: Batch add continues after file error
- **GIVEN** `--batch` flag is set, `valid.md` exists, and `invalid.md` has parse errors
- **WHEN** `library add --batch valid.md invalid.md` is called
- **THEN** `valid.md` is processed and `invalid.md` failure is collected
- **AND** the command returns successfully

#### Scenario: Batch add collects all failures
- **GIVEN** `--batch` flag is set and files `a.md`, `b.md`, `c.md` all have errors
- **WHEN** `library add --batch a.md b.md c.md` is called
- **THEN** all three failures are collected in the result

### Requirement: Exit code is always 0 for batch

The system SHALL return exit code 0 for batch operations regardless of individual failures.

#### Scenario: Batch add with failures returns exit 0
- **GIVEN** `--batch` flag is set and some files fail to process
- **WHEN** `library add --batch file1.md file2.md` is called
- **THEN** exit code is 0

#### Scenario: Batch add with all failures returns exit 0
- **GIVEN** `--batch` flag is set and all files fail to process
- **WHEN** `library add --batch bad1.md bad2.md` is called
- **THEN** exit code is 0

### Requirement: Summary output at end of batch

The system SHALL output a summary after processing all inputs.

#### Scenario: Batch add shows summary
- **GIVEN** `--batch` flag is set and inputs were processed
- **WHEN** `library add --batch file1.md file2.md` is called
- **THEN** output includes "Added N, skipped M, failed K" summary

#### Scenario: Batch add summary with all success
- **GIVEN** `--batch` flag is set and all files added successfully
- **WHEN** `library add --batch skill-a.md skill-b.md` is called
- **THEN** output shows "Added 2, skipped 0, failed 0"

#### Scenario: Batch add summary with mixed results
- **GIVEN** `--batch` flag is set, `skill-a.md` added, `skill-b.md` skipped (conflict), `skill-c.md` failed
- **WHEN** `library add --batch skill-a.md skill-b.md skill-c.md` is called
- **THEN** output shows "Added 1, skipped 1, failed 1"

### Requirement: Skipped resources are distinguished from failures

The system SHALL categorize "skipped" (business logic) separately from "failed" (unexpected errors).

#### Scenario: Skip on existing resource
- **GIVEN** library already contains `skill/commit` and `--batch` is set
- **WHEN** `library add --batch skill-commit.md` is called
- **THEN** the resource is marked as skipped with issue "already_exists"

#### Scenario: Skip on conflict
- **GIVEN** orphan file exists with name matching existing resource
- **WHEN** `library add --batch --discover` is called
- **THEN** the orphan is marked as skipped with issue "conflict"

#### Scenario: Failure on invalid file
- **GIVEN** file `broken.md` contains invalid YAML frontmatter
- **WHEN** `library add --batch broken.md` is called
- **THEN** the file is marked as failed with error message

### Requirement: JSON output via parent --json flag

The system SHALL output structured JSON when `--json` flag is set on parent `library` command.

#### Scenario: Batch add with --json flag
- **GIVEN** `--batch` flag is set and `--json` is set on parent
- **WHEN** `library add --batch file1.md file2.md` is called
- **THEN** output is JSON with `added`, `skipped`, `failed`, `summary` fields

#### Scenario: JSON output structure
- **GIVEN** `--batch` and `--json` flags are set
- **WHEN** `library add --batch file.md` is called
- **THEN** output JSON conforms to BatchAddResult schema:
```json
{
  "added": [{"ref": "skill/commit", "path": "skills/commit.md"}],
  "skipped": [{"source": "skill-existing.md", "issue": "already_exists"}],
  "failed": [{"source": "bad.md", "error": "validation failed"}],
  "summary": {"total": 3, "added": 1, "skipped": 1, "failed": 1}
}
```

### Requirement: Discover integration with batch

The system SHALL support `--discover --batch` to find orphans and add all of them.

#### Scenario: Discover batch finds and adds orphans
- **GIVEN** library at `/tmp/lib` with orphan files not in library.yaml
- **WHEN** `library add --discover --batch` is called
- **THEN** all orphans are processed as if added individually

#### Scenario: Discover batch with force
- **GIVEN** library with orphan `skill/new.md` not in library.yaml
- **WHEN** `library add --discover --batch --force` is called
- **THEN** the orphan is registered in library.yaml

#### Scenario: Discover batch dry-run
- **GIVEN** library with orphans
- **WHEN** `library add --discover --batch --dry-run` is called
- **THEN** no changes are made and orphans are shown as would-be-added

### Requirement: Dry-run in batch mode

The system SHALL preview changes without modifying library when `--dry-run` is set in batch mode.

#### Scenario: Batch dry-run shows additions
- **GIVEN** `--batch` and `--dry-run` flags are set with valid sources
- **WHEN** `library add --batch --dry-run skill-a.md skill-b.md` is called
- **THEN** no files are modified
- **AND** summary shows what would be added

#### Scenario: Batch dry-run shows failures
- **GIVEN** `--batch` and `--dry-run` flags are set with one invalid source
- **WHEN** `library add --batch --dry-run valid.md invalid.md` is called
- **AND** no files are modified
- **AND** the invalid file is shown as would-fail

### Requirement: Force flag applies to all in batch

The system SHALL apply `--force` to all files in batch mode.

#### Scenario: Batch with force overwrites existing
- **GIVEN** library already contains `skill/commit` and `--batch --force` is set
- **WHEN** `library add --batch --force skill-commit.md` is called
- **THEN** the existing resource is replaced

#### Scenario: Batch without force keeps existing
- **GIVEN** library already contains `skill/commit` and `--batch` is set without `--force`
- **WHEN** `library add --batch skill-commit.md` is called
- **THEN** the resource is skipped with issue "already_exists"

### Requirement: Resource type and platform detection in batch

The system SHALL detect resource type and platform the same way as single-file add.

#### Scenario: Batch detects type from frontmatter
- **GIVEN** `agent.md` contains `type: agent` in frontmatter
- **WHEN** `library add --batch agent.md` is called
- **THEN** type is detected as `agent`

#### Scenario: Batch detects platform from frontmatter
- **GIVEN** `skill.md` contains `platform: opencode` in frontmatter
- **WHEN** `library add --batch skill.md` is called
- **THEN** platform is detected as `opencode`
- **AND** document is canonicalized before adding

#### Scenario: Batch with explicit type override
- **GIVEN** `unknown.md` has no type indication
- **WHEN** `library add --batch --type skill unknown.md` is called
- **THEN** type is `skill` regardless of detection
