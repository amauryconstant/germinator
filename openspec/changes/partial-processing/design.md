## Context

The initializer service (`internal/service/initializer.go`) handles `germinator init`, which installs resources from the library to a target project. Currently, it uses fail-fast error handling—returning immediately on the first error (lines 42, 51, 58, 67, 82, 90, 98, 105).

This makes batch operations frustrating: users with multiple resources must fix errors one at a time, re-running after each fix to discover the next issue.

## Goals / Non-Goals

**Goals:**
- Process all resources in a request, collecting individual results
- Continue on error (partial processing) rather than fail-fast
- Return all results including successes and failures
- Maintain backward compatibility for single-resource operations
- Provide clear summary output indicating success/failure counts

**Non-Goals:**
- Do not change the `InitializeResult` structure (already supports per-resource errors)
- Do not modify the CLI command structure (same flags, same usage)
- Do not add retry logic or exponential backoff
- Do not change library resource resolution (that remains fail-fast on missing resources)

## Decisions

### 1. Remove early returns in the processing loop

**Choice:** Replace all `return results, fmt.Errorf(...)` statements with `results = append(results, result); continue`

**Rationale:** The current code has 8 early return points. Changing them to append the result with the error and continue allows the loop to process all resources.

**Alternatives considered:**
- Collect errors in a separate slice and return at end: Adds complexity, the result structure already has an Error field
- Return error only if ALL resources fail: Could be confusing; users expect to know when any failure occurs

### 2. Track success/failure count for exit status

**Choice:** Return `nil` error if at least one resource succeeded; return error only if ALL resources failed

**Rationale:** Aligns with user expectation—partial success should not be an error condition for the command.

**Alternatives considered:**
- Always return error if any failure: Too strict, defeats the purpose of partial processing
- Always return nil: Makes it impossible to detect total failure in scripts without parsing output

### 3. No new result structure needed

**Choice:** Use existing `[]InitializeResult` with per-resource Error field

**Rationale:** The `InitializeResult` struct (Ref, InputPath, OutputPath, Error) already supports per-resource error reporting. The service signature returns `([]InitializeResult, error)`—only the error return value behavior changes.

**Alternatives considered:**
- Add SuccessCount/FailureCount to a wrapper struct: Unnecessary, the results slice provides this information
- Change to a map[string]InitializeResult: Breaking change to consumers, no benefit

### 4. CLI output unchanged at service level

**Choice:** CLI layer (cmd/init.go) will be updated to display summary

**Rationale:** The service returns results identically, just with different error behavior. The CLI handles formatting and output decisions.

**Alternatives considered:**
- Embed summary in service result: Violates separation of concerns
- Return error containing summary: Makes programmatic consumption harder

## Risks / Trade-offs

**[Risk] Breaking scripts that parse error on partial failure**
→ **Mitigation:** Scripts using `--json` flag get structured results where individual errors are visible. Scripts relying on non-zero exit code should check if ALL resources failed.

**[Risk] Users might not notice failures in large batches**
→ **Mitigation:** CLI output will show clear summary: "Initialized 2 resources, 1 failed" with individual error messages.

**[Risk] Some errors are not recoverable (e.g., missing library)**
→ **Mitigation:** Library resolution errors (resource not found) are per-resource and will be collected like others. Infrastructure errors (disk full, permissions) are also collected per-resource.

**[Trade-off] Simpler code vs. detailed error reporting**
→ The new code is actually simpler—fewer nested error handling paths, just one loop with continue on error.

**[Trade-off] Backward compatibility**
→ Single-resource calls behave identically (result with error or success). Batch calls now continue instead of stopping—users who relied on fail-fast may need to adjust.
