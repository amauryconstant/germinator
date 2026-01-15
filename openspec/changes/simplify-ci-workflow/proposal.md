# Simplify CI Workflow

## Why

The current CI workflow has disconnected validation, silent failures, poor cache management, and manual release tagging:

- **Silent failures** when GitHub mirror is misconfigured (GITHUB_ACCESS_TOKEN missing) - confusing behavior
- **No validation** that Git tags match version.go or that git state is clean
- **Poor cache keys** don't invalidate when .mise/config.toml changes (tool version updates)
- **Potential cache corruption** from concurrent writes across multiple pipelines
- **No git state validation** - releases can be created with uncommitted changes
- **No branch validation** - releases can be created from non-main branches
- **Confusing validation tasks** - both release:check and release:validate exist
- **Manual tagging** - Developers must remember to create and push tags manually after version bumps, leading to forgotten or mistagged releases

These issues lead to wasted CI time, undetected configuration problems, inconsistent build behavior, and release process friction.

## What Changes

This change simplifies CI workflow with better validation, error handling, and automated release tagging:

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
- **Add automatic tag creation** - Create Git tags when internal/version/version.go changes, replacing manual tagging workflow
- **Add tag stage** - New stage after test that creates tags idempotently
- **Integrate with release workflow** - Tags trigger release stage automatically

## Impact

**Affected Specs:**
- New spec for `ci-workflow` covering validation, cache management, and pipeline optimization
- Delta changes to `release-management` spec for validation requirements

**Affected Code:**
- `.gitlab-ci.yml` - Add CI optimization rules to skip expensive jobs on openspec-only changes, improve cache configuration, standardize base images, add tag stage with automatic tag creation
- `.mise/tasks/` - Add release:validate task, remove release:check task
- `AGENTS.md` - Update release workflow documentation, remove manual tagging steps

**Note**: Delta changes to `release-management` spec are included in `simplify-ci-workflow/specs/release-management/spec.md`. When this proposal is archived, these deltas will be applied to `openspec/specs/release-management/spec.md` as part of the archive process. This allows all spec changes to be visible in one location while maintaining clear proposal scope.

**Affected Workflows:**
- GitHub mirror job fails clearly when variables missing (or skips via CI rules)
- Cache invalidates properly when .mise/config.toml changes (tool version updates)
- All stages use consistent tooling environment (CI image everywhere)
- Concurrent pipelines handle cache safely (serialized writes via resource_group)
- Release validation catches uncommitted changes, wrong branch, tag mismatches
- Single validation task reduces confusion
- Expensive CI jobs (lint, test, release, mirror) automatically skipped when only documentation changes, saving CI resources and time
- Automatic tag creation when internal/version/version.go changes, eliminating manual tagging steps
- Developers only need to run `mise run version:*` to bump version, push to main, and watch CI create tag and release

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
- Tag stage runs when internal/version/version.go changes on main branch
- Tag stage creates tags with format v<VERSION> (e.g., v0.3.0)
- Tag stage is idempotent - skips creation if tag already exists
- Tag stage runs after test stage completes successfully
- Tag stage uses $GITLAB_USER_EMAIL and $GITLAB_USER_NAME for git config
- Release stage triggers automatically after tag stage creates tag
- Manual tagging workflow removed from documentation
