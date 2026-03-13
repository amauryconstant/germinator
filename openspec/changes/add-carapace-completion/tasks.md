## 1. Dependencies

- [x] 1.1 Add `github.com/carapace-sh/carapace v1.11.1` to go.mod
- [x] 1.2 Run `go mod tidy` to update dependencies

## 2. Configuration Extension

- [x] 2.1 Add `CompletionConfig` struct to `internal/config/config.go` with `Timeout` and `CacheTTL` fields
- [x] 2.2 Add completion config defaults in `DefaultConfig()` (timeout: "500ms", cache_ttl: "5s")
- [x] 2.3 Add koanf tags for completion config fields
- [x] 2.4 Add config tests for completion settings

## 3. Completion Command

- [x] 3.1 Create `cmd/completion.go` with `newCompletionCommand(rootCmd *cobra.Command)`
- [x] 3.2 Implement completion subcommands for all shells (bash, zsh, fish, powershell, elvish, nushell, oil, tcsh, xonsh, cmd-clink)
- [x] 3.3 Add shell-specific installation instructions in help text
- [x] 3.4 Use `carapace.Gen(rootCmd).Snippet(shell)` to generate completion scripts

## 4. Completion Actions

- [x] 4.1 Create `cmd/completions.go` with package-level cache (map with mutex, TTL support)
- [x] 4.2 Implement `getCompletionTimeout(cfg *config.Config) time.Duration` helper
- [x] 4.3 Implement `getCacheTTL(cfg *config.Config) time.Duration` helper
- [x] 4.4 Implement `resolveLibraryPath(cmd *cobra.Command, cfg *config.Config) string` for path resolution
- [x] 4.5 Implement `actionPlatforms() carapace.Action` for static platform completions
- [x] 4.6 Implement `actionResources(cmd *cobra.Command) carapace.Action` with caching and timeout
- [x] 4.7 Implement `actionPresets(cmd *cobra.Command) carapace.Action` with caching and timeout
- [x] 4.8 Implement `actionLibraryRefs(cmd *cobra.Command) carapace.Action` combining resources and presets

## 5. Wire Completions into Commands

- [x] 5.1 Modify `cmd/root.go` to initialize carapace with `carapace.Gen(cmd)`
- [x] 5.2 Replace Cobra's default completion with carapace completion command in `cmd/root.go`
- [x] 5.3 Add `--platform` flag completion to `cmd/adapt.go`
- [x] 5.4 Add `--platform` flag completion to `cmd/validate.go`
- [x] 5.5 Add `--platform` flag completion to `cmd/canonicalize.go`
- [x] 5.6 Add `--platform` and `--resources` flag completions to `cmd/init.go`
- [x] 5.7 Add `--preset` flag completion to `cmd/init.go`
- [x] 5.8 Add positional completion for `library show <ref>` in `cmd/library.go`

## 6. Tests

- [x] 6.1 Create `cmd/completions_test.go` with tests for `getCompletionTimeout`
- [x] 6.2 Add tests for `getCacheTTL`
- [x] 6.3 Add tests for `actionPlatforms` returns valid platforms
- [x] 6.4 Add tests for `actionResources` with mock library
- [x] 6.5 Add tests for `actionPresets` with mock library
- [x] 6.6 Add tests for `actionLibraryRefs` combines resources and presets
- [x] 6.7 Add tests for cache hit/miss scenarios
- [x] 6.8 Add tests for silent failure on library load errors

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
