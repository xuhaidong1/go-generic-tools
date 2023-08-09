package queue

import (
	"context"
)

type Queue[T any] interface {
	Enqueue(ctx context.Context, data T) error
	Dequeue(ctx context.Context) (T, error)
	IsFull() bool
	IsEmpty() bool
	Len() uint64
}
