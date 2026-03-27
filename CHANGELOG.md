# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.0] - 2026-03-27

### Added

- Extract `Parser` and `Serializer` interfaces from infrastructure layer for testable services with dependency injection via constructor (extract-infrastructure-interfaces)

### Changed

- Refactor release workflow with shared library (`release-lib.sh`) and phase-based tasks: `release:check` for validation-only, `release:prepare` for validation + preview (refactor-release-workflow)

## [0.7.0] - 2026-03-20

### Added

- Add shell completion via carapace with dynamic, library-aware suggestions for bash, zsh, fish, and powershell with path resolution and caching (add-carapace-completion)
- Add testify/mock implementations for Transformer, Validator, Canonicalizer, and Initializer interfaces with mock generation and test helper patterns (mock-infrastructure)
- Add MIT license file to the project root

### Changed

- Migrate CLI commands to RunE pattern with centralized error handling and expanded semantic exit codes (cli-rune-migration)
- Expand golangci-lint from 8 to 25+ linters with depguard domain purity enforcement, complexity thresholds, and comprehensive test exclusions (comprehensive-linting)
- Reorganize domain layer moving models, errors, and validation into `internal/domain/` package following DDD-light principles (domain-restructure)
- Reorganize infrastructure packages under `internal/infrastructure/` with unified structure: parsing, serialization, adapters, config, and library subpackages (infrastructure-restructure)
- Update Go to 1.26.1 and openspec-extended to 0.18.1 for improved tooling and automation

### Fixed

- Suppress false-positive gosec warnings for intentional file operations in CLI context

## [0.6.0] - 2026-03-11

### Added

- Add `canonicalize` command to convert platform documents (Claude Code, OpenCode) to canonical YAML format for reverse transformation workflows (reverse-transformation)
- Add `init` command for batch transformation and installation of library resources to projects (add-library-init-system)
- Add `library` command with `resources`, `presets`, and `show` subcommands for managing the canonical resource library (add-library-init-system)
- Add global `-v, --verbose` flag for increased output verbosity with multiple escalation levels (cli-infrastructure)
- Add global configuration system with Koanf-based TOML loading at XDG-compliant locations (`~/.config/germinator/config.toml`) (add-configuration-system)
- Add dependency injection pattern with ServiceContainer for cleaner command architecture and testability (di-foundation)
- Introduce service interfaces in `internal/application/` for Transformer, Validator, Canonicalizer, and Initializer with request/response types (introduce-service-interfaces)

### Changed

- Add functional validation pipeline with `Result[T]` pattern for composable, early-exit validation with rich error aggregation (add-validation-pipeline)
- Enhance all error types with immutable builder pattern supporting `WithSuggestions()`, `WithContext()`, and `WithDetails()` for progressive error enrichment (enhance-all-errors)
- Refactor CLI commands for consistency across adapt, validate, canonicalize, init, and library using RunE pattern with centralized error handling (cli-infrastructure)
- Expand E2E test coverage for all CLI commands and platform adapters using Ginkgo v2 with parallel execution support (e2e-test-coverage, e2e-testing-ginkgo)

## [0.5.0] - 2026-02-09

### Added

- Add OpenCode as a target platform with canonical source format, platform-agnostic models, and comprehensive OpenCode templates and validation (add-opencode-platform)
- Add platform adapters for bidirectional conversion between canonical format and Claude Code documents (canonical-format-redesign)
- Add retrieval-led reasoning guidance for improved AI agent document handling (canonical-format-redesign)

### Changed

- Redesign canonical format to be domain-driven with `permissionPolicy` enum (allow/deny/require), `behavior` objects for action configuration, and `targets` section for platform-specific overrides (canonical-format-redesign)
- Refactor adapters to use canonical models with unified `Steps` field across all platform formats (canonical-format-redesign)

## [0.4.0] - 2026-02-03

### Added

- Add OpenCode as a target platform with templates, validation functions, permission transformation, and tool name case conversion (PascalCase to lowercase) (add-opencode-platform)
- Add golden file test suite with `UPDATE_GOLDEN` environment variable for simple test regeneration (add-opencode-platform)
- Add OpenSpec concepts skill for AI agents to understand spec-driven development workflow (add-opencode-platform)
- Add installation documentation with curl-based install script supporting Linux and macOS (add-opencode-platform)
- Add teaching instructions to memory template for improved AI guidance (add-opencode-platform)
- Add `pre-commit` to mise for automated validation hooks
- Add version bump enforcement, GoReleaser dry-run validation, and git tag serialization in release workflow
- Add hash-based CI image tagging for reliable cache invalidation

### Changed

- Migrate Docker CI image to Alpine Linux achieving 73% size reduction with improved build reliability and enhanced caching strategy using checksum-based approach (optimize-ci-infrastructure)
- Simplify CI workflow with better validation, automatic Git tag creation, hash-based CI image tagging, parallel job execution, and version bump enforcement (simplify-ci-workflow)
- Refactor documentation into hierarchical package-specific structure with AGENTS.md guides for each layer (add-opencode-platform)
- Rationalize mise tasks by removing duplicates and consolidating file-based tasks as source of truth
- Optimize Docker image build by checking for existing images before rebuilding
- Consolidate CI pipeline stages and optimize job execution for faster builds

### Fixed

- Correct OpenCode command tool support and field name inconsistencies (add-opencode-platform)
- Fix release hanging caused by duplicate SBOM filenames in GoReleaser output
- Fix CI job issues with entrypoint override and force-push strategy
- Fix mirror job dependency on optional tag creation job

## [0.3.0] - 2026-01-14

### Added

- Build foundational document models (Agent, Command, Memory, Skill) with YAML parsing, struct validation, and file loading infrastructure (add-core-infrastructure)
- Build minimal CLI infrastructure to enable core workflows: `validate` and `adapt` commands with template rendering pipeline (add-document-processing-infrastructure)
- Implement industry-standard release management using GoReleaser for automated cross-platform builds, checksums, SBOMs, and GitLab releases (implement-release-management)
- Add version command with enhancements for better version reporting

### Changed

- Move models to internal package structure for better encapsulation (add-document-processing-infrastructure)
- Implement version management system with `--version` flag and version command (add-document-processing-infrastructure)

## [0.2.0] - 2026-01-13

### Added

- Add `validate` command for AI coding assistant document validation with template rendering pipeline (add-document-processing-infrastructure)
- Add `adapt` command for document transformation between AI coding assistant platforms (add-document-processing-infrastructure)
- Add release infrastructure with GoReleaser integration for cross-platform binary builds, checksums, and custom Docker CI image (implement-release-management)
- Add installation documentation and curl-based install script (implement-release-management)

## [0.1.0] - 2026-01-13

### Added

- Establish the Go project structure with Cobra CLI framework, standard Go layout, and foundational configuration for the germinator CLI tool (initialize-project-structure)
- Create README documentation and minimal placeholder files for configuration and test directories (setup-configuration-structure)
- Configure minimal development tooling including golangci-lint for linting and mise task runner for validation and tool installation (setup-development-tooling)
- Build core document models (Agent, Command, Memory, Skill) with YAML parsing, struct validation, and file loading infrastructure (add-core-infrastructure)
