package ihttp

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// convertToStringMap 将map[string]any转换为map[string]string
func convertToStringMap(data map[string]any) map[string]string {
	result := make(map[string]string)
	for k, v := range data {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// updateCookiesFromResponse 从响应中更新Cookies
func updateCookiesFromResponse(cookies *map[string]string,
	resp *http.Response) (ckList []*http.Cookie) {

	ckList = []*http.Cookie{}
	if cookies == nil {
		return
	}

	// 初始化cookies map如果为nil
	if *cookies == nil {
		*cookies = make(map[string]string)
	}

	// 从响应中提取Set-Cookie头
	for _, cookieStr := range resp.Header["Set-Cookie"] {
		if cookieStr == "" {
			continue
		}

		// 解析cookie字符串
		cookie := ParseCookieString(cookieStr)
		if cookie != nil {
			(*cookies)[cookie.Name] = cookie.Value
			ckList = append(ckList, cookie)
		}
	}
	return ckList
}

// parseCookieString 解析cookie字符串
func ParseCookieString(cookieStr string) *http.Cookie {
	parts := strings.Split(cookieStr, ";")
	if len(parts) == 0 {
		return nil
	}

	// 解析name=value部分
	nameValue := strings.TrimSpace(parts[0])
	if nameValue == "" {
		return nil
	}

	equalIndex := strings.Index(nameValue, "=")
	if equalIndex == -1 {
		return nil
	}

	name := strings.TrimSpace(nameValue[:equalIndex])
	value := strings.TrimSpace(nameValue[equalIndex+1:])

	if name == "" {
		return nil
	}

	return &http.Cookie{
		Name:  name,
		Value: value,
	}
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
