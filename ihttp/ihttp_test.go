package ihttp

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHttp(t *testing.T) {
	var out map[string]any
	cookies := make(map[string]string)

	// 使用链式调用创建配置
	opt := NewOpt()
	opt.Cookies = &cookies
	opt.RespOut = &out
	opt.Headers = map[string]string{
		"User-Agent": "ihttp-test",
	}
	opt.HttpCLi = &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := Get("https://clob.polymarket.com/book?token_id=70224002415726915146697406828863644162763565870559027191380082229342088681891", opt)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if !resp.IsSuccess() {
		t.Fatalf("Expected success status, got %d", resp.StatusCode)
	}

	fmt.Printf("Cookies after request: %+v\n", cookies)
	fmt.Printf("Response: %+v\n", out)
}
