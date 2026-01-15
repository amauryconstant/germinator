# Optimize CI Infrastructure

## Why

The current CI infrastructure suffers from inefficiencies that impact developer experience and operational costs:

- **Large CI image size** (~807MB Debian) causes slow startup times, higher storage costs, and inefficient resource utilization
- **Unreliable CI image build process** pulls entire images for existence checks, uses brittle GitHub API parsing, and lacks proper error handling
- **Poor cache management** leads to stale caches, missed invalidations, and potential corruption from concurrent writes

These issues result in longer feedback loops, wasted CI resources, and inconsistent build behavior across pipelines.

## What Changes

This change optimizes CI infrastructure to improve efficiency, reliability, and maintainability:

- **Migrate CI Docker image to Alpine** (golang:1.25.5-alpine3.23) for 73% size reduction (807MB → 215MB)
- **Improve CI image build reliability** with better existence checking, race condition handling, and robust error recovery
- **Enhance cache strategy** with proper invalidation policies, safe concurrent access, and serialized writes

### Cache Strategy Implementation

**Constraint:** GitLab CI has a hard limit of 2 files maximum for `cache:key:files`. This prevents directly listing all 5 critical files (go.mod, go.sum, .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml).

**Solution:** Use a combined checksum file approach that:
- Generates a `.cache-key` file containing SHA256 checksums of all 5 critical files
- Uses `.cache-key` as the single `cache:key:files` reference
- Ensures cache invalidation when any of the 5 files change

**Implementation:**
- Setup job calculates checksums for all 5 files using Alpine's `sha256sum` command
- Checksums are combined deterministically into `.cache-key` file
- Cache key changes when `.cache-key` content differs
- Resource group (`cache_updates`) serializes cache write operations to prevent concurrent corruption

## Impact

**Affected Specs:**
- `ci-image` - Update image requirements for Alpine base and build reliability
- New capabilities for cache management with serialization

**Affected Code:**
- `Dockerfile.ci` - Migrate to Alpine base image, optimize layers
- `.gitlab-ci.yml` - Improve build-ci-image job, add resource_group for cache writes, update cache configuration

**Affected Workflows:**
- CI pipeline startup time decreases (faster image pulls)
- CI image builds become more reliable
- Cache invalidation prevents stale dependencies
- Concurrent pipeline runs handle cache safely via serialization

## Dependencies

None - this change is self-contained and can be implemented independently

## Migration Path

No breaking changes to existing workflows. The optimized CI image will be built and deployed automatically, with rollback procedure documented in tasks.md if issues occur.

## Risks

- **Alpine base image** requires verification that all tools work correctly (mise, golangci-lint, GoReleaser)
- **Cache strategy changes** may cause initial cache misses as new keys are established
- **Build process changes** may need validation in concurrent pipeline scenarios

## Success Criteria

- CI image size reduced to ~215MB (target golang:1.25.5-alpine3.23)
- CI pipeline startup time improved by at least 30%
- CI image build reliability >99% (no failures from race conditions or API issues)
- Cache invalidation properly handles config changes (no stale caches)
- Cache key respects GitLab's 2-file limit for `cache:key:files`
- All existing tools (mise, golangci-lint, GoReleaser) remain functional on Alpine
