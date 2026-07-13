package config

// loadFn is a package-level seam for testing. Tests override it to inject a
// stub Manager without re-running NewConfigManager(). Production code MUST
// NOT modify it. The variable is one mutable package-level binding, which
// is the documented cost of the test-injection seam (see design.md
// Decision 4 alternatives).
var loadFn = NewConfigManager

// Load is the top-level convenience wrapper that callers (notably
// `main.go`'s Factory wiring) can use without instantiating a
// Manager directly. It has the same precedence contract as
// `NewConfigManager().Load()` — defaults → file → env vars.
//
// Contract: the returned `*Config` is always non-nil. On error it
// holds the `DefaultConfig()`-seeded values (the koanf unmarshal
// happens after defaults seeding in `NewConfigManager`); the error
// chain (`*core.FileError` / `*core.ParseError` / `*core.ConfigError`,
// dispatched via `errors.As` by `output.FormatError`) is the
// authoritative signal.
func Load() (*Config, error) {
	mgr := loadFn()
	if err := mgr.Load(); err != nil {
		return mgr.GetConfig(), err
	}
	return mgr.GetConfig(), nil
}
