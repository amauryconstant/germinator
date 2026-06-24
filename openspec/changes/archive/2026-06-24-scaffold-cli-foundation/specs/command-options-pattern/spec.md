# command-options-pattern Specification

## Purpose

Define the canonical command-file template: each command declares its own `Options` struct, a constructor `NewCmdXxx(f *Factory, runF func(*XxxOptions) error)`, and a `runXxx(opts)` function. The `runF` parameter enables test injection of the `run` function without a Cobra invocation.

## ADDED Requirements

### Requirement: Options struct per command

Each command file SHALL declare its own `XxxOptions` struct holding parsed flags, the `*iostreams.IOStreams`, lazy dependency functions, and a context.

#### Scenario: Options struct fields

- **GIVEN** a command file `cmd/<name>.go`
- **WHEN** the file is inspected
- **THEN** it SHALL declare a `type <name>Options struct` with at least:
  - `IO *iostreams.IOStreams`
  - `Ctx context.Context`
  - one or more lazy dependency fields typed `func() (T, error)` matching the command's needs
  - one field per Cobra flag (string, bool, []string, etc.)

### Requirement: NewCmdXxx constructor signature

Each command file SHALL export a `NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command` constructor.

#### Scenario: Constructor signature

- **WHEN** a command file is inspected
- **THEN** it SHALL export a `NewCmdXxx` function with exactly two parameters: `f *cmdutil.Factory` and `runF func(*XxxOptions) error`
- **AND** it SHALL return `*cobra.Command`

#### Scenario: RunE populates Options and calls runF

- **GIVEN** a command file's `NewCmdXxx(f, runF)` constructor
- **WHEN** Cobra invokes the command's `RunE` closure
- **THEN** the closure SHALL construct an `XxxOptions` struct populated from `f` (for `IO`, `Ctx`, lazy fns) and from the parsed flags
- **AND** the closure SHALL call `runF(opts)` if `runF != nil`, otherwise `runXxx(opts)`

### Requirement: runXxx function

Each command file SHALL define a package-private `runXxx(opts *XxxOptions) error` function that contains the command's actual logic.

#### Scenario: runXxx is package-private

- **WHEN** the command file is inspected
- **THEN** the `runXxx` function SHALL be lowercase (package-private)
- **AND** it SHALL take exactly one parameter: `opts *XxxOptions`
- **AND** it SHALL return `error`

#### Scenario: runXxx is testable directly

- **GIVEN** a test wants to exercise a command's logic without invoking Cobra
- **WHEN** the test constructs an `XxxOptions` struct directly and calls `runXxx(opts)`
- **THEN** the command's logic SHALL execute
- **AND** the test SHALL NOT need a Cobra invocation or a real Factory

### Requirement: Interface declarations live where consumed

Each command file SHALL declare the tiny interface it needs (e.g. `type Transformer interface { Transform(ctx, *TransformRequest) (*core.TransformResult, error) }`).

#### Scenario: Interface declared in command file

- **WHEN** a command file declares a lazy dependency field (e.g. `Transformer func() (Transformer, error)`)
- **THEN** the `Transformer` interface SHALL be declared in the same file as the command
- **AND** the interface SHALL expose only the methods the command actually calls
- **AND** the interface SHALL NOT expose the full concrete type's API

#### Scenario: Production wiring in main.go

- **GIVEN** `main.go` populates the Factory's `Transformer` function
- **WHEN** the function is called
- **THEN** it SHALL construct the concrete service and return it as the interface type
- **AND** the concrete type's package SHALL be imported only by `main.go` (for production wiring) and by the corresponding `*_test.go` file of the concrete implementation — never by any command file

### Requirement: No mutable shared state

Command files SHALL NOT use any package-level mutable state. The only state is on `opts`.

#### Scenario: No package-level variables

- **WHEN** a command file under `cmd/` is inspected
- **THEN** it SHALL NOT declare any package-level `var` (except unexported constants)
- **AND** it SHALL NOT call any `init()` function for registration purposes

### Requirement: Context flows from Factory.RootContext

`opts.Ctx` SHALL be set from `c.Context()` (the signal-aware context set on the root command in `main.go` via `rootCmd.SetContext(ctx)`).

#### Scenario: opts.Ctx is signal-aware

- **GIVEN** `main.go` sets `rootCmd.SetContext(signalAwareCtx)`
- **WHEN** a command's `RunE` populates `opts`
- **THEN** `opts.Ctx` SHALL be set from `c.Context()`
- **AND** it SHALL be cancelled on SIGINT/SIGTERM
- **AND** it SHALL NOT be `context.Background()`
