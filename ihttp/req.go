package ihttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Covsj/gokit/ilog"
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
	if opt.Json != nil {
		bodyKinds++
	}
	if len(opt.Files) > 0 {
		bodyKinds++
	}
	if bodyKinds > 1 {
		return nil, errors.New("只能设置Data/Json/Files其中一种请求")
	}

	// 准备请求体用于日志记录
	var logBody any
	if len(opt.Data) > 0 {
		logBody = opt.Data
	} else if opt.Json != nil {
		logBody = opt.Json
	} else if len(opt.Files) > 0 {
		logBody = opt.Files
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
				"请求体", logBody,
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

	var client *http.Client

	if opt.HttpCLi != nil {
		client = opt.HttpCLi
	} else {
		// 创建HTTP客户端
		client = &http.Client{}
	}

	if opt.TimeOut > 0 {
		client.Timeout = time.Duration(opt.TimeOut) * time.Second
	}

	// 创建请求
	var reqBody io.Reader
	var contentType string

	// 设置请求体
	if len(opt.Data) > 0 {
		// Form数据
		formData := url.Values{}
		for k, v := range opt.Data {
			formData.Set(k, fmt.Sprintf("%v", v))
		}
		reqBody = strings.NewReader(formData.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if opt.Json != nil {
		// JSON数据
		jsonData, err := json.Marshal(opt.Json)
		if err != nil {
			return nil, fmt.Errorf("JSON序列化失败: %v", err)
		}
		reqBody = bytes.NewReader(jsonData)
		contentType = "application/json"
	} else if len(opt.Files) > 0 {
		// 文件上传
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		for fieldName, file := range opt.Files {
			var fileReader io.Reader
			var fileName string

			if file.Reader != nil {
				fileReader = file.Reader
				fileName = file.FileName
			} else if file.Path != "" {
				fileObj, err := os.Open(file.Path)
				if err != nil {
					return nil, fmt.Errorf("打开文件失败: %v", err)
				}
				defer fileObj.Close()
				fileReader = fileObj
				fileName = file.FileName
				if fileName == "" {
					fileName = filepath.Base(file.Path)
				}
			}

			if fileReader != nil {
				part, err := writer.CreateFormFile(fieldName, fileName)
				if err != nil {
					return nil, fmt.Errorf("创建文件字段失败: %v", err)
				}
				_, err = io.Copy(part, fileReader)
				if err != nil {
					return nil, fmt.Errorf("复制文件内容失败: %v", err)
				}
			}
		}

		writer.Close()
		reqBody = bytes.NewReader(buf.Bytes())
		contentType = writer.FormDataContentType()
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, opt.URL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	// 设置Cookies
	if opt.Cookies != nil && len(*opt.Cookies) > 0 {
		for name, value := range *opt.Cookies {
			req.AddCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 处理响应
	response = &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Text:       string(body),
		Headers:    resp.Header,
		CookieList: []*http.Cookie{},
	}

	// 自动更新Cookies
	if opt.Cookies != nil {
		cklist := updateCookiesFromResponse(opt.Cookies, resp)

		response.CookieList = cklist
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
