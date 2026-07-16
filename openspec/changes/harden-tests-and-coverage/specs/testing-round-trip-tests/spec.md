# testing-round-trip-tests Specification (new)

## Purpose

Define the parse→render round-trip contract for adapter tests. The pattern
catches adapter drift and tag collisions in canonical struct embedding (e.g.,
`core.Agent`, `core.Skill`) by exercising both the parser and renderer on real
fixtures and verifying that the resulting canonical form preserves all field
values. The canonical round-trip is a **semantic** equality test (key fields
compared), not byte equality. It lives in the default test suite with no build
tag. A second round-trip test (the **platform round-trip**) exercises the
forward and reverse adapters on a platform fixture; it shares the same
semantic-equality expectation but uses the platform parser entry point.

## ADDED Requirements

### Requirement: Canonical round-trip preserves fields

Adapter tests SHALL include a `TestParseRenderRoundTrip` test in `internal/renderer/serializer_test.go` that:

1. Reads a canonical fixture file from `test/fixtures/canonical/agent-<scenario>.md`.
2. Parses it via `parser.ParseDocument(t.Context(), inputPath, "agent")`, returning `*parser.CanonicalAgent`.
3. Marshals the result via `renderer.MarshalCanonical(t.Context(), doc)`.
4. Re-parses the marshaled string by writing it to a temporary file and calling `parser.ParseDocument(t.Context(), tmpPath, "agent")`.
5. Asserts that the canonical structural fields of `core.Agent` are equal between the two parses: `Name`, `Description`, `Tools`, `DisallowedTools`, `PermissionPolicy`, `Behavior.Mode`, `Behavior.Temperature`, `Behavior.Steps`, `Behavior.Prompt`, `Behavior.Hidden`, `Behavior.Disabled`, `Model`, `Targets`, and `Extensions`.

The test SHALL live in the default test suite (no build tag) because semantic equality is deterministic.

**Change**: NEW requirement codifying the pattern introduced in change `harden-tests-and-coverage`. Pre-change tests start from struct literals (per the review's D-013); the new pattern reads from real fixtures and asserts the full canonical field set (which mirrors `core.Agent` at `internal/core/agent.go:18-33`).

#### Scenario: Round-trip preserves Agent fields

- **GIVEN** a canonical agent fixture at `test/fixtures/canonical/agent-permission-balanced.md` with non-default values for `name`, `description`, `mode`, `temperature`, `steps`, `tools`, and `permissionPolicy`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the parse→marshal→re-parse cycle SHALL produce a struct with equal field values
- **AND** the test SHALL fail if any key field differs between the two parses

#### Scenario: Round-trip preserves Skill fields

- **GIVEN** a canonical skill fixture with non-default `name`, `description`, and `tools`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

#### Scenario: Round-trip preserves Command fields

- **GIVEN** a canonical command fixture with non-default `name`, `description`, and `Execution` fields
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

#### Scenario: Round-trip preserves Memory fields

- **GIVEN** a canonical memory fixture with non-default `Paths` and `Content`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

### Requirement: Platform round-trip preserves semantic fields

Adapter tests SHALL include a `TestPlatformRoundTrip` test for each platform (`opencode`, `claude-code`) that:

1. Reads a platform fixture (e.g., `test/fixtures/<platform>/agent-balanced.md`).
2. Parses it via `parser.ParsePlatformDocument(t.Context(), inputPath, "<platform>", "agent")`, returning `*parser.CanonicalAgent`.
3. Marshals the result via `renderer.MarshalCanonical(t.Context(), doc)` to obtain canonical YAML.
4. Re-parses the marshaled canonical YAML by writing it to a temporary file and calling `parser.ParseDocument(t.Context(), tmpPath, "agent")`.
5. Asserts that the canonical structural fields (the same set as the canonical round-trip) are equal between the platform-parsed struct and the re-parsed canonical struct.

The test SHALL live in the default test suite. It validates that no field is lost when a platform document is converted through canonical marshaling.

**Change**: NEW requirement codifying the platform→canonical→canonical round-trip; complements the canonical-only round-trip and exercises the platform adapter path.

#### Scenario: Platform fixture round-trips through canonical

- **GIVEN** a `claude-code` agent fixture at `test/fixtures/claude-code/agent-permission-balanced.md` with a non-default `permissionMode`
- **WHEN** `TestPlatformRoundTrip` runs for `claude-code`
- **THEN** the parsed `PermissionPolicy` and behavior fields SHALL match after the canonical-marshal re-parse
