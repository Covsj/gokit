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
	err = cache.SetWithTTL("key2", "value2", 100*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to set cache with TTL: %v", err)
	}

	ok = cache.Get("key2", &val)
	if !ok || val != "value2" {
		t.Error("Failed to get cached value with TTL")
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)
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
