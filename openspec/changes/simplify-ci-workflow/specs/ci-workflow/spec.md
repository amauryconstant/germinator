# ci-workflow Spec Delta

**Note**: This is a new capability spec. All requirements below are **ADDED** (not modifications to existing specs). Release validation changes are tracked as delta changes in `specs/release-management/spec.md` within this proposal.

## ADDED Requirements

### Requirement: CI Pipeline Optimization for Documentation-Only Changes

The CI pipeline SHALL skip expensive jobs when only documentation changes occur in `openspec/` directory.

#### Scenario: Full pipeline runs on code changes
**Given** commit includes code changes (not openspec/)
**When** pipeline runs
**Then** all jobs SHALL run normally
**And** expensive jobs (lint, test, release) SHALL NOT be skipped

#### Scenario: Expensive jobs skipped on doc-only changes
**Given** commit only includes openspec/ directory changes
**When** pipeline runs
**Then** openspec-validation job SHALL detect openspec-only changes
**And** lint, test, release jobs SHALL be skipped
**And** setup job SHALL still run
**And** pipeline SHALL complete quickly

### Requirement: CI Variable Validation

The CI pipeline SHALL validate required variables using GitLab CI rules before jobs that depend on them.

#### Scenario: GitHub mirror validates variables via rules
**Given** mirror-to-github job is triggered
**When** CI_COMMIT_BRANCH is main
**And** GITHUB_ACCESS_TOKEN is set
**And** GITHUB_REPO_URL is set
**Then** job SHALL run
**And** job SHALL attempt git push to GitHub

#### Scenario: GitHub mirror skipped when variables missing
**Given** mirror-to-github job is triggered
**When** GITHUB_ACCESS_TOKEN is not set
**Or** GITHUB_REPO_URL is not set
**Then** job SHALL be skipped by GitLab CI rules
**And** job SHALL NOT appear in pipeline as failed
**And** job SHALL appear as skipped in pipeline UI
**And** pipeline SHALL complete successfully without errors

#### Scenario: No manual variable checks in script
**Given** mirror-to-github job runs
**When** job script executes
**Then** script SHALL NOT check for GITHUB_ACCESS_TOKEN
**And** script SHALL NOT check for GITHUB_REPO_URL
**And** validation SHALL be handled by GitLab CI rules only

---

### Requirement: Consistent Base Images

All CI jobs SHALL use the same base image for consistent tooling and environment.

#### Scenario: All jobs use CI image
**Given** .gitlab-ci.yml is inspected
**When** job configurations are reviewed
**Then** all jobs SHALL use `registry.gitlab.com/amoconst/germinator/ci:latest`
**And** no jobs SHALL use alpine:latest or other images
**And** tool versions SHALL be consistent across all jobs

#### Scenario: Mirror job uses CI image
**Given** mirror-to-github job runs
**When** job base image is inspected
**Then** image SHALL be `registry.gitlab.com/amoconst/germinator/ci:latest`
**And** git SHALL be pre-installed in image
**And** job SHALL NOT install git via apk
**And** job SHALL NOT include apk add in script

#### Scenario: Tools available in CI image
**Given** job uses CI image
**When** job executes
**Then** git SHALL be available
**And** curl SHALL be available
**And** all tools from mise config SHALL be available
**And** job SHALL NOT need to install tools at runtime

---

### Requirement: Comprehensive Cache Keys

The cache configuration SHALL invalidate when any configuration or dependency file changes.

#### Scenario: Cache key includes all configuration files
**Given** .gitlab-ci.yml cache configuration is defined
**When** cache key is computed
**Then** key SHALL include .gitlab-ci.yml
**And** key SHALL include Dockerfile.ci
**And** key SHALL include .mise/config.toml
**And** key SHALL include go.mod
**And** key SHALL include go.sum
**And** any change to these files SHALL invalidate cache

#### Scenario: Cache invalidates on tool version changes
**Given** .mise/config.toml is modified
**When** tool version is bumped (e.g., golangci-lint)
**Then** cache key SHALL change
**And** old cache SHALL not be used
**And** new tools SHALL be installed by mise
**And** new cache SHALL be populated

#### Scenario: Cache invalidates on pipeline config changes
**Given** .gitlab-ci.yml is modified
**When** pipeline runs after change
**Then** cache key SHALL change
**And** old cache SHALL not be used
**And** new cache SHALL be populated

---

### Requirement: Safe Cache Access Policies

Cache policies SHALL prevent corruption from concurrent writes via serialization.

#### Scenario: Setup job writes cache exclusively
**Given** setup job runs
**When** job completes successfully
**Then** cache policy SHALL be pull-push
**And** job SHALL update Go module cache
**And** other jobs SHALL not write to cache

#### Scenario: Lint and test jobs read only
**Given** lint or test job runs
**When** job accesses cache
**Then** cache policy SHALL be pull only
**And** job SHALL NOT write to cache
**And** no concurrent writes SHALL occur

#### Scenario: Serialized cache writes prevent conflicts
**Given** multiple pipelines run simultaneously on main branch
**When** setup job from one pipeline writes cache
**Then** resource_group: cache_updates SHALL serialize writes
**And** only one pipeline SHALL write cache at a time
**And** other pipelines SHALL read consistent cache
**And** cache SHALL not be corrupted

#### Scenario: No cache corruption
**Given** multiple pipelines run concurrently
**When** jobs access cache
**Then** cache SHALL not be corrupted
**And** jobs SHALL not fail due to cache conflicts
**And** cache SHALL remain consistent

---

### Requirement: Standardized Artifact Lifetime

Artifacts SHALL have consistent lifetime across all proposals to support multi-stage pipelines.

#### Scenario: Artifacts last 24 hours
**Given** pipeline creates artifacts
**When** artifact expiration is configured
**Then** expiration SHALL be 24 hours
**And** artifacts SHALL be available to all downstream stages
**And** artifacts SHALL not expire mid-pipeline

#### Scenario: Go module cache available to all stages
**Given** setup stage completes
**When** lint and test stages run
**Then** Go module cache SHALL be available
**And** dependencies SHALL not be re-downloaded
**And** pipeline SHALL complete successfully

---

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

The CI pipeline SHALL automatically create Git tags when the version file changes, eliminating manual tagging.

#### Scenario: Tag stage creates version tag
**Given** commit includes changes to internal/version/version.go
**When** pipeline runs on main branch
**And** test stage completes successfully
**Then** tag stage SHALL run
**And** tag stage SHALL extract version from internal/version/version.go
**And** tag stage SHALL create Git tag with format v<VERSION>
**And** tag stage SHALL push tag to origin
**And** release stage SHALL trigger with new tag

#### Scenario: Tag stage idempotent behavior
**Given** tag vX.Y.Z already exists
**When** pipeline runs again with same version.go
**Then** tag stage SHALL detect existing tag
**And** tag stage SHALL skip tag creation
**And** tag stage SHALL report "Tag already exists, skipping"
**And** pipeline SHALL continue normally

#### Scenario: Tag stage skips when version unchanged
**Given** commit does NOT change internal/version/version.go
**When** pipeline runs
**Then** tag stage SHALL be skipped
**And** no tag SHALL be created
**And** other stages SHALL run normally

#### Scenario: Tag format validation
**Given** version in internal/version/version.go is "0.3.0"
**When** tag is created
**Then** tag SHALL be "v0.3.0"
**And** tag SHALL include 'v' prefix
**And** tag SHALL match semantic version format

#### Scenario: Git configuration for tagging
**Given** tag stage is running
**When** git config is set
**Then** git user.email SHALL be set from $GITLAB_USER_EMAIL
**And** git user.name SHALL be set from $GITLAB_USER_NAME
**And** git remote SHALL be configured for push with CI_JOB_TOKEN
**And** tag SHALL be pushed successfully

#### Scenario: Release stage integration
**Given** tag stage creates tag v0.3.0
**When** tag push completes
**Then** release stage SHALL trigger automatically
**And** release stage SHALL use tag vX.Y.Z for version
**And** release:validate SHALL find the tag
**And** GoReleaser SHALL create release artifacts

---

### Requirement: Manual Tagging Workflow Removal

The release documentation SHALL remove manual tagging instructions, as tagging is now automatic.

#### Scenario: Documentation reflects automatic tagging
**Given** developer reads AGENTS.md release workflow
**When** documentation is reviewed
**Then** manual `git tag` and `git push origin` steps SHALL NOT be documented
**And** automatic tag creation SHALL be documented
**And** version bump workflow SHALL be clearly explained
**And** tag stage SHALL be included in pipeline stages

#### Scenario: Release workflow simplification
**Given** developer wants to create a release
**When** developer bumps version using mise tasks
**And** commits and pushes to main
**Then** CI SHALL automatically create tag
**And** CI SHALL automatically create release
**And** no manual git commands SHALL be required

---

### Requirement: Hash-Based CI Image Tagging

The CI image build process SHALL use content-based hashing to generate unique image tags that capture changes to Dockerfile.ci and .mise/config.toml.

#### Scenario: Calculate content hash for CI image tag
**Given** build-ci-image job is running
**When** hash is calculated
**Then** hash SHALL be SHA256 hash of Dockerfile.ci + .mise/config.toml
**And** hash SHALL be truncated to 12 characters (first 48 bits)
**And** hash SHALL be combined with mise version in format mise-version-hash
**And** tag format SHALL be "2026.1.2-abc123def456"

#### Scenario: CI image rebuilds on Dockerfile.ci changes
**Given** Dockerfile.ci has been modified
**When** build-ci-image job runs
**Then** hash SHALL be different from previous build
**Then** image tag SHALL be different (e.g., 2026.1.2-xyz789abc123)
**And** CI image SHALL be rebuilt with new Dockerfile.ci
**And** image SHALL be pushed with new tag
**And** latest tag SHALL be updated

#### Scenario: CI image rebuilds on config.toml changes
**Given** .mise/config.toml has been modified
**When** build-ci-image job runs
**Then** hash SHALL be different from previous build
**Then** image tag SHALL be different
**And** CI image SHALL be rebuilt with new config
**And** image SHALL be pushed with new tag
**And** latest tag SHALL be updated

#### Scenario: CI image build skipped on no changes
**Given** Dockerfile.ci and .mise/config.toml are unchanged
**When** build-ci-image job runs
**Then** hash SHALL be identical to previous build
**Then** image tag SHALL already exist in registry
**And** build process SHALL be skipped
**And** "CI image already exists, skipping build" message SHALL be displayed
**And** no new image SHALL be built

#### Scenario: Hash calculation is deterministic
**Given** build-ci-image job runs twice with same files
**When** hashes are calculated
**Then** both hashes SHALL be identical
**And** hash order SHALL be consistent (Dockerfile.ci before config.toml)
**And** same files SHALL always generate same hash

#### Scenario: CI image tag with new mise version
**Given** mise version has updated to new release
**When** build-ci-image job runs
**Then** mise version in tag SHALL be new version (e.g., 2026.2.0)
**And** hash SHALL be calculated based on current files
**Then** image tag SHALL be "2026.2.0-abc123def456"
**And** CI image SHALL be built with new mise version

#### Scenario: Docker CLI available in CI image
**Given** release job is running
**When** docker login command is executed
**Then** docker command SHALL be available
**And** docker login SHALL succeed with CI_JOB_TOKEN
**And** release job SHALL authenticate with registry.gitlab.com

---

### Requirement: Docker CLI in CI Image

The CI image SHALL include the docker-cli package to enable docker commands in release jobs.

#### Scenario: Docker CLI installed in CI image
**Given** CI image is built from Dockerfile.ci
**When** image is inspected
**Then** docker-cli package SHALL be installed via apk
**And** docker command SHALL be available in PATH
**And** docker --version SHALL execute successfully
