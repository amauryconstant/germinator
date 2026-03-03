## 1. ParseError Enhancement (29 call sites: 9 production, 20 test)

- [ ] 1.1 Make ParseError fields private (path, message, cause)
- [ ] 1.2 Add new ParseError fields (suggestions []string, context string)
- [ ] 1.3 Add ParseError getter methods (Path(), Message(), Cause())
- [ ] 1.4 Add ParseError new getter methods (Suggestions(), Context())
- [ ] 1.5 Add ParseError WithSuggestions() builder (returns new instance)
- [ ] 1.6 Add ParseError WithContext() builder (returns new instance)
- [ ] 1.7 Update ParseError Error() method to use getters
- [ ] 1.8 Update ParseError call sites in internal/core/loader.go (2 sites)
- [ ] 1.9 Update ParseError call sites in internal/services/transformer.go (3 sites)
- [ ] 1.10 Update ParseError call sites in internal/services/canonicalizer.go (2 sites)
- [ ] 1.11 Update ParseError call sites in internal/config/manager.go (1 site)
- [ ] 1.12 Update ParseError call sites in internal/library/library.go (1 site)
- [ ] 1.13 Update ParseError call sites in test files (20 sites)
- [ ] 1.14 Update formatParseError() in cmd/error_formatter.go to use getters
- [ ] 1.15 Add context and suggestions to formatParseError() output (use "Hint:" format)
- [ ] 1.16 Add unit tests for ParseError getters and builders
- [ ] 1.17 Add unit tests for ParseError immutability
- [ ] 1.18 Run tests and verify ParseError changes

## 2. TransformError Enhancement (18 call sites: 3 production, 15 test)

- [ ] 2.1 Make TransformError fields private (operation, platform, message, cause)
- [ ] 2.2 Add new TransformError fields (suggestions []string, context string)
- [ ] 2.3 Add TransformError getter methods (Operation(), Platform(), Message(), Cause())
- [ ] 2.4 Add TransformError new getter methods (Suggestions(), Context())
- [ ] 2.5 Add TransformError WithSuggestions() builder (returns new instance)
- [ ] 2.6 Add TransformError WithContext() builder (returns new instance)
- [ ] 2.7 Update TransformError Error() method to use getters
- [ ] 2.8 Update TransformError call sites in internal/services/transformer.go (1 site)
- [ ] 2.9 Update TransformError call sites in internal/services/canonicalizer.go (1 site)
- [ ] 2.10 Update TransformError call sites in internal/config/config.go (1 site)
- [ ] 2.11 Update TransformError call sites in test files (15 sites)
- [ ] 2.12 Update formatTransformError() in cmd/error_formatter.go to use getters
- [ ] 2.13 Add context and suggestions to formatTransformError() output (use "Hint:" format)
- [ ] 2.14 Add unit tests for TransformError getters and builders
- [ ] 2.15 Add unit tests for TransformError immutability
- [ ] 2.16 Run tests and verify TransformError changes

## 3. FileError Enhancement (17 call sites: 6 production, 11 test)

- [ ] 3.1 Make FileError fields private (path, operation, message, cause)
- [ ] 3.2 Add new FileError fields (suggestions []string, context string)
- [ ] 3.3 Add FileError getter methods (Path(), Operation(), Message(), Cause())
- [ ] 3.4 Add FileError new getter methods (Suggestions(), Context())
- [ ] 3.5 Add FileError WithSuggestions() builder (returns new instance)
- [ ] 3.6 Add FileError WithContext() builder (returns new instance)
- [ ] 3.7 Update FileError Error() method to use getters
- [ ] 3.8 Update FileError call sites in internal/config/manager.go (1 site)
- [ ] 3.9 Update FileError call sites in internal/core/parser.go (1 site)
- [ ] 3.10 Update FileError call sites in internal/services/initializer.go (2 sites)
- [ ] 3.11 Update FileError call sites in internal/services/transformer.go (1 site)
- [ ] 3.12 Update FileError call sites in internal/services/canonicalizer.go (1 site)
- [ ] 3.13 Update FileError call sites in test files (11 sites)
- [ ] 3.14 Update formatFileError() in cmd/error_formatter.go to use getters
- [ ] 3.15 Add context and suggestions to formatFileError() output (use "Hint:" format)
- [ ] 3.16 Add unit tests for FileError getters and builders
- [ ] 3.17 Add unit tests for FileError immutability
- [ ] 3.18 Run tests and verify FileError changes

## 4. ConfigError Enhancement (26 call sites: 16 production, 10 test - BREAKING)

- [ ] 4.1 Make ConfigError fields private (field, value, message)
- [ ] 4.2 Rename ConfigError.Available field to suggestions (private)
- [ ] 4.3 Add new ConfigError field (context string)
- [ ] 4.4 Change NewConfigError() constructor signature (remove available parameter)
- [ ] 4.5 Add ConfigError getter methods (Field(), Value(), Message())
- [ ] 4.6 Add ConfigError Suggestions() getter (returns copy of suggestions)
- [ ] 4.7 Add ConfigError Context() getter
- [ ] 4.8 Add ConfigError WithSuggestions() builder (returns new instance)
- [ ] 4.9 Add ConfigError WithContext() builder (returns new instance)
- [ ] 4.10 Update ConfigError Error() method to use getters
- [ ] 4.11 Update ConfigError call sites in cmd/init.go (4 sites - constructor)
- [ ] 4.12 Update ConfigError call sites in cmd/canonicalize.go (4 sites - constructor)
- [ ] 4.13 Update ConfigError call sites in cmd/adapt.go (1 site - constructor)
- [ ] 4.14 Update ConfigError call sites in cmd/validate.go (1 site - constructor)
- [ ] 4.15 Update ConfigError call sites in internal/config/config.go (2 sites - constructor)
- [ ] 4.16 Update ConfigError call sites in internal/services/transformer.go (2 sites - constructor)
- [ ] 4.17 Update ConfigError call sites in internal/core/loader.go (2 sites - constructor)
- [ ] 4.18 Update ConfigError call sites in test files (10 sites - constructor)
- [ ] 4.19 Update formatConfigError() in cmd/error_formatter.go to use getters
- [ ] 4.20 Change formatConfigError() output from "Available:" to "Hint:" for suggestions
- [ ] 4.21 Add context to formatConfigError() output
- [ ] 4.22 Add unit tests for ConfigError getters and builders
- [ ] 4.23 Add unit tests for ConfigError immutability
- [ ] 4.24 Add unit tests for ConfigError constructor breaking change
- [ ] 4.25 Run tests and verify ConfigError changes

## 5. Final Verification and Cleanup

- [ ] 5.1 Run full test suite (mise run test)
- [ ] 5.2 Run linter (mise run lint)
- [ ] 5.3 Run format check (mise run format)
- [ ] 5.4 Run full check (mise run check)
- [ ] 5.5 Verify all error types have consistent API (getters, builders)
- [ ] 5.6 Verify all error types are immutable (builders return new instances)
- [ ] 5.7 Verify error_formatter.go uses getters consistently
- [ ] 5.8 Verify no direct field access remains (rg '\.(Path|Message|Cause|Field|Value|Available|Operation|Platform)\b' internal/ cmd/ should only show types.go)
- [ ] 5.9 Update AGENTS.md if error handling patterns changed
- [ ] 5.10 Create commit with descriptive message
