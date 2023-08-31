package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type BulidinMapCache struct {
	lock          sync.RWMutex
	data          map[string]*item
	close         chan struct{}
	closed        bool
	onEvicted     func(key string, val any)
	cycleInterval time.Duration
}

func NewBuildinMapCache(opts ...CacheOption) *BulidinMapCache {
	res := &BulidinMapCache{
		data:          make(map[string]*item),
		cycleInterval: time.Second * 10,
	}
	for _, opt := range opts {
		opt(res)
	}
	res.checkCycle()
	return res
}

func (c *BulidinMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return errors.New("缓存已经被关闭")
	}
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	c.data[key] = &item{
		Val:      val,
		Deadline: dl,
	}
	return nil
}

func (c *BulidinMapCache) Get(ctx context.Context, key string) (any, error) {
	if c.closed {
		return nil, ErrCacheClosed
	}
	c.lock.RLock()
	itm, ok := c.data[key]
	c.lock.RUnlock()
	if !ok {
		return nil, ErrCacheKeyNotExist
	}
	// double check 以防别的 goroutine 设置值了
	now := time.Now()
	if itm.deadlineBefore(now) {
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.closed {
			return nil, ErrCacheClosed
		}
		itm, ok = c.data[key]
		if !ok {
			return nil, ErrCacheKeyNotExist
		}
		if itm.deadlineBefore(now) {
			c.delete(key)
			return nil, ErrCacheKeyNotExist
		}
	}
	return itm.Val, nil
}

func (c *BulidinMapCache) OnEvicted(fn func(key string, val any)) {
	oldfn := c.onEvicted
	c.onEvicted = func(key string, val any) {
		if oldfn != nil {
			oldfn(key, val)
		}
		fn(key, val)
	}
}

type CacheOption func(b *BulidinMapCache)

func WithCycleInterval(interval time.Duration) CacheOption {
	return func(b *BulidinMapCache) {
		b.cycleInterval = interval
	}
}

func WithOnEvicted(onEvicted func(key string, val any)) CacheOption {
	return func(b *BulidinMapCache) {
		b.onEvicted = onEvicted
	}
}

func (c *BulidinMapCache) checkCycle() {
	go func() {
		ticker := time.NewTicker(c.cycleInterval)
		for {
			select {
			case now := <-ticker.C:
				c.lock.Lock()
				for key, itm := range c.data {
					// 设置了过期时间，并且已经过期
					if itm.deadlineBefore(now) {
						c.delete(key)
					}
				}
				c.lock.Unlock()
			case <-c.close:
				close(c.close)
				return
			}
		}
	}()
}

func (c *BulidinMapCache) Delete(ctx context.Context, key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return ErrCacheClosed
	}
	c.delete(key)
	return nil
}

func (c *BulidinMapCache) delete(key string) {
	//log.Printf("mapCache 中的delete %s\n", key)
	itm, ok := c.data[key]
	if ok {
		delete(c.data, key)
		if c.onEvicted != nil {
			c.onEvicted(key, itm.Val)
		}
	}
}

func (c *BulidinMapCache) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.close <- struct{}{}
	c.closed = true
	if c.onEvicted != nil {
		for key, itm := range c.data {
			c.onEvicted(key, itm.Val)
		}
	}
	c.data = nil
	return nil
}

func (i *item) deadlineBefore(t time.Time) bool {
	return !i.Deadline.IsZero() && i.Deadline.Before(t)
}

func (c *BulidinMapCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.closed {
		return 0, ErrCacheClosed
	}
	c.lock.RLock()
	itm, ok := c.data[key]
	c.lock.RUnlock()
	if !ok {
		return 0, ErrCacheKeyNotExist
	}
	// double check 以防别的 goroutine 设置值了
	now := time.Now()
	if itm.deadlineBefore(now) {
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.closed {
			return 0, ErrCacheClosed
		}
		itm, ok = c.data[key]
		if !ok {
			return 0, ErrCacheKeyNotExist
		}
		if itm.deadlineBefore(now) {
			c.delete(key)
			return 0, ErrCacheKeyNotExist
		}
	}
	return itm.Deadline.Sub(time.Now()), nil
}

func (c *BulidinMapCache) Expire(key string, expiration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return ErrCacheClosed
	}
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	itm, ok := c.data[key]
	if !ok {
		return ErrCacheKeyNotExist
	}
	c.data[key] = &item{
		Val:      itm.Val,
		Deadline: dl,
	}
	return nil
}

// KeysAsSlice 仅供单元测试使用
func (c *BulidinMapCache) keysAsSlice() []any {
	var res []any
	for k := range c.data {
		res = append(res, k)
	}
	return res
}
