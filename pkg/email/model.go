package email

// PrefixRequest 描述设置电子邮件前缀的请求格式。
type PrefixRequest struct {
	EmPrefix string `json:"em_prefix"`
}

// SetResponse 描述从WxEmail服务器返回的设置电子邮件响应。
type SetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DomainName string `json:"domain_name"`
		EmPrefix   string `json:"em_prefix"`
		Fulldomain string `json:"fulldomain"`
	} `json:"data"`
}

// FetchResponse 描述从WxEmail服务器返回的获取邮件消息的响应。
type FetchResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []Detail `json:"data"`
}

// Detail 描述一封电子邮件的基本内容。
type Detail struct {
	Id              int    `json:"id"`
	From            string `json:"from"`
	FromEmail       string `json:"fromemail"`
	To              string `json:"to"`
	Subject         string `json:"subject"`
	ToFormat        string `json:"toFormat"`
	MailContentType string `json:"mail_content_type"`
	MailContent     string `json:"mail_content"`
	CreateTime      string `json:"create_time"`
	HostUrl         string `json:"hosturl"`
	UpdateTime      string `json:"update_time"`
	Uid             int    `json:"uid"`
	Paid            int    `json:"paid"`
}

// ImapData 描述从imap服务器返回的具体邮件消息数据。
type ImapData struct {
	Id              int    `json:"id"`
	From            string `json:"from"`
	FromEmail       string `json:"fromemail"`
	To              string `json:"to"`
	Subject         string `json:"subject"`
	ToFormat        string `json:"toFormat"`
	MailContentType string `json:"mail_content_type"`
	MailContent     string `json:"mail_content"`
	CreateTime      string `json:"create_time"`
	HostURL         string `json:"hosturl"`
	UpdateTime      string `json:"update_time"`
	Uid             int    `json:"uid"`
	Paid            int    `json:"paid"`
}

// IMAPDetail 描述从IMAP服务器获取的电子邮件数据。
type IMAPDetail struct {
	TimeStamp int64  `json:"time_stamp"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}
