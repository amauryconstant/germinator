# library-library-json-output Specification (delta)

> **Stream contract:** All `--output` formats (plain, json, table) write primary data to **stdout** (`opts.IO.Out`). Verbose progress writes to **stderr** (`opts.IO.ErrOut` via `opts.IO.Verbosef`). Errors write to **stderr** via `output.FormatError`. Never mix diagnostic output into stdout — this preserves `germinator library resources --output json | jq '.'`.

## REMOVED Requirements

### Requirement: Library parent command accepts --json flag

The `germinator library` parent command SHALL accept a `--json` flag that is inherited by all subcommands.

**Reason:** The parent-inherited `--json` flag mechanism is replaced by per-command `--output json|table|plain` flags. Each library sub-command now opts in to structured output via `cmdutil.AddOutputFlags`. The migration is staged: this change adds `--output` to `library resources`; subsequent changes (per the status blockquote at the bottom of this spec) add it to the remaining library commands.

### Requirement: Library resources outputs JSON when --json is set

The `germinator library resources --json` command SHALL output JSON format when the `--json` flag is set.

**Reason:** Superseded by `## MODIFIED Requirements` → "library resources supports --output flag" below. The `--json` flag is rejected as an unknown flag (see scenario "Old --json flag is rejected").

### Requirement: Library presets outputs JSON when --json is set

The `germinator library presets --json` command SHALL output JSON format when the `--json` flag is set.

**Reason:** `library presets` is migrated in change-4; its `--output` flag lands with that change's delta spec.

### Requirement: Library remove outputs JSON when --json is set

The `germinator library remove` subcommands SHALL output JSON format when the `--json` flag is set.

**Reason:** `library remove` is migrated in change-6; its `--output` flag lands with that change's delta spec.

### Requirement: Library add outputs JSON when --json is set

The `germinator library add --json` command SHALL output JSON format when the `--json` flag is set.

**Reason:** `library add` is migrated in change-6; its `--output` flag lands with that change's delta spec.

### Requirement: Library show outputs JSON when --json is set

The `germinator library show <ref> --json` command SHALL output JSON format when the `--json` flag is set.

**Reason:** `library show` is migrated in change-4; its `--output` flag lands with that change's delta spec.

### Requirement: Library init outputs JSON when --json is set

The `germinator library init --json` command SHALL output JSON format when the `--json` flag is set.

**Reason:** `library init` is migrated in change-7 (cleanup-and-finalize); its `--output` flag lands with that change's delta spec.

## MODIFIED Requirements

### Requirement: library resources supports --output flag

The `library resources` command SHALL expose a per-command string flag accepting the values `json`, `table`, or `plain` (default `plain`).

#### Scenario: Plain is the default

- **WHEN** `germinator library resources` is invoked without `--output`
- **THEN** the output SHALL be plain text (byte-identical to the pre-change `library resources` output)
- **AND** the output SHALL be written to **stdout** (`opts.IO.Out`)

#### Scenario: JSON output via --output json

- **WHEN** `germinator library resources --output json` is invoked
- **THEN** the output SHALL be JSON-formatted (2-space indent, trailing newline) via `output.NewJSONExporter().Write(opts.IO, data)`
- **AND** the JSON output SHALL be written to **stdout**
- **AND** the JSON structure SHALL match the previous `--json` flag output: `{"resources": [{"type": "...", "name": "...", "description": "...", "path": "..."}, ...]}`

#### Scenario: Table output via --output table

- **WHEN** `germinator library resources --output table` is invoked
- **THEN** the output SHALL be a tab-aligned table via `output.NewTableExporter().Write(opts.IO, data)`
- **AND** the table SHALL be written to **stdout**
- **AND** the column order SHALL be derived from `tab:"HEADER"` struct tags on the resource type, falling back to field names when no tag is present (per `internal/output/exporter.go`)

#### Scenario: Old --json flag is rejected

- **WHEN** `germinator library resources --json` is invoked
- **THEN** the command SHALL return a usage error
- **AND** the process SHALL exit with code 2 (`ExitCodeUsage`)

> **Status (this change):** the `--output` flag is added to `library resources` only. Other library commands (`presets`, `show`, `add`, `create`, `init`, `refresh`, `remove`, `validate`) get `--output` in their respective migration changes.
