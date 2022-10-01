package utils

import (
	"context"
	"sync"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/anzdata"
)

// NewCache is a helper to create instance of the Cache struct
func NewCache(ctx context.Context, cacheTimeInSeconds int64) *Cache {
	cache := &Cache{
		data:               make(map[string]*CacheValue),
		cacheTimeInSeconds: cacheTimeInSeconds,
	}
	cache.startExpirationProcessing(ctx)
	return cache
}

// Cache that will invalidate cached item after specified cacheTimeInSeconds
type Cache struct {
	sync.Mutex
	data               map[string]*CacheValue
	cacheTimeInSeconds int64
}

type CacheValue struct {
	data      anzdata.User
	timestamp int64
}

func (c *Cache) Set(key string, value anzdata.User) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = &CacheValue{
		data:      value,
		timestamp: time.Now().Unix(),
	}
}

func (c *Cache) Get(key string) interface{} {
	c.Lock()
	defer c.Unlock()
	if item, ok := c.data[key]; ok {
		return item.data
	}
	return nil
}

func (c *Cache) Dump() map[string]*CacheValue {
	c.Lock()
	defer c.Unlock()

	return c.data
}

// Clean up items in cache that are old than cacheTimeInSeconds
func (c *Cache) startExpirationProcessing(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Duration(c.cacheTimeInSeconds) * time.Second)

		for t := range ticker.C {
			c.Lock()
			logf.Debug(ctx, "start cleaning data... time: %s, cache size: %d \n", t, len(c.data))
			currentTimeStamp := time.Now().Unix()
			for k, item := range c.data {
				if currentTimeStamp-item.timestamp > c.cacheTimeInSeconds {
					delete(c.data, k)
				}
			}
			c.Unlock()
		}
	}()
}
