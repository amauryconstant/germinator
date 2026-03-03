## Purpose

Define service interfaces and request/result types for document transformation operations, enabling dependency injection, testability, and clean architecture separation.

## ADDED Requirements

### Requirement: Transformer interface with request/result types

The system SHALL provide a `Transformer` interface in `internal/application/` for document transformation operations.

#### Scenario: TransformRequest contains required fields

- **WHEN** a TransformRequest is created
- **THEN** it SHALL have `InputPath string` field
- **AND** it SHALL have `OutputPath string` field
- **AND** it SHALL have `Platform string` field

#### Scenario: TransformResult contains output information

- **WHEN** a TransformResult is returned
- **THEN** it SHALL have `OutputPath string` field
- **AND** it SHALL have `BytesWritten int` field

#### Scenario: Transformer.Transform accepts context and request

- **WHEN** `Transform(ctx context.Context, req *TransformRequest)` is called
- **THEN** it SHALL return `(*TransformResult, error)`
- **AND** it SHALL write transformed document to OutputPath
- **AND** it SHALL return error on failure (parse, validation, write)

---

### Requirement: Validator interface with request/result types

The system SHALL provide a `Validator` interface in `internal/application/` for document validation operations.

#### Scenario: ValidateRequest contains required fields

- **WHEN** a ValidateRequest is created
- **THEN** it SHALL have `InputPath string` field
- **AND** it SHALL have `Platform string` field

#### Scenario: ValidateResult contains validation errors

- **WHEN** a ValidateResult is returned
- **THEN** it SHALL have `Errors []error` field
- **AND** it SHALL have `Valid() bool` method returning `len(r.Errors) == 0`

#### Scenario: Validator.Validate accepts context and request

- **WHEN** `Validate(ctx context.Context, req *ValidateRequest)` is called
- **THEN** it SHALL return `(*ValidateResult, error)`
- **AND** `error` return SHALL indicate fatal errors (file not found, parse failure)
- **AND** `result.Errors` SHALL contain validation issues

#### Scenario: ValidateResult.Valid method works correctly

- **WHEN** `result.Valid()` is called
- **THEN** it SHALL return `true` if `len(result.Errors) == 0`
- **AND** it SHALL return `false` if `len(result.Errors) > 0`

---

### Requirement: Canonicalizer interface with request/result types

The system SHALL provide a `Canonicalizer` interface in `internal/application/` for platform-to-canonical conversion operations.

#### Scenario: CanonicalizeRequest contains required fields

- **WHEN** a CanonicalizeRequest is created
- **THEN** it SHALL have `InputPath string` field
- **AND** it SHALL have `OutputPath string` field
- **AND** it SHALL have `Platform string` field
- **AND** it SHALL have `DocType string` field

#### Scenario: CanonicalizeResult contains output information

- **WHEN** a CanonicalizeResult is returned
- **THEN** it SHALL have `OutputPath string` field
- **AND** it SHALL have `BytesWritten int` field

#### Scenario: Canonicalizer.Canonicalize accepts context and request

- **WHEN** `Canonicalize(ctx context.Context, req *CanonicalizeRequest)` is called
- **THEN** it SHALL return `(*CanonicalizeResult, error)`
- **AND** it SHALL write canonical YAML to OutputPath
- **AND** it SHALL return error on failure (parse, validation, marshal, write)

---

### Requirement: Initializer interface with request/result types

The system SHALL provide an `Initializer` interface in `internal/application/` for resource installation operations.

#### Scenario: InitializeRequest contains all required fields

- **WHEN** an InitializeRequest is created
- **THEN** it SHALL have `Library *library.Library` field
- **AND** it SHALL have `Platform string` field
- **AND** it SHALL have `OutputDir string` field
- **AND** it SHALL have `DryRun bool` field
- **AND** it SHALL have `Force bool` field
- **AND** it SHALL have `Refs []string` field

#### Scenario: InitializeResult contains per-resource information

- **WHEN** an InitializeResult is returned
- **THEN** it SHALL have `Ref string` field
- **AND** it SHALL have `InputPath string` field
- **AND** it SHALL have `OutputPath string` field
- **AND** it SHALL have `Error error` field

#### Scenario: Initializer.Initialize accepts context and request

- **WHEN** `Initialize(ctx context.Context, req *InitializeRequest)` is called
- **THEN** it SHALL return `([]InitializeResult, error)`
- **AND** it SHALL return partial results even on error (fail-fast with progress)
- **AND** it SHALL process refs in order

#### Scenario: Initialize supports dry-run mode

- **WHEN** `Initialize` is called with `req.DryRun == true`
- **THEN** it SHALL NOT write any files
- **AND** it SHALL return results with InputPath and OutputPath populated

#### Scenario: Initialize supports force mode

- **WHEN** `Initialize` is called with `req.Force == true`
- **THEN** it SHALL overwrite existing files
- **AND** it SHALL NOT return file-exists errors

---

### Requirement: Request/result types live in application package

All request and result types SHALL be defined in `internal/application/` alongside interfaces.

#### Scenario: Requests and results are importable from application package

- **WHEN** a package imports `internal/application`
- **THEN** it SHALL have access to `TransformRequest`, `TransformResult`
- **AND** it SHALL have access to `ValidateRequest`, `ValidateResult`
- **AND** it SHALL have access to `CanonicalizeRequest`, `CanonicalizeResult`
- **AND** it SHALL have access to `InitializeRequest`, `InitializeResult`

---

### Requirement: Services implement interfaces

Concrete service implementations in `internal/services/` SHALL implement the corresponding interfaces.

#### Scenario: Transformer implementation satisfies interface

- **WHEN** `services.NewTransformer()` is called
- **THEN** the returned value SHALL implement `application.Transformer`

#### Scenario: Validator implementation satisfies interface

- **WHEN** `services.NewValidator()` is called
- **THEN** the returned value SHALL implement `application.Validator`

#### Scenario: Canonicalizer implementation satisfies interface

- **WHEN** `services.NewCanonicalizer()` is called
- **THEN** the returned value SHALL implement `application.Canonicalizer`

#### Scenario: Initializer implementation satisfies interface

- **WHEN** `services.NewInitializer()` is called
- **THEN** the returned value SHALL implement `application.Initializer`
