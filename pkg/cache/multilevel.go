package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

type L1Cache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
	size  int
}

func NewL1Cache(size int) *L1Cache {
	return &L1Cache{
		items: make(map[string]*CacheItem),
		size:  size,
	}
}

func (c *L1Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}

func (c *L1Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.size {
		c.evict()
	}

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

func (c *L1Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *L1Cache) evict() {
	oldestKey := ""
	oldestTime := time.Now()

	for k, v := range c.items {
		if v.Expiration.Before(oldestTime) {
			oldestTime = v.Expiration
			oldestKey = k
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

type MultiLevelCache struct {
	l1    *L1Cache
	l2    *redis.Client
	l2TTL time.Duration
	l1TTL time.Duration
}

func NewMultiLevelCache(l1Size int, l2Addr string, l1TTL, l2TTL time.Duration) *MultiLevelCache {
	l2 := redis.NewClient(&redis.Options{
		Addr:     l2Addr,
		Password: "",
		DB:       0,
	})

	return &MultiLevelCache{
		l1:    NewL1Cache(l1Size),
		l2:    l2,
		l1TTL: l1TTL,
		l2TTL: l2TTL,
	}
}

func (c *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	if val, ok := c.l1.Get(key); ok {
		return val, nil
	}

	val, err := c.l2.Get(ctx, key).Result()
	if err == nil {
		c.l1.Set(key, val, c.l1TTL)
		return val, nil
	}

	return nil, err
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}) error {
	c.l1.Set(key, value, c.l1TTL)
	return c.l2.Set(ctx, key, value, c.l2TTL).Err()
}

func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
	c.l1.Delete(key)
	return c.l2.Del(ctx, key).Err()
}

func (c *MultiLevelCache) GetOrSet(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
	if val, ok := c.l1.Get(key); ok {
		return val, nil
	}

	val, err := c.l2.Get(ctx, key).Result()
	if err == nil {
		c.l1.Set(key, val, c.l1TTL)
		return val, nil
	}

	data, err := loader()
	if err != nil {
		return nil, err
	}

	c.Set(ctx, key, data)
	return data, nil
}
