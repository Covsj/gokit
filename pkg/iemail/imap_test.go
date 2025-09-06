package iemail

import (
	"fmt"
	"testing"
)

func TestDragonsmailImapClient_Creation(t *testing.T) {
	// 测试创建IMAP客户端
	client := NewDragonsmailClient("ThanhWuTTk@gutsmail.com", "QmoPPOYOQc.")
	if client == nil {
		t.Error("NewImapClient() 返回 nil")
	}

	msgs, err := client.SearchMessages(&ImapSearchCriteria{From: "noreply@memoscan.org"})
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, msg := range msgs {
		fmt.Println(msg.From, msg.To, msg.Subject, msg.TextContent)
	}
}
