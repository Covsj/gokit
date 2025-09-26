package ihttp

import (
	"fmt"
	"testing"
)

type ss struct {
	CContext string `json:"@context"`
}

func TestHttp1(t *testing.T) {
	var out any
	resp, err := Do(&Opt{
		Method:  "GET",
		URL:     "https://api.mail.gw/domains",
		RespOut: &out,
	})
	if resp != nil {
		fmt.Println(resp.Text, err, out)
	} else {
		fmt.Println("resp is nil", err, out)
	}
}
