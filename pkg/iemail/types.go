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
