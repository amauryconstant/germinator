# cli-stdin-composability Specification

## Purpose

Define how the CLI composes with shell pipelines: a `-` filename convention for "read from stdin", and the rule that the CLI SHALL NEVER silently hang on empty stdin in a TTY context.

## Requirements

### Requirement: - filename convention

When a command accepts a file path and the user passes `-` as the path, the CLI SHALL read the file contents from stdin instead of opening a file named `-`. This is the universal Unix convention for "use stdin".

#### Scenario: - reads from stdin

- **GIVEN** the user has YAML on stdin (e.g., piped from `cat`)
- **WHEN** `germinator validate - --platform opencode` is invoked
- **THEN** the CLI SHALL read YAML from stdin
- **AND** parse it as a canonical document
- **AND** validate it against the opencode platform rules

#### Scenario: explicit path opens a file

- **GIVEN** a file `/path/to/file.yaml` exists
- **WHEN** `germinator validate /path/to/file.yaml --platform opencode` is invoked
- **THEN** the CLI SHALL open and read `/path/to/file.yaml`
- **AND** `stdin` SHALL NOT be read

#### Scenario: empty path is treated as -

- **WHEN** a command receives an empty string as the file path (e.g., from a misconfigured script)
- **THEN** the CLI MAY treat empty string as equivalent to `-` (read from stdin)
- **OR** SHALL return a clear error: "no input: provide a file path or pipe data via stdin"

### Requirement: No silent hang on empty stdin

The CLI SHALL NOT silently block on stdin when stdin is a TTY with no data. If stdin is a TTY (i.e., the user is running interactively) and no data is piped, the CLI SHALL either:

- Print a brief prompt or error immediately, OR
- Read stdin with a short timeout and treat EOF as empty input

The CLI SHALL NEVER block indefinitely waiting for a TTY user to type a YAML document.

#### Scenario: Interactive stdin with no input

- **GIVEN** stdin is a TTY (the user invoked the CLI interactively)
- **AND** no data is piped to stdin
- **WHEN** `germinator validate - --platform opencode` is invoked
- **THEN** the CLI SHALL print an error or help text within 1 second
- **AND** exit code SHALL be non-zero
- **AND** the CLI SHALL NOT hang waiting for keyboard input

#### Scenario: Piped empty stdin

- **GIVEN** stdin is a pipe (e.g., `echo "" | germinator validate -`)
- **WHEN** the CLI reads from stdin
- **THEN** it SHALL treat EOF immediately as empty input
- **AND** return an error: `no input: empty stdin`
- **AND** exit code SHALL be non-zero

#### Scenario: Piped non-empty stdin

- **GIVEN** stdin is a pipe with valid YAML
- **WHEN** `cat doc.yaml | germinator validate - --platform opencode` is invoked
- **THEN** the CLI SHALL read until EOF
- **AND** validate the streamed YAML
- **AND** exit code SHALL be 0 on success

### Requirement: Stdin read has a sensible default chunker

For large YAML inputs from stdin, the CLI SHALL stream-read in chunks (default 64 KiB) rather than buffering the entire input. Document loading (`internal/parser/LoadDocument`) is responsible for the actual decode.

#### Scenario: Large stdin input

- **GIVEN** a 10 MB YAML document is piped via stdin
- **WHEN** `germinator validate - --platform opencode` is invoked
- **THEN** the CLI SHALL stream-read in 64 KiB chunks
- **AND** memory usage SHALL NOT exceed a small multiple of the chunk size plus the decoded document size

### Requirement: Stderr is unaffected

Reading from stdin SHALL NOT produce output on `ErrOut` for the normal success path. Diagnostic output (verbose progress, errors) MAY appear on `ErrOut` but SHALL NOT interleave with the read loop.

#### Scenario: stdin read keeps stderr quiet on success

- **GIVEN** stdin is piped with a valid document
- **WHEN** `germinator validate - --platform opencode` is invoked (without `--verbose`)
- **THEN** `ErrOut` SHALL be empty (no progress noise)
- **AND** `Out` SHALL contain the result

### Requirement: Adapt and validate accept stdin

The `adapt` and `validate` commands SHALL accept `-` as the input path argument. Canonicalize MAY accept `-` as the input path argument for round-trip symmetry (parse platform doc from stdin, emit canonical YAML on stdout).

#### Scenario: adapt reads canonical YAML from stdin

- **WHEN** `cat source.yaml | germinator adapt - /tmp/out.md --platform opencode` is invoked
- **THEN** the CLI SHALL read canonical YAML from stdin
- **AND** write the rendered platform file to `/tmp/out.md`
- **AND** exit code SHALL be 0 on success
