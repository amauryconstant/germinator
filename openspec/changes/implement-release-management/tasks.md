# Tasks for Implement Release Management

This document tracks all tasks for release management implementation, organized by commit.

---

## Commit 1: Version command enhancements (0.2.0 → 0.2.1)

### Code Changes
- [x] 1.1 Update `internal/version/version.go` - change Version from `const` to `var`
- [x] 1.2 Add `Commit` variable to `internal/version/version.go` (default "dev")
- [x] 1.3 Add `Date` variable to `internal/version/version.go` (default "unknown")
- [x] 1.4 Export all variables (Version, Commit, Date) in version package
- [x] 1.5 Update `cmd/version.go` - import version package
- [x] 1.6 Modify version command to display: `germinator {version} ({commit}) {date}`
- [ ] 1.7 Bump version to 0.2.1 in `internal/version/version.go`

### Documentation Updates
- [x] 1.8 Update this tasks.md file

### Testing & Validation
- [x] 1.9 Verify compilation: `go build ./internal/version/...`
- [x] 1.10 Verify version command output format
- [x] 1.11 Test version command locally: `./bin/germinator version`

---

## Commit 2: Release infrastructure (0.2.1 → 0.3.0)

### Code Changes
- [x] 2.1 Create `.goreleaser.yml` configuration file
- [x] 2.2 Configure builds section with 5 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- [x] 2.3 Configure ldflags for version, commit, date injection
- [x] 2.4 Configure archives (.tar.gz for Unix, .zip for Windows)
- [x] 2.5 Configure checksum generation (SHA256)
- [x] 2.6 Configure SBOM generation (Syft, SPDX)
- [x] 2.7 Configure changelog with filters (exclude docs, test, ci, chore, build, merges)
- [x] 2.8 Configure GitLab URLs for releases
- [x] 2.9 Create `Dockerfile.ci` base stage
- [x] 2.10 Install prerequisites (curl, git, ca-certificates, sudo) in Dockerfile
- [x] 2.11 Install mise (latest stable) via curl from mise.run
- [x] 2.12 Configure mise environment variables (MISE_DATA_DIR, MISE_CACHE_DIR, etc.)
- [x] 2.13 Create `Dockerfile.ci` tools stage
- [x] 2.14 Copy `.mise/config.toml` into tools stage
- [x] 2.15 Run `mise install --yes` in tools stage
- [x] 2.16 Create `Dockerfile.ci` final stage
- [x] 2.17 Set workspace directory and GOMODCACHE
- [x] 2.18 Update `.gitlab-ci.yml` - add `build-ci` stage
- [x] 2.19 Add `build-ci-image` job to `.gitlab-ci.yml`
- [x] 2.20 Configure job to run on Dockerfile.ci or .mise/config.toml changes
- [x] 2.21 Use docker:dind service for building
- [x] 2.22 Tag image with mise version from `.mise/config.toml`
- [x] 2.23 Push to GitLab Container Registry
- [x] 2.24 Update `.gitlab-ci.yml` default image to `registry.gitlab.com/amoconst/germinator/ci:latest`
- [x] 2.25 Remove `create-version-tag` job from pipeline
- [x] 2.26 Add `release` stage to stages list
- [x] 2.27 Create `release` job with GoReleaser Docker-in-Docker
- [x] 2.28 Configure release job to run only on tags (`if: '$CI_COMMIT_TAG'`)
- [x] 2.29 Configure release job with Docker service (`docker:dind`)
- [x] 2.30 Configure release job to use GITLAB_TOKEN (CI_JOB_TOKEN)
- [x] 2.31 Update pipeline rules to include tag pushes
- [x] 2.32 Add GoReleaser to `.mise/config.toml` [tools] section
- [x] 2.33 Create `release:dry-run` task in `.mise/config.toml`
- [x] 2.34 Create `release:check` task in `.mise/config.toml`
- [ ] 2.35 Bump version to 0.3.0 in `internal/version/version.go`
- [x] 2.36 Update `internal/core/parser.go` - change yaml import to alias
- [x] 2.37 Update `.gitignore` - add `dist/`, `artifacts.json`, `metadata.json`

### Documentation Updates
- [x] 2.38 Update this tasks.md file

### Testing & Validation
- [x] 2.39 Validate GoReleaser config: `goreleaser check`
- [x] 2.40 Verify tools work in Docker image (mise ls, go version, goreleaser --version, golangci-lint version)
- [x] 2.41 Test local Docker build: `docker build -t test -f Dockerfile.ci .`
- [x] 2.42 Verify YAML syntax: `cat .gitlab-ci.yml | mise exec -- yamllint`
- [x] 2.43 Verify TOML syntax: `cat .mise/config.toml | mise exec -- toml2json`
- [x] 2.44 Run `go mod tidy`
- [x] 2.45 Run `go vet ./...`
- [x] 2.46 Run `golangci-lint run`
- [x] 2.47 Run `mise run format`

---

## Commit 3: Tool management (0.3.0 → 0.3.1)

### Code Changes
- [ ] 3.1 Pin golangci-lint version in `.mise/config.toml` (currently 2.8.0)
- [ ] 3.2 Add `build:local` task in `.mise/config.toml`
- [ ] 3.3 Update `build` task to depend on `build:clean`
- [ ] 3.4 Update `build:clean` to include `dist/` directory
- [ ] 3.5 Reorganize tasks section in `.mise/config.toml`
- [ ] 3.6 Update `check` task to depend on lint, format, test, build
- [ ] 3.7 Remove standalone check task (now depends on other tasks)
- [ ] 3.8 Remove standalone format script (now inline in config)
- [ ] 3.9 Remove standalone test section header
- [ ] 3.10 Create `.mise/tasks/tools/check.sh` script
- [ ] 3.11 Implement GitHub API query for golangci-lint latest version
- [ ] 3.12 Implement GitHub API query for GoReleaser latest version
- [ ] 3.13 Implement version comparison logic
- [ ] 3.14 Implement formatted output (current vs latest)
- [ ] 3.15 Make check.sh executable: `chmod +x .mise/tasks/tools/check.sh`
- [ ] 3.16 Create `.mise/tasks/tools/update.sh` script
- [ ] 3.17 Implement GitHub API query for latest versions
- [ ] 3.18 Implement `.mise/config.toml` update logic (use sed for cross-platform compatibility)
- [ ] 3.19 Implement next steps documentation (git diff, mise install, commit)
- [ ] 3.20 Make update.sh executable: `chmod +x .mise/tasks/tools/update.sh`
- [ ] 3.21 Update `.mise/tasks/version.sh` - use grep instead of sed for version extraction
- [ ] 3.22 Update version.sh to also update Commit and Date fields
- [ ] 3.23 Update version.sh next steps to reference tag creation
- [ ] 3.24 Update `.golangci.yml` - use `disable-all: true` instead of `default: none`
- [ ] 3.25 Replace staticcheck with typecheck in linters
- [ ] 3.26 Remove gosec linter
- [ ] 3.27 Remove gosec settings section
- [ ] 3.28 Remove formatters section (use mise run format instead)
- [ ] 3.29 Delete `.mise/tasks/format.sh`
- [ ] 3.30 Delete `.mise/tasks/smoke-test.sh`
- [ ] 3.31 Delete `.mise/tasks/validate.sh`
- [ ] 3.32 Delete `scripts/.gitkeep`
- [ ] 3.33 Bump version to 0.3.1 in `internal/version/version.go`

### Documentation Updates
- [ ] 3.34 Update this tasks.md file

### Testing & Validation
- [ ] 3.35 Test tool check script: `mise run tools:check`
- [ ] 3.36 Test tool update script: `mise run tools:update`
- [ ] 3.37 Verify .mise/config.toml updates correctly after running update script
- [ ] 3.38 Verify tools install: `mise install --yes`
- [ ] 3.39 Run `go build ./...`
- [ ] 3.40 Run `go vet ./...`
- [ ] 3.41 Run `golangci-lint run`
- [ ] 3.42 Run `mise run test`

---

## Commit 4: Installation and documentation (0.3.1 → 0.3.2)

### Code Changes
- [ ] 4.1 Create `install.sh` script at project root (for end users)
- [ ] 4.2 Implement OS detection (Linux, macOS, Windows)
- [ ] 4.3 Implement architecture detection (amd64, arm64)
- [ ] 4.4 Implement version detection (latest vs specific)
- [ ] 4.5 Implement binary download from GitLab releases
- [ ] 4.6 Implement archive extraction (tar.gz vs zip)
- [ ] 4.7 Implement installation to /usr/local/bin or ~/bin
- [ ] 4.8 Add executable permission to binary
- [ ] 4.9 Display installed version
- [ ] 4.10 Make install.sh executable: `chmod +x install.sh`
- [ ] 4.11 Bump version to 0.3.2 in `internal/version/version.go`

### Documentation Updates
- [ ] 4.12 Create `INSTALL.md` with installation instructions
- [ ] 4.13 Document quick install (curl | bash script)
- [ ] 4.14 Document manual download for Linux (amd64, arm64)
- [ ] 4.15 Document manual download for macOS (Intel, Apple Silicon)
- [ ] 4.16 Document manual download for Windows
- [ ] 4.17 Document checksum verification
- [ ] 4.18 Document GPG signature verification (optional)
- [ ] 4.19 Document version command for verification
- [ ] 4.20 Update `README.md` with quick install section
- [ ] 4.21 Add link to `INSTALL.md` in README
- [ ] 4.22 Update `AGENTS.md` - add Release Management section
- [ ] 4.23 Document tool update workflow in AGENTS.md
- [ ] 4.24 Document CI image maintenance in AGENTS.md
- [ ] 4.25 Document release process (version bump, tag, push) in AGENTS.md
- [ ] 4.26 Update AGENTS.md technology stack section
- [ ] 4.27 Update AGENTS.md CI/CD section
- [ ] 4.28 Update this tasks.md file

### Testing & Validation
- [ ] 4.29 Test install script locally: `bash install.sh`
- [ ] 4.30 Test version detection in install script
- [ ] 4.31 Test installation to /usr/local/bin or ~/bin
- [ ] 4.32 Verify code formatting: `mise run format`
- [ ] 4.33 Review INSTALL.md for clarity and completeness
- [ ] 4.34 Verify README.md mentions installation methods correctly
- [ ] 4.35 Run `mise run check` - verify all validation checks pass
- [ ] 4.36 Create test release tag: `git tag v0.3.2-test`
- [ ] 4.37 Push main branch and test tag
- [ ] 4.38 Monitor full release pipeline
- [ ] 4.39 Verify GitLab release with all artifacts
- [ ] 4.40 Download and test binary on Linux
- [ ] 4.41 Download and test binary on macOS (if possible)
- [ ] 4.42 Verify version command shows correct info
- [ ] 4.43 Verify SBOM is valid JSON
- [ ] 4.44 Verify checksums match binaries
- [ ] 4.45 Test install script with release URL
- [ ] 4.46 Clean up test tag: `git push origin :refs/tags/v0.3.2-test`

---

## Notes

- All version bumps are included in their respective commits
- Testing sections include only validation relevant to that commit's changes
- Commit 4 includes final integration testing after all implementation is complete
- After completing all commits, tag and push v0.3.2 to create production release
