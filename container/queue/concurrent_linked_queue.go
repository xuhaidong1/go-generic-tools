package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"unsafe"
)

// ConcurrentLinkedQueue 基于CAS操作的无锁并发队列，采用链表实现
type ConcurrentLinkedQueue[T any] struct {
	head  unsafe.Pointer //head是哨兵
	tail  unsafe.Pointer
	count int32
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkedQueue[T] {
	head := &node[T]{}
	ptr := unsafe.Pointer(head)
	return &ConcurrentLinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

type node[T any] struct {
	Val  T
	Next unsafe.Pointer
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, data T) error {
	newNode := &node[T]{
		Val: data,
	}
	newNodePtr := unsafe.Pointer(newNode)
	for {
		//超时判断不准确 误差在纳秒级
		if ctx.Err() != nil {
			return ctx.Err()
		}
		tailPtr := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPtr)
		tailNext := atomic.LoadPointer(&tail.Next)
		if tailNext != nil {
			//别人抢先enqueue了 下一轮即可
			continue
		}
		//并发安全的c.tail.Next=newNode
		if atomic.CompareAndSwapPointer(&tail.Next, tailNext, newNodePtr) {
			//并发安全的c.tail = c.tail.Next
			atomic.CompareAndSwapPointer(&c.tail, tailPtr, newNodePtr)
			atomic.AddInt32(&c.count, 1)
			return nil
		}
	}
}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	for {
		//超时判断不准确 误差在纳秒级
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		//res := c.head
		headPtr := atomic.LoadPointer(&c.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPtr)
		if head == tail {
			//如果进来了就认为这一刻队列为空，即使有人在这一刻正在调整指针进行入队
			//未来的事情在这一刻无法预测，未来CAS失败了可以等待下一次循环
			var t T
			return t, errors.New("队列为空")
		}
		headNextPtr := atomic.LoadPointer(&head.Next)
		//并发安全的c.head = c.head.Next
		//如果在这里队列突然变空了 下面的CAS会失败
		if atomic.CompareAndSwapPointer(&c.head, headPtr, headNextPtr) {
			atomic.AddInt32(&c.count, -1)
			return (*node[T])(headNextPtr).Val, nil
		}
	}
}

func (c *ConcurrentLinkedQueue[T]) IsFull() bool {
	panic("未提供实现")
}

// IsEmpty 这种实现返回不准确的结果，可能入队了1个元素但未执行到tail的CAS操作
func (c *ConcurrentLinkedQueue[T]) IsEmpty() bool {
	return atomic.LoadPointer(&c.head) == atomic.LoadPointer(&c.tail)
}

// Len 这种实现返回不准确的结果，可能有元素执行完成入队/出队CAS操作。但还没count+1
func (c *ConcurrentLinkedQueue[T]) Len() int {
	return int(atomic.LoadInt32(&c.count))
}
