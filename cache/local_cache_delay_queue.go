package cache

import (
	"context"
	"fmt"
	"github.com/xuhaidong1/go-generic-tools/container/queue"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// LocalCacheDelayQueue 轮询过期和延时队列二选一，如下是延时队列实现
type LocalCacheDelayQueue struct {
	lock      sync.RWMutex
	data      map[string]*itemDelay
	size      int32
	count     int32
	close     chan struct{}
	closed    bool
	onEvicted func(key string, val any)
	//轮询时间间隔
	//cycleInterval time.Duration
	//延时队列
	delayQueue *queue.DelayQueue[itemDelay]
}

type itemDelay struct {
	key      string
	val      any
	deadline time.Time
}

func (i itemDelay) Delay() time.Duration {
	return i.deadline.Sub(time.Now())
}

func NewLocalCacheDelayQueue(size int) *LocalCacheDelayQueue {
	res := &LocalCacheDelayQueue{
		data: make(map[string]*itemDelay),
		//cycleInterval: time.Second * 10,
		delayQueue: queue.NewDelayQueue[itemDelay](size),
	}
	res.onEvicted = func(key string, val any) {
		atomic.AddInt32(&res.count, -1)
	}
	return res
}

func (c *LocalCacheDelayQueue) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return ErrCacheClosed
	}

	if _, ok := c.data[key]; !ok {
		cnt := atomic.AddInt32(&c.count, 1)
		//先加再减
		if cnt > c.size {
			atomic.AddInt32(&c.count, -1)
			return ErrCacheFull
		}
	}
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	itm := itemDelay{
		key:      key,
		val:      val,
		deadline: dl,
	}
	err := c.delayQueue.Enqueue(ctx, itm)
	if err != nil {
		return err
	}
	c.data[key] = &itm
	return nil
}

func (c *LocalCacheDelayQueue) Get(ctx context.Context, key string) (any, error) {
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
	return itm.val, nil
}

// AutoExpire 用延时队列处理过期key
func (c *LocalCacheDelayQueue) AutoExpire() {
	go func() {
		for {
			select {
			case <-c.close:
				close(c.close)
				return
			default:
				itm, err := c.delayQueue.Dequeue(context.Background())
				if err != nil {
					log.Fatalln(fmt.Sprintf("localcache delayqueue err %s", err))
				}
				err = c.Delete(context.Background(), itm.key)
				if err != nil {
					log.Fatalln(fmt.Sprintf("localcache delayqueue delete err %s", err))
				}
			}
		}
	}()
}

func (c *LocalCacheDelayQueue) Delete(ctx context.Context, key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return ErrCacheClosed
	}
	c.delete(key)
	return nil
}

func (c *LocalCacheDelayQueue) delete(key string) {
	//log.Printf("mapCache 中的delete %s\n", key)
	itm, ok := c.data[key]
	if ok {
		delete(c.data, key)
		if c.onEvicted != nil {
			c.onEvicted(key, itm.val)
		}
	}
}

func (c *LocalCacheDelayQueue) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.close <- struct{}{}
	c.closed = true
	if c.onEvicted != nil {
		for key, itm := range c.data {
			c.onEvicted(key, itm.val)
		}
	}
	c.data = nil
	return nil
}

func (i itemDelay) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

// KeysAsSlice 仅供单元测试使用
func (c *LocalCacheDelayQueue) keysAsSlice() []any {
	var res []any
	for k := range c.data {
		res = append(res, k)
	}
	return res
}
