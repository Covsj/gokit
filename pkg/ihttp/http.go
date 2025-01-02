package ihttp

import (
	"fmt"

	"github.com/imroc/req/v3"
)

// RequestOptions 定义 HTTP 请求的配置选项
type RequestOptions struct {
	Method  string            // HTTP 方法: GET, POST, PUT 等
	URL     string            // 请求 URL
	Headers map[string]string // 请求头
	Body    interface{}       // 请求体

	// 响应相关
	ResponseOut interface{} // 用于解析响应的结构体指针

	// 请求配置
	RetryCount     int                    // 重试次数
	FormData       map[string]interface{} // 表单数据
	ForceMultipart bool                   // 强制使用 multipart 格式

	// TLS 和代理配置
	SkipTLSVerify bool   // 是否跳过 TLS 验证
	ProxyURL      string // 代理服务器 URL
	UseCycleTLS   bool   // 是否使用 CycleTLS
}

// 默认的 HTTP 客户端实例
var defaultClient = req.C().
	ImpersonateSafari().
	SetCookieJar(nil).
	DisableDumpAll()

// DoRequest 执行 HTTP 请求
func DoRequest(opts *RequestOptions) (*req.Response, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	// 配置请求客户端
	if opts.SkipTLSVerify {
		defaultClient.TLSClientConfig.InsecureSkipVerify = true
	}

	if opts.ProxyURL != "" {
		defaultClient.SetProxyURL(opts.ProxyURL)
	}

	// 构建请求
	req := defaultClient.NewRequest()

	// 设置请求选项
	if opts.RetryCount > 0 {
		req = req.SetRetryCount(opts.RetryCount)
	}

	if opts.Headers != nil {
		req = req.SetHeaders(opts.Headers)
	}

	if opts.ResponseOut != nil {
		req = req.SetSuccessResult(opts.ResponseOut)
	}

	if opts.Body != nil {
		req = req.SetBody(opts.Body)
	}

	if opts.ForceMultipart {
		req = req.EnableForceMultipart()
	}

	if opts.FormData != nil {
		req = req.SetFormDataAnyType(opts.FormData)
	}

	// 执行请求
	switch opts.Method {
	case "POST":
		return req.Post(opts.URL)
	case "PUT":
		return req.Put(opts.URL)
	case "DELETE":
		return req.Delete(opts.URL)
	case "OPTIONS":
		return req.Options(opts.URL)
	case "HEAD":
		return req.Head(opts.URL)
	default:
		return req.Get(opts.URL)
	}
}
