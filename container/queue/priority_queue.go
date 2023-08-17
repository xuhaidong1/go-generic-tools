package queue

import (
	"context"
	go_generic_tools "github.com/xuhaidong1/go-generic-tools"
	"github.com/xuhaidong1/go-generic-tools/container/errs"
)

// PriorityQueue 优先队列。容量固定 小顶堆
type PriorityQueue[T any] struct {
	data       []T
	maxsize    int
	comparator go_generic_tools.Comparator[T]
}

func NewPriorityQueue[T any](maxsize int, comparator go_generic_tools.Comparator[T]) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		data:       make([]T, 0, maxsize),
		maxsize:    maxsize,
		comparator: comparator,
	}
}

func (p *PriorityQueue[T]) Enqueue(ctx context.Context, data T) error {
	if p.IsFull() {
		return errs.NewErrFullQueue()
	}
	//if ctx.Err() != nil {
	//	return ctx.Err()
	//}
	p.data = append(p.data, data)
	p.shiftUp(len(p.data) - 1)
	return nil
}

// Dequeue 这里暂不做超时控制，因为delayqueue需要当超时的时候也要正常出队元素，所以需要优先队列进行正常出队
func (p *PriorityQueue[T]) Dequeue(ctx context.Context) (T, error) {
	var res T
	if p.IsEmpty() {
		return res, errs.NewErrEmptyQueue()
	}
	//if ctx.Err() != nil {
	//	return res, ctx.Err()
	//}
	res = p.data[0]
	p.swap(0, len(p.data)-1)
	p.data = p.data[:len(p.data)-1]
	p.shiftDown(0)
	return res, nil
}

func (p *PriorityQueue[T]) Peek() (T, error) {
	var res T
	if p.IsEmpty() {
		return res, errs.NewErrEmptyQueue()
	}
	//if ctx.Err() != nil {
	//	return res, ctx.Err()
	//}
	res = p.data[0]
	return res, nil
}

func (p *PriorityQueue[T]) IsFull() bool {
	return len(p.data) == p.maxsize
}

func (p *PriorityQueue[T]) IsEmpty() bool {
	return len(p.data) == 0
}

func (p *PriorityQueue[T]) Len() int {
	return len(p.data)
}

func (p *PriorityQueue[T]) shiftUp(i int) {
	for i >= 0 && (i-1)/2 >= 0 {
		parent := (i - 1) / 2
		if p.comparator(p.data[i], p.data[parent]) < 0 {
			p.swap(i, parent)
			i = parent
		} else {
			break
		}
	}
}

func (p *PriorityQueue[T]) shiftDown(i int) {
	for i < p.Len() {
		left, right, smallest := 2*i+1, 2*i+2, i
		if left < p.Len() && p.comparator(p.data[left], p.data[i]) < 0 {
			smallest = left
		}
		if right < p.Len() && p.comparator(p.data[right], p.data[smallest]) < 0 {
			smallest = right
		}
		if smallest == i {
			break
		} else {
			p.swap(smallest, i)
			i = smallest
		}
	}
}

func (p *PriorityQueue[T]) swap(i, j int) {
	p.data[i], p.data[j] = p.data[j], p.data[i]
}
