package queue

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 高并发入队 结果全成功（中间过程无法校验）
func TestConcurrentLinkedQueueEnqueue(t *testing.T) {
	var inSucc int32 = 0
	var inFail int32 = 0
	//var outSucc int32 = 0
	//var outFail int32 = 0
	q := NewConcurrentLinkedQueue[int]()
	var wg sync.WaitGroup
	goroutineNum := 200000
	wg.Add(goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		j := i
		go func() {
			//这个超时时间对于20w并发来说够用（Apple M1），200w会出现少量超时
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			err := q.Enqueue(ctx, j)
			if err != nil {
				atomic.AddInt32(&inFail, 1)
			} else {
				atomic.AddInt32(&inSucc, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(goroutineNum), inSucc)
	assert.Equal(t, int32(0), inFail)
	assert.Equal(t, int32(goroutineNum), q.count)
	//fmt.Printf("insucc: %d\n", inSucc)
	//fmt.Printf("infail: %d\n", inFail)
	//fmt.Printf("outsucc: %d\n", outSucc)
	//fmt.Printf("outfail: %d\n", outFail)
}

// 高并发出队 结果全成功（中间过程无法校验）
func TestConcurrentLinkedQueueDequeue(t *testing.T) {
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := NewConcurrentLinkedQueue[int]()
	var wg sync.WaitGroup
	goroutineNum := 200000
	wg.Add(goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		j := i
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err := q.Enqueue(ctx, j)
			if err != nil {
				atomic.AddInt32(&inFail, 1)
			} else {
				atomic.AddInt32(&inSucc, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	//验证出队
	assert.Equal(t, int32(goroutineNum), inSucc)
	assert.Equal(t, int32(0), inFail)
	assert.Equal(t, int32(goroutineNum), q.count)
	var wg2 sync.WaitGroup
	wg2.Add(goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			_, err := q.Dequeue(ctx)
			if err != nil {
				atomic.AddInt32(&outFail, 1)
				//atomic.
			} else {
				atomic.AddInt32(&outSucc, 1)
			}
			wg2.Done()
		}()
	}
	wg2.Wait()
	assert.Equal(t, int32(goroutineNum), outSucc)
	assert.Equal(t, int32(0), outFail)
	assert.Equal(t, int32(0), q.count)
	//fmt.Printf("insucc: %d\n", inSucc)
	//fmt.Printf("infail: %d\n", inFail)
	//fmt.Printf("outsucc: %d\n", outSucc)
	//fmt.Printf("outfail: %d\n", outFail)
}

// 高并发出入队 结果在预期时间内跑完且符合预期（中间过程无法校验）
func TestConcurrentLinkedQueueInOut(t *testing.T) {
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := NewConcurrentLinkedQueue[int]()
	var wg sync.WaitGroup
	goroutineNum := 100000
	wg.Add(goroutineNum)
	for i := 0; i < goroutineNum/2; i++ {
		j := i
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			err := q.Enqueue(ctx, j)
			if err != nil {
				atomic.AddInt32(&inFail, 1)
			} else {
				atomic.AddInt32(&inSucc, 1)
			}
			wg.Done()
		}()
	}
	for i := 0; i < goroutineNum/2; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			_, err := q.Dequeue(ctx)
			if err != nil {
				atomic.AddInt32(&outFail, 1)
				//atomic.
			} else {
				atomic.AddInt32(&outSucc, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	//assert.Equal(t, int32(goroutineNum), outSucc)
	//assert.Equal(t, int32(0), outFail)
	//assert.Equal(t, int32(0), q.count)
	fmt.Printf("insucc: %d\n", inSucc)
	fmt.Printf("infail: %d\n", inFail)
	fmt.Printf("outsucc: %d\n", outSucc)
	fmt.Printf("outfail: %d\n", outFail)
}
