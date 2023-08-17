package queue

import (
	"context"
	"sync"
)

// ConcurrentBlockingQueue 基于切片的并发阻塞队列，使用支持timeout的cond来实现信号的广播，不保证FIFO唤醒等待者
type ConcurrentBlockingQueue[T any] struct {
	data     []T
	maxSize  int
	mutex    *sync.Mutex
	notFull  *CondAtomic
	notEmpty *CondAtomic
}

func NewConcurrentBlockingQueue[T any](maxsize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:     make([]T, 0, maxsize),
		mutex:    m,
		maxSize:  maxsize,
		notFull:  NewCondAtomic(m),
		notEmpty: NewCondAtomic(m),
	}
}

func (c *ConcurrentBlockingQueue[T]) Enqueue(ctx context.Context, data T) error {
	//先判断有没有过期
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	//isFull采用无锁方法，因为已经加过锁了
	for c.isFull() {
		//基于【广播信号cond】的方法
		err := c.notFull.WaitWithTimeout(ctx)
		if err != nil {
			c.mutex.Unlock()
			return err
		}
	}
	c.data = append(c.data, data)
	c.notEmpty.Broadcast()
	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	for c.isEmpty() {
		//基于【超时转发信号的cond】的方法
		err := c.notEmpty.WaitWithTimeout(ctx)
		if err != nil {
			c.mutex.Unlock()
			var t T
			return t, err
		}
	}
	res := c.data[0]
	c.data = c.data[1:]
	c.notFull.Broadcast()
	c.mutex.Unlock()
	return res, nil
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == 0
}
func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data)
}
