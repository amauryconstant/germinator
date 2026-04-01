## Context

The `germinator library add` command currently supports adding a single resource at a time. Users migrating existing skill/agent collections or curating large directories need to add multiple files in one command. The current workflow requires running `add` once per file, which is inefficient and creates friction.

**Current behavior:**
- Single source file argument required (unless `--discover` is used)
- Stops on first error
- Exit code reflects first failure

**Constraints:**
- Must not break existing single-file workflow
- Batch mode must be opt-in via `--batch` flag
- Directory scanning must be recursive
- Error resilience: continue processing on error

## Goals / Non-Goals

**Goals:**
- Enable adding multiple files/directories in a single `library add` invocation
- Support recursive directory scanning for `.md` files
- Process all inputs even if some fail (error resilience)
- Always exit 0 for batch operations (partial success = success)
- Provide summary output: "Added N, skipped M, failed K"
- Enable `--json` output on parent `library` command for scripting
- Integrate `--discover --batch` to discover orphans and add all

**Non-Goals:**
- Not changing non-batch behavior (single file add still fails on error)
- Not implementing parallel processing (sequential for simplicity)
- Not adding batch support to other library subcommands

## Decisions

### 1. Add `--batch` flag to `library add` command

**Choice:** Add a `--batch` boolean flag that changes argument parsing from single-file to variadic.

**Rationale:** The batch behavior is fundamentally different (multiple args, error collection, exit code handling). A flag keeps the existing workflow intact while enabling the new mode.

**Alternatives considered:**
- New `library batch-add` subcommand: Would require duplicating flags, more code
- Positional args only (no flag): Ambiguous - couldn't distinguish batch from single-file mode

### 2. Result type structure

**Choice:** `BatchAddResult` with nested `Added`, `Skipped`, `Failed`, and `Summary` slices.

```go
type BatchAddResult struct {
    Added   []AddSuccess  `json:"added"`
    Skipped []SkipInfo    `json:"skipped"`
    Failed  []FailureInfo `json:"failed"`
    Summary BatchSummary  `json:"summary"`
}

type AddSuccess   { Ref, Path string }
type SkipInfo     { Source, Issue string }  // Issue: "conflict", "already_exists"
type FailureInfo  { Source, Error string }
type BatchSummary { Total, Added, Skipped, Failed int }
```

**Rationale:** Clear categorization matches user mental model. JSON structure enables scripting. `SkipInfo.Issue` distinguishes "already exists" (informational) from "conflict" (name collision).

**Alternatives considered:**
- Flat result with status enum per file: Less explicit, harder to parse
- Only returning errors: Loses the "skipped" distinction (user wants to know what was intentionally skipped vs failed)

### 3. Skip vs Failure distinction

**Choice:** Skip = file processed but not added due to business logic (conflict, already exists). Failure = unexpected error during processing.

**Rationale:** Skip is "expected" and doesn't indicate broken state. Failure indicates something unexpected happened. This matches user expectations: "I know this file might conflict, but I want to see the rest of the batch."

**Implementation:**
- Skip: `library.AddResource` returns `ErrResourceExists` or conflict detected before calling Add
- Failure: File read error, canonicalization failure, validation error, etc.

### 4. Exit code handling

**Choice:** Batch mode always exits 0, even if some failed.

**Rationale:** The user explicitly opted into batch error-resilience mode. Exit 0 with summary output is the contract. Scripts can inspect the result (or use `--json`) to determine if issues occurred.

**Implementation:** In `cmd/library_add.go`, when `--batch` is true, ignore errors from `BatchAddResources()` and always return `nil`.

### 5. Directory scanning

**Choice:** Recursively scan directories for `*.md` files using `filepath.Walk`.

**Rationale:** Common use case is `germinator library add --batch ./skills/` where skills may have nested organization.

**Implementation:** In `BatchAddResources()`, for each path argument:
- If file: process directly
- If directory: `filepath.Walk` to find all `*.md` files

### 6. BatchAddResources placement

**Choice:** Add `BatchAddResources()` function in `internal/infrastructure/library/` alongside existing `AddResource()`.

**Rationale:** Keeps batch logic with other library operations. Allows future reuse by other commands if needed.

**Function signature:**
```go
func BatchAddResources(opts BatchAddOptions) (*BatchAddResult, error)
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| User confusion if batch fails silently on some files | Summary output always shows counts; verbose mode shows details |
| Batch with 1000s of files takes long time | Sequential processing is simple and predictable; no timeout needed for now |
| Directory with mixed file types (non-.md) | Only process `.md` files; skip others silently |
| JSON output on parent command affects all subcommands | Use `--json` persist on `library` command as designed |

## Open Questions

**Q:** Should `--batch` change how `--dry-run` behaves?
**A:** Yes. Dry-run in batch mode should show what would be added/skipped/failed without modifying library.

**Q:** Does batch mode canonicalize platform documents?
**A:** Yes, same as single-file mode. Each file is canonicalized if needed before adding.

**Q:** What order are files processed?
**A:** Sequential in argument order. Directories are scanned depth-first.

**Q:** How does `--force` interact with batch?
**A:** `--force` applies to all files in batch. Existing resources are overwritten.
