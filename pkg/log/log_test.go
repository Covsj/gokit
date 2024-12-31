package log

import (
	"errors"
	"testing"
)

func TestLogger(t *testing.T) {
	// 测试 JSON 格式（默认）
	Debug("这是一条调试日志")
	Info("这是一条信息日志")
	Warn("这是一条警告日志")
	Error("这是一条错误日志")

	// 测试文本格式
	Debug("这是一条文本格式的调试日志")
	Info("这是一条文本格式的信息日志")
	Warn("这是一条文本格式的警告日志")
	Error("这是一条文本格式的错误日志")

	// 测试格式化日志
	DebugF("格式化的调试日志: %s", "debug")
	InfoF("格式化的信息日志: %d", 1)
	WarnF("格式化的警告日志: %v", true)
	ErrorF("格式化的错误日志: %v", errors.New("test error"))
}
