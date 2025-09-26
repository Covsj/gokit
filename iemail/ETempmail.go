package iemail

import (
	"github.com/Covsj/gokit/ihttp"
	"github.com/Covsj/gokit/ilog"
	"github.com/Covsj/requests/models"
)

// https://etempmail.com/
type ETempMailCli struct {
	BaseUrl string

	CookieMap map[string]string

	Email      string
	EmailId    string
	RecoverKey string
}

// CliName implements IEmail
func (t *ETempMailCli) CliName() string {
	return "ETempMail"
}

func (t *ETempMailCli) Disconnect() error {
	return nil
}

func (t *ETempMailCli) dohttp(reqUrl, method string, rawBody map[string]any, out any) (*models.Response, error) {
	headers := map[string]string{
		"user-agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
		"x-requested-with": "XMLHttpRequest",
	}
	var err error
	var resp *models.Response
	resp, err = ihttp.Do(&ihttp.Opt{
		URL:     reqUrl,
		Method:  method,
		Json:    rawBody,
		Cookies: t.CookieMap,
		Headers: headers,
		RespOut: out,
	})
	if err != nil {
		return resp, err
	}

	for _, ck := range resp.Cookies {
		t.CookieMap[ck.Name] = ck.Value
	}

	return resp, err
}

// NewEmailCli implements IEmail
func (t *ETempMailCli) NewEmailCli(opt map[string]any) (IEmail, error) {
	t = &ETempMailCli{
		CookieMap: map[string]string{},
		BaseUrl:   ETempMailBaseURL,
	}

	if opt != nil {
		if baseUrl, ok := opt["baseUrl"].(string); ok && baseUrl != "" {
			t.BaseUrl = baseUrl
		}
	}

	_, err := t.dohttp(joinURL(t.BaseUrl, "/zh"), "GET", nil, nil)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "主页", "Error", err.Error())
		return nil, err
	}
	out := map[string]string{}
	_, err = t.dohttp(joinURL(t.BaseUrl, "/getEmailAddress"), "POST", map[string]any{}, &out)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "getEmailAddress", "Error", err.Error())
		return nil, err
	}
	t.Email = out["address"]
	t.EmailId = out["id"]
	t.RecoverKey = out["recover_key"]

	ilog.Info("邮箱初始化成功", "客户端类型", t.CliName(),
		"邮箱", t.Email, "邮箱ID", t.EmailId,
		"恢复密钥", t.RecoverKey)
	return t, nil
}

// GetDomains implements IEmail. ETempMail uses a fixed domain.
func (t *ETempMailCli) GetDomains() ([]string, error) {
	return []string{"ohm.edu.pl", "cross.edu.pl", "usa.edu.pl", "beta.edu.pl"}, nil
}

// GetEmailMsgs implements IEmail
func (t *ETempMailCli) GetEmailMsgs() (msgs []Msg, err error) {
	type eTempMailMsg struct {
		Subject string `json:"subject"`
		From    string `json:"from"`
		Date    string `json:"date"`
		Body    string `json:"body"`
	}
	list := []eTempMailMsg{}
	_, err = t.dohttp(joinURL(t.BaseUrl, "/getInbox"), "POST", map[string]any{}, &list)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "getInbox", "Error", err.Error())
		return nil, err
	}
	msgs = []Msg{}
	for _, m := range list {
		msgs = append(msgs, Msg{
			From:    m.From,
			To:      t.Email,
			Subject: m.Subject,
			Date:    m.Date,
			Body:    m.Body,
		})
	}
	ilog.Info("邮箱获取邮件成功", "邮箱", t.Email, "邮件数量", len(msgs))
	return msgs, nil
}

// Data implements IEmail
func (t *ETempMailCli) Data() map[string]any {
	return map[string]any{
		"email":       t.Email,
		"id":          t.EmailId,
		"recover_key": t.RecoverKey,
	}
}
