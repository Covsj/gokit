package isms

import (
	"fmt"
	"time"

	"github.com/Covsj/gokit/ihttp"
	log "github.com/Covsj/gokit/ilog"
)

type HaoZhuMa struct {
	Token string
	User  string
}

const (
	domain = "https://api.haozhuma.com"
)

func (h *HaoZhuMa) Login(user, password string) {
	type Res struct {
		Msg   string `json:"msg"`
		Code  int    `json:"code"`
		Token string `json:"token"`
	}
	reqUrl := fmt.Sprintf("%s/sms/?api=login&user=%s&pass=%s", domain, user, password)
	res := Res{}

	_, err := ihttp.Do(&ihttp.Opt{URL: reqUrl, RespOut: &res})
	if err != nil {
		log.Error("豪猪码登陆失败", "user", user, "password", password,
			"error", err.Error())
		return
	}

	log.Info("豪猪码登陆成功", "user", user, "password", password, "res", res)
	h.Token = res.Token
	h.User = user
}

func (h *HaoZhuMa) GetPhone(sid string) string {
	type Res struct {
		Code        string      `json:"code"`
		Msg         string      `json:"msg"`
		Sid         string      `json:"sid"`
		ShopName    string      `json:"shop_name"`
		CountryName string      `json:"country_name"`
		CountryCode string      `json:"country_code"`
		CountryQu   string      `json:"country_qu"`
		Uid         interface{} `json:"uid"`
		Phone       string      `json:"phone"`
		Sp          string      `json:"sp"`
		PhoneGsd    string      `json:"phone_gsd"`
	}

	reqUrl := fmt.Sprintf("%s/sms/?api=getPhone&token=%s"+
		"&sid=%s&ascription=2&isp=&isp=&Province=&sp=2&paragraph=&isp=1", domain, h.Token, sid)
	for i := 0; i < 5; i++ {
		res := Res{}
		_, err := ihttp.Do(&ihttp.Opt{URL: reqUrl, RespOut: &res})
		if err != nil {
			log.Error("获取号码失败", "sid", sid,
				"error", err.Error())
			continue
		}
		mobile := res.Phone
		if mobile != "" {
			log.Info("获取手机号成功", "mobile", mobile)
			return mobile
		}
	}
	return ""
}

func (h *HaoZhuMa) GetMessage(sid, phone string) (string, string) {
	type Res struct {
		Code interface{} `json:"code,omitempty"`
		Msg  string      `json:"msg,omitempty"`
		Sms  string      `json:"sms,omitempty"`
		Yzm  string      `json:"yzm,omitempty"`
	}
	reqUrl := fmt.Sprintf("%s/sms/?api=getMessage&token=%s&sid=%s&phone=%s", domain, h.Token, sid, phone)
	for i := 0; i < 20; i++ {
		res := Res{}
		_, err := ihttp.Do(&ihttp.Opt{URL: reqUrl, RespOut: &res})
		if err != nil {
			log.Error("获取验证码失败", "sid", sid,
				"error", err.Error())
			continue
		}
		log.Info("正在获取验证码", "phone", phone, "res", res)
		time.Sleep(5 * time.Second)
		if res.Yzm != "" {
			log.Info("获取验证码成功", "sid", sid, "mobile", phone, "code", res.Yzm)
			return res.Yzm, res.Sms
		}
	}
	return "", ""
}
