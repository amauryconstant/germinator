package cmd

import (
	"os"
	"sync"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/config"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
	"gitlab.com/amoconst/germinator/internal/models"
)

// cachedLibraryData holds cached library data with an expiration time.
type cachedLibraryData struct {
	data      *library.Library
	expiresAt time.Time
}

// completionCache is a package-level cache for library data during completions.
// It uses a mutex for thread-safe access and supports TTL-based expiration.
var completionCache struct {
	mu      sync.RWMutex
	entries map[string]*cachedLibraryData
}

// init initializes the cache map.
func init() {
	completionCache.entries = make(map[string]*cachedLibraryData)
}

// getCompletionTimeout parses the timeout from config, returning a default if invalid.
func getCompletionTimeout(cfg *config.Config) time.Duration {
	if cfg == nil || cfg.Completion.Timeout == "" {
		return 500 * time.Millisecond
	}

	d, err := time.ParseDuration(cfg.Completion.Timeout)
	if err != nil {
		return 500 * time.Millisecond
	}

	return d
}

// getCacheTTL parses the cache TTL from config, returning a default if invalid.
func getCacheTTL(cfg *config.Config) time.Duration {
	if cfg == nil || cfg.Completion.CacheTTL == "" {
		return 5 * time.Second
	}

	d, err := time.ParseDuration(cfg.Completion.CacheTTL)
	if err != nil {
		return 5 * time.Second
	}

	return d
}

// resolveLibraryPath determines the library path using the priority chain:
// flag > env > default
func resolveLibraryPath(cmd *cobra.Command, cfg *config.Config) string {
	// First, check if --library flag was provided
	if flag := cmd.Flag("library"); flag != nil && flag.Changed {
		return expandTildeInPath(flag.Value.String())
	}

	// Second, check environment variable
	if envPath := os.Getenv("GERMINATOR_LIBRARY"); envPath != "" {
		return expandTildeInPath(envPath)
	}

	// Third, check config file (if available)
	if cfg != nil && cfg.Library != "" {
		return expandTildeInPath(cfg.Library)
	}

	// Finally, use default
	return library.DefaultLibraryPath()
}

// expandTildeInPath expands ~ in a path to the user's home directory.
func expandTildeInPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return homeDir + path[1:]
	}
	return path
}

// actionPlatforms returns a static completion action for platform values.
func actionPlatforms() carapace.Action {
	return carapace.ActionValuesDescribed(
		models.PlatformClaudeCode, "Claude Code document format",
		models.PlatformOpenCode, "OpenCode document format",
	)
}

// actionResources returns a dynamic completion action for library resources.
// It loads the library with caching and timeout support.
func actionResources(cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		libPath := resolveLibraryPath(cmd, nil)
		timeout := getCompletionTimeout(nil)
		cacheTTL := getCacheTTL(nil)

		// Check cache first
		cached := getCachedLibrary(libPath, cacheTTL)
		if cached != nil {
			return resourceActionFromLibrary(cached)
		}

		// Load library with timeout
		libCh := make(chan *library.Library, 1)
		errCh := make(chan error, 1)

		go func() {
			lib, err := library.LoadLibrary(libPath)
			if err != nil {
				errCh <- err
				return
			}
			libCh <- lib
		}()

		select {
		case lib := <-libCh:
			// Cache the result
			setCachedLibrary(libPath, lib, cacheTTL)
			return resourceActionFromLibrary(lib)
		case <-errCh:
			// Silent failure - return empty completions
			return carapace.ActionValues()
		case <-time.After(timeout):
			// Timeout - return empty completions
			return carapace.ActionValues()
		}
	})
}

// actionPresets returns a dynamic completion action for library presets.
// It loads the library with caching and timeout support.
func actionPresets(cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		libPath := resolveLibraryPath(cmd, nil)
		timeout := getCompletionTimeout(nil)
		cacheTTL := getCacheTTL(nil)

		// Check cache first
		cached := getCachedLibrary(libPath, cacheTTL)
		if cached != nil {
			return presetActionFromLibrary(cached)
		}

		// Load library with timeout
		libCh := make(chan *library.Library, 1)
		errCh := make(chan error, 1)

		go func() {
			lib, err := library.LoadLibrary(libPath)
			if err != nil {
				errCh <- err
				return
			}
			libCh <- lib
		}()

		select {
		case lib := <-libCh:
			// Cache the result
			setCachedLibrary(libPath, lib, cacheTTL)
			return presetActionFromLibrary(lib)
		case <-errCh:
			// Silent failure - return empty completions
			return carapace.ActionValues()
		case <-time.After(timeout):
			// Timeout - return empty completions
			return carapace.ActionValues()
		}
	})
}

// actionLibraryRefs returns a dynamic completion action combining resources and presets.
// It loads the library with caching and timeout support.
func actionLibraryRefs(cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		libPath := resolveLibraryPath(cmd, nil)
		timeout := getCompletionTimeout(nil)
		cacheTTL := getCacheTTL(nil)

		// Check cache first
		cached := getCachedLibrary(libPath, cacheTTL)
		if cached != nil {
			return libraryRefActionFromLibrary(cached)
		}

		// Load library with timeout
		libCh := make(chan *library.Library, 1)
		errCh := make(chan error, 1)

		go func() {
			lib, err := library.LoadLibrary(libPath)
			if err != nil {
				errCh <- err
				return
			}
			libCh <- lib
		}()

		select {
		case lib := <-libCh:
			// Cache the result
			setCachedLibrary(libPath, lib, cacheTTL)
			return libraryRefActionFromLibrary(lib)
		case <-errCh:
			// Silent failure - return empty completions
			return carapace.ActionValues()
		case <-time.After(timeout):
			// Timeout - return empty completions
			return carapace.ActionValues()
		}
	})
}

// getCachedLibrary retrieves cached library data if it exists and hasn't expired.
func getCachedLibrary(libPath string, _ time.Duration) *library.Library {
	completionCache.mu.RLock()
	defer completionCache.mu.RUnlock()

	entry, exists := completionCache.entries[libPath]
	if !exists {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.data
}

// setCachedLibrary stores library data in the cache with the given TTL.
func setCachedLibrary(libPath string, lib *library.Library, ttl time.Duration) {
	completionCache.mu.Lock()
	defer completionCache.mu.Unlock()

	completionCache.entries[libPath] = &cachedLibraryData{
		data:      lib,
		expiresAt: time.Now().Add(ttl),
	}
}

// resourceActionFromLibrary creates a completion action from a library's resources.
func resourceActionFromLibrary(lib *library.Library) carapace.Action {
	var values []string

	for typ, names := range lib.Resources {
		for name, res := range names {
			ref := library.FormatRef(typ, name)
			values = append(values, ref, res.Description)
		}
	}

	if len(values) == 0 {
		return carapace.ActionValues()
	}

	return carapace.ActionValuesDescribed(values...)
}

// presetActionFromLibrary creates a completion action from a library's presets.
func presetActionFromLibrary(lib *library.Library) carapace.Action {
	var values []string

	for name, preset := range lib.Presets {
		values = append(values, name, preset.Description)
	}

	if len(values) == 0 {
		return carapace.ActionValues()
	}

	return carapace.ActionValuesDescribed(values...)
}

// libraryRefActionFromLibrary creates a completion action combining resources and preset references.
func libraryRefActionFromLibrary(lib *library.Library) carapace.Action {
	var values []string

	// Add resources
	for typ, names := range lib.Resources {
		for name, res := range names {
			ref := library.FormatRef(typ, name)
			values = append(values, ref, res.Description)
		}
	}

	// Add presets with "preset/" prefix
	for name, preset := range lib.Presets {
		ref := "preset/" + name
		values = append(values, ref, preset.Description)
	}

	if len(values) == 0 {
		return carapace.ActionValues()
	}

	return carapace.ActionValuesDescribed(values...)
}
