package queue

import (
	"context"
	"errors"
	"github.com/xuhaidong1/go-generic-tools/container/errs"
	"sync"
	"time"
)

type Delayable interface {
	Delay() time.Duration
}

type DelayQueue[T Delayable] struct {
	q         *PriorityQueue[T]
	mutex     *sync.RWMutex
	inSignal  *Cond
	outSignal *Cond
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	m := &sync.RWMutex{}
	return &DelayQueue[T]{
		mutex:     m,
		inSignal:  NewCond(m),
		outSignal: NewCond(m),
		q: NewPriorityQueue[T](capacity, func(src, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay < dstDelay {
				return -1
			} else if srcDelay == dstDelay {
				return 0
			} else {
				return 1
			}
		}),
	}
}

// Enqueue 入队与并发阻塞队列区别不大
func (d *DelayQueue[T]) Enqueue(ctx context.Context, data T) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		d.mutex.Lock()
		err := d.q.Enqueue(ctx, data)
		switch err {
		case nil:
			d.inSignal.Broadcast()
			return nil
		case errs.NewErrFullQueue():
			//阻塞 开始等待 这里面会释放锁
			ch := d.outSignal.SignalCh()
			select {
			case <-ch:
			//被唤醒了 回到for开始
			case <-ctx.Done():
				return ctx.Err()
			}
		default:
			d.mutex.Unlock()
			return err
		}
	}
}

// Dequeue 出队就有讲究了：
//  1. Delay() 返回 <= 0 的时候才能出队ok
//  2. 如果队首的 Delay()=300ms >0，要是 sleep，等待 Delay() 降下去ok
//  3. 如果正在 sleep 的过程，有新元素来了，
//     并且 Delay() = 200 比你正在sleep 的时间还要短，你要调整你的 sleep 时间ok
//  4. 如果 sleep 的时间还没到，就超时了，那么就返回元素，没有元素可返回就报empty
//
// sleep 本质上是阻塞（你可以用 time.Sleep，你也可以用 channel）
func (d *DelayQueue[T]) Dequeue(ctx context.Context) (T, error) {
	var res T
	var timer *time.Timer
	for {
		//先看超时
		select {
		case <-ctx.Done():
			return d.timeoutDequeue(ctx)
		default:
		}
		//取出队头
		d.mutex.Lock()
		head, err := d.q.Peek()
		//确保每个分支都有解锁操作
		switch err {
		case nil:
			delay := head.Delay()
			if delay < 0 {
				res, _ = d.q.Dequeue(ctx)
				d.outSignal.Broadcast()
				return res, nil
			}
			signal := d.inSignal.SignalCh()
			if timer == nil {
				timer = time.NewTimer(delay)
			} else {
				timer.Reset(delay)
			}
			//开始等待 等待过程有元素入队就回到for
			select {
			case <-ctx.Done():
				return d.timeoutDequeue(ctx)
			case <-timer.C:
				// 到了时间 原队头可能已经被其他协程先出队，故再次检查队头
				d.mutex.Lock()
				res, err = d.q.Peek()
				if err != nil || res.Delay() > 0 {
					d.mutex.Unlock()
					continue
				}
				// 验证元素过期后将其出队
				res, err = d.q.Dequeue(ctx)
				d.outSignal.Broadcast()
				return res, err
			case <-signal:
				//收到了新元素入队信号 回到循环开始处
			}
		case errs.NewErrEmptyQueue():
			ch := d.inSignal.SignalCh()
			select {
			case <-ch:
			case <-ctx.Done():
				return d.timeoutDequeue(ctx)
			}
		default:
			d.mutex.Unlock()
			//未知错误
			return res, errors.New("未知错误")
		}
	}
}

func (d *DelayQueue[T]) timeoutDequeue(ctx context.Context) (T, error) {
	d.mutex.Lock()
	res, err := d.q.Dequeue(ctx)
	if err != nil {
		d.mutex.Unlock()
		return res, ctx.Err()
	}
	d.mutex.Unlock()
	return res, ctx.Err()
}

func (d *DelayQueue[T]) IsFull() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.q.IsFull()
}

func (d *DelayQueue[T]) IsEmpty() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.q.IsEmpty()
}

func (d *DelayQueue[T]) Len() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.q.Len()
}
