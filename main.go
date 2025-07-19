package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/avast/retry-go"
)

func main() {
	url := "http://example.com"
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(body)
}
