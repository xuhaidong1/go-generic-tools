package cache

import (
	"context"
	"fmt"
	"github.com/xuhaidong1/go-generic-tools/cache/errs"
	"log"
	"time"
)

type ReadThroughCache struct {
	Cache
	Expiration time.Duration
	//把捞DB抽象为“加载数据”
	LoadFunc
}

func NewReadThroughCache(cache Cache, Expiration time.Duration, loadFunc LoadFunc) *ReadThroughCache {
	return &ReadThroughCache{
		Cache:      cache,
		Expiration: Expiration,
		LoadFunc:   loadFunc,
	}
}

// Get 加锁问题：先穿透读 再有人写数据库，数据就会不一致；加了写锁也会有不一致，保证了读，但直接写cache就会不一致
func (c *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	//先捞缓存 再捞db
	val, err := c.Cache.Get(ctx, key)
	//不知道哪里出问题了
	if err != nil && err != errs.NewErrKeyNotFound(key) {
		return nil, err
	}
	if err == errs.NewErrKeyNotFound(key) {

		val, err = c.LoadFunc(ctx, key)
		if err != nil {
			//包一层错误信息 方便定位
			return nil, fmt.Errorf("cache:无法加载数据 %w", err)
		}

		err = c.Set(ctx, key, val, c.Expiration)
		//这里err可以考虑忽略掉
		if err != nil {
			log.Fatalln(err)
		}
		return val, nil
	}
	return val, nil

}
