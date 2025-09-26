package icache

import (
	"encoding/json"
	"time"

	"github.com/coocood/freecache"
)

// FreeCache 是对freecache的简单封装，支持存储结构体
type FreeCache struct {
	cache *freecache.Cache
}

// NewFreeCache 创建一个新的FreeCache实例
// size: 缓存大小，单位为字节
func NewFreeCache(size int) *FreeCache {
	return &FreeCache{
		cache: freecache.NewCache(size),
	}
}

// Set 将键值对存入缓存，永不过期
// value可以是任意类型，内部会进行JSON序列化
func (c *FreeCache) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.cache.Set([]byte(key), data, 0)
}

// SetWithTTL 将键值对存入缓存，并设置过期时间
// value可以是任意类型，内部会进行JSON序列化
// ttl: 过期时间，如果小于等于0则永不过期
func (c *FreeCache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// freecache要求过期时间以秒为单位
	// 如果TTL小于1秒但大于0，至少设置为1秒，避免0值（永不过期）
	var expireSeconds int
	if ttl <= 0 {
		expireSeconds = 0 // 永不过期
	} else {
		expireSeconds = int(ttl.Seconds())
		if expireSeconds <= 0 {
			expireSeconds = 1 // 至少1秒
		}
	}
	return c.cache.Set([]byte(key), data, expireSeconds)
}

// Get 获取键对应的值并解析到目标结构体中
// valuePtr应该是指向结构体的指针，例如：&myStruct
func (c *FreeCache) Get(key string, valuePtr interface{}) bool {
	data, err := c.cache.Get([]byte(key))
	if err != nil {
		return false
	}
	if valuePtr == nil {
		return true
	}
	err = json.Unmarshal(data, valuePtr)
	return err == nil
}

// GetRaw 获取键对应的原始字节数据
func (c *FreeCache) GetRaw(key string) ([]byte, bool) {
	value, err := c.cache.Get([]byte(key))
	if err != nil {
		return nil, false
	}
	return value, true
}

// Delete 删除键对应的值
func (c *FreeCache) Delete(key string) {
	c.cache.Del([]byte(key))
}

// Clear 清空缓存
func (c *FreeCache) Clear() {
	c.cache.Clear()
}

// Len 返回缓存中的条目数
func (c *FreeCache) Len() int {
	return int(c.cache.EntryCount())
}

// GetStats 获取缓存统计信息
func (c *FreeCache) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["EntryCount"] = c.cache.EntryCount()
	stats["HitCount"] = c.cache.HitCount()
	stats["MissCount"] = c.cache.MissCount()
	stats["LookupCount"] = c.cache.LookupCount()
	stats["HitRate"] = c.cache.HitRate()
	stats["EvacuateCount"] = c.cache.EvacuateCount()
	stats["ExpiredCount"] = c.cache.ExpiredCount()
	return stats
}
