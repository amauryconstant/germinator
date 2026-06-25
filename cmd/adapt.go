package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Transformer is the local command-side contract for document
// transformation. Defined in cmd/ per the target architecture
// ("interfaces where consumed" — golang-cli-architecture skill
// principle 8). Will move to internal/core/contracts.go in change-7
// when internal/application/ is deleted.
type Transformer interface {
	Transform(ctx context.Context, req *TransformRequest) (*core.TransformResult, error)
}

// adaptOptions holds the runtime state for an `adapt` invocation.
// IO, Ctx, and Transformer come from the Factory; the rest come
// from parsed flags and positional args. Transformer is the local
// command-side interface (defined above) which matches the
// application.Transformer implementation at runtime.
type adaptOptions struct {
	IO          *iostreams.IOStreams
	Transformer func() (Transformer, error)
	Ctx         context.Context
	InputPath   string
	OutputPath  string
	Platform    string
}

// TransformRequest is the application-side request type. Imported
// from internal/application/requests.go (still alive in this change).
type TransformRequest = application.TransformRequest

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
				IO:          f.IOStreams,
				Transformer: adaptTransformer(f),
				Ctx:         c.Context(),
				InputPath:   args[0],
				OutputPath:  args[1],
				Platform:    platform,
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
		"platform": actionPlatforms(),
	})

	return cmd
}

// adaptTransformer wraps the Factory's application.Transformer lazy
// field behind the local Transformer interface. The wrapper does no
// work beyond the type assertion; it exists so the adapt command
// stays decoupled from the application package's concrete type.
func adaptTransformer(f *cmdutil.Factory) func() (Transformer, error) {
	if f == nil || f.Transformer == nil {
		return nil
	}
	return func() (Transformer, error) {
		t, err := f.Transformer()
		if err != nil {
			return nil, fmt.Errorf("resolving transformer: %w", err)
		}
		return t, nil
	}
}

// runAdapt executes the adapt logic against the resolved options.
// It is the production wiring for NewCmdAdapt's runF parameter.
func runAdapt(opts *adaptOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	transformer, err := opts.Transformer()
	if err != nil {
		return fmt.Errorf("resolving transformer: %w", err)
	}

	opts.IO.Verbosef("transforming %s → %s", opts.InputPath, opts.OutputPath)

	if _, err := transformer.Transform(opts.Ctx, &TransformRequest{
		InputPath:  opts.InputPath,
		OutputPath: opts.OutputPath,
		Platform:   opts.Platform,
	}); err != nil {
		return fmt.Errorf("transforming document: %w", err)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "wrote %s\n", opts.OutputPath)
	return nil
}
