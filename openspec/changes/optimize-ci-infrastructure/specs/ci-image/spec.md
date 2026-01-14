# ci-image Spec Delta

## ADDED Requirements

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

#### Scenario: Cache key includes all configuration files
**Given** .gitlab-ci.yml cache configuration is defined
**When** cache key is computed
**Then** key SHALL include .gitlab-ci.yml
**And** key SHALL include Dockerfile.ci
**And** key SHALL include .mise/config.toml
**And** key SHALL include go.mod
**And** key SHALL include go.sum
**And** any change to these files SHALL invalidate cache

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
