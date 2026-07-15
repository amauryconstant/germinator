// Package canonicalize provides document-canonicalization as an I/O
// shell-package service. The package exists to satisfy the
// imperative-shell boundary defined by the
// golang-cli-architecture skill: any code that performs filesystem
// reads, document validation, and YAML output lives here at the
// package edge, not in cmd/.
//
// The Service interface, Request type, and NewService constructor are
// the canonical contract. cmd/canonicalize.go declares a local
// Canonicalizer interface that is structurally identical to
// canonicalize.Service (same parameter / return types); *canonicalizeService
// satisfies both via structural typing so cmd/ does not have to import
// this package's adapter directly.
package canonicalize

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// Request carries the inputs for document canonicalization. Lifted
// from cmd/canonicalize.go so the shell-package owns its types per
// the golang-cli-architecture "interfaces where consumed" rule (the
// consumer, cmd/canonicalize.go, declares its Canonicalizer
// interface using this type via the shared import).
type Request struct {
	InputPath  string
	OutputPath string
	Platform   string
	DocType    string
}

// Service is the per-call contract for document canonicalization.
// Returns a *core.CanonicalizeResult on success; on failure returns a
// *core.ParseError / *core.ValidationError / *core.TransformError /
// *core.FileError so cmd/cmdutil.ExitCodeFor maps them to exit 1 via
// errors.As dispatch.
type Service interface {
	Canonicalize(ctx context.Context, req *Request) (*core.CanonicalizeResult, error)
}

// canonicalizeService is the production implementation. Zero-size
// because the orchestration (parse → validate → marshal → write) is
// fully delegated to the imported shell units below.
type canonicalizeService struct{}

// Compile-time confirmation that *canonicalizeService satisfies the
// Service interface declared in this package.
var _ Service = (*canonicalizeService)(nil)

// NewService returns the production wiring for document
// canonicalization. Mirrors validate.NewService at
// internal/validate/validate.go.
func NewService() Service {
	return &canonicalizeService{}
}

// Canonicalize implements Service. Orchestrates parse-platform-doc →
// validate (flatten) → marshal-canonical → write-to-output. Each
// step wraps its failure into a typed *core.*Error so the cmd layer
// can render via output.FormatError without re-wrapping.
//
// The docType field on Request MUST be pre-validated by the caller
// via core.ValidateDocumentType. Platform is also pre-validated
// upstream by core.ValidatePlatform in cmd/canonicalize.go's
// runCanonicalize.
func (canonicalizeService) Canonicalize(ctx context.Context, req *Request) (*core.CanonicalizeResult, error) {
	doc, err := parser.ParsePlatformDocument(ctx, req.InputPath, req.Platform, req.DocType)
	if err != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return nil, core.NewValidationError("", "", "", errs[0].Error())
	}

	yamlBytes, err := renderer.MarshalCanonical(ctx, doc)
	if err != nil {
		return nil, core.NewTransformError("marshal", req.Platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(yamlBytes), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
		return nil, core.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &core.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

// validateCanonicalDoc validates a canonical document and returns
// any validation errors. Lifted verbatim from cmd/canonicalize.go.
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
// errors. The name preserves the legacy cmd-side identifier to keep
// the rename footprint minimal across the merge of stage 1.
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
