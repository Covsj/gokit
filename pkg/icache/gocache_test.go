package icache

import (
	"testing"
	"time"
)

func TestGoCache(t *testing.T) {
	// 创建一个默认过期时间为5分钟，每10分钟清理一次的缓存
	cache := NewGoCache(5*time.Minute, 10*time.Minute)

	// 测试基本的设置和获取
	cache.Set("key1", "value1")
	if val, ok := cache.Get("key1"); !ok || val.(string) != "value1" {
		t.Error("Failed to get cached value")
	}

	// 测试带TTL的设置
	cache.SetWithTTL("key2", "value2", 100*time.Millisecond)
	if val, ok := cache.Get("key2"); !ok || val.(string) != "value2" {
		t.Error("Failed to get cached value with TTL")
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)
	if _, ok := cache.Get("key2"); ok {
		t.Error("Cache item should have expired")
	}

	// 测试删除
	cache.Set("key3", "value3")
	cache.Delete("key3")
	if _, ok := cache.Get("key3"); ok {
		t.Error("Cache item should have been deleted")
	}

	// 测试清空
	cache.Set("key4", "value4")
	cache.Clear()
	if cache.Len() != 0 {
		t.Error("Cache should be empty after clear")
	}
}
