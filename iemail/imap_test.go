package iemail

import (
	"testing"
	"time"

	"github.com/Covsj/gokit/ilog"
	"github.com/sirupsen/logrus"
)

func TestDragonsmailImapClient_Creation(t *testing.T) {

	ilog.Log.SetLevel(logrus.TraceLevel)
	// Use random account generation for testing
	cli := &ImapCli{}
	iCLi, err := cli.NewEmailCli(map[string]any{
		"addr":     DragonsmailImapAddr,
		"username": "BonitaSantosf1L6rJ@dragonsmail.com",
		"password": "4lAU.LWVsai.",
		"folder":   "INBOX",
	})
	if err != nil {
		ilog.Error("ImapCli NewEmailCli失败", "错误", err)
		return
	}
	for {
		msgs, err := iCLi.GetEmailMsgs()
		if err != nil {
			ilog.Error("错误", err)
			return
		}
		for _, msg := range msgs {
			ilog.Info("获取成功", "消息", msg)
		}
		time.Sleep(3 * time.Second)
	}
}
