# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-07-17


Two-and-a-half months of work since `v0.9.0`: seven new library subcommands and flags, a full migration of every command to the Functional Core / Imperative Shell architecture, plus a hardening pass that consolidates typed errors, removes the legacy exit-code canary, and extracts I/O adapters into dedicated shell packages. See the **Breaking** section for upgrade guidance.

### Added

**Library features (April 2026)**

- Add `library refresh` to re-sync `description`/`path` metadata from resource files into `library.yaml` (skips on name conflict, never deletes entries) (library-refresh-and-discover)
- Add `library validate` to audit `library.yaml` against the filesystem across four issue types (missing file, ghost preset ref, orphaned file, malformed frontmatter) with `--fix` for safe auto-cleanup (library-validate)
- Add `library remove resource <ref>` (deletes file + YAML entry; errors if a preset references it) and `library remove preset <name>` (YAML-only) (library-remove)
- Add `--discover` to `library add` for orphan registration across `skills/`, `agents/`, `commands/`, `memory/`; `--discover` later extended to recurse into subdirectories and return a structured `DiscoverResult` (library-refresh-and-discover, discover-orphans)
- Add `--batch` to `library add` to accept multiple files/directories (recursively scanned for `.md`), continuing on per-item errors and emitting an aggregate summary (`Added N, skipped M, failed K`); batch exit code is always 0 (partial success = success) (batch-add)
- Add `--json` flag on the parent `library` command, inheriting to every subcommand (`resources`, `presets`, `show`, `add`, `init`, `remove`, `validate`, `refresh`) with consistent JSON schemas (library-json-flag)
- Add partial-processing semantics to `germinator init`: every resource is attempted (no fail-fast on the first error), each receives its own `InitializeResult`, and exit status reflects aggregate outcome (success if any succeed, error only if all fail) (partial-processing)

**Foundation packages (June-July 2026 migration)**

- Introduce `internal/iostreams/`, `internal/output/`, and `internal/cmdutil/` foundation packages: `IOStreams` terminal abstraction (TTY detection, lipgloss `Styles`, `Verbosef`, `GERMINATOR_DEBUG` logger), `Exporter` (`json`/`table`/`plain`) interface with `JSONExporter` and `TableExporter`, and the lazy `cmdutil.Factory` dependency-injection mechanism with `sync.OnceValues` caching (scaffold-cli-foundation)
- Introduce `internal/transform/`, `internal/validate/`, `internal/canonicalize/`, and `internal/install/` shell packages holding the service-style `Transformer`/`Validator`/`Canonicalizer`/`Initializer` I/O adapters; `install` (not `init`) avoids Go's reserved word (extract-io-adapters)
- Introduce `internal/permission/` with typed `Allow`/`Deny`/`Ask`/`Action` constants and shared validator rejecting unknown permission actions at runtime (harden-tests-and-coverage)

**Flags & commands (migration and hardening)**

- Add `--output json|table|plain` flag on every library subcommand, replacing the parent-inherited `--json` (wire-factory-and-pilots, migrate-library-readonly, migrate-library-add-create, migrate-library-rest)
- Add `--force` flag on `library remove resource` and `library remove preset` (migrate-library-rest)
- Add `Config.Debug` field and `debug` config-file key (or `GERMINATOR_DEBUG` env var) to drive the diagnostic logger (wire-factory-config-pipeline)
- Add 3-arg `library.FindLibrary(flag, env, cfg)` codifying the canonical precedence `flag > env > config > XDG default`; introduces `GERMINATOR_LIBRARY` env var as a supported tier (wire-factory-config-pipeline)
- Add typed errors `*core.NotFoundError`, `*core.OperationError`, `*core.UsageError`, `*core.CobraUsageError`, plus the pure `core.CanInstallResource` rule for pre-flight ref validation (migrate-library-readonly, migrate-library-add-create, enforce-error-discipline)
- Add `(*library.Library).Refresh`, `RemoveResource`, `RemovePreset`, `Validate`, `Fix`, and additional `AddResource`/`BatchAddResources`/`DiscoverOrphans` methods mirroring the `CreatePreset` precedent (migrate-library-rest, extract-io-adapters)
- Add `MarshalJSON` on every `core.*Error` type returning `{"error": "<Error()>"}` so typed errors survive JSON round-trip (enforce-error-discipline)

### Changed

**Library behaviour**

- Migrate `library add`, `library create preset`, `library refresh`, `library validate`, `library remove`, `library resources`, `library presets`, `library show`, and `library init` to the `NewCmdXxx(f, runF) + runXxx(opts)` pattern with `*cmdutil.Factory` dependency injection and `iostreams.IOStreams`-based I/O (migrate-library-readonly, migrate-library-add-create, migrate-library-rest)
- Unify `library show` not-found errors under `*core.NotFoundError` rendered as `Error: not found: <ref>` (was two distinct strings for resources vs. presets) (migrate-library-readonly)
- Collapse the `library create` Cobra group wrapper; `library create preset` is now registered directly under `library` (migrate-library-add-create)
- Migrate every `library` package function (`RefreshLibrary`, `RemoveResource`, `RemovePreset`, `CreateLibrary`, `LoadDocument`, `DetectType`, `ParsePlatformDocument`, `RenderDocument`, `MarshalCanonical`) to accept `ctx context.Context` as the first parameter so cancellation reaches blocking I/O (propagate-context-through-shell)
- Stop `internal/library` from writing directly to `os.Stdout`: `CreateOptions`, `AddRequest`, `BatchAddOptions`, and `InitRequest` each gain a `Stdout io.Writer` field, and cmd-side callers wire `opts.IO.Out` (fix-library-io-discipline)
- Serialize library writes with `withFileLock` across `save`/`add`/`remove`/`refresh` to prevent concurrent-write corruption
- Parallelize `scanDirectory` via `errgroup.SetLimit(8)` and document the Unix permission convention (fix-library-io-discipline)
- Adopt atomic library writes via new `library.atomicWriteFile(path, data, mode)` helper with `EXDEV` cross-filesystem fallback so cross-device renames (e.g., `/tmp` → `/home`) succeed (fix-library-io-discipline)

**Configuration pipeline**

- Wire `Factory.Config` end-to-end in `main.go` (previously declared but never assigned) so config-file values, `GERMINATOR_DEBUG`, and the koanf env-vars tier actually take effect (wire-factory-config-pipeline)
- Add the env-vars tier as the third of the four-tier merge (defaults → file → env → flags) and fail-fast on malformed `~/.config/germinator/config.toml` (was silently ignored) (wire-factory-config-pipeline)
- Adopt `github.com/adrg/xdg` for XDG path resolution; `Config.Library` default switches from a literal tilde-prefix string to an empty string that falls through to the XDG default (`~/.local/share/germinator/library/`) (wire-factory-config-pipeline)
- Rename `Config.Platform` → `Config.PlatformDefault` (wire-factory-config-pipeline)
- Convert platform adapters (`internal/opencode`, `internal/claude-code`) to package singletons with explicit `var _ permission.Adapter = (*Adapter)(nil)` compile-time contract checks (harden-tests-and-coverage)

**Error handling & exit codes**

- Migrate every command to the `NewCmdXxx(f, runF) + runXxx(opts)` pattern with `*cmdutil.Factory` dependency injection and `iostreams.IOStreams`-based I/O: `adapt`, `validate`, `canonicalize`, `init`, every `library` subcommand, `config init`, `config validate`, `completion`, and `version` (scaffold-cli-foundation through migrate-completion-cleanup)
- Rewrite `main.go` as the single composition root constructing `IOStreams` + `Factory` with signal-aware root context; centralized error handling via `output.FormatError` + `os.Exit(int(cmdutil.ExitCodeFor(err)))` (wire-factory-and-pilots)
- Drop the 12-substring Cobra-prefix dispatch in favour of typed `*pflag.*Error` + `*core.CobraUsageError` so pflag validation errors render consistently (enforce-error-discipline)
- Migrate `internal/library` to typed errors (`NotFoundError`, `OperationError`, `ValidationError`) so error classification works uniformly across the library and cmd layers (enforce-error-discipline)
- Extend `BatchFailureInfo` with `ErrorType`/`Cause` fields to preserve the typed-error chain through JSON output (enforce-error-discipline)
- Wire `--verbose` flag through `PersistentPreRunE` to `IOStreams.Verbose` so all commands honor the global verbosity flag

**Internal restructuring**

- Rename `internal/domain/` → `internal/core/` and flatten `internal/infrastructure/{parsing,serialization,adapters,config,library}/` into top-level `internal/{parser,renderer,claude-code,opencode,config,library}/` (scaffold-cli-foundation)
- Simplify `cmdutil.Factory` and drop the signal double-wrap
- Demote `cobra` output globals to explicit writer wiring to keep `xdgReload` test seams deterministic (harden-tests-and-coverage)

### Removed

- Remove `internal/application/`, `internal/service/`, `internal/models/`, the empty `internal/infrastructure/` tree, and the `internal/permission.Adapter` interface (parser keeps a local `platformAdapter` definition) (scaffold-cli-foundation, migrate-library-rest, migrate-completion-cleanup, harden-tests-and-coverage)
- Remove `ServiceContainer`, `CommandConfig`, `Verbosity`/`VerbosePrint`, `ErrorFormatter`, and `CategorizeError`/`Category*` enum (replaced by `cmdutil.Factory`, `IOStreams.Verbosef`, `output.FormatError`, `cmdutil.ExitCodeFor`) (wire-factory-and-pilots, migrate-library-rest)
- Remove `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`, `cmd/error_formatter.go`, `cmd/verbose.go`, and the temporary `cmd/legacy_bridge.go` shim (wire-factory-and-pilots, migrate-library-rest)
- Remove duplicate `PlatformClaudeCode`/`PlatformOpenCode` definitions from `internal/parser/loader.go` (single source in `internal/core/platform.go`) (migrate-completion-cleanup)
- Remove `cmdutil.AddOutputFlags` re-export shim (callers now use `output.AddOutputFlags` directly) (remove-cmdutil-output-reexport)
- Remove the entire `internal/warning` package including the `MaybeWarnLegacyExitCode` exit-code-3-6 canary — the migration is now complete and the deprecation warning is no longer needed (enforce-error-discipline)

### Fixed

- Fix AGENTS.md drift with respect to `NewFactory`/`FindLibrary` signatures, `RefreshError`/`SkipInfo` struct shapes, `FormatError` dispatch, and the exit-code canary reference (reconcile-agents-md-drift)
- Add missing `//go:build integration` tag on `internal/cmdutil/integration_test.go` so the default `go test ./...` run skips it (must be invoked via `mise run test:integration`) (reconcile-agents-md-drift)
- Delete four unused JSON output types (`PresetsJSONOutput`, `PresetInfoJSON`, `ShowResourceJSONOutput`, `ShowPresetJSONOutput`) (reconcile-agents-md-drift)
- Reject unknown permission actions at runtime via the shared `internal/permission` validator returning `*core.ConfigError` instead of silently mapping them (harden-tests-and-coverage)
- Guard `(*Library).CreatePreset` against nil receivers (harden-tests-and-coverage)
- Add `Unwrap` to `NotFoundError` so typed-error symmetry with `errors.Is`/`errors.As` works as expected
- Harden `xdgReload` read path and demote Cobra output globals so concurrent reads during file watching no longer race (harden-tests-and-coverage)
- Add `go.uber.org/goleak` `TestMain` guard to `internal/library` to detect goroutine leaks in tests (harden-tests-and-coverage)
- Eliminate mutable package vars from `cmd/` to avoid data races (harden-tests-and-coverage)

### Breaking

- **NotFoundError exit code 2 → 1**: `*core.NotFoundError` now maps to exit code 1 (was 2 — semantic correction; "not found" is a runtime state, not a usage error). Combined with the previous migration collapse of codes 3–6 to 1, every typed error now exits 1 (or 2 for usage errors). Scripts that special-cased `exit 2` for not-found must update to `exit 1`. (enforce-error-discipline)
- **Exit codes 3–6 collapsed to 1**: the error exit code is now `1` for all error types (was previously `3` Config, `4` Git, `5` Validation, `6` NotFound); usage errors remain `2`. The exit code is no longer semantic — check stderr for the typed error message. (wire-factory-and-pilots, migrate-domain-commands, migrate-library-readonly, migrate-library-add-create, migrate-library-rest, migrate-config-commands)
- **`Config.Platform` → `Config.PlatformDefault`**: the TOML config field is renamed. In-repo impact is zero (no struct-literal consumers); external consumers using raw `toml.Decode` on their own config files must update the key. (wire-factory-config-pipeline)
- **`Config.Library` default behaviour**: the default value of `Config.Library` changes from a literal tilde-prefix string to the empty string, falling through to XDG default resolution via `github.com/adrg/xdg`. The resolved path is identical for typical users (`~/.local/share/germinator/library/`) but consumers who explicitly set the empty string now get XDG behaviour instead of a literal tilde. (wire-factory-config-pipeline)
- **`internal/warning` package removed**: the `MaybeWarnLegacyExitCode` canary and its one-time stderr warning are gone — the exit-code 3–6 → 1 migration is complete. Any caller importing `internal/warning` (in-repo callers all gone) breaks at compile time. (enforce-error-discipline)
- **`--output`/`-o` flag renamed to `--output-dir` on `init`**: `germinator init --output <dir>` is now `germinator init --output-dir <dir>`. Disambiguates from the new `--output` format flag and aligns the flag name with its semantic. (migrate-init-command)
- **`--output` flag renamed to `--output-path` on config commands**: `config init --output <path>` is now `config init --output-path <path>`; same for `config validate`. Disambiguates from the new `--output` format flag. (migrate-config-commands)
- **`--json` flag replaced by `--output json` on library commands**: the parent-inherited `--json` is removed; use `--output json` on every library subcommand. (wire-factory-and-pilots, migrate-library-readonly, migrate-library-add-create, migrate-library-rest)
- **`init` exit-code semantics**: partial processing means `germinator init` exits 0 if any resource in the request succeeds, and exits 1 only if every resource fails (was: first failure aborts with non-zero). (partial-processing)
- **`BatchFailureInfo` JSON shape**: `errorType` and `cause` fields added to the JSON output for `library add --batch`. Consumers parsing the previous two-field shape must be widened. (enforce-error-discipline)
- **`core.*Error` JSON wire shape**: every typed error now serializes as `{"error": "<Error()>"}` when marshaled via `MarshalJSON`; consumers that previously parsed the raw struct fields must read the `error` key. (enforce-error-discipline)
- **Removed packages**: `internal/application/`, `internal/service/`, and `internal/models/` directories are deleted. Any Go code importing these breaks (germinator is a CLI application, not a library, so external importers are rare). (migrate-library-rest, migrate-completion-cleanup)

## [0.9.0] - 2026-03-31


### Added

- Add `config init` and `config validate` commands for scaffolding and validating configuration files (config-init-command)
- Add `library init` command for scaffolding library directory structures with `library.yaml` and resource subdirectories (library-init)
- Add `library create preset` command for creating presets in the library with validation and overwrite protection (create-library-preset-command)
- Add `library add` command for importing resources into the library with auto-detection, canonicalization, and validation (library-add)

### Changed

- Update release validation to allow unstaged CHANGELOG.md for same-commit changelog updates and release

## [0.8.0] - 2026-03-30

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
