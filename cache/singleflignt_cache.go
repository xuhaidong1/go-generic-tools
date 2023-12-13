package cache

import (
	"context"
	"fmt"
	"github.com/xuhaidong1/go-generic-tools/cache/errs"
	"golang.org/x/sync/singleflight"
	"log"
)

// SingleFlightCache 采用singleFlight调用去访问数据库，解决缓存击穿的问题
// 保证相同的key只有一个goroutine会实际查询数据库，调用完成forget一下让其它等待的goroutine不等了
type SingleFlightCache struct {
	ReadThroughCache
	g *singleflight.Group
}

func (c *SingleFlightCache) Get(ctx context.Context, key string) (any, error) {
	//先捞缓存 再捞db
	val, err := c.Cache.Get(ctx, key)
	//不知道哪里出问题了
	if err != nil && err != errs.NewErrKeyNotFound(key) {
		return nil, err
	}
	if err == errs.NewErrKeyNotFound(key) {
		defer c.g.Forget(key)
		//采用singleFlight调用去访问数据库，保证相同的key只有一个goroutine会实际查询数据库，调用完成forget一下让其它等待的goroutine不等了
		val, err, _ = c.g.Do(key, func() (interface{}, error) {
			value, err := c.LoadFunc(ctx, key)
			if err != nil {
				//包一层错误信息 方便定位
				return nil, fmt.Errorf("cache:无法加载数据 %w", err)
			}
			if err := c.Set(ctx, key, value, c.Expiration); err != nil {
				log.Fatalf("刷新缓存失败%v\n", err)
			}
			//这里err可以考虑忽略掉
			return val, nil
		})

	}
	return val, err
}
