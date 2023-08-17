package queue

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type DelayElem struct {
	deadline time.Time
	val      int
}

func NewDelayElem(val int, duration time.Duration) DelayElem {
	return DelayElem{
		deadline: time.Now().Add(duration),
		val:      val,
	}
}

func (d DelayElem) Delay() time.Duration {
	return d.deadline.Sub(time.Now())
}
func TestDelayQueue_Enqueue(t *testing.T) {
	type testCases struct {
		name    string
		elems   []DelayElem
		d       *DelayQueue[DelayElem]
		wantErr error
	}
	tests := []testCases{
		{
			name: "批量入队",
			elems: []DelayElem{
				NewDelayElem(132, 100*time.Second), NewDelayElem(23, 500*time.Second), NewDelayElem(45, 240*time.Second),
				NewDelayElem(6, 400*time.Second), NewDelayElem(42, 200*time.Second), NewDelayElem(12, 600*time.Second),
				NewDelayElem(71, 400*time.Second), NewDelayElem(54, 700*time.Second), NewDelayElem(91, 900*time.Second),
			},
			d:       NewDelayQueue[DelayElem](10),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, e := range tt.elems {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := tt.d.Enqueue(ctx, e)
				require.NoError(t, err)
			}
			for i := 0; i < tt.d.Len(); i++ {
				fmt.Printf("val:%d,delay:%v\n", tt.d.q.data[i].val, tt.d.q.data[i].Delay())
			}
			fmt.Println("-------------------------------------------------")
			for tt.d.q.Len() > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				dequeue, err := tt.d.q.Dequeue(ctx)
				require.NoError(t, err)
				fmt.Printf("val:%d,delay:%v\n", dequeue.val, dequeue.Delay())
			}
		})
	}
}

func TestDelayQueue_Dequeue(t *testing.T) {
	type testCases struct {
		name     string
		elems    []DelayElem
		d        *DelayQueue[DelayElem]
		wantvals []int
		wantErr  error
	}
	tests := []testCases{
		{
			name: "批量按规定时间出队",
			elems: []DelayElem{
				//NewDelayElem(132, 1*time.Second), NewDelayElem(23, 500*time.Second), NewDelayElem(45, 240*time.Second),
				//NewDelayElem(6, 400*time.Second), NewDelayElem(42, 200*time.Second), NewDelayElem(12, 600*time.Second),
				//NewDelayElem(71, 400*time.Second), NewDelayElem(54, 700*time.Second), NewDelayElem(91, 900*time.Second),
				NewDelayElem(132, 100*time.Millisecond), NewDelayElem(23, 500*time.Millisecond), NewDelayElem(45, 240*time.Millisecond),
				NewDelayElem(6, 400*time.Millisecond), NewDelayElem(42, 200*time.Millisecond), NewDelayElem(12, 600*time.Millisecond),
				NewDelayElem(71, 440*time.Millisecond), NewDelayElem(54, 700*time.Millisecond), NewDelayElem(91, 900*time.Millisecond),
			},
			d:        NewDelayQueue[DelayElem](10),
			wantvals: []int{132, 42, 45, 6, 71, 23, 12, 54, 91},
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now() //开始时间统一计算了。。
			for _, e := range tt.elems {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := tt.d.Enqueue(ctx, e)
				require.NoError(t, err)
			}
			timemiss := time.Now().Sub(start)
			//for i := 0; i < tt.d.Len(); i++ {
			//	fmt.Printf("val:%d,delay:%v\n", tt.d.q.data[i].val, tt.d.q.data[i].Delay())
			//}
			var outtime []time.Duration
			var outres []int
			fmt.Println(tt.d.Len())
			for tt.d.Len() > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
				defer cancel()
				dequeue, err := tt.d.Dequeue(ctx)
				outtime = append(outtime, time.Now().Sub(start))
				outres = append(outres, dequeue.val)
				require.NoError(t, err)
			}
			fmt.Printf("timemiss:%v\n", timemiss)
			fmt.Printf("outtime:%v\n", outtime)
			assert.Equal(t, tt.wantvals, outres)
		})
	}
}

func TestDelayQueue_DequeueInsert(t *testing.T) {
	type testCases struct {
		name     string
		elems    []DelayElem
		d        *DelayQueue[DelayElem]
		wantvals []int
		//insert   DelayElem
		wantErr error
	}
	tests := []testCases{
		{
			name: "出队睡觉时有delay时长更短的元素加入，需要减少睡觉时间",
			elems: []DelayElem{
				//NewDelayElem(132, 1*time.Second), NewDelayElem(23, 500*time.Second), NewDelayElem(45, 240*time.Second),
				//NewDelayElem(6, 400*time.Second), NewDelayElem(42, 200*time.Second), NewDelayElem(12, 600*time.Second),
				//NewDelayElem(71, 400*time.Second), NewDelayElem(54, 700*time.Second), NewDelayElem(91, 900*time.Second),
				NewDelayElem(132, 100*time.Millisecond), NewDelayElem(23, 500*time.Millisecond), NewDelayElem(45, 250*time.Millisecond),
			},
			d: NewDelayQueue[DelayElem](10),
			//insert:   NewDelayElem(489, 100*time.Millisecond), elem的保质期从new的一刻就开始减少了
			wantvals: []int{132, 45, 489, 23},
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now() //开始时间统一计算了。。
			for _, e := range tt.elems {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := tt.d.Enqueue(ctx, e)
				require.NoError(t, err)
			}
			timemiss := time.Now().Sub(start)
			//for i := 0; i < tt.d.Len(); i++ {
			//	fmt.Printf("val:%d,delay:%v\n", tt.d.q.data[i].val, tt.d.q.data[i].Delay())
			//}
			var outtime []time.Duration
			var outres []int
			fmt.Println(tt.d.Len())
			//在这里进行delay时间更短的元素的插入
			go func() {
				time.Sleep(300 * time.Millisecond)
				ctxx, cancell := context.WithTimeout(context.Background(), 1*time.Second)
				err := tt.d.Enqueue(ctxx, NewDelayElem(489, 97*time.Millisecond))
				require.NoError(t, err)
				cancell()
			}()
			for tt.d.Len() > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
				dequeue, err := tt.d.Dequeue(ctx)
				outtime = append(outtime, time.Now().Sub(start))
				outres = append(outres, dequeue.val)
				require.NoError(t, err)

				cancel()
			}
			fmt.Printf("timemiss:%v\n", timemiss)
			//100 250 400 500
			fmt.Printf("outtime:%v\n", outtime)
			assert.Equal(t, tt.wantvals, outres)
		})
	}
}

func TestDelayQueue_DequeueTimeout(t *testing.T) {
	type testCases struct {
		name     string
		elems    []DelayElem
		d        *DelayQueue[DelayElem]
		wantvals []int
		//insert   DelayElem
		wantErr error
	}
	tests := []testCases{
		{
			name: "个别元素超时提前出队，不影响后面元素的出队时间",
			elems: []DelayElem{
				//NewDelayElem(132, 1*time.Second), NewDelayElem(23, 500*time.Second), NewDelayElem(45, 240*time.Second),
				//NewDelayElem(6, 400*time.Second), NewDelayElem(42, 200*time.Second), NewDelayElem(12, 600*time.Second),
				//NewDelayElem(71, 400*time.Second), NewDelayElem(54, 700*time.Second), NewDelayElem(91, 900*time.Second),
				NewDelayElem(132, 100*time.Millisecond), NewDelayElem(489, 400*time.Millisecond), NewDelayElem(23, 500*time.Millisecond), NewDelayElem(45, 150*time.Millisecond),
			},
			d: NewDelayQueue[DelayElem](10),
			//insert:   NewDelayElem(489, 100*time.Millisecond), elem的保质期从new的一刻就开始减少了
			wantvals: []int{132, 45, 489, 23},
			wantErr:  context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//start := time.Now() //开始时间统一计算了。。
			for _, e := range tt.elems {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := tt.d.Enqueue(ctx, e)
				require.NoError(t, err)
			}
			start := time.Now()
			//timemiss := time.Now().Sub(start)
			var outtime []time.Duration
			var outres []int
			fmt.Println(tt.d.Len())
			//在这里进行delay时间更短的元素的插入
			//go func() {
			//	time.Sleep(300 * time.Millisecond)
			//	ctxx, cancell := context.WithTimeout(context.Background(), 1*time.Second)
			//	err := tt.d.Enqueue(ctxx, NewDelayElem(489, 100*time.Millisecond))
			//	require.NoError(t, err)
			//	cancell()
			//}()
			for i := 0; i < 4; i++ {
				if i == 2 {
					//此时队列头部val应该是489，应该在400ms出队，走到这个分支时刻应该是150ms 第2个元素刚出去
					//在180ms超时
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
					dequeue, err := tt.d.Dequeue(ctx)
					//应该是加了180ms这个点
					outtime = append(outtime, time.Now().Sub(start))
					outres = append(outres, dequeue.val)
					assert.Error(t, err, context.DeadlineExceeded)
					cancel()
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
				dequeue, err := tt.d.Dequeue(ctx)
				outtime = append(outtime, time.Now().Sub(start))
				outres = append(outres, dequeue.val)
				require.NoError(t, err)
				cancel()
			}
			//fmt.Printf("timemiss:%v\n", timemiss)
			//100 150 400 500
			//400这个超时了 要在180的时候就出队。 结果100 150 180 500
			fmt.Printf("outtime:%v\n", outtime)
			assert.Equal(t, tt.wantvals, outres)
		})
	}
}
