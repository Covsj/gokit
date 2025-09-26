package iemail

import (
	"strings"

	"github.com/Covsj/requests/models"
)

const (
	TMP_MAIL_TM_API = "https://api.mail.tm"
	TMP_MAIL_GW_API = "https://api.mail.gw"
)

const (
	ETempMailBaseURL = "https://etempmail.com"

	FakeMailBaseURL = "https://www.fakemail.net"
)

// 默认配置常量
const (
	QQImapAddr          = "imap.qq.com:993"
	DragonsmailImapAddr = "pop.dragonsmail.com:993"
	FirstmailImapAddr   = "imap.firstmail.ltd:993"
	GmailImapAddr       = "imap.gmail.com:993"
	OutlookImapAddr     = "outlook.office365.com:993"
)

type Msg struct {
	From    string
	To      string
	Subject string
	Date    string
	Body    string
	Extra   map[string]any
}

type IEmail interface {
	CliName() string
	dohttp(reqUrl, method string, rawBody map[string]any, out any) (*models.Response, error)
	NewEmailCli(opt map[string]any) (IEmail, error)
	GetDomains() ([]string, error)
	GetEmailMsgs() ([]Msg, error)
	Data() map[string]any
	Disconnect() error
}

func joinURL(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}
