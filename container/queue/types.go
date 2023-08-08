package queue

import (
	"context"
	"sync"
	"unsafe"
)

type Queue[T any] interface {
	Enqueue(ctx context.Context, data T) error
	Dequeue(ctx context.Context) (T, error)
	IsFull() bool
	IsEmpty() bool
	Len() uint64
}

// 转发信号的的cond
type CondV1 struct {
	*sync.Cond
}

// 广播信号的cond 通过channel使用广播唤醒等待者
// 实现不了singal方法，会出现漏信号的情况
type Cond struct {
	L sync.Locker
	n unsafe.Pointer
}
