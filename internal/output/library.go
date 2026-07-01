// Package output also hosts shared library-formatter helpers,
// starting with FormatResourcesList (slice-7 7.4.7). The full
// library-formatter surface (FormatResourcesList +
// FormatBatchAddSummary) lived in cmd/library_formatters.go until
// slice 7 began migrating read-only commands off the shared helper;
// this file is the new home for the read-only side. The mutating
// helper FormatBatchAddSummary remains in cmd/library_formatters.go
// until 7.5.12 deletes the file.
package output

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gitlab.com/amoconst/germinator/internal/library"
)

// FormatResourcesList formats the library resources list as a
// human-readable string suitable for `library resources` (and other
// read-only consumers that need a stable byte-identical plain text
// rendering).
//
// The output groups resources by type in canonical order
// (skill, agent, command, memory) and renders each entry as
// "<type>/<name>" followed by a description when present. The
// "No resources found." sentinel is returned (with a trailing
// newline) when the library holds no resources so the caller can
// distinguish the empty case from a successful run that emitted no
// data.
func FormatResourcesList(lib *library.Library) string {
	var sb strings.Builder

	resources := library.ListResources(lib)

	typeOrder := []string{
		string(library.ResourceTypeSkill),
		string(library.ResourceTypeAgent),
		string(library.ResourceTypeCommand),
		string(library.ResourceTypeMemory),
	}

	hasContent := false
	for _, typ := range typeOrder {
		infos, ok := resources[typ]
		if !ok || len(infos) == 0 {
			continue
		}

		if hasContent {
			sb.WriteString("\n")
		}
		hasContent = true

		header := cases.Title(language.English).String(typ) + "s"
		sb.WriteString(header + ":\n")

		for _, info := range infos {
			ref := library.FormatRef(info.Type, info.Name)
			if info.Description != "" {
				fmt.Fprintf(&sb, "  %s - %s\n", ref, info.Description)
			} else {
				fmt.Fprintf(&sb, "  %s\n", ref)
			}
		}
	}

	if !hasContent {
		return "No resources found.\n"
	}

	return sb.String()
}
