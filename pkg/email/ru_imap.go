package email

import "time"

// RuMailProvider Mail.ru的IMAP提供者实现
type RuMailProvider struct{}

func (r *RuMailProvider) GetServer() string {
	return "imap.mail.ru:993"
}

func (r *RuMailProvider) GetTimeout() time.Duration {
	return 3 * time.Second
}

// RuMailClient Mail.ru邮件客户端
type RuMailClient struct {
	*BaseIMAPClient
}

// NewRuMailClient 创建新的Mail.ru邮件客户端
func NewRuMailClient() *RuMailClient {
	return &RuMailClient{
		BaseIMAPClient: NewBaseIMAPClient(&RuMailProvider{}),
	}
}

// FetchLatestEmails 获取最新的邮件
func (r *RuMailClient) FetchLatestEmails(numEmails int) ([]IMAPDetail, error) {
	return r.FetchEmails(numEmails)
}
