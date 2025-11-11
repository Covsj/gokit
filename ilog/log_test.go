package ilog

import (
	"errors"
	"testing"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestLogger(t *testing.T) {
	SetCallerConfig(3, true)
	// 测试带字段的日志✅
	Info("用户登录", "用户ID", "12345", "IP", "192.168.1.1",
		"struct", User{Name: "John", Age: 30},
		"map", map[string]interface{}{"name": "John", "age": 30},
		"slice", []string{"a", "b", "c"},
		"array", [3]string{"a", "b", "c"},
		"int", 123,
		"float", 123.456,
		"bool", true,
		"nil", nil,
	)

	// 测试多个字段
	Warn("系统警告",
		"模块", "认证服务",
		"错误码", "AUTH001",
		"详情", "密码尝试次数过多",
	)

	// 测试格式化日志
	InfoF("用户 %s 执行了 %s 操作", "admin", "删除文件")

	Success("系统警告",
		"模块", "认证服务",
		"错误码", "AUTH001",
		"详情", "密码尝试次数过多",
	)

	// 测试错误日志
	Failed("操作失败",
		errors.New("数据库错误"),
		"表名", "users",
		"SQL", "SELECT * FROM users",
	)

	Debug("测试DEBUG", "1", 2)
}

func TestFieldOrder(t *testing.T) {
	// 专门测试字段顺序
	Info("字段顺序测试",
		"第一个", "1",
		"第二个", "2",
		"第三个", "3",
		"第四个", "4",
		"第五个", "5",
	)

	// 测试字母顺序与输入顺序的区别
	Info("字母顺序测试",
		"zebra", "最后",
		"apple", "第一",
		"banana", "第二",
		"cherry", "第三",
	)
}
