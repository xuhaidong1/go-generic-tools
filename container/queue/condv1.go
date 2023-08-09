package queue

import (
	"context"
	"sync"
)

// CondV1 转发信号的的cond 基于sync.Cond实现
type CondV1 struct {
	*sync.Cond
}

func NewCondV1(m sync.Locker) *CondV1 {
	return &CondV1{
		Cond: sync.NewCond(m),
	}
}

func (c *CondV1) WaitwithTimeout(ctx context.Context) error {
	//这个channel负责从下面的goroutine向外部函数传递等待结果
	ch := make(chan struct{})
	go func() {
		c.Cond.Wait() //wait被唤醒后是拿到了锁
		select {
		case ch <- struct{}{}:
			//发不进去信号说明外部函数返回了，ch被回收了，所以会继续走default
		default:
			//这里已经超时返回了，要转发信号唤醒别人
			c.Cond.Signal()
			//转发之后要释放掉锁，不能占着锁不用
			//这句有死锁隐患
			c.Cond.L.Unlock()
		}
	}()
	select {
	//收到这个说明超时了
	case <-ctx.Done():
		return ctx.Err()
		//收到这个说明goroutine里的wait被唤醒，ch还在并被填入了消息，函数还没返回，说明没超时
	case <-ch:
		//真的被唤醒了
		return nil
	}
}
