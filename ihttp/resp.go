package ihttp

import "net/http"

// Response 响应结构
type Response struct {
	StatusCode int
	Body       []byte
	Text       string
	Headers    map[string][]string
	CookieList []*http.Cookie
}

// IsSuccess 检查响应是否成功
func (r *Response) IsSuccess() bool {
	if r == nil {
		return false
	}
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return false
	}
	return true
}
