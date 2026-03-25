## 1. Define Infrastructure Interfaces

- [x] 1.1 Add `Parser` interface to `internal/application/interfaces.go`
- [x] 1.2 Add `Serializer` interface to `internal/application/interfaces.go`

## 2. Create Infrastructure Adapters

- [x] 2.1 Create `parsingParser` struct implementing `Parser` in `internal/infrastructure/parsing/`
- [x] 2.2 Create `serializationSerializer` struct implementing `Serializer` in `internal/infrastructure/serialization/`
- [x] 2.3 Add compile-time interface checks for adapters

## 3. Update Service Constructors

- [x] 3.1 Update `NewTransformer(parser Parser, serializer Serializer)` signature in `internal/service/`
- [x] 3.2 Update `NewInitializer(parser Parser, serializer Serializer)` signature in `internal/service/`
- [x] 3.3 Update `Transformer` implementation to use injected interfaces instead of direct calls
- [x] 3.4 Update `Initializer` implementation to use injected interfaces instead of direct calls
- [x] 3.5 Add compile-time interface checks for services

## 4. Update ServiceContainer Wiring

- [x] 4.1 Update `NewServiceContainer()` to create Parser and Serializer instances
- [x] 4.2 Pass infrastructure to `NewTransformer` and `NewInitializer` constructors
- [x] 4.3 Verify all services compile and wire correctly (Canonicalizer unchanged)

## 5. Create Mocks for Testing

- [x] 5.1 Create `MockParser` in `test/mocks/parser_mock.go`
- [x] 5.2 Create `MockSerializer` in `test/mocks/serializer_mock.go`

## 6. Add Unit Tests with Mocks

- [x] 6.1 Add unit test for Transformer using MockParser and MockSerializer
- [x] 6.2 Add unit test for Initializer using MockParser and MockSerializer

## 7. Verify and Clean Up

- [x] 7.1 Run `mise run lint` and fix any issues
- [x] 7.2 Run `mise run test` and verify all tests pass
- [x] 7.3 Run `mise run build` and verify binary builds
