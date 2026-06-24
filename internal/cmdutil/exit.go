package cmdutil

import (
	"errors"
	"strings"

	"github.com/spf13/pflag"

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

// cobraUsagePrefixes are the substrings used to detect Cobra-emitted
// usage errors that pflag does not wrap in a typed error.
var cobraUsagePrefixes = []string{
	"unknown flag",
	"flag needs an argument",
	"invalid argument",
	"bad flag syntax",
	"no such flag",
	"invalid syntax",
	"unknown shorthand flag",
}

// ExitCodeFor maps an error to an ExitCode.
//
//	nil                                   -> 0
//	*pflag.NotExistError                  -> 2
//	*pflag.ValueRequiredError             -> 2
//	*pflag.InvalidValueError              -> 2
//	*pflag.InvalidSyntaxError             -> 2
//	Cobra string-prefix match             -> 2
//	*core.PartialSuccessError (S>0)       -> 0
//	*core.PartialSuccessError (S==0)      -> 1
//	all other errors                      -> 1
func ExitCodeFor(err error) ExitCode {
	if err == nil {
		return ExitCodeSuccess
	}
	var (
		notExist   *pflag.NotExistError
		valueReq   *pflag.ValueRequiredError
		invalidVal *pflag.InvalidValueError
		invalidSyn *pflag.InvalidSyntaxError
		partial    *core.PartialSuccessError
	)
	if errors.As(err, &notExist) ||
		errors.As(err, &valueReq) ||
		errors.As(err, &invalidVal) ||
		errors.As(err, &invalidSyn) {
		return ExitCodeUsage
	}
	if hasCobraUsagePrefix(err) {
		return ExitCodeUsage
	}
	if errors.As(err, &partial) {
		if partial.Succeeded() > 0 {
			return ExitCodeSuccess
		}
		return ExitCodeError
	}
	return ExitCodeError
}

func hasCobraUsagePrefix(err error) bool {
	msg := err.Error()
	for _, p := range cobraUsagePrefixes {
		if strings.Contains(msg, p) {
			return true
		}
	}
	return false
}
