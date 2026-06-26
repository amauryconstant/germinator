package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/core/opencode"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/output"
	"gitlab.com/amoconst/germinator/internal/parser"
)

// Validator is the local command-side contract for document
// validation. Defined in cmd/ per the target architecture
// ("interfaces where consumed" — golang-cli-architecture principle 8).
// Will move to internal/core/contracts.go in change-7 when
// internal/application/ is deleted.
type Validator interface {
	Validate(ctx context.Context, req *ValidateRequest) (*core.ValidateResult, error)
}

// ValidateRequest is the application-side request type. Imported
// from internal/application/requests.go (still alive in this change).
type ValidateRequest = application.ValidateRequest

// validateOptions holds the runtime state for a `validate` invocation.
// IO, Ctx, and Validator come from the Factory; the rest come from
// parsed flags and positional args.
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
				Validator: validateValidator(f),
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
		"platform": actionPlatforms(),
	})

	return cmd
}

// validateValidator wraps the Factory's application.Validator lazy
// field behind the local Validator interface (matches
// cmd/adapt.go:91-102 pattern).
func validateValidator(f *cmdutil.Factory) func() (Validator, error) {
	if f == nil || f.Validator == nil {
		return nil
	}
	return func() (Validator, error) {
		v, err := f.Validator()
		if err != nil {
			return nil, fmt.Errorf("resolving validator: %w", err)
		}
		return v, nil
	}
}

// runValidate executes the validate logic against the resolved options.
// It is the production wiring for NewCmdValidate's runF parameter.
func runValidate(opts *validateOptions) error {
	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	validator, err := opts.Validator()
	if err != nil {
		return fmt.Errorf("resolving validator: %w", err)
	}

	opts.IO.Verbosef("validating %s (platform: %s)", opts.InputPath, opts.Platform)

	result, err := validator.Validate(opts.Ctx, &ValidateRequest{
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

// validateDocument contains the validation logic migrated from
// internal/service/validator.go (deleted in slice 3). Platform is
// already validated by runValidate.
func validateDocument(_ context.Context, req *ValidateRequest) (*core.ValidateResult, error) {
	docType := parser.DetectType(req.InputPath)
	if docType == "" {
		return nil, core.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := parser.ParseDocument(req.InputPath, docType)
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

// NewValidator returns the production wiring for application.Validator.
// Replaces service.NewValidator (deleted in slice 3); keeps the
// legacyBridge functional until slice 7 deletes application.Validator.
func NewValidator() application.Validator {
	return validatorAdapter{}
}

// validatorAdapter wraps validateDocument to satisfy
// application.Validator. The implementation lives in cmd/validate.go
// per slice-3 design decision 2 (no new internal/validator/ package).
type validatorAdapter struct{}

var _ application.Validator = (*validatorAdapter)(nil)

func (validatorAdapter) Validate(ctx context.Context, req *ValidateRequest) (*core.ValidateResult, error) {
	return validateDocument(ctx, req)
}
