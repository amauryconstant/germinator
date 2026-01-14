# Proposal: Implement Professional Release Management

## Summary

Implement industry-standard release management using GoReleaser for automated cross-platform builds, checksums, SBOMs, and GitLab releases. Enhance mise to manage all development tools (GoReleaser, golangci-lint). Create custom Docker image for CI with pre-installed tools. Replace custom GitLab CI ldflags injection with professional-grade release automation.

## Prerequisites

Before implementing this proposal, ensure:

- **GitLab Container Registry** is enabled in project settings
- **Container Registry push permissions** are configured for CI/CD
- **GitLab CI/CD variables** include `GITLAB_TOKEN` with `api` scope
- **GitLab project** has `releases` permission enabled for CI job

## Motivation

Current implementation has manual version management via hardcoded constants and custom GitLab CI with ldflags injection. This approach:
- Lacks professional release artifacts (checksums, SBOMs, archives)
- Requires maintaining complex CI scripts for builds
- No package manager integration support
- Reproducible builds not guaranteed
- Technical debt in maintenance of shell scripts

We need:
- **Automated Release Management**: GoReleaser for cross-platform builds, archives, checksums, SBOMs
- **Tool Versioning**: mise manages GoReleaser, golangci-lint with easy updates
- **CI Image**: Custom Docker image with pre-installed tools for faster CI
- **Industry Standard**: Used by Google, Microsoft, AWS, HashiCorp, Kubernetes

This approach follows industry best practices for 2025 Go CLI tools, eliminates technical debt, and provides professional-grade releases.

## Proposed Change

**Feature 1: GoReleaser Integration (release-management)**
- Create `.goreleaser.yml` configuration
- Configure cross-platform builds: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Generate archives: .tar.gz (Linux/macOS), .zip (Windows)
- Generate SHA256 checksums for all artifacts
- Generate SBOMs using Syft (SPDX format)
- Auto-generate release notes from commit history
- Inject version, commit, date via ldflags
- Configure GitLab release integration

**Feature 2: mise Tool Management (mise-tools)**
- Add GoReleaser to `.mise/config.toml`
- Add golangci-lint to `.mise/config.toml` (pin version: 1.60.1)
- Add version injection tasks to `.mise/config.toml`
- Create `.mise/tasks/tools/check.sh` for checking tool updates
- Create `.mise/tasks/tools/update.sh` for updating tool versions
- Add release-related tasks: `release:dry-run`, `release:check`

**Feature 3: Custom CI Docker Image (ci-image)**
- Create `Dockerfile.ci` based on `golang:1.25.5-bookworm`
- Install mise (latest stable) in image
- Install GoReleaser, golangci-lint via mise in image
- Add `build-ci-image` job to `.gitlab-ci.yml` for automated image building
- Version image with mise version tag
- Verify all tools work in image

**Feature 4: Enhanced Version Management (cli-framework modification)**
- Update `internal/version/version.go`: change `Version` from `const` to `var`
- Add `Commit` and `Date` variables to version package
- Update `cmd/version.go` to display full version info
- Format: `germinator v0.2.0 (abc1234) 2025-01-13`

**Feature 5: Documentation and Installation (release-management)**
- Create `INSTALL.md` with installation instructions
- Create `install.sh` at project root for one-line user installation
- Add `install:local` task to `.mise/config.toml` (simple build and install)
- Document checksum verification
- Document GPG signature verification (optional)
- Add SBOM documentation

**Feature 6: GitLab CI Simplification (ci-modification)**
- Remove `create-version-tag` job from `.gitlab-ci.yml` (users push tags manually)
- Replace custom build jobs with single GoReleaser job
- Update default image to use custom CI Docker image
- Configure GoReleaser Docker-in-Docker integration
- Simplify pipeline stages: setup → lint → test → release → mirror

## Alternatives Considered

1. **Keep custom ldflags approach**: Could maintain existing implementation, but this would:
   - Require manual checksum creation
   - No SBOM generation
   - No package manager support
   - Complex shell script maintenance
   - Not industry standard

2. **Use GitHub Actions over GitLab CI**: Could migrate, but this would:
   - Require changing CI platform entirely
   - Not leverage existing GitLab infrastructure
   - Lose GitLab Container Registry integration
   - Team already uses GitLab

3. **Skip custom CI image**: Could install mise in each job, but this would:
   - Add installation time to every job (30-60s)
   - Slower CI runs
   - Network dependency on mise.run
   - Less reproducible

4. **Bootstrap script approach**: Could commit mise binary to repo, but this would:
   - Increase repo size (10-15MB binary)
   - Require version control updates
   - Not leverage Docker caching
   - Longer CI setup time

5. **Manual GPG signing setup**: Could implement immediately, but this would:
   - Add complexity to initial implementation
   - Require GPG key generation and management
   - Add CI variable setup
   - Defer until basic release flow works

## Impact

**Affected Specs**:
- Add new capability: release-management
- Add new capability: ci-image
- Modify existing capability: mise-tools (add GoReleaser, update scripts)
- Modify existing capability: cli-framework (enhanced version command)

**Affected Code**:
- Create `.goreleaser.yml` (GoReleaser configuration)
- Create `Dockerfile.ci` (CI Docker image)
- Create `.mise/tasks/tools/check.sh` (tool update checker)
- Create `.mise/tasks/tools/update.sh` (tool version updater)
- Create `install.sh` (one-line user install script)
- Create `INSTALL.md` (installation guide)
- Modify `internal/version/version.go` (const → var, add Commit/Date)
- Modify `cmd/version.go` (enhanced version display)
- Modify `.mise/config.toml` (add GoReleaser, tasks)
- Modify `.gitlab-ci.yml` (simplify, use GoReleaser, use CI image)
- Remove: GitLab CI `create-version-tag` job
- Remove: `scripts/` directory (all scripts moved to `.mise/tasks/`)

**Positive Impacts**:
- Professional-grade release artifacts (archives, checksums, SBOMs)
- Automated cross-platform builds
- Eliminates technical debt in CI scripts
- Industry-standard approach (GoReleaser)
- Faster CI runs with pre-installed tools
- Easy tool version management via mise
- Reproducible builds with mod_timestamp
- One-line installation for users
- Better security (SBOMs, optional GPG signing)

**Neutral Impacts**:
- Manual tag creation instead of automatic (more control)
- Requires Docker image maintenance (infrastructure trade-off)
- Requires initial learning of GoReleaser config

**No Negative Impacts**

## Dependencies

None - this is independent infrastructure improvement.

## Success Criteria

1. GoReleaser builds all 5 platform binaries successfully
2. Archives (.tar.gz, .zip) are created correctly
3. SHA256 checksums are generated for all artifacts
4. SBOMs are generated in SPDX format
5. GitLab releases are created on tag push with:
   - All 5 binary archives
   - Checksums file
   - SBOM file
   - Auto-generated release notes
6. Custom CI Docker image builds and pushes successfully
7. GitLab CI runs faster with pre-installed tools
8. Version command displays full version info
9. Install script works for all platforms
10. Tool version check and update scripts work correctly

## Validation Plan

- Run `mise run release:check` to validate GoReleaser configuration
- Run `mise run release:dry-run` to test full release pipeline locally (requires clean git state and tag)
- Build Docker image: `docker build -t test -f Dockerfile.ci .`
- Test Docker image: `docker run --rm test mise run check`
- Build and push Docker image: `mise run build-ci-image` (or let CI auto-build on changes)
- Test GitLab CI on main branch (should NOT trigger release)
- Create test tag and verify full pipeline
- Verify release artifacts on GitLab
- Test install script locally
- Run `mise run check` for final quality check

## Related Changes

This is Release Management Milestone - infrastructure improvement for professional releases.

## Timeline Estimate

1.0-1.5 days for implementation, testing, and documentation.

## Open Questions

None - scope is clear and based on industry best practices.
