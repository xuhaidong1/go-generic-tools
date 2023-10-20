package queue

import (
	"context"
	"sync"
)

// ConcurrentQueue 基于切片的并发队列
type ConcurrentQueue[T any] struct {
	data  []T
	mutex *sync.Mutex
	size  int
}

func NewConcurrentQueue[T any](size int) *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{
		data:  make([]T, 0, size),
		mutex: &sync.Mutex{},
		size:  size,
	}
}

func (c *ConcurrentQueue[T]) Enqueue(ctx context.Context, data T) error {
	//先判断有没有过期
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = append(c.data, data)
	return nil
}

func (c *ConcurrentQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isEmpty() {
		var t T
		return t, ErrQueueEmpty
	}
	res := c.data[0]
	c.data = c.data[1:]
	return res, nil
}

func (c *ConcurrentQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == c.size
}

func (c *ConcurrentQueue[T]) isFull() bool {
	return len(c.data) == c.size
}

func (c *ConcurrentQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == 0
}
func (c *ConcurrentQueue[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentQueue[T]) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data)
}
