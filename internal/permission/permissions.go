package permission

import (
	"strings"
	"unicode"

	"gitlab.com/amoconst/germinator/internal/core"
)

// ToPascalCase converts a string to PascalCase.
// It splits the input on hyphens, underscores, and whitespace,
// capitalizes the first letter of each word, and joins them together.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || unicode.IsSpace(r)
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) == 0 {
			continue
		}
		result.WriteRune(unicode.ToUpper(rune(word[0])))
		result.WriteString(strings.ToLower(word[1:]))
	}

	return result.String()
}

// ToLowerCase converts a string to lowercase.
func ToLowerCase(s string) string {
	return strings.ToLower(s)
}

// Action represents the action to take for a tool permission.
type Action string

const (
	// Allow allows the tool to be used without confirmation.
	Allow Action = "allow"
	// Ask requires user confirmation before using the tool.
	Ask Action = "ask"
	// Deny prevents the tool from being used.
	Deny Action = "deny"
)

// Map maps tool names to their permission actions.
type Map struct {
	Edit      Action `json:"edit,omitempty"`
	Bash      Action `json:"bash,omitempty"`
	Read      Action `json:"read,omitempty"`
	Grep      Action `json:"grep,omitempty"`
	Glob      Action `json:"glob,omitempty"`
	List      Action `json:"list,omitempty"`
	WebFetch  Action `json:"webfetch,omitempty"`
	WebSearch Action `json:"websearch,omitempty"`
}

// Mapping maps a Claude Code permission policy to OpenCode permissions.
type Mapping struct {
	ClaudeCode string `json:"claudeCode"`
	OpenCode   Map    `json:"openCode"`
}

// PermissionPolicyMappings maps canonical permission policy names to their platform-specific permission mappings.
// The keys match the canonical PermissionPolicy enum values (restrictive, balanced, permissive, analysis, unrestricted).
var PermissionPolicyMappings = map[string]Mapping{
	"restrictive": {
		ClaudeCode: "default",
		OpenCode: Map{
			Edit:      Ask,
			Bash:      Ask,
			Read:      Ask,
			Grep:      Ask,
			Glob:      Ask,
			List:      Ask,
			WebFetch:  Ask,
			WebSearch: Ask,
		},
	},
	"balanced": {
		ClaudeCode: "acceptEdits",
		OpenCode: Map{
			Edit:      Allow,
			Bash:      Ask,
			Read:      Allow,
			Grep:      Allow,
			Glob:      Allow,
			List:      Allow,
			WebFetch:  Allow,
			WebSearch: Allow,
		},
	},
	"permissive": {
		ClaudeCode: "dontAsk",
		OpenCode: Map{
			Edit:      Allow,
			Bash:      Allow,
			Read:      Allow,
			Grep:      Allow,
			Glob:      Allow,
			List:      Allow,
			WebFetch:  Allow,
			WebSearch: Allow,
		},
	},
	"analysis": {
		ClaudeCode: "plan",
		OpenCode: Map{
			Edit:      Deny,
			Bash:      Deny,
			Read:      Allow,
			Grep:      Allow,
			Glob:      Allow,
			List:      Allow,
			WebFetch:  Allow,
			WebSearch: Allow,
		},
	},
	"unrestricted": {
		ClaudeCode: "bypassPermissions",
		OpenCode: Map{
			Edit:      Allow,
			Bash:      Allow,
			Read:      Allow,
			Grep:      Allow,
			Glob:      Allow,
			List:      Allow,
			WebFetch:  Allow,
			WebSearch: Allow,
		},
	},
}

// ValidateActionStrings scans a nested permission object for unknown
// action strings and returns *core.ConfigError listing the valid
// permission.Action values if any are found. The expected shape is:
//
//	{ "<tool>": { "<pattern>": "<action>" }, ... }
//
// where <action> must be one of Allow, Ask, Deny. Action values that
// are not strings (e.g. bool, int) are skipped silently, matching the
// type-assertion tolerance of mapPermissionObjectToPolicy in the
// OpenCode adapter. The tool name is not validated; only the action
// strings are checked.
//
// Used by adapter code that consumes raw YAML permission maps before
// inferring a canonical PermissionPolicy.
func ValidateActionStrings(perm map[string]interface{}) error {
	for _, toolPerms := range perm {
		inner, ok := toolPerms.(map[string]interface{})
		if !ok {
			continue
		}
		for _, raw := range inner {
			actionStr, ok := raw.(string)
			if !ok {
				continue
			}
			switch Action(actionStr) {
			case Allow, Ask, Deny:
				continue
			default:
				return core.NewConfigError(
					"permission-action",
					actionStr,
					"unknown permission action: "+actionStr,
				).WithSuggestions([]string{string(Allow), string(Ask), string(Deny)})
			}
		}
	}
	return nil
}
