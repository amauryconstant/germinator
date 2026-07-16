# testing-adapter-golden-tests Specification (new)

## Purpose

Define the byte-equality golden test contract for platform adapters. These tests
exercise the full adapter chain end-to-end and compare the rendered output
byte-for-byte (frontmatter + Markdown body) against a checked-in fixture.
They live in the E2E tier (`test/e2e/`, `//go:build e2e`) because
byte-equality is sensitive to renderer dependency drift (Cobra, sprig,
yaml.v3) and belongs in a separate CI stage from the default suite. The
renderer currently emits YAML frontmatter wrapped around a Markdown body for
both Claude Code and OpenCode (`config/templates/<platform>/agent.tmpl`), so
both fixtures use the `.md` extension.

Per `golang-spf13-cobra` Testing section: cobra accumulates flag state across
`Execute()` calls. The new E2E tests are authored as **Ginkgo `Describe`/`It`
blocks within `package e2e_test`**, NOT standalone `Test*` functions. The
Ginkgo runner (`TestE2E` in `test/e2e/e2e_suite_test.go`) registers a single
test entry point; bare `Test*` functions would either be ignored by Ginkgo
or run alongside it (state-leak risk). Fixtures live under
`test/e2e/fixtures/<platform>/` per the existing E2E convention in
`test/AGENTS.md`.

This is distinct from `testing-round-trip-tests`, which exercises semantic
equality in the default suite.

## ADDED Requirements

### Requirement: Adapter golden test runs as E2E

Adapter golden tests SHALL be authored as **Ginkgo `Describe`/`It` specs** in
`test/e2e/` with `//go:build e2e` and `package e2e_test`. They SHALL use the
existing E2E infrastructure (`e2e.BinaryPath()` from `gexec.Build`,
`helpers.NewGerminatorCLI()`) to invoke `germinator adapt <fixture> <out>
--platform <platform>` against a checked-in fixture, and SHALL compare the
rendered output against `test/e2e/fixtures/<platform>/<fixture>.md`
byte-for-byte.

**Change**: NEW requirement codifying the pattern introduced in change
`harden-tests-and-coverage`. The adapter byte-equality tests live alongside
the existing E2E suite and use the same `e2e` build tag. The pre-change
canonicalize golden tests at `internal/canonicalize/canonicalize_golden_test.go`
use the `golden` build tag and live elsewhere; the new adapter tests
deliberately avoid that pre-existing tag to keep the byte-equality sensitive
checks in a single, gated CI stage.

#### Scenario: OpenCode adapter golden spec

- **GIVEN** a canonical agent fixture with `permissionPolicy=balanced`
- **AND** the fixture at `test/e2e/fixtures/opencode/agent-balanced.md`
- **WHEN** the Ginkgo spec `Describe("OpenCode adapter byte-equality rendering", func() { It("renders permission-balanced agent byte-equally", ...) })` runs under `//go:build e2e`
- **THEN** the rendered OpenCode output SHALL match the fixture byte-for-byte
- **AND** the spec SHALL fail if the rendered output drifts (e.g., a new tool is added to the permission map)

#### Scenario: Claude Code adapter golden spec

- **GIVEN** a canonical agent fixture with `permissionPolicy=balanced`
- **AND** the fixture at `test/e2e/fixtures/claude-code/agent-balanced.md`
- **WHEN** the Ginkgo spec for Claude Code runs under `//go:build e2e`
- **THEN** the rendered Claude Code output SHALL match the fixture byte-for-byte

#### Scenario: Byte-equality golden test build tag

- **WHEN** `go test ./...` is run (default)
- **THEN** adapter golden tests SHALL be skipped (build tag excludes them)
- **WHEN** `go test -tags=e2e ./...` is run
- **THEN** adapter golden tests SHALL run via the Ginkgo runner (`TestE2E`)

#### Scenario: Golden fixture drift detection

- **WHEN** a renderer dependency update (Cobra, sprig, yaml.v3) changes the output format
- **THEN** the adapter golden spec SHALL fail with a clear diff against the fixture
- **AND** the contributor SHALL refresh the fixture via the project's standard E2E golden-refresh mechanism