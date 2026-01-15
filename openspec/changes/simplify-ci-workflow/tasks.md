# Implementation Tasks

## Task 1: Update Cache Configuration
- [x] Add .mise/config.toml to cache key files list
- [x] Change setup job cache policy to pull-push
- [x] Change lint and test job cache policies to pull
- [x] Add resource_group: cache_updates to setup job
- [x] Set artifact expiration to 24 hours
- [x] Ensure cache key includes: .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum

## Task 2: Standardize Base Images
- [x] Update mirror-to-github job to use `registry.gitlab.com/amoconst/germinator/ci:latest`
- [x] Remove alpine:latest base image usage
- [x] Ensure git is available in CI image
- [x] Test mirror job with new base image

## Task 3: Implement Variable Validation Rules
- [x] Add GitLab CI rule to mirror-to-github job:
  - if: '$CI_COMMIT_BRANCH == "main" && $GITHUB_ACCESS_TOKEN && $GITHUB_REPO_URL'
- [x] Remove silent exit code 0 when variables missing
- [x] Let GitLab CI automatically skip job when rule doesn't match
- [x] Remove apk add git from mirror-to-github (git in CI image)

## Task 4: Implement release:validate Task
- [x] Create .mise/tasks/release/validate.sh script
- [x] Check git status for uncommitted changes
- [x] Check current branch is main
- [x] Extract version from Git tag using simple grep/sed (strip 'v' prefix)
- [x] Read version from internal/version/version.go
- [x] Compare tag version with code version
- [x] Validate .goreleaser.yml syntax
- [x] Provide clear error messages for all validation failures
- [x] Add task to .mise/config.toml as [tasks."release:validate"]

## Dependencies

- Task 4: Depends on Task 1, 13

## Task 5: Remove release:check Task
- [x] Remove release:check task from .mise/config.toml
- [x] Update documentation to reference release:validate instead

## Task 6: Add Tag Stage to CI Pipeline
- [x] Add `tag` to stages list in .gitlab-ci.yml (after test, before release)
- [x] Create create-version-tag job in .gitlab-ci.yml
- [x] Configure create-version-tag job:
  - Stage: tag
  - Needs: [test]
  - Git config: Use $GITLAB_USER_EMAIL and $GITLAB_USER_NAME
  - Version extraction: Read from internal/version/version.go using grep/sed
  - Tag format: v<VERSION>
  - Tag creation: Use git tag -a and git push
  - Idempotent check: Skip if tag already exists
  - Remote URL: Configure for push with CI_JOB_TOKEN
- [x] Add rules to create-version-tag:
  - Run on main branch pushes when internal/version/version.go changes
  - Skip otherwise
- [x] Update workflow rules to allow tag stage on main pushes
- [x] Remove manual tagging instructions from AGENTS.md
- [x] Update pipeline documentation with tag stage

## Task 7: Test Tag Stage
- [ ] Test tag stage with version.go change (should create tag)
- [ ] Test tag stage without version.go change (should skip)
- [ ] Test tag stage when tag already exists (should skip idempotently)
- [ ] Verify tag format is correct (vX.Y.Z)
- [ ] Verify release stage triggers after tag creation
- [ ] Verify git config variables work correctly
- [ ] Test idempotent behavior (re-run pipeline with same version)

## Task 8: Test Mirror Job Behavior
- [ ] Test mirror job with all variables set (should run)
- [ ] Test mirror job with GITHUB_ACCESS_TOKEN missing (should skip)
- [ ] Test mirror job with GITHUB_REPO_URL missing (should skip)
- [ ] Verify job appears in pipeline but is skipped, not failed

## Task 9: Test release:validate Locally
- [x] Test with clean git state on main branch
- [x] Test with uncommitted changes (should fail)
- [x] Test on non-main branch (should fail)
- [x] Test with matching tag and version (should pass)
- [x] Test with mismatched tag and version (should fail)
- [x] Test with invalid GoReleaser config (should fail)

## Task 10: Update Release Job with Validation
- [x] Add mise run release:validate to release job's before_script in .gitlab-ci.yml
- [x] Ensure validation runs before GoReleaser
- [x] Add rules to skip job when only openspec files changed
- [ ] Test release job with matching tag (should succeed)
- [ ] Test release job with mismatched tag (should fail early)

## Dependencies

- Task 10: Depends on Task 4, 13

## Task 11: Test Cache Invalidation
- [ ] Modify .mise/config.toml to bump tool version
- [ ] Verify cache key changes
- [ ] Verify old cache is not used
- [ ] Verify Go modules are re-downloaded with new tools

## Task 12: Test Concurrent Pipeline Cache Safety
- [ ] Trigger multiple pipelines simultaneously on main branch
- [ ] Verify resource_group serializes setup job writes
- [ ] Verify no cache corruption occurs
- [ ] Verify all pipelines complete successfully

## Task 13: Update Documentation
- [x] Update AGENTS.md to reflect new CI configuration
- [x] Document cache key composition (5 files)
- [x] Document GitLab CI rules behavior for mirror job
- [x] Document resource_group for cache serialization
- [x] Update release workflow section with release:validate task
- [x] Document validation checks: git state, branch, tag match
- [x] Add troubleshooting section for CI and releases
- [x] Update release workflow to remove manual tagging steps
- [x] Document automatic tag creation workflow
- [x] Add tag stage to pipeline documentation

## Task 14: Create CI Workflow Spec
- [x] Define requirements for cache key composition (all 5 files)
- [x] Define requirements for serialized cache writes
- [x] Define requirements for GitLab CI rules variable validation
- [x] Define requirements for base image consistency
- [x] Define requirements for 24-hour artifact lifetime
- [x] Add scenario for concurrent pipeline cache safety
- [x] Add requirements for tag validation (simple grep/sed approach)
- [x] Add requirements for git state validation
- [x] Add requirements for branch validation
- [x] Add requirements for release:validate task
- [x] Add requirements for automatic tag creation
- [x] Add requirements for tag stage behavior
- [x] Add scenario for tag stage idempotency
- [x] Add scenario for version.go change triggering tag creation

## Task 15: Test Complete Release Workflow
- [ ] Perform end-to-end release with new validation
- [ ] Verify validation catches uncommitted changes
- [ ] Verify validation catches wrong branch
- [ ] Verify validation catches tag mismatches
- [ ] Verify validation catches invalid GoReleaser config
- [ ] Verify successful release creates proper artifacts
- [ ] Verify tag stage creates tag automatically
- [ ] Verify release stage triggers after tag creation
- [x] Verify AGENTS.md documentation is accurate

## Task 16: Test CI Optimization
- [ ] Trigger test MR with only openspec changes
- [ ] Verify pipeline shows only setup job running
- [ ] Verify lint, test, release, mirror jobs are skipped
- [ ] Verify all jobs complete successfully
- [x] Document expected behavior in AGENTS.md

## Task 17: Test Mixed Changes
- [ ] Trigger test MR with both openspec and code changes
- [ ] Verify pipeline runs all jobs normally
- [ ] Verify no unexpected job skipping occurs
- [x] Document full pipeline behavior in AGENTS.md

## Task 18: Optimize CI for openspec-only Changes
- [x] Add rules to skip jobs when only openspec files change
- [x] Apply rules to lint, test, release, mirror jobs
- [ ] Test that code changes trigger full pipeline (all jobs run)
- [ ] Test that openspec-only changes skip expensive jobs (lint, test, release)
- [ ] Verify setup job still runs for both cases

## Task 19: Create CI Workflow Optimization Spec
- [x] Define requirements for CI optimization using rules:changes
- [x] Document openspec-only change detection and job skipping behavior
- [x] Add scenarios for code changes triggering full pipeline
- [x] Add scenarios for documentation-only changes skipping expensive jobs

## Task 20: Test CI Optimization
- [ ] Trigger test MR with only openspec changes
- [ ] Verify pipeline shows only setup job running
- [ ] Verify lint, test, release, mirror jobs are skipped
- [ ] Trigger test MR with code changes
- [ ] Verify all jobs run normally
- [x] Document expected behavior in AGENTS.md

## Dependencies

- Task 1: No dependencies
- Task 2: Can run in parallel with Task 1
- Task 3: Can run in parallel with Task 1
- Task 4: Depends on Task 1, 13
- Task 5: No dependencies
- Task 6: Can run in parallel with Task 2, 3, 4
- Task 7: Depends on Task 6
- Task 8: Depends on Task 6
- Task 9: Depends on Task 4, 5
- Task 10: Depends on Task 4, 13
- Task 11: Depends on Task 1
- Task 12: Depends on Task 1, 2, 3
- Task 13: No dependencies
- Task 14: Depends on Task 1, 4, 10, 12, 13, 20
- Task 15: Depends on Tasks 7, 10
- Task 16: Depends on Task 13, 20
- Task 17: Depends on Tasks 13, 20
- Task 18: No dependencies
- Task 19: Depends on Tasks 18
- Task 20: Depends on Task 13, 19
