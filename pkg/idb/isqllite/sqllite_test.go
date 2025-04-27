package isqllite

import (
	"testing"

	"github.com/Covsj/gokit/pkg/ilog"
)

func TestSQLiteClient(t *testing.T) {
	// 创建SQLite客户端
	config := DefaultConfig()
	config.Path = "E:\\GOPATH\\src\\EagleEye-Core\\cmd\\eagleeye.db"
	client, err := New(config)
	if err != nil {
		ilog.ErrorF("创建SQLite客户端失败: %v", err)
		return
	}

	// 测试表是否存在
	exists, err := client.TableExists("sys_user")
	if err != nil {
		ilog.ErrorF("测试表是否存在失败: %v", err)
		return
	}
	ilog.InfoF("表是否存在: %v", exists)
	
}
