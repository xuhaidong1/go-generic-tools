package cache

import (
	"context"
	"log"
	"time"
)

// PreloadCache 预加载cache，set的时候同时set哨兵cache和实际cache，哨兵cache只存储key，
// 且过期时间要比主cache早，哨兵cache过期的时候执行回调： 执行主cache的loadfunc刷新主cache的缓存。
// 适合对读性能（缓存命中率）要求较高的场景
type PreloadCache struct {
	SentinelCache Cache
	ReadThroughCache
	expiration time.Duration
}

func NewPreloadCache(expiration time.Duration, onEvicted func(key string, val any), loadFunc LoadFunc) *PreloadCache {
	c := ReadThroughCache{NewLocalCache(onEvicted), expiration, loadFunc}
	return &PreloadCache{
		expiration:       expiration,
		ReadThroughCache: c,
		SentinelCache: NewLocalCache(func(key string, val any) {
			res, err := loadFunc(context.Background(), key)
			if err != nil {
				log.Fatalf("cache:sentinel预加载失败%v\n", err)
				return
			}
			err = c.Set(context.Background(), key, res, expiration)
			if err != nil {
				log.Fatalf("cache:主cache预存失败%v\n", err)
			}
		}),
	}
}

func (c *PreloadCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// sentinelExpiration 的设置原则是：
	// 确保 expiration - sentinelExpiration 这段时间内，来得及加载数据刷新缓存
	// 要注意 OnEvicted 的时机，尤其是懒删除，但是轮询删除效果又不是很好的时候
	err := c.SentinelCache.Set(ctx, key, nil, expiration-time.Second*3)
	if err != nil {
		log.Fatalf("cache:写入sentinel cache失败%v\n", err)
	}
	err = c.ReadThroughCache.Set(ctx, key, val, expiration)
	if err != nil {
		return err
	}
	return nil
}
