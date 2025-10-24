package iemail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Covsj/gokit/ihttp"
	"github.com/Covsj/gokit/ilog"
	"github.com/Covsj/gokit/iutil"
)

type FakeCli struct {
	BaseUrl string

	CookieMap map[string]string

	Email     string
	EmailName string
}

// CliName implements IEmail
func (t *FakeCli) CliName() string {
	return "FakeMail"
}

func (t *FakeCli) Disconnect() error {
	return nil
}

func (t *FakeCli) dohttp(reqUrl, method string, rawBody map[string]any,
	out any) (*ihttp.Response, error) {
	headers := map[string]string{
		"user-agent": iutil.Random(),
	}
	if method != "GET" {
		headers["x-requested-with"] = "XMLHttpRequest"
	}

	var err error
	var resp *ihttp.Response

	for i := 0; i < 3; i++ {
		resp, err = ihttp.Do(&ihttp.Opt{
			URL:     reqUrl,
			Method:  method,
			Data:    rawBody,
			Cookies: &t.CookieMap,
			Headers: headers,
			RespOut: out,
		})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		resBody := resp.Text
		if resBody == "" {
			time.Sleep(2 * time.Second)
			continue
		}
		return resp, err
	}
	return resp, err
}

func (t *FakeCli) NewEmailCli(opt map[string]any) (IEmail, error) {
	t = &FakeCli{
		CookieMap: map[string]string{},
		BaseUrl:   FakeMailBaseURL,
	}
	if opt != nil {
		if baseUrl, ok := opt["baseUrl"].(string); ok && baseUrl != "" {
			t.BaseUrl = baseUrl
		}
	}
	_, err := t.dohttp(joinURL(t.BaseUrl, ""), "GET", nil, nil)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "首页", "Error", err.Error())
		return nil, err
	}
	// 生成邮箱并在服务端创建
	name := strings.ToLower(iutil.GenerateRandomStr(8, ""))
	t.EmailName = name
	// 默认域名；如需扩展，可从 GetDomains() 中随机选择
	t.Email = name + "@fontfee.com"
	_, err = t.dohttp(
		joinURL(t.BaseUrl, "/index/new-email/"),
		"POST", map[string]any{
			"emailInput": t.EmailName,
			"format":     "json",
		}, nil)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "new-email", "Error", err.Error())
		return nil, err
	}

	ilog.Info("邮箱初始化成功", "客户端类型", t.CliName(),
		"邮箱", t.Email)

	return t, nil
}

func (t *FakeCli) GetDomains() ([]string, error) {
	return []string{"fontfee.com"}, nil
}

func (t *FakeCli) GetEmailMsgs() (msgs []Msg, err error) {

	type fakeMsg struct {
		Akce            string `json:"akce"`
		OD              string `json:"od"`
		Predmet         string `json:"predmet"`
		PredmetZkraceny string `json:"predmetZkraceny"`
	}
	out := []fakeMsg{}

	resp, err := t.dohttp(
		joinURL(t.BaseUrl, "/index/refresh"), "GET", nil, nil)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "refresh", "Error", err.Error())
		return nil, err
	}
	if resp != nil && resp.IsSuccess() {
		respBytes := resp.Body
		err = safeJSONUnmarshal(respBytes, &out)
		if err != nil {
			ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
				"逻辑接口", "safeJSONUnmarshal", "Error", err.Error())
			return nil, err
		}
	}

	msgs = []Msg{}
	for _, m := range out {
		msgs = append(msgs, Msg{
			From:    m.OD,
			To:      t.Email,
			Subject: m.Predmet,
			Body:    m.PredmetZkraceny,
		})
	}

	ilog.Info("邮箱获取邮件成功", "邮箱", t.Email, "邮件数量", len(msgs))
	return msgs, nil
}

// Data implements IEmail
func (t *FakeCli) Data() map[string]any {
	return map[string]any{
		"email":     t.Email,
		"emailName": t.EmailName,
	}
}

// safeJSONUnmarshal 安全地解析 JSON，提供详细的错误信息
func safeJSONUnmarshal(data []byte, v interface{}) error {
	// 首先尝试直接解析
	err := json.Unmarshal(data, v)
	if err == nil {
		return nil
	}
	// cleanJSONResponse 清理响应内容，移除 BOM 和其他非 JSON 字符
	cleanJSONResponse := func(data []byte) []byte {
		// 移除 BOM (Byte Order Mark)
		if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
			data = data[3:]
		}

		// 转换为字符串并清理
		str := string(data)

		// 移除前后空白字符
		str = strings.TrimSpace(str)

		// 查找第一个 '{' 或 '[' 字符
		start := -1
		for i, r := range str {
			if r == '{' || r == '[' {
				start = i
				break
			}
		}

		if start == -1 {
			return data // 如果没有找到 JSON 开始字符，返回原数据
		}

		// 从第一个 JSON 字符开始截取
		str = str[start:]

		// 查找最后一个 '}' 或 ']' 字符
		end := -1
		for i := len(str) - 1; i >= 0; i-- {
			if str[i] == '}' || str[i] == ']' {
				end = i + 1
				break
			}
		}

		if end == -1 {
			return data // 如果没有找到 JSON 结束字符，返回原数据
		}

		str = str[:end]

		// 验证是否为有效的 UTF-8
		if !utf8.ValidString(str) {
			return data // 如果不是有效的 UTF-8，返回原数据
		}

		return []byte(str)
	}

	// 如果失败，尝试清理后解析
	cleanData := cleanJSONResponse(data)
	err = json.Unmarshal(cleanData, v)
	if err == nil {
		//ilog.Info("JSON解析成功", "原始数据长度", len(data), "清理后数据长度", len(cleanData))
		return nil
	}

	return fmt.Errorf("JSON解析失败: %v, 原始数据: %s, 清理后数据: %s", err, string(data), string(cleanData))
}
