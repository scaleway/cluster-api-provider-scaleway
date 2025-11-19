package scaleway

import (
	"errors"
	"fmt"
	"time"
)

// ReconcileError represents an error that may be fixed by waiting.
type ReconcileError struct {
	error
	requestAfter time.Duration
}

// Error returns the error message for a ReconcileError.
func (t *ReconcileError) Error() string {
	var errStr string
	if t.error != nil {
		errStr = t.error.Error()
	}

	if t.requestAfter != 0 {
		return fmt.Sprintf("%s. Object will be requeued after %s", errStr, t.requestAfter.String())
	}

	return fmt.Sprintf("reconcile error: %s", errStr)
}

// RequeueAfter returns requestAfter value.
func (t *ReconcileError) RequeueAfter() time.Duration {
	if t == nil {
		return 0
	}

	return t.requestAfter
}

// WithTransientError wraps the error in a ReconcileError with requeueAfter duration.
func WithTransientError(err error, requeueAfter time.Duration) *ReconcileError {
	return &ReconcileError{error: err, requestAfter: requeueAfter}
}

// IsTransientReconcileError returns true if the provided error is a transient ReconcileError.
func IsTransientReconcileError(err error) bool {
	var reconcileErr *ReconcileError
	return errors.As(err, &reconcileErr) && reconcileErr.RequeueAfter() != 0
}
