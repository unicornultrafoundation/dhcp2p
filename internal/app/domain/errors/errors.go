package errors

import "errors"

var (
	ErrMissingDependencies = errors.New("missing dependencies")
)
