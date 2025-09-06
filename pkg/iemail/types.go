package iemail

import "time"

type Account struct {
	Id         string                 `json:"id"`
	Address    string                 `json:"address"`
	Password   string                 `json:"password"`
	Token      string                 `json:"token"`
	IsDeleted  bool                   `json:"isDeleted"`
	IsDisabled bool                   `json:"isDisabled"`
	Properties map[string]interface{} `json:"properties"`
}

type domainResp struct {
	List []Domain `json:"hydra:member"`
}
type Domain struct {
	Id        string    `json:"id"`
	Domain    string    `json:"domain"`
	IsActive  bool      `json:"isActive"`
	IsPrivate bool      `json:"isPrivate"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Options struct {
	Random   bool
	Domain   string
	Username string
	Password string
}
type messageResp struct {
	List []Message `json:"hydra:member"`
}
type Message struct {
	ID          string       `json:"id"`
	From        Addressee    `json:"from"`
	To          []Addressee  `json:"to"`
	Subject     string       `json:"subject"`
	Intro       string       `json:"intro"`
	Seen        bool         `json:"seen"`
	IsDeleted   bool         `json:"isDeleted"`
	Size        int          `json:"size"`
	Text        string       `json:"text"`
	HTML        []string     `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type Addressee struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Attachment struct {
	Id               string `json:"id"`
	Filename         string `json:"filename"`
	ContentType      string `json:"contentType"`
	Disposition      string `json:"disposition"`
	TransferEncoding string `json:"transferEncoding"`
	Related          bool   `json:"related"`
	Size             int    `json:"size"`
	DownloadURL      string `json:"downloadUrl"`
}

type Service interface {
	AvailableDomains() ([]Domain, error)
	NewAccount(*Options) (*Account, error)
	MessagesList(*Account, int) ([]Message, error)
}

// IMAP相关类型定义
type ImapOptions struct {
	Addr     string // IMAP服务器地址，如 "imap.qq.com:993"
	Username string // 邮箱地址
	Password string // IMAP密码（不是邮箱登录密码）
	Folder   string // 邮箱文件夹，默认 "INBOX"
}

type ImapMessage struct {
	UID         uint32           `json:"uid"`
	SeqNum      uint32           `json:"seqNum"`
	From        string           `json:"from"`
	To          []string         `json:"to"`
	Subject     string           `json:"subject"`
	Date        time.Time        `json:"date"`
	TextContent string           `json:"textContent"`
	HTMLContent string           `json:"htmlContent"`
	Attachments []ImapAttachment `json:"attachments"`
	Flags       []string         `json:"flags"`
	Size        uint32           `json:"size"`
}

type ImapAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	Data        []byte `json:"data,omitempty"`
}

type ImapFolder struct {
	Name     string `json:"name"`
	Messages uint32 `json:"messages"`
	Unseen   uint32 `json:"unseen"`
}

type ImapSearchCriteria struct {
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
	Subject string    `json:"subject,omitempty"`
	Body    string    `json:"body,omitempty"`
	Since   time.Time `json:"since,omitempty"`
	Before  time.Time `json:"before,omitempty"`
	Unseen  bool      `json:"unseen,omitempty"`
	Seen    bool      `json:"seen,omitempty"`
}

// IMAP服务接口
type ImapService interface {
	Connect(opt *ImapOptions) error
	Disconnect() error
	GetFolders() ([]ImapFolder, error)
	GetMessages(folder string, page int, pageSize int) ([]ImapMessage, error)
	GetMessageByUID(uid uint32) (*ImapMessage, error)
	MarkAsRead(uids []uint32) error
	MarkAsUnread(uids []uint32) error
	DeleteMessages(uids []uint32) error
	SearchMessages(criteria *ImapSearchCriteria) ([]uint32, error)
}
