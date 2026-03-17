## 1. Testing Infrastructure - Mocks

- [ ] 1.1 Create `test/mocks/` directory
- [ ] 1.2 Create `test/mocks/transformer_mock.go` implementing `application.Transformer`
- [ ] 1.3 Create `test/mocks/validator_mock.go` implementing `application.Validator`
- [ ] 1.4 Create `test/mocks/canonicalizer_mock.go` implementing `application.Canonicalizer`
- [ ] 1.5 Create `test/mocks/initializer_mock.go` implementing `application.Initializer`
- [ ] 1.6 Create `test/mocks/doc.go` with package documentation

## 2. Testing Infrastructure - Helpers

- [ ] 2.1 Create `test/helpers/` directory
- [ ] 2.2 Create `test/helpers/doc.go` with package documentation (directory for future shared utilities)

## 3. Mock Usage Example

- [ ] 3.1 Create example unit test in `cmd/validate_test.go` demonstrating `MockValidator` usage

## 4. Update Documentation

- [ ] 4.1 Create `test/mocks/AGENTS.md` with mock inventory
- [ ] 4.2 Update `test/AGENTS.md` with mock usage patterns (setup with On(), assertions with AssertCalled())
