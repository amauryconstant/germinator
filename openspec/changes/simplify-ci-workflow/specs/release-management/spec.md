# release-management Spec Delta

**Note**: These are **ADDED** requirements (not modifications). The main `openspec/specs/release-management/spec.md` will be updated during archive process.

## ADDED Requirements

### Requirement: Automated Release Builds

The system SHALL build cross-platform release artifacts automatically when a Git tag is pushed, tag matches version in code, and validation passes.

#### Scenario: Create GitLab Release with validation
**Given** GoReleaser builds all artifacts
**And** Git tag matches version in internal/version/version.go
**And** mise run release:validate passes in before_script
**When** release job completes
**Then** it SHALL create a GitLab release with tag name as release name
**And** it SHALL attach all 5 binary archives
**And** it SHALL attach checksums file
**And** it SHALL attach SBOM file
**And** it SHALL include auto-generated release notes

#### Scenario: Tag validation in release job
**Given** release job is triggered by Git tag
**When** release job starts (before_script)
**Then** it SHALL run mise run release:validate
**And** it SHALL validate tag format
**And** it SHALL validate tag version matches internal/version/version.go
**And** it SHALL fail immediately if versions don't match
**And** error message SHALL indicate mismatch between tag and version.go

---

## ADDED Requirements

### Requirement: Git State Validation

The system SHALL validate git working directory state before allowing release.

#### Scenario: Git state must be clean
**Given** a developer attempts to create a release
**When** validation runs
**And** working directory has uncommitted changes
**Then** validation SHALL fail
**And** error SHALL list uncommitted files
**And** release SHALL not proceed

#### Scenario: Working directory is clean
**Given** a developer attempts to create a release
**When** validation runs
**And** working directory has no uncommitted changes
**Then** validation SHALL pass for git state check
**And** release SHALL proceed if other validations pass

---

### Requirement: Branch Validation

The system SHALL validate that releases only occur from main branch.

#### Scenario: Must be on main branch
**Given** a developer attempts to create a release
**When** validation runs
**And** current branch is not main
**Then** validation SHALL fail
**And** error SHALL indicate main branch is required
**And** release SHALL not proceed

#### Scenario: On main branch
**Given** a developer attempts to create a release
**When** validation runs
**And** current branch is main
**Then** validation SHALL pass for branch check
**And** release SHALL proceed if other validations pass

---

### Requirement: Git Tag Validation

The system SHALL validate Git tags before attempting release builds.

#### Scenario: Tag matches version.go
**Given** tag is properly formatted with 'v' prefix
**When** tag version is compared to code
**Then** tag version (without 'v') SHALL equal Version in internal/version/version.go
**And** mismatch SHALL cause immediate failure
**And** error message SHALL show both versions

#### Scenario: Validation handles 'v' prefix
**Given** Git tag is created with format vX.Y.Z
**When** version is extracted for comparison
**Then** validation SHALL strip 'v' prefix from tag
**And** comparison SHALL use semantic version only
**And** version.go Version SHALL be compared without 'v' prefix

#### Scenario: Validation runs in CI
**Given** release job runs
**When** job starts (before_script)
**Then** validation SHALL occur before GoReleaser
**And** validation SHALL include git state check
**And** validation SHALL include branch check
**And** validation SHALL include tag match check
**And** validation failure SHALL stop job
**And** GoReleaser SHALL not run on invalid tags or states

---

### Requirement: Release Validation Task

The system SHALL provide a single consolidated validation task for release operations (manual, not automatic).

#### Scenario: release:validate task checks all conditions
**Given** developer wants to check release readiness
**When** developer runs mise run release:validate
**Then** task SHALL check git state is clean
**And** task SHALL check current branch is main
**And** task SHALL validate Git tag format
**And** task SHALL validate tag matches version.go
**And** task SHALL validate .goreleaser.yml
**And** task SHALL report all issues found with clear error messages

#### Scenario: release:validate with uncommitted changes
**Given** developer runs mise run release:validate
**When** uncommitted changes exist
**Then** task SHALL fail
**And** error SHALL list uncommitted files
**And** task SHALL suggest committing changes first

#### Scenario: release:validate on wrong branch
**Given** developer runs mise run release:validate
**When** current branch is not main
**Then** task SHALL fail
**And** error SHALL indicate main branch is required
**And** task SHALL suggest checking out main branch

#### Scenario: release:validate with tag mismatch
**Given** developer runs mise run release:validate
**When** tag version doesn't match version.go
**Then** task SHALL fail
**And** error SHALL show both versions
**And** task SHALL suggest creating correct tag

#### Scenario: No automatic tag creation
**Given** developer runs release:validate task
**When** task executes
**Then** task SHALL NOT automatically create Git tags
**And** task SHALL NOT automatically push tags
**And** developer SHALL manually create and push tags
**And** AGENTS.md SHALL reinforce manual tag creation

#### Scenario: Deprecated release:check task removed
**Given** .mise/config.toml is inspected
**When** tasks are reviewed
**Then** release:check task SHALL NOT exist
**And** all documentation SHALL reference release:validate

---

### Requirement: Automatic Tag Creation

The system SHALL automatically create Git tags when the version file changes, eliminating manual tagging workflow.

#### Scenario: Tag stage triggers on version change
**Given** internal/version/version.go is modified
**When** change is pushed to main branch
**And** test stage completes successfully
**Then** tag stage SHALL run automatically
**And** tag SHALL be created with format v<VERSION>
**And** tag SHALL be pushed to repository

#### Scenario: Manual tagging workflow removed
**Given** developer reads release documentation
**When** release workflow instructions are reviewed
**Then** manual git tag commands SHALL NOT be documented
**And** automatic tag creation SHALL be clearly documented
**And** developer SHALL only need to run `mise run version:*` tasks
