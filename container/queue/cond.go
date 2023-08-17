package queue

import "sync"

type Cond struct {
	l      sync.Locker
	signal chan struct{}
}

func NewCond(l sync.Locker) *Cond {
	return &Cond{
		l:      l,
		signal: make(chan struct{}),
	}
}

// Broadcast 唤醒等待者
// 如果没有人等待，那么什么也不会发生
// 必须加锁之后才能调用这个方法
// 广播之后锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *Cond) Broadcast() {
	n := make(chan struct{})
	old := c.signal
	c.signal = n
	c.l.Unlock()
	close(old)
}

// SignalCh 返回一个 channel，用于监听广播信号
// 必须在锁范围内使用
// 调用后，锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *Cond) SignalCh() <-chan struct{} {
	res := c.signal
	c.l.Unlock()
	return res
}
