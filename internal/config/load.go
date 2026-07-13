package config

import "sync"

// loadFn is the package-level seam for Load. Tests override it via
// swapLoadFn to inject a stub Manager. Production code MUST NOT modify
// it directly — use swapLoadFn which is mutex-protected. The mutable
// binding is the documented cost of the test-injection seam (see
// design.md Decision 4 alternatives).
var (
	loadFnMu sync.RWMutex
	loadFn   = NewConfigManager
)

// getLoadFn returns the currently registered loadFn under read lock.
// Used by Load() on the hot path; the RWMutex allows concurrent readers.
func getLoadFn() func() Manager {
	loadFnMu.RLock()
	defer loadFnMu.RUnlock()
	return loadFn
}

// swapLoadFn replaces the package-level loader and returns a restore
// function suitable for t.Cleanup. Use in tests only; production code
// MUST NOT call this. The restore function re-acquires the write lock
// to guarantee that concurrent readers cannot observe a torn state.
func swapLoadFn(fn func() Manager) func() {
	loadFnMu.Lock()
	defer loadFnMu.Unlock()
	prev := loadFn
	loadFn = fn
	return func() {
		loadFnMu.Lock()
		defer loadFnMu.Unlock()
		loadFn = prev
	}
}

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
	mgr := getLoadFn()()
	if err := mgr.Load(); err != nil {
		return mgr.GetConfig(), err
	}
	return mgr.GetConfig(), nil
}
