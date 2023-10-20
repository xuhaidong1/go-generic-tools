package queue

import (
	"context"
	"errors"
)

type Queue[T any] interface {
	Enqueue(ctx context.Context, data T) error
	Dequeue(ctx context.Context) (T, error)
	IsFull() bool
	IsEmpty() bool
	Len() int
}

var ErrQueueEmpty = errors.New("队列空")
