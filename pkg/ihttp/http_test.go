package ihttp

import (
	"fmt"
	"testing"
)

func TestHttp(t *testing.T) {
	resp, err := DoRequest(&RequestOptions{
		Method: "GET",
		URL:    "https://debot.ai/api/dashboard/chain/recommend/hot_token?chain=solana&duration=1H&sort_field=score&sort_order=desc&filter=%7B%7D&is_hide_honeypot=true",
	})
	fmt.Println(resp, err)
}
