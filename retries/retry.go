package retries

import (
	"context"
	"fmt"
	"net/http"

	"github.com/avast/retry-go"
)

// count tracks the number of retry attempts for the current operation.
// This is used primarily for testing purposes to verify retry behavior.
// Note: This is a global variable, which means it's not thread-safe, but since
// it's intended for testing scenarios, this is acceptable.
var count int

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
	var res T
	var err error
	count = 0
	err = retry.Do(
		func() error {
			res, err = f()
			res, err = checkResAndErr(res, err)
			return err
		},
		retry.Attempts(5),
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
	switch r := any(res).(type) {
	case *http.Response:
		switch {
		case r.StatusCode >= 500:
			return res, err
		case r.StatusCode >= 400:
			err = fmt.Errorf("HTTP failed - %d, Bad Request", r.StatusCode)
			return res, retry.Unrecoverable(err)
		}
	default:
		return res, nil
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
