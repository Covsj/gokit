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
	SetColors(LogColorConfig{
		// 调用时间 - 晨雾灰
		Timestamp: func(text string) string {
			return RGBForeground(130, 140, 150, text)
		},
		// 日志级别
		LogLevel: func(level string) string {
			colors := map[string]struct{ r, g, b int }{
				"DBUG": {120, 170, 220}, // 晨空蓝
				"INFO": {100, 180, 140}, // 嫩芽绿
				"WARN": {220, 180, 100}, // 晨光黄
				"ERRO": {220, 130, 130}, // 朝霞红
			}
			if color, exists := colors[level]; exists {
				return RGBForeground(color.r, color.g, color.b, level)
			}
			return RGBForeground(150, 160, 170, level)
		},
		// 调用堆栈 - 薄暮紫
		StackTrace: func(text string) string {
			return RGBForeground(170, 150, 210, text)
		},
		// 日志消息key - 远山蓝
		MessageKey: func(text string) string {
			return RGBForeground(100, 150, 200, text)
		},
		// 日志其他字段key - 暖沙橙
		FieldKey: func(text string) string {
			return RGBForeground(200, 150, 100, text)
		},
		// 日志其他字段value - 湖水绿
		FieldValue: func(text string) string {
			return RGBForeground(90, 170, 160, text)
		},
	})
	Set5()
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
