package iemail

import (
	"testing"
	"time"

	"github.com/Covsj/gokit/pkg/ilog"
)

// TestTmIntegration performs an integration test for the Mail.tm client.
// It creates a new random account and fetches the initial message list.
func TestTmIntegration(t *testing.T) {
	// Use random account generation for testing
	opt := &Options{Random: true}

	// Instantiate the TM client with the Base URL
	// Consider making the Base URL configurable for tests, e.g., via env vars
	cfg := TM{
		BaseUrl: "https://api.mail.gw", // Using the previous default
	}

	// Test account creation
	account, err := cfg.NewAccount(opt)
	if err != nil {
		t.Fatalf("测试 NewAccount 失败: %v", err)
	}
	if account == nil || account.Id == "" || account.Address == "" || account.Token == "" {
		t.Fatalf("测试 NewAccount 返回的 account 无效: %+v", account)
	}
	ilog.Info("测试: 随机账户创建成功", "邮箱", account.Address, "ID", account.Id) // Password no longer available in Account struct by default

	// Allow some time for potential welcome message or propagation
	time.Sleep(2 * time.Second)

	// Test fetching message list (page 1)
	list, err := cfg.MessagesList(account, 1)
	if err != nil {
		t.Fatalf("测试 MessagesList 失败: %v", err)
	}
	// We don't know how many messages to expect initially (could be 0 or 1 welcome message),
	// so we just check for errors. A successful call returns a non-nil slice.
	if list == nil {
		t.Fatalf("测试 MessagesList 返回了 nil 列表，预期为 non-nil (可能为空)")
	}
	ilog.Info("测试: 邮件列表获取成功 (第一页)", "获取到的邮件数量", len(list))

	// Optional: Add more assertions here, e.g., fetch a specific message if one exists.
}
