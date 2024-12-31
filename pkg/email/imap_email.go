package email

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// ConnectToIMAP 登录到IMAP服务器并返回客户端连接。
func ConnectToIMAP(username, password string) (*client.Client, error) {
	server := "imap.mail.ru:993"
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	c, err := client.DialWithDialerTLS(dialer, server, nil)
	if err != nil {
		log.Printf("Unable to establish TLS connection: %v", err)
		c, err = client.DialWithDialer(dialer, server) // Unencrypted login
		if err != nil {
			return nil, err
		}
	}

	if err = c.Login(username, password); err != nil {
		log.Printf("IMAP login failed: %v", err)
		return nil, err
	}

	return c, nil
}

// FetchLatestIMAPEmails 从IMAP服务器获取最新的电子邮件。
func FetchLatestIMAPEmails(c *client.Client, numEmails int) ([]IMAPDetail, error) {

	var emails []IMAPDetail
	mbox, err := c.Select("INBOX", false)
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
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
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

// ParseEmailBodyAndAttachments 解析电子邮件主体和附件。
func ParseEmailBodyAndAttachments(mr *mail.Reader) (string, map[string][]byte) {
	var body string
	attachments := make(map[string][]byte)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error reading next part:", err)
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
