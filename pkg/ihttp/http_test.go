package ihttp

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	resp, err := DoCycleTLSRequest(&ReqOpt{
		Method: "GET",
		Url:    "https://gmgn.ai/api/v1/gas_price/eth",
	})
	fmt.Println(resp, err)
}
