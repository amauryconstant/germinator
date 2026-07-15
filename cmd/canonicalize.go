package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/canonicalize"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Canonicalizer is the local command-side contract for document
// canonicalization. Declared in cmd/ per the "interfaces where
// consumed" rule; the production wiring lives in
// internal/canonicalize/ and satisfies this interface via
// structural typing on *canonicalize.Request.
type Canonicalizer interface {
	Canonicalize(ctx context.Context, req *canonicalize.Request) (*core.CanonicalizeResult, error)
}

// canonicalizeOptions holds the runtime state for a `canonicalize`
// invocation. IO and Ctx come from the Factory; the rest come from
// parsed flags and positional args. The Canonicalizer lazy field is
// the per-call injection seam for tests — production wires it to a
// closure that invokes canonicalize.NewService(); tests substitute a
// fake.
type canonicalizeOptions struct {
	IO            *iostreams.IOStreams
	Canonicalizer func() (Canonicalizer, error)
	Ctx           context.Context
	InputPath     string
	OutputPath    string
	Platform      string
	DocType       string
}

// NewCmdCanonicalize creates the `canonicalize` command via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 3.
// runF is the test-injection seam; production wires it to
// runCanonicalize, tests substitute a stub.
func NewCmdCanonicalize(f *cmdutil.Factory, runF func(*canonicalizeOptions) error) *cobra.Command {
	var platform, docType string

	cmd := &cobra.Command{
		Use:   "canonicalize <input> <output>",
		Short: "Convert a platform document to canonical format",
		Long: fmt.Sprintf(`Convert a platform document to canonical YAML format.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Supported document types:
  agent   - Agent configuration
  command - Command configuration
  skill   - Skill configuration
  memory  - Memory configuration

Example:
  germinator canonicalize agent.md canonical-agent.yaml --platform %s --type agent`, core.PlatformClaudeCode, core.PlatformOpenCode, core.PlatformOpenCode),
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			opts := &canonicalizeOptions{
				IO:         f.IOStreams,
				Ctx:        c.Context(),
				InputPath:  args[0],
				OutputPath: args[1],
				Platform:   platform,
				DocType:    docType,
			}
			if runF != nil {
				return runF(opts)
			}
			return runCanonicalize(opts)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("Source platform (required: %s, %s)", core.PlatformClaudeCode, core.PlatformOpenCode))
	cmd.Flags().StringVar(&docType, "type", "", "Document type (required: agent, command, skill, memory)")
	_ = cmd.MarkFlagRequired("platform")
	_ = cmd.MarkFlagRequired("type")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform": actionPlatforms(f),
		"type":     carapace.ActionValuesDescribed("agent", "command", "skill", "memory"),
	})

	return cmd
}

// runCanonicalize executes the canonicalize logic against the
// resolved options. It is the production wiring for
// NewCmdCanonicalize's runF parameter.
//
// Pre-flight validation: ValidatePlatform guards the platform
// segment; ValidateDocumentType guards the --type flag value.
// Both produce *core.ValidationError dispatched via output.FormatError
// → ExitCodeError (1). ValidateDocumentType catches the case where
// --type is provided but has an unknown value (e.g., the plural
// "skills" or empty string); MarkFlagRequired only catches the
// missing-flag case.
//
// Canonicalizer resolution: production wires opts.Canonicalizer to a
// closure that calls canonicalize.NewService(); tests may inject a
// fake via the same field. A nil opts.Canonicalizer falls back to
// the production constructor.
func runCanonicalize(opts *canonicalizeOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}
	if err := core.ValidateDocumentType(opts.DocType); err != nil {
		return fmt.Errorf("validating document type: %w", err)
	}

	opts.IO.Verbosef("canonicalizing %s → %s (platform: %s, type: %s)",
		opts.InputPath, opts.OutputPath, opts.Platform, opts.DocType)

	resolve := opts.Canonicalizer
	if resolve == nil {
		resolve = func() (Canonicalizer, error) { return canonicalize.NewService(), nil }
	}
	c, err := resolve()
	if err != nil {
		return fmt.Errorf("resolving canonicalizer: %w", err)
	}

	if _, err := c.Canonicalize(opts.Ctx, &canonicalize.Request{
		InputPath:  opts.InputPath,
		OutputPath: opts.OutputPath,
		Platform:   opts.Platform,
		DocType:    opts.DocType,
	}); err != nil {
		return fmt.Errorf("canonicalizing document: %w", err)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "Successfully canonicalized document to: %s\n", opts.OutputPath)
	return nil
}
