package queue

import (
	"context"
	"sync"
)

// ConcurrentBlockingQueueV1 基于切片的并发阻塞队列，封装sync.Cond增加了超时信号转发机制，不保证FIFO唤醒等待者
type ConcurrentBlockingQueueV1[T any] struct {
	data     []T
	maxSize  int
	mutex    *sync.Mutex
	notFull  *CondV1
	notEmpty *CondV1
	//notFull  chan struct{} channel会出现信号不能按预期接收到，可能出现队列不满但排队情况
	//notEmpty chan struct{}
	//notFull  *sync.Cond 使用cond会一睡不醒，无法兼顾超时
	//notEmpty *sync.Cond
}

func NewConcurrentBlockingQueueV1[T any](maxsize int) *ConcurrentBlockingQueueV1[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueueV1[T]{
		data:     make([]T, 0, maxsize),
		mutex:    m,
		maxSize:  maxsize,
		notFull:  NewCondV1(m),
		notEmpty: NewCondV1(m),
	}
}

func (c *ConcurrentBlockingQueueV1[T]) Enqueue(ctx context.Context, data T) error {
	//先判断有没有过期
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	//采用无锁方法，因为已经加过锁了
	for c.isFull() {
		//基于【超时转发信号的cond】的方法
		err := c.notFull.WaitwithTimeout(ctx)
		if err != nil {
			return err
		}
		//以下是基于channel的实现
		//满了，我阻塞住我自己，直到有人唤醒我
		//c.mutex.Unlock()
		//超时检测和唤醒channel平级，都可以兼顾到
		//select {
		//case <-ctx.Done():
		//	return ctx.Err()
		//case <-c.notFull:
		//	//如果不满，再次拿到锁往下走，到for那里就会退出循环了
		//	c.mutex.Lock()
		//}
	}
	c.data = append(c.data, data)
	c.notEmpty.Signal()
	//没有人等这个信号，这句就会阻塞住
	//c.notEmpty <- struct{}{}
	c.mutex.Unlock()
	return nil

	//以下为cond的实现，存在不能兼顾超时的问题
	//select {
	//case <-ctx.Done():
	//	c.mutex.Unlock()
	//	return ctx.Err()
	//
	//default:
	//	//容量判断：为什么不同if？---g1，g2入队被阻塞，此时释放一个位置，g1g2同时被唤醒，g2先占坑之后，若用if，g1也会接着入队，容量超出了
	//	for c.IsFull() {
	//	//满了，我阻塞住我自己，知道有人唤醒我
	//	//wait做的事情check加锁了，加入到cond的队列里面，释放锁，开始等待，被唤醒后拿到锁锁住
	//	c.notFull.Wait()
	//	//再次拿到锁之后并没有考虑超时问题
	//	}
	//}

}

func (c *ConcurrentBlockingQueueV1[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	for c.isEmpty() {
		//基于【超时转发信号的cond】的方法
		err := c.notEmpty.WaitwithTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
		//以下为基于channel的实现
		////进select之前先解锁，不然会死锁
		//c.mutex.Unlock()
		////阻塞住自己，等待元素入队
		//select {
		//case <-ctx.Done():
		//	var t T
		//	return t, ctx.Err()
		//case <-c.notEmpty:
		//	c.mutex.Lock()
		//}
	}
	res := c.data[0]
	c.data = c.data[1:]
	c.notFull.Signal()
	//只有len从maxsize到maxsize-1变化时，才会有满到不满的状态变化，因此chan的容量设置为1
	//也只有状态转换的场景下，才需要给信号  不对。。。 g1出去两个才给1个信号，g2由于只收到1个信号所以只能进入一个，虽然没满，但其他想进的还需要阻塞
	//if len(c.data) == c.maxSize-1 {
	//	c.notFull <- struct{}{}
	//}
	//可能在Enqueue出现没满但被阻塞住了 在enqueue的select执行之前 这里走过了default分支
	//select {
	//case c.notFull <- struct{}{}:
	//default:
	//}
	c.mutex.Unlock()
	return res, nil
}

func (c *ConcurrentBlockingQueueV1[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueueV1[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueueV1[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == 0
}
func (c *ConcurrentBlockingQueueV1[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentBlockingQueueV1[T]) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data)
}
