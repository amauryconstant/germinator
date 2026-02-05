package adapters

import (
	"strings"
	"unicode"
)

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

func ToLowerCase(s string) string {
	return strings.ToLower(s)
}

type PermissionAction string

const (
	PermissionActionAllow PermissionAction = "allow"
	PermissionActionAsk   PermissionAction = "ask"
	PermissionActionDeny  PermissionAction = "deny"
)

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

type PermissionMapping struct {
	ClaudeCode string        `json:"claudeCode"`
	OpenCode   PermissionMap `json:"openCode"`
}

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
