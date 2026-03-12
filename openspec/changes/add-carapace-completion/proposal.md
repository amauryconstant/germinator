## Why

Users have to manually type or remember resource references (e.g., `skill/commit`), preset names, and platform values when using the CLI. Adding shell completion improves discoverability, reduces errors, and provides a better developer experience by suggesting valid options dynamically.

## What Changes

- Add `github.com/carapace-sh/carapace` dependency for multi-shell completion support
- Create `germinator completion <shell>` command supporting bash, zsh, fish, powershell, and additional shells
- Add dynamic completions for:
  - `init --resources` → available resources from library (e.g., `skill/commit`, `agent/reviewer`)
  - `init --preset` → available presets from library
  - `library show <ref>` → resource and preset references
- Add static completions for `--platform` flag on relevant commands
- Extend config with `completion.timeout` and `completion.cache_ttl` settings
- Implement in-memory caching for library data during completion (5s default TTL)

## Capabilities

### New Capabilities

- `shell-completion`: Carapace-based shell completion system with dynamic library-aware suggestions and configurable timeout/caching

### Modified Capabilities

None. This is a new feature that doesn't change existing spec-level behavior.

## Impact

**Files Created:**
- `cmd/completion.go` - Completion command for all shells
- `cmd/completions.go` - Action functions for dynamic/static completions
- `cmd/completions_test.go` - Tests for completion actions

**Files Modified:**
- `go.mod` - Add carapace dependency
- `internal/config/config.go` - Add `CompletionConfig` struct
- `cmd/root.go` - Wire carapace, replace Cobra's default completion
- `cmd/init.go` - Add flag completions for `--resources`, `--preset`, `--platform`
- `cmd/library.go` - Add positional completion for `show` command

**Dependencies:**
- `github.com/carapace-sh/carapace v1.11.1`

**Configuration:**
```toml
[completion]
timeout = "500ms"    # Max time for library loading during completion
cache_ttl = "5s"     # Cache duration for library data
```
