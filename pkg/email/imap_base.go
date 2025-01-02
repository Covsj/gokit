package email

import (
	"fmt"
	"io"
	"io/ioutil"

	"time"

	"github.com/Covsj/gokit/pkg/log"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// IMAPProvider 定义IMAP服务提供者接口
type IMAPProvider interface {
	GetServer() string
	GetTimeout() time.Duration
}

// BaseIMAPClient 基础IMAP客户端结构体
type BaseIMAPClient struct {
	client   *client.Client
	provider IMAPProvider
}

// NewBaseIMAPClient 创建新的基础IMAP客户端
func NewBaseIMAPClient(provider IMAPProvider) *BaseIMAPClient {
	return &BaseIMAPClient{
		provider: provider,
	}
}

// Connect 连接到IMAP服务器
func (b *BaseIMAPClient) Connect(username, password string) error {
	server := b.provider.GetServer()
	var err error
	var c *client.Client
	for i := 0; i <= 5; i++ {
		c, err = client.DialTLS(server, nil)
		if err != nil {
			log.ErrorF("Unable to establish TLS connection,server:%s ,error:%v", server, err)
			continue
		}

		if err = c.Login(username, password); err != nil {
			log.ErrorF("IMAP login failed: %v", err)
			continue
		}
	}
	if err != nil {
		return err
	}
	b.client = c
	return nil
}

// FetchEmails 获取邮件的基础方法
func (b *BaseIMAPClient) FetchEmails(numEmails int) ([]IMAPDetail, error) {
	var emails []IMAPDetail

	mbox, err := b.client.Select("INBOX", false)
	if err != nil {
		return emails, fmt.Errorf("failed to select INBOX: %v", err)
	}

	if mbox.Messages == 0 {
		return emails, nil
	}

	start := uint32(1)
	if mbox.Messages > uint32(numEmails) {
		start = mbox.Messages - uint32(numEmails) + 1
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(start, mbox.Messages)

	messages := make(chan *imap.Message, numEmails)
	done := make(chan error, 1)
	go func() {
		done <- b.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()

	for msg := range messages {
		detail := IMAPDetail{
			TimeStamp: msg.Envelope.Date.Unix(),
			From:      msg.Envelope.From[0].Address(),
			To:        msg.Envelope.To[0].Address(),
			Subject:   msg.Envelope.Subject,
		}
		if body := msg.GetBody(&imap.BodySectionName{Peek: true}); body != nil {
			bytes, _ := ioutil.ReadAll(body)
			detail.Body = string(bytes)
		}
		emails = append(emails, detail)
	}

	if err := <-done; err != nil {
		return emails, fmt.Errorf("failed to fetch emails: %v", err)
	}
	return emails, nil
}

// Close 关闭IMAP连接
func (b *BaseIMAPClient) Close() error {
	if b.client != nil {
		b.client.Logout()
		return b.client.Close()
	}
	return nil
}

// ParseEmailBodyAndAttachments 解析邮件内容和附件
func ParseEmailBodyAndAttachments(mr *mail.Reader) (string, map[string][]byte) {
	var body string
	attachments := make(map[string][]byte)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.ErrorF("Error reading next part: %s", err.Error())
			break
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			b, _ := ioutil.ReadAll(part.Body)
			body = string(b)
		case *mail.AttachmentHeader:
			filename, _ := header.Filename()
			fileContent, _ := ioutil.ReadAll(part.Body)
			attachments[filename] = fileContent
		}
	}

	return body, attachments
}
