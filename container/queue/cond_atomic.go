package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

// CondAtomic 广播信号的cond 通过channel使用广播唤醒等待者
// 实现不了singal方法，会出现漏信号的情况
type CondAtomic struct {
	L sync.Locker
	n unsafe.Pointer
}

func NewCondAtomic(m sync.Locker) *CondAtomic {
	n := make(chan struct{})
	return &CondAtomic{
		L: m,
		n: unsafe.Pointer(&n),
	}
}

// NotifyChan 返回一个用来等待下一次广播的chan,其实就是原子的读出来Newcond里面建的chan
func (c *CondAtomic) NotifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

// Broadcast 需要持有锁才能调用
func (c *CondAtomic) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	//把老的chan关闭掉，关闭后的通道会立即返回零值。
	close(*(*chan struct{})(ptrOld))
}

func (c *CondAtomic) Wait() {
	n := c.NotifyChan()
	c.L.Unlock()
	<-n
	c.L.Lock()
}

// WaitWithTimeout 在退出之后，应该保持加锁的状态。所以在调用者返回之前，要释放锁
func (c *CondAtomic) WaitWithTimeout(ctx context.Context) error {
	n := c.NotifyChan() //取出通道
	c.L.Unlock()        //释放锁
	select {
	//通道被关闭了，收到了广播
	case <-n:
		c.L.Lock()
		return nil
	case <-ctx.Done():
		c.L.Lock()
		return ctx.Err()
	}

}
