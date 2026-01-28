# ci-image Specification

## Purpose
TBD - created by archiving change implement-release-management. Update Purpose after archive.
## Requirements
### Requirement: CI Dockerfile Exists

The project SHALL have a custom Dockerfile for CI image.

#### Scenario: Dockerfile exists
**Given** CI image is set up
**When** a developer inspects project root
**Then** `Dockerfile.ci` SHALL exist
**And** it SHALL be a valid Dockerfile
**And** it SHALL be based on `golang:1.25.5-slim`

---

### Requirement: mise Pre-Installed

The CI image SHALL have mise pre-installed for tool management.

#### Scenario: mise installed in image
**Given** `Dockerfile.ci` is inspected
**When** installation steps are reviewed
**Then** mise SHALL be installed via curl from mise.run
**And** mise SHALL be available in PATH
**And** mise version SHALL be latest stable

#### Scenario: mise configured for CI
**Given** CI image is built
**When** environment variables are inspected
**Then** `MISE_DATA_DIR` SHALL be set to `/mise`
**And** `MISE_CACHE_DIR` SHALL be set to `/mise/cache`
**And** `MISE_CONFIG_DIR` SHALL be set to `/mise`
**And** `PATH` SHALL include `/mise/shims:$PATH`

---

### Requirement: Development Tools Pre-Installed

The CI image SHALL have GoReleaser and golangci-lint pre-installed via mise.

#### Scenario: GoReleaser installed in image
**Given** `Dockerfile.ci` is inspected
**When** installation steps are reviewed
**Then** `.mise/config.toml` SHALL be copied into image
**And** mise install SHALL run for all tools
**And** GoReleaser SHALL be available in PATH
**And** `goreleaser --version` SHALL work

#### Scenario: golangci-lint installed in image
**Given** `Dockerfile.ci` is inspected
**When** installation steps are reviewed
**Then** `.mise/config.toml` SHALL be copied into image
**And** mise install SHALL run for all tools
**And** golangci-lint SHALL be available in PATH
**And** `golangci-lint version` SHALL work

#### Scenario: All tools verified
**Given** CI image is built
**When** verification steps run
**Then** `mise ls` SHALL list installed tools
**And** `go version` SHALL output Go 1.25.5
**And** `goreleaser --version` SHALL succeed
**And** `golangci-lint version` SHALL succeed

---

### Requirement: Multi-Stage Build Optimization

The Dockerfile SHALL use multi-stage builds for layer caching.

#### Scenario: Base stage with mise
**Given** `Dockerfile.ci` is inspected
**When** stages are reviewed
**Then** `base` stage SHALL install prerequisites (curl, git, ca-certificates)
**And** `base` stage SHALL install mise
**And** `base` stage SHALL configure mise environment

#### Scenario: Tools stage with cached layer
**Given** `Dockerfile.ci` is inspected
**When** stages are reviewed
**Then** `tools` stage SHALL copy `.mise/config.toml`
**And** `tools` stage SHALL run `mise install --yes`
**And** this layer SHALL be cacheable (tools only change when config changes)

#### Scenario: Final stage with verification
**Given** `Dockerfile.ci` is inspected
**When** stages are reviewed
**Then** `final` stage SHALL copy from `tools` stage
**And** `final` stage SHALL verify all tools work
**And** `final` stage SHALL set workspace directory

---

### Requirement: Image Hosted on GitLab Registry

The CI image SHALL be hosted on GitLab Container Registry.

#### Scenario: Image accessible from GitLab CI
**Given** CI image is pushed
**When** `.gitlab-ci.yml` uses the image
**Then** it SHALL pull from `registry.gitlab.com/amoconst/germinator/ci:latest`
**And** it SHALL be accessible to GitLab CI jobs
**And** it SHALL have proper permissions

#### Scenario: Image versioning
**Given** CI image is pushed
**When** multiple versions exist
**Then** `:latest` tag SHALL always point to newest version
**And** versioned tags (e.g., `:v2026.1.1`) SHALL be immutable
**And** both tags SHALL be available

---

### Requirement: Image Validation

The CI image SHALL be validated before pushing to registry.

#### Scenario: Image builds successfully
**Given** a developer runs `docker build -t test -f Dockerfile.ci .`
**When** build completes
**Then** image SHALL be created
**And** build logs SHALL show no errors

#### Scenario: Tools work in image
**Given** CI image is built
**When** developer tests image locally
**Then** `docker run --rm test mise run check` SHALL succeed
**And** `docker run --rm test go version` SHALL succeed
**And** `docker run --rm test goreleaser --version` SHALL succeed
**And** `docker run --rm test golangci-lint version` SHALL succeed

### Requirement: Alpine-Based CI Image

The CI image SHALL use Alpine Linux base for reduced size while maintaining tool functionality.

#### Scenario: Image uses Alpine 3.23
**Given** CI image is built
**When** base image is inspected
**Then** image SHALL use golang:1.25.5-alpine3.23 as base
**And** final image size SHALL be approximately 215MB
**And** size reduction SHALL be at least 70% compared to Debian 807MB

#### Scenario: All tools remain functional on Alpine
**Given** Alpine-based CI image is built
**When** tool functionality is tested
**Then** mise SHALL install and run correctly
**And** golangci-lint SHALL install and run correctly
**And** GoReleaser SHALL install and run correctly
**And** Go 1.25.5 SHALL be available
**And** static binary compilation SHALL work

#### Scenario: Fast image startup
**Given** CI job starts
**When** image is pulled from registry
**Then** pull time SHALL be less than current Debian baseline by 30%

---

### Requirement: Reliable CI Image Build Process

The CI image build process SHALL be reliable and handle concurrent builds safely.

#### Scenario: Check image existence without full pull
**Given** build-ci-image job runs
**When** checking if image exists
**Then** docker manifest inspect SHALL be used
**And** image SHALL NOT be fully pulled just for existence check
**And** check SHALL complete quickly without wasting bandwidth

#### Scenario: Handle GitHub API failures gracefully
**Given** build-ci-image job runs
**When** GitHub API for mise version fails
**Then** job SHALL retry up to 3 times with exponential backoff
**And** SHALL use cached mise version from previous run as fallback
**And** SHALL NOT fail the entire pipeline

#### Scenario: Build reliability >99%
**Given** build-ci-image job runs over time
**When** success rate is measured
**Then** success rate SHALL be greater than 99%
**And** failures SHALL only be due to infrastructure issues, not code

---

### Requirement: Cache Management with Serialization

The cache configuration SHALL prevent corruption and handle concurrent pipelines safely.

#### Scenario: Cache key uses checksum file to track configuration changes
**Given** .gitlab-ci.yml cache configuration is defined
**When** cache key is computed
**Then** key SHALL use .cache-key as single file reference
**And** .cache-key SHALL contain SHA256 checksums of all critical files
**And** .cache-key SHALL include checksums for .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum
**And** any change to critical files SHALL change .cache-key
**And** cache SHALL be invalidated when .cache-key changes
**And** implementation SHALL respect GitLab's 2-file limit for cache:key:files

#### Scenario: Setup job writes cache exclusively
**Given** setup job runs
**When** job completes successfully
**Then** cache policy SHALL be pull-push
**And** job SHALL update Go module cache
**And** other jobs SHALL not write to cache

#### Scenario: Other jobs read cache only
**Given** lint or test job runs
**When** job accesses cache
**Then** cache policy SHALL be pull only
**And** job SHALL NOT write to cache
**And** job SHALL not modify cache contents

#### Scenario: Serialized cache writes prevent conflicts
**Given** multiple pipelines run simultaneously on main branch
**When** setup job from one pipeline writes cache
**Then** resource_group: cache_updates SHALL serialize writes
**And** only one pipeline SHALL write cache at a time
**And** other pipelines SHALL read consistent cache
**And** cache SHALL not be corrupted

#### Scenario: Artifact lifetime supports multi-stage pipelines
**Given** pipeline creates artifacts
**When** artifact expiration is configured
**Then** expiration SHALL be 24 hours
**And** artifacts SHALL be available to all downstream stages
**And** artifacts SHALL not expire mid-pipeline

---

### Requirement: Rollback Capability

The CI configuration SHALL support rollback to previous image version if issues occur.

#### Scenario: Rollback procedure documented
**Given** Alpine image is deployed
**When** issues are detected
**Then** tasks.md SHALL document rollback procedure
**And** procedure SHALL include reverting .gitlab-ci.yml image tag
**And** rollback conditions SHALL be specified:
  - >5% image pull failure rate
  - Tool errors (mise, golangci-lint, GoReleaser)
  - >20% build time increase

#### Scenario: Previous image tag available
**Given** new image is pushed to registry
**When** rollback is needed
**Then** previous Debian image tag SHALL still be available
**And** both latest and versioned tags SHALL exist for new image
**And** older versioned tags SHALL remain available

