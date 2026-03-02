## 1. Core Types

- [x] 1.1 Create `internal/library/library.go` with `Library`, `Resource`, `Preset` types
- [x] 1.2 Library.Resources as `map[string]map[string]Resource` (type → name → resource)
- [x] 1.3 Add `ResourceType` enum (skill, agent, command, memory)
- [x] 1.4 Add `Resource.Validate()` and `Preset.Validate()` methods

## 2. Library Loader

- [x] 2.1 Create `internal/library/loader.go` with `LoadLibrary(path string) (*Library, error)`
- [x] 2.2 Implement `library.yaml` parsing with gopkg.in/yaml.v3
- [x] 2.3 Add library.yaml structure validation (version, resources, presets)
- [x] 2.4 Add resource field validation (type, path, description)
- [x] 2.5 Handle missing library directory error
- [x] 2.6 Handle missing library.yaml error
- [x] 2.7 Add error handling for unsupported version

## 3. Resolver

- [x] 3.1 Create `internal/library/resolver.go`
- [x] 3.2 Implement `ResolveResource(lib *Library, ref string) (path string, error)` where ref is `type/name`
- [x] 3.3 Implement `ParseRef(ref string) (typ, name string, error)` for splitting `type/name`
- [x] 3.4 Implement `ResolveResources(lib *Library, refs []string) ([]string, error)` for batch resolution
- [x] 3.5 Implement `ResolvePreset(lib *Library, name string) ([]string, error)` returning `type/name` refs
- [x] 3.6 Implement `GetOutputPath(typ, name, platform, outputDir string) string` with platform-specific path derivation
- [x] 3.7 Add output path mapping for all type/platform combinations (skill, agent, command, memory × opencode, claude-code)
- [x] 3.8 Add "resource not found", "preset not found", and "invalid resource reference" errors

## 4. Lister

- [x] 4.1 Create `internal/library/lister.go`
- [x] 4.2 Implement `ListResources(lib *Library) map[string][]Resource` (grouped by type)
- [x] 4.3 Implement `ListPresets(lib *Library) []Preset`
- [x] 4.4 Add `FormatResourcesList(lib *Library) string` for CLI output formatting
- [x] 4.5 Add `FormatPresetsList(lib *Library) string` for CLI output formatting

## 5. Discovery

- [x] 5.1 Create `internal/library/discovery.go`
- [x] 5.2 Implement `FindLibrary(flagPath, envPath string) string` with priority chain
- [x] 5.3 Add `DefaultLibraryPath()` returning `~/.config/germinator/library/`
- [x] 5.4 Implement default path using `os.UserConfigDir` for XDG compliance

## 6. Service Layer - Initializer

- [x] 6.1 Create `internal/services/initializer.go` with `InitializeResources` function
- [x] 6.2 Implement resource loading using `core.LoadDocument`
- [x] 6.3 Implement resource transformation using `core.RenderDocument`
- [x] 6.4 Implement file writing with directory creation
- [x] 6.5 Implement dry-run mode (print only, no writes)
- [x] 6.6 Implement force overwrite logic
- [x] 6.7 Implement file exists check without force
- [x] 6.8 Add fail-fast error handling

## 7. CLI - Library Commands

- [x] 7.1 Create `cmd/library.go` with root `library` command
- [x] 7.2 Add `--library` flag to library command
- [x] 7.3 Implement `library resources` subcommand (output grouped by type: Skills/Agents/Commands sections)
- [x] 7.4 Implement `library presets` subcommand
- [x] 7.5 Implement `library show <ref>` subcommand parsing `type/name` format
- [x] 7.6 Handle invalid ref format (no slash) with clear error message
- [x] 7.7 Wire commands in `cmd/root.go`

## 8. CLI - Init Command

- [x] 8.1 Create `cmd/init.go` with `initCmd` Cobra command
- [x] 8.2 Add `--platform` flag (required, with validation)
- [x] 8.3 Add `--resources` flag (comma-separated list)
- [x] 8.4 Add `--preset` flag (mutually exclusive with `--resources`)
- [x] 8.5 Add `--library` flag for custom library path
- [x] 8.6 Add `--output` flag for output directory (default: .)
- [x] 8.7 Add `--dry-run` flag to preview changes
- [x] 8.8 Add `--force` flag to overwrite existing files
- [x] 8.9 Implement flag validation (platform required, resources or preset required)
- [x] 8.10 Wire `initCmd` to `rootCmd` in main
- [x] 8.11 Implement success output formatting
- [x] 8.12 Implement error output formatting with file paths

## 9. Testing - Library Package

- [x] 9.1 Create `internal/library/library_test.go` with type validation tests
- [x] 9.2 Create `internal/library/loader_test.go` with table-driven tests
- [x] 9.3 Create `internal/library/resolver_test.go` with table-driven tests
- [x] 9.4 Add tests for all platform/type output path combinations
- [x] 9.5 Create `internal/library/lister_test.go` with listing tests
- [x] 9.6 Create `internal/library/discovery_test.go` with priority chain tests
- [x] 9.7 Add test fixtures in `test/fixtures/library/` with sample library.yaml

## 10. Testing - Service Layer

- [x] 10.1 Create `internal/services/initializer_test.go`
- [x] 10.2 Add integration test with sample library resources
- [x] 10.3 Add tests for dry-run mode
- [x] 10.4 Add tests for force overwrite behavior
- [x] 10.5 Add tests for file exists error handling

## 11. Testing - CLI

- [x] 11.1 Create `cmd/library_test.go` with CLI integration tests
- [x] 11.2 Add init command tests to `cmd/cmd_test.go`
- [x] 11.3 Test required flag validation
- [x] 11.4 Test successful resource installation
- [x] 11.5 Test error handling for missing resources
- [x] 11.6 Test `library resources` output
- [x] 11.7 Test `library presets` output
- [x] 11.8 Test `library show` output

## 12. Documentation

- [ ] 12.1 Update `cmd/AGENTS.md` with library and init command documentation
- [ ] 12.2 Create `internal/library/AGENTS.md` with package documentation
- [ ] 12.3 Update root `AGENTS.md` with init command in Essential Commands

## 13. Verification

- [ ] 13.1 Run `mise run check` (lint, format, test, build)
- [ ] 13.2 Run `mise run test:coverage` and verify 70%+ coverage for `internal/library/`
- [ ] 13.3 Verify all spec scenarios are covered by tests
