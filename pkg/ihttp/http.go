package ihttp

import (
	"encoding/json"
	"fmt"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/imroc/req/v3"
)

// https://req.cool/zh/docs/prologue/introduction/

type ReqOpt struct {
	Method               string
	Url                  string
	Headers              map[string]string
	RespOut              interface{}
	RetryCount           int
	Body                 interface{}
	FormData             map[string]interface{}
	EnableForceMultipart bool
	NeedSkipTls          bool
	ProxyUrl             string
	UseCycleTLS          bool // 是否使用 CycleTLS
}

// 保持原有的 req client 用于普通请求
var cli = req.C().
	ImpersonateChrome().
	SetCookieJar(nil).
	DisableDumpAll()

// CycleTLS client
var cycleClient = cycletls.Init()

func DoRequest(opt *ReqOpt) (*req.Response, error) {
	// 原有的请求逻辑
	if opt.NeedSkipTls {
		cli.TLSClientConfig.InsecureSkipVerify = true
	}
	if opt.ProxyUrl != "" {
		cli.SetProxyURL(opt.ProxyUrl)
	}

	request := cli.NewRequest()
	if opt.RetryCount != 0 {
		request = request.SetRetryCount(opt.RetryCount)
	}
	if opt.Headers != nil {
		request = request.SetHeaders(opt.Headers)
	}
	if opt.RespOut != nil {
		request = request.SetSuccessResult(opt.RespOut)
	}
	if opt.Body != nil {
		request = request.SetBody(opt.Body)
	}
	if opt.EnableForceMultipart {
		request = request.EnableForceMultipart()
	}
	if opt.FormData != nil {
		request = request.SetFormDataAnyType(opt.FormData)
	}

	switch opt.Method {
	case "POST":
		return request.Post(opt.Url)
	case "PUT":
		return request.Put(opt.Url)
	case "DELETE":
		return request.Delete(opt.Url)
	case "OPTIONS":
		return request.Options(opt.Url)
	case "HEAD":
		return request.Head(opt.Url)
	default:
		return request.Get(opt.Url)
	}
}

// DoCycleTLSRequest 使用 CycleTLS 发送请求
func DoCycleTLSRequest(opt *ReqOpt) (cycletls.Response, error) {
	var response cycletls.Response

	// 准备请求数据
	var bodyStr string
	if opt.Body != nil {
		bodyBytes, err := json.Marshal(opt.Body)
		if err != nil {
			return response, fmt.Errorf("marshal body failed: %v", err)
		}
		bodyStr = string(bodyBytes)
	}

	// 设置 JA3 指纹，默认使用 Chrome 的指纹
	ja3 := "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-51-43-13-45-28-21,29-23-24-25-256-257,0"

	// 发送请求
	response, err := cycleClient.Do(opt.Url, cycletls.Options{
		Body:    bodyStr,
		Headers: opt.Headers,
		Ja3:     ja3,
		Method:  opt.Method,
		Proxy:   opt.ProxyUrl,
		Timeout: 30, // 默认30秒超时
	}, "chrome")

	if err != nil {
		return response, fmt.Errorf("cycletls request failed: %v", err)
	}

	// 如果需要解析响应到结构体
	if opt.RespOut != nil {
		err = json.Unmarshal([]byte(response.Body), opt.RespOut)
		if err != nil {
			return response, fmt.Errorf("unmarshal response failed: %v", err)
		}
	}

	return response, nil
}
