package cmdutil

import (
	"errors"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
)

func triggerPflagError(t *testing.T, args []string) error {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("platform", "", "platform")
	return fs.Parse(args)
}

func TestExitCodeFor(t *testing.T) {
	t.Parallel()

	pflagNotExist := triggerPflagError(t, []string{"--unknown"})
	require.NotNil(t, pflagNotExist)
	var notExist *pflag.NotExistError
	require.True(t, errors.As(pflagNotExist, &notExist))

	pflagValueRequired := triggerPflagError(t, []string{"--platform"})
	require.NotNil(t, pflagValueRequired)
	var valueRequired *pflag.ValueRequiredError
	require.True(t, errors.As(pflagValueRequired, &valueRequired))

	pflagInvalidSyntax := triggerPflagError(t, []string{"---foo"})
	require.NotNil(t, pflagInvalidSyntax)
	var invalidSyntax *pflag.InvalidSyntaxError
	require.True(t, errors.As(pflagInvalidSyntax, &invalidSyntax))

	tests := []struct {
		name string
		err  error
		want ExitCode
	}{
		{name: "nil", err: nil, want: ExitCodeSuccess},
		{name: "cobra unknown flag", err: errors.New("unknown flag: --foo"), want: ExitCodeUsage},
		{name: "cobra flag needs arg", err: errors.New("flag needs an argument: --platform"), want: ExitCodeUsage},
		{name: "cobra invalid argument", err: errors.New("invalid argument \"foo\" for \"--platform\" flag"), want: ExitCodeUsage},
		{name: "cobra bad flag syntax", err: errors.New("bad flag syntax: ---foo"), want: ExitCodeUsage},
		{name: "pflag NotExistError", err: pflagNotExist, want: ExitCodeUsage},
		{name: "pflag ValueRequiredError", err: pflagValueRequired, want: ExitCodeUsage},
		{name: "pflag InvalidSyntaxError", err: pflagInvalidSyntax, want: ExitCodeUsage},
		{name: "core ValidationError", err: core.NewValidationError("x", "y", "z", "msg"), want: ExitCodeError},
		{name: "core ParseError", err: core.NewParseError("/tmp/x", "bad", nil), want: ExitCodeError},
		{name: "core TransformError", err: core.NewTransformError("op", "plat", "msg", nil), want: ExitCodeError},
		{name: "core FileError", err: core.NewFileError("/tmp/x", "op", "msg", nil), want: ExitCodeError},
		{name: "core ConfigError", err: core.NewConfigError("f", "v", "msg"), want: ExitCodeError},
		{name: "generic error", err: errors.New("boom"), want: ExitCodeError},
		{name: "PartialSuccessError S>0", err: core.NewPartialSuccessError(3, 1, nil), want: ExitCodeSuccess},
		{name: "PartialSuccessError S==0", err: core.NewPartialSuccessError(0, 1, nil), want: ExitCodeError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ExitCodeFor(tt.err))
		})
	}
}

func TestExitCodeForWrapped(t *testing.T) {
	t.Parallel()

	wrapped := &wrappedError{inner: core.NewValidationError("x", "y", "z", "msg")}
	assert.Equal(t, ExitCodeError, ExitCodeFor(wrapped))
}

type wrappedError struct {
	inner error
}

func (w *wrappedError) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrappedError) Unwrap() error { return w.inner }
