# release-management Specification

## Purpose
TBD - created by archiving change implement-release-management. Update Purpose after archive.
## Requirements
### Requirement: GoReleaser Configuration

The project SHALL have a GoReleaser configuration file for automated release management.

#### Scenario: GoReleaser configuration exists
**Given** release management is implemented
**When** a developer inspects the project root
**Then** `.goreleaser.yml` SHALL exist
**And** it SHALL be valid YAML syntax
**And** it SHALL configure cross-platform builds

#### Scenario: Cross-platform builds configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL specify 5 build targets:
**And** it SHALL include linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
**And** it SHALL use `CGO_ENABLED=0` for static linking

#### Scenario: Version injection configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure ldflags for version injection
**And** it SHALL inject `{{.Version}}` into `gitlab.com/amoconst/germinator/internal/version.Version`
**And** it SHALL inject `{{.Commit}}` into `gitlab.com/amoconst/germinator/internal/version.Commit`
**And** it SHALL inject `{{.Date}}` into `gitlab.com/amoconst/germinator/internal/version.Date`

#### Scenario: Archive generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL generate `.tar.gz` archives for Linux/macOS
**And** it SHALL generate `.zip` archives for Windows
**And** it SHALL include LICENSE and README.md in archives

#### Scenario: Checksum generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure checksum generation
**And** it SHALL use SHA256 algorithm
**And** it SHALL name checksum file as `checksums.txt`

#### Scenario: SBOM generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure SBOM generation using Syft
**And** it SHALL use SPDX format
**And** it SHALL name SBOM file as `germinator_{{.Version}}_sbom.spdx.json`

#### Scenario: Changelog generation configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL configure auto-generated release notes
**And** it SHALL filter out docs, test, ci, chore, build commits
**And** it SHALL filter out merge commits

---

### Requirement: Automated Release Builds

The system SHALL build cross-platform release artifacts automatically when a Git tag is pushed.

#### Scenario: Build all platform binaries
**Given** a Git tag is pushed (e.g., v0.3.0)
**When** GitLab CI release job runs
**Then** GoReleaser SHALL build 5 binaries:
**And** it SHALL build `germinator_0.3.0_linux_amd64`
**And** it SHALL build `germinator_0.3.0_linux_arm64`
**And** it SHALL build `germinator_0.3.0_darwin_amd64`
**And** it SHALL build `germinator_0.3.0_darwin_arm64`
**And** it SHALL build `germinator_0.3.0_windows_amd64.exe`

#### Scenario: Generate archives
**Given** GoReleaser builds binaries
**When** archives are generated
**Then** it SHALL create `.tar.gz` archives for Linux/macOS platforms
**And** it SHALL create `.zip` archive for Windows platform
**And** archives SHALL contain binary, LICENSE, README.md

#### Scenario: Generate checksums
**Given** GoReleaser creates release artifacts
**When** checksums are generated
**Then** it SHALL create `checksums.txt`
**And** it SHALL contain SHA256 hashes for all 5 archives
**And** checksums.txt SHALL include SBOM file

#### Scenario: Generate SBOMs
**Given** GoReleaser creates release artifacts
**When** SBOMs are generated
**Then** it SHALL create `germinator_0.3.0_sbom.spdx.json`
**And** it SHALL list all Go dependencies
**And** it SHALL include metadata (version, commit, date)

#### Scenario: Create GitLab Release
**Given** GoReleaser builds all artifacts
**When** release job completes
**Then** it SHALL create a GitLab release with tag name as release name
**And** it SHALL attach all 5 binary archives
**And** it SHALL attach checksums file
**And** it SHALL attach SBOM file
**And** it SHALL include auto-generated release notes

#### Scenario: Release notes from commits
**Given** GitLab release is created
**When** release notes are generated
**Then** it SHALL show one commit message per line
**And** it SHALL exclude commits starting with docs:, test:, ci:, chore:, build:
**And** it SHALL exclude merge commits

---

### Requirement: Local Development Builds

The system SHALL support local development builds with version information.

#### Scenario: Build snapshot locally
**Given** a developer runs `mise run build`
**When** simple build runs
**Then** it SHALL create binary in `bin/` directory
**And** version SHALL be from `internal/version/version.go`

#### Scenario: Quick local build without GoReleaser
**Given** a developer runs `mise run build:local`
**When** quick build runs
**Then** it SHALL build single binary in `bin/germinator`
**And** version SHALL be derived from git describe
**And** commit SHA SHALL be from current HEAD
**And** date SHALL be current date

---

### Requirement: Installation Support

The project SHALL provide one-line installation script and comprehensive installation guide.

#### Scenario: Install script for end users
**Given** installation is documented
**When** a user inspects project root
**Then** `install.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Install script detects platform
**Given** a user runs install script
**When** script executes
**Then** it SHALL detect OS (Linux/macOS/Windows)
**And** it SHALL detect architecture (amd64/arm64)
**And** it SHALL select appropriate binary

#### Scenario: Install script downloads and installs
**Given** a user runs install script
**When** script executes
**Then** it SHALL download binary from GitLab releases
**And** it SHALL extract archive
**And** it SHALL install to `/usr/local/bin/` or `~/bin/`
**And** it SHALL display installed version

#### Scenario: Local install task for developers
**Given** installation is documented
**When** a developer inspects `.mise/tasks/` directory
**Then** `.mise/tasks/install-local.sh` SHALL exist
**And** it SHALL be executable

#### Scenario: Local install task
**Given** a developer runs `mise run install:local`
**When** task executes
**Then** it SHALL build germinator locally
**And** it SHALL install to `/usr/local/bin/` or `~/.local/bin/`
**And** it SHALL display installation verification message

#### Scenario: Installation guide exists
**Given** installation is documented
**When** a user reads documentation
**Then** `INSTALL.md` SHALL exist
**And** it SHALL document quick install (curl | bash)
**And** it SHALL document manual download for all platforms
**And** it SHALL document checksum verification
**And** it SHALL document GPG signature verification (optional)

---

### Requirement: Reproducible Builds

The system SHALL support reproducible builds via GoReleaser configuration.

#### Scenario: Build reproducibility configured
**Given** `.goreleaser.yml` exists
**When** configuration is inspected
**Then** it SHALL use `-trimpath` flags
**And** it SHALL set `mod_timestamp: "{{ .CommitTimestamp }}"`
**And** it SHALL set `gcflags: all=-trimpath={{.Env.GOPATH}}`

#### Scenario: Same commit produces same binary
**Given** same GoReleaser config and same commit
**When** binary is built twice
**Then** binaries SHALL have identical checksums
**And** build date SHALL use commit date, not build date

---

### Requirement: Dry-Run Support

The system SHALL support testing release configuration without creating actual releases.

#### Scenario: Validate GoReleaser config
**Given** a developer runs `mise run release:check`
**When** configuration is validated
**Then** GoReleaser SHALL check configuration syntax
**And** it SHALL report any errors or deprecated properties
**And** it SHALL exit with 0 if valid (works with uncommitted changes)

#### Scenario: Dry-run release
**Given** a developer has a clean git state
**And** a developer has a git tag checked out (e.g., v0.3.0)
**When** developer runs `mise run release:dry-run`
**Then** GoReleaser SHALL build all artifacts
**And** it SHALL skip publishing to GitLab
**And** it SHALL create artifacts locally in `dist/` directory
**And** it SHALL use version from git tag
**And** it SHALL display what would be released

#### Scenario: Windows binary testing is out of scope
**Given** GoReleaser generates Windows binary
**When** Windows binary is created
**Then** cross-compilation SHALL work (GOOS=windows GOARCH=amd64)
**And** Windows testing is out of scope for initial implementation
**And** documentation SHALL note that Windows testing is planned for future iteration

