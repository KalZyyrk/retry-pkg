package retries

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	tt := []struct {
		name          string
		f             func() (any, error)
		expectedRetry int
		isError       bool
	}{
		{
			name: "functional error",
			f: func() (any, error) {
				res := http.Response{
					StatusCode: 400,
				}
				return &res, nil
			},
			isError:       true,
			expectedRetry: 0,
		},
		{
			name: "Network issue 3 retries",
			f: func() (any, error) {
				count = 0
				if count < 3 {
					return &http.Response{
						StatusCode: 200,
					}, nil
				}
				count++
				return &http.Response{
					StatusCode: 500,
				}, nil

			},
		},
	}

	ctx := context.TODO()

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			res, err := Retry(ctx, test.f)

			if test.isError {
				assert.Equal(t, test.expectedRetry, GetCount())
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expectedRetry, GetCount())
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
			}
		})
	}
}
