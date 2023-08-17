package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	go_generic_tools "github.com/xuhaidong1/go-generic-tools"
	"github.com/xuhaidong1/go-generic-tools/container/errs"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	data := []int{6, 5, 4, 3, 2, 1}
	testCases := []struct {
		name     string
		q        *PriorityQueue[int]
		maxsize  int
		data     []int
		expected []int
	}{
		{
			name:     "新建队列",
			q:        NewPriorityQueue[int](len(data), go_generic_tools.ComparatorRealNumber[int]),
			maxsize:  len(data),
			data:     data,
			expected: []int{1, 2, 3, 4, 5, 6},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, 0, tc.q.Len())
			for _, d := range data {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
				defer cancel()
				err := tc.q.Enqueue(ctx, d)
				assert.NoError(t, err)
				if err != nil {
					return
				}
			}
			assert.Equal(t, tc.maxsize, tc.q.Len())
			assert.Equal(t, len(data), tc.q.Len())
			res := make([]int, 0, len(data))
			for tc.q.Len() > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
				defer cancel()
				el, err := tc.q.Dequeue(ctx)
				assert.NoError(t, err)
				if err != nil {
					return
				}
				res = append(res, el)
			}
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestPriorityQueue_Dequeue(t *testing.T) {
	testCases := []struct {
		name    string
		maxsize int
		data    []int
		wantErr error
	}{
		{
			name:    "有数据",
			maxsize: 6,
			data:    []int{6, 5, 4, 3, 2, 1},
			wantErr: errs.NewErrEmptyQueue(),
		},
		{
			name:    "无数据",
			maxsize: 6,
			data:    []int{},
			wantErr: errs.NewErrEmptyQueue(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewPriorityQueue[int](tc.maxsize, go_generic_tools.ComparatorRealNumber[int])
			for _, el := range tc.data {
				err := q.Enqueue(context.Background(), el)
				require.NoError(t, err)
			}
			
			if tc.name == "有数据" {
				err := q.Enqueue(context.Background(), 9)
				assert.Equal(t, errs.NewErrFullQueue(), err)
			}
			for q.Len() > 0 {
				_, err := q.Dequeue(context.Background())
				assert.NoError(t, err)
			}
			_, err := q.Dequeue(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})

	}
}

func TestPriorityQueue_EnqueueElement(t *testing.T) {
	testCases := []struct {
		name      string
		data      []int
		element   int
		wantSlice []int
	}{
		{
			name:      "新加入的元素是最大的",
			data:      []int{10, 8, 7, 6, 2},
			element:   20,
			wantSlice: []int{2, 6, 8, 10, 7, 20},
		},
		{
			name:      "新加入的元素是最小的",
			data:      []int{10, 8, 7, 6, 2},
			element:   1,
			wantSlice: []int{1, 6, 2, 10, 7, 8},
		},
		{
			name:      "新加入的元素子区间中",
			data:      []int{10, 8, 7, 6, 2},
			element:   5,
			wantSlice: []int{2, 6, 5, 10, 7, 8},
		},
		{
			name:      "新加入的元素与已有元素相同",
			data:      []int{10, 8, 7, 6, 2},
			element:   6,
			wantSlice: []int{2, 6, 6, 10, 7, 8},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewPriorityQueue[int](10, go_generic_tools.ComparatorRealNumber[int])
			for _, d := range tc.data {
				err := q.Enqueue(context.Background(), d)
				if err != nil {
					require.NoError(t, err)
				}
			}
			require.NotNil(t, q)
			err := q.Enqueue(context.Background(), tc.element)
			require.NoError(t, err)
			assert.Equal(t, tc.wantSlice, q.data)
		})

	}
}
