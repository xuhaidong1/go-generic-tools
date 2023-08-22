package cache

import (
	"context"
	"fmt"
	"github.com/xuhaidong1/go-generic-tools/cache/errs"
	"log"
)

// BloomFilterCache 布隆过滤器cache，用于解决缓存穿透的问题，当攻击者伪造大量不同key攻击时会直接打到数据库，
// 这个cache再查库之前先调用用户传进来的Exist方法看有没有这个key
type BloomFilterCache struct {
	bf BloomFilter
	ReadThroughCache
}

func (c *BloomFilterCache) Get(ctx context.Context, key string) (any, error) {
	val, err := c.ReadThroughCache.Get(ctx, key)
	if err != nil && err != errs.NewErrKeyNotFound(key) {
		return nil, err
	}
	if ok := c.bf(ctx, key); !ok {
		return nil, errs.NewErrKeyNotFound(key)
	}
	val, err = c.LoadFunc(ctx, key)
	if err != nil {
		//包一层错误信息 方便定位
		return nil, fmt.Errorf("cache:无法加载数据 %w", err)
	}
	if err := c.Set(ctx, key, val, c.Expiration); err != nil {
		log.Fatalln(err) //这里err可以考虑忽略掉
	}
	return val, nil
}
