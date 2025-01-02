package email

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Covsj/gokit/pkg/ihttp"
	"github.com/Covsj/gokit/pkg/log"
)

type Config struct {
	Name   string `json:"name"`
	Token  string `json:"token"`
	Host   string `json:"host"`
	Domain string `json:"domain"`
}

func NewConfig(name, token, host, domain string) *Config {
	return &Config{
		Name:   name,
		Token:  token,
		Host:   host,
		Domain: domain,
	}
}

func (cfg *Config) ConfigureEmail(isRand bool, emailName string) (string, error) {
	requestBody := ""
	if !isRand && emailName != "" {
		requestBody = strings.ReplaceAll(`{"em_prefix":"REPLACE_EMAIL"}`, "REPLACE_EMAIL", emailName)
	}

	resp, err := ihttp.DoRequest(
		&ihttp.RequestOptions{
			URL:     "https://" + cfg.Domain + "/api/mailbox/rand_emprefix",
			Method:  "POST",
			Body:    requestBody,
			Headers: map[string]string{"token": cfg.Token},
		},
	)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("request failed: %v", err)
	}
	log.Info("邮箱初始化成功","响应",resp)
	response := &SetResponse{}
	if err := json.Unmarshal(resp.Bytes(), response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Data.Fulldomain, nil
}

func (cfg *Config) FetchEmails(targetEmail, targetSubject string) ([]Detail, error) {
	var err error
	var emails []Detail

	response := &FetchResponse{}

	for i := 0; i < 5; i++ {
		resp, err := ihttp.DoRequest(
			&ihttp.RequestOptions{
				URL:     "https://" + cfg.Host + "/api/mailbox/getnewest5",
				Method:  "POST",
				Headers: map[string]string{"token": cfg.Token},
				ResponseOut: response,
			},
		)

		if err != nil || resp.StatusCode != 200 {
			continue
		}

		for _, item := range response.Data {
			if targetEmail != "" {
				if strings.Contains(item.To, targetEmail) {
					emails = append(emails, item)
				}
			} else if targetSubject != "" {
				if strings.Contains(item.Subject, targetSubject) {
					emails = append(emails, item)
				}
			}
		}
		if len(emails) > 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return emails, err
}
