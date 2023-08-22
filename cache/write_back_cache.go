package cache

import (
	"context"
	"log"
)

// WriteBackCache 采用localcache的删除回调机制实现缓存过期刷新到db
type WriteBackCache struct {
	*LocalCache
	StoreFunc
}

func NewWriteBackCache(storeFunc StoreFunc) *WriteBackCache {
	return &WriteBackCache{
		StoreFunc: storeFunc,
		LocalCache: NewLocalCache(func(key string, val any) {
			//context 和err不好处理
			err := storeFunc(context.Background(), key, val)
			if err != nil {
				log.Fatalln(err)
			}
		}),
	}
}

func (c *WriteBackCache) Close() error {
	//遍历所有的key，刷新到数据库
	return nil
}
