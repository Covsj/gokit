// 使用imap协议获取邮件

package iemail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/Covsj/gokit/ilog"
	"github.com/Covsj/requests/models"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

const (
	ReadBatchSize = 10 // 批量读取邮件数量
)

// ImapCli 实现 ImapService 接口
type ImapCli struct {
	client   *client.Client
	Addr     string // IMAP服务器地址，如 "imap.qq.com:993"
	Username string // 邮箱地址
	Password string // IMAP密码（不是邮箱登录密码）
	Folder   string // 邮箱文件夹，默认 "INBOX"
}

type imapMsg struct {
	UID         uint32     `json:"uid"`
	SeqNum      uint32     `json:"seqNum"`
	From        string     `json:"from"`
	To          []string   `json:"to"`
	Subject     string     `json:"subject"`
	Date        time.Time  `json:"date"`
	TextContent string     `json:"textContent"`
	HTMLContent string     `json:"htmlContent"`
	Attachments []imapAtts `json:"attachments"`
	Flags       []string   `json:"flags"`
	Size        uint32     `json:"size"`
}

type imapAtts struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	Data        []byte `json:"data,omitempty"`
}

// Connect 连接到IMAP服务器
func (t *ImapCli) connect() error {
	client, err := client.DialTLS(t.Addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	t.client = client
	// 登录
	if err := client.Login(t.Username, t.Password); err != nil {
		return fmt.Errorf("IMAP登录失败: %w", err)
	}
	ilog.Info("邮箱连接服务器成功", "客户端类型", t.CliName(),
		"邮箱", t.Addr)
	return nil
}

// Disconnect 断开IMAP连接
func (t *ImapCli) Disconnect() error {
	if t.client != nil {
		if err := t.client.Logout(); err != nil {
			ilog.Error("IMAP登出失败", "Error", err.Error())
			return err
		}
		ilog.Info("IMAP连接已断开")
	}
	return nil
}

// getMessageByUID 根据UID获取单个邮件
func (t *ImapCli) getMessageByUID(uid uint32) (*imapMsg, error) {

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchFlags,
		imap.FetchUid,
		imap.FetchRFC822Size,
		imap.FetchRFC822}
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)

	go func() {
		done <- t.client.UidFetch(seqset, items, messages)
	}()

	var result *imapMsg
	if msg, ok := <-messages; ok {
		ImapMsg, err := t.parseMessageWithContent(msg)
		if err != nil {
			return nil, fmt.Errorf("解析邮件内容失败: %w", err)
		}
		result = &ImapMsg
	}

	// Wait for fetch to complete
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件失败: %w", err)
	}

	if result != nil {
		return result, nil
	}

	return nil, errors.New("未找到指定UID的邮件")
}

// SearchMessages 搜索邮件
func (t *ImapCli) searchMessages() ([]uint32, error) {
	if t.client == nil {
		return nil, errors.New("IMAP客户端未连接")
	}

	// 选择邮箱文件夹
	folder := t.Folder

	_, err := t.client.Select(folder, false)
	if err != nil {
		return nil, fmt.Errorf("选择邮箱文件夹失败: %w", err)
	}

	searchCriteria := imap.NewSearchCriteria()

	// if criteria.From != "" {
	// 	searchCriteria.Header.Add("From", criteria.From)
	// }
	// if criteria.To != "" {
	// 	searchCriteria.Header.Add("To", criteria.To)
	// }
	// if criteria.Subject != "" {
	// 	searchCriteria.Header.Add("Subject", criteria.Subject)
	// }
	// if criteria.Body != "" {
	// 	searchCriteria.Text = []string{criteria.Body}
	// }
	// if !criteria.Since.IsZero() {
	// 	searchCriteria.Since = criteria.Since
	// }
	// if !criteria.Before.IsZero() {
	// 	searchCriteria.Before = criteria.Before
	// }
	// if criteria.Unseen {
	// 	searchCriteria.WithoutFlags = []string{imap.SeenFlag}
	// }
	// if criteria.Seen {
	// 	searchCriteria.WithFlags = []string{imap.SeenFlag}
	// }

	uids, err := t.client.UidSearch(searchCriteria)
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败: %w", err)
	}
	return uids, nil
}

// parseMessageWithContent 解析邮件完整内容
func (t *ImapCli) parseMessageWithContent(msg *imap.Message) (imapMsg, error) {
	ImapMsg := imapMsg{
		UID:    msg.Uid,
		SeqNum: msg.SeqNum,
		Size:   msg.Size,
	}

	if msg.Envelope != nil {
		ImapMsg.From = msg.Envelope.From[0].Address()
		ImapMsg.Subject = msg.Envelope.Subject
		ImapMsg.Date = msg.Envelope.Date

		for _, to := range msg.Envelope.To {
			ImapMsg.To = append(ImapMsg.To, to.Address())
		}
	}

	if msg.Flags != nil {
		ImapMsg.Flags = msg.Flags
	}

	// 获取邮件正文
	r := msg.GetBody(&imap.BodySectionName{})
	if r == nil {
		return ImapMsg, errors.New("服务器没有返回消息内容")
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		return ImapMsg, fmt.Errorf("创建邮件读取器失败: %w", err)
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			ilog.Warn("邮箱读取邮件失败", "客户端类型", t.CliName(),
				"邮箱", t.Addr, "错误", err)
			continue
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			contentType := h.Get("Content-Type")
			b, _ := ioutil.ReadAll(p.Body)
			if strings.HasPrefix(contentType, "text/plain") {
				ImapMsg.TextContent = string(b)
			} else if strings.HasPrefix(contentType, "text/html") {
				ImapMsg.HTMLContent = string(b)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			contentType := h.Get("Content-Type")
			b, _ := ioutil.ReadAll(p.Body)

			attachment := imapAtts{
				Filename:    filename,
				ContentType: contentType,
				Size:        len(b),
				Data:        b,
			}
			ImapMsg.Attachments = append(ImapMsg.Attachments, attachment)
		}
	}

	return ImapMsg, nil
}

// CliName 与 IEmail 系列保持一致的客户端标识
func (t *ImapCli) CliName() string {
	return "IMAP"
}

func (t *ImapCli) GetDomains() ([]string, error) {
	return []string{}, nil
}

// Data 返回关键信息，便于统一展示
func (t *ImapCli) Data() map[string]any {
	return map[string]any{
		"addr":     t.Addr,
		"username": t.Username,
		"folder":   t.Folder,
	}
}
func (t *ImapCli) dohttp(reqUrl, method string, rawBody map[string]any, out any) (*models.Response, error) {
	return nil, nil
}

// NewImapClient 通用构造器：根据传入的 ImapOpt 建立连接
func (t *ImapCli) NewEmailCli(opt map[string]any) (IEmail, error) {
	t = &ImapCli{}
	if addr, ok := opt["addr"].(string); ok {
		t.Addr = addr
	} else {
		return nil, errors.New("连接服务器为空")
	}
	if username, ok := opt["username"].(string); ok {
		t.Username = username
	} else {
		return nil, errors.New("用户名为空")
	}
	if password, ok := opt["password"].(string); ok {
		t.Password = password
	}
	if folder, ok := opt["folder"].(string); ok {
		t.Folder = folder
	} else {
		t.Folder = "INBOX"
	}
	err := t.connect()
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "connect", "Error", err.Error())
		return nil, err
	}

	ilog.Info("IMAP连接成功", "地址", t.Addr, "用户名", t.Username)

	return t, nil
}

// GetEmailMsgs 拉取最近若干封邮件并转换为通用 Msg
func (t *ImapCli) GetEmailMsgs() (msgs []Msg, err error) {
	uids, err := t.searchMessages()
	if err != nil {
		ilog.Error("邮箱内部逻辑失败", "客户端类型", t.CliName(),
			"逻辑接口", "searchMessages", "Error", err.Error())
		return nil, err
	}

	ilog.Info("邮箱查询邮件ID列表成功", "客户端类型", t.CliName(),
		"邮箱", t.Addr, "ID列表", uids)

	if len(uids) == 0 {
		return []Msg{}, nil
	}

	start := 0
	if len(uids) > ReadBatchSize {
		start = len(uids) - ReadBatchSize
	}

	msgs = []Msg{}
	for _, uid := range uids[start:] {
		im, err := t.getMessageByUID(uid)
		if err != nil {
			ilog.Warn("邮箱获取邮件详情失败", "客户端类型", t.CliName(),
				"邮箱", t.Addr, "邮件ID", uid, "错误", err)
			continue
		}
		// 优先文本正文
		body := im.TextContent
		if body == "" {
			body = im.HTMLContent
		}
		// To 列表串联
		toJoined := ""
		if len(im.To) > 0 {
			toJoined = im.To[0]
			for i := 1; i < len(im.To); i++ {
				toJoined += "," + im.To[i]
			}
		}
		msgs = append(msgs, Msg{
			From:    im.From,
			To:      toJoined,
			Subject: im.Subject,
			Date:    im.Date.Format(time.RFC3339),
			Body:    body,
		})
	}

	ilog.Info("邮箱获取邮件成功", "客户端类型", t.CliName(),
		"邮箱", t.Addr, "邮件数量", len(msgs))
	return msgs, nil
}
