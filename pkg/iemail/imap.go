// 使用imap协议获取邮件

package iemail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/Covsj/gokit/pkg/ilog"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

const (
	ReadBatchSize = 10 // 批量读取邮件数量
)

// 默认配置常量
const (
	DefaultQQImapAddr          = "imap.qq.com:993"
	DefaultDragonsmailImapAddr = "pop.dragonsmail.com:993"
	DefaultFirstmailImapAddr   = "imap.firstmail.ltd:993"
	DefaultGmailImapAddr       = "imap.gmail.com:993"
	DefaultOutlookImapAddr     = "outlook.office365.com:993"
)

// ImapClient 实现 ImapService 接口
type ImapClient struct {
	client *client.Client
	opt    *ImapOptions
}

// Connect 连接到IMAP服务器
func (c *ImapClient) Connect() error {
	opt := c.opt
	if opt == nil {
		return errors.New("IMAP选项不能为空")
	}
	if opt.Addr == "" || opt.Username == "" || opt.Password == "" {
		return errors.New("IMAP服务器地址、用户名和密码不能为空")
	}

	ilog.Info("开始连接IMAP服务器", "地址", opt.Addr, "用户名", opt.Username)

	// 建立与 IMAP 服务器的连接，跳过证书验证
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	client, err := client.DialTLS(opt.Addr, tlsConfig)
	if err != nil {
		ilog.Error("连接IMAP服务器失败", "Error", err.Error(), "地址", opt.Addr)
		return fmt.Errorf("连接IMAP服务器失败: %w", err)
	}

	c.client = client
	c.opt = opt

	// 登录
	if err := client.Login(opt.Username, opt.Password); err != nil {
		ilog.Error("IMAP登录失败", "Error", err.Error(), "用户名", opt.Username)
		return fmt.Errorf("IMAP登录失败: %w", err)
	}

	ilog.Info("IMAP连接成功", "地址", opt.Addr, "用户名", opt.Username)
	return nil
}

// Disconnect 断开IMAP连接
func (c *ImapClient) Disconnect() error {
	if c.client != nil {
		if err := c.client.Logout(); err != nil {
			ilog.Error("IMAP登出失败", "Error", err.Error())
			return err
		}
		ilog.Info("IMAP连接已断开")
	}
	return nil
}

// getMessageByUID 根据UID获取单个邮件
func (c *ImapClient) getMessageByUID(uid uint32) (*ImapMessage, error) {
	if c.client == nil {
		return nil, errors.New("IMAP客户端未连接")
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid, imap.FetchRFC822Size, imap.FetchRFC822}
	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)

	go func() {
		done <- c.client.UidFetch(seqset, items, messages)
	}()

	var result *ImapMessage
	for msg := range messages {
		imapMsg, err := c.parseMessageWithContent(msg)
		if err != nil {
			ilog.Error("解析邮件内容失败", "Error", err.Error(), "UID", uid)
			return nil, fmt.Errorf("解析邮件内容失败: %w", err)
		}
		result = &imapMsg
		// Only process first message since we're looking for a specific UID
		break
	}

	// Wait for fetch to complete
	if err := <-done; err != nil {
		ilog.Error("获取邮件失败", "Error", err.Error(), "UID", uid)
		return nil, fmt.Errorf("获取邮件失败: %w", err)
	}

	if result != nil {
		return result, nil
	}

	return nil, errors.New("未找到指定UID的邮件")
}

// SearchMessages 搜索邮件
func (c *ImapClient) searchMessages(criteria *ImapSearchCriteria) ([]uint32, error) {
	if c.client == nil {
		return nil, errors.New("IMAP客户端未连接")
	}

	if criteria == nil {
		return nil, errors.New("搜索条件不能为空")
	}

	// 选择邮箱文件夹
	folder := "INBOX"
	if c.opt != nil && c.opt.Folder != "" {
		folder = c.opt.Folder
	}

	_, err := c.client.Select(folder, false)
	if err != nil {
		ilog.Error("选择邮箱文件夹失败", "Error", err.Error(), "文件夹", folder)
		return nil, fmt.Errorf("选择邮箱文件夹失败: %w", err)
	}

	searchCriteria := imap.NewSearchCriteria()

	if criteria.From != "" {
		searchCriteria.Header.Add("From", criteria.From)
	}
	if criteria.To != "" {
		searchCriteria.Header.Add("To", criteria.To)
	}
	if criteria.Subject != "" {
		searchCriteria.Header.Add("Subject", criteria.Subject)
	}
	if criteria.Body != "" {
		searchCriteria.Text = []string{criteria.Body}
	}
	if !criteria.Since.IsZero() {
		searchCriteria.Since = criteria.Since
	}
	if !criteria.Before.IsZero() {
		searchCriteria.Before = criteria.Before
	}
	if criteria.Unseen {
		searchCriteria.WithoutFlags = []string{imap.SeenFlag}
	}
	if criteria.Seen {
		searchCriteria.WithFlags = []string{imap.SeenFlag}
	}

	uids, err := c.client.UidSearch(searchCriteria)
	if err != nil {
		ilog.Error("搜索邮件失败", "Error", err.Error())
		return nil, fmt.Errorf("搜索邮件失败: %w", err)
	}

	ilog.Info("搜索邮件成功", "结果数量", len(uids))
	return uids, nil
}

// parseMessage 解析邮件基本信息
func (c *ImapClient) parseMessage(msg *imap.Message) (ImapMessage, error) {
	imapMsg := ImapMessage{
		UID:    msg.Uid,
		SeqNum: msg.SeqNum,
		Size:   msg.Size,
	}

	if msg.Envelope != nil {
		imapMsg.From = msg.Envelope.From[0].Address()
		imapMsg.Subject = msg.Envelope.Subject
		imapMsg.Date = msg.Envelope.Date

		for _, to := range msg.Envelope.To {
			imapMsg.To = append(imapMsg.To, to.Address())
		}
	}

	if msg.Flags != nil {
		imapMsg.Flags = msg.Flags
	}

	return imapMsg, nil
}

// parseMessageWithContent 解析邮件完整内容
func (c *ImapClient) parseMessageWithContent(msg *imap.Message) (ImapMessage, error) {
	imapMsg, err := c.parseMessage(msg)
	if err != nil {
		return imapMsg, err
	}

	// 获取邮件正文
	r := msg.GetBody(&imap.BodySectionName{})
	if r == nil {
		return imapMsg, errors.New("服务器没有返回消息内容")
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		return imapMsg, fmt.Errorf("创建邮件读取器失败: %w", err)
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			ilog.Warn("读取邮件内容时出现错误", "Error", err.Error())
			continue
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			contentType := h.Get("Content-Type")
			b, _ := ioutil.ReadAll(p.Body)
			if strings.HasPrefix(contentType, "text/plain") {
				imapMsg.TextContent = string(b)
			} else if strings.HasPrefix(contentType, "text/html") {
				imapMsg.HTMLContent = string(b)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			contentType := h.Get("Content-Type")
			b, _ := ioutil.ReadAll(p.Body)

			attachment := ImapAttachment{
				Filename:    filename,
				ContentType: contentType,
				Size:        len(b),
				Data:        b,
			}
			imapMsg.Attachments = append(imapMsg.Attachments, attachment)
		}
	}

	return imapMsg, nil
}

func NewDragonsmailClient(username, password string) *ImapClient {
	cli := &ImapClient{
		opt: &ImapOptions{
			Addr:     DefaultDragonsmailImapAddr,
			Username: username,
			Password: password,
			Folder:   "INBOX",
		},
	}
	cli.Connect()
	return cli
}

func NewFirstmailClient(username, password string) *ImapClient {
	cli := &ImapClient{
		opt: &ImapOptions{
			Addr:     DefaultFirstmailImapAddr,
			Username: username,
			Password: password,
			Folder:   "INBOX",
		},
	}
	cli.Connect()
	return cli
}

// ConnectAndSearchMessages 连接并搜索邮件的便捷方法
func (c *ImapClient) SearchMessages(criteria *ImapSearchCriteria) ([]ImapMessage, error) {

	uids, err := c.searchMessages(criteria)
	if err != nil {
		return nil, err
	}

	var messages []ImapMessage
	for _, uid := range uids {
		msg, err := c.getMessageByUID(uid)
		if err != nil {
			ilog.Warn("获取邮件失败，跳过", "Error", err.Error(), "UID", uid)
			continue
		}
		messages = append(messages, *msg)
	}

	return messages, nil
}
