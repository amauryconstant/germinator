# testing-round-trip-tests Specification (new)

## Purpose

Define the parse→render round-trip contract for adapter golden tests. The pattern catches adapter drift and tag collisions in canonical struct embedding (e.g., `core.Agent`, `core.Skill`) by exercising both the parser and renderer on real fixtures and verifying that the resulting canonical form preserves all field values.

## ADDED Requirements

### Requirement: Round-trip test pattern

Adapter tests SHALL include a `TestParseRenderRoundTrip` test that:

1. Reads a fixture file from `test/fixtures/<platform>/` (canonical or platform-specific).
2. Parses it via `parser.ParsePlatformDocument(inputPath, platform, docType)`.
3. Marshals the result via `renderer.MarshalCanonical(doc)`.
4. Re-parses the marshaled output via the same parser.
5. Asserts that key fields (`Name`, `Description`, `Mode`, `Temperature`, `Steps`, `Hidden`, `Disabled`, `Tools`, `PermissionPolicy`, etc.) are equal between the two parses (semantic equality, not byte equality).

The test SHALL run under the `golden` build tag (`//go:build golden`) so it is excluded from the default `go test ./...` target but available for CI verification via `go test -tags=golden ./...`.

**Change**: NEW requirement codifying the pattern introduced in change `harden-tests-and-coverage`. Pre-change tests start from struct literals (per the review's D-013); the new pattern reads from real fixtures.

#### Scenario: Round-trip preserves Agent fields

- **GIVEN** a canonical agent fixture at `test/fixtures/canonical/agent.md` with non-default values for `name`, `description`, `mode`, `temperature`, `steps`, `tools`, and `permissionPolicy`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the parse→marshal→re-parse cycle SHALL produce a struct with equal field values
- **AND** the test SHALL fail if any key field differs between the two parses

#### Scenario: Round-trip preserves Skill fields

- **GIVEN** a canonical skill fixture with non-default `name`, `description`, and `tools`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

#### Scenario: Round-trip preserves Command fields

- **GIVEN** a canonical command fixture with non-default `name`, `description`, `template`, and `agent`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

#### Scenario: Round-trip preserves Memory fields

- **GIVEN** a canonical memory fixture with non-default `name`, `description`, and `content`
- **WHEN** `TestParseRenderRoundTrip` runs
- **THEN** the cycle SHALL preserve all fields

#### Scenario: Round-trip test build tag

- **WHEN** `go test ./...` is run (default)
- **THEN** the round-trip test SHALL be skipped (build tag excludes it)
- **WHEN** `go test -tags=golden ./...` is run
- **THEN** the round-trip test SHALL run

#### Scenario: Adapter golden test fixture

- **GIVEN** a canonical agent with `permissionPolicy=balanced`
- **WHEN** `TestOpenCodeAdapter_GoldenPermissionRendering` runs
- **THEN** the rendered OpenCode YAML SHALL match `test/golden/opencode/agent-balanced.yaml` byte-for-byte
- **AND** the test SHALL fail if the YAML drifts (e.g., a new tool is added to the permission map)
