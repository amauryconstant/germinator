## Why

The initializer currently uses fail-fast error handling, stopping immediately when any resource fails. This makes it difficult for users to identify all issues in a batch of resources—they must fix one error, re-run, discover the next, and repeat. Users want to see all successes and failures in a single execution.

## What Changes

- **Remove fail-fast behavior**: The initializer will process all resources in the request, regardless of individual errors
- **Collect all results**: Each resource gets its own `InitializeResult` with individual error status
- **Return partial results**: Even if some resources fail, successful resources are still installed
- **Summary reporting**: Exit status reflects overall outcome (success if any, error if all fail)
- **JSON output**: `--json` flag outputs structured results for scripting

## Capabilities

### New Capabilities
<!-- Capabilities being introduced. Replace <name> with kebab-case identifier (e.g., user-auth, data-export, api-rate-limiting). Each creates specs/<name>/spec.md -->
- `partial-initialization`: Continue-on-error processing for resource installation, collecting individual results for each resource

### Modified Capabilities
<!-- Existing capabilities whose REQUIREMENTS are changing (not just implementation).
     Only list here if spec-level behavior changes. Each needs a delta spec file.
     Use existing spec names from openspec/specs/. Leave empty if no requirement changes. -->
- `resource-installation`: Change from fail-fast to partial processing (continue-on-error)

## Impact

- **internal/service/initializer.go**: Remove early returns, collect all results
- **openspec/specs/library/resource-installation/**: Add delta spec for behavior change
- **CLI output**: Update to show per-resource status and summary
- **Tests**: Update to verify partial processing behavior
