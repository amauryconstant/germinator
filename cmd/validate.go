package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/core/opencode"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/output"
	"gitlab.com/amoconst/germinator/internal/parser"
)

// Validator is the local command-side contract for document
// validation. Defined in cmd/ per the target architecture
// (interfaces are declared where consumed; see the
// golang-cli-architecture skill).
type Validator interface {
	Validate(ctx context.Context, req *ValidateRequest) (*core.ValidateResult, error)
}

// ValidateRequest carries the inputs for document validation. Local
// to this package since the cross-package type alias was removed
// when the legacy shell was deleted.
type ValidateRequest struct {
	InputPath string
	Platform  string
}

// validateOptions holds the runtime state for a `validate`
// invocation. IO and Ctx come from the Factory; the rest come from
// parsed flags and positional args. The Validator lazy field is the
// per-call injection seam for tests — production wires it to a
// closure that invokes cmd.NewValidator(); tests substitute a fake.
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
// that calls cmd.NewValidator(); tests may inject a fake via the same
// field. A nil opts.Validator falls back to the production
// constructor.
func runValidate(opts *validateOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	opts.IO.Verbosef("validating %s (platform: %s)", opts.InputPath, opts.Platform)

	resolve := opts.Validator
	if resolve == nil {
		resolve = func() (Validator, error) { return NewValidator(), nil }
	}
	v, err := resolve()
	if err != nil {
		return fmt.Errorf("resolving validator: %w", err)
	}

	result, err := v.Validate(opts.Ctx, &ValidateRequest{
		InputPath: opts.InputPath,
		Platform:  opts.Platform,
	})
	if err != nil {
		return fmt.Errorf("validating document: %w", err)
	}

	if !result.Valid() {
		for _, e := range result.Errors {
			output.FormatError(opts.IO, e)
		}
		return result.Errors[0]
	}

	_, _ = fmt.Fprintln(opts.IO.Out, "Document is valid")
	return nil
}

// validateDocument contains the document validation logic. Platform is
// already validated by runValidate.
func validateDocument(ctx context.Context, req *ValidateRequest) (*core.ValidateResult, error) {
	docType := parser.DetectType(ctx, req.InputPath)
	if docType == "" {
		return nil, core.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := parser.ParseDocument(ctx, req.InputPath, docType)
	if parseErr != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *parser.CanonicalAgent:
		if result := core.ValidateAgent(&d.Agent); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateAgentOpenCode(&d.Agent); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *parser.CanonicalCommand:
		if result := core.ValidateCommand(&d.Command); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateCommandOpenCode(&d.Command); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *parser.CanonicalMemory:
		if result := core.ValidateMemory(&d.Memory); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
	case *parser.CanonicalSkill:
		if result := core.ValidateSkill(&d.Skill); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == core.PlatformOpenCode {
			if result := opencode.ValidateSkillOpenCode(&d.Skill); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	default:
		return nil, core.NewParseError(req.InputPath, "unknown document type", nil)
	}

	return &core.ValidateResult{Errors: errs}, nil
}

// unwrapErrors unwraps a joined error into individual errors.
// If the error is not a joined error, returns a slice with just that error.
func unwrapErrors(err error) []error {
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

// NewValidator returns the production wiring for the cmd-side
// Validator interface. Mirrors NewCanonicalizer at cmd/canonicalize.go.
func NewValidator() Validator {
	return &validatorAdapter{}
}

// validatorAdapter adapts the package-private validateDocument
// helper to the local Validator interface. The implementation lives
// in cmd/validate.go per slice-3 design decision 2 (no new
// internal/validator/ package).
type validatorAdapter struct{}

// Compile-time confirmation that *validatorAdapter satisfies the
// local Validator interface.
var _ Validator = (*validatorAdapter)(nil)

// Validate delegates to validateDocument.
func (validatorAdapter) Validate(ctx context.Context, req *ValidateRequest) (*core.ValidateResult, error) {
	return validateDocument(ctx, req)
}
