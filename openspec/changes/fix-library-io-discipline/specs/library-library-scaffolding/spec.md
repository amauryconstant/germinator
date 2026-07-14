# library-library-scaffolding Specification (delta)

## MODIFIED Requirements

### Requirement: Support dry-run mode

The system SHALL preview changes without creating files when `--dry-run` is specified. The dry-run preview SHALL be written to the `Stdout io.Writer` field on `CreateOptions` (populated by the cmd layer from `opts.IO.Out`). The library package SHALL NOT write to `os.Stdout` directly; the cmd layer is responsible for selecting the writer.

**Change**: ADDED the "writer field" requirement. The pre-change implementation wrote dry-run output to `os.Stdout` directly, which polluted piped consumers' stdout. The cmd layer now populates `CreateOptions.Stdout` via the delegation chain `cmd/library_init.go:161 → library.Init(ctx, &InitRequest{Stdout: opts.IO.Out, ...}) → creator.go:93 Init → creator.go:101 CreateLibrary(CreateOptions{Stdout: req.Stdout, ...})`. The dry-run output respects the cmd-side I/O discipline (e.g., `--output json` writes JSON to the same writer; `--output plain` writes plain text).

#### Scenario: Dry-run does not create files

- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --dry-run` is executed
- **THEN** no files or directories are created
- **AND** a message is printed to `opts.IO.Out` (via the `Stdout` field on `CreateOptions`) indicating what would be created

#### Scenario: Library package does not write to os.Stdout

- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** the dry-run output SHALL be written via the `Stdout io.Writer` field on `CreateOptions` (gated on `opts.Stdout != nil`)
