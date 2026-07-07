package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// Canonicalizer is the local command-side contract for document
// canonicalization. Defined in cmd/ per the target architecture
// ("interfaces where consumed" — golang-cli-architecture skill
// principle 8).
type Canonicalizer interface {
	Canonicalize(ctx context.Context, req *CanonicalizeRequest) (*core.CanonicalizeResult, error)
}

// CanonicalizeRequest carries the inputs for document
// canonicalization. Local to this package since the cross-package
// type alias was removed when the legacy shell was deleted.
type CanonicalizeRequest struct {
	InputPath  string
	OutputPath string
	Platform   string
	DocType    string
}

// canonicalizeOptions holds the runtime state for a `canonicalize`
// invocation. IO and Ctx come from the Factory; the rest come from
// parsed flags and positional args. The Canonicalizer lazy field is
// the per-call injection seam for tests — production wires it to a
// closure that invokes cmd.NewCanonicalizer(); tests substitute a
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
// Canonicalizer resolution: production wires opts.Canonicalizer to a
// closure that calls cmd.NewCanonicalizer(); tests may inject a fake
// via the same field. A nil opts.Canonicalizer falls back to the
// production constructor.
func runCanonicalize(opts *canonicalizeOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	opts.IO.Verbosef("canonicalizing %s → %s (platform: %s, type: %s)",
		opts.InputPath, opts.OutputPath, opts.Platform, opts.DocType)

	resolve := opts.Canonicalizer
	if resolve == nil {
		resolve = func() (Canonicalizer, error) { return NewCanonicalizer(), nil }
	}
	c, err := resolve()
	if err != nil {
		return fmt.Errorf("resolving canonicalizer: %w", err)
	}

	if _, err := c.Canonicalize(opts.Ctx, &CanonicalizeRequest{
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

// canonicalizeDocument performs the actual canonicalization: parse
// the platform-specific document, validate it, render to canonical
// YAML, and write to the output file.
func canonicalizeDocument(_ context.Context, req *CanonicalizeRequest) (*core.CanonicalizeResult, error) {
	doc, err := parser.ParsePlatformDocument(req.InputPath, req.Platform, req.DocType)
	if err != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return nil, core.NewValidationError("", "", "", errs[0].Error())
	}

	yamlBytes, err := renderer.MarshalCanonical(doc)
	if err != nil {
		return nil, core.NewTransformError("marshal", req.Platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(yamlBytes), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
		return nil, core.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &core.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

// validateCanonicalDoc validates a canonical document and returns
// any validation errors.
func validateCanonicalDoc(doc interface{}) []error {
	switch d := doc.(type) {
	case *parser.CanonicalAgent:
		if result := core.ValidateAgent(&d.Agent); result.IsError() {
			return unwrapCanonicalErrors(result.Error)
		}
	case *parser.CanonicalCommand:
		if result := core.ValidateCommand(&d.Command); result.IsError() {
			return unwrapCanonicalErrors(result.Error)
		}
	case *parser.CanonicalSkill:
		if result := core.ValidateSkill(&d.Skill); result.IsError() {
			return unwrapCanonicalErrors(result.Error)
		}
	case *parser.CanonicalMemory:
		if result := core.ValidateMemory(&d.Memory); result.IsError() {
			return unwrapCanonicalErrors(result.Error)
		}
	default:
		return []error{core.NewParseError("", "unknown document type", nil)}
	}
	return nil
}

// unwrapCanonicalErrors unwraps a joined error into individual
// errors. Renamed to avoid symbol collision with the validate
// command which independently owns its own unwrap helper.
func unwrapCanonicalErrors(err error) []error {
	if err == nil {
		return nil
	}
	type multipleUnwrapper interface {
		Unwrap() []error
	}
	if unwrapper, ok := err.(multipleUnwrapper); ok {
		return unwrapper.Unwrap()
	}
	return []error{err}
}

// NewCanonicalizer returns the production Canonicalizer
// implementation backed by canonicalizeDocument. Used by main.go
// (and tests) to wire Factory.Canonicalizer.
func NewCanonicalizer() Canonicalizer {
	return &canonicalizerAdapter{}
}

// canonicalizerAdapter adapts the package-private canonicalizeDocument
// helper to the local Canonicalizer interface. It is a zero-size type
// because the implementation is a free function.
type canonicalizerAdapter struct{}

// Compile-time confirmation that *canonicalizerAdapter satisfies the
// local Canonicalizer interface.
var _ Canonicalizer = (*canonicalizerAdapter)(nil)

// Canonicalize delegates to canonicalizeDocument.
func (canonicalizerAdapter) Canonicalize(ctx context.Context, req *CanonicalizeRequest) (*core.CanonicalizeResult, error) {
	return canonicalizeDocument(ctx, req)
}
