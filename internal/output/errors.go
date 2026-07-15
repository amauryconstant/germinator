// Package output centralizes error formatting, output format flags,
// and exporters (JSON, table) for commands that produce structured output.
package output

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// FormatError writes a formatted representation of err to io.ErrOut.
// It dispatches on typed errors via errors.As so each error type
// produces a category-specific render.
//
// The dispatch set covers all 11 core typed errors (9 existing + 2
// new: UsageError, CobraUsageError) plus the Imperative Shell typed
// error *config.WriteError. InitializeError, UsageError, and
// WriteError were added by change enforce-error-discipline (Phase 1).
func FormatError(io *iostreams.IOStreams, err error) {
	if err == nil || io == nil {
		return
	}
	var (
		parse      *core.ParseError
		valid      *core.ValidationError
		transform  *core.TransformError
		file       *core.FileError
		configErr  *core.ConfigError
		notFound   *core.NotFoundError
		partial    *core.PartialSuccessError
		operation  *core.OperationError
		initialize *core.InitializeError
		usage      *core.UsageError
		writeErr   *config.WriteError
	)
	switch {
	case errors.As(err, &parse):
		writeErrOut(io, formatParseError(io, parse))
	case errors.As(err, &valid):
		writeErrOut(io, formatValidationError(io, valid))
	case errors.As(err, &transform):
		writeErrOut(io, formatTransformError(io, transform))
	case errors.As(err, &file):
		writeErrOut(io, formatFileError(io, file))
	case errors.As(err, &configErr):
		writeErrOut(io, formatConfigError(io, configErr))
	case errors.As(err, &notFound):
		writeErrOut(io, formatNotFoundError(io, notFound))
	case errors.As(err, &partial):
		writeErrOut(io, formatPartialSuccessError(partial))
	case errors.As(err, &operation):
		writeErrOut(io, formatOperationError(io, operation))
	case errors.As(err, &initialize):
		writeErrOut(io, formatInitializeError(io, initialize))
	case errors.As(err, &usage):
		writeErrOut(io, formatUsageError(io, usage))
	case errors.As(err, &writeErr):
		writeErrOut(io, formatWriteError(io, writeErr))
	default:
		writeErrOut(io, io.Styles.Error("Error: "+err.Error())+"\n")
	}
}

func writeErrOut(io *iostreams.IOStreams, msg string) {
	_, _ = io.ErrOut.Write([]byte(msg))
}

func formatParseError(io *iostreams.IOStreams, e *core.ParseError) string {
	body := fmt.Sprintf("parse failed at %s: %s", e.Path(), e.Message())
	if e.Cause() != nil {
		body += fmt.Sprintf(": %v", e.Cause())
	}
	return io.Styles.Error("Error: ") + body + "\n"
}

func formatValidationError(io *iostreams.IOStreams, e *core.ValidationError) string {
	var sb strings.Builder
	sb.WriteString(io.Styles.Error("Error: "))
	sb.WriteString("validation failed: ")
	sb.WriteString(e.Message())
	if e.Field() != "" {
		fmt.Fprintf(&sb, " (field: %s)", e.Field())
	}
	sb.WriteString("\n")
	for _, s := range e.Suggestions() {
		fmt.Fprintf(&sb, "  Hint: %s\n", s)
	}
	return sb.String()
}

func formatTransformError(io *iostreams.IOStreams, e *core.TransformError) string {
	body := "transform failed: " + e.Message()
	if e.Platform() != "" {
		body = fmt.Sprintf("transform failed (%s for %s): %s", e.Operation(), e.Platform(), e.Message())
	}
	if e.Cause() != nil {
		body += fmt.Sprintf(": %v", e.Cause())
	}
	return io.Styles.Error("Error: ") + body + "\n"
}

func formatFileError(io *iostreams.IOStreams, e *core.FileError) string {
	body := fmt.Sprintf("%s %s: %s", e.Operation(), e.Path(), e.Message())
	if e.Cause() != nil {
		body += fmt.Sprintf(": %v", e.Cause())
	}
	return io.Styles.Error("Error: ") + body + "\n"
}

func formatConfigError(io *iostreams.IOStreams, e *core.ConfigError) string {
	body := "config: " + e.Message()
	if e.Field() != "" {
		body = fmt.Sprintf("config (%s): %s", e.Field(), e.Message())
	}
	return io.Styles.Error("Error: ") + body + "\n"
}

func formatNotFoundError(io *iostreams.IOStreams, e *core.NotFoundError) string {
	return io.Styles.Error("Error: ") + "not found: " + e.Key + "\n"
}

func formatPartialSuccessError(e *core.PartialSuccessError) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "partial success: %d succeeded, %d failed\n", e.Succeeded(), e.Failed())
	for _, ie := range e.Errors() {
		fmt.Fprintf(&sb, "  - %s\n", ie.Error())
	}
	return sb.String()
}

func formatOperationError(io *iostreams.IOStreams, e *core.OperationError) string {
	var sb strings.Builder
	sb.WriteString(io.Styles.Error("Error: "))
	sb.WriteString(e.Op)
	sb.WriteString(": ")
	sb.WriteString(e.Resource)
	sb.WriteString("\n")
	if e.Cause != nil {
		sb.WriteString("  ")
		sb.WriteString(io.Styles.Dim(e.Cause.Error()))
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatInitializeError(io *iostreams.IOStreams, e *core.InitializeError) string {
	return io.Styles.Error("Error: ") + e.Error() + "\n"
}

func formatUsageError(io *iostreams.IOStreams, e *core.UsageError) string {
	return io.Styles.Error("Error: ") + e.Flag() + ": " + e.Reason() + "\n"
}

func formatWriteError(io *iostreams.IOStreams, e *config.WriteError) string {
	return io.Styles.Error("Error: ") + e.Op() + " " + e.Path() + ": " + e.Error() + "\n"
}
