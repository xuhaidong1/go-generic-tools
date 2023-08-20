package cache

import (
	"context"
	"time"
)

type WriteThroughCache struct {
	Cache
	Expiration time.Duration
	//把捞DB抽象为“加载数据”
	StoreFunc
}

func NewWriteThroughCache(cache Cache, Expiration time.Duration, storeFunc StoreFunc) *WriteThroughCache {
	return &WriteThroughCache{
		Cache:      cache,
		Expiration: Expiration,
		StoreFunc:  storeFunc,
	}
}

func (c *WriteThroughCache) Get(ctx context.Context, key string) (any, error) {
	return c.Cache.Get(ctx, key)
}

func (c *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	//在这里开goroutine是全异步
	err := c.StoreFunc(ctx, key, val)
	if err != nil {
		return err
	}
	//在这里开goroutine是半异步
	return c.Cache.Set(ctx, key, val, expiration)
}

func (c *WriteThroughCache) Delete(ctx context.Context, key string) error {
	return c.Cache.Delete(ctx, key)
}
