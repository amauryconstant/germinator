# Simplify CI Workflow

## Why

The current CI workflow has disconnected validation, silent failures, and poor cache management:

- **Silent failures** when GitHub mirror is misconfigured (GITHUB_ACCESS_TOKEN missing) - confusing behavior
- **No validation** that Git tags match version.go or that git state is clean
- **Poor cache keys** don't invalidate when .mise/config.toml changes (tool version updates)
- **Potential cache corruption** from concurrent writes across multiple pipelines
- **No git state validation** - releases can be created with uncommitted changes
- **No branch validation** - releases can be created from non-main branches
- **Confusing validation tasks** - both release:check and release:validate exist

These issues lead to wasted CI time, undetected configuration problems, and inconsistent build behavior.

## What Changes

This change simplifies CI workflow with better validation and error handling:

- **Add .mise/config.toml to cache key** - Invalidate cache when tool versions change
- **Validate GitHub mirror variables** - Fail-fast with clear errors or skip gracefully
- **Standardize base images** - All jobs use CI image instead of alpine:latest
- **Improve cache policies** - Serialize writes to prevent corruption
- **Consolidate validation tasks** - Single release:validate task
- **Add git state validation** - Ensure clean working directory and main branch
- **Add tag validation** - Validate Git tags against version.go (simple grep/sed approach)
- **Add CI integration** - Run validation in release job's before_script
- **Set artifact lifetime** - 24 hours across all stages
- **Scope out prerelease support** - Keep validation simple

## Impact

**Affected Specs:**
- New spec for `ci-workflow` covering validation, cache management, and pipeline optimization
- Delta changes to `release-management` spec for validation requirements

**Affected Code:**
- `.gitlab-ci.yml` - Add CI optimization rules to skip expensive jobs on openspec-only changes, improve cache configuration, standardize base images
- `.mise/tasks/` - Add release:validate task, remove release:check task
- `AGENTS.md` - Update release workflow documentation

**Note**: Delta changes to `release-management` spec are included in `simplify-ci-workflow/specs/release-management/spec.md`. When this proposal is archived, these deltas will be applied to `openspec/specs/release-management/spec.md` as part of the archive process. This allows all spec changes to be visible in one location while maintaining clear proposal scope.

**Affected Workflows:**
- GitHub mirror job fails clearly when variables missing (or skips via CI rules)
- Cache invalidates properly when .mise/config.toml changes (tool version updates)
- All stages use consistent tooling environment (CI image everywhere)
- Concurrent pipelines handle cache safely (serialized writes via resource_group)
- Release validation catches uncommitted changes, wrong branch, tag mismatches
- Single validation task reduces confusion
- Expensive CI jobs (lint, test, release, mirror) automatically skipped when only documentation changes, saving CI resources and time

## Dependencies

None - this change is independent and can be implemented at any time.

## Migration Path

No breaking changes. Existing workflows continue to work, with better validation and error detection.

## Risks

- **Cache key changes** may cause temporary cache misses during transition
- **GitLab CI rules** must correctly identify when to skip mirror job
- **Standardizing base images** requires ensuring all jobs work with CI image
- **Validation logic** must correctly parse version strings using simple grep/sed

## Success Criteria

- Cache invalidates on .mise/config.toml changes
- GitHub mirror fails clearly when variables missing (or skips via CI rules)
- All pipeline stages use `registry.gitlab.com/amoconst/germinator/ci:latest` as base image
- Cache writes are serialized via resource_group to prevent corruption
- Artifact lifetime is 24 hours across all stages
- Single release:validate task validates git state, tag format, and GoReleaser config
- Invalid Git tags (mismatched with version.go) detected and rejected before release
- Uncommitted changes detected and rejected before release
- Releases only allowed from main branch
- Release job validates with mise run release:validate in before_script
- AGENTS.md updated with validation checks and troubleshooting
