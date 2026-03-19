package adapters

import (
	"strings"
	"unicode"
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

// PermissionAction represents the action to take for a tool permission.
type PermissionAction string

const (
	// PermissionActionAllow allows the tool to be used without confirmation.
	PermissionActionAllow PermissionAction = "allow"
	// PermissionActionAsk requires user confirmation before using the tool.
	PermissionActionAsk PermissionAction = "ask"
	// PermissionActionDeny prevents the tool from being used.
	PermissionActionDeny PermissionAction = "deny"
)

// PermissionMap maps tool names to their permission actions.
type PermissionMap struct {
	Edit      PermissionAction `json:"edit,omitempty"`
	Bash      PermissionAction `json:"bash,omitempty"`
	Read      PermissionAction `json:"read,omitempty"`
	Grep      PermissionAction `json:"grep,omitempty"`
	Glob      PermissionAction `json:"glob,omitempty"`
	List      PermissionAction `json:"list,omitempty"`
	WebFetch  PermissionAction `json:"webfetch,omitempty"`
	WebSearch PermissionAction `json:"websearch,omitempty"`
}

// PermissionMapping maps a Claude Code permission policy to OpenCode permissions.
type PermissionMapping struct {
	ClaudeCode string        `json:"claudeCode"`
	OpenCode   PermissionMap `json:"openCode"`
}

// PermissionPolicyMappings maps canonical permission policy names to their platform-specific permission mappings.
// The keys match the canonical PermissionPolicy enum values (restrictive, balanced, permissive, analysis, unrestricted).
var PermissionPolicyMappings = map[string]PermissionMapping{
	"restrictive": {
		ClaudeCode: "default",
		OpenCode: PermissionMap{
			Edit:      PermissionActionAsk,
			Bash:      PermissionActionAsk,
			Read:      PermissionActionAsk,
			Grep:      PermissionActionAsk,
			Glob:      PermissionActionAsk,
			List:      PermissionActionAsk,
			WebFetch:  PermissionActionAsk,
			WebSearch: PermissionActionAsk,
		},
	},
	"balanced": {
		ClaudeCode: "acceptEdits",
		OpenCode: PermissionMap{
			Edit:      PermissionActionAllow,
			Bash:      PermissionActionAsk,
			Read:      PermissionActionAllow,
			Grep:      PermissionActionAllow,
			Glob:      PermissionActionAllow,
			List:      PermissionActionAllow,
			WebFetch:  PermissionActionAllow,
			WebSearch: PermissionActionAllow,
		},
	},
	"permissive": {
		ClaudeCode: "dontAsk",
		OpenCode: PermissionMap{
			Edit:      PermissionActionAllow,
			Bash:      PermissionActionAllow,
			Read:      PermissionActionAllow,
			Grep:      PermissionActionAllow,
			Glob:      PermissionActionAllow,
			List:      PermissionActionAllow,
			WebFetch:  PermissionActionAllow,
			WebSearch: PermissionActionAllow,
		},
	},
	"analysis": {
		ClaudeCode: "plan",
		OpenCode: PermissionMap{
			Edit:      PermissionActionDeny,
			Bash:      PermissionActionDeny,
			Read:      PermissionActionAllow,
			Grep:      PermissionActionAllow,
			Glob:      PermissionActionAllow,
			List:      PermissionActionAllow,
			WebFetch:  PermissionActionAllow,
			WebSearch: PermissionActionAllow,
		},
	},
	"unrestricted": {
		ClaudeCode: "bypassPermissions",
		OpenCode: PermissionMap{
			Edit:      PermissionActionAllow,
			Bash:      PermissionActionAllow,
			Read:      PermissionActionAllow,
			Grep:      PermissionActionAllow,
			Glob:      PermissionActionAllow,
			List:      PermissionActionAllow,
			WebFetch:  PermissionActionAllow,
			WebSearch: PermissionActionAllow,
		},
	},
}
