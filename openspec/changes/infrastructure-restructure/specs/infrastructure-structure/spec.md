# infrastructure-structure Specification (Delta)

## Purpose

Organize all infrastructure concerns (parsing, serialization, adapters, config, library) under a unified `internal/infrastructure/` package and rename `services/` to `service/` to match DDD-light naming conventions.

## ADDED Requirements

### Requirement: Infrastructure Package Structure

The system SHALL provide an `internal/infrastructure/` package with subdirectories for infrastructure concerns.

#### Scenario: Infrastructure package exists

- **WHEN** the project structure is inspected
- **THEN** an `internal/infrastructure/` directory SHALL exist
- **AND** it SHALL contain `parsing/`, `serialization/`, `adapters/`, `config/`, `library/` subdirectories

#### Scenario: Parsing subdirectory contains loader and parser

- **WHEN** `internal/infrastructure/parsing/` is inspected
- **THEN** `loader.go` SHALL exist
- **AND** `parser.go` SHALL exist
- **AND** `platform_parser.go` SHALL exist

#### Scenario: Serialization subdirectory contains serializer and templates

- **WHEN** `internal/infrastructure/serialization/` is inspected
- **THEN** `serializer.go` SHALL exist
- **AND** `template_funcs.go` SHALL exist

---

### Requirement: Core Package Migrated

All files from `internal/core/` SHALL be migrated to `internal/infrastructure/`.

#### Scenario: Loader in infrastructure

- **WHEN** `internal/infrastructure/parsing/loader.go` is inspected
- **THEN** it SHALL contain the document loading logic from `internal/core/loader.go`

#### Scenario: Parser in infrastructure

- **WHEN** `internal/infrastructure/parsing/parser.go` is inspected
- **THEN** it SHALL contain the parsing logic from `internal/core/parser.go`

#### Scenario: Serializer in infrastructure

- **WHEN** `internal/infrastructure/serialization/serializer.go` is inspected
- **THEN** it SHALL contain the serialization logic from `internal/core/serializer.go`

---

### Requirement: Adapters Package Migrated

The `internal/adapters/` package SHALL be migrated to `internal/infrastructure/adapters/`.

#### Scenario: Adapter interface in infrastructure

- **WHEN** `internal/infrastructure/adapters/` is inspected
- **THEN** `adapter.go` SHALL exist with Adapter interface
- **AND** `helpers.go` SHALL exist with helper functions

#### Scenario: Claude adapter in infrastructure

- **WHEN** `internal/infrastructure/adapters/claude-code/` is inspected
- **THEN** adapter files SHALL exist with Claude adapter implementation

#### Scenario: OpenCode adapter in infrastructure

- **WHEN** `internal/infrastructure/adapters/opencode/` is inspected
- **THEN** adapter files SHALL exist with OpenCode adapter implementation

---

### Requirement: Config Package Migrated

The `internal/config/` package SHALL be migrated to `internal/infrastructure/config/`.

#### Scenario: Config manager in infrastructure

- **WHEN** `internal/infrastructure/config/` is inspected
- **THEN** config loading and management files SHALL exist
- **AND** imports SHALL reference `internal/infrastructure/config`

---

### Requirement: Library Package Migrated

The `internal/library/` package SHALL be migrated to `internal/infrastructure/library/`.

#### Scenario: Library loader in infrastructure

- **WHEN** `internal/infrastructure/library/` is inspected
- **THEN** library loading and resource management files SHALL exist
- **AND** imports SHALL reference `internal/infrastructure/library`

---

### Requirement: Service Package Renamed

The `internal/services/` package SHALL be renamed to `internal/service/`.

#### Scenario: Service package exists

- **WHEN** the project structure is inspected
- **THEN** `internal/service/` SHALL exist
- **AND** `internal/services/` SHALL NOT exist

#### Scenario: Service files renamed

- **WHEN** `internal/service/` is inspected
- **THEN** all service implementation files SHALL exist
- **AND** package declarations SHALL declare `package service`

---

### Requirement: Import Paths Updated

All import paths SHALL be updated to reflect new package locations.

#### Scenario: Core imports updated

- **WHEN** any file imports from `internal/core`
- **THEN** it SHALL import from `internal/infrastructure/parsing` or `internal/infrastructure/serialization`

#### Scenario: Adapters imports updated

- **WHEN** any file imports from `internal/adapters`
- **THEN** it SHALL import from `internal/infrastructure/adapters`

#### Scenario: Config imports updated

- **WHEN** any file imports from `internal/config`
- **THEN** it SHALL import from `internal/infrastructure/config`

#### Scenario: Library imports updated

- **WHEN** any file imports from `internal/library`
- **THEN** it SHALL import from `internal/infrastructure/library`

#### Scenario: Services imports updated

- **WHEN** any file imports from `internal/services`
- **THEN** it SHALL import from `internal/service`

---

## REMOVED Requirements

### Requirement: Core Package

**Reason:** Split into `infrastructure/parsing/` and `infrastructure/serialization/` for clearer organization.

**Migration:**
- `internal/core/loader.go` → `internal/infrastructure/parsing/loader.go`
- `internal/core/parser.go` → `internal/infrastructure/parsing/parser.go`
- `internal/core/platform_parser.go` → `internal/infrastructure/parsing/platform_parser.go`
- `internal/core/serializer.go` → `internal/infrastructure/serialization/serializer.go`
- `internal/core/template_funcs.go` → `internal/infrastructure/serialization/template_funcs.go`

### Requirement: Separate Adapters Package

**Reason:** Moved under `infrastructure/` for consistent layer organization.

**Migration:**
- `internal/adapters/` → `internal/infrastructure/adapters/`

### Requirement: Separate Config Package

**Reason:** Moved under `infrastructure/` as it's an infrastructure concern.

**Migration:**
- `internal/config/` → `internal/infrastructure/config/`

### Requirement: Separate Library Package

**Reason:** Moved under `infrastructure/` as it's an infrastructure concern.

**Migration:**
- `internal/library/` → `internal/infrastructure/library/`

### Requirement: Services Package (plural)

**Reason:** Renamed to singular `service/` to match DDD naming conventions.

**Migration:**
- `internal/services/` → `internal/service/`
