package cache

import (
	"context"
	"math/rand"
	"time"
)

// RandomExpirationCache 增加随机偏移过期时间，以防key统一过期造成缓存雪崩
type RandomExpirationCache struct {
	Cache
	offset func() time.Duration
}

func NewRandomExpirationCache(cache Cache, offset func() time.Duration) *RandomExpirationCache {
	return &RandomExpirationCache{
		Cache:  cache,
		offset: offset,
	}
}

func (c RandomExpirationCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return c.Cache.Set(ctx, key, val, expiration+c.offset())
}

func offset() time.Duration {
	i := rand.Intn(10000)
	return time.Duration(i) * time.Millisecond
}
