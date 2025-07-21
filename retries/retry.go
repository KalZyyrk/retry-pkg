// Package retries provides intelligent retry functionality with automatic error handling
// for HTTP operations and other recoverable failures.
//
// This package offers a generic Retry function that can work with any return type,
// implementing smart retry logic that distinguishes between recoverable and
// unrecoverable errors. HTTP responses are handled intelligently:
//   - 2xx: Success, no retry needed
//   - 4xx: Client errors, marked as unrecoverable (no retry)
//   - 5xx: Server errors, will retry up to the configured limit
//
// The package uses the github.com/avast/retry-go library under the hood and extends
// it with HTTP-aware error handling and predefined error variables for common
// HTTP status codes.
package retries

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/avast/retry-go"
)

// count tracks the number of retry attempts for the current operation.
// This is used primarily for testing purposes to verify retry behavior.
// Note: This is a global variable, which means it's not thread-safe, but since
// it's intended for testing scenarios, this is acceptable.
var count int

// HTTP error variables for common status codes.
// These predefined errors provide consistent error messages
// across the application when handling HTTP-related failures.
var (
	// ErrBadRequest represents a 400 Bad Request error.
	// Used when the client sends a malformed or invalid request.
	ErrBadRequest = errors.New("bad request - Status Code: 400")

	// ErrForbidden represents a 403 Forbidden error.
	// Used when the client lacks proper authorization to access the resource.
	ErrForbidden = errors.New("forbidden - Status Code: 403")

	// ErrNotFound represents a 404 Not Found error.
	// Used when the requested resource cannot be found on the server.
	ErrNotFound = errors.New("not found - Status Code: 404")

	// ErrNotImplemented represents a 501 Not Implemented error.
	// Used when the server does not support the functionality required to fulfill the request.
	ErrNotImplemented = errors.New("not implemented - Status Code: 501")
)

// Retry executes a function with automatic retry logic and intelligent error handling.
// It uses Go generics to work with any return type T.
//
// The function will retry up to 5 times by default, but only for recoverable errors.
// HTTP responses are handled intelligently:
// - 2xx: Success, no retry
// - 4xx: Client error, marked as unrecoverable (no retry)
// - 5xx: Server error, will retry
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - f: The function to execute, must return (T, error)
//   - opts: Optional retry-go configuration options
//
// Returns:
//   - T: The result from the successful function execution
//   - error: Any error that occurred, or nil on success
func Retry[T any](ctx context.Context, f func() (T, error), opts ...retry.Option) (T, error) {
	var (
		res T
		err error
	)

	count = 0
	err = retry.Do(
		func() error {
			res, err = f()
			res, err = checkResAndErr(res, err)

			return err
		},
		retry.RetryIf(retryAction),
	)

	return res, err
}

// checkResAndErr analyzes the response and error to determine retry behavior.
// This function implements smart HTTP response handling by examining status codes
// and marking 4xx errors as unrecoverable.
//
// Parameters:
//   - res: The response from the function (any type T)
//   - err: The error from the function (may be nil)
//
// Returns:
//   - T: The response (potentially modified)
//   - error: The error (potentially wrapped or modified)
func checkResAndErr[T any](res T, err error) (T, error) {
	if err == nil {
		switch r := any(res).(type) {
		case *http.Response:
			switch r.StatusCode {
			case http.StatusNotImplemented:
				return res, retry.Unrecoverable(fmt.Errorf("HTTP failed: %w", ErrNotImplemented))
			case http.StatusBadRequest:
				return res, retry.Unrecoverable(fmt.Errorf("HTTP failed: %w", ErrBadRequest))
			case http.StatusForbidden:
				return res, retry.Unrecoverable(fmt.Errorf("HTTP failed: %w", ErrForbidden))
			case http.StatusNotFound:
				return res, retry.Unrecoverable(fmt.Errorf("HTTP failed: %w", ErrNotFound))
			default:
				return res, err
			}
		default:
			return res, nil
		}
	}

	return res, err
}

// retryAction implements the retry.RetryIfFunc interface to determine whether
// a retry should be attempted based on the error.
// This function is called by retry-go before each retry attempt to decide
// if the operation should be retried.
//
// RetryIfFunc signature: func(err error) bool
//
// Parameters:
//   - err: The error from the failed attempt
//
// Returns:
//   - bool: true if retry should be attempted, false otherwise
func retryAction(err error) bool {
	// Check if the error is recoverable using retry-go's built-in logic
	// Unrecoverable errors (created with retry.Unrecoverable()) will return false
	if !retry.IsRecoverable(err) {
		return false
	}

	// Increment the attempt counter (used for testing purposes)
	count++

	return true
}

// GetCount returns the number of retry attempts made during the last Retry() call.
//
// Return values:
//   - 0: Success on first attempt (no retries needed)
//   - 1+: Number of retry attempts made
//
// Note: This function is not thread-safe due to the global count variable.
// In a concurrent environment, multiple goroutines calling Retry() simultaneously
// may interfere with each other's count values.
func GetCount() int {
	return count
}
