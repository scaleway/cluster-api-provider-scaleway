package scaleway

import (
	"fmt"
	"time"
)

// ReconcileErrorType represents the type of a ReconcileError.
type ReconcileErrorType string

const (
	// TransientErrorType can be recovered, will be requeued after a configured time interval.
	TransientErrorType ReconcileErrorType = "Transient"
	// TerminalErrorType cannot be recovered, will not be requeued.
	TerminalErrorType ReconcileErrorType = "Terminal"
)

// ReconcileError represents an error that is not automatically recoverable
// errorType indicates what type of action is required to recover. It can take two values:
// 1. `Transient` - Can be recovered through manual intervention or by waiting, will be requeued after.
// 2. `Terminal` - Cannot be recovered, will not be requeued.
type ReconcileError struct {
	error
	errorType    ReconcileErrorType
	requestAfter time.Duration
}

// Error returns the error message for a ReconcileError.
func (t *ReconcileError) Error() string {
	var errStr string
	if t.error != nil {
		errStr = t.error.Error()
	}
	switch t.errorType {
	case TransientErrorType:
		return fmt.Sprintf("%s. Object will be requeued after %s", errStr, t.requestAfter.String())
	case TerminalErrorType:
		return fmt.Sprintf("reconcile error that cannot be recovered occurred: %s. Object will not be requeued", errStr)
	default:
		return fmt.Sprintf("reconcile error occurred with unknown recovery type. The actual error is: %s", errStr)
	}
}

// IsTransient returns true if the ReconcileError is recoverable.
func (t *ReconcileError) IsTransient() bool {
	return t.errorType == TransientErrorType
}

// IsTerminal returns true if the ReconcileError is not recoverable.
func (t *ReconcileError) IsTerminal() bool {
	return t.errorType == TerminalErrorType
}

// RequeueAfter returns requestAfter value.
func (t *ReconcileError) RequeueAfter() time.Duration {
	return t.requestAfter
}

// WithTransientError wraps the error in a ReconcileError with errorType as `Transient`.
func WithTransientError(err error, requeueAfter time.Duration) *ReconcileError {
	return &ReconcileError{error: err, errorType: TransientErrorType, requestAfter: requeueAfter}
}

// WithTerminalError wraps the error in a ReconcileError with errorType as `Terminal`.
func WithTerminalError(err error) *ReconcileError {
	return &ReconcileError{error: err, errorType: TerminalErrorType}
}
