package ihttp

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	resp, err := DoRequest(&Options{
		Method: "GET",
		URL:    "https://api.mail.gw/domains?page=1",
	})
	fmt.Println(resp, err)
}
