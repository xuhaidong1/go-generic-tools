package queue

import (
	"context"
	"sync"
)

// ConcurrentBlockingQueueRingBuffer 基于ringBuffer的并发阻塞队列，使用支持timeout的cond来实现信号的广播，不保证FIFO唤醒等待者
type ConcurrentBlockingQueueRingBuffer[T any] struct {
	data     []T
	maxSize  int
	count    int
	head     int
	tail     int
	zero     T
	mutex    *sync.Mutex
	notFull  *Cond
	notEmpty *Cond
}

func NewConcurrentBlockingQueueRingBuffer[T any](maxsize int) *ConcurrentBlockingQueueRingBuffer[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueueRingBuffer[T]{
		//使用ringbuffer时应该一开始将所有内存都分配好
		data:     make([]T, maxsize, maxsize),
		mutex:    m,
		maxSize:  maxsize,
		notFull:  NewCond(m),
		notEmpty: NewCond(m),
	}
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) Enqueue(ctx context.Context, data T) error {
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
	c.data[c.tail] = data
	c.tail++
	if c.tail == c.maxSize {
		c.tail = 0
	}
	c.count++
	c.notEmpty.Broadcast()
	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) Dequeue(ctx context.Context) (T, error) {
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
	res := c.data[c.head]
	c.data[c.head] = c.zero //空地填充类型0值
	c.head++
	if c.head == c.maxSize {
		c.head = 0
	}
	c.count--
	c.notFull.Broadcast()
	c.mutex.Unlock()
	return res, nil
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count == c.maxSize
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) isFull() bool {
	return c.count == c.maxSize
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count == 0
}
func (c *ConcurrentBlockingQueueRingBuffer[T]) isEmpty() bool {
	return c.count == 0
}

func (c *ConcurrentBlockingQueueRingBuffer[T]) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.count
}
