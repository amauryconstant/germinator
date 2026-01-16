# release-management Specification

## Purpose
TBD - created by archiving change implement-release-management. Update Purpose after archive.
## Requirements
### Requirement: GoReleaser Configuration

The project SHALL have a GoReleaser configuration file for automated release management.

#### Scenario: GoReleaser configuration exists
**Given** release management is implemented
**When** a developer inspects the project root
**Then** `.goreleaser.yml` SHALL exist
**And** it SHALL be valid YAML syntax
**And** it SHALL configure cross-platform builds

#### Scenario: Cross-platform builds configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL specify 5 build targets:
**And** it SHALL include linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
**And** it SHALL use `CGO_ENABLED=0` for static linking

#### Scenario: Version injection configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure ldflags for version injection
**And** it SHALL inject `{{.Version}}` into `gitlab.com/amoconst/germinator/internal/version.Version`
**And** it SHALL inject `{{.Commit}}` into `gitlab.com/amoconst/germinator/internal/version.Commit`
**And** it SHALL inject `{{.Date}}` into `gitlab.com/amoconst/germinator/internal/version.Date`

#### Scenario: Archive generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL generate `.tar.gz` archives for Linux/macOS
**And** it SHALL generate `.zip` archives for Windows
**And** it SHALL include LICENSE and README.md in archives

#### Scenario: Checksum generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure checksum generation
**And** it SHALL use SHA256 algorithm
**And** it SHALL name checksum file as `checksums.txt`

#### Scenario: SBOM generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure SBOM generation using Syft
**And** it SHALL use SPDX format
**And** it SHALL name SBOM file as `germinator_{{.Version}}_sbom.spdx.json`

#### Scenario: Changelog generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure auto-generated release notes
**And** it SHALL filter out docs, test, ci, chore, build commits
**And** it SHALL filter out merge commits

---

### Requirement: Automated Release Builds

The system SHALL build cross-platform release artifacts automatically when a Git tag is pushed.

#### Scenario: Build all platform binaries
**Given** a Git tag is pushed (e.g., v0.3.0)
**When** GitLab CI release job runs
**Then** GoReleaser SHALL build 5 binaries:
**And** it SHALL build `germinator_0.3.0_linux_amd64`
**And** it SHALL build `germinator_0.3.0_linux_arm64`
**And** it SHALL build `germinator_0.3.0_darwin_amd64`
**And** it SHALL build `germinator_0.3.0_darwin_arm64`
**And** it SHALL build `germinator_0.3.0_windows_amd64.exe`

#### Scenario: Generate archives
**Given** GoReleaser builds binaries
**When** archives are generated
**Then** it SHALL create `.tar.gz` archives for Linux/macOS platforms
**And** it SHALL create `.zip` archive for Windows platform
**And** archives SHALL contain binary, LICENSE, README.md

#### Scenario: Generate checksums
**Given** GoReleaser creates release artifacts
**When** checksums are generated
**Then** it SHALL create `checksums.txt`
**And** it SHALL contain SHA256 hashes for all 5 archives
**And** checksums.txt SHALL include SBOM file

#### Scenario: Generate SBOMs
**Given** GoReleaser creates release artifacts
**When** SBOMs are generated
**Then** it SHALL create `germinator_0.3.0_sbom.spdx.json`
**And** it SHALL list all Go dependencies
**And** it SHALL include metadata (version, commit, date)

#### Scenario: Create GitLab Release
**Given** GoReleaser builds all artifacts
**When** release job completes
**Then** it SHALL create a GitLab release with tag name as release name
**And** it SHALL attach all 5 binary archives
**And** it SHALL attach checksums file
**And** it SHALL attach SBOM file
**And** it SHALL include auto-generated release notes

#### Scenario: Release notes from commits
**Given** GitLab release is created
**When** release notes are generated
**Then** it SHALL show one commit message per line
**And** it SHALL exclude commits starting with docs:, test:, ci:, chore:, build:
**And** it SHALL exclude merge commits

---

### Requirement: Local Development Builds

The system SHALL support local development builds with version information.

#### Scenario: Build snapshot locally
**Given** a developer runs `mise run build`
**When** simple build runs
**Then** it SHALL create binary in `bin/` directory
**And** version SHALL be from `internal/version/version.go`

#### Scenario: Quick local build without GoReleaser
**Given** a developer runs `mise run build:local`
**When** quick build runs
**Then** it SHALL build single binary in `bin/germinator`
**And** version SHALL be derived from git describe
**And** commit SHA SHALL be from current HEAD
**And** date SHALL be current date

---

### Requirement: Installation Support

The project SHALL provide one-line installation script and comprehensive installation guide.

#### Scenario: Install script for end users
**Given** installation is documented
**When** a user inspects project root
**Then** `install.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Install script detects platform
**Given** a user runs install script
**When** script executes
**Then** it SHALL detect OS (Linux/macOS/Windows)
**And** it SHALL detect architecture (amd64/arm64)
**And** it SHALL select appropriate binary

#### Scenario: Install script downloads and installs
**Given** a user runs install script
**When** script executes
**Then** it SHALL download binary from GitLab releases
**And** it SHALL extract archive
**And** it SHALL install to `/usr/local/bin/` or `~/bin/`
**And** it SHALL display installed version

#### Scenario: Local install task for developers
**Given** installation is documented
**When** a developer inspects `.mise/tasks/` directory
**Then** `.mise/tasks/install-local.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Local install task
**Given** a developer runs `mise run install:local`
**When** task executes
**Then** it SHALL build germinator locally
**And** it SHALL install to `/usr/local/bin/` or `~/.local/bin/`
**And** it SHALL display installation verification message

#### Scenario: Installation guide exists
**Given** installation is documented
**When** a user reads documentation
**Then** `INSTALL.md` SHALL exist
**And** it SHALL document quick install (curl | bash)
**And** it SHALL document manual download for all platforms
**And** it SHALL document checksum verification
**And** it SHALL document GPG signature verification (optional)

---

### Requirement: Reproducible Builds

The system SHALL support reproducible builds via GoReleaser configuration.

#### Scenario: Build reproducibility configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL use `-trimpath` flags
**And** it SHALL set `mod_timestamp: "{{ .CommitTimestamp }}"`
**And** it SHALL set `gcflags: all=-trimpath={{.Env.GOPATH}}`

#### Scenario: Same commit produces same binary
**Given** same GoReleaser config and same commit
**When** binary is built twice
**Then** binaries SHALL have identical checksums
**And** build date SHALL use commit date, not build date

---

### Requirement: Dry-Run Support

The system SHALL support testing release configuration without creating actual releases.

#### Scenario: Validate GoReleaser config
**Given** a developer runs `mise run release:check`
**When** configuration is validated
**Then** GoReleaser SHALL check configuration syntax
**And** it SHALL report any errors or deprecated properties
**And** it SHALL exit with 0 if valid (works with uncommitted changes)

#### Scenario: Dry-run release
**Given** a developer has a clean git state
**And** a developer has a git tag checked out (e.g., v0.3.0)
**When** developer runs `mise run release:dry-run`
**Then** GoReleaser SHALL build all artifacts
**And** it SHALL skip publishing to GitLab
**And** it SHALL create artifacts locally in `dist/` directory
**And** it SHALL use version from git tag
**And** it SHALL display what would be released

#### Scenario: Windows binary testing is out of scope
**Given** GoReleaser generates Windows binary
**When** Windows binary is created
**Then** cross-compilation SHALL work (GOOS=windows GOARCH=amd64)
**And** Windows testing is out of scope for initial implementation
**And** documentation SHALL note that Windows testing is planned for future iteration

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

### Requirement: Automatic Tag Creation with Pipeline Triggering

The system SHALL automatically create Git tags when version file changes, eliminating manual tagging workflow, and SHALL trigger a new pipeline to ensure release stage executes.

#### Scenario: Tag stage creates version tag and triggers pipeline
**Given** internal/version/version.go is modified
**When** change is pushed to main branch
**And** test stage completes successfully
**Then** tag stage SHALL run automatically
**And** tag SHALL be created with format v<VERSION>
**And** tag SHALL be pushed to repository
**And** tag stage SHALL trigger a new pipeline using GitLab API
**And** tag stage SHALL use $CI_JOB_TOKEN for API authentication
**And** tag stage SHALL use /trigger/pipeline endpoint
**And** triggered pipeline SHALL set $CI_COMMIT_TAG
**And** triggered pipeline SHALL set $CI_PIPELINE_SOURCE to "trigger"
**And** release stage SHALL execute automatically with new tag

#### Scenario: Tag stage idempotent behavior with pipeline trigger
**Given** tag vX.Y.Z already exists
**When** pipeline runs again with same version.go
**Then** tag stage SHALL detect existing tag
**And** tag stage SHALL skip tag creation
**And** tag stage SHALL report "Tag already exists — skipping tag creation"
**And** tag stage SHALL STILL trigger pipeline for existing tag (in case tag was deleted and recreated)
**And** tag stage SHALL use $CI_JOB_TOKEN for API authentication
**And** tag stage SHALL use /trigger/pipeline endpoint
**And** triggered pipeline SHALL set $CI_PIPELINE_SOURCE to "trigger"
**And** pipeline SHALL continue normally

#### Scenario: Manual tagging workflow removed
**Given** developer reads release documentation
**When** release workflow instructions are reviewed
**Then** manual git tag commands SHALL NOT be documented
**And** automatic tag creation SHALL be clearly documented
**And** developer SHALL only need to run `mise run version:*` tasks

---

### Requirement: Git State Validation in CI Context

The release:validate task SHALL handle CI-specific conditions when running in GitLab CI environment.

#### Scenario: Git state validation ignores untracked files
**Given** release:validate task is running in CI
**When** git state is checked
**Then** validation SHALL use `git status --porcelain --untracked-files=no`
**And** untracked files (e.g., .cache/) SHALL be ignored
**And** only tracked changes SHALL be detected

#### Scenario: Branch validation accepts detached HEAD in CI
**Given** release job is triggered by tag
**When** release:validate checks branch
**Then** validation SHALL accept detached HEAD state
**And** validation SHALL use git describe to find tag when CI_COMMIT_TAG is not set
**And** validation SHALL NOT fail when on detached HEAD with valid tag
**And** validation SHALL display "Note: Running on detached HEAD (checking for tag in next step)"

#### Scenario: Tag detection fallback when CI_COMMIT_TAG not set
**Given** release:validate is running in CI
**When** CI_COMMIT_TAG environment variable is not set
**Then** validation SHALL use git describe --tags --exact-match to find tag
**And** validation SHALL set GIT_TAG variable from git describe
**And** validation SHALL continue with tag detection process

#### Scenario: Final validation requires tag when on detached HEAD
**Given** release:validate checks branch and finds HEAD
**When** validation reaches final check
**And** no tag is found
**Then** validation SHALL fail
**And** validation SHALL display "On detached HEAD but no tag found" error
**And** validation SHALL require tag for release on detached HEAD

