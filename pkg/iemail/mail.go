package iemail

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Covsj/gokit/pkg/ihttp"
	"github.com/Covsj/gokit/pkg/ilog"
)

const (
	domainsPath  = "/domains"
	messagesPath = "/messages"
	accountsPath = "/accounts"
	tokenPath    = "/token"
)

const (
	MAIL_TM_API = "https://api.mail.tm"
	MAIL_GW_API = "https://api.mail.gw"
)

type TM struct {
	BaseUrl string
	Domains []Domain
}

func joinURL(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}

func (t *TM) AvailableDomains() ([]Domain, error) {
	tmp, res := domainResp{}, []Domain{}
	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:         joinURL(t.BaseUrl, domainsPath),
		Method:      "GET",
		ResponseOut: &tmp,
	})
	if err != nil {
		ilog.Error("email.tm 获取可用域名失败", "Error", err.Error())
		return nil, err
	}
	for _, item := range tmp.List {
		if item.IsActive {
			res = append(res, item)
		}
	}
	ilog.Info("email.tm 获取邮箱列表成功", "总", len(tmp.List), "可用", len(res))
	return res, nil
}

func (t *TM) NewAccount(opt *Options) (*Account, error) {
	var address, password string
	if opt.Random {
		var domain, username string
		domains, err := t.AvailableDomains()
		if err != nil {
			return nil, err
		}
		foundActive := false
		for _, item := range domains {
			if item.IsActive {
				domain = item.Domain
				foundActive = true
				break
			}
		}
		if !foundActive {
			return nil, errors.New("暂无可用Domain")
		}
		username, password = generateRandom(8)
		address = username + "@" + domain
	} else {
		if opt.Username == "" || opt.Domain == "" || opt.Password == "" {
			return nil, errors.New("非随机模式下，用户名、域名和密码不能为空")
		}
		address = opt.Username + "@" + opt.Domain
		password = opt.Password
	}

	err := t.regist(address, password)
	if err != nil {
		ilog.Error("email.tm 注册邮箱失败", "Error", err.Error(),
			"邮箱", address)
		return nil, err
	}
	account, err := t.login(address, password)
	if err != nil {
		ilog.Error("email.tm 邮箱登录失败", "Error", err.Error(),
			"邮箱", address)
		return nil, err
	}
	return account, nil
}

func (t *TM) MessagesList(account *Account, page int) ([]Message, error) {
	var data messageResp
	messagesURL := joinURL(t.BaseUrl, messagesPath) + "?page=" + strconv.Itoa(page)
	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:         messagesURL,
		Method:      "GET",
		ResponseOut: &data,
		Headers: map[string]string{
			"Authorization": "Bearer " + account.Token,
		},
	})
	if err != nil {
		ilog.Error("email.tm 邮件列表获取失败", "Error", err.Error())
		return nil, err
	}
	res := []Message{}
	for _, item := range data.List {
		msg, err := t.messageById(account, item.ID)
		if err != nil {
			ilog.Warn("email.tm 获取单个邮件失败，跳过", "Error", err.Error(), "msgId", item.ID)
			continue
		}
		res = append(res, msg)
	}
	return res, nil
}

func (t *TM) messageById(account *Account, id string) (Message, error) {
	var msg Message
	messageURL := joinURL(t.BaseUrl, messagesPath+"/"+id)
	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:         messageURL,
		Method:      "GET",
		ResponseOut: &msg,
		Headers: map[string]string{
			"Authorization": "Bearer " + account.Token,
		},
	})
	if err != nil {
		ilog.Error("email.tm 获取邮件内容失败", "Error", err.Error(), "id", id)
		return msg, err
	}
	return msg, nil
}

func (t *TM) regist(address string, password string) error {
	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:    joinURL(t.BaseUrl, accountsPath),
		Method: "POST",
		JSONBody: map[string]string{
			"address":  address,
			"password": password,
		},
	})
	if err != nil {
		return err
	}
	ilog.Info("email.tm 注册成功", "邮箱", address)
	return nil
}

func (t *TM) login(address string, password string) (*Account, error) {
	id, token, err := t.getIdAndToken(address, password)
	if err != nil {
		return nil, err
	}
	account, err := t.loginWithIdAndToken(id, token)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (t *TM) getIdAndToken(address string, password string) (string, string, error) {
	res := map[string]string{}
	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:    joinURL(t.BaseUrl, tokenPath),
		Method: "POST",
		JSONBody: map[string]string{
			"address":  address,
			"password": password,
		},
		ResponseOut: &res,
	})
	if err != nil {
		ilog.Error("email.tm 邮箱获取Token失败", "Error", err.Error(), "邮箱", address)
		return "", "", err
	}
	id, okId := res["id"]
	token, okToken := res["token"]
	if !okId || !okToken {
		ilog.Error("email.tm 邮箱获取Token响应缺少id或token", "响应", res)
		return "", "", errors.New("获取id或token失败，响应格式不正确")
	}
	return id, token, nil
}

func (t *TM) loginWithIdAndToken(id string, token string) (*Account, error) {
	account := &Account{}
	accountURL := joinURL(t.BaseUrl, accountsPath+"/"+id)

	_, err := ihttp.DoRequest(&ihttp.Options{
		URL:    accountURL,
		Method: "GET",
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
		ResponseOut: account,
	})
	if err != nil {
		ilog.Error("email.tm 邮箱激活/获取账户信息失败", "Error", err.Error(),
			"ID", id)
		return nil, err
	}
	account.Token = token
	return account, nil
}
