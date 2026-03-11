**Location**: `internal/models/`
**Parent**: See `/internal/AGENTS.md` for core patterns

---

# Models Package

Canonical document models representing AI coding assistant configurations.

---

# Structure

```
internal/models/
├── canonical/
│   ├── models.go      # Agent, Command, Skill, Memory structs
│   └── doc.go         # Package documentation
└── constants.go       # Platform and permission constants
```

---

# Canonical Models (`canonical/models.go`)

## PermissionPolicy

```go
type PermissionPolicy string

const (
    PermissionPolicyRestrictive  PermissionPolicy = "restrictive"
    PermissionPolicyBalanced     PermissionPolicy = "balanced"
    PermissionPolicyPermissive   PermissionPolicy = "permissive"
    PermissionPolicyAnalysis     PermissionPolicy = "analysis"
    PermissionPolicyUnrestricted PermissionPolicy = "unrestricted"
)

func (p PermissionPolicy) IsValid() bool
```

Canonical enum for tool permissions. Transformed to platform-specific formats by adapters.

## Agent

```go
type Agent struct {
    Name        string           `yaml:"name" json:"name"`
    Description string           `yaml:"description" json:"description"`
    Content     string           `yaml:"-" json:"-"`
    FilePath    string           `yaml:"-" json:"-"`

    Tools            []string         `yaml:"tools,omitempty" json:"tools,omitempty"`
    DisallowedTools  []string         `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`
    PermissionPolicy PermissionPolicy `yaml:"permissionPolicy,omitempty" json:"permissionPolicy,omitempty"`
    Behavior         AgentBehavior    `yaml:"behavior,omitempty" json:"behavior,omitempty"`
    Targets          PlatformConfig   `yaml:"targets,omitempty" json:"targets,omitempty"`
    Extensions       AgentExtensions  `yaml:"extensions,omitempty" json:"extensions,omitempty"`
    Model            string           `yaml:"model,omitempty" json:"model,omitempty"`
}
```

| Field | Purpose |
|-------|---------|
| `Name` | Agent identifier |
| `Description` | Human-readable description |
| `Content` | Prompt/narrative content (not serialized) |
| `FilePath` | Source file path (not serialized) |
| `Tools` | Allowed tool names |
| `DisallowedTools` | Explicitly denied tools |
| `PermissionPolicy` | Canonical permission enum |
| `Behavior` | Mode, temperature, steps, prompt, hidden, disabled |
| `Targets` | Platform-specific overrides |
| `Extensions` | Hooks and custom extensions |
| `Model` | Platform-specific model ID |

## Command

```go
type Command struct {
    Name        string           `yaml:"name" json:"name"`
    Description string           `yaml:"description" json:"description"`
    Content     string           `yaml:"-" json:"-"`
    FilePath    string           `yaml:"-" json:"-"`

    Tools     []string         `yaml:"tools,omitempty" json:"tools,omitempty"`
    Execution CommandExecution `yaml:"execution,omitempty" json:"execution,omitempty"`
    Arguments CommandArguments `yaml:"arguments,omitempty" json:"arguments,omitempty"`
    Targets   PlatformConfig   `yaml:"targets,omitempty" json:"targets,omitempty"`
    Model     string           `yaml:"model,omitempty" json:"model,omitempty"`
}
```

| Field | Purpose |
|-------|---------|
| `Execution.Context` | Execution context (fork) |
| `Execution.Subtask` | Run as subtask |
| `Execution.Agent` | Agent reference |
| `Arguments.Hint` | Argument hint text |

## Skill

```go
type Skill struct {
    Name        string           `yaml:"name" json:"name"`
    Description string           `yaml:"description" json:"description"`
    Content     string           `yaml:"-" json:"-"`
    FilePath    string           `yaml:"-" json:"-"`

    Tools      []string        `yaml:"tools,omitempty" json:"tools,omitempty"`
    Extensions SkillExtensions `yaml:"extensions,omitempty" json:"extensions,omitempty"`
    Execution  SkillExecution  `yaml:"execution,omitempty" json:"execution,omitempty"`
    Targets    PlatformConfig  `yaml:"targets,omitempty" json:"targets,omitempty"`
    Model      string          `yaml:"model,omitempty" json:"model,omitempty"`
}
```

| Field | Purpose |
|-------|---------|
| `Extensions.License` | Skill license |
| `Extensions.Compatibility` | Compatible platforms |
| `Extensions.Metadata` | Custom key-value pairs |
| `Extensions.Hooks` | Lifecycle hooks |
| `Execution.Context` | Execution context (fork) |
| `Execution.Agent` | Agent reference |
| `Execution.UserInvocable` | User can invoke directly |

## Memory

```go
type Memory struct {
    Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
    Content  string   `yaml:"content,omitempty" json:"content,omitempty"`
    FilePath string   `yaml:"-" json:"-"`
}
```

| Field | Purpose |
|-------|---------|
| `Paths` | File references (@-style) |
| `Content` | Narrative context |

---

# Supporting Types

## AgentBehavior

```go
type AgentBehavior struct {
    Mode        string   `yaml:"mode,omitempty" json:"mode,omitempty"`
    Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    Steps       int      `yaml:"steps,omitempty" json:"steps,omitempty"`
    Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
    Hidden      bool     `yaml:"hidden,omitempty" json:"hidden,omitempty"`
    Disabled    bool     `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}
```

## PlatformConfig

```go
type PlatformConfig map[string]map[string]interface{}
```

Platform-specific field overrides. Keyed by platform name (`claude-code`, `opencode`).

---

# Constants (`constants.go`)

| Constant | Value |
|----------|-------|
| `PlatformClaudeCode` | `"claude-code"` |
| `PlatformOpenCode` | `"opencode"` |

---

# Validation

**Note**: Validation methods removed in v0.5.0. Use `internal/validation/` package instead.

| Document Type | Validator Function |
|---------------|-------------------|
| Agent | `validation.ValidateAgent()` |
| Command | `validation.ValidateCommand()` |
| Skill | `validation.ValidateSkill()` |
| Memory | `validation.ValidateMemory()` |

See `internal/validation/AGENTS.md` for `Result[T]` pattern and composable validators.

---

# Field Tags

- `yaml:"-"` / `json:"-"` - Exclude from serialization (Content, FilePath)
- `yaml:",omitempty"` - Omit empty values
- JSON tags mirror YAML for potential future use

---

# See Also

- `internal/validation/AGENTS.md` - Validation patterns
- `internal/adapters/AGENTS.md` - How adapters transform these models
- `config/AGENTS.md` - How templates render these models
