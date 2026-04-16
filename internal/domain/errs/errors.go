package errs

import (
	"errors"
	"time"
)

// Sentinel errors for task control flow.
var (
	ErrRetry           = errors.New("retry")
	ErrSkip            = errors.New("skip")
	ErrStop            = errors.New("stop")
	ErrCancel          = errors.New("cancel")
	ErrMissingResource = errors.New("missing resource")
	ErrStorageLimit    = errors.New("storage limit reached")
	ErrLackOfFreeCrop  = errors.New("lack of free crop")
	ErrQueueFull       = errors.New("building queue full")
)

// TaskError wraps a sentinel error with additional metadata.
type TaskError struct {
	Err         error
	Message     string
	NextExecute time.Time
}

func (e *TaskError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *TaskError) Unwrap() error {
	return e.Err
}

// NewSkipError creates a skip error with a scheduled next execution time.
func NewSkipError(msg string, next time.Time) *TaskError {
	return &TaskError{
		Err:         ErrSkip,
		Message:     msg,
		NextExecute: next,
	}
}

// NewStopError creates a stop error that will pause the account.
func NewStopError(msg string) *TaskError {
	return &TaskError{
		Err:     ErrStop,
		Message: msg,
	}
}

// NewRetryError creates a retry error.
func NewRetryError(msg string) *TaskError {
	return &TaskError{
		Err:     ErrRetry,
		Message: msg,
	}
}
