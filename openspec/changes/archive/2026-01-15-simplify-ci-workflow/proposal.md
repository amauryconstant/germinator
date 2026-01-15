# Simplify CI Workflow

## Why

The current CI workflow has disconnected validation, silent failures, poor cache management, manual release tagging, and unreliable CI image rebuilding:

- **Silent failures** when GitHub mirror is misconfigured (GITHUB_ACCESS_TOKEN missing) - confusing behavior
- **No version bump enforcement** - Code can be merged without version bump, leading to version drift
- **No validation** that Git tags match version.go or that git state is clean
- **Poor cache keys** don't invalidate when .mise/config.toml changes (tool version updates)
- **Potential cache corruption** from concurrent writes across multiple pipelines
- **No git state validation** - releases can be created with uncommitted changes
- **No branch validation** - releases can be created from non-main branches
- **Confusing validation tasks** - both release:check and release:validate exist
- **Manual tagging** - Developers must remember to create and push tags manually after version bumps, leading to forgotten or mistagged releases
- **CI image rebuild failures** - Current skip check uses mise version alone, which doesn't capture changes to Dockerfile.ci or .mise/config.toml, causing the CI image to not rebuild when it should

These issues lead to wasted CI time, undetected configuration problems, inconsistent build behavior, release process friction, and CI job failures due to stale CI images.

## What Changes

This change simplifies CI workflow with better validation, error handling, automated release tagging, reliable CI image rebuilding, and improved code quality:

- **Add .mise/config.toml to cache key** - Invalidate cache when tool versions change
- **Validate GitHub mirror variables** - Fail-fast with clear errors or skip gracefully
- **Standardize base images** - All jobs use CI image instead of alpine:latest
- **Improve cache policies** - Serialize writes to prevent corruption
- **Consolidate validation tasks** - Single release:validate task
- **Add git state validation** - Ensure clean working directory and main branch, ignore untracked files in CI context
- **Enforce version bump on code changes** - Require version.go update when cmd/ files change (prevents version drift)
- **Add GoReleaser dry-run validation** - Validate GoReleaser configuration on MRs before release
- **Add tag validation** - Validate Git tags against version.go (simple grep/sed approach)
- **Add CI integration** - Run validation in release job's before_script, accept detached HEAD when tag is found via git describe
- **Fix tag detection** - Use git describe as fallback when CI_COMMIT_TAG is not set, ensure detached HEAD is only accepted with valid tag
- **Set artifact lifetime** - 24 hours across all stages
- **Scope out prerelease support** - Keep validation simple
- **Add automatic tag creation** - Create Git tags when internal/version/version.go changes, replacing manual tagging workflow
- **Add tag stage** - New stage after test that creates tags idempotently
- **Integrate with release workflow** - Tags trigger release stage automatically
- **Add hash-based CI image tagging** - Tag CI images with mise version + content hash (format: 2026.1.2-abc123def456) to ensure CI image rebuilds when Dockerfile.ci or .mise/config.toml changes
- **Add docker CLI to CI image** - Include docker-cli package to enable docker commands in release job
- **Upgrade DIND service version** - Update docker:dind service from 24.0.5 to latest (29.1.4) to support newer docker CLI API version
- **Add tag job resource group** - Serialize tag creation to prevent race conditions from concurrent pipelines
- **Improve CI code quality** with YAML anchors for shared cache configuration, interruptible flags for long-running jobs, and dead code removal
- **Consolidate pipeline stages** - Reduce from 7 stages to 5 stages by merging lint/test into "validate" stage and release/mirror into "distribute" stage
- **Parallelize validation jobs** - Lint and test run in parallel in same stage to reduce pipeline duration
- **Parallelize distribution jobs** - Release and mirror run in parallel in same stage
- **Update tag job dependencies** - Tag job runs independently (no dependency on test), allowing version tracking regardless of test state
- **Update mirror job dependencies** - Mirror job depends on tag, only running when version.go changes (GitHub mirror syncs on releases only)
- **Rename stages for clarity** - "lint" → "validate", "release/mirror" → "distribute"
- **Simplify release job** - Use goreleaser/goreleaser official image with inline validation for CI simplicity
- **Add GoReleaser dry-run job** - Validate GoReleaser configuration on MRs to catch config issues early
- **Add version bump enforcement job** - Require version.go update when cmd/ files change on MRs
- **Override goreleaser/goreleaser entrypoint** - Add entrypoint: [""] to prevent image's default entrypoint from interfering with GitLab CI script execution
- **Remove release:validate task** - Deprecate local validation script (.mise/tasks/release/validate.sh removed)
- **Add GIT_DEPTH support** - Enable GoReleaser changelog generation with full git history
- **Remove Docker service from release** - Remove unnecessary Docker-in-Docker service (no Docker images being pushed)
- **Simplify mirror job** - Use force-push strategy (--force) since GitLab is source of truth, fetch github remote before pushing
- **Simplify release job rules** - Release job only runs on tags (removed MR/openspec skip rules)

## Impact

**Affected Specs:**
- New spec for `ci-workflow` covering validation, cache management, and pipeline optimization
- Delta changes to `release-management` spec for validation requirements

**Affected Code:**
- `.gitlab-ci.yml` - Add CI optimization rules to skip expensive jobs on openspec-only changes, improve cache configuration, standardize base images, add tag stage with automatic tag creation, implement hash-based CI image tagging, consolidate pipeline stages, simplify release job with goreleaser/goreleaser image and entrypoint override, simplify mirror job with force-push strategy, add version bump enforcement check, add GoReleaser dry-run validation, add tag job resource group, simplify release job rules to only run on tags
- `.mise/tasks/release/validate.sh` - Remove (deprecated, replaced by inline validation)
- `.mise/config.toml` - Remove release:validate task
- `AGENTS.md` - Update release workflow documentation, remove manual tagging steps, document docker CLI in CI image
- `Dockerfile.ci` - Add docker-cli package to enable docker commands in release job
- `.goreleaser.yml` - Use GitLab job token for releases

**Note**: Delta changes to `release-management` spec are included in `simplify-ci-workflow/specs/release-management/spec.md`. When this proposal is archived, these deltas will be applied to `openspec/specs/release-management/spec.md` as part of the archive process. This allows all spec changes to be visible in one location while maintaining clear proposal scope.

**Affected Workflows:**
- GitHub mirror job fails clearly when variables missing (or skips via CI rules)
- Cache invalidates properly when .mise/config.toml changes (tool version updates)
- All stages use consistent tooling environment (CI image everywhere)
- Concurrent pipelines handle cache safely (serialized writes via resource_group)
- Concurrent tag creation handled safely (serialized writes via resource_group: version_tagging)
- Release validation catches uncommitted changes, wrong branch, tag mismatches
- Version bump enforcement prevents merging code changes without version.go updates
- GoReleaser dry-run validates configuration on MRs before release
- Single validation task reduces confusion
- Expensive CI jobs (lint, test, release, mirror) automatically skipped when only documentation changes, saving CI resources and time
- Automatic tag creation when internal/version/version.go changes, eliminating manual tagging steps
- Developers only need to run `mise run version:*` to bump version, push to main, and watch CI create tag and release
- CI image rebuilds automatically when Dockerfile.ci or .mise/config.toml changes via content-based hash tagging
- CI image tags use format mise-version-content-hash (e.g., 2026.1.2-abc123def456) for unique identification
- CI image build is skipped when Dockerfile.ci and .mise/config.toml are unchanged, saving CI time
- Docker CLI available in CI image, enabling docker login and docker commands in release job
- Improved code maintainability with YAML anchors reducing duplicate cache configuration
- Better CI resource management with interruptible flags on lint and mirror jobs
- Cleaner pipeline configuration with dead code (when: never) removed
- Version bump enforced when cmd/ files change on MRs
- GoReleaser configuration validated on MRs before release
- Tag creation serialized to prevent duplicate tags from concurrent pipelines
- Reduced pipeline stage count from 7 to 5 (build-ci, setup, validate, tag, distribute)
- Lint and test jobs run in parallel in "validate" stage, reducing validation time
- Release and mirror jobs run in parallel in "distribute" stage, reducing distribution time
- Tag job creates version tags independently of test state (allows version tracking even with failed tests)
- GitHub mirror job only runs when version.go changes (on releases), not on every main push
- GitHub mirror job force-pushes to GitHub (--force) since GitLab is source of truth
- Clearer stage semantics with renamed stages (validate, distribute)
- Release job uses goreleaser/goreleaser official image for simplicity
- Release job overrides entrypoint with [""] to prevent script execution errors
- Release job validates version match inline (only check needed in CI)
- Release job no longer has Docker service overhead (simpler, faster)
- GoReleaser generates changelogs properly with GIT_DEPTH=0
- Mirror job fetches github remote and force-pushes with --tags --force
- Mirror job uses HEAD:$CI_COMMIT_BRANCH reference for consistent branch pushing
- Release job rules simplified to only run on tags
- Deprecated release:validate task removed from configuration
- GoReleaser dry-run job validates configuration on MRs
- Version bump enforcement job checks for version.go updates on MRs

## Dependencies

None - this change is independent and can be implemented at any time.

## Migration Path

**Breaking Changes:**
- GitHub mirror job behavior changes from mirroring on every main push to mirroring only when version.go changes (on releases)
- GitHub mirror job uses force-push (--force) strategy (GitLab is source of truth, any GitHub changes will be overwritten)
- MRs with cmd/ changes require version.go update (enforced by check-version-bump job)
- `mise run release:validate` task is removed (deprecated, no longer maintained)
- Developers can no longer run local validation via mise task (inline validation in CI only)

**Migration Impact:**
- GitHub mirror will no longer sync code on normal main pushes (without version changes)
- GitHub mirror will only run when a version tag is created
- This may affect workflows that rely on continuous GitHub mirroring
- MRs that change cmd/ files must also update internal/version/version.go (new enforcement)
- GoReleaser dry-run job will run on MRs to validate configuration
- Local release validation is removed (not used by developers)

**No other breaking changes.** Existing workflows continue to work with better validation, error detection, and improved pipeline structure.

## Risks

- **Cache key changes** may cause temporary cache misses during transition
- **GitLab CI rules** must correctly identify when to skip mirror job
- **Standardizing base images** requires ensuring all jobs work with CI image
- **Validation logic** must correctly parse version strings using simple grep/sed
- **Mirror job dependency on tag** means GitHub will no longer mirror on every main push, only on releases
- **Tag job independence** means version tags may be created for code with failing tests
- **Parallel jobs in same stage** may race for shared resources (currently no shared resources between lint/test or release/mirror)
- **Reduced stage visibility** may make it harder to track which specific job failed in pipeline UI
- **Version bump enforcement** may cause friction if version updates are forgotten or inappropriate for minor changes
- **GoReleaser dry-run** job may fail on MRs with configuration issues, blocking merging until fixed

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
- Release job validates tag version match with inline validation in before_script
- AGENTS.md updated with validation checks and troubleshooting
- Deprecated release:validate task removed (unused, replaced by inline validation)
- Tag stage runs when internal/version/version.go changes on main branch
- Tag stage creates tags with format v<VERSION> (e.g., v0.3.0)
- Tag stage is idempotent - skips creation if tag already exists
- Tag stage runs independently (no dependency on test) to allow version tracking regardless of test state
- Tag stage creates tags even if tests fail
- Tag stage uses $GITLAB_USER_EMAIL and $GITLAB_USER_NAME for git config
- Release stage triggers automatically after tag stage creates tag
- Manual tagging workflow removed from documentation
- CI image tagged with format mise-version-content-hash (e.g., 2026.1.2-abc123def456)
- CI image rebuilds when Dockerfile.ci or .mise/config.toml changes
- CI image build skipped when Dockerfile.ci and .mise/config.toml are unchanged
- Docker CLI available in CI image for docker commands
- Release job can successfully run docker login with docker CLI
- Docker DIND service upgraded to version 29.1.4 (latest) to support docker CLI API version
- No API version mismatch errors between docker CLI and DIND service
- Version bump enforcement job prevents merging code changes without version.go updates
- Version bump enforcement job checks for cmd/ file changes and requires version.go update
- GoReleaser dry-run job validates configuration on MRs
- GoReleaser dry-run job runs on MRs when .goreleaser.yml or Go files change
- GoReleaser dry-run job uses snapshot mode and skips publish (no artifacts created)
- Pipeline has exactly 5 stages: build-ci, setup, validate, tag, distribute
- Lint and test jobs run in parallel in "validate" stage
- Release and mirror jobs run in parallel in "distribute" stage
- Tag job has no dependencies (runs independently after setup)
- Mirror job depends on tag (only runs when version.go changes)
- Release job depends on lint and test (ensures validation passes before release)
- Pipeline execution time reduced by ~50% for validation phase (lint + test parallel)
- Pipeline execution time reduced by ~40% for distribution phase (release + mirror parallel)
- GitHub mirror only runs when version.go changes (on releases)
- Release job uses goreleaser/goreleaser image with entrypoint override
- Mirror job uses force-push strategy (--force)
- Mirror job fetches github remote before pushing
- Mirror job uses HEAD:$CI_COMMIT_BRANCH --tags --force to push branch and tags together
- Tag job uses resource_group: version_tagging to serialize concurrent tag creation
- Release job rules simplified to only run on tags (removed MR/openspec skip rules)
