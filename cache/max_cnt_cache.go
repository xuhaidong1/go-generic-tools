package cache

import (
	"context"
	"errors"
	"github.com/xuhaidong1/go-generic-tools/cache/errs"
	"sync"
	"sync/atomic"
	"time"
)

type MaxCntCache struct {
	Cache
	mutex    sync.Mutex
	MaxCount int32
	count    int32
}

func NewMaxCntCache(maxcount int32) *MaxCntCache {
	res := &MaxCntCache{
		MaxCount: maxcount,
	}
	//因为localcache有多个地方都会delete，应该里面delete的时候也要控制外面计数
	res.Cache = NewLocalCache(func(key string, val any) {
		atomic.AddInt32(&res.MaxCount, -1)
	})
	return res
}

func (m *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, err := m.Cache.Get(ctx, key)
	if err != nil && err == errs.NewErrKeyNotFound(key) {
		cnt := atomic.AddInt32(&m.count, 1)
		//先加再减
		if cnt > m.MaxCount {
			atomic.AddInt32(&m.count, -1)
			return errors.New("")
		}
	}
	return m.Cache.Set(ctx, key, val, expiration)
}
