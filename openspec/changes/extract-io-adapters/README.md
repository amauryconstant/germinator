# extract-io-adapters

Extract 5 I/O adapters (Transformer, Validator, Canonicalizer, Initializer, libraryAdapter) from `cmd/` into dedicated `internal/<x>/` shell packages, completing the Functional Core / Imperative Shell architecture's `cmd/ → internal/<dep>/` pattern.

## What

This change moves the concrete I/O adapter implementations from `cmd/` into 4 new shell packages:

| Package | Purpose | Source |
|---|---|---|
| `internal/validate/` | Document validation with platform rules | `cmd/validate.go:130-218` |
| `internal/canonicalize/` | Canonicalize + validate documents | `cmd/canonicalize.go:158-240` |
| `internal/transform/` | Parse → transform → serialize pipeline | `cmd/transformer.go:33-60` |
| `internal/install/` | Per-ref init orchestration | `cmd/initializer.go:38-125` |

Stage 2 additionally converts `libraryAdapter` into direct `*library.Library` methods (`Add`, `BatchAddResources`, `DiscoverOrphans`), retiring the adapter bridge.

## Stages

- **Stage 1**: Extract `validate` + `canonicalize` (~3h). No public API change.
- **Stage 2**: Convert library adders to `*library.Library` methods (~2h). `libraryAdapter` deleted.
- **Stage 3**: Extract `transform` + `install` (~4h). `cmd/transformer.go` and `cmd/initializer.go` deleted.

Documentation (AGENTS.md updates, spec sync) and archive are handled separately after implementation.

## Impact

- **cmd/ shrinks** by ~421 LOC; each command file becomes focused on parse/execute/respond.
- **4 new shell packages** follow the `internal/library` convention (`Service` interface, `Request`/`Result` types, `NewService` constructor, `AGENTS.md`).
- **`*library.Library`** gains 3 methods, completing the slice-7 forward path.
- **CLI behavior unchanged**; end users see no difference.
- **Test surface additive**; new shell packages get table-driven unit tests with `t.TempDir()` fixtures.

## Key design decisions

- [`design.md:Decision 1`](design.md#1-package-naming-internalinstall-not-internalinit) — `install` (not `init`) avoids Go identifier collision.
- [`design.md:Decision 4`](design.md#4-import-direction-shell-packages-depend-on-internallibrary-not-vice-versa) — shell packages depend on `internal/library`; `library` is a leaf.
- [`design.md:Decision 5`](design.md#5-preserve-runf-injection-seam-per-options-lazy-field-pattern-unchanged) — `runF` injection seam preserved; per-options lazy field pattern unchanged.
