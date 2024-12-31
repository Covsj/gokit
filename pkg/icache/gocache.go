package icache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GoCache struct {
	cache *gocache.Cache
}

// NewGoCache 创建一个新的缓存实例
// defaultExpiration: 默认的过期时间
// cleanupInterval: 清理过期项的时间间隔
func NewGoCache(defaultExpiration, cleanupInterval time.Duration) *GoCache {
	return &GoCache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *GoCache) Set(key string, value interface{}) {
	c.cache.Set(key, value, gocache.DefaultExpiration)
}

func (c *GoCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *GoCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *GoCache) Delete(key string) {
	c.cache.Delete(key)
}

func (c *GoCache) Clear() {
	c.cache.Flush()
}

func (c *GoCache) Len() int {
	return c.cache.ItemCount()
}
