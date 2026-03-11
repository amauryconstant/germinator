# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [0.6.0] - 2026-03-11

### Added

- Add `canonicalize` command to convert platform documents (Claude Code, OpenCode) to canonical YAML format (reverse-transformation)
- Add `init` command for batch transformation and installation of library resources to projects (add-library-init-system)
- Add `library` command with `resources`, `presets`, and `show` subcommands for managing the canonical resource library (add-library-init-system)
- Add global `-v, --verbose` flag for increased output verbosity (cli-infrastructure)
- Add global configuration system with Koanf-based TOML loading at XDG-compliant location (add-configuration-system)
- Add dependency injection pattern with ServiceContainer for cleaner command architecture (di-foundation)
- Introduce service interfaces in `internal/application/` for testability and separation of concerns (introduce-service-interfaces)

### Changed

- Add functional validation pipeline with `Result[T]` pattern for composable, early-exit validation (add-validation-pipeline)
- Enhance error types with immutable builder pattern: `WithSuggestions()`, `WithContext()` (enhance-all-errors)
- Refactor CLI commands for consistency across adapt, validate, canonicalize, init, and library (cli-infrastructure)
- Expand E2E test coverage for all CLI commands and platform adapters (e2e-test-coverage, e2e-testing-ginkgo)

## [0.5.0] - 2026-02-09

### Added

- Add OpenCode as a target platform with canonical source format, platform-agnostic models, and OpenCode templates and validation (add-opencode-platform)

## [0.4.1] - 2026-02-04

### Added

- Add installation documentation and curl-based install script
- Add OpenSpec concepts skill for AI agents

### Changed

- Refactor documentation into hierarchical package-specific structure

### Fixed

- Correct OpenCode command tool support and field name

## [0.4.0] - 2026-02-03

### Changed

- Migrate Docker image to Alpine with 73% size reduction, improve build reliability, and enhance cache strategy (optimize-ci-infrastructure)
- Simplify CI workflow with better validation, error handling, automated release tagging, and reliable CI image rebuilding (simplify-ci-workflow)

## [0.3.2] - 2026-01-14

### Added

- Add installation documentation and curl-based install script

## [0.3.0] - 2026-01-14

### Added

- Build foundational components: document models, YAML parsing, struct validation, and file loading (add-core-infrastructure)
- Build minimal infrastructure to enable core workflows: validate and adapt AI coding assistant documents (add-document-processing-infrastructure)
- Implement industry-standard release management using GoReleaser for automated cross-platform builds, checksums, SBOMs, and GitLab releases (implement-release-management)

## [0.2.0] - 2026-01-13

### Changed

- Move models to internal package
- Implement version management system

## [0.1.0] - 2026-01-13

### Added

- Establish the Go project structure, tooling, and foundational configuration for the germinator CLI tool (initialize-project-structure)
- Create README documentation and minimal placeholder files for configuration and test directories (setup-configuration-structure)
- Configure minimal development tooling including golangci-lint and mise task runner for validation and tool installation (setup-development-tooling)
