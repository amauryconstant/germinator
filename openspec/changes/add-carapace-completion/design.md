## Context

Germinator currently uses Cobra's built-in completion system, which provides basic static completions but lacks dynamic library-aware suggestions. Users must remember resource references (`skill/commit`), preset names, and platform values.

The completion system needs to:
- Query the library dynamically for available resources and presets
- Work with multiple shells (bash, zsh, fish, powershell, etc.)
- Handle library path resolution (flag > env > config > default)
- Perform efficiently with caching and timeouts

**Constraints:**
- Completion runs in a subprocess before flag parsing completes
- Library loading can be slow (filesystem reads)
- Must not block the shell if library is unavailable

## Goals / Non-Goals

**Goals:**
- Provide dynamic completions for resources, presets, and platforms
- Support bash, zsh, fish, powershell, and additional shells via carapace
- Cache library data with configurable TTL to avoid repeated loads
- Gracefully degrade when library is unavailable or timeout expires

**Non-Goals:**
- Multi-part progressive completion (e.g., `project/branch` like twiggit) - germinator doesn't need this
- Context-aware completions based on current directory
- Custom completion scripts per-project

## Decisions

### Decision 1: Use Carapace over Cobra's native completion

**Choice:** `github.com/carapace-sh/carapace`

**Rationale:**
- Supports 10+ shells vs Cobra's 4 (bash, zsh, fish, powershell)
- Better dynamic completion support with `ActionCallback`
- Built-in caching and timeout mechanisms
- Proven in twiggit codebase with similar patterns

**Alternatives considered:**
- Cobra native completion: Simpler but less shell support, weaker dynamic completion
- Custom completion scripts: More maintenance, shell-specific code

### Decision 2: In-memory cache with TTL

**Choice:** Package-level cache variable with 5-second TTL

**Rationale:**
- Completions are short-lived subprocess calls
- 5s is long enough for repeated TAB presses during a session
- Simple to implement, no external dependencies
- Configurable via `completion.cache_ttl`

**Alternatives considered:**
- No caching: Too slow for large libraries
- File-based cache: Overkill for subprocess completions
- Redis/shared cache: Unnecessary complexity

### Decision 3: Library path resolution in completion context

**Choice:** Check in order: `--library` flag > `GERMINATOR_LIBRARY` env > config file > default

**Rationale:**
- Carapace provides access to already-parsed flags via `carapace.Context`
- Environment variable is always available
- Config file may not be loaded yet (completion runs early)
- Consistent with existing `library.FindLibrary()` logic

### Decision 4: Silent failure on library load errors

**Choice:** Return empty completions instead of errors

**Rationale:**
- Shell completion is a UX enhancement, not critical functionality
- Error messages in completion context are confusing to users
- Silent degradation is standard practice (see twiggit patterns)

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COMPLETION ARCHITECTURE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   cmd/root.go                                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  carapace.Gen(cmd)                     // Initialize carapace        │   │
│   │  cmd.AddCommand(newCompletionCommand()) // Replace Cobra completion  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   cmd/completion.go                                                         │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  newCompletionCommand()                                              │   │
│   │    └─► Subcommands: bash, zsh, fish, powershell, elvish, nushell,   │   │
│   │                       oil, tcsh, xonsh, cmd-clink                    │   │
│   │    └─► Each generates snippet via carapace.Gen().Snippet(shell)     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   cmd/completions.go                                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  actionPlatforms()        → static ["claude-code", "opencode"]      │   │
│   │  actionResources(libPath) → dynamic from library.Resources          │   │
│   │  actionPresets(libPath)   → dynamic from library.Presets            │   │
│   │  actionLibraryRefs()      → resources + preset/ prefix              │   │
│   │                                                                       │   │
│   │  Cache: package-level map with mutex, 5s TTL                         │   │
│   │  Timeout: 500ms via carapace.ActionCallback().Timeout()              │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   internal/config/config.go                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  type CompletionConfig struct {                                      │   │
│   │      Timeout  string `koanf:"timeout"`   // default: "500ms"         │   │
│   │      CacheTTL string `koanf:"cache_ttl"` // default: "5s"            │   │
│   │  }                                                                   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Slow library loading blocks shell | 500ms timeout with silent fallback to empty completions |
| Stale cache after library changes | Short 5s TTL, user can press TAB again after changes |
| Config not loaded during completion | Use env/flag paths as primary, config as fallback |
| Carapace dependency adds ~2MB to binary | Acceptable trade-off for multi-shell support |

## Open Questions

None. The design is straightforward based on proven twiggit patterns.
