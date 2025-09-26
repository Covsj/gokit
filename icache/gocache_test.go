package icache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestFreeCache(t *testing.T) {
	// 创建一个10MB大小的缓存
	cacheSize := 10 * 1024 * 1024 // 10MB
	cache := NewFreeCache(cacheSize)

	// 测试基本的设置和获取
	err := cache.Set("key1", "value1")
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	var val string
	ok := cache.Get("key1", &val)
	if !ok || val != "value1" {
		t.Error("Failed to get cached value")
	}

	// 测试带TTL的设置
	err = cache.SetWithTTL("key2", "value2", 1*time.Second)
	if err != nil {
		t.Errorf("Failed to set cache with TTL: %v", err)
	}

	ok = cache.Get("key2", &val)
	if !ok || val != "value2" {
		t.Error("Failed to get cached value with TTL")
	}

	// 等待过期
	time.Sleep(2 * time.Second)
	if ok := cache.Get("key2", nil); ok {
		t.Error("Cache item should have expired")
	}

	// 测试删除
	cache.Set("key3", "value3")
	cache.Delete("key3")
	if ok := cache.Get("key3", nil); ok {
		t.Error("Cache item should have been deleted")
	}

	// 测试清空
	cache.Set("key4", "value4")
	cache.Clear()
	if cache.Len() != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestFreeCacheConcurrency(t *testing.T) {
	// 创建一个10MB大小的缓存
	cacheSize := 10 * 1024 * 1024 // 10MB
	cache := NewFreeCache(cacheSize)

	var wg sync.WaitGroup
	routines := 100
	operations := 100

	// 并发写入
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := []byte("key-" + strconv.Itoa(routineID) + "-" + strconv.Itoa(j))
				value := []byte("value-" + strconv.Itoa(routineID) + "-" + strconv.Itoa(j))
				cache.Set(string(key), value)
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := "key-" + strconv.Itoa(routineID) + "-" + strconv.Itoa(j)
				cache.Get(key, nil)
			}
		}(i)
	}

	// 并发删除
	for i := 0; i < routines/2; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < operations/2; j++ {
				key := "key-" + strconv.Itoa(routineID) + "-" + strconv.Itoa(j)
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// 检查缓存状态
	stats := cache.GetStats()
	t.Logf("Cache stats after concurrent operations: %+v", stats)
}

func TestFreeCacheEdgeCases(t *testing.T) {
	cacheSize := 10 * 1024 * 1024 // 10MB
	cache := NewFreeCache(cacheSize)

	// 测试零TTL（永不过期）
	err := cache.SetWithTTL("zero_ttl", "value", 0)
	if err != nil {
		t.Errorf("Failed to set cache with zero TTL: %v", err)
	}

	// 等待一段时间后检查是否仍然存在
	time.Sleep(100 * time.Millisecond)
	var val string
	if !cache.Get("zero_ttl", &val) || val != "value" {
		t.Error("Zero TTL should not expire")
	}

	// 测试负TTL（永不过期）
	err = cache.SetWithTTL("negative_ttl", "value", -1*time.Second)
	if err != nil {
		t.Errorf("Failed to set cache with negative TTL: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !cache.Get("negative_ttl", &val) || val != "value" {
		t.Error("Negative TTL should not expire")
	}

	// 测试极小TTL（应该至少设置为1秒）
	err = cache.SetWithTTL("tiny_ttl", "value", 100*time.Microsecond)
	if err != nil {
		t.Errorf("Failed to set cache with tiny TTL: %v", err)
	}

	// 立即检查应该存在
	if !cache.Get("tiny_ttl", &val) || val != "value" {
		t.Error("Tiny TTL should be set to at least 1 second")
	}

	// 等待1.5秒后应该过期
	time.Sleep(1500 * time.Millisecond)
	if cache.Get("tiny_ttl", nil) {
		t.Error("Tiny TTL should have expired after 1.5 seconds")
	}
}
