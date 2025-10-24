package ihttp

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Covsj/gokit/ilog"
	"github.com/go-resty/resty/v2"
)

// Do 执行HTTP请求
func Do(opt *Opt) (*Response, error) {
	if opt == nil {
		return nil, errors.New("空配置项")
	}
	if opt.URL == "" {
		return nil, errors.New("空请求链接")
	}

	start := time.Now()
	if opt.Method == "" {
		opt.Method = "GET"
	}
	method := strings.ToUpper(opt.Method)

	// 校验请求体类型互斥
	bodyKinds := 0
	if len(opt.Data) > 0 {
		bodyKinds++
	}
	if len(opt.Json) > 0 {
		bodyKinds++
	}
	if len(opt.Files) > 0 {
		bodyKinds++
	}
	if bodyKinds > 1 {
		return nil, errors.New("只能设置Data/Json/Files其中一种请求")
	}

	// 准备请求体用于日志记录
	var reqBody any
	if len(opt.Data) > 0 {
		reqBody = opt.Data
	} else if len(opt.Json) > 0 {
		reqBody = opt.Json
	} else if len(opt.Files) > 0 {
		reqBody = opt.Files
	}

	// 使用defer确保日志记录
	var response *Response
	var err error
	defer func() {
		if !opt.NotLog {
			elapsed := time.Since(start)
			args := []any{
				"方法", opt.Method,
				"URL", opt.URL,
				"请求体", reqBody,
				"请求Header", opt.Headers,
				"耗时", elapsed.String(),
			}

			if response != nil {
				args = append(args, "响应码", response.StatusCode)

				// 仅在错误或非2xx时记录响应体
				if err != nil || response.StatusCode >= 400 {
					rb := response.Text
					if len(rb) > 1024 {
						rb = rb[:1024] + "...(省略)"
					}
					args = append(args, "响应体", rb)
				}
			} else {
				args = append(args, "响应码", 0)
			}

			if err != nil {
				args = append(args, "错误信息", err)
			}

			ilog.Debug("调用请求", args...)
		}
	}()

	// 创建resty客户端
	client := resty.New()

	// 设置超时
	if opt.TimeOut == 0 {
		opt.TimeOut = 600
	}
	client.SetTimeout(time.Duration(opt.TimeOut) * time.Second)

	// 设置代理
	proxy := opt.Proxy
	// if proxy == "" {
	// 	proxy = getProxyFromEnv(opt.URL)
	// }
	if proxy != "" {
		client.SetProxy(proxy)
	}

	// 设置重定向
	if !opt.AllowRedirects {
		client.SetRedirectPolicy(resty.NoRedirectPolicy())
	}

	// 设置SSL验证
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: opt.SkipVerify,
	})

	// 设置请求头
	if len(opt.Headers) > 0 {
		client.SetHeaders(opt.Headers)
	}

	// 创建请求
	req := client.R()

	// 设置Cookies
	if opt.Cookies != nil && len(*opt.Cookies) > 0 {
		for name, value := range *opt.Cookies {
			req.SetCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}

	// 设置请求体
	if len(opt.Data) > 0 {
		req.SetFormData(convertToStringMap(opt.Data))
	} else if len(opt.Json) > 0 {
		req.SetBody(opt.Json)
	} else if len(opt.Files) > 0 {
		// 处理文件上传
		for fieldName, file := range opt.Files {
			if file.Reader != nil {
				// 使用Reader
				req.SetFileReader(fieldName, file.FileName, file.Reader)
			} else if file.Path != "" {
				// 使用文件路径
				req.SetFile(fieldName, file.Path)
			}
		}
	}

	// 执行请求
	var resp *resty.Response

	switch method {
	case "GET":
		resp, err = req.Get(opt.URL)
	case "POST":
		resp, err = req.Post(opt.URL)
	case "PUT":
		resp, err = req.Put(opt.URL)
	case "PATCH":
		resp, err = req.Patch(opt.URL)
	case "DELETE":
		resp, err = req.Delete(opt.URL)
	case "HEAD":
		resp, err = req.Head(opt.URL)
	case "OPTIONS":
		resp, err = req.Options(opt.URL)
	case "CONNECT":
		// CONNECT方法在resty中不直接支持，使用Execute方法
		resp, err = req.Execute("CONNECT", opt.URL)
	case "TRACE":
		// TRACE方法在resty中不直接支持，使用Execute方法
		resp, err = req.Execute("TRACE", opt.URL)
	default:
		resp, err = req.Get(opt.URL)
	}

	// 处理响应
	if resp != nil {
		response = &Response{
			StatusCode: resp.StatusCode(),
			Body:       resp.Body(),
			Text:       resp.String(),
			Headers:    resp.Header(),
		}

		// 自动更新Cookies
		if opt.Cookies != nil {
			updateCookiesFromResponse(opt.Cookies, resp)
		}
	}

	// 处理响应体反序列化
	if opt.RespOut != nil && response != nil &&
		response.IsSuccess() &&
		len(response.Body) > 0 {
		if jsonErr := json.Unmarshal(response.Body,
			opt.RespOut); jsonErr != nil {
			ilog.Error("反序列化失败", "原始数据", response.Text,
				"传入结构", opt.RespOut, "错误信息", jsonErr)
			return response, jsonErr
		}
	}

	return response, err
}

// 便捷方法 - 支持多种HTTP方法

// Get 执行GET请求
func Get(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "GET"
	return Do(opt)
}

// Post 执行POST请求
func Post(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "POST"
	return Do(opt)
}

// Put 执行PUT请求
func Put(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "PUT"
	return Do(opt)
}

// Patch 执行PATCH请求
func Patch(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "PATCH"
	return Do(opt)
}

// Delete 执行DELETE请求
func Delete(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "DELETE"
	return Do(opt)
}

// Head 执行HEAD请求
func Head(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "HEAD"
	return Do(opt)
}

// Options 执行OPTIONS请求
func Options(url string, opt *Opt) (*Response, error) {
	if opt == nil {
		opt = NewOpt()
	}
	opt.URL = url
	opt.Method = "OPTIONS"
	return Do(opt)
}

// 便捷的JSON请求方法

// PostJSON 发送JSON POST请求
func PostJSON(url string, jsonData map[string]any, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Json = jsonData
	opt.RespOut = respOut
	opt.Headers["Content-Type"] = "application/json"
	return Post(url, opt)
}

// PutJSON 发送JSON PUT请求
func PutJSON(url string, jsonData map[string]any, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Json = jsonData
	opt.RespOut = respOut
	opt.Headers["Content-Type"] = "application/json"
	return Put(url, opt)
}

// PatchJSON 发送JSON PATCH请求
func PatchJSON(url string, jsonData map[string]any, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Json = jsonData
	opt.RespOut = respOut
	opt.Headers["Content-Type"] = "application/json"
	return Patch(url, opt)
}

// 便捷的Form请求方法

// PostForm 发送Form POST请求
func PostForm(url string, formData map[string]any, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Data = formData
	opt.RespOut = respOut
	opt.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return Post(url, opt)
}

// PutForm 发送Form PUT请求
func PutForm(url string, formData map[string]any, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Data = formData
	opt.RespOut = respOut
	opt.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return Put(url, opt)
}

// 便捷的文件上传方法

// PostFile 发送文件POST请求
func PostFile(url string, files map[string]File, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Files = files
	opt.RespOut = respOut
	return Post(url, opt)
}

// PutFile 发送文件PUT请求
func PutFile(url string, files map[string]File, respOut any) (*Response, error) {
	opt := NewOpt()
	opt.Files = files
	opt.RespOut = respOut
	return Put(url, opt)
}
