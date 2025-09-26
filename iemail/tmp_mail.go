package iemail

import (
	"errors"
	"strings"
	"time"

	"github.com/Covsj/gokit/ihttp"
	"github.com/Covsj/gokit/ilog"
	"github.com/Covsj/gokit/iutil"
	"github.com/Covsj/requests/models"
)

type TmpCli struct {
	BaseUrl   string
	Domains   []string
	CookieMap map[string]string

	Email    string
	Password string

	EmailId    string
	EmailToken string
}

func (t *TmpCli) CliName() string {
	return "TM系列"
}

func (t *TmpCli) dohttp(reqUrl, method string, rawBody map[string]any, out any) (*models.Response, error) {
	headers := map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
	}
	if t.EmailToken != "" {
		headers["Authorization"] = "Bearer " + t.EmailToken
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

func (t *TmpCli) GetDomains() ([]string, error) {
	if len(t.Domains) > 0 {
		return t.Domains, nil
	}

	type domainResp struct {
		List []struct {
			Id        string    `json:"id"`
			Domain    string    `json:"domain"`
			IsActive  bool      `json:"isActive"`
			IsPrivate bool      `json:"isPrivate"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
		} `json:"hydra:member"`
	}

	tmp, res := domainResp{}, []string{}
	_, err := t.dohttp(joinURL(t.BaseUrl, "/domains"), "GET", nil, &tmp)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "domains", "Error", err.Error())
		return nil, err
	}
	for _, item := range tmp.List {
		if item.IsActive {
			res = append(res, item.Domain)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("可用列表为空")
	}

	t.Domains = res
	ilog.Info("邮箱获取邮箱列表成功", "客户端类型", t.CliName(), "可用列表", res)
	return res, nil
}

func (t *TmpCli) messageById(id string) (Msg, error) {
	type tmpMessage struct {
		ID   string `json:"id"`
		From struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"from"`
		To []struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"to"`
		Subject string `json:"subject"`
		Intro   string `json:"intro"`
	}

	var tmpMsg tmpMessage
	messageURL := joinURL(t.BaseUrl, "/messages/"+id)
	_, err := ihttp.Do(&ihttp.Opt{
		URL:     messageURL,
		Method:  "GET",
		RespOut: &tmpMsg,
		Headers: map[string]string{
			"Authorization": "Bearer " + t.EmailToken,
		},
	})
	if err != nil {
		ilog.Error("TM/GW 邮箱 获取邮件内容失败", "Error", err.Error(), "id", id)
		return Msg{}, err
	}

	msg := Msg{
		From:    tmpMsg.From.Address,
		Subject: tmpMsg.Subject,
		Body:    tmpMsg.Intro,
	}
	if len(tmpMsg.To) != 0 {
		msg.To = tmpMsg.To[0].Address
	}

	return msg, nil
}

func (t *TmpCli) accounts() error {
	out := map[string]any{}

	_, err := t.dohttp(joinURL(t.BaseUrl, "/accounts"), "POST", map[string]any{
		"address":  t.Email,
		"password": t.Password,
	}, &out)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "accounts", "Error", err.Error())
		return err
	}

	t.Email = out["address"].(string)
	t.EmailId = out["id"].(string)

	// ilog.Info("邮箱内部逻辑成功", "客户端类型", t.CliName(),
	// 	"逻辑接口", "accounts", "响应", resp.Text)
	return nil
}

func (t *TmpCli) token() error {
	out := map[string]string{}
	_, err := t.dohttp(joinURL(t.BaseUrl, "/token"), "POST", map[string]any{
		"address":  t.Email,
		"password": t.Password,
	}, &out)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "token", "Error", err.Error())
		return err
	}
	// ilog.Info("邮箱内部逻辑成功", "客户端类型", t.CliName(),
	// 	"逻辑接口", "accounts", "响应", resp.Text)
	id, okId := out["id"]
	token, okToken := out["token"]
	if !okId || !okToken {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "token", "Error", "获取id或token失败")
		return errors.New("获取id或token失败")
	}

	t.EmailId = id
	t.EmailToken = token

	return nil
}

// NewEmailCli 实现 IEmail 接口的 NewEmailCli 方法
func (t *TmpCli) NewEmailCli(opt map[string]any) (IEmail, error) {
	// 创建新的 TmpCli 实例
	t = &TmpCli{}
	if opt != nil {
		if baseUrl, ok := opt["baseUrl"].(string); ok {
			t.BaseUrl = baseUrl
		}
	}

	domains, err := t.GetDomains()
	if err != nil {
		return nil, err
	}
	activeDomain := domains[iutil.GlobalRandom.Intn(len(domains))]

	emailName := strings.ToLower(iutil.GenerateRandomStr(8, ""))
	emailPwd := strings.ToLower(iutil.GenerateRandomStr(8, ""))

	email := emailName + "@" + activeDomain

	t.Email = email
	t.Password = emailPwd

	err = t.accounts()
	if err != nil {
		return nil, err
	}
	err = t.token()
	if err != nil {
		return nil, err
	}
	ilog.Info("邮箱初始化成功", "客户端类型", t.CliName(),
		"邮箱", email, "密码", emailPwd)
	return t, nil
}

// GetEmailMsgs 实现 IEmail 接口的 GetEmailMsgs 方法
func (t *TmpCli) GetEmailMsgs() (msgs []Msg, err error) {
	type addressee struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	type tmpMessage struct {
		ID      string      `json:"id"`
		From    addressee   `json:"from"`
		To      []addressee `json:"to"`
		Subject string      `json:"subject"` // 主题
		Intro   string      `json:"intro"`   // 内容
		//Seen      bool        `json:"seen"`
		//IsDeleted bool        `json:"isDeleted"`
		//Size      int         `json:"size"`
		//Text      string      `json:"text"`
		//HTML      []string    `json:"html"`
		//Attachments []Attachment `json:"attachments"`
	}

	type messageResp struct {
		List []tmpMessage `json:"hydra:member"`
	}

	var out messageResp

	_, err = t.dohttp(joinURL(t.BaseUrl, "/messages"),
		"GET", map[string]any{}, &out)
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "messages", "Error", err.Error())
		return nil, err
	}
	msgs = []Msg{}
	for _, item := range out.List {
		msg := Msg{
			From:    item.From.Address,
			Subject: item.Subject,
			Body:    item.Intro,
		}
		if len(item.To) != 0 {
			msg.To = item.To[0].Address
		}
		msgs = append(msgs, msg)
	}
	ilog.Info("邮箱获取邮件成功", "客户端类型", t.CliName(),
		"邮箱", t.Email, "邮件数量", len(msgs))
	return msgs, nil
}

func (t *TmpCli) Data() map[string]any {
	return map[string]any{
		"email":    t.Email,
		"password": t.Password,
	}
}
