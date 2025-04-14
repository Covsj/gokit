package ihttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/valyala/fasthttp"
)

// New 创建一个新的 HTTP 客户端
func New() *Client {
	client := &fasthttp.Client{
		MaxConnsPerHost:               100,
		MaxIdleConnDuration:           30 * time.Second,
		ReadTimeout:                   30 * time.Second,
		WriteTimeout:                  30 * time.Second,
		MaxResponseBodySize:           10 * 1024 * 1024, // 10MB
		DisableHeaderNamesNormalizing: false,
		DisablePathNormalizing:        false,
	}
	cli := &Client{
		client: client,
		options: &Options{
			UserAgent:     DefaultUA,
			RetryCount:    3,
			RetryInterval: 2 * time.Second,
			Headers: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
		},
	}
	cli.SetTimeout(60 * time.Second)

	return cli
}

// R 创建一个新的请求
func (c *Client) R(opt *Options) *Request {
	return &Request{
		client:  c,
		options: opt,
	}
}

// 请求配置方法

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.ReadTimeout = timeout
	c.client.WriteTimeout = timeout
	c.options.Timeout = timeout
	return c
}

// SetSkipTLSVerify 设置是否跳过 TLS 验证
func (c *Client) SetSkipTLSVerify(skip bool) *Client {
	c.client.TLSConfig = &tls.Config{
		InsecureSkipVerify: skip,
	}
	c.options.SkipTLSVerify = skip
	return c
}

// SetDebugMode 设置是否启用调试模式
func (c *Client) SetDebugMode(debugMode bool) *Client {
	c.options.DebugMode = debugMode
	return c
}

// Send 发送请求
func (r *Request) Send() (*Response, error) {
	opts := r.options
	// 验证必要参数
	if opts.URL == "" {
		return nil, errors.New("URL cannot be empty")
	}

	var err error
	var resp *Response

	// 执行重试逻辑
	for i := 0; i <= opts.RetryCount; i++ {
		if i > 0 {
			// 等待重试
			time.Sleep(opts.RetryInterval)
		}

		// 执行请求
		resp, err = r.doRequest(opts)

		// 检查是否需要重试
		if err == nil {
			break
		}
	}

	return resp, err
}

// 执行 HTTP 请求
func (r *Request) doRequest(opts *Options) (*Response, error) {
	// 创建请求对象
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// 创建响应对象
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 设置请求 URL
	reqURL := opts.URL

	// 添加查询参数
	if len(opts.Query) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return nil, err
		}

		q := u.Query()
		for k, v := range opts.Query {
			q.Set(k, v)
		}

		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	req.SetRequestURI(reqURL)

	// 设置 HTTP 方法
	req.Header.SetMethod(opts.Method)

	// 设置请求头
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	// 设置用户代理
	if opts.UserAgent != "" {
		req.Header.SetUserAgent(opts.UserAgent)
	}

	if opts.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.BearerToken)
	}

	// 设置请求体
	if opts.Body != nil {
		// 根据类型转换成字节
		switch body := opts.Body.(type) {
		case string:
			req.SetBodyString(body)
		case []byte:
			req.SetBody(body)
		default:
			// 尝试 JSON 序列化
			b, err := json.Marshal(opts.Body)
			if err != nil {
				return nil, err
			}
			req.SetBody(b)
			req.Header.SetContentType("application/json")
		}
	} else if opts.JSONBody != nil {
		b, err := json.Marshal(opts.JSONBody)
		if err != nil {
			return nil, err
		}
		req.SetBody(b)
		req.Header.SetContentType("application/json")
	} else if opts.XMLBody != nil {
		b, err := xml.Marshal(opts.XMLBody)
		if err != nil {
			return nil, err
		}
		req.SetBody(b)
		req.Header.SetContentType("application/xml")
	} else if opts.FormData != nil || len(opts.Files) > 0 {
		// 处理 multipart/form-data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加表单字段
		if opts.FormData != nil {
			for k, v := range opts.FormData {
				switch val := v.(type) {
				case string:
					_ = writer.WriteField(k, val)
				default:
					b, err := json.Marshal(val)
					if err != nil {
						return nil, err
					}
					_ = writer.WriteField(k, string(b))
				}
			}
		}

		// 添加文件
		for field, filePath := range opts.Files {
			file, err := os.Open(filePath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(field, filepath.Base(filePath))
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return nil, err
			}
		}

		// 关闭 writer
		err := writer.Close()
		if err != nil {
			return nil, err
		}

		req.SetBody(body.Bytes())
		req.Header.SetContentType(writer.FormDataContentType())
	} else if opts.FormURLEncoded != nil {
		// 处理 application/x-www-form-urlencoded
		form := url.Values{}
		for k, v := range opts.FormURLEncoded {
			form.Add(k, v)
		}
		req.SetBodyString(form.Encode())
		req.Header.SetContentType("application/x-www-form-urlencoded")
	} else if opts.RawBody != "" {
		req.SetBodyString(opts.RawBody)
	} else if opts.BinaryBody != nil {
		req.SetBody(opts.BinaryBody)
	} else if opts.GraphQLQuery != "" {
		// 处理 GraphQL 请求
		graphqlBody := map[string]interface{}{
			"query": opts.GraphQLQuery,
		}

		if opts.GraphQLVars != nil {
			graphqlBody["variables"] = opts.GraphQLVars
		}

		b, err := json.Marshal(graphqlBody)
		if err != nil {
			return nil, err
		}

		req.SetBody(b)
		req.Header.SetContentType("application/json")
	}

	// 设置客户端超时
	client := r.client.client
	if opts.Timeout > 0 {
		client.ReadTimeout = opts.Timeout
		client.WriteTimeout = opts.Timeout
	}

	// 发送请求
	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	// 创建响应对象
	response := &Response{
		StatusCode:    resp.StatusCode(),
		Body:          resp.Body(),
		ContentLength: resp.Header.ContentLength(),
		Headers:       make(map[string]string),
		Raw:           resp,
	}

	// 复制响应头
	resp.Header.VisitAll(func(key, value []byte) {
		response.Headers[string(key)] = string(value)
	})

	// 检查状态码是否在预期内
	if len(opts.ExpectedStatus) > 0 {
		statusOK := false
		for _, status := range opts.ExpectedStatus {
			if response.StatusCode == status {
				statusOK = true
				break
			}
		}

		if !statusOK {
			return response, fmt.Errorf("unexpected status code: %d", response.StatusCode)
		}
	}

	// 如果提供了 ResponseOut，解析响应
	if opts.ResponseOut != nil {
		contentType := string(resp.Header.ContentType())

		if contentType == "application/json" || contentType == "application/json; charset=utf-8" {
			err = json.Unmarshal(response.Body, opts.ResponseOut)
		} else if contentType == "application/xml" || contentType == "application/xml; charset=utf-8" {
			err = xml.Unmarshal(response.Body, opts.ResponseOut)
		} else {
			// 尝试 JSON 解析
			err = json.Unmarshal(response.Body, opts.ResponseOut)
		}

		if err != nil {
			return response, err
		}
	}

	return response, nil
}

// DoRequest 执行 HTTP 请求 (兼容旧的 API)
func DoRequest(opts *Options) (*Response, error) {
	client := New()
	req := client.R(opts)

	if opts.SkipTLSVerify {
		client.SetSkipTLSVerify(true)
	}

	if opts.DebugMode {
		client.SetDebugMode(true)
	}

	// 执行请求
	return req.Send()
}

// JSON 解析响应为 JSON
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// XML 解析响应为 XML
func (r *Response) XML(v interface{}) error {
	return xml.Unmarshal(r.Body, v)
}

// String 返回响应体作为字符串
func (r *Response) String() string {
	return string(r.Body)
}

// GetHeader 获取指定响应头
func (r *Response) GetHeader(key string) string {
	return r.Headers[key]
}

// IsSuccess 检查响应是否成功 (2xx)
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError 检查响应是否为错误 (4xx 或 5xx)
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// IsClientError 检查响应是否为客户端错误 (4xx)
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 检查响应是否为服务器错误 (5xx)
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500
}
