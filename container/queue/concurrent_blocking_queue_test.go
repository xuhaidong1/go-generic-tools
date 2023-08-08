package queue

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewConcurrentBlockingQueue_Enqueue(t *testing.T) {
	testCases := []struct {
		name     string
		q        *ConcurrentBlockingQueue[int]
		value    int
		wantErr  error
		wantData []int
		timeout  time.Duration
	}{
		{
			name:     "入队",
			q:        NewConcurrentBlockingQueue[int](10),
			value:    1,
			wantData: []int{1},
			timeout:  time.Minute,
		},

		{
			name: "入队阻塞 然后超时",
			q: func() *ConcurrentBlockingQueue[int] {
				res := NewConcurrentBlockingQueue[int](2)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := res.Enqueue(ctx, 1)
				require.NoError(t, err)
				err = res.Enqueue(ctx, 2)
				require.NoError(t, err)
				return res
			}(),
			value:    3,
			wantData: []int{1, 2},
			timeout:  time.Second,
			wantErr:  context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			err := tc.q.Enqueue(ctx, tc.value)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, tc.q.data)
		})
	}
}

func TestNewConcurrentBlockingQueue_Dequeue(t *testing.T) {
	testCases := []struct {
		name      string
		q         *ConcurrentBlockingQueue[int]
		wantValue int
		wantErr   error
		wantData  []int
		timeout   time.Duration
	}{
		{
			name: "出队",
			q: func() *ConcurrentBlockingQueue[int] {
				res := NewConcurrentBlockingQueue[int](2)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := res.Enqueue(ctx, 1)
				require.NoError(t, err)
				err = res.Enqueue(ctx, 2)
				require.NoError(t, err)
				return res
			}(),
			wantValue: 1,
			wantData:  []int{2},
			timeout:   time.Minute,
			wantErr:   nil,
		},

		{
			name:      "出队阻塞 然后超时",
			q:         NewConcurrentBlockingQueue[int](2),
			wantValue: 0,
			wantData:  []int{},
			timeout:   time.Second,
			wantErr:   context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			got, err := tc.q.Dequeue(ctx)
			assert.Equal(t, tc.wantValue, got)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, tc.q.data)
		})
	}
}

// 入队阻塞n 然后出队x个 然后入队成功n-x个 剩余入队请求超时
func TestConcurrentBlockingQueueInblock(t *testing.T) {
	//初始化队列size 20（满），想入50，出10，最终应该入成功10，入失败40，出成功10，出失败0
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := func() *ConcurrentBlockingQueue[int] {
		res := NewConcurrentBlockingQueue[int](20)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for i := 0; i < 20; i++ {
			err := res.Enqueue(ctx, i)
			require.NoError(t, err)
		}
		return res
	}()
	var wg sync.WaitGroup
	wg.Add(60)
	for i := 20; i < 70; i++ {
		j := i
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
	for i := 0; i < 10; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := q.Dequeue(ctx)
			if err != nil {
				atomic.AddInt32(&outFail, 1)
			} else {
				atomic.AddInt32(&outSucc, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 20, q.Len())
	fmt.Print(q.data)
	assert.Equal(t, int32(10), inSucc)
	assert.Equal(t, int32(40), inFail)
	assert.Equal(t, int32(10), outSucc)
	assert.Equal(t, int32(0), outFail)
}

// 入队阻塞n 然后出队x个 然后入队成功n-x个 剩余入队请求超时
func TestConcurrentBlockingQueueInblockVisable(t *testing.T) {
	//初始化队列size 20（满），想入30，出10，最终应该入成功10，入失败20，出成功10，出失败0（超时时间足够长
	var inarr uint64 = 0 //自右向左看，1-入队成功，0-未知入队情况（若位置小于lenq则是入队失败，大于lenq则是还没操作）
	var outarr uint64 = 0
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := func() *ConcurrentBlockingQueue[int] {
		res := NewConcurrentBlockingQueue[int](20)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for i := 0; i < 20; i++ {
			err := res.Enqueue(ctx, i)
			atomic.AddUint64(&inarr, 1<<i)
			require.NoError(t, err)
		}
		return res
	}()
	var wg sync.WaitGroup
	wg.Add(40)
	for i := 20; i < 50; i++ {
		j := i
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := q.Enqueue(ctx, j)
			if err != nil {
				atomic.AddInt32(&inFail, 1)

			} else {
				atomic.AddInt32(&inSucc, 1)
				atomic.AddUint64(&inarr, 1<<j)
			}
			wg.Done()
		}()
	}
	for i := 0; i < 10; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			res, err := q.Dequeue(ctx)
			if err != nil {
				atomic.AddInt32(&outFail, 1)
			} else {
				atomic.AddInt32(&outSucc, 1)
				atomic.AddUint64(&outarr, 1<<res)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 20, q.Len())
	fmt.Println(q.data)
	fmt.Printf("inarr:%b\n", inarr)
	fmt.Printf("outarr:%b\n", outarr)
	assert.Equal(t, int32(10), inSucc)
	assert.Equal(t, int32(20), inFail)
	assert.Equal(t, int32(10), outSucc)
	assert.Equal(t, int32(0), outFail)
}

// 出队阻塞n个 然后入队x（x<n）个 然后出队成功x个，n-x个出队请求超时
func TestConcurrentBlockingQueueOutBLock(t *testing.T) {
	//初始化队列size 20（空），想入30，出50，最终应该入成功30，入失败0，出成功30，出失败20（超时时间够长
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := func() *ConcurrentBlockingQueue[int] {
		res := NewConcurrentBlockingQueue[int](20)
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		//defer cancel()
		//for i := 0; i < 20; i++ {
		//	err := res.Enqueue(ctx, i)
		//	require.NoError(t, err)
		//}
		return res
	}()
	var wg sync.WaitGroup
	wg.Add(80)
	for i := 0; i < 30; i++ {
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
	for i := 0; i < 50; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			_, err := q.Dequeue(ctx)
			if err != nil {
				atomic.AddInt32(&outFail, 1)
			} else {
				atomic.AddInt32(&outSucc, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 0, q.Len())
	fmt.Print(q.data)
	assert.Equal(t, int32(30), inSucc)
	assert.Equal(t, int32(0), inFail)
	assert.Equal(t, int32(30), outSucc)
	assert.Equal(t, int32(20), outFail)
}

// 高并发出队入队同时进行，结果没有死锁（中间过程无法校验）
func TestConcurrentBlockingQueueInOut(t *testing.T) {
	//初始化队列maxsize 200，想入1000，出1000
	var inSucc int32 = 0
	var inFail int32 = 0
	var outSucc int32 = 0
	var outFail int32 = 0
	q := func() *ConcurrentBlockingQueueV1[int] {
		res := NewConcurrentBlockingQueueV1[int](200)
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		//defer cancel()
		//for i := 0; i < 20; i++ {
		//	err := res.Enqueue(ctx, i)
		//	require.NoError(t, err)
		//}
		return res
	}()
	var wg sync.WaitGroup
	wg.Add(20000)
	for i := 0; i < 10000; i++ {
		j := i
		go func() {
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
	for i := 0; i < 10000; i++ {
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
			wg.Done()
		}()
	}
	wg.Wait()
	//assert.Equal(t, 20, q.Len())
	fmt.Println(q.data)
	//assert.Equal(t, int32(10), inSucc)
	//assert.Equal(t, int32(40), inFail)
	//assert.Equal(t, int32(10), outSucc)
	//assert.Equal(t, int32(0), outFail)
	fmt.Printf("insucc: %d\n", inSucc)
	fmt.Printf("infail: %d\n", inFail)
	fmt.Printf("outsucc: %d\n", outSucc)
	fmt.Printf("outfail: %d\n", outFail)

}
