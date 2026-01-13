# Design: Professional Release Management

## Context

The current codebase uses manual version management via hardcoded constants in `internal/version/version.go` and custom GitLab CI with ldflags injection for builds. This approach lacks professional release artifacts, requires maintaining complex CI scripts, and doesn't follow industry standards.

### Constraints
- Follow existing mise-based tool management pattern
- Use GoReleaser (industry standard for Go CLI releases)
- Custom Docker image for CI with pre-installed tools
- Maintain GitLab CI infrastructure
- Minimal learning curve for team

## Goals / Non-Goals

**Goals**:
- Automated cross-platform release builds (Linux, macOS, Windows)
- Professional release artifacts (archives, checksums, SBOMs)
- Easy tool version management via mise
- Faster CI runs with pre-installed tools
- One-line installation for users
- Reproducible builds

**Non-Goals**:
- GPG signing in initial implementation (deferred)
- Package manager integration (Homebrew, Scoop, etc. - deferred)
- UPX compression (deferred)
- macOS code signing/notarization (deferred)
- Multiple CI/CD platforms (GitLab only)
- Automated tag creation (manual push)

## Decisions

### Decision 1: GoReleaser Over Custom CI

**What**: Use GoReleaser for all release builds instead of custom GitLab CI ldflags injection.

**Why**:
- Industry standard (Google, Microsoft, AWS, HashiCorp, Kubernetes)
- Single configuration file handles everything
- Automatic checksums, SBOMs, archives
- Auto-generated release notes from commits
- Supports reproducible builds
- Package manager integration (future-proof)
- Eliminates technical debt

**Alternatives considered**:
- Custom ldflags in GitLab CI: Manual maintenance, no SBOMs, no package managers
- GitHub Actions over GitLab CI: Would require changing entire CI platform
- Makefiles: Shell script maintenance, no built-in features

**Trade-offs**:
- **Pro**: Industry standard, comprehensive features, no script maintenance
- **Con**: Initial learning curve for GoReleaser config
- **Mitigation**: GoReleaser is well-documented, config is simple YAML

---

### Decision 2: Custom Docker Image with mise

**What**: Create custom Docker image (`Dockerfile.ci`) with Go 1.25.5, mise, and all tools pre-installed. Host on GitLab Container Registry.

**Why**:
- Faster CI runs (tools pre-installed, no installation time)
- Consistent environment (dev and CI use same versions)
- Better caching (tools are in image, not in CI cache)
- Version controlled (mise config baked into image)
- GitLab Container Registry integration

**Alternatives considered**:
- Bootstrap script in each job: Slower, network dependency, less reproducible
- Install mise via curl in before_script: 30-60s per job, network dependency
- Use official mise image: Would need to add Go and golangci-lint anyway
- Commit mise binary to repo: Increases repo size 10-15MB, version control overhead

**Trade-offs**:
- **Pro**: Fast CI, consistent env, version controlled, good caching
- **Con**: Requires Docker image maintenance (rebuild on tool updates)
- **Mitigation**: GitLab CI job automatically rebuilds image when Dockerfile.ci or .mise/config.toml changes, clear documentation
**Dockerfile structure**:
```dockerfile
FROM golang:1.25.5-bookworm AS base
# Install mise
ENV MISE_DATA_DIR=/mise
ENV MISE_CONFIG_DIR=/mise
ENV MISE_CACHE_DIR=/mise/cache
ENV PATH="/mise/shims:$PATH"
# Copy .mise/config.toml
# mise install --yes
FROM tools AS final
```

**Image naming**: `registry.gitlab.com/amoconst/germinator/ci:latest` and `:v2026.1.2` (mise version)

---

### Decision 3: Pinned Tool Versions in mise

**What**: Pin golangci-lint to specific version (1.60.1) and GoReleaser to specific version (2.4.0) in `.mise/config.toml`. Provide update scripts.

**Why**:
- Reproducible builds (same versions across dev and CI)
- Clear upgrade path (scripts to check and update)
- Version control for tool versions
- Easy to see when updates are available

**Alternatives considered**:
- Use "latest" for all tools: Unpredictable, breaks CI when versions change
- Environment variables: Not version controlled, harder to manage

**Trade-offs**:
- **Pro**: Reproducible, version controlled, clear upgrades
- **Con**: Manual updates required
- **Mitigation**: Simple scripts (`.mise/tasks/tools/check.sh`, `.mise/tasks/tools/update.sh`)

**Update workflow**:
1. Run `mise run tools:check` to see available updates
2. Run `mise run tools:update` to update `.mise/config.toml`
3. Commit and push
4. GitLab CI will automatically rebuild the CI image when `.mise/config.toml` changes

---

### Decision 4: Manual Tag Creation

**What**: Remove automatic GitLab CI tag creation. Users manually create tags: `git tag v0.3.0 && git push origin v0.3.0`.

**Why**:
- More control over when releases happen
- Prevents accidental releases from main branch pushes
- Standard practice in open source
- Simple and explicit

**Alternatives considered**:
- Keep automatic tag creation: Could trigger releases unexpectedly
- Tag on every commit to main: Too many releases, confusing
- Semantic release automation: Overkill for simple versioning

**Trade-offs**:
- **Pro**: Control, explicit, standard practice
- **Con**: Manual step required (1-2 commands)
- **Mitigation**: Documented workflow, simple commands

---

### Decision 5: Version Injection via ldflags

**What**: Change `internal/version/version.go` from `const Version` to `var Version`. Add `Commit` and `Date` variables. Inject via GoReleaser ldflags.

**Why**:
- GoReleaser standard approach
- Simple and direct
- Works for both local and CI builds
- Clear separation (version package, CLI display)

**Alternatives considered**:
- Build-time .version file: Requires extra file, less standard
- Environment variables: Not version controlled
- Compile-time code generation: Overkill

**Trade-offs**:
- **Pro**: Standard, simple, works everywhere
- **Con**: `var` instead of `const` (minor)
- **Mitigation**: GoReleaser always injects at build time

**Variable defaults**:
```go
var (
    Version = "0.2.0"  // Fallback for local builds
    Commit  = "dev"    // Fallback for local builds
    Date    = "unknown" // Fallback for local builds
)
```

**Version bump script update**:
- `.mise/tasks/version.sh` now updates all three variables (Version, Commit, Date)
- Version is bumped based on patch/minor/major argument
- Commit is set to current git commit SHA (7 characters)
- Date is set to current date (YYYY-MM-DD format)
- During GoReleaser release, Commit and Date will be overwritten with build-time values


---

### Decision 6: Simplified GitLab CI Pipeline

**What**: Replace complex build matrix with single GoReleaser job. Remove `create-version-tag` job. Use custom CI image.

**Why**:
- GoReleaser handles all cross-platform builds
- Simpler configuration (1 job vs. 6 jobs)
- Faster CI (Docker-in-Docker, parallel builds)
- Clearer pipeline stages
- Less maintenance

**Alternatives considered**:
- Keep build matrix jobs: More complex, GoReleaser does it internally
- Keep tag creation job: Manual tags are better control
- Use different image per stage: Inconsistent env, slower

**Trade-offs**:
- **Pro**: Simple, fast, maintainable
- **Con**: Single point of failure (GoReleaser job)
- **Mitigation**: GoReleaser is battle-tested, mature

**Pipeline stages**:
```
setup → lint → test → release → mirror
```

**Release job only runs on tags**: `if: '$CI_COMMIT_TAG'`

---

### Decision 7: Archive Formats (tar.gz + zip)

**What**: Generate .tar.gz for Linux/macOS, .zip for Windows.

**Why**:
- Standard practice (most tools provide both)
- Windows users prefer .zip
- Unix users prefer .tar.gz
- GoReleaser supports format overrides

**Alternatives considered**:
- Only .tar.gz: Windows users unfamiliar
- Only .zip: Unix users prefer .tar.gz
- Single format for all: Not standard

**Trade-offs**:
- **Pro**: Standard, familiar to users, GoReleaser support
- **Con**: Two artifacts per platform
- **Mitigation**: Clear naming, users choose appropriate format

---

### Decision 8: SBOM Generation

**What**: Generate SBOMs using Syft in SPDX format for all release artifacts.

**Why**:
- Required by enterprises in 2025
- GoReleaser has native Syft integration
- Provides dependency transparency
- Security best practice
- Zero configuration needed

**Alternatives considered**:
- Skip SBOMs: Not enterprise-ready
- Custom SBOM generation: Reinvents wheel
- CycloneDX instead of SPDX: Less common

**Trade-offs**:
- **Pro**: Enterprise-ready, security best practice, automatic
- **Con**: Extra artifact file
- **Mitigation**: SBOM is small, optional to download

**SBOM naming**: `germinator_0.3.0_sbom.spdx.json`

---

### Decision 9: Auto-Generated Release Notes

**What**: Generate release notes from commit history using GoReleaser's changelog feature. Filter out docs, test, ci, chore, build commits.

**Why**:
- Automatic, no manual entry
- Clear format (one line per commit)
- Filter out noise (internal commits)
- Standard practice (Helm, GoReleaser itself)

**Alternatives considered**:
- Manual release notes: Time-consuming, easy to forget
- Custom changelog generation: GoReleaser does it
- Include all commits: Too noisy

**Trade-offs**:
- **Pro**: Automatic, filtered, standard
- **Con**: Less control over format
- **Mitigation**: Clear filtering rules, manual override if needed

**Filter rules**:
```yaml
filters:
  exclude:
    - "^docs:"
    - "^test:"
    - "^ci:"
    - "^chore:"
    - "^build:"
    - Merge pull request
    - Merge branch
```

---

### Decision 10: One-Line Install Script

**What**: Create `install.sh` at project root for end-user curl-based installation: `curl -sSL https://gitlab.com/amoconst/germinator/-/raw/main/install.sh | bash`.

**Why**:
- Standard practice (Homebrew, Helm, kubectl, etc.)
- Easy for users
- Cross-platform
- Automatic platform/arch detection

**Alternatives considered**:
- No install script: Manual download for each platform
- Package manager only: More complex, deferred
- Multi-step instructions: Higher friction

**Trade-offs**:
- **Pro**: Standard, easy, cross-platform
- **Con**: Security concern (curl | bash)
- **Mitigation**: Document checksum verification, users can inspect script

---

## Risks / Trade-offs

### Risk 1: Docker Image Maintenance Overhead

**Risk**: Custom Docker image requires rebuilding and pushing when tools update.

**Mitigation**: GitLab CI job automatically rebuilds image when needed, clear documentation.

**Acceptance**: Low risk, high benefit (faster CI, consistent env).

---

### Risk 2: GoReleaser Learning Curve

**Risk**: Team needs to learn GoReleaser configuration and workflow.

**Mitigation**: Well-documented, simple YAML config, examples from GoReleaser docs and Helm project.

**Acceptance**: Low risk, one-time learning effort.

---

### Risk 3: Manual Tag Creation

**Risk**: Users might forget to create tags, delaying releases.

**Mitigation**: Documented workflow, simple commands, GitLab CI shows release status clearly.

**Acceptance**: Low risk, standard practice in open source.

---

### Risk 4: Security of Install Script

**Risk**: `curl | bash` pattern raises security concerns (MITM, script changes).

**Mitigation**: Document checksum verification, users can inspect script before running, use HTTPS, GitLab is trusted.

**Acceptance**: Acceptable risk (standard practice, documented verification).

---

## Migration Plan

### Phase 1: Core Changes
1. Update `internal/version/version.go` (const → var, add Commit/Date)
2. Update `cmd/version.go` (enhanced display)
3. Update `.mise/tasks/version.sh` (handle var format, update Commit/Date)
4. Create `.goreleaser.yml` configuration
5. Test local GoReleaser builds

### Phase 2: mise Integration
1. Update `.mise/config.toml` (add GoReleaser, tasks, scripts)
2. Create `.mise/tasks/tools/check.sh`
3. Create `.mise/tasks/tools/update.sh`
4. Test tool update scripts

### Phase 3: CI Image
1. Create `Dockerfile.ci`
2. Add `build-ci-image` job to `.gitlab-ci.yml`
3. Build and test image locally
4. Push to GitLab Container Registry

### Phase 4: GitLab CI Update
1. Update `.gitlab-ci.yml` (simplify, use GoReleaser, use CI image)
2. Test pipeline on main branch
3. Test pipeline with merge request
4. Create test tag, verify release

### Phase 5: Documentation
1. Create `INSTALL.md`
2. Create `.mise/tasks/install.sh`
3. Test install script locally
4. Update README.md with installation link

### Rollback Plan
If any phase introduces issues:
- Revert to previous working state via git
- Phase isolation allows targeted rollback
- Can use custom ldflags approach temporarily if GoReleaser fails

---

## Edge Case Handling

### No Git Repository

**Scenario**: Build runs in directory without git history or shallow clone.

**Handling**:
- Use default values from `internal/version/version.go`
- Version: "0.2.0", Commit: "dev", Date: "unknown"
- Version command displays defaults
- GoReleaser fails gracefully on missing git info
- Documentation: Note that git is required for version injection

**Implementation**: No special handling needed - GoReleaser and git describe default gracefully

---

### Tag Conflicts

**Scenario**: Tag already exists but points to different commit.

**Handling**:
- Git tag force not allowed (standard git safety)
- User must delete old tag: `git tag -d v0.3.0 && git push origin :refs/tags/v0.3.0`
- Then recreate and push new tag
- Documentation: Include tag deletion workflow in release process

**Implementation**: Manual process documented in release workflow

---

### Partial Release Failure

**Scenario**: GoReleaser creates GitLab release but fails on some artifacts.

**Handling**:
- Release exists on GitLab with partial artifacts
- Manual cleanup required: Delete GitLab release via UI or API
- Re-run CI job: Retry button in GitLab UI or push new tag
- Monitoring: Check job logs to identify failure point

**Implementation**: Document cleanup process in troubleshooting section

---

### Tool Installation Failure

**Scenario**: mise fails to install golangci-lint or GoReleaser during CI image build.

**Handling**:
- Docker build fails immediately
- Error message shows which tool failed
- Fix: Update `.mise/config.toml` with correct version
- GitLab CI will automatically rebuild the image on push

**Implementation**: Clear error messages, automatic CI rebuild

---

### GitHub API Rate Limiting

**Scenario**: Tool update scripts hit GitHub API rate limits.

**Handling**:
- Scripts use unauthenticated API calls (60 requests/hour limit)
- Minimal queries (2 tools)
- Error message: "API rate limit exceeded, wait 1 hour and retry"
- Alternative: Use `latest` versions or manually specify

**Implementation**: Rate limit check, retry logic with backoff

---

## Open Questions

None - design is clear and based on industry best practices.
