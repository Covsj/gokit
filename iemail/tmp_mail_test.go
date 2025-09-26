package iemail

import (
	"testing"
	"time"

	"github.com/Covsj/gokit/ilog"
	"github.com/sirupsen/logrus"
)

// TestTmIntegration performs an integration test for the Mail.tm client.
// It creates a new random account and fetches the initial message list.
func TestTmIntegration(t *testing.T) {
	ilog.Log.SetLevel(logrus.TraceLevel)
	// Use random account generation for testing
	cli := &TmpCli{}
	iCLi, err := cli.NewEmailCli(nil)
	if err != nil {
		ilog.Error("TmpCli NewEmailCli失败", "错误", err)
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
