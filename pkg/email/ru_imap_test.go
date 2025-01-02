package email

import (
	"testing"

	"github.com/Covsj/gokit/pkg/log"
)

func TestRuImap(t *testing.T) {
	ruClient := NewRuMailClient()

	// 连接到服务器
	err := ruClient.Connect("default@covsj.top", "a4FK9tXYNU0DGtNHHnJ8")
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer ruClient.Close()

	// 获取最新的10封邮件
	emails, err := ruClient.FetchLatestEmails(10)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for i, v := range emails {
		log.Info(i, "时间戳", v.TimeStamp,
			"发件人", v.From, "收件人", v.To,
			"主题", v.Subject)
	}
}
