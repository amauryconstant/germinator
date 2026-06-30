**Location**: `cmd/canonical-examples/`
**Parent**: See [`/cmd/AGENTS.md`](../AGENTS.md) for CLI architecture (DI, errors, exit codes, verbosity, lint enforcement)

---

# Canonical Migration Examples

Worked examples of `cmd/` migrations, organized by slice. Each entry
captures the pattern variants introduced in that slice and the
differences from earlier slices. Start with the earliest slice
containing a pattern you need; read forward only if your command
combines multiple patterns.

---

# Canonical examples (slice 2)

The two pilot migrations in slice 2 (`adapt` and `library resources`)
are the canonical references for the new pattern. See:

- `cmd/adapt.go` — `NewCmdAdapt(f, runF)` + `runAdapt(opts)`; uses
  `core.ValidatePlatform`, `opts.IO.Out`, `opts.IO.Verbosef`. Defines
  the `Transformer` interface inline (interfaces where consumed).
- `cmd/resources.go` — `NewCmdResources(f, libraryPath, runF)` +
  `runResources(opts)`; dispatches on `opts.Output` to the JSON or
  table exporter, or plain output via the shared `formatResourcesList`
  helper. The `libraryPath *string` parameter is the parent's
  shared `--library` pointer so the parent's flag value is honored.

---

# Canonical examples (slice 4)

Slice 4 extends the library-command pattern to `presets` and `show`.
Both follow the slice-3 `resources` template, but each carries its
own options struct (the per-command runF is typed to that struct,
not to `*libraryResourcesOptions`). The parent's `NewLibraryCommand`
passes `nil` for the per-command runF (the production wiring is done
by main.go at the composition root, not by the parent). See:

- `cmd/presets.go` — `NewCmdPresets(f, libraryPath, runF)` +
  `runPresets(opts)`. Uses `library.ListPresets` (a package function,
  not a method on `*library.Library`); dispatches on `opts.Output`
  to plain/JSON/table. Owns its own `formatPresetsList`,
  `flattenPresets`, and `presetsRow` types — no shared formatter file.
- `cmd/show.go` — `NewCmdShow(f, libraryPath, runF)` + `runShow(opts)`.
  Resolves the ref via `strings.HasPrefix(opts.Ref, "preset/")` +
  `strings.TrimPrefix` for presets, and `library.ParseRef` for
  resources; on miss returns `core.NewNotFoundError("library ref", opts.Ref)`
  which `output.FormatError` renders as `Error: not found: <ref>` to
  stderr. Owns its own `formatResourceDetails`, `formatPresetDetails`,
  and `showResourceRow`/`showPresetRow` types.

The pattern of per-command `runF` (typed to the command's options
struct) and the shared `(f, libraryPath *string, runF)` signature
holds for all current and future library sub-commands (`add`, `create`,
`init`, `refresh`, `remove`, `validate` migrate in slice 6/7). The
parent never threads the `runF` between sub-commands because the
option types differ.

---

# Canonical example (slice 5)

Slice 5 migrates the top-level `init` command — the first command
with a per-resource partial-success aggregate — to the
`NewCmdXxx(f, runF)` pattern. Key differences from the slice-2/4
templates:

- **No parent `--library` sharing.** Top-level `init` owns its own
  `--library` flag via `initLibrary(f, explicitPath)`; the lazy loader
  resolves per-call so changes to `--library` between invocations
  are honored.
- **`--output-dir` rename (breaking).** The legacy `--output` / `-o`
  flag is gone; the new flag is `--output-dir` only.
- **`--resources` is `StringSliceVar`.** No comma-splitting is done
  in the run body.
- **`Initializer` lazy field.** `initInitializer(f)` returns the
  Factory's `application.Initializer` lazy field; nil field yields
  nil (runInit surfaces an error rather than a nil dereference).
- **Partial success.** `runInit` returns `*core.PartialSuccessError`
  on partial / all-failed outcomes; `cmdutil.ExitCodeFor` returns 0
  for `Succeeded > 0` and 1 for `Succeeded == 0`. Preset-not-found
  returns `*core.NotFoundError` → exit 2 (mapped by the slice-5 §5.0.1
  extension to `cmdutil.ExitCodeFor`).
- **Per-resource errors.** `renderResults` writes successes to
  `opts.IO.Out` and failures to `opts.IO.ErrOut` via
  `output.FormatError(io, *core.InitializeError)`.

See:
- `cmd/init.go` — `NewCmdInit(f, runF)` + `runInit(opts)`;
  uses `(*library.Library).ResolvePreset` (slice-5 §5.0.2) and
  `core.ValidatePlatform` for validation.

- `cmd/legacy_bridge.go` — `LegacyBridge` shim (transitional; slice 7
  deletes it). `legacyCfgFrom(bridge)` builds the per-command
  `CommandConfig` consumed by non-migrated commands during the
  migration window.

---

# Canonical example (slice 6)

Slice 6 migrates the mutating library commands `library add` (three
modes + a legacy `--batch` mode for explicit files) and
`library create preset`. The migrated files stay **flat** in `cmd/`
(per Decision 7 in `openspec/changes/migrate-library-add-create/design.md`)
— no `cmd/library/` subdirectory is created. Key differences from the
slice-2/4/5 templates:

- **Three (plus one legacy) modes in `library add`.** Mode dispatch
  happens in `runAdd` based on the `Discover` and `Batch` flags:
  - **Mode 1 (explicit files):** for each `InputPath`, call
    `lib.AddResource(opts.Ctx, ...)`; collect into a partial-success
    aggregate.
  - **Mode 2 (`--discover`):** call `lib.DiscoverOrphans(opts.Ctx, ...)`;
    for each orphan, validate ref via `core.CanInstallResource`; on
    `name_conflict`, record `*core.OperationError{Op: "register",
    Resource: <ref>, Cause: library.ErrNameConflict}` and increment
    `PartialSuccessError.Failed`. Return `*core.PartialSuccessError`
    on partial success.
  - **Mode 3 (`--discover --batch --force`):** continuous loop; on
    cancellation, collect partial successes and return wrapped
    `ctx.Err()`.
  - **Mode 4 (legacy `--batch` with explicit InputPaths):** routed
    through `library.BatchAddResources`; preserved for the
    pre-change behavior that `e2e/library_add_test.go` still
    exercises. Not in the new spec but kept for compat.
- **Per-resource `name_conflict` is distinct from skip.** Conflicts
  produce a `*core.OperationError` wrapping `library.ErrNameConflict`
  as the `Cause` and count as failures (matches the pre-change
  semantics; see design Decision 3). `errors.Is(err,
  library.ErrNameConflict)` works through the chain.
- **`Ctx context.Context` in `addOptions`.** Threaded into every
  call to `library.DiscoverOrphans`, `library.BatchAddResources`,
  `library.LoadLibrary`. Cancellation during batch mode surfaces
  as wrapped `ctx.Err()` with partial results.
- **`--output` only on `library add`.** `library create preset` does
  not get `--output` (legacy did not have `--json`). Plain output
  is byte-identical to the legacy `library add` output per design
  Decision 9 (golden-file pinned at
  `cmd/testdata/library_add_plain.golden`); JSON output uses the
  net-new `discoverJSONPayload`/`explicitJSONPayload` shapes via
  `output.NewJSONExporter`; table output uses `tab:"HEADER"` struct
  tags via `output.NewTableExporter`. Spec delta:
  `openspec/changes/migrate-library-add-create/specs/library-library-json-output/spec.md`.
- **`library create` collapses to a leaf** via a thin routing
  parent at `cmd/library.go:56-65`. The pre-change
  `NewLibraryCreateCommand` Cobra group wrapper is deleted (per
  design Decision 8); `NewCmdCreatePreset` is registered under a
  one-line `createCmd` Cobra parent so the user-facing command path
  `germinator library create preset <name> --resources ...` remains
  routable. The thin parent has no `RunE` of its own (it just
  shows the `preset` subcommand when `library create` is invoked
  bare), matching the spec scenario "library create has no
  subcommand list" intent even though the parent exists for
  routing.
- **`core.CanInstallResource` (pure, in `internal/core/rules.go`).**
  String-only ref validation; depguard-compatible (no `library`
  import). The authoritative validation still happens in the
  library; this is a fast-fail check before I/O.
- **Library type renames.** `library.AddOptions` → `library.AddRequest`
  and `library.OrphanInfo` → `library.Orphan` (per design Decision 6)
  to align the public types with the request/result convention.
- **Inline interfaces are named after their behavior**, not after
  the `library.Library` struct they substitute for. The
  `resourceAdder` interface (3 methods: `AddResource`,
  `DiscoverOrphans`, `BatchAddResources`) is satisfied via a
  small `libraryAdapter` shim because the library package exposes
  package-level functions, not methods on `*Library`. The
  `presetWriter` interface (1 method: `CreatePreset`) is satisfied
  directly by `*library.Library` because `CreatePreset` is also
  a method on `*Library` (introduced alongside the package-level
  function for symmetry). Compile-time interface checks at
  `cmd/library_add.go` and `cmd/library_create.go`:
  ```go
  var _ resourceAdder = (*libraryAdapter)(nil)
  var _ presetWriter  = (*library.Library)(nil)
  ```

See:
- `cmd/library_add.go` (rewritten in place) — `NewCmdAdd(f,
  libraryPath, runF)` + `runAdd(opts)` dispatcher → `runAddExplicit`
  (Mode 1), `runAddBatchFiles` (Mode 4 legacy), or
  `runAddDiscover` (Modes 2/3). Uses `core.CanInstallResource` for
  ref validation; partial-success aggregation via
  `*core.PartialSuccessError`; per-resource errors via
  `*core.OperationError`.
- `cmd/library_create.go` (rewritten in place; group wrapper
  removed) — `NewCmdCreatePreset(f, libraryPath, runF)` +
  `runCreatePreset(opts)`. Pre-flight validation via
  `core.CanInstallResource`; empty `--resources ""` mapped to a
  Cobra-style positional-arg error so `cmdutil.ExitCodeFor`
  returns 2 via the `cobraUsagePrefixes` branch.
- `cmd/library.go` — rewires `NewLibraryAddCommand(bridge, ...)` to
  `NewCmdAdd(f, ...)`; the `library create` parent is a thin
  `createCmd` (lines 56-65) wrapping `NewCmdCreatePreset(f, ...)`.
- `internal/library/adder.go` — adds `ctx context.Context` to
  `AddResource`, `BatchAddResources`, `DiscoverOrphans`; renames
  `AddOptions` → `AddRequest` and `OrphanInfo` → `Orphan`; adds
  `library.ErrNameConflict` sentinel; renames `hasNameConflict` →
  `checkNameConflict` (now returns `error` instead of `bool`).
- `internal/library/loader.go` — adds `ctx context.Context` to
  `LoadLibrary`.
- `internal/library/creator.go` — adds `library.CreatePresetRequest`
  type + `library.CreatePreset` package function +
  `(*library.Library).CreatePreset` method (symmetric form; the
  method lets `*library.Library` satisfy `presetWriter` without an
  adapter shim).
