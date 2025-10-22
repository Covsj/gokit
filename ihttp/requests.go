package ihttp

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Covsj/gokit/ilog"
	"github.com/Covsj/requests"
	"github.com/Covsj/requests/models"
	requests_url "github.com/Covsj/requests/url"
)

type Opt struct {
	URL     string
	Proxy   string
	Method  string
	TimeOut int

	Data  map[string]any
	Json  map[string]any
	Files *requests_url.Files

	Headers map[string]string
	Cookies map[string]string

	RespOut any

	// 默认安全：开启证书校验与重定向；指纹随机化默认关闭
	AllowRedirects int // 0为true，默认打开
	Verify         int // 0为true，默认打开
	RandomJA3      int // 0为false，默认关闭

	NotLog bool
}

// getProxyFromEnv 从环境变量获取代理设置
func getProxyFromEnv(targetURL string) string {
	// 解析目标URL以确定协议
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	var proxyEnv string
	if parsedURL.Scheme == "https" {
		proxyEnv = os.Getenv("HTTPS_PROXY")
		if proxyEnv == "" {
			proxyEnv = os.Getenv("https_proxy")
		}
	} else {
		proxyEnv = os.Getenv("HTTP_PROXY")
		if proxyEnv == "" {
			proxyEnv = os.Getenv("http_proxy")
		}
	}

	// 如果没有找到协议特定的代理，尝试通用代理
	if proxyEnv == "" {
		proxyEnv = os.Getenv("ALL_PROXY")
		if proxyEnv == "" {
			proxyEnv = os.Getenv("all_proxy")
		}
	}

	return proxyEnv
}

func Do(opt *Opt) (resp *models.Response, err error) {
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
	var reqBody any
	if opt.Data != nil {
		reqBody = opt.Data
	} else if opt.Json != nil {
		reqBody = opt.Json
	} else if opt.Files != nil {
		reqBody = opt.Files
	} else {
		reqBody = nil
	}

	// 校验 Data/Json/Files 互斥
	bodyKinds := 0
	if len(opt.Data) > 0 {
		bodyKinds++
	}
	if len(opt.Json) > 0 {
		bodyKinds++
	}
	if opt.Files != nil {
		bodyKinds++
	}
	if bodyKinds > 1 {
		return nil, errors.New("只能设置Data/Json/Files其中一种请求")
	}

	URL := opt.URL
	var jsonErr error
	defer func() {
		status := 500
		respBody := ""
		if resp != nil {
			status = resp.StatusCode
			// 仅在错误或非2xx时记录响应体，并限制长度
			if err != nil || status >= 400 {
				rb := resp.Text
				if len(rb) > 1024 {
					rb = rb[:1024] + "...(省略)"
				}
				respBody = rb
			}
		}
		elapsed := time.Since(start)
		args := []any{
			"方法", opt.Method,
			"URL", opt.URL,
			"请求体", reqBody,
			"请求Header", opt.Headers,
			"响应码", status,
			"耗时", elapsed.String(),
		}
		if respBody != "" {
			args = append(args, "响应体", respBody)
		}
		if err != nil {
			args = append(args, "错误信息", err)
		}
		if jsonErr != nil {
			args = append(args, "Unmarshal失败", jsonErr)
		}
		if !opt.NotLog {
			ilog.Debug(
				"调用请求", args...,
			)
		}

	}()

	req := &requests_url.Request{
		AllowRedirects: opt.AllowRedirects == 0,
		Verify:         opt.Verify == 0,
		RandomJA3:      opt.RandomJA3 != 0,
		Headers:        requests_url.NewHeaders(),
	}

	for k, v := range opt.Headers {
		req.Headers.Set(k, v)
	}

	if len(opt.Cookies) > 0 {
		req.Cookies = requests_url.ParseCookies(URL, opt.Cookies)
	}

	// 设置代理：优先使用显式设置的代理，否则从环境变量获取
	proxy := opt.Proxy
	if proxy == "" {
		proxy = getProxyFromEnv(URL)
	}
	if proxy != "" {
		req.Proxies = proxy
	}
	if opt.TimeOut == 0 {
		opt.TimeOut = 600
	}
	req.Timeout = time.Duration(opt.TimeOut) * time.Second

	if len(opt.Data) != 0 {
		req.Data = requests_url.ParseData(opt.Data)
	} else if len(opt.Json) != 0 {
		req.Json = opt.Json
	} else if opt.Files != nil {
		// SetFile(name,fileName,filePath,contentType)
		// name为字段名，fileName为上传的文件名，filePath为上传文件的绝对路径，contentType为上传的文件类型
		// 如果contentType设置为""，则默认为"application/octet-stream"
		// files := url.NewFiles()
		// files.SetFile("api", "api", "D:\\Go\\github.com\\wangluozhe\\requests\\api.go", "")
		// req := url.NewRequest()
		req.Files = opt.Files
	}

	switch method {
	case "GET":
		resp, err = requests.Get(URL, req)
	case "POST":
		resp, err = requests.Post(URL, req)
	case "OPTIONS":
		resp, err = requests.Options(URL, req)
	case "HEAD":
		resp, err = requests.Head(URL, req)
	case "PUT":
		resp, err = requests.Put(URL, req)
	case "PATCH":
		resp, err = requests.Patch(URL, req)
	case "DELETE":
		resp, err = requests.Delete(URL, req)
	case "CONNECT":
		resp, err = requests.Connect(URL, req)
	case "TRACE":
		resp, err = requests.Trace(URL, req)
	default:
		resp, err = requests.Get(URL, req)
	}
	if err != nil || resp == nil {
		return
	}
	if opt.RespOut != nil && resp.Ok() && resp.Content != nil {
		// var js *simplejson.Json
		// var b []byte
		// js, jsonErr := resp.SimpleJson()
		// if jsonErr != nil {
		// 	return
		// }
		// b, jsonErr = js.MarshalJSON()
		// if jsonErr != nil {
		// 	return
		// }
		jsonErr = json.Unmarshal(resp.Content, opt.RespOut)
		if jsonErr != nil {
			ilog.Error("反序列化失败", "原始数据", string(resp.Content), "传入结构", opt.RespOut, "错误信息", jsonErr)
			return
		}
	}
	return resp, err
}
