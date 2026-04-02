## Context

The `library add --discover` command currently performs single-level directory scanning and lacks batch operation support. The existing `DiscoverOrphans` function in `internal/infrastructure/library/adder.go` scans only immediate children of `skills/`, `agents/`, `commands/`, and `memory/` directories, missing nested structures.

The current `DiscoverResult` structure is minimal, missing summary statistics needed for batch integration and clear user output.

## Goals / Non-Goals

**Goals:**
- Add recursive directory scanning to discover orphans in nested subdirectories
- Enhance `DiscoverResult` with summary statistics (`TotalScanned`, `TotalOrphans`)
- Add batch mode support via `--batch` flag for continuous orphan processing
- Improve human-readable and JSON output with descriptive summaries

**Non-Goals:**
- Modifying how resources are registered (use existing `registerOrphan` flow)
- Adding parallel processing for large libraries
- Supporting exclusion patterns for discovery

## Decisions

### Decision 1: Recursive scanning approach

**Choice:** Use `filepath.WalkDir` with `WalkDirFunc` for recursive directory traversal.

**Rationale:** `filepath.WalkDir` is the most efficient approach in Go 1.16+ as it:
- Avoids unnecessary stat calls
- Handles symlinks appropriately (can skip with `WalkDirFunc` returning `filepath.SkipDir`)
- Supports skipping directories explicitly

**Alternatives considered:**
- `filepath.Glob` with `**/*.md` pattern: Simpler but less control over traversal
- Manual `Readdirnames` recursion: More code, same functionality

### Decision 2: Result structure enhancement

**Choice:** Extend existing `DiscoverResult` with `Summary` field rather than replacing it.

**Rationale:**
- Backward compatible with existing consumers
- Avoids breaking changes in any external code
- New `DiscoverSummary` aggregates existing counts plus adds `TotalScanned`

**Proposed structure:**
```go
type DiscoverResult struct {
    Orphans   []OrphanInfo    `json:"orphans"`
    Added     []AddSuccess    `json:"added"`
    Conflicts []ConflictInfo  `json:"conflicts"`
    Summary   DiscoverSummary `json:"summary"`
}

type OrphanInfo struct {
    Path string `json:"path"`
    Type string `json:"type"`
    Name string `json:"name"`
    Issue string `json:"issue,omitempty"` // "name_conflict" or empty
}

type DiscoverSummary struct {
    TotalScanned  int `json:"totalScanned"`
    TotalOrphans  int `json:"totalOrphans"`
    TotalAdded    int `json:"totalAdded"`
    TotalSkipped  int `json:"totalSkipped"`
    TotalFailed   int `json:"totalFailed"`
}
```

### Decision 3: Batch mode implementation

**Choice:** Add `--batch` flag that processes all orphans continuously, skipping errors without stopping.

**Rationale:**
- Follows Unix convention for batch operations
- Allows partial success when some orphans fail validation or conflict
- `--force` continues on conflict (skip conflicting, process rest)

**Behavior:**
- `--discover --batch --force`: Discover all, add all (skip conflicts)
- `--discover --batch` (no force): Report-only, show what would happen
- `--discover --dry-run`: Same as current (show orphans without adding)

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Recursive scan slows down large libraries | Performance | Default to recursive, acceptable for typical library sizes (<1000 files) |
| Breaking JSON output structure | API consumers | New fields are additive, existing fields preserved |
| Orphan in wrong directory type | User confusion | Report `Issue: "type_mismatch"` if detected |

## Open Questions

- Should `--batch` support a limit on failures before stopping? (Not in scope for v1)
- Should nested directory names affect type detection? (No - type still from top-level directory)
