# ci-workflow Specification

## Purpose
Define GitLab CI pipeline with 4 stages (build-ci, setup, validate, distribute), efficient caching, and parallel job execution.

## Requirements
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

---

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
**Then** key SHALL use .cache-key as single file reference
**And** .cache-key SHALL contain SHA256 checksums of all critical files
**And** .cache-key SHALL include checksums for .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum
**And** any change to these files SHALL change .cache-key
**And** cache SHALL be invalidated when .cache-key changes

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
**Then** cache policy SHALL be pull
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

### Requirement: Hash-Based CI Image Tagging

The CI image build process SHALL use content-based hashing to generate unique image tags that capture changes to Dockerfile.ci and .mise/config.toml.

#### Scenario: Calculate content hash for CI image tag
**Given** build-ci-image job is running
**When** hash is calculated
**Then** hash SHALL be SHA256 hash of Dockerfile.ci + .mise/config.toml
**And** hash SHALL be truncated to 12 characters (first 48 bits)
**And** hash SHALL be combined with mise version in format mise-version-hash
**And** tag format SHALL be "2026.1.2-abc123def456"

#### Scenario: Job runs only when Dockerfile.ci or config.toml changes on main
**Given** pipeline is running on main branch
**When** Dockerfile.ci or .mise/config.toml has been modified
**Then** build-ci-image job SHALL run
**And** job SHALL rebuild the CI image with changes
**When** neither Dockerfile.ci nor .mise/config.toml has been modified
**Then** build-ci-image job SHALL be skipped
**And** pipeline SHALL continue without rebuilding image

#### Scenario: Job runs only when Dockerfile.ci or config.toml changes in MR
**Given** pipeline is running on merge request
**When** Dockerfile.ci or .mise/config.toml has been modified
**Then** build-ci-image job SHALL run
**And** job SHALL rebuild the CI image with changes
**When** neither Dockerfile.ci nor .mise/config.toml has been modified
**Then** build-ci-image job SHALL be skipped
**And** pipeline SHALL continue without rebuilding image

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

---

### Requirement: Latest DIND Service Version

The CI pipeline SHALL use the latest stable version of the docker:dind service to ensure API compatibility with the docker CLI in the CI image.

#### Scenario: DIND service uses latest version
**Given** build-ci-image or release job is configured
**When** docker:dind service is defined
**Then** service SHALL use latest stable version (e.g., 29.1.4)
**And** docker CLI image SHALL match DIND service version
**And** API version SHALL be compatible between CLI and DIND

#### Scenario: No API version mismatch errors
**Given** release job is running with docker:dind service
**When** docker login command is executed
**Then** no "client version is too new" errors SHALL occur
**Then** docker commands SHALL succeed
**And** API version SHALL be supported by DIND service

#### Scenario: Build job uses latest DIND
**Given** build-ci-image job is configured
**When** job uses docker:dind service
**Then** service SHALL be latest stable version
**And** docker build commands SHALL succeed
**And** docker push commands SHALL succeed

---

### Requirement: YAML Anchors for Cache Configuration

The CI pipeline SHALL use YAML anchors to eliminate duplicate cache configuration.

#### Scenario: Shared cache configuration defined as anchor
**Given** .gitlab-ci.yml is inspected
**When** cache configuration is defined
**Then** .cache_config anchor SHALL exist at top level
**And** anchor SHALL contain key, files, and paths configuration
**And** anchor SHALL not contain policy (policy is job-specific)

#### Scenario: Setup job uses cache anchor with pull-push policy
**Given** setup job is defined
**When** cache configuration is reviewed
**Then** job SHALL reference .cache_config anchor
**And** policy SHALL be pull-push
**And** cache SHALL be writable by setup job only

#### Scenario: Lint job uses cache anchor with pull policy
**Given** lint job is defined
**When** cache configuration is reviewed
**Then** job SHALL reference .cache_config anchor
**And** policy SHALL be pull
**And** cache SHALL be read-only

#### Scenario: Test job uses cache anchor with pull policy
**Given** test job is defined
**When** cache configuration is reviewed
**Then** job SHALL reference .cache_config anchor
**And** policy SHALL be pull
**And** cache SHALL be read-only

---

### Requirement: Interruptible Jobs for CI Efficiency

Long-running CI jobs SHALL be marked as interruptible to allow cancellation when new pipelines start.

#### Scenario: Lint job is interruptible
**Given** lint job is running
**When** a new pipeline is triggered on the same branch
**Then** lint job SHALL be interruptible
**And** job SHALL be cancelled if configured
**And** CI resources SHALL be freed for new pipeline

#### Scenario: Mirror job is interruptible
**Given** mirror-to-github job is running
**When** a new pipeline is triggered on the same branch
**Then** mirror job SHALL be interruptible
**And** job SHALL be cancelled if configured
**And** CI resources SHALL be freed for new pipeline

#### Scenario: Setup job is NOT interruptible
**Given** setup job is running
**When** a new pipeline is triggered on the same branch
**Then** setup job SHALL NOT be interruptible
**And** job SHALL run to completion
**And** cache writes SHALL not be interrupted

#### Scenario: Release job is NOT interruptible
**Given** release job is running
**When** a new pipeline is triggered
**Then** release job SHALL NOT be interruptible
**And** job SHALL run to completion
**And** release artifacts SHALL be completed

---

### Requirement: Consolidated Pipeline Stages

The CI pipeline SHALL use a minimal number of stages with parallel job execution to reduce pipeline duration.

#### Scenario: Pipeline has exactly 4 stages
**Given** .gitlab-ci.yml is inspected
**When** stages are listed
**Then** pipeline SHALL have exactly 4 stages: build-ci, setup, validate, distribute
**And** stages SHALL be in the following order
**And** no additional stages SHALL exist

#### Scenario: Validate stage runs lint and test in parallel
**Given** pipeline is running in validate stage
**When** jobs are executing
**Then** lint job SHALL be running in validate stage
**And** test job SHALL be running in validate stage
**And** both jobs SHALL execute in parallel
**And** both jobs SHALL depend on setup job
**And** both jobs SHALL NOT depend on each other

#### Scenario: Distribute stage runs release and mirror in parallel
**Given** pipeline is running in distribute stage
**When** jobs are executing
**Then** release job SHALL be running in distribute stage
**And** mirror-to-github job SHALL be running in distribute stage
**And** both jobs SHALL execute in parallel
**And** release job SHALL depend on lint and test jobs
**And** mirror job SHALL run in parallel without dependencies

---

### Requirement: Mirror Job Runs on Main Pushes

The mirror job SHALL run on main branch pushes to mirror code to GitHub, excluding tag pushes.

#### Scenario: Mirror job has no dependencies
**Given** mirror-to-github job is defined
**When** job configuration is reviewed
**Then** job SHALL have no needs dependencies
**And** job SHALL run in distribute stage

#### Scenario: Mirror runs on main pushes, not tags
**Given** pipeline is triggered by main branch push
**When** pipeline reaches distribute stage
**Then** mirror-to-github job SHALL run
**And** code SHALL be mirrored to GitHub

#### Scenario: Mirror job skipped on tag pushes
**Given** pipeline is triggered by tag push
**When** pipeline reaches distribute stage
**Then** mirror-to-github job SHALL be skipped
**And** code SHALL NOT be mirrored again (tags already included in mirror)

#### Scenario: Release job has no tag dependency
**Given** release job is defined
**When** job configuration is reviewed
**Then** job SHALL have needs dependency on lint job
**And** job SHALL have needs dependency on test job
**And** job SHALL NOT have needs dependency on any tag job
**And** release job SHALL wait for both lint and test to succeed

---

### Requirement: Simplified Release Job

The release job SHALL use goreleaser/goreleaser official image with inline validation for simplicity.

#### Scenario: Release job uses official image
**Given** release job is defined
**When** job configuration is reviewed
**Then** job SHALL use goreleaser/goreleaser image
**And** job SHALL have entrypoint set to [""]
**And** job SHALL NOT use Docker service
**And** job SHALL NOT have Docker-related variables (DOCKER_HOST, DOCKER_TLS_CERTDIR)
**And** job SHALL NOT have docker login commands

#### Scenario: Release job enables changelog generation
**Given** release job is defined
**When** job variables are reviewed
**Then** GIT_DEPTH SHALL be 0
**And** GoReleaser SHALL be able to diff tags for changelog
**And** GoReleaser SHALL generate release notes from git history

#### Scenario: Release job validates tag format inline
**Given** release job is running
**When** before_script executes
**Then** job SHALL validate tag matches format vX.Y.Z using regex
**And** invalid tags SHALL cause immediate failure
**And** job SHALL NOT use mise or external validation script
**And** job SHALL NOT validate git state (redundant in CI)
**And** job SHALL NOT validate branch (redundant in CI)

#### Scenario: GoReleaser uses project access token
**Given** .goreleaser.yml is configured
**When** release.gitlab section is reviewed
**Then** GITLAB_TOKEN environment variable SHALL be set to $GITLAB_RELEASE_TOKEN
**And** GoReleaser SHALL use GitLab project access token with api scope for API access

#### Scenario: Release validation task removed from CI
**Given** .mise/tasks/release/validate.sh is inspected
**When** file exists
**Then** file SHALL remain for manual validation
**And** .mise/config.toml SHALL have release:validate task
**And** release validation SHALL be available for pre-tag checks

---

The CI pipeline SHALL validate GoReleaser configuration on merge requests before release.

#### Scenario: GoReleaser dry-run runs on MRs with relevant changes
**Given** MR is created
**And** MR changes .goreleaser.yml
**Or** MR changes go.mod or go.sum
**Or** MR changes cmd/**/* or internal/**/*
**When** pipeline runs
**Then** goreleaser-dry-run job SHALL run
**And** job SHALL validate GoReleaser configuration
**And** job SHALL NOT publish artifacts

#### Scenario: GoReleaser dry-run uses snapshot mode
**Given** goreleaser-dry-run job is running
**When** goreleaser release command executes
**Then** command SHALL use --snapshot flag
**And** command SHALL use --skip-publish flag
**And** command SHALL use --clean flag
**And** no artifacts SHALL be published
**And** no GitLab release SHALL be created

#### Scenario: GoReleaser dry-run uses official image
**Given** goreleaser-dry-run job is defined
**When** job configuration is reviewed
**Then** job SHALL use goreleaser/goreleaser image
**And** job SHALL have entrypoint set to [""]
**And** job SHALL enable full git history with GIT_DEPTH: 0

#### Scenario: GoReleaser dry-run catches configuration errors
**Given** .goreleaser.yml has syntax error
**Or** .goreleaser.yml has invalid configuration
**When** goreleaser-dry-run job runs
**Then** job SHALL fail
**And** job SHALL display GoReleaser error message
**And** MR SHALL NOT be mergeable
**And** developer SHALL fix configuration before merging

#### Scenario: GoReleaser dry-run validates on code changes
**Given** MR includes changes to cmd/**/* or internal/**/*
**When** goreleaser-dry-run job runs
**Then** job SHALL validate GoReleaser configuration
**And** job SHALL ensure builds configuration is valid
**And** job SHALL ensure release configuration is valid
