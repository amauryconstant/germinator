package models

import (
	"fmt"
)

// ValidatePlatform checks if platform parameter is valid.
func ValidatePlatform(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, fmt.Errorf("platform is required (available: %s, %s)", PlatformClaudeCode, PlatformOpenCode))
		return errs
	}

	if platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: %s, %s)", platform, PlatformClaudeCode, PlatformOpenCode))
		return errs
	}

	return nil
}
