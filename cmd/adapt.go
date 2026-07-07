package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Transformer is the local command-side contract for document
// transformation. Defined in cmd/ per the target architecture
// (interfaces are declared where consumed; see the
// golang-cli-architecture skill).
type Transformer interface {
	Transform(ctx context.Context, req *TransformRequest) (*core.TransformResult, error)
}

// TransformRequest carries the inputs for a document transformation.
// Local to this package since the cross-package type alias was
// removed when the legacy shell was deleted.
type TransformRequest struct {
	InputPath  string
	OutputPath string
	Platform   string
}

// adaptOptions holds the runtime state for an `adapt` invocation. IO
// and Ctx come from the Factory; the rest come from parsed flags and
// positional args. The Transformer lazy field is the per-call
// injection seam for tests — production wires it to a closure that
// invokes cmd.NewTransformer(); tests substitute a fake.
type adaptOptions struct {
	IO          *iostreams.IOStreams
	Transformer func() (Transformer, error)
	Ctx         context.Context
	InputPath   string
	OutputPath  string
	Platform    string
}

// NewCmdAdapt creates the `adapt` command via the canonical
// NewCmdXxx(f, runF) pattern. runF is the test-injection seam;
// production wires it to runAdapt, tests substitute a stub.
func NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "adapt <input> <output>",
		Short: "Transform a document to another platform",
		Long: `Transform a document from Germinator source format to another platform's format.

Supported platforms:
  claude-code  - Claude Code document format
  opencode     - OpenCode document format

Example:
  germinator adapt agent.yaml opencode-agent.md --platform opencode`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			opts := &adaptOptions{
				IO:         f.IOStreams,
				Ctx:        c.Context(),
				InputPath:  args[0],
				OutputPath: args[1],
				Platform:   platform,
			}
			if runF != nil {
				return runF(opts)
			}
			return runAdapt(opts)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Target platform (required: claude-code, opencode)")
	_ = cmd.MarkFlagRequired("platform")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform": actionPlatforms(f),
	})

	return cmd
}

// runAdapt executes the adapt logic against the resolved options.
// It is the production wiring for NewCmdAdapt's runF parameter.
//
// Transformer resolution: production wires opts.Transformer to a
// closure that calls cmd.NewTransformer(); tests may inject a fake
// via the same field. A nil opts.Transformer falls back to the
// production constructor so callers that don't populate the field
// still get correct behavior.
func runAdapt(opts *adaptOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	opts.IO.Verbosef("transforming %s → %s", opts.InputPath, opts.OutputPath)

	resolve := opts.Transformer
	if resolve == nil {
		resolve = func() (Transformer, error) { return NewTransformer(), nil }
	}
	t, err := resolve()
	if err != nil {
		return fmt.Errorf("resolving transformer: %w", err)
	}

	if _, err := t.Transform(opts.Ctx, &TransformRequest{
		InputPath:  opts.InputPath,
		OutputPath: opts.OutputPath,
		Platform:   opts.Platform,
	}); err != nil {
		return fmt.Errorf("transforming document: %w", err)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "wrote %s\n", opts.OutputPath)
	return nil
}
