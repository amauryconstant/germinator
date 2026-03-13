## 1. Dependencies

- [ ] 1.1 Add `github.com/carapace-sh/carapace v1.11.1` to go.mod
- [ ] 1.2 Run `go mod tidy` to update dependencies

## 2. Configuration Extension

- [ ] 2.1 Add `CompletionConfig` struct to `internal/config/config.go` with `Timeout` and `CacheTTL` fields
- [ ] 2.2 Add completion config defaults in `DefaultConfig()` (timeout: "500ms", cache_ttl: "5s")
- [ ] 2.3 Add koanf tags for completion config fields
- [ ] 2.4 Add config tests for completion settings

## 3. Completion Command

- [ ] 3.1 Create `cmd/completion.go` with `newCompletionCommand(rootCmd *cobra.Command)`
- [ ] 3.2 Implement completion subcommands for all shells (bash, zsh, fish, powershell, elvish, nushell, oil, tcsh, xonsh, cmd-clink)
- [ ] 3.3 Add shell-specific installation instructions in help text
- [ ] 3.4 Use `carapace.Gen(rootCmd).Snippet(shell)` to generate completion scripts

## 4. Completion Actions

- [ ] 4.1 Create `cmd/completions.go` with package-level cache (map with mutex, TTL support)
- [ ] 4.2 Implement `getCompletionTimeout(cfg *config.Config) time.Duration` helper
- [ ] 4.3 Implement `getCacheTTL(cfg *config.Config) time.Duration` helper
- [ ] 4.4 Implement `resolveLibraryPath(c carapace.Context, cfg *config.Config) string` for path resolution
- [ ] 4.5 Implement `actionPlatforms() carapace.Action` for static platform completions
- [ ] 4.6 Implement `actionResources(libPath string, cfg *config.Config) carapace.Action` with caching and timeout
- [ ] 4.7 Implement `actionPresets(libPath string, cfg *config.Config) carapace.Action` with caching and timeout
- [ ] 4.8 Implement `actionLibraryRefs(libPath string, cfg *config.Config) carapace.Action` combining resources and presets

## 5. Wire Completions into Commands

- [ ] 5.1 Modify `cmd/root.go` to initialize carapace with `carapace.Gen(cmd)`
- [ ] 5.2 Replace Cobra's default completion with carapace completion command in `cmd/root.go`
- [ ] 5.3 Add `--platform` flag completion to `cmd/adapt.go`
- [ ] 5.4 Add `--platform` flag completion to `cmd/validate.go`
- [ ] 5.5 Add `--platform` flag completion to `cmd/canonicalize.go`
- [ ] 5.6 Add `--platform` and `--resources` flag completions to `cmd/init.go`
- [ ] 5.7 Add `--preset` flag completion to `cmd/init.go`
- [ ] 5.8 Add positional completion for `library show <ref>` in `cmd/library.go`

## 6. Tests

- [ ] 6.1 Create `cmd/completions_test.go` with tests for `getCompletionTimeout`
- [ ] 6.2 Add tests for `getCacheTTL`
- [ ] 6.3 Add tests for `actionPlatforms` returns valid platforms
- [ ] 6.4 Add tests for `actionResources` with mock library
- [ ] 6.5 Add tests for `actionPresets` with mock library
- [ ] 6.6 Add tests for `actionLibraryRefs` combines resources and presets
- [ ] 6.7 Add tests for cache hit/miss scenarios
- [ ] 6.8 Add tests for silent failure on library load errors

## 7. Documentation

- [ ] 7.1 Update `internal/config/AGENTS.md` with completion config documentation
- [ ] 7.2 Update `cmd/AGENTS.md` with completion command documentation

## 8. Verification

- [ ] 8.1 Run `mise run check` to verify lint, format, tests pass
- [ ] 8.2 Manually test bash completion: `source <(./bin/germinator completion bash)`
- [ ] 8.3 Manually test zsh completion: `source <(./bin/germinator completion zsh)`
- [ ] 8.4 Manually test fish completion: `./bin/germinator completion fish | source`
- [ ] 8.5 Verify dynamic completions work for `germinator init --resources <TAB>`
- [ ] 8.6 Verify dynamic completions work for `germinator init --preset <TAB>`
- [ ] 8.7 Verify dynamic completions work for `germinator library show <TAB>`
- [ ] 8.8 Verify static completions work for `germinator adapt --platform <TAB>`
