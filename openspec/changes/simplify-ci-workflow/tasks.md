# Implementation Tasks

## Task 1: Update Cache Configuration
- Add .mise/config.toml to cache key files list
- Change setup job cache policy to pull-push
- Change lint and test job cache policies to pull
- Add resource_group: cache_updates to setup job
- Set artifact expiration to 24 hours
- Ensure cache key includes: .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum

## Task 2: Standardize Base Images
- Update mirror-to-github job to use `registry.gitlab.com/amoconst/germinator/ci:latest`
- Remove alpine:latest base image usage
- Ensure git is available in CI image
- Test mirror job with new base image

## Task 3: Implement Variable Validation Rules
- Add GitLab CI rule to mirror-to-github job:
  - if: '$CI_COMMIT_BRANCH == "main" && $GITHUB_ACCESS_TOKEN && $GITHUB_REPO_URL'
- Remove silent exit code 0 when variables missing
- Let GitLab CI automatically skip job when rule doesn't match
- Remove apk add git from mirror-to-github (git in CI image)

## Task 4: Implement release:validate Task
- Create .mise/tasks/release-validate.sh script
- Check git status for uncommitted changes
- Check current branch is main
- Extract version from Git tag using simple grep/sed (strip 'v' prefix)
- Read version from internal/version/version.go
- Compare tag version with code version
- Validate .goreleaser.yml syntax
- Provide clear error messages for all validation failures
- Add task to .mise/config.toml as [tasks."release:validate"]

## Dependencies

- Task 4: Depends on Task 1, 11

## Task 5: Remove release:check Task
- Remove release:check task from .mise/config.toml
- Update documentation to reference release:validate instead

## Task 6: Test Mirror Job Behavior
- Test mirror job with all variables set (should run)
- Test mirror job with GITHUB_ACCESS_TOKEN missing (should skip)
- Test mirror job with GITHUB_REPO_URL missing (should skip)
- Verify job appears in pipeline but is skipped, not failed

## Task 7: Test release:validate Locally
- Test with clean git state on main branch
- Test with uncommitted changes (should fail)
- Test on non-main branch (should fail)
- Test with matching tag and version (should pass)
- Test with mismatched tag and version (should fail)
- Test with invalid GoReleaser config (should fail)

## Task 8: Update Release Job with Validation
- Add mise run release:validate to release job's before_script in .gitlab-ci.yml
- Ensure validation runs before GoReleaser
- Add rules: !reference [.skip_on_openspec_only] to skip job when only openspec files changed
- Test release job with matching tag (should succeed)
- Test release job with mismatched tag (should fail early)

## Dependencies

- Task 8: Depends on Task 4, 11

## Task 9: Test Cache Invalidation
- Modify .mise/config.toml to bump tool version
- Verify cache key changes
- Verify old cache is not used
- Verify Go modules are re-downloaded with new tools

## Task 10: Test Concurrent Pipeline Cache Safety
- Trigger multiple pipelines simultaneously on main branch
- Verify resource_group serializes setup job writes
- Verify no cache corruption occurs
- Verify all pipelines complete successfully

## Task 11: Update Documentation
- Update AGENTS.md to reflect new CI configuration
- Document cache key composition (5 files)
- Document GitLab CI rules behavior for mirror job
- Document resource_group for cache serialization
- Update release workflow section with release:validate task
- Document validation checks: git state, branch, tag match
- Add troubleshooting section for CI and releases
- Document that version bumping remains manual

## Task 12: Create CI Workflow Spec
- Define requirements for cache key composition (all 5 files)
- Define requirements for serialized cache writes
- Define requirements for GitLab CI rules variable validation
- Define requirements for base image consistency
- Define requirements for 24-hour artifact lifetime
- Add scenario for concurrent pipeline cache safety
- Add requirements for tag validation (simple grep/sed approach)
- Add requirements for git state validation
- Add requirements for branch validation
- Add requirements for release:validate task

## Task 13: Test Complete Release Workflow
- Perform end-to-end release with new validation
- Verify validation catches uncommitted changes
- Verify validation catches wrong branch
- Verify validation catches tag mismatches
- Verify validation catches invalid GoReleaser config
- Verify successful release creates proper artifacts
- Verify AGENTS.md documentation is accurate

## Task 14: Test CI Optimization
- Trigger test MR with only openspec changes
- Verify pipeline shows only setup job running
- Verify lint, test, release, mirror jobs are skipped
- Verify all jobs complete successfully
- Document expected behavior in AGENTS.md

## Task 15: Test Mixed Changes
- Trigger test MR with both openspec and code changes
- Verify pipeline runs all jobs normally
- Verify no unexpected job skipping occurs
- Document full pipeline behavior in AGENTS.md

## Task 11: Optimize CI for openspec-only Changes
- Add `.skip_on_openspec_only` anchor to .gitlab-ci.yml
- Apply `rules: !reference [.skip_on_openspec_only]` to lint, test, release, mirror jobs
- Test that code changes trigger full pipeline (all jobs run)
- Test that openspec-only changes skip expensive jobs (lint, test, release)
- Verify setup job still runs for both cases

## Task 12: Create CI Workflow Optimization Spec
- Define requirements for CI optimization using rules:changes
- Document openspec-only change detection and job skipping behavior
- Add scenarios for code changes triggering full pipeline
- Add scenarios for documentation-only changes skipping expensive jobs

## Task 13: Test CI Optimization
- Trigger test MR with only openspec changes
- Verify pipeline shows only setup job running
- Verify lint, test, release, mirror jobs are skipped
- Trigger test MR with code changes
- Verify all jobs run normally
- Document expected behavior in AGENTS.md

## Dependencies

- Task 1: No dependencies
- Task 2: Can run in parallel with Task 1
- Task 3: Can run in parallel with Task 1
- Task 4: Depends on Task 1, 11
- Task 5: No dependencies
- Task 6: Can run in parallel with Task 2, 3, 4
- Task 7: Depends on Task 4, 5
- Task 8: Depends on Task 4, 11
- Task 9: Depends on Task 1
- Task 10: Depends on Task 1, 2, 3
- Task 11: No dependencies
- Task 12: Depends on Task 1, 4, 8, 10, 11, 13
- Task 13: Depends on Task 11, 12
- Task 14: Depends on Task 13
- Task 15: Depends on Task 13, 14
- Task 16: Depends on Tasks 11, 15
