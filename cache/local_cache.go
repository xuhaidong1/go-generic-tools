package cache

import (
	"context"
	"github.com/xuhaidong1/go-generic-tools/cache/errs"
	"sync"
	"time"
)

// LocalCache 过期key删除采用懒惰关闭+轮询关闭
type LocalCache struct {
	//m         sync.Map
	data      map[string]any
	mutex     sync.RWMutex
	close     chan struct{}
	closeOnce sync.Once
	onEvicted func(key string, val any)
}

type item struct {
	Val      any
	Deadline time.Time
}

func NewLocalCache(onEvicted func(key string, val any)) *LocalCache {
	ticker := time.NewTicker(time.Second)
	ch := make(chan struct{})
	res := &LocalCache{
		close:     ch,
		onEvicted: onEvicted,
	}
	//开一个goroutine用于删除过期的key
	go func() {
		for {
			select {
			case <-ticker.C:
				res.mutex.Lock()
				cnt := 0
				for key, val := range res.data {
					itm := val.(*item)
					if itm.Deadline.Before(time.Now()) {
						res.delete(key, itm)
					}
					cnt++
					if cnt > 2000 {
						break
					}
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()
	return res
}

func (l *LocalCache) Get(ctx context.Context, key string) (any, error) {
	//用sync.Map不行，从捞出来到检查过期可能过了很久了，需要加锁
	l.mutex.RLock()
	itm, ok := l.data[key]
	l.mutex.RUnlock()
	if !ok {
		return nil, errs.NewErrKeyNotFound(key)
	}
	res := itm.(*item)
	//有可能别人在这里调用了set，double check
	if res.Deadline.Before(time.Now()) {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		itm, ok := l.data[key]
		if !ok {
			return nil, errs.NewErrKeyNotFound(key)
		}
		res := itm.(*item)
		if res.Deadline.Before(time.Now()) {
			delete(l.data, key)
		}
		return nil, errs.NewErrKeyNotFound(key)
	}
	return res.Val, nil
}

func (l *LocalCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.data[key] = &item{
		Val:      val,
		Deadline: time.Now().Add(expiration),
	}
	return nil
}

func (l *LocalCache) Delete(ctx context.Context, key string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	val, ok := l.data[key]
	if !ok {
		return nil
	}
	l.delete(key, val)
	return nil
}

func (l *LocalCache) delete(key string, val any) {
	delete(l.data, key)
	if l.onEvicted != nil {
		l.onEvicted(key, val.(*item).Val)
	}
}

// Close 需要考虑用户重复close
func (l *LocalCache) Close() error {
	l.closeOnce.Do(func() {
		l.close <- struct{}{}
	})
	return nil
	//select{
	//case l.close<- struct{}{}:
	//default:
	//}
	//return nil
}
