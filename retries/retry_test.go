package retries_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"retry-pkg/retries"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		name          string
		f             func() (any, error)
		expectedRetry int
		isError       bool
	}{
		{
			name: "functional error",
			f: func() (any, error) {
				res := http.Response{
					StatusCode: http.StatusBadRequest,
				}

				return &res, nil
			},
			isError:       true,
			expectedRetry: 0,
		},
		{
			name: "Network issue 3 retries",
			f: func() (any, error) {
				count := 0
				if count < 3 {
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				}
				count++

				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil
			},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := retries.Retry(ctx, test.f)

			if test.isError {
				assert.Equal(t, test.expectedRetry, retries.GetCount())
				require.Error(t, err)
			} else {
				assert.Equal(t, test.expectedRetry, retries.GetCount())
				require.NoError(t, err)
				assert.NotEmpty(t, res)
			}
		})
	}
}
