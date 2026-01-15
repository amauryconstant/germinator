# Implementation Tasks

## Task 1: Research Alpine Base Image Compatibility
- [x] Test mise, golangci-lint, GoReleaser on golang:1.25.5-alpine3.23
- [x] Document any compatibility issues or workarounds needed
- [x] Verify static binary compilation works correctly
- [x] Document findings in inline comments or temporary notes

### Research Results:
- Base image: golang:1.25.5-alpine3.23 works correctly
- mise 2026.1.2: Installs and runs correctly
- golangci-lint 2.8.0: Installs and runs correctly
- GoReleaser 2.13.3: Installs and runs correctly
- Go 1.25.5: Available and functional
- Binary compilation: Works correctly
- Required packages: git, curl, ca-certificates, bash (standard Alpine packages)
- No special workarounds needed

## Task 2: Document Alpine Migration Decision
- [x] Document tool compatibility test results
- [x] Confirm final base image selection: golang:1.25.5-alpine3.23
- [x] Note any package dependencies required on Alpine
- [x] Archive research findings for future reference

## Task 3: Migrate Dockerfile.ci to Alpine
- [x] Update base image from golang:1.25.5-bookworm to golang:1.25.5-alpine3.23
- [x] Replace apt-get commands with apk equivalents
- [x] Install required Alpine packages (ca-certificates, git, curl)
- [x] Optimize RUN commands to reduce layers
- [x] Clean up build artifacts in each layer

## Task 4: Verify Alpine Image Builds Successfully
- [x] Build new Alpine-based CI image locally
- [x] Verify mise installs and runs correctly
- [x] Verify golangci-lint installs and runs correctly
- [x] Verify GoReleaser installs and runs correctly
- [x] Verify Go 1.25.5 is available

## Task 5: Implement Reliable Image Build Process
- [x] Replace docker pull existence check with docker manifest inspect
- [x] Add retry logic for GitHub API failures (up to 3 retries)
- [x] Use cached mise version as fallback if API fails
- [x] Implement proper error handling and messages

## Task 6: Update Cache Configuration
- [x] Generate .cache-key checksum file combining all 5 critical files
- [x] Use .cache-key as single cache:key:files reference (respects GitLab's 2-file limit)
- [x] Change setup job cache policy to pull-push
- [x] Change lint and test job cache policies to pull
- [x] Add resource_group: cache_updates to setup job
- [x] Set artifact expiration to 24 hours

### Cache Strategy Details

**Constraint:** GitLab CI limits `cache:key:files` to maximum 2 files.

**Solution:** Use combined checksum approach:
- Setup job generates `.cache-key` file containing SHA256 checksums of .gitlab-ci.yml, Dockerfile.ci, .mise/config.toml, go.mod, go.sum
- Use `.cache-key` as single `cache:key:files` reference
- Any change to 5 source files changes `.cache-key` → new cache created

## Task 7: Update CI Image to Use Alpine
- [x] Push Alpine-based CI image to GitLab registry with mise version tag
- [x] Update .gitlab-ci.yml default image to new Alpine version
- [x] Keep both latest and versioned tags for rollback capability

## Task 8: Test CI Pipeline with New Configuration
- [x] Run full CI pipeline with Alpine image
- [x] Verify cache invalidation works when config files change
- [x] Verify resource_group prevents concurrent write conflicts
- [x] Verify all jobs complete successfully
- [x] Monitor for any tool errors or failures

## Task 9: Document Rollback Procedure
- [x] Document rollback steps in tasks.md comments:
  - Revert .gitlab-ci.yml to previous image tag
  - Rollback conditions: >5% image pull failure rate, tool errors, >20% build time increase
- [x] Document how to verify if rollback is needed
- [x] Add to AGENTS.md release management section

### Rollback Procedure

If issues are detected after deploying the Alpine-based CI image, use this rollback procedure:

#### Rollback Steps
1. Revert .gitlab-ci.yml default image to previous Debian version:
   ```yaml
   default:
     image: registry.gitlab.com/amoconst/germinator/ci:latest
   ```
2. Rebuild and push the Debian-based CI image if needed (from commit before Alpine migration)
3. Commit and push the rollback changes
4. Run a pipeline to verify the rollback resolves issues

#### Rollback Conditions
Initiate rollback if any of the following conditions are met:
- Image pull failure rate exceeds 5%
- Tool errors occur (mise, golangci-lint, or GoReleaser failures)
- Pipeline build time increases by more than 20%
- Critical functionality breaks in CI pipeline

#### Verification
To determine if rollback is needed:
1. Monitor pipeline job success rates in GitLab CI
2. Check job duration trends for significant increases
3. Review job logs for tool-specific errors
4. Compare metrics against Debian baseline (before Alpine migration)

#### Note
- Previous Debian image tags remain available in GitLab registry
- Both latest and versioned tags exist for rollback capability
- Alpine and Debian images can coexist in registry for testing

## Dependencies

- Task 1: No dependencies
- Task 2: Depends on Task 1
- Task 3: Depends on Task 2 (confirmed Alpine compatibility)
- Task 4: Depends on Task 3
- Task 5: Can run in parallel with Tasks 3, 4
- Task 6: Can run in parallel with Tasks 3, 4
- Task 7: Depends on Tasks 4, 5, 6
- Task 8: Depends on Task 7
- Task 9: Depends on Task 8
