package main

import (
	"errors"
	"fmt"
	"os"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

type ExitCode int

const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
	ExitCodeUsage   ExitCode = 2
	ExitCodeParse   ExitCode = 3
)

type ErrorCategory int

const (
	CategoryCobra ErrorCategory = iota
	CategoryConfig
	CategoryParse
	CategoryValidation
	CategoryTransform
	CategoryFile
	CategoryGeneric
)

func CategorizeError(err error) ErrorCategory {
	var parseErr *gerrors.ParseError
	var validationErr *gerrors.ValidationError
	var transformErr *gerrors.TransformError
	var fileErr *gerrors.FileError
	var configErr *gerrors.ConfigError

	if errors.As(err, &parseErr) {
		return CategoryParse
	}
	if errors.As(err, &validationErr) {
		return CategoryValidation
	}
	if errors.As(err, &transformErr) {
		return CategoryTransform
	}
	if errors.As(err, &fileErr) {
		return CategoryFile
	}
	if errors.As(err, &configErr) {
		return CategoryConfig
	}

	return CategoryGeneric
}

func GetExitCodeForError(err error) ExitCode {
	category := CategorizeError(err)

	switch category {
	case CategoryParse:
		return ExitCodeParse
	case CategoryConfig, CategoryValidation, CategoryCobra:
		return ExitCodeUsage
	case CategoryTransform, CategoryFile, CategoryGeneric:
		return ExitCodeError
	default:
		return ExitCodeError
	}
}

func HandleError(cfg *CommandConfig, err error) {
	fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(err))
	os.Exit(int(GetExitCodeForError(err)))
}
