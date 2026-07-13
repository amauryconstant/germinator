package cmd

import (
	"context"
	"os"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
)

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
// The Factory parameter is reserved for future use; static values today.
func actionPlatforms(_ *cmdutil.Factory) carapace.Action {
	return carapace.ActionValuesDescribed(
		core.PlatformClaudeCode, "Claude Code document format",
		core.PlatformOpenCode, "OpenCode document format",
	)
}

// loadLibraryForCompletion returns the library for libPath, consulting
// the Factory's CompletionCache first. On cache miss it loads the
// library directly via library.LoadLibrary (bypassing f.Library, which
// uses sync.OnceValues and would permanently cache the first error),
// wrapping f.RootContext with a per-lookup timeout. The cache stores
// the successful result with a TTL so subsequent completions within
// the TTL are fast. Errors (including timeouts) return nil so the
// caller surfaces an empty completion rather than an error.
//
// cfg is loaded via the explicit nil-safe pattern (per task 5.1): if
// f.Config is wired (production main.go path) and returns a non-nil
// *Config, cfg is consulted for the timeout / TTL knobs. If f.Config
// is nil or returns nil/err, the helpers fall back to their defaults
// (`500ms` / `5s`). Per golang-error-handling Rule 7, a non-nil
// cfgErr is logged at debug level so failure is observable in
// --verbose runs.
func loadLibraryForCompletion(f *cmdutil.Factory, libPath string) *library.Library {
	if f.CompletionCache != nil {
		if cached := f.CompletionCache.Get(libPath); cached != nil {
			return cached
		}
	}

	var cfg *config.Config
	if f.Config != nil {
		if c, cfgErr := f.Config(); cfgErr == nil && c != nil {
			cfg = c
		} else if cfgErr != nil && f.IOStreams != nil && f.IOStreams.Logger != nil {
			f.IOStreams.Logger.Debug("config load failed; using defaults", "error", cfgErr)
		}
	}

	loadCtx, cancel := context.WithTimeout(f.RootContext, getCompletionTimeout(cfg))
	defer cancel()
	lib, err := library.LoadLibrary(loadCtx, libPath)
	if err != nil {
		return nil
	}

	if f.CompletionCache != nil {
		f.CompletionCache.Set(libPath, lib, getCacheTTL(cfg))
	}
	return lib
}

// actionResources returns a dynamic completion action for library resources.
// It loads the library (with caching and timeout support) via the Factory.
func actionResources(f *cmdutil.Factory, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		var cfg *config.Config
		if f.Config != nil {
			if c, cfgErr := f.Config(); cfgErr == nil && c != nil {
				cfg = c
			} else if cfgErr != nil && f.IOStreams != nil && f.IOStreams.Logger != nil {
				f.IOStreams.Logger.Debug("config load failed; using defaults", "error", cfgErr)
			}
		}
		libPath := resolveLibraryPath(cmd, cfg)
		if lib := loadLibraryForCompletion(f, libPath); lib != nil {
			return resourceActionFromLibrary(lib)
		}
		return carapace.ActionValues()
	})
}

// actionPresets returns a dynamic completion action for library presets.
// It loads the library (with caching and timeout support) via the Factory.
func actionPresets(f *cmdutil.Factory, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		var cfg *config.Config
		if f.Config != nil {
			if c, cfgErr := f.Config(); cfgErr == nil && c != nil {
				cfg = c
			} else if cfgErr != nil && f.IOStreams != nil && f.IOStreams.Logger != nil {
				f.IOStreams.Logger.Debug("config load failed; using defaults", "error", cfgErr)
			}
		}
		libPath := resolveLibraryPath(cmd, cfg)
		if lib := loadLibraryForCompletion(f, libPath); lib != nil {
			return presetActionFromLibrary(lib)
		}
		return carapace.ActionValues()
	})
}

// actionLibraryRefs returns a dynamic completion action combining resources and presets.
// It loads the library (with caching and timeout support) via the Factory.
func actionLibraryRefs(f *cmdutil.Factory, cmd *cobra.Command) carapace.Action {
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		var cfg *config.Config
		if f.Config != nil {
			if c, cfgErr := f.Config(); cfgErr == nil && c != nil {
				cfg = c
			} else if cfgErr != nil && f.IOStreams != nil && f.IOStreams.Logger != nil {
				f.IOStreams.Logger.Debug("config load failed; using defaults", "error", cfgErr)
			}
		}
		libPath := resolveLibraryPath(cmd, cfg)
		if lib := loadLibraryForCompletion(f, libPath); lib != nil {
			return libraryRefActionFromLibrary(lib)
		}
		return carapace.ActionValues()
	})
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
	values := make([]string, 0, 2*len(lib.Presets))

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
