## 1. Define Infrastructure Interfaces

- [ ] 1.1 Add `Parser` interface to `internal/application/interfaces.go`
- [ ] 1.2 Add `Serializer` interface to `internal/application/interfaces.go`

## 2. Create Infrastructure Adapters

- [ ] 2.1 Create `parsingParser` struct implementing `Parser` in `internal/infrastructure/parsing/`
- [ ] 2.2 Create `serializationSerializer` struct implementing `Serializer` in `internal/infrastructure/serialization/`
- [ ] 2.3 Add compile-time interface checks for adapters

## 3. Update Service Constructors

- [ ] 3.1 Update `NewTransformer(parser Parser, serializer Serializer)` signature in `internal/service/`
- [ ] 3.2 Update `NewCanonicalizer(parser Parser, serializer Serializer)` signature in `internal/service/`
- [ ] 3.3 Update `NewInitializer(parser Parser, serializer Serializer)` signature in `internal/service/`
- [ ] 3.4 Update service implementations to use injected interfaces instead of direct calls
- [ ] 3.5 Add compile-time interface checks for services

## 4. Update ServiceContainer Wiring

- [ ] 4.1 Update `NewServiceContainer()` to create Parser and Serializer instances
- [ ] 4.2 Pass infrastructure to service constructors
- [ ] 4.3 Verify all services compile and wire correctly

## 5. Create Mocks for Testing

- [ ] 5.1 Create `MockParser` in `test/mocks/parser_mock.go`
- [ ] 5.2 Create `MockSerializer` in `test/mocks/serializer_mock.go`

## 6. Add Unit Tests with Mocks

- [ ] 6.1 Add unit test for Transformer using MockParser and MockSerializer
- [ ] 6.2 Add unit test for Canonicalizer using MockParser and MockSerializer
- [ ] 6.3 Add unit test for Initializer using MockParser and MockSerializer

## 7. Verify and Clean Up

- [ ] 7.1 Run `mise run lint` and fix any issues
- [ ] 7.2 Run `mise run test` and verify all tests pass
- [ ] 7.3 Run `mise run build` and verify binary builds
- [ ] 7.4 Update AGENTS.md files if needed
