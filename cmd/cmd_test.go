package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/canonicalize"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
	"gitlab.com/amoconst/germinator/internal/transform"
	"gitlab.com/amoconst/germinator/internal/validate"
)

// TestAdaptCommand exercises the production Transformer adapter via
// transform.NewService(parser.NewParser(), renderer.NewSerializer()).Transform
// end-to-end. Replaces the legacy TestAdaptCommand that routed through
// bridge.Services.Transformer.
func TestAdaptCommand(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/input-agent.md"
	outputFile := tmpDir + "/output-agent.md"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
---
This is test content`

	require.NoError(t, os.WriteFile(inputFile, []byte(content), 0o644))

	tests := []struct {
		name        string
		platform    string
		expectError bool
		errorMsg    string
	}{
		{name: "invalid platform", platform: "invalid-platform", expectError: true, errorMsg: "unknown platform"},
		{name: "valid claude-code platform", platform: "claude-code", expectError: false},
		{name: "valid opencode platform", platform: "opencode", expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := transform.NewService(parser.NewParser(), renderer.NewSerializer()).Transform(context.Background(), &transform.Request{
				InputPath:  inputFile,
				OutputPath: outputFile,
				Platform:   tt.platform,
			})
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestValidateCommandWithPlatformVariants exercises the production
// Validator adapter via validate.NewService().Validate for both
// supported platforms. Replaces the legacy
// TestValidateCommandWithPlatformVariants that routed through
// bridge.Services.Validator.
func TestValidateCommandWithPlatformVariants(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := tmpDir + "/test-command.md"

	content := `---
name: test-command
description: A test command
template: command template
---
Command content`

	require.NoError(t, os.WriteFile(validFile, []byte(content), 0o644))

	tests := []struct {
		name     string
		platform string
	}{
		{"claude-code platform", "claude-code"},
		{"opencode platform", "opencode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validate.NewService().Validate(context.Background(), &validate.Request{
				InputPath: validFile,
				Platform:  tt.platform,
			})
			require.NoError(t, err)
			assert.True(t, result.Valid(),
				"expected no validation errors for %s, got %d: %v",
				tt.platform, len(result.Errors), result.Errors)
		})
	}
}

// TestRootCommand_RunHelp verifies the root command renders its Long
// description (help text) when invoked without subcommands. The
// legacy os.Pipe hack was replaced with iostreams.Test + cmd.SetOut,
// matching the slice-2+ Factory.IOStreams pattern.
func TestRootCommand_RunHelp(t *testing.T) {
	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios)
	rootCmd := NewRootCommand(f)
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	rootCmd.Run(rootCmd, []string{})

	assert.Contains(t, out.String(), "Germinator is a configuration adapter",
		"Root command should show help")
}

// TestCanonicalizeCommandWithAllFlags drives the production
// canonicalize path end-to-end and asserts the YAML payload contains
// the parsed name and description fields. Replaces the legacy
// TestCanonicalizeCommandWithAllFlags (which routed through
// bridge.Services.Canonicalizer).
func TestCanonicalizeCommandWithAllFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	outputFile := tmpDir + "/canonical-agent.yaml"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
---
Agent content`

	require.NoError(t, os.WriteFile(inputFile, []byte(content), 0o644))

	_, err := canonicalize.NewService().Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	require.NoError(t, err)

	result, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	resultStr := string(result)

	assert.Contains(t, resultStr, "name: test-agent")
	assert.Contains(t, resultStr, "description: A test agent")
}

// The four Canonicalize error-path tests below mirror the legacy
// TestCanonicalizeCommand{Missing,Invalid}{Platform,Type} shapes.
// Each asserts the right core error class is surfaced for the
// corresponding invalid input. The production Canonicalize service
// returns typed core errors via errors.As; tests invoke the public
// canonicalize.NewService().Canonicalize entry point directly.

func TestCanonicalizeCommandMissingPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	require.NoError(t, os.WriteFile(inputFile, []byte(`---
name: test-agent
description: A test agent
---
Body`), 0o644))

	_, err := canonicalize.NewService().Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/out.yaml",
		Platform:   "",
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported platform")
}

func TestCanonicalizeCommandInvalidPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	require.NoError(t, os.WriteFile(inputFile, []byte(`---
name: test-agent
description: A test agent
---
Body`), 0o644))

	_, err := canonicalize.NewService().Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/out.yaml",
		Platform:   "invalid-platform",
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported platform")
}

func TestCanonicalizeCommandInvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	require.NoError(t, os.WriteFile(inputFile, []byte(`---
name: test-agent
description: A test agent
---
Body`), 0o644))

	_, err := canonicalize.NewService().Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/out.yaml",
		Platform:   "claude-code",
		DocType:    "invalid-type",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")
}

func TestCanonicalizeCommandFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := canonicalize.NewService().Canonicalize(context.Background(), &canonicalize.Request{
		InputPath:  tmpDir + "/non-existent-file.md",
		OutputPath: tmpDir + "/out.yaml",
		Platform:   "claude-code",
		DocType:    "agent",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// TestExitCodeForErrorTypes — smoke test that all legacy typed core
// errors map to ExitCodeError (1) under the slice-1 0/1/2 mapping.
// The legacy 7-code scheme collapsed in slice 1; per-error-code
// coverage lives in TestExitCodeForTypedErrors.
func TestExitCodeForErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "ParseError", err: core.NewParseError("test.yaml", "parse failed", nil)},
		{name: "ValidationError", err: core.NewValidationError("", "name", "", "invalid field")},
		{name: "ConfigError", err: core.NewConfigError("platform", "invalid", "unknown platform").WithSuggestions([]string{"claude-code"})},
		{name: "TransformError", err: core.NewTransformError("render", "opencode", "failed", nil)},
		{name: "FileError", err: core.NewFileError("test.yaml", "read", "not found", nil)},
		{name: "generic error", err: fmt.Errorf("something went wrong")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := cmdutil.ExitCodeFor(tt.err)
			assert.Equal(t, cmdutil.ExitCodeError, code,
				"cmdutil.ExitCodeFor() = %d, want ExitCodeError (1)", code)
		})
	}
}

// TestExitCodeForTypedErrors verifies the slice-1 0/1/2 exit-code
// mapping for typed core errors. cmdutil.ExitCodeFor collapses typed
// errors to ExitCodeError (1) except PartialSuccess which keeps the
// Succeeded>0 → ExitCodeSuccess (0) shortcut.
func TestExitCodeForTypedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want cmdutil.ExitCode
	}{
		{name: "ParseError returns ExitCodeError", err: core.NewParseError("test.yaml", "bad", nil), want: 1},
		{name: "ValidationError returns ExitCodeError", err: core.NewValidationError("", "field", "", "invalid"), want: 1},
		{name: "TransformError returns ExitCodeError", err: core.NewTransformError("render", "opencode", "failed", nil), want: 1},
		{name: "FileError returns ExitCodeError", err: core.NewFileError("test.yaml", "read", "failed", nil), want: 1},
		{name: "ConfigError returns ExitCodeError", err: core.NewConfigError("f", "v", "msg"), want: 1},
		{name: "PartialSuccess S>0 returns ExitCodeSuccess", err: core.NewPartialSuccessError(3, 1, nil), want: 0},
		{name: "PartialSuccess S==0 returns ExitCodeError", err: core.NewPartialSuccessError(0, 1, nil), want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cmdutil.ExitCodeFor(tt.err))
		})
	}
}

// TestHandleErrorWithNilError pins cmdutil.ExitCodeFor(nil) == 0.
// (The legacy GetExitCodeForError(nil) returned 1, which was a bug —
// a nil error means "nothing went wrong".)
func TestHandleErrorWithNilError(t *testing.T) {
	assert.Equal(t, cmdutil.ExitCodeSuccess, cmdutil.ExitCodeFor(nil),
		"cmdutil.ExitCodeFor(nil) must map to ExitCodeSuccess (0)")
}

// Legacy tests deleted in slice 7.5 (cmd_test.go rewrite):
//
// - TestValidateCommandWithActualServices — exercised the
//   bridge.Services.Validator plumbing; equivalent behavior is now
//   covered by cmd/validate_test.go's TestValidateDocument_HappyPath
//   and TestValidateService_AdapterSatisfiesInterface.
// - TestValidateCommandValidDocument — duplicate of the new
//   TestValidateCommandWithPlatformVariants shape.
// - TestCLIPlatformFlagValidation — platform validation moved to
//   runValidate via core.ValidatePlatform in slice 3; covered by
//   cmd/validate_test.go's TestRunValidate_InvalidPlatform_ReturnsValidationError.
// - TestCLIDescriptiveErrorMessages — superseded by table-driven
//   coverage in cmd/validate_test.go and the document-shape tests.
// - TestCLIAdaptEndToEnd, TestCLIValidateEndToEnd — end-to-end
//   integration tests that drove the production transformer/validator
//   directly through bridge.Services; replaced by the
//   transform.NewService() / validate.NewService() call sites in
//   TestAdaptCommand / TestValidateCommandWithPlatformVariants.
// - TestCanonicalizeCommandMissingType — duplicate of
//   TestCanonicalizeCommandInvalidType (both assert "unknown
//   document type" via different invalid inputs).
// - TestCanonicalizeCommandSuccessfulConversion — duplicate of
//   TestCanonicalizeCommandWithAllFlags.
// - TestValidateCommandVerboseFlag, TestAdaptCommandVerboseFlag,
//   TestCanonicalizeCommandVerboseFlag — verbose-output tests
//   replaced by the per-command TestRun*_VerboseProgressToStderr
//   tests in cmd/{validate,adapt,canonicalize}_test.go.
// - TestValidateCommandExitCodes, TestAdaptCommandExitCodes —
//   exit-code tests replaced by cmd/{validate,adapt}_test.go's
//   direct assertion of cmdutil.ExitCodeFor on typed core errors
//   (TestRunValidate_*, TestRunAdapt_*, TestExitCodeForTypedErrors).
// - Several slice-1 tests (covering the deleted legacy shell's
//   configuration container and the typed error formatter that
//   preceded output.FormatError) were removed with the legacy shell
//   in slice 7.5. Equivalent coverage lives in
//   TestExitCodeForTypedErrors and the per-command TestRun* tests.
