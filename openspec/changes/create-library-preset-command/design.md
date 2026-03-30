## Context

The library system manages canonical resources (skills, agents, commands, memory) organized in a `library.yaml` file. Users can list and view presets, but cannot create new presets through the CLI—currently requiring manual YAML editing.

The existing library infrastructure provides:
- `LoadLibrary()` for reading library.yaml
- `ResolvePreset()` for expanding preset names to resource lists
- `ListPresets()` for displaying preset details
- No write/persistence capability

## Goals / Non-Goals

**Goals:**
- Enable preset creation via `germinator library create preset <name>`
- Validate referenced resources exist before creating preset
- Support overwrite with `--force` flag
- Follow existing CLI patterns (help text, error messages, output formatting)

**Non-Goals:**
- Interactive/prompt-based creation (non-interactive flags only)
- Creating resources themselves (only creating presets that reference existing resources)
- Bulk import/export of presets
- Preset modification (update existing preset resources/description)

## Decisions

### 1. Command Structure: `library create preset`

**Choice:** `germinator library create preset <name> --resources ... --description ...`

**Rationale:** Groups creation under a `create` subcommand, leaving room for `library create resource` future extension. Follows existing patterns where `library` is parent with subcommands (`init`, `resources`, `presets`, `show`).

**Alternatives Considered:**
- `library create-preset` (flat): Would require separate subcommand for each creation type
- `library preset <name>` (verb-less): Less clear intent, conflates read (show) and write (create)

### 2. Persistence: YAML Rewrite on Save

**Choice:** `SaveLibrary()` re-marshals entire library.yaml using `yaml.Marshal()`

**Rationale:** Simple implementation. Current loader uses `yaml.Unmarshal` which doesn't preserve formatting anyway, so round-trip preservation is not possible. Standard YAML marshaling is sufficient.

**Alternatives Considered:**
- YAML library with comment preservation: Over-engineered for v1; formatting changes on save are acceptable
- Partial update (read-modify-write of only presets section): More complex, same end result

### 3. Strict Resource Validation

**Choice:** Fail if any `--resources` ref doesn't exist in library

**Rationale:** Prevents creating broken presets. The preset's value is in referencing valid resources. Fail-fast with clear error message is better than letting `germinator init --preset` fail later.

**Alternatives Considered:**
- Lenient (create anyway): Would cause confusing errors at init time
- Validate resources exist in filesystem: Overkill; library.yaml index is source of truth

### 4. Error on Duplicate Without Force

**Choice:** Return error "preset '<name>' already exists (use --force to overwrite)"

**Rationale:** Matches existing pattern in `config init --force` and `library init --force`. Clear, actionable message.

**Alternatives Considered:**
- Silent overwrite: Dangerous, could lose intended preset
- Interactive prompt: Out of scope for non-interactive design

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| YAML formatting changes on save | Acceptable trade-off; library.yaml is auto-generated file, formatting not semantically important |
| Race condition if library.yaml edited concurrently | Unlikely in CLI usage; add file locking if needed in future |
| Force overwrite could lose intentional changes | User must explicitly pass `--force` flag; error message reminds about force |

## Open Questions

None at this time.
