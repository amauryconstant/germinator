---
opencode-version: 1.0
created: 2026-03-24
author: amaury
---

# Proposal: Extract Infrastructure Interfaces

## Why

The service layer currently directly instantiates infrastructure dependencies (`parsing.LoadDocument`, `serialization.RenderDocument`), making unit testing impossible without the full parsing/serialization stack. Services cannot be tested in isolation because they hard-code concrete infrastructure calls rather than depending on interfaces. This blocks effective unit testing of business logic and violates the dependency inversion principle.

## What Changes

- Extract `Parser` interface from `parsing.LoadDocument` function
- Extract `Serializer` interface from `serialization.RenderDocument` function
- Inject `Parser` and `Serializer` into `Transformer` and `Initializer` services via constructor
- Update service constructors to accept infrastructure interfaces as parameters
- Create mocks for `Parser` and `Serializer` interfaces to enable unit testing

## Capabilities

### New Capabilities

- `infrastructure-interfaces`: Define `Parser` and `Serializer` interfaces in the application layer for infrastructure abstraction

### Modified Capabilities

- `application/dependency-injection`: Extend to cover infrastructure interface injection into services (not just service interfaces into commands)
- `application/service-contracts`: May need updates if service signatures change to accept infrastructure dependencies

## Impact

- **Affected packages**: `internal/service/`, `internal/application/`, `internal/infrastructure/parsing/`, `internal/infrastructure/serialization/`
- **Breaking**: Service constructors change signatures (`NewTransformer(parser, serializer)` instead of `NewTransformer()`)
- **Test impact**: Enables unit testing of `Transformer` and `Initializer` without filesystem or YAML parsing
- **CLI impact**: None - internal refactor only
- **Note**: `Canonicalizer` is excluded from this change because it uses `ParsePlatformDocument` and `MarshalCanonical` (different infrastructure functions requiring separate interfaces)
