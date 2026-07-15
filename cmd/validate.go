package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/validate"
)

// Validator is the local command-side contract for document
// validation. Declared in cmd/ per the "interfaces where consumed"
// rule; the production wiring lives in internal/validate/ and
// satisfies this interface via structural typing on *validate.Request.
type Validator interface {
	Validate(ctx context.Context, req *validate.Request) (*core.ValidateResult, error)
}

// validateOptions holds the runtime state for a `validate`
// invocation. IO and Ctx come from the Factory; the rest come from
// parsed flags and positional args. The Validator lazy field is the
// per-call injection seam for tests — production wires it to a
// closure that invokes validate.NewService(); tests substitute a
// fake.
type validateOptions struct {
	IO        *iostreams.IOStreams
	Validator func() (Validator, error)
	Ctx       context.Context
	InputPath string
	Platform  string
}

// NewCmdValidate creates the `validate` command via the canonical
// NewCmdXxx(f, runF) pattern. runF is the test-injection seam;
// production wires it to runValidate, tests substitute a stub.
func NewCmdValidate(f *cmdutil.Factory, runF func(*validateOptions) error) *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a document file",
		Long: `Validate a document file and display any errors found.

Supported platforms:
  claude-code  - Claude Code document format
  opencode     - OpenCode document format

Example:
  germinator validate agent.yaml --platform claude-code`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := &validateOptions{
				IO:        f.IOStreams,
				Ctx:       c.Context(),
				InputPath: args[0],
				Platform:  platform,
			}
			if runF != nil {
				return runF(opts)
			}
			return runValidate(opts)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Target platform (required: claude-code, opencode)")
	_ = cmd.MarkFlagRequired("platform")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform": actionPlatforms(f),
	})

	return cmd
}

// runValidate executes the validate logic against the resolved
// options. It is the production wiring for NewCmdValidate's runF
// parameter.
//
// Validator resolution: production wires opts.Validator to a closure
// that calls validate.NewService(); tests may inject a fake via the
// same field. A nil opts.Validator falls back to the production
// constructor.
func runValidate(opts *validateOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	opts.IO.Verbosef("validating %s (platform: %s)", opts.InputPath, opts.Platform)

	resolve := opts.Validator
	if resolve == nil {
		resolve = func() (Validator, error) { return validate.NewService(), nil }
	}
	v, err := resolve()
	if err != nil {
		return fmt.Errorf("resolving validator: %w", err)
	}

	result, err := v.Validate(opts.Ctx, &validate.Request{
		InputPath: opts.InputPath,
		Platform:  opts.Platform,
	})
	if err != nil {
		return fmt.Errorf("validating document: %w", err)
	}

	if !result.Valid() {
		return result.Errors[0]
	}

	_, _ = fmt.Fprintln(opts.IO.Out, "Document is valid")
	return nil
}
