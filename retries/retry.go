package retries

import (
	"context"
	"fmt"
	"net/http"

	"github.com/avast/retry-go"
)

var count int

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

func retryAction(err error) bool {
	if !retry.IsRecoverable(err) {
		return false
	}
	count++
	return true
}

func GetCount() int {
	return count
}
