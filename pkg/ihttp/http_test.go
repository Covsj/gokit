package ihttp

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	resp, err := DoRequest(&RequestOptions{
		Method: "GET",
		URL:    "https://gmgn.ai/defi/quotation/v1/signals?size=10&better=true",
	})
	fmt.Println(resp, err)
}
