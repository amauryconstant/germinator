# Design — Migrate library presets and library show

## Context

After changes 2 and 3 migrate the core domain commands and `library resources`, change-4 migrates the two remaining read-only library commands (`library presets` and `library show`). These commands have no `--json` legacy flag (they previously printed plain text only); the new `--output` flag is added fresh.

Slice-1's `internal/AGENTS.md` line 31 lists `NotFoundError` as a typed domain error, but no implementation exists in `internal/core/errors.go`. Task group 4.0 introduces the type as a precondition in this slice (rather than a separate foundation slice) because `library show`'s not-found behavior depends on it; without it, the delta spec's "Ref not found" scenario is unsatisfiable.

## Goals / Non-Goals

**Goals:**

- `cmd/presets.go` and `cmd/show.go` follow the `NewCmdXxx(f, libraryPath, runF) + runXxx(opts)` pattern (note the `libraryPath *string` second parameter; matches `NewCmdResources` at `cmd/resources.go:47`).
- Both commands gain `--output json|table|plain` via `cmdutil.AddOutputFlags`.
- `library show` resolves refs correctly (resource or preset) and returns `*core.NotFoundError` on miss.
- Six formatters move from `cmd/library_formatters.go` into the per-command files; the file itself stays alive until slice-7.
- `cmd/library_formatters.go` is NOT deleted in this slice (consumed by slice-3 `cmd/resources.go` and slice-6 `cmd/library_add.go`/`cmd/library_create.go`).
- `core.NotFoundError` is introduced as a typed domain error (task group 4.0) and wired into `output.FormatError`.
- Output remains byte-identical for the default `plain` format.

**Non-Goals:**

- Migrating mutating library commands — changes 6, 7.
- Migrating `init` — change-5 (partial-success semantics).
- Restructuring the `Library` package (adding methods to `*library.Library`) — deferred; the package-function API stays as-is.
- Deleting `cmd/library_formatters.go` — slice-7 (after `library_add`/`library_create` migrate).
- A separate foundation slice for `core.NotFoundError` — deferred; folded into this slice as task group 4.0.

## Decisions

### 1. No `Library` interface in command files; concrete `*library.Library` + package functions

**Choice**: Both `presets.go` and `show.go` use the concrete `*library.Library` returned by `opts.Library()` and call package functions directly. `presets.go` calls `library.ListPresets(lib)` (a package function at `internal/library/lister.go:41`). `show.go` calls `library.ParseRef(opts.Ref)` for resource refs and direct `lib.Presets[name]` map lookup for preset refs. **Neither file declares a `Library` interface.**

**Rationale**: matches slice-3's documented precedent verbatim. Slice-3 archive task 2.3.3 states explicitly: *"Do not declare a `Library` interface in this file; use the concrete `*library.Library` returned from `opts.Library()` and call `library.ListResources(lib)` directly. The returned `map[string][]ResourceInfo` is then flattened into a slice of structs (with `tab:"HEADER"` struct tags for the table exporter) before being passed to the exporters."* The reason is that the helpers (`ListResources`, `ListPresets`, `ParseRef`) are package functions, not methods on `*library.Library`. An interface would require inventing methods that don't exist.

**Alternatives considered**:

- Declare per-command `Library` interfaces → rejected; would require either (a) adding methods to `*library.Library` (a foundation-unit change, out of scope), or (b) wrapping the package functions in interface methods that exist only in the cmd package (a leaky abstraction that contradicts `golang-cli-architecture` principle 8: "interfaces where consumed", and would only have one implementation — the production `*library.Library`).
- Reference `cli-command-options-pattern` "interfaces where consumed" → considered; rejected because the capability's spec scenario ("Interface declared in command file") presumes the methods exist. They don't, in this slice. If a future slice adds `(*Library).Resolve(...)` as a method, the interface pattern becomes viable.

### 2. Partial move of `cmd/library_formatters.go`; file kept until slice-7

**Choice**: Six helpers move to per-command files. The file itself is **not deleted** in this slice.

- `cmd/presets.go` (from `cmd/library_formatters.go`): `formatPresetsList` (~22 LOC), `outputPresetsJSON` (~20 LOC), `PresetsJSONOutput` (~3 LOC), `PresetInfoJSON` (~5 LOC).
- `cmd/show.go` (from `cmd/library_formatters.go`): `formatResourceDetails` (~25 LOC), `formatPresetDetails` (~18 LOC), `outputShowResourceJSON` (~28 LOC), `outputShowPresetJSON` (~20 LOC), `ShowResourceJSONOutput` (~5 LOC), `ShowPresetJSONOutput` (~5 LOC).
- **Remain in `cmd/library_formatters.go` until slice-7**: `formatResourcesList` (used by `cmd/resources.go:119`, slice-3), `formatPresetOutput` (used by `cmd/library_create.go:151`, slice-6), `FormatBatchAddSummary` (used by `cmd/library_add.go:311,474`, slice-6).
- Add a one-line TODO at the top of `cmd/library_formatters.go`: `// TODO(slice-7): delete this file once slice-3/slice-6 consumers migrate off the remaining helpers.`

**Rationale**: a single file move (six helpers, six types, ~150 LOC of net change) is the proportional scope. Deleting the file would break three live consumers outside this slice's migration window; deferring the deletion to slice-7 keeps the build green throughout the migration. This is consistent with how slice-3 deferred its own deletions (e.g., `cmd/verbose.go`, `cmd/error_formatter.go`, `internal/service/`, `internal/application/`) to slice-7.

**Alternatives considered**:

- Full deletion in this slice → rejected; would break slice-3 (`cmd/resources.go`) and block slice-6 (`cmd/library_add.go`, `cmd/library_create.go`).
- Move all helpers including the residual ones → rejected; `formatResourcesList` is owned by slice-3's resources migration, and the create/add helpers are owned by slice-6's mutating-command migrations. Cross-slice touches add coordination cost without proportional benefit.
- Move formatters to `internal/output/` → rejected; they're library-specific, not generic output helpers (per slice-3 design Decision 2 reasoning at `archive/2026-06-26-wire-factory-and-pilots/design.md`).

### 3. `library show` ref resolution lives in `cmd/show.go`; brittle legacy check is replaced

**Choice**: The ref resolution logic (parses `type/name` or `preset/name`) moves from `cmd/library.go:111-155` into `cmd/show.go` as a private helper. The brittle legacy string check at `cmd/library.go:129` (`len(ref) > 7 && ref[:7] == "preset/"`) is replaced with `strings.HasPrefix(opts.Ref, "preset/")` + `strings.TrimPrefix`. Resource refs use `library.ParseRef(opts.Ref)` followed by direct `lib.Resources[typ][name]` map lookup (matches the existing `formatResourceDetails` pattern at `cmd/library_formatters.go:83-107`; this slice keeps the lookup local to the command file to avoid inventing a new method on `*library.Library` out of scope for slice-4). Preset refs strip the `preset/` prefix with `strings.TrimPrefix` and look up `lib.Presets[name]`. On miss, both paths return `core.NewNotFoundError("library ref", opts.Ref)`. A future slice may extract a `(*library.Library).Resolve(ref) (any, error)` method that returns `*core.NotFoundError` natively, at which point this command-local lookup is replaced.

**Rationale**: the resolution logic is `show`-specific; encapsulating it in the command file makes the file self-contained. The legacy `len() > 7` check silently misroutes refs shorter than 8 characters (e.g., `preset/a` would not match); the modern `strings.HasPrefix` is the idiomatic replacement per `golang-modernize` skill.

### 4. Add `*core.NotFoundError` as a typed domain error

**Choice**: Introduce `*core.NotFoundError{Entity, Key string}` in `internal/core/errors.go` with constructor `NewNotFoundError(entity, key string) *NotFoundError` and `Error() string` returning `"not found: " + e.Key`. Wire `output.FormatError` (`internal/output/errors.go`) to dispatch on the new type and render `Error: not found: <key>\n` to stderr via the existing `Styles.Error` channel. Update `internal/AGENTS.md` line 31 ("typed domain errors" bullet) to confirm the type exists.

**Rationale**: the type is already documented as a slice-1 deliverable in `internal/AGENTS.md` line 31 (`"Typed domain errors (ValidationError, NotFoundError, OperationError, etc.)"`). The implementation never landed. `library show`'s not-found behavior needs the type to satisfy the delta spec's "Ref not found" scenario. Folding the work into this slice (task group 4.0) avoids creating a separate foundation slice that only touches one type.

**Alternatives considered**:

- Separate foundation slice → rejected; ~30 LOC across two files plus tests is below the threshold that warrants its own change. The slice would have no other consumers until slice-7 (`internal/service/` deletion) at the earliest.
- Use `*core.FileError{IsNotFound(): true}` → rejected; the `FileError` type is for I/O errors (path + operation), not for domain-entity lookups. The `IsNotFound()` method exists, but reusing the type for non-I/O misses is a category error.
- Use plain `fmt.Errorf("not found: %s", ref)` → rejected; loses the typed-error semantics the architecture targets (`errors.As` dispatch in `FormatError`, structural equality in tests). The delta spec scenario explicitly requires `*core.NotFoundError`.

## Risks / Trade-offs

- **Ref resolution edge cases** — unusual ref formats (e.g., empty string, no slash, `preset/` with no name after the prefix) could be lost in the move. **Mitigation:** task 4.2.4 adds table-driven tests covering `""`, `"no-slash"`, `"agent/"`, `"preset/git-workflow"`, `"skill/commit"`, `"nonexistent-ref"`. The legacy `len(ref) > 7` brittle check is replaced with `strings.HasPrefix` to eliminate an existing latent bug (refs shorter than 8 characters were silently misrouted to resource-ref parsing).
- **Partial formatter move leaves residual helpers** — `formatResourcesList`, `formatPresetOutput`, `FormatBatchAddSummary` stay in `cmd/library_formatters.go` because they're consumed by commands not migrated in this slice. **Mitigation:** the TODO comment marks the slice-7 deletion; task 4.4.2 verifies unit tests pass; task 4.4.8 smoke-tests the unmigrated commands end-to-end via `legacyBridge`.
- **Foundation-unit work in a command-migration slice** — `core.NotFoundError` is technically a foundation unit but lands in this slice. **Mitigation:** task group 4.0 is self-contained (~30 LOC + tests); numeric ordering enforces execution before task 4.2; the new type is fully unit-tested in isolation. The alternative (separate foundation slice) was rejected as adding ceremony without proportional benefit.
- **Constructor signature consistency across library sub-commands** — `NewCmdResources(f, libraryPath, runF)` (slice-3), `NewCmdPresets(f, libraryPath, runF)`, `NewCmdShow(f, libraryPath, runF)` all take the shared `libraryPath *string`. The `cmd/AGENTS.md` "Canonical examples" prose mentions `NewCmdXxx(f, runF)` generically; the library-command group is an exception. **Mitigation:** the per-library-command signature pattern is consistent and documented in `cmd/AGENTS.md` (slice-3 update); non-library commands (`NewCmdAdapt`, `NewCmdValidate`) keep the simpler `NewCmdXxx(f, runF)` signature.
