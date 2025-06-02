package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

var (
	// ErrNoItemFound is returned when trying to find a resource that does not exist.
	ErrNoItemFound = errors.New("no item found")
	// ErrTooManyItemsFound is returned when trying to find a resource but more than one are found.
	ErrTooManyItemsFound = errors.New("expected to find only one item")
)

// IsForbiddenError returns true if err is an HTTP 403 error.
func IsForbiddenError(err error) bool {
	var respError *scw.ResponseError
	return errors.As(err, &respError) && respError.StatusCode == http.StatusForbidden
}

// IsNotFoundError returns true if err is an HTTP 404 error or ErrNoItemFound.
func IsNotFoundError(err error) bool {
	if errors.Is(err, ErrNoItemFound) {
		return true
	}

	var notFoundError *scw.ResourceNotFoundError
	return errors.As(err, &notFoundError)
}

// IsPreconditionFailedError returns true if err is a PreconditionFailedError.
func IsPreconditionFailedError(err error) bool {
	var preconditionFailedError *scw.PreconditionFailedError
	return errors.As(err, &preconditionFailedError)
}

func newCallError(method string, err error) error {
	return fmt.Errorf("error occurred while calling %s: %w", method, err)
}
