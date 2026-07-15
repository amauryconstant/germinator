package cmdutil

import (
	"errors"

	"github.com/spf13/pflag"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
)

// ExitCode is the process exit code returned by the CLI.
type ExitCode int

// Exit code constants.
const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
	ExitCodeUsage   ExitCode = 2
)

// ExitCodeFor maps an error to an ExitCode.
//
// The dispatch set is intentionally typed (errors.As) — no substring
// matching against Cobra/pflag error text. Substring matching is brittle
// to upstream wording drift; typed dispatch matches the user's intent
// without depending on Cobra's wording.
//
//	nil                                            -> 0
//	*pflag.NotExistError                           -> 2
//	*pflag.ValueRequiredError                      -> 2
//	*pflag.InvalidValueError                       -> 2
//	*pflag.InvalidSyntaxError                      -> 2
//	*core.UsageError                               -> 2
//	*core.CobraUsageError                          -> 2
//	*core.NotFoundError                            -> 1 (CORRECTED — was 2 prior to enforce-error-discipline)
//	*core.PartialSuccessError (S>0)                -> 0
//	*core.PartialSuccessError (S==0)               -> 1
//	*config.WriteError                             -> 1
//	all other errors                               -> 1
func ExitCodeFor(err error) ExitCode {
	if err == nil {
		return ExitCodeSuccess
	}
	var (
		notExist   *pflag.NotExistError
		valueReq   *pflag.ValueRequiredError
		invalidVal *pflag.InvalidValueError
		invalidSyn *pflag.InvalidSyntaxError
		usage      *core.UsageError
		cobraUsage *core.CobraUsageError
		notFound   *core.NotFoundError
		partial    *core.PartialSuccessError
		writeErr   *config.WriteError
	)
	if errors.As(err, &notExist) ||
		errors.As(err, &valueReq) ||
		errors.As(err, &invalidVal) ||
		errors.As(err, &invalidSyn) {
		return ExitCodeUsage
	}
	if errors.As(err, &usage) {
		return ExitCodeUsage
	}
	if errors.As(err, &cobraUsage) {
		return ExitCodeUsage
	}
	if errors.As(err, &notFound) {
		return ExitCodeError
	}
	if errors.As(err, &partial) {
		if partial.Succeeded() > 0 {
			return ExitCodeSuccess
		}
		return ExitCodeError
	}
	if errors.As(err, &writeErr) {
		return ExitCodeError
	}
	return ExitCodeError
}
