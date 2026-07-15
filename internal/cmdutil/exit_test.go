package cmdutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
)

func triggerPflagError(t *testing.T, args []string) error {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("platform", "", "platform")
	return fs.Parse(args)
}

func triggerPflagInvalidValue(t *testing.T) error {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Int("count", 0, "count")
	return fs.Parse([]string{"--count", "not-a-number"})
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

	pflagInvalidValue := triggerPflagInvalidValue(t)
	require.NotNil(t, pflagInvalidValue)
	var invalidValue *pflag.InvalidValueError
	require.True(t, errors.As(pflagInvalidValue, &invalidValue))

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
		{name: "pflag NotExistError", err: pflagNotExist, want: ExitCodeUsage},
		{name: "pflag ValueRequiredError", err: pflagValueRequired, want: ExitCodeUsage},
		{name: "pflag InvalidValueError", err: pflagInvalidValue, want: ExitCodeUsage},
		{name: "pflag InvalidSyntaxError", err: pflagInvalidSyntax, want: ExitCodeUsage},
		{name: "core ValidationError", err: core.NewValidationError("x", "y", "z", "msg"), want: ExitCodeError},
		{name: "core ParseError", err: core.NewParseError("/tmp/x", "bad", nil), want: ExitCodeError},
		{name: "core TransformError", err: core.NewTransformError("op", "plat", "msg", nil), want: ExitCodeError},
		{name: "core FileError", err: core.NewFileError("/tmp/x", "op", "msg", nil), want: ExitCodeError},
		{name: "core ConfigError", err: core.NewConfigError("f", "v", "msg"), want: ExitCodeError},
		{name: "core NotFoundError", err: core.NewNotFoundError("library ref", "missing"), want: ExitCodeError},
		{name: "core UsageError", err: core.NewUsageError("--resources", "must be non-empty list of refs"), want: ExitCodeUsage},
		{name: "core CobraUsageError", err: core.MustNewCobraUsageError(errors.New("requires at least 1 arg(s)")), want: ExitCodeUsage},
		{name: "config WriteError", err: config.NewWriteError("write", "/tmp/config.toml", errors.New("permission denied")), want: ExitCodeError},
		{name: "generic error", err: errors.New("boom"), want: ExitCodeError},
		{name: "PartialSuccessError S>0", err: core.NewPartialSuccessError(3, 1, nil), want: ExitCodeSuccess},
		{name: "PartialSuccessError S==0", err: core.NewPartialSuccessError(0, 1, nil), want: ExitCodeError},
		{name: "core OperationError", err: core.NewOperationError("register", "skill/commit", nil), want: ExitCodeError},
		{name: "OperationError wrapped in PartialSuccessError S>0", err: core.NewPartialSuccessError(1, 1, []core.InitializeError{
			*core.NewInitializeError("skill/commit", "/in", "/out", core.NewOperationError("register", "skill/commit", nil)),
		}), want: ExitCodeSuccess},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ExitCodeFor(tt.err))
		})
	}

	t.Run("typed dispatch traverses wrap chain", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			err  error
			want ExitCode
		}{
			{"wrapped pflag NotExistError", fmt.Errorf("cobra: %w", pflagNotExist), ExitCodeUsage},
			{"wrapped pflag ValueRequiredError", fmt.Errorf("cobra: %w", pflagValueRequired), ExitCodeUsage},
			{"wrapped pflag InvalidValueError", fmt.Errorf("cobra: %w", pflagInvalidValue), ExitCodeUsage},
			{"wrapped pflag InvalidSyntaxError", fmt.Errorf("cobra: %w", pflagInvalidSyntax), ExitCodeUsage},
			{"wrapped core NotFoundError", fmt.Errorf("resolving resource: %w", core.NewNotFoundError("library ref", "missing")), ExitCodeError},
			{"wrapped core UsageError", fmt.Errorf("validating flags: %w", core.NewUsageError("--resources", "must be non-empty list of refs")), ExitCodeUsage},
			{"wrapped core CobraUsageError", fmt.Errorf("cobra: %w", core.MustNewCobraUsageError(errors.New("requires at least 1 arg(s)"))), ExitCodeUsage},
			{"wrapped core WriteError", fmt.Errorf("saving: %w", config.NewWriteError("write", "/tmp/cfg.toml", errors.New("permission denied"))), ExitCodeError},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.want, ExitCodeFor(tt.err),
					"typed-error dispatch must traverse %%w wraps via errors.As")
			})
		}
	})

	t.Run("no substring matching on legacy Cobra phrasing", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			err  error
		}{
			{"plain unknown flag", errors.New("unknown flag: --foo")},
			{"plain flag needs argument", errors.New("flag needs an argument: --foo")},
			{"plain accepts at most N", errors.New("accepts at most 1 arg(s), only received 2")},
			{"plain requires at least", errors.New("requires at least 1 arg(s), only received 0")},
			{"plain required flag(s)", errors.New(`required flag(s) "type" not set`)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, ExitCodeError, ExitCodeFor(tt.err),
					"plain errors.New with legacy Cobra phrasing must fall through to ExitCodeError (1); the substring dispatch fallback was removed in change enforce-error-discipline")
			})
		}
	})
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
