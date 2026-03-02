## Context

Germinator's CLI commands currently use a simple pattern:
- Global command variables (`var validateCmd = &cobra.Command{...}`)
- `init()` functions to register commands with `rootCmd`
- Direct calls to service functions (`services.ValidateDocument()`)
- `main()` function inside `cmd/root.go`

This works for a simple tool but creates issues as complexity grows:
- No way to inject mock services for testing
- Implicit dependencies via global state
- Service wiring scattered across `init()` functions

## Goals / Non-Goals

**Goals:**
- Establish ServiceContainer pattern for holding service instances
- Convert commands from `init()` pattern to constructor pattern (`NewXCommand`)
- Make `main.go` the composition root where all dependencies are wired
- Enable future testability through dependency injection

**Non-Goals:**
- Creating actual services (ServiceContainer may be initially sparse)
- Changing command behavior or CLI semantics
- Adding test infrastructure (future change)
- Interface-based service abstractions (future change)

## Decisions

### Decision 1: ServiceContainer over Individual Service Passing

**Choice:** Use a `ServiceContainer` struct to group services passed to commands.

**Rationale:**
- Single parameter to pass through command tree
- Easy to add new services without changing all command constructors
- Clear ownership of service instances
- Matches twiggit reference implementation

**Alternatives:**
- Pass individual services to each command: Combinatorial explosion as services grow
- Use global service registry: Reintroduces implicit dependencies we're trying to avoid
- Context-based service lookup: Over-engineering for current needs

### Decision 2: main.go as Composition Root

**Choice:** Extract `main()` to root-level `main.go`, making it the single place where services are wired.

**Rationale:**
- Clear visibility of all dependencies
- Single location for service lifecycle management
- Enables future test setups to wire alternative configurations
- Follows Dependency Injection best practices

**Alternatives:**
- Keep main() in cmd/root.go: Loses clear separation of wiring vs command definition
- Use wire/dig DI frameworks: Over-engineering for a CLI with ~4 commands

### Decision 3: Constructor Pattern for Commands

**Choice:** Each command has `NewXCommand(config *CommandConfig) *cobra.Command`.

**Rationale:**
- Explicit dependency declaration
- Commands are self-contained, testable units
- Eliminates `init()` ordering concerns
- Enables parallel command construction

**Alternatives:**
- Keep init() pattern: Implicit dependencies, harder to test
- Use functional options: Unnecessary complexity for current needs

## Risks / Trade-offs

**Risk:** ServiceContainer is initially sparse (may have no real services yet)
→ **Mitigation:** Pattern is established; container grows naturally as services are identified

**Risk:** Large diff touching all command files
→ **Mitigation:** Changes are mechanical; each file follows same pattern

**Risk:** Dependency on cli-infrastructure change for CommandConfig
→ **Mitigation:** Explicit dependency declaration; this change extends that foundation
