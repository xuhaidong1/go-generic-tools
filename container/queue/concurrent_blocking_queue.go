package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConcurrentBlockingQueue[T any] struct {
	data     []T
	maxSize  int
	mutex    *sync.Mutex
	notFull  *Cond
	notEmpty *Cond
}

func NewConcurrentBlockingQueue[T any](maxsize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:     make([]T, 0, maxsize),
		mutex:    m,
		maxSize:  maxsize,
		notFull:  NewCond(m),
		notEmpty: NewCond(m),
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

func NewCond(m sync.Locker) *Cond {
	n := make(chan struct{})
	return &Cond{
		L: m,
		n: unsafe.Pointer(&n),
	}
}

// NotifyChan 返回一个用来等待下一次广播的chan,其实就是原子的读出来Newcond里面建的chan
func (c *Cond) NotifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

func (c *Cond) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	//把老的chan关闭掉，关闭后的通道会立即返回零值。
	close(*(*chan struct{})(ptrOld))
}

func (c *Cond) Wait() {
	n := c.NotifyChan()
	c.L.Unlock()
	<-n
	c.L.Lock()
}

func (c *Cond) WaitWithTimeout(ctx context.Context) error {
	n := c.NotifyChan() //取出通道
	c.L.Unlock()        //释放锁
	select {
	//通道被关闭了，收到了广播
	case <-n:
		c.L.Lock()
		return nil
	case <-ctx.Done():
		//c.L.Lock()
		return ctx.Err()
	}

}
