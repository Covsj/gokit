package ihttp

import (
	"fmt"
	"testing"
)

// TestProxyPriority 测试代理优先级
func TestProxyPriority(t *testing.T) {
	var out any
	// 测试显式代理优先于环境变量
	resp, err := Do(&Opt{
		Method:  "GET",
		URL:     "https://clob.polymarket.com/",
		RespOut: &out,
	})
	if resp != nil {
		fmt.Println(resp.Text, err, out)
	} else {
		fmt.Println("resp is nil", err, out)
	}
}
