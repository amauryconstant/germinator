# ci-image Specification

## Purpose
Provide custom Docker image for GitLab CI with pre-installed Go, mise, and development tools for faster and consistent builds.

## ADDED Requirements

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
