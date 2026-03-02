## 1. Setup

- [x] 1.1 Add Koanf dependencies to go.mod (koanf/v2, parsers/toml, providers/file)

## 2. Config Types

- [x] 2.1 Create `internal/config/` package directory
- [x] 2.2 Create `internal/config/config.go` with Config struct (Library, Platform fields)
- [x] 2.3 Add DefaultConfig() function returning defaults (Library: ~/.config/germinator/library, Platform: "")
- [x] 2.4 Add Validate() method on Config to validate platform value
- [x] 2.5 Add expandTilde helper function for path expansion

## 3. Manager Implementation

- [x] 3.1 Create `internal/config/manager.go` with ConfigManager interface (Load, GetConfig)
- [x] 3.2 Implement resolveConfigPath() for XDG location discovery
- [x] 3.3 Implement koanfConfigManager struct with Koanf-based Load()
- [x] 3.4 Implement NewConfigManager() constructor
- [x] 3.5 Implement GetConfig() returning loaded config

## 4. Tests

- [x] 4.1 Create `internal/config/config_test.go` with tests for DefaultConfig and Validate
- [x] 4.2 Create `internal/config/manager_test.go` with tests for location discovery, parsing, and loading
- [x] 4.3 Add test fixtures for valid and invalid config files

## 5. Verification

- [x] 5.1 Run `mise run check` to verify lint, format, and tests pass
